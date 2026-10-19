package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"schedule-system-v2/backend/internal/update"
)

// ==================== 阶段与进度 ====================

func TestPhaseProgress(t *testing.T) {
	cases := map[string]struct {
		want int
		desc string
	}{
		phaseVerifyPkg:  {10, "校验安装包"},
		phaseStopping:   {20, "停止服务"},
		phaseBackingUp:  {45, "备份"},
		phaseInstalling: {70, "安装"},
		phaseRestarting: {90, "重启"},
		phaseVerifying:  {96, "验证服务"},
		phaseDone:       {100, "完成"},
		phaseFailed:     {0, "失败归零"},
		phaseRolledBack: {0, "回滚归零"},
	}
	for phase, tc := range cases {
		if got := phaseProgress(phase); got != tc.want {
			t.Errorf("phaseProgress(%s) = %d, want %d（%s）", phase, got, tc.want, tc.desc)
		}
	}
}

func TestSetPhase(t *testing.T) {
	u := &Updater{}

	for _, phase := range []string{phaseVerifyPkg, phaseStopping, phaseBackingUp, phaseInstalling, phaseRestarting, phaseVerifying} {
		u.setPhase(phase, "msg-"+phase)
		s := u.snapshot()
		if s.Phase != phase || s.Message != "msg-"+phase {
			t.Errorf("setPhase(%s) 后 phase=%s message=%s", phase, s.Phase, s.Message)
		}
		if s.Progress <= 0 {
			t.Errorf("阶段 %s 的进度应大于 0，实际 %d", phase, s.Progress)
		}
	}

	// 失败阶段：进度归零、错误保留
	u.setPhase(phaseInstalling, "安装中")
	u.fail("模拟失败")
	s := u.snapshot()
	if s.Phase != phaseFailed || s.Error != "模拟失败" {
		t.Errorf("fail 后 phase=%s error=%q", s.Phase, s.Error)
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
		target: "v3.0.1",
	}
	u.writeResult(phaseDone, "更新成功，已升级到 v3.0.1")

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

// ==================== 镜像列表处理 ====================

func TestSplitMirrors(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"   ", nil},
		{"https://a.com", []string{"https://a.com"}},
		{"https://a.com, https://b.com/", []string{"https://a.com", "https://b.com/"}},
		{" https://a.com ,, https://b.com ", []string{"https://a.com", "https://b.com"}},
	}
	for _, c := range cases {
		got := splitMirrors(c.in)
		if len(got) != len(c.want) {
			t.Errorf("splitMirrors(%q) = %v, want %v", c.in, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("splitMirrors(%q)[%d] = %q, want %q", c.in, i, got[i], c.want[i])
			}
		}
	}
}

func TestPrioritizeMirror(t *testing.T) {
	mirrors := []string{"https://a.com", "https://b.com", "https://c.com"}

	// 上次成功的镜像移到最前，其余保序
	got := prioritizeMirror(mirrors, "https://c.com")
	want := []string{"https://c.com", "https://a.com", "https://b.com"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("prioritizeMirror = %v, want %v", got, want)
		}
	}

	// 空 best 时原样返回
	got = prioritizeMirror(mirrors, "")
	if len(got) != 3 || got[0] != "https://a.com" {
		t.Errorf("空 best 应原样返回，实际 %v", got)
	}
}

// ==================== 下载阶段状态文件 ====================

func TestDownloaderFailWritesState(t *testing.T) {
	dir := t.TempDir()
	d := &Downloader{
		dir: dir,
		state: update.DownloadState{
			Phase:   update.DownloadPhaseChecking,
			Current: "v3.0.1",
			Target:  "v3.0.2",
		},
	}
	if code := d.fail("检查更新失败: boom"); code != 1 {
		t.Errorf("fail 退出码 = %d, want 1", code)
	}

	st := update.ReadDownloadState(dir)
	if st == nil {
		t.Fatal("失败后应写入状态文件")
	}
	if st.Phase != update.DownloadPhaseFailed || st.Error != "检查更新失败: boom" {
		t.Errorf("状态 = %s / %q, want failed / 检查更新失败: boom", st.Phase, st.Error)
	}
	if !st.TerminalPhase() {
		t.Error("失败应为终态")
	}
	if st.FinishedAt == "" {
		t.Error("失败应记录 finishedAt")
	}
}

func TestDownloaderSetPhaseFlushes(t *testing.T) {
	dir := t.TempDir()
	d := &Downloader{dir: dir, state: update.DownloadState{Phase: update.DownloadPhaseChecking}}
	d.setPhase(update.DownloadPhaseDownloading, "正在下载更新包… 42%")

	st := update.ReadDownloadState(dir)
	if st == nil || st.Phase != update.DownloadPhaseDownloading {
		t.Fatalf("setPhase 后应立即落盘，实际 %+v", st)
	}
	if st.Message != "正在下载更新包… 42%" {
		t.Errorf("message = %q", st.Message)
	}
	if st.Error != "" {
		t.Errorf("非失败阶段 error 应为空，实际 %q", st.Error)
	}
}

func TestDownloadStateStale(t *testing.T) {
	now, err := time.Parse("2006-01-02 15:04:05", "2026-10-17 12:00:00")
	if err != nil {
		t.Fatalf("解析基准时间: %v", err)
	}

	fresh := &update.DownloadState{Phase: update.DownloadPhaseDownloading, UpdatedAt: "2026-10-17 11:59:50"}
	if update.DownloadStateStale(fresh, now) {
		t.Error("10 秒内刷新过不应判停滞")
	}
	stale := &update.DownloadState{Phase: update.DownloadPhaseDownloading, UpdatedAt: "2026-10-17 11:59:00"}
	if !update.DownloadStateStale(stale, now) {
		t.Error("超过 30 秒未刷新应判停滞")
	}
	// 终态不判停滞
	ready := &update.DownloadState{Phase: update.DownloadPhaseReady, UpdatedAt: "2026-10-17 10:00:00"}
	if update.DownloadStateStale(ready, now) {
		t.Error("终态不应判停滞")
	}
	// 时间戳不可解析按停滞处理
	broken := &update.DownloadState{Phase: update.DownloadPhaseDownloading, UpdatedAt: "not-a-time"}
	if !update.DownloadStateStale(broken, now) {
		t.Error("时间戳不可解析应按停滞处理")
	}
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
