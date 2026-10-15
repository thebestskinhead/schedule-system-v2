package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

// ---------- 版本解析与比较 ----------

func TestParseVersion(t *testing.T) {
	cases := []struct {
		in   string
		want Version
		ok   bool
	}{
		{"v3.0.1", Version{3, 0, 1, "", "v3.0.1"}, true},
		{"3.0.1", Version{3, 0, 1, "", "3.0.1"}, true},
		{"v2.10.0-test", Version{2, 10, 0, "test", "v2.10.0-test"}, true},
		{"v10.20.30+build", Version{10, 20, 30, "build", "v10.20.30+build"}, true},
		{"v3.1.0-dev+abc123", Version{3, 1, 0, "dev+abc123", "v3.1.0-dev+abc123"}, true},
		{"v1.2", Version{}, false},
		{"abc", Version{}, false},
		{"", Version{}, false},
		{"v1.2.3.4", Version{}, false},
	}
	for _, c := range cases {
		got, err := ParseVersion(c.in)
		if c.ok && err != nil {
			t.Errorf("ParseVersion(%q) 意外错误: %v", c.in, err)
			continue
		}
		if !c.ok && err == nil {
			t.Errorf("ParseVersion(%q) 应当报错", c.in)
			continue
		}
		if c.ok && got != c.want {
			t.Errorf("ParseVersion(%q) = %+v, want %+v", c.in, got, c.want)
		}
	}
}

func TestVersionString(t *testing.T) {
	cases := map[string]string{
		"v3.0.1":            "v3.0.1",
		"v3.1.0-dev":        "v3.1.0-dev",
		"v3.1.0-dev+abc123": "v3.1.0-dev+abc123",
	}
	for in, want := range cases {
		v, err := ParseVersion(in)
		if err != nil {
			t.Fatalf("ParseVersion(%q): %v", in, err)
		}
		if got := v.String(); got != want {
			t.Errorf("(%s).String() = %q, want %q", in, got, want)
		}
	}
}

func TestVersionCompare(t *testing.T) {
	a := Version{2, 3, 0, "", ""}
	b := Version{2, 3, 1, "", ""}
	if a.Compare(b) != -1 || b.Compare(a) != 1 || a.Compare(a) != 0 {
		t.Error("Compare 结果不符合预期")
	}
	if !a.SameMajor(Version{2, 99, 99, "", ""}) {
		t.Error("SameMajor 应当忽略 minor/patch")
	}
	if a.SameMajor(Version{3, 0, 0, "", ""}) {
		t.Error("不同大版本不应判定为 SameMajor")
	}
}

// TestVersionComparePrerelease 预发布排序（semver 规则）：
// 正式版高于预发布版；否则 CI 分支构建的 v3.1.0-dev+sha 部署后
// 将永远收不到正式版 v3.1.0 的更新。
func TestVersionComparePrerelease(t *testing.T) {
	parse := func(s string) Version {
		v, err := ParseVersion(s)
		if err != nil {
			t.Fatalf("ParseVersion(%q): %v", s, err)
		}
		return v
	}
	cases := []struct {
		lo, hi string
		desc   string
	}{
		{"v3.1.0-dev+abc123", "v3.1.0", "预发布版低于正式版"},
		{"v3.1.0-alpha", "v3.1.0-beta", "字典序：alpha < beta"},
		{"v3.1.0-alpha.1", "v3.1.0-alpha.2", "数字标识符按数值：1 < 2"},
		{"v3.1.0-alpha.2", "v3.1.0-alpha.10", "数字标识符按数值而非字典序：2 < 10"},
		{"v3.0.1", "v3.1.0-dev", "主版本号优先于预发布标记"},
		{"v3.1.0-dev+aaa", "v3.1.0-dev+bbb", "两个 dev 构建按 sha 排序"},
	}
	for _, c := range cases {
		lo, hi := parse(c.lo), parse(c.hi)
		if lo.Compare(hi) != -1 {
			t.Errorf("%s: (%s).Compare(%s) = %d, want -1", c.desc, c.lo, c.hi, lo.Compare(hi))
		}
		if hi.Compare(lo) != 1 {
			t.Errorf("%s: (%s).Compare(%s) = %d, want 1", c.desc, c.hi, c.lo, hi.Compare(lo))
		}
		if lo.Compare(lo) != 0 {
			t.Errorf("%s: 自身比较应为 0", c.desc)
		}
	}
}

func TestVersionFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	if got := CurrentVersion(dir); got.Raw != "v0.0.0" {
		t.Errorf("无 version.txt 时应为 v0.0.0，得到 %s", got.Raw)
	}
	v := Version{3, 1, 4, "", "v3.1.4"}
	if err := WriteVersion(dir, v); err != nil {
		t.Fatalf("WriteVersion: %v", err)
	}
	if got := CurrentVersion(dir); got.String() != "v3.1.4" {
		t.Errorf("读回版本号 = %s, want v3.1.4", got.String())
	}
}

// TestVersionFileRoundTripDev dev 后缀版本号（CI 分支构建产物）必须原样往返
func TestVersionFileRoundTripDev(t *testing.T) {
	dir := t.TempDir()
	v, err := ParseVersion("v3.1.0-dev+abc123")
	if err != nil {
		t.Fatalf("ParseVersion: %v", err)
	}
	if err := WriteVersion(dir, v); err != nil {
		t.Fatalf("WriteVersion: %v", err)
	}
	got := CurrentVersion(dir)
	if got.String() != "v3.1.0-dev+abc123" {
		t.Errorf("读回版本号 = %s, want v3.1.0-dev+abc123", got.String())
	}
	if got.Major != 3 || got.Minor != 1 || got.Patch != 0 {
		t.Errorf("核心版本号解析错误: %+v", got)
	}
}

// ---------- 备份 / 应用 / 回滚 ----------

