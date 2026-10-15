package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"schedule-system-v2/backend/internal/update"

	"github.com/gin-gonic/gin"
)

// ---------- 测试辅助 ----------

// callHandler 用 gin 测试上下文调用处理器，返回 HTTP 状态码与响应信封
func callHandler(t *testing.T, method, path string, h gin.HandlerFunc) (int, map[string]interface{}) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, nil)

	h(c)

	var env struct {
		Code    int                    `json:"code"`
		Message string                 `json:"message"`
		Data    map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("响应解析失败: %v (body=%s)", err, w.Body.String())
	}
	if env.Data == nil {
		env.Data = map[string]interface{}{}
	}
	// JSON 数字反序列化为 float64，统一按 float64 存放便于断言
	env.Data["_code"] = float64(env.Code)
	env.Data["_message"] = env.Message
	return w.Code, env.Data
}

// withInstallDir 将 installDirFn 指向临时目录，测试结束后还原
func withInstallDir(t *testing.T, dir string) {
	t.Helper()
	orig := installDirFn
	installDirFn = func() string { return dir }
	t.Cleanup(func() { installDirFn = orig })
}

// resetCheckCache 清空检查结果缓存（包级变量，同包测试可直接操作）
func resetCheckCache() {
	checkCacheMu.Lock()
	defer checkCacheMu.Unlock()
	checkCacheInfo, checkCacheErr, checkCacheAt = nil, nil, time.Time{}
}

// mockGitHub 启动一个模拟 GitHub API 服务器，返回 releases JSON。
// 测试期间覆盖 update 包的 GitHub API 地址，结束后自动还原。
func mockGitHub(t *testing.T, releases string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(releases))
	}))
	restore := update.SetGitHubAPIBaseForTest(srv.URL)
	t.Cleanup(func() {
		restore()
		srv.Close()
	})
	return srv
}

// releaseJSON 构造 GitHub releases 响应
func releaseJSON(t *testing.T, tags ...string) string {
	t.Helper()
	type asset struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	}
	type rel struct {
		TagName    string  `json:"tag_name"`
		Prerelease bool    `json:"prerelease"`
		Draft      bool    `json:"draft"`
		Body       string  `json:"body"`
		Assets     []asset `json:"assets"`
	}
	var list []rel
	for _, tag := range tags {
		list = append(list, rel{
			TagName: tag,
			Body:    "release " + tag,
			Assets: []asset{
				{Name: "schedule-system-v2-windows-amd64.zip", BrowserDownloadURL: "https://example.invalid/" + tag + ".zip", Size: 1024},
				{Name: "schedule-system-v2-linux-amd64.tar.gz", BrowserDownloadURL: "https://example.invalid/" + tag + ".tar.gz", Size: 2048},
			},
		})
	}
	data, err := json.Marshal(list)
	if err != nil {
		t.Fatalf("构造 releases JSON 失败: %v", err)
	}
	return string(data)
}

// ---------- GetVersion ----------

func TestGetVersion(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "version.txt"), []byte("v3.1.0\n"), 0644); err != nil {
		t.Fatal(err)
	}
	withInstallDir(t, dir)

	_, data := callHandler(t, http.MethodGet, "/system/version", NewUpdateHandler().GetVersion)

	if got := data["version"]; got != "v3.1.0" {
		t.Errorf("version = %v, want v3.1.0", got)
	}
	if got, _ := data["platform"].(string); got == "" || !contains(got, "/") {
		t.Errorf("platform = %v, 期望形如 windows/amd64", got)
	}
	if got := data["updaterPresent"]; got != false {
		t.Errorf("updaterPresent = %v, want false（临时目录无 update.exe）", got)
	}
}

func TestGetVersionNoVersionFile(t *testing.T) {
	withInstallDir(t, t.TempDir()) // 无 version.txt

	_, data := callHandler(t, http.MethodGet, "/system/version", NewUpdateHandler().GetVersion)

	if got := data["version"]; got != "v0.0.0" {
		t.Errorf("无 version.txt 时 version = %v, want v0.0.0", got)
	}
}

// ---------- CheckUpdate ----------

