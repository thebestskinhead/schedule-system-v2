package update

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------- 测试辅助 ----------

// mockReleases 启动模拟 GitHub API 服务器，返回指定 tag 的 releases 列表。
// 测试结束后自动还原 API 地址。
func mockReleases(t *testing.T, tags ...string) {
	t.Helper()
	type asset struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	}
	type rel struct {
		TagName string  `json:"tag_name"`
		Assets  []asset `json:"assets"`
	}
	list := make([]rel, 0, len(tags))
	for _, tag := range tags {
		list = append(list, rel{
			TagName: tag,
			Assets: []asset{
				{Name: "schedule-system-v2-windows-amd64.zip", BrowserDownloadURL: "https://example.invalid/" + tag + ".zip", Size: 123},
				{Name: "schedule-system-v2-linux-amd64.tar.gz", BrowserDownloadURL: "https://example.invalid/" + tag + ".tar.gz", Size: 456},
			},
		})
	}
	body, err := json.Marshal(list)
	if err != nil {
		t.Fatalf("构造 releases JSON 失败: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	}))
	restore := SetGitHubAPIBaseForTest(srv.URL)
	t.Cleanup(func() {
		restore()
		srv.Close()
	})
}

// withVersionFile 在临时目录生成 version.txt 并返回目录
func withVersionFile(t *testing.T, version string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, VersionFileName), []byte(version), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// ---------- 检查更新结果语义 ----------

// TestCheckUpdateNoUpdateTarget 无更新时 Target 必须回落到当前版本。
// 回归：曾返回结构体零值，前端「最新版本」显示为 v0.0.0，与「已是最新」标签自相矛盾。
func TestCheckUpdateNoUpdateTarget(t *testing.T) {
	dir := withVersionFile(t, "v3.1.0")
	mockReleases(t, "v3.1.0", "v3.0.1")

	info, err := CheckUpdate(dir, DefaultRepo, PlatformAssetName(), false)
	if err != nil {
		t.Fatalf("CheckUpdate: %v", err)
	}
	if info.HasUpdate {
		t.Fatalf("hasUpdate = true, want false")
	}
	if info.Target.String() != "v3.1.0" {
		t.Errorf("无更新时 target = %s, want v3.1.0（应为当前版本而非零值 v0.0.0）", info.Target)
	}
	if !info.SameMajor {
		t.Error("无更新时 sameMajor 应为 true（当前即最新同大版本）")
	}
	if info.CrossMajor {
		t.Error("无更新时 crossMajor 应为 false")
	}
}

// TestCheckUpdateNoUpdateDevVersion 领先于所有 release 的 dev 版本同样不误报更新
func TestCheckUpdateNoUpdateDevVersion(t *testing.T) {
	dir := withVersionFile(t, "v3.1.0-dev+e40455c")
	mockReleases(t, "v3.0.1", "v3.0.0")

	info, err := CheckUpdate(dir, DefaultRepo, PlatformAssetName(), false)
	if err != nil {
		t.Fatalf("CheckUpdate: %v", err)
	}
	if info.HasUpdate {
		t.Errorf("本地构建领先于最新 release 时不应提示更新，target=%s", info.Target)
	}
	if info.Target.String() != "v3.1.0-dev+e40455c" {
		t.Errorf("target = %s, want 当前版本", info.Target)
	}
}

// TestCheckUpdateCrossMajorFlag 找不到可用的同大版本更新、但存在更高大版本时：
// 标记 crossMajor 并给出目标版本，供前端提示手动下载整包（本地无可用热更新链路）。
func TestCheckUpdateCrossMajorFlag(t *testing.T) {
	dir := withVersionFile(t, "v3.1.0")
	// v4.0.0 未发布本平台资产 → 无法热更新；v3.1.0 与当前相同
	body := `[
		{"tag_name":"v4.0.0","assets":[]},
		{"tag_name":"v3.1.0","assets":[{"name":"schedule-system-v2-windows-amd64.zip","browser_download_url":"https://example.invalid/w.zip","size":9},{"name":"schedule-system-v2-linux-amd64.tar.gz","browser_download_url":"https://example.invalid/l.tgz","size":9}]}
	]`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	restore := SetGitHubAPIBaseForTest(srv.URL)
	t.Cleanup(func() { restore(); srv.Close() })

	info, err := CheckUpdate(dir, DefaultRepo, PlatformAssetName(), false)
	if err != nil {
		t.Fatalf("CheckUpdate: %v", err)
	}
	if info.HasUpdate {
		t.Error("跨大版本不应判定为可热更新")
	}
	if !info.CrossMajor {
		t.Error("存在更高大版本时应标记 crossMajor = true")
	}
	if info.SameMajor {
		t.Error("跨大版本时 sameMajor 应为 false")
	}
	if info.Target.String() != "v4.0.0" {
		t.Errorf("crossMajor 时 target = %s, want v4.0.0", info.Target)
	}
}

// TestCheckUpdateHigherMajorWithAsset 更高大版本且带本平台资产时，
// 走 hasUpdate=true / sameMajor=false 分支（前端据此同样提示跨大版本）。
func TestCheckUpdateHigherMajorWithAsset(t *testing.T) {
	dir := withVersionFile(t, "v3.1.0")
	mockReleases(t, "v4.0.0", "v3.1.0")

	info, err := CheckUpdate(dir, DefaultRepo, PlatformAssetName(), false)
	if err != nil {
		t.Fatalf("CheckUpdate: %v", err)
	}
	if !info.HasUpdate {
		t.Fatal("更高大版本带资产时应判定为存在更新")
	}
	if info.SameMajor {
		t.Error("与更高大版本之间 sameMajor 应为 false（前端提示跨大版本）")
	}
	if info.Target.String() != "v4.0.0" {
		t.Errorf("target = %s, want v4.0.0", info.Target)
	}
}

// TestCheckUpdateReleaseWithoutPlatformAsset 缺少本平台资产的 release 应被跳过，
// 继续向更早的 release 查找可用版本。
func TestCheckUpdateReleaseWithoutPlatformAsset(t *testing.T) {
	dir := withVersionFile(t, "v3.0.0")

	// 只有无资产的 v3.1.0 与带资产的 v3.0.1
	body := `[
		{"tag_name":"v3.1.0","assets":[]},
		{"tag_name":"v3.0.1","assets":[{"name":"schedule-system-v2-windows-amd64.zip","browser_download_url":"https://example.invalid/w.zip","size":9},{"name":"schedule-system-v2-linux-amd64.tar.gz","browser_download_url":"https://example.invalid/l.tgz","size":9}]}
	]`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	restore := SetGitHubAPIBaseForTest(srv.URL)
	t.Cleanup(func() { restore(); srv.Close() })

	info, err := CheckUpdate(dir, DefaultRepo, PlatformAssetName(), false)
	if err != nil {
		t.Fatalf("CheckUpdate: %v", err)
	}
	if !info.HasUpdate || info.Target.String() != "v3.0.1" {
		t.Errorf("应跳过无资产的 v3.1.0 并选中 v3.0.1，实际 hasUpdate=%v target=%s", info.HasUpdate, info.Target)
	}
}

// TestCheckUpdateSkipsPrereleaseAndDraft 预发布与草稿版本默认不参与更新判断
func TestCheckUpdateSkipsPrereleaseAndDraft(t *testing.T) {
	dir := withVersionFile(t, "v3.0.0")
	body := `[
		{"tag_name":"v3.2.0","prerelease":true,"assets":[{"name":"schedule-system-v2-windows-amd64.zip","browser_download_url":"https://example.invalid/a.zip","size":1},{"name":"schedule-system-v2-linux-amd64.tar.gz","browser_download_url":"https://example.invalid/a.tgz","size":1}]},
		{"tag_name":"v3.1.0","draft":true,"assets":[{"name":"schedule-system-v2-windows-amd64.zip","browser_download_url":"https://example.invalid/b.zip","size":1},{"name":"schedule-system-v2-linux-amd64.tar.gz","browser_download_url":"https://example.invalid/b.tgz","size":1}]},
		{"tag_name":"v3.0.1","assets":[{"name":"schedule-system-v2-windows-amd64.zip","browser_download_url":"https://example.invalid/c.zip","size":1},{"name":"schedule-system-v2-linux-amd64.tar.gz","browser_download_url":"https://example.invalid/c.tgz","size":1}]}
	]`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	restore := SetGitHubAPIBaseForTest(srv.URL)
	t.Cleanup(func() { restore(); srv.Close() })

	info, err := CheckUpdate(dir, DefaultRepo, PlatformAssetName(), false)
	if err != nil {
		t.Fatalf("CheckUpdate: %v", err)
	}
	if info.Target.String() != "v3.0.1" {
		t.Errorf("应跳过预发布 v3.2.0 与草稿 v3.1.0，选中 v3.0.1，实际 %s", info.Target)
	}
}

// ---------- 网络错误友好化 ----------

type timeoutErr struct{}

func (timeoutErr) Error() string   { return "i/o timeout" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return true }

func TestFriendlyNetErr(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		contains []string
	}{
		{
			name:     "DNS 解析失败（用户实际遇到的 no such host）",
			err:      &net.DNSError{Err: "no such host", Name: "api.github.com", IsNotFound: true},
			contains: []string{"DNS 解析失败", "no such host", "HTTPS_PROXY"},
		},
		{
			name:     "连接超时",
			err:      timeoutErr{},
			contains: []string{"超时", "HTTPS_PROXY"},
		},
		{
			name:     "context 超时",
			err:      context.DeadlineExceeded,
			contains: []string{"超时", "HTTPS_PROXY"},
		},
		{
			name:     "其它网络错误",
			err:      errors.New("connection reset by peer"),
			contains: []string{"无法连接 GitHub", "connection reset by peer", "HTTPS_PROXY"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := friendlyNetErr(c.err)
			if got == nil {
				t.Fatal("不应返回 nil")
			}
			msg := got.Error()
			for _, want := range c.contains {
				if !strings.Contains(msg, want) {
					t.Errorf("错误信息 %q 未包含 %q", msg, want)
				}
			}
			// 原始英文错误不再原样抛出（已转为可操作提示）
			if strings.HasPrefix(msg, "Get \"https://api.github.com") {
				t.Errorf("错误信息仍为原始 Go 网络错误: %q", msg)
			}
		})
	}

	if got := friendlyNetErr(nil); got != nil {
		t.Errorf("nil 应返回 nil，实际 %v", got)
	}
}
