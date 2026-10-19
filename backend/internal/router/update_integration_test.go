package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// ==================== 更新相关接口：走真实路由的集成测试 ====================

// newTestServer 用真实 SetupRouter 启动一个测试服务器（不依赖数据库：
// 更新相关接口均不访问 DB，管理端接口在无 token 时即被 401 拦截）。
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	gin.SetMode(gin.TestMode)
	srv := httptest.NewServer(SetupRouter())
	t.Cleanup(srv.Close)
	return srv
}

// getJSON 发起 GET 并解析响应信封
func getJSON(t *testing.T, url string) (int, map[string]interface{}) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s 失败: %v", url, err)
	}
	defer resp.Body.Close()
	var env struct {
		Code    int                    `json:"code"`
		Message string                 `json:"message"`
		Data    map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		t.Fatalf("GET %s 响应解析失败: %v", url, err)
	}
	return resp.StatusCode, map[string]interface{}{
		"httpCode": resp.StatusCode,
		// JSON 数字反序列化为 float64，统一按 float64 存放便于断言
		"code":    float64(env.Code),
		"message": env.Message,
		"data":    env.Data,
	}
}

func TestUpdatePublicEndpointsThroughRouter(t *testing.T) {
	srv := newTestServer(t)

	cases := []struct {
		path        string
		wantDataKey string
	}{
		{"/api/v1/system/version", "version"},
		{"/api/v1/system/health", ""},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			httpCode, env := getJSON(t, srv.URL+tc.path)
			if httpCode != http.StatusOK {
				t.Fatalf("http 状态码 = %d, want 200", httpCode)
			}
			if env["code"] != float64(200) {
				t.Fatalf("业务码 = %v, want 200 (msg=%v)", env["code"], env["message"])
			}
			if tc.wantDataKey != "" {
				data, _ := env["data"].(map[string]interface{})
				if _, ok := data[tc.wantDataKey]; !ok {
					t.Errorf("响应缺少字段 %q: %v", tc.wantDataKey, env["data"])
				}
			}
		})
	}
}

func TestUpdateAdminEndpointsRequireAuth(t *testing.T) {
	srv := newTestServer(t)

	// 未携带 token 的管理端接口必须被认证中间件拦截（401），
	// 且绝不能穿透到业务逻辑（否则会误杀服务/触发下载）。
	for _, path := range []string{
		"/api/v1/admin/update/check",
		"/api/v1/admin/update/progress",
		"/api/v1/admin/update/status",
		"/api/v1/admin/update/log",
		"/api/v1/admin/update/mirrors",
	} {
		t.Run(path, func(t *testing.T) {
			httpCode, _ := getJSON(t, srv.URL+path)
			if httpCode != http.StatusUnauthorized {
				t.Fatalf("http 状态码 = %d, want 401", httpCode)
			}
		})
	}

	// POST 触发下载/安装/取消的接口同样必须拦截
	for _, path := range []string{
		"/api/v1/admin/update/download",
		"/api/v1/admin/update/install",
		"/api/v1/admin/update/cancel-download",
	} {
		resp, err := http.Post(srv.URL+path, "application/json", nil)
		if err != nil {
			t.Fatalf("POST %s 失败: %v", path, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("POST %s 状态码 = %d, want 401", path, resp.StatusCode)
		}
	}

	// PUT 保存镜像同样必须拦截
	req, err := http.NewRequest(http.MethodPut, srv.URL+"/api/v1/admin/update/mirrors", nil)
	if err != nil {
		t.Fatalf("构造请求失败: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT /admin/update/mirrors 失败: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("PUT /admin/update/mirrors 状态码 = %d, want 401", resp.StatusCode)
	}
}

// ==================== 前后端 API 契约测试 ====================
//
// 解析 frontend-v2/src/api/*.js 中的 request.<method>('<path>') 调用，
// 逐一验证后端路由表（SetupRouter 实际注册的路由）中存在对应的
//「方法 + /api/v1 + 路径」。任一端改名/漏注册，此测试立即失败——
// 前后端接口漂移在 CI 阶段即可发现，不会留到运行时 404。

// frontendAPIDir 前端 API 目录（相对本测试包：backend/internal/router）
const frontendAPIDir = "../../../frontend-v2/src/api"

// callRe 匹配 request.get('/x')、request.post(`/x/${id}`) 等调用（含多行调用的首行）
var callRe = regexp.MustCompile("request\\.(get|post|put|delete|patch)\\(\\s*[`'\"]([^`'\"]+)[`'\"]")

// normalizePath 归一化路径中的动态片段，使前后端可以比对：
//   前端 `/admin/users/${id}`   → `/admin/users/:param`
//   后端 `/admin/users/:id`     → `/admin/users/:param`
func normalizePath(p string) string {
	segs := strings.Split(p, "/")
	for i, s := range segs {
		if strings.HasPrefix(s, "${") || strings.HasPrefix(s, ":") {
			segs[i] = ":param"
		}
	}
	return strings.Join(segs, "/")
}

// registeredRoutes 返回后端实际注册的路由集合：key = "GET /api/v1/xxx"
func registeredRoutes(t *testing.T) map[string]bool {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := SetupRouter()
	set := make(map[string]bool)
	for _, ri := range r.Routes() {
		set[strings.ToUpper(ri.Method)+" "+normalizePath(ri.Path)] = true
	}
	return set
}

func TestFrontendBackendAPIContract(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(frontendAPIDir, "*.js"))
	if err != nil || len(files) == 0 {
		t.Fatalf("未找到前端 API 文件 %s: %v (err=%v)", frontendAPIDir, files, err)
	}
	sort.Strings(files)

	routes := registeredRoutes(t)

	type call struct {
		method string
		path   string
		file   string
	}
	var calls []call
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("读取 %s 失败: %v", f, err)
		}
		for _, m := range callRe.FindAllStringSubmatch(string(data), -1) {
			calls = append(calls, call{
				method: strings.ToUpper(m[1]),
				path:   m[2],
				file:   filepath.Base(f),
			})
		}
	}
	if len(calls) < 30 {
		t.Fatalf("解析到的前端调用过少（%d 个），解析规则可能已失效", len(calls))
	}

	var missing []string
	for _, c := range calls {
		key := c.method + " " + normalizePath("/api/v1"+c.path)
		if !routes[key] {
			missing = append(missing, c.file+": "+c.method+" "+c.path+" （期望路由 "+key+"）")
		}
	}
	if len(missing) > 0 {
		t.Fatalf("前端调用了 %d 个后端未注册的接口：\n  - %s",
			len(missing), strings.Join(missing, "\n  - "))
	}
	t.Logf("契约校验通过：前端 %d 个 API 调用全部在后端路由表中登记（来自 %d 个文件）",
		len(calls), len(files))
}