func TestCheckUpdateSameMajor(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "version.txt"), []byte("v3.0.0"), 0644); err != nil {
		t.Fatal(err)
	}
	withInstallDir(t, dir)
	mockGitHub(t, releaseJSON(t, "v3.0.1", "v3.0.0", "v2.9.0"))
	resetCheckCache()

	_, data := callHandler(t, http.MethodGet, "/admin/update/check", NewUpdateHandler().CheckUpdate)

	if data["hasUpdate"] != true {
		t.Fatalf("hasUpdate = %v, want true", data["hasUpdate"])
	}
	if data["sameMajor"] != true {
		t.Errorf("sameMajor = %v, want true", data["sameMajor"])
	}
	if data["current"] != "v3.0.0" || data["target"] != "v3.0.1" {
		t.Errorf("current/target = %v/%v, want v3.0.0/v3.0.1", data["current"], data["target"])
	}
	// 体积按平台取值（mock: windows zip=1024 / linux tar.gz=2048），
	// 否则在 Linux CI 上会因选中 tar.gz 而断言失败
	wantSize := float64(1024)
	if update.PlatformAssetName() != "schedule-system-v2-windows-amd64.zip" {
		wantSize = 2048
	}
	if data["assetSize"] != wantSize {
		t.Errorf("assetSize = %v, want %v", data["assetSize"], wantSize)
	}
}

func TestCheckUpdateCrossMajor(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "version.txt"), []byte("v2.2.0"), 0644); err != nil {
		t.Fatal(err)
	}
	withInstallDir(t, dir)
	mockGitHub(t, releaseJSON(t, "v3.0.1", "v2.2.0"))
	resetCheckCache()

	_, data := callHandler(t, http.MethodGet, "/admin/update/check", NewUpdateHandler().CheckUpdate)

	if data["hasUpdate"] != true || data["sameMajor"] != false {
		t.Errorf("跨大版本: hasUpdate=%v sameMajor=%v, want true/false", data["hasUpdate"], data["sameMajor"])
	}
	if data["target"] != "v3.0.1" {
		t.Errorf("target = %v, want v3.0.1", data["target"])
	}
}

func TestCheckUpdateAlreadyLatest(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "version.txt"), []byte("v3.0.1"), 0644); err != nil {
		t.Fatal(err)
	}
	withInstallDir(t, dir)
	mockGitHub(t, releaseJSON(t, "v3.0.1", "v3.0.0"))
	resetCheckCache()

	_, data := callHandler(t, http.MethodGet, "/admin/update/check", NewUpdateHandler().CheckUpdate)

	if data["hasUpdate"] != false {
		t.Errorf("hasUpdate = %v, want false（已是最新）", data["hasUpdate"])
	}
}

