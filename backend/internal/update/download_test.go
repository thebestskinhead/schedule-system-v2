package update

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------- 镜像地址拼接 ----------

func TestMirrorURL(t *testing.T) {
	raw := "https://github.com/o/r/releases/download/v3.2.0/pkg.zip"
	cases := []struct {
		mirror string
		want   string
	}{
		{"https://ghfast.top", "https://ghfast.top/https://github.com/o/r/releases/download/v3.2.0/pkg.zip"},
		{"https://ghfast.top/", "https://ghfast.top/https://github.com/o/r/releases/download/v3.2.0/pkg.zip"},
		{" https://ghproxy.com/ ", "https://ghproxy.com/https://github.com/o/r/releases/download/v3.2.0/pkg.zip"},
	}
	for _, c := range cases {
		if got := MirrorURL(c.mirror, raw); got != c.want {
			t.Errorf("MirrorURL(%q) = %q, want %q", c.mirror, got, c.want)
		}
	}
}

// ---------- 镜像回退下载 ----------

// newPkgServer 启动一个返回固定内容的下载服务器，返回服务器与内容 sha256
func newPkgServer(t *testing.T, content string) (*httptest.Server, string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(content))
	}))
	sum := sha256.Sum256([]byte(content))
	return srv, hex.EncodeToString(sum[:])
}

func TestDownloadWithMirrorsFirstMirrorOK(t *testing.T) {
	good, sum := newPkgServer(t, "GOOD-PACKAGE-CONTENT")
	defer good.Close()
	_, _ = newPkgServer(t, "unused")

	dest := filepath.Join(t.TempDir(), "pkg.zip")
	var attempts []string
	used, err := DownloadWithMirrors(good.URL+"/pkg.zip", []string{good.URL}, dest,
		nil, nil, func(p string) error { return VerifyChecksum(p, sum) },
		func(mirror string, err error) { attempts = append(attempts, mirror) })
	if err != nil {
		t.Fatalf("DownloadWithMirrors: %v", err)
	}
	if used != good.URL {
		t.Errorf("usedMirror = %q, want %q", used, good.URL)
	}
	if len(attempts) != 0 {
		t.Errorf("首个镜像即成功时不应有失败尝试，实际 %v", attempts)
	}
	data, err := os.ReadFile(dest)
	if err != nil || string(data) != "GOOD-PACKAGE-CONTENT" {
		t.Errorf("下载内容 = %q (err=%v)", data, err)
	}
}

func TestDownloadWithMirrorsChecksumMismatchSwitchesMirror(t *testing.T) {
	bad, _ := newPkgServer(t, "CORRUPT-PACKAGE-CONTENT")
	defer bad.Close()
	good, sum := newPkgServer(t, "GOOD-PACKAGE-CONTENT")
	defer good.Close()

	dest := filepath.Join(t.TempDir(), "pkg.zip")
	var attempts []string
	used, err := DownloadWithMirrors(good.URL+"/pkg.zip", []string{bad.URL, good.URL}, dest,
		nil, nil, func(p string) error { return VerifyChecksum(p, sum) },
		func(mirror string, err error) { attempts = append(attempts, mirror) })
	if err != nil {
		t.Fatalf("DownloadWithMirrors: %v", err)
	}
	// 校验不符必须切换下一个镜像，而不是把损坏包装进去
	if used != good.URL {
		t.Errorf("usedMirror = %q, want %q（损坏镜像应被跳过）", used, good.URL)
	}
	if len(attempts) != 1 || attempts[0] != bad.URL {
		t.Errorf("失败尝试 = %v, want [%s]", attempts, bad.URL)
	}
	data, _ := os.ReadFile(dest)
	if string(data) != "GOOD-PACKAGE-CONTENT" {
		t.Errorf("最终内容 = %q, want 好镜像的内容", data)
	}
}

func TestDownloadWithMirrorsAllFailFallsBackToDirect(t *testing.T) {
	bad, _ := newPkgServer(t, "CORRUPT-PACKAGE-CONTENT")
	defer bad.Close()
	direct, sum := newPkgServer(t, "DIRECT-PACKAGE-CONTENT")
	defer direct.Close()

	dest := filepath.Join(t.TempDir(), "pkg.zip")
	used, err := DownloadWithMirrors(direct.URL+"/pkg.zip", []string{bad.URL}, dest,
		nil, nil, func(p string) error { return VerifyChecksum(p, sum) }, nil)
	if err != nil {
		t.Fatalf("全部镜像失败后应回退直连: %v", err)
	}
	if used != "" {
		t.Errorf("usedMirror = %q, want 空（回退直连）", used)
	}
	data, _ := os.ReadFile(dest)
	if string(data) != "DIRECT-PACKAGE-CONTENT" {
		t.Errorf("最终内容 = %q, want 直连内容", data)
	}
}

