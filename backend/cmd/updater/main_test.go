package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// ==================== 阶段与进度 ====================

func TestPhaseProgress(t *testing.T) {
	cases := map[string]struct {
		want int
		desc string
	}{
		phaseStopping:    {5, "停止服务"},
		phaseChecking:    {12, "检查更新"},
		phaseDownloading: {0, "下载进度由字节数驱动，不设固定值"},
		phaseBackingUp:   {70, "备份"},
		phaseInstalling:  {85, "安装"},
		phaseRestarting:  {95, "重启"},
		phaseVerifying:   {98, "验证"},
		phaseDone:        {100, "完成"},
		phaseFailed:      {0, "失败归零"},
		phaseAborted:     {0, "中止归零"},
	}
	for phase, tc := range cases {
		if got := phaseProgress(phase); got != tc.want {
			t.Errorf("phaseProgress(%s) = %d, want %d（%s）", phase, got, tc.want, tc.desc)
		}
	}
}

func TestSetPhase(t *testing.T) {
	u := &Updater{}

	// 下载/备份/安装阶段允许中止（用户可随时终止更新恢复旧版本）
	for _, phase := range []string{phaseChecking, phaseDownloading, phaseBackingUp, phaseInstalling} {
		u.setPhase(phase, "msg-"+phase)
		s := u.snapshot()
		if s.Phase != phase || s.Message != "msg-"+phase {
			t.Errorf("setPhase(%s) 后 phase=%s message=%s", phase, s.Phase, s.Message)
		}
		if !s.CanAbort {
			t.Errorf("阶段 %s 应允许中止", phase)
		}
	}

	// 重启/验证阶段禁止中止（产物已替换，中止语义不安全）
	for _, phase := range []string{phaseRestarting, phaseVerifying} {
		u.setPhase(phase, "msg-"+phase)
		if u.snapshot().CanAbort {
			t.Errorf("阶段 %s 不应允许中止", phase)
		}
	}

	// 失败阶段保持可中止（等价于恢复旧版本）
	u.setPhase(phaseDownloading, "下载中")
	u.fail("模拟失败")
	s := u.snapshot()
	if s.Phase != phaseFailed || s.Error != "模拟失败" {
		t.Errorf("fail 后 phase=%s error=%q", s.Phase, s.Error)
	}
	if !s.CanAbort {
		t.Error("失败阶段应允许中止（恢复旧版本）")
	}
	if s.Progress != 0 {
		t.Errorf("失败后 progress = %d, want 0", s.Progress)
	}
}

// ==================== 日志环形缓冲 ====================

func TestLogRing(t *testing.T) {
	u := &Updater{}

	// 超过 500 条后内部只保留末尾 500 条
	for i := 0; i < 520; i++ {
		u.log("line")
	}
	u.mu.Lock()
	internal := len(u.state.Logs)
	u.mu.Unlock()
	if internal != 500 {
		t.Fatalf("内部日志应裁剪到 500 条，实际 %d", internal)
	}

	// 快照对外最多暴露 200 条
	if got := len(u.snapshot().Logs); got != 200 {
		t.Errorf("快照日志应裁剪到 200 条，实际 %d", got)
	}

	// 快照必须是拷贝：修改快照不得影响内部状态
	s := u.snapshot()
	s.Logs[0] = "MUTATED"
	if u.snapshot().Logs[0] == "MUTATED" {
		t.Error("snapshot 返回了内部切片的引用，存在数据竞争风险")
	}
}

func TestLogfFormat(t *testing.T) {
	u := &Updater{}
	u.logf("版本 %s -> %s（%d MB）", "v3.0.0", "v3.0.1", 16)
	got := u.snapshot().Logs
	if len(got) != 1 {
		t.Fatalf("应有 1 条日志，实际 %d", len(got))
	}
	want := "版本 v3.0.0 -> v3.0.1（16 MB）"
	if got[0][len(got[0])-len(want):] != want {
		t.Errorf("日志内容 = %q, 期望末尾为 %q", got[0], want)
	}
}

// ==================== 结果落盘 ====================

func TestWriteResult(t *testing.T) {
	u := &Updater{
		dir: t.TempDir(),
		state: State{
			CurrentVersion: "v3.0.0",
			TargetVersion:  "v3.0.1",
		},
	}
	u.writeResult(phaseDone, "更新成功，已升级到 v3.0.1", nil)

	data, err := os.ReadFile(filepath.Join(u.dir, "updates", "last-result.json"))
	if err != nil {
		t.Fatalf("读取 last-result.json 失败: %v", err)
	}
	var r LastResult
	if err := json.Unmarshal(data, &r); err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if r.Phase != phaseDone || r.From != "v3.0.0" || r.To != "v3.0.1" {
		t.Errorf("结果内容 = %+v", r)
	}
	if r.At == "" {
		t.Error("结果缺少时间戳")
	}
}

// ==================== 中止语义 ====================

func TestRequestAbortGating(t *testing.T) {
	u := &Updater{abortCh: make(chan struct{})}

	// 不可中止阶段：点击无效
	u.setPhase(phaseRestarting, "重启中")
	u.requestAbort()
	if u.abortRequested() {
		t.Fatal("重启阶段不应响应中止请求")
	}

	// 可中止阶段：点击生效
	u.setPhase(phaseDownloading, "下载中")
	u.requestAbort()
	if !u.abortRequested() {
		t.Fatal("下载阶段应响应中止请求")
	}

	// 重复点击不 panic（abortOnce 保护）
	u.requestAbort()
	u.requestAbort()
}

func TestResolveDir(t *testing.T) {
	// 指定目录时原样使用（转绝对路径）
	if got := resolveDir("."); got == "" {
		t.Error("resolveDir(\".\") 不应为空")
	}
	// 未指定时取可执行文件目录
	if got := resolveDir(""); got == "" {
		t.Error("resolveDir(\"\") 不应为空")
	}
}