// TestUpdateRoutesRegistered 更新功能接口的方法与路径必须精确登记
func TestUpdateRoutesRegistered(t *testing.T) {
	routes := registeredRoutes(t)

	want := []string{
		"GET /api/v1/system/version",
		"GET /api/v1/system/health",
		"GET /api/v1/admin/update/check",
		"POST /api/v1/admin/update/download",
		"GET /api/v1/admin/update/progress",
		"POST /api/v1/admin/update/cancel-download",
		"POST /api/v1/admin/update/install",
		"GET /api/v1/admin/update/mirrors",
		"PUT /api/v1/admin/update/mirrors",
		"GET /api/v1/admin/update/status",
		"GET /api/v1/admin/update/log",
	}
	for _, key := range want {
		if !routes[key] {
			t.Errorf("路由未注册: %s", key)
		}
	}

	// 旧的一键更新端点（先停服再下载）已被两阶段流程取代，不应残留
	for _, gone := range []string{"POST /api/v1/admin/update/start"} {
		if routes[gone] {
			t.Errorf("路由应已删除: %s", gone)
		}
	}
}

// ==================== 更新程序安装进度页契约 ====================
//
// update.exe 的安装进度页（前端轮询后跳转落地）依赖两个端点：
// GET /api/state、POST /api/retry。
// 这些端点由 cmd/updater 提供（不走 gin），这里校验页面 HTML 中的契约。
// 安装阶段不提供终止入口：/api/abort 与「终止更新」按钮均已移除。

func TestProgressPageEndpointsContract(t *testing.T) {
	html := readUpdaterPageHTML(t)

	for _, ep := range []string{"/api/state", "/api/retry"} {
		if !strings.Contains(html, ep) {
			t.Errorf("安装进度页未引用端点 %s，前端轮询/重试逻辑将失效", ep)
		}
	}
	// 安装阶段不可终止：不得再暴露中止入口
	for _, gone := range []string{"/api/abort", "终止更新"} {
		if strings.Contains(html, gone) {
			t.Errorf("安装进度页不应再包含 %q（安装阶段不提供终止）", gone)
		}
	}
	// 前端自动跳转依赖的 phase 关键字（安装六步 + 终态）
	for _, phase := range []string{"verifying-pkg", "stopping", "backing-up", "installing", "restarting", "verifying", "done", "failed"} {
		if !strings.Contains(html, phase) {
			t.Errorf("安装进度页缺少 phase %q，状态展示将不完整", phase)
		}
	}
}

// readUpdaterPageHTML 读取 cmd/updater 使用的进度页 HTML 常量。
// 通过源码文件读取而不是 import，避免测试包反向依赖 main 包。
func readUpdaterPageHTML(t *testing.T) string {
	t.Helper()
	// 本测试包位于 backend/internal/router，页面常量在 backend/internal/update/page.go
	data, err := os.ReadFile(filepath.Join("..", "..", "internal", "update", "page.go"))
	if err != nil {
		t.Fatalf("读取 page.go 失败: %v", err)
	}
	src := string(data)
	start := strings.Index(src, "ProgressPageHTML = `")
	if start < 0 {
		t.Fatal("未找到 ProgressPageHTML 常量")
	}
	start += len("ProgressPageHTML = `")
	end := strings.Index(src[start:], "`")
	if end < 0 {
		t.Fatal("ProgressPageHTML 常量未正确闭合")
	}
	return src[start : start+end]
}