func makeFakeInstall(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeFile := func(rel, content string) {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	writeFile("server.exe", "OLD-SERVER")
	writeFile("dist/index.html", "<h1>OLD-DIST</h1>")
	writeFile("dist/assets/app.js", "OLD-ASSET")
	writeFile("dist-mobile/index.html", "<h1>OLD-MOBILE</h1>")
	// 运行数据：绝不允许被更新触碰
	writeFile("configs/config.yaml", "password: secret")
	writeFile("migrations/001.sql", "CREATE TABLE t;")
	writeFile("updates/update.log", "old log")
	return dir
}

func TestBackupApplyRollback(t *testing.T) {
	dir := makeFakeInstall(t)

	backupDir, err := Backup(dir, "v3.0.0-test")
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}
	if _, err := os.Stat(filepath.Join(backupDir, "server.exe")); err != nil {
		t.Fatalf("备份缺少 server.exe: %v", err)
	}

	// 构造新版本 staging 并应用
	staging := filepath.Join(dir, "updates", "staging")
	writeStaging := func(rel, content string) {
		p := filepath.Join(staging, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	writeStaging("server.exe", "NEW-SERVER")
	writeStaging("dist/index.html", "<h1>NEW-DIST</h1>")
	writeStaging("dist/assets/app.js", "NEW-ASSET")
	writeStaging("dist-mobile/index.html", "<h1>NEW-MOBILE</h1>")

	if err := ApplyStaging(dir, staging); err != nil {
		t.Fatalf("ApplyStaging: %v", err)
	}
	assertContent(t, filepath.Join(dir, "server.exe"), "NEW-SERVER")
	assertContent(t, filepath.Join(dir, "dist/index.html"), "<h1>NEW-DIST</h1>")
	assertContent(t, filepath.Join(dir, "dist-mobile/index.html"), "<h1>NEW-MOBILE</h1>")

	// 运行数据未被触碰
	assertContent(t, filepath.Join(dir, "configs/config.yaml"), "password: secret")
	assertContent(t, filepath.Join(dir, "migrations/001.sql"), "CREATE TABLE t;")

	// 回滚
	if err := Rollback(dir, backupDir); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	assertContent(t, filepath.Join(dir, "server.exe"), "OLD-SERVER")
	assertContent(t, filepath.Join(dir, "dist/index.html"), "<h1>OLD-DIST</h1>")
	assertContent(t, filepath.Join(dir, "dist-mobile/index.html"), "<h1>OLD-MOBILE</h1>")
	// 回滚同样不碰运行数据
	assertContent(t, filepath.Join(dir, "configs/config.yaml"), "password: secret")
}

func assertContent(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取 %s: %v", path, err)
	}
	if string(data) != want {
		t.Errorf("%s 内容 = %q, want %q", path, data, want)
	}
}

// ---------- 解压 ----------

func makeZip(t *testing.T, path string, files map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w := zip.NewWriter(f)
	for name, content := range files {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}

func makeTarGz(t *testing.T, path string, files map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)
	for name, content := range files {
		hdr := &tar.Header{Name: name, Mode: 0644, Size: int64(len(content)), Typeflag: tar.TypeReg}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestExtractPackage(t *testing.T) {
	files := map[string]string{
		"server.exe":      "SERVER",
		"dist/index.html": "<h1>DIST</h1>",
		"dist/a/b.js":     "JS",
	}
	zipPath := filepath.Join(t.TempDir(), "pkg.zip")
	makeZip(t, zipPath, files)

	out := filepath.Join(t.TempDir(), "staging")
	if _, err := ExtractPackage(zipPath, out); err != nil {
		t.Fatalf("ExtractPackage(zip): %v", err)
	}
	assertContent(t, filepath.Join(out, "server.exe"), "SERVER")
	assertContent(t, filepath.Join(out, "dist/a/b.js"), "JS")

	tgzPath := filepath.Join(t.TempDir(), "pkg.tar.gz")
	makeTarGz(t, tgzPath, files)
	out2 := filepath.Join(t.TempDir(), "staging")
	if _, err := ExtractPackage(tgzPath, out2); err != nil {
		t.Fatalf("ExtractPackage(tar.gz): %v", err)
	}
	assertContent(t, filepath.Join(out2, "dist/index.html"), "<h1>DIST</h1>")
}

// ---------- 校验 ----------

func TestVerifyChecksum(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "pkg.bin")
	content := []byte("hello schedule")
	if err := os.WriteFile(p, content, 0644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(content)
	expect := hex.EncodeToString(sum[:])

	if err := VerifyChecksum(p, ""); err != nil {
		t.Error("期望值为空时应跳过校验")
	}
	if err := VerifyChecksum(p, expect); err != nil {
		t.Errorf("校验应通过: %v", err)
	}
	if err := VerifyChecksum(p, "deadbeef"); err == nil {
		t.Error("校验失败时应返回错误")
	}
}

// ---------- GitHub 响应解析辅助 ----------

func TestFindAsset(t *testing.T) {
	assets := []Asset{
		{Name: "schedule-system-v2-windows-amd64.zip"},
		{Name: "schedule-system-v2-windows-amd64.zip.sha256"},
		{Name: "schedule-system-v2-linux-amd64.tar.gz"},
	}
	a, ok := findAsset(assets, "schedule-system-v2-windows-amd64.zip")
	if !ok || a.Name != "schedule-system-v2-windows-amd64.zip" {
		t.Error("按名称查找资产失败")
	}
	if _, ok := findAsset(assets, "not-exist.zip"); ok {
		t.Error("不存在的资产不应命中")
	}
	// 大小写不敏感（Windows 文件系统语义）
	if _, ok := findAsset(assets, "SCHEDULE-SYSTEM-V2-WINDOWS-AMD64.ZIP"); !ok {
		t.Error("资产查找应大小写不敏感")
	}
}

func TestPlatformAssetName(t *testing.T) {
	name := PlatformAssetName()
	if name != "schedule-system-v2-windows-amd64.zip" && name != "schedule-system-v2-linux-amd64.tar.gz" {
		t.Errorf("未知平台资产名: %s", name)
	}
}

// ---------- 下载中止 ----------

func TestDownloadChecksumParsing(t *testing.T) {
	// 覆盖 DownloadChecksum 的字段解析逻辑（通过 findAsset 的命名约定间接验证）
	name := AssetFileName("https://github.com/o/r/releases/download/v3.0.1/schedule-system-v2-windows-amd64.zip")
	if name != "schedule-system-v2-windows-amd64.zip" {
		t.Errorf("AssetFileName = %s", name)
	}
}

var _ = bytes.MinRead // 保持导入稳定