func TestDownloadWithMirrorsUnreachableMirrorSkipped(t *testing.T) {
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	deadURL := dead.URL
	dead.Close() // 关闭后连接被拒绝，等价于镜像失效

	good, sum := newPkgServer(t, "GOOD-PACKAGE-CONTENT")
	defer good.Close()

	dest := filepath.Join(t.TempDir(), "pkg.zip")
	used, err := DownloadWithMirrors(good.URL+"/pkg.zip", []string{deadURL, good.URL}, dest,
		nil, nil, func(p string) error { return VerifyChecksum(p, sum) }, nil)
	if err != nil {
		t.Fatalf("失效镜像应被跳过: %v", err)
	}
	if used != good.URL {
		t.Errorf("usedMirror = %q, want %q", used, good.URL)
	}
}

func TestDownloadWithMirrorsNoMirrorsGoesDirect(t *testing.T) {
	direct, sum := newPkgServer(t, "DIRECT-ONLY")
	defer direct.Close()

	dest := filepath.Join(t.TempDir(), "pkg.zip")
	used, err := DownloadWithMirrors(direct.URL+"/pkg.zip", nil, dest,
		nil, nil, func(p string) error { return VerifyChecksum(p, sum) }, nil)
	if err != nil {
		t.Fatalf("DownloadWithMirrors: %v", err)
	}
	if used != "" {
		t.Errorf("无镜像配置时 usedMirror 应为空，实际 %q", used)
	}
}

// ---------- 下载状态文件 ----------

func TestDownloadStateRoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := &DownloadState{
		Phase:          DownloadPhaseReady,
		Message:        "更新包已就绪（v3.0.2），可在管理页执行安装",
		Received:       1024,
		Total:          2048,
		Current:        "v3.0.1",
		Target:         "v3.0.2",
		AssetName:      "schedule-system-v2-windows-amd64.zip",
		AssetSize:      17112839,
		Checksum:       "abcdef0123456789",
		Mirror:         "https://ghfast.top",
		LastGoodMirror: "https://ghfast.top",
		PID:            4321,
		StartedAt:      "2026-10-17 10:00:00",
		UpdatedAt:      "2026-10-17 10:05:00",
		FinishedAt:     "2026-10-17 10:05:00",
	}
	if err := WriteDownloadState(dir, want); err != nil {
		t.Fatalf("WriteDownloadState: %v", err)
	}

	got := ReadDownloadState(dir)
	if got == nil {
		t.Fatal("ReadDownloadState 返回 nil")
	}
	if *got != *want {
		t.Errorf("读回状态 = %+v, want %+v", got, want)
	}

	// 原子写：不得残留 .tmp 文件
	if _, err := os.Stat(DownloadStatePath(dir) + ".tmp"); !os.IsNotExist(err) {
		t.Error("原子写残留了 .tmp 文件")
	}
}

func TestReadDownloadStateMissingAndGarbage(t *testing.T) {
	if st := ReadDownloadState(t.TempDir()); st != nil {
		t.Errorf("无状态文件时应返回 nil，实际 %+v", st)
	}
	dir := t.TempDir()
	if err := os.MkdirAll(DownloadTmpDir(dir), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(DownloadStatePath(dir), []byte("{not json"), 0644); err != nil {
		t.Fatal(err)
	}
	if st := ReadDownloadState(dir); st != nil {
		t.Errorf("损坏的状态文件应返回 nil，实际 %+v", st)
	}
}

// ---------- 安装前预检 ----------

func TestCheckInstallPreconditions(t *testing.T) {
	dir := t.TempDir()
	pkg := filepath.Join(dir, "pkg.zip")
	if err := os.WriteFile(pkg, []byte("0123456789"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := CheckInstallPreconditions(dir, 10); err != nil {
		t.Errorf("常规目录应通过预检: %v", err)
	}
}

func TestCheckDiskSpaceInsufficient(t *testing.T) {
	dir := t.TempDir()
	// 需要一个不可能满足的空间：必须报错且信息可读
	err := CheckDiskSpace(dir, 1<<62)
	if err == nil {
		t.Fatal("空间不足应返回错误")
	}
	if !strings.Contains(err.Error(), "磁盘剩余空间不足") {
		t.Errorf("错误信息 = %q, 应包含「磁盘剩余空间不足」", err)
	}
}

func TestCheckDiskSpaceSufficient(t *testing.T) {
	dir := t.TempDir()
	if err := CheckDiskSpace(dir, 1024); err != nil {
		t.Errorf("1KB 需求应满足: %v", err)
	}
}

// ---------- 下载中止（取消标志） ----------

func TestDownloadCancelFlagPath(t *testing.T) {
	dir := t.TempDir()
	flagPath := DownloadCancelFlagPath(dir)
	if !strings.HasPrefix(flagPath, DownloadTmpDir(dir)) {
		t.Errorf("取消标志 %q 应位于下载临时目录内", flagPath)
	}
	if err := os.MkdirAll(DownloadTmpDir(dir), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(flagPath, []byte("cancel"), 0644); err != nil {
		t.Fatal(err)
	}
	if !fileExists(flagPath) {
		t.Error("取消标志写入后应存在")
	}
	_ = os.Remove(flagPath)
	if fileExists(flagPath) {
		t.Error("取消标志删除后不应存在")
	}
}