func TestCheckUpdateGitHubError(t *testing.T) {
	dir := t.TempDir()
	withInstallDir(t, dir)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"boom"}`))
	}))
	restore := update.SetGitHubAPIBaseForTest(srv.URL)
	t.Cleanup(func() { restore(); srv.Close() })
	resetCheckCache()

	code, data := callHandler(t, http.MethodGet, "/admin/update/check", NewUpdateHandler().CheckUpdate)

	if code != http.StatusOK || data["_code"] != float64(500) {
		t.Errorf("GitHub 故障时应返回业务码 500，得到 http=%d code=%v", code, data["_code"])
	}
	if msg, _ := data["_message"].(string); !contains(msg, "检查更新失败") {
		t.Errorf("message = %q, 期望包含「检查更新失败」", msg)
	}
}

func TestCheckUpdateCache(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "version.txt"), []byte("v3.0.0"), 0644); err != nil {
		t.Fatal(err)
	}
	withInstallDir(t, dir)

	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(releaseJSON(t, "v3.0.1")))
	}))
	restore := update.SetGitHubAPIBaseForTest(srv.URL)
	t.Cleanup(func() { restore(); srv.Close() })
	resetCheckCache()

	h := NewUpdateHandler()
	callHandler(t, http.MethodGet, "/admin/update/check", h.CheckUpdate)
	callHandler(t, http.MethodGet, "/admin/update/check", h.CheckUpdate)

	if hits != 1 {
		t.Errorf("5 分钟缓存内应只请求 GitHub 1 次，实际 %d 次", hits)
	}
}

// ---------- StartUpdate ----------

func TestStartUpdateNoUpdaterBinary(t *testing.T) {
	withInstallDir(t, t.TempDir()) // 目录中没有 update.exe
	resetCheckCache()

	code, data := callHandler(t, http.MethodPost, "/admin/update/start", NewUpdateHandler().StartUpdate)

	if data["_code"] != float64(400) {
		t.Fatalf("code = %v, want 400 (http=%d)", data["_code"], code)
	}
	if msg, _ := data["_message"].(string); !contains(msg, "未找到更新程序") {
		t.Errorf("message = %q, 期望提示缺少 update.exe", msg)
	}
}

func TestStartUpdateAlreadyLatest(t *testing.T) {
	dir := t.TempDir()
	// 放置一个假的更新程序，绕过「文件存在」检查，让流程走到版本判断
	if err := os.WriteFile(filepath.Join(dir, update.UpdaterExeName()), []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "version.txt"), []byte("v3.0.1"), 0644); err != nil {
		t.Fatal(err)
	}
	withInstallDir(t, dir)
	mockGitHub(t, releaseJSON(t, "v3.0.1"))
	resetCheckCache()

	_, data := callHandler(t, http.MethodPost, "/admin/update/start", NewUpdateHandler().StartUpdate)

	if data["_code"] != float64(400) {
		t.Fatalf("code = %v, want 400", data["_code"])
	}
	if msg, _ := data["_message"].(string); !contains(msg, "已是最新版本") {
		t.Errorf("message = %q, 期望「当前已是最新版本」", msg)
	}
}

func TestStartUpdateCrossMajorBlocked(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, update.UpdaterExeName()), []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "version.txt"), []byte("v2.2.0"), 0644); err != nil {
		t.Fatal(err)
	}
	withInstallDir(t, dir)
	mockGitHub(t, releaseJSON(t, "v3.0.1"))
	resetCheckCache()

	_, data := callHandler(t, http.MethodPost, "/admin/update/start", NewUpdateHandler().StartUpdate)

	if data["_code"] != float64(400) {
		t.Fatalf("code = %v, want 400", data["_code"])
	}
	if msg, _ := data["_message"].(string); !contains(msg, "跨大版本") {
		t.Errorf("message = %q, 期望提示跨大版本需手动升级", msg)
	}
}

// ---------- GetUpdateStatus / UpdateLogTail ----------

func TestGetUpdateStatus(t *testing.T) {
	t.Run("无结果文件", func(t *testing.T) {
		withInstallDir(t, t.TempDir())
		_, data := callHandler(t, http.MethodGet, "/admin/update/status", NewUpdateHandler().GetUpdateStatus)
		if data["hasResult"] != false {
			t.Errorf("hasResult = %v, want false", data["hasResult"])
		}
	})

	t.Run("有结果文件", func(t *testing.T) {
		dir := t.TempDir()
		result := `{"phase":"done","message":"更新成功，已升级到 v3.0.1","from":"v3.0.0","to":"v3.0.1","at":"2026-10-15 09:29:21"}`
		if err := os.MkdirAll(filepath.Join(dir, "updates"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "updates", "last-result.json"), []byte(result), 0644); err != nil {
			t.Fatal(err)
		}
		withInstallDir(t, dir)

		_, data := callHandler(t, http.MethodGet, "/admin/update/status", NewUpdateHandler().GetUpdateStatus)
		if data["hasResult"] != true {
			t.Fatalf("hasResult = %v, want true", data["hasResult"])
		}
		if data["phase"] != "done" || data["to"] != "v3.0.1" {
			t.Errorf("解析结果 = %v/%v, want done/v3.0.1", data["phase"], data["to"])
		}
	})
}

func TestUpdateLogTail(t *testing.T) {
	t.Run("无日志文件", func(t *testing.T) {
		withInstallDir(t, t.TempDir())
		_, data := callHandler(t, http.MethodGet, "/admin/update/log", NewUpdateHandler().UpdateLogTail)
		if log, _ := data["log"].(string); log != "" {
			t.Errorf("log = %q, want 空", log)
		}
	})

	t.Run("超长日志截断到 8KB", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, "updates"), 0755); err != nil {
			t.Fatal(err)
		}
		big := make([]byte, 20*1024)
		for i := range big {
			big[i] = 'x'
		}
		if err := os.WriteFile(filepath.Join(dir, "updates", "update.log"), big, 0644); err != nil {
			t.Fatal(err)
		}
		withInstallDir(t, dir)

		_, data := callHandler(t, http.MethodGet, "/admin/update/log", NewUpdateHandler().UpdateLogTail)
		log, _ := data["log"].(string)
		if len(log) != 8*1024 {
			t.Errorf("日志长度 = %d, want 8192（截断到末尾 8KB）", len(log))
		}
	})
}

// ---------- HealthCheck ----------

func TestHealthCheck(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/system/health", nil)

	NewUpdateHandler().HealthCheck(c)

	if w.Code != http.StatusOK {
		t.Errorf("状态码 = %d, want 200", w.Code)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
