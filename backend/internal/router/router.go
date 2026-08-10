package router

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"schedule-system-v2/backend/internal/config"
	"schedule-system-v2/backend/internal/dao"
	"schedule-system-v2/backend/internal/handler"
	"schedule-system-v2/backend/internal/middleware"
	"schedule-system-v2/backend/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
)

// getDistPath 获取电脑端前端路径
func getDistPath() string {
	paths := []string{
		"./dist",
		"./frontend/dist",
		"./static",
		"dist",
		"frontend/dist",
		"static",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "./dist"
}

// getMobileDistPath 获取移动端前端路径
func getMobileDistPath() string {
	paths := []string{
		"./dist-mobile",
		"./frontend-mobile/dist",
		"dist-mobile",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "./dist-mobile"
}

// isMobileUA 判断 User-Agent 是否为移动端设备
func isMobileUA(ua string) bool {
	lower := strings.ToLower(ua)
	keywords := []string{
		"mobile", "android", "iphone", "ipod", "ipad",
		"windows phone", "iemobile", "opera mini",
		"blackberry", "bb10", "webos", "palm", "midp", "symbian", "nokia",
	}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// CORS中间件
	r.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		if len(path) >= 4 && path[:4] == "/api" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}
		}
		c.Next()
	})

	// 初始化服务和 Handler
	handlers := &Handlers{
		System:       handler.NewSystemHandler(),
		User:         handler.NewUserHandler(),
		Availability: handler.NewAvailabilityHandler(),
		Schedule:     handler.NewScheduleHandler(),
		Template:     handler.NewTemplateHandler(),
		WeeklyDuty:   handler.NewWeeklyDutyHandler(),
		TempPerm:     handler.NewTempPermissionHandler(),
		SMTP:         handler.NewSMTPHandler(),
		QRLogin:      handler.NewQRLoginHandler(),
	}

	// 初始化应用系统相关服务
	if config.IsInstalled() {
		userDao := dao.NewUserDAO()
		tempPermissionDao := dao.NewTempPermissionDAO()
		applicationDao := dao.NewApplicationDao()

		appManager := service.NewApplicationManager()
		applicationService := service.NewApplicationService(applicationDao, userDao, appManager)

		// 注册临权申请执行器
		tempPermExecutor := service.NewTempPermissionExecutor(userDao, tempPermissionDao, applicationDao)
		appManager.Register(tempPermExecutor)

		handlers.Application = handler.NewApplicationHandler(applicationService)
	}

	// 表驱动注册全部 API 路由（接口 → 权限 单一事实来源）
	api := r.Group("/api/v1")
	for _, rt := range buildRoutes(handlers) {
		chain := make([]gin.HandlerFunc, 0, 3)
		if rt.Auth {
			chain = append(chain, middleware.AuthMiddleware())
			if len(rt.Perms) == 1 {
				chain = append(chain, middleware.PermissionMiddleware(rt.Perms[0]))
			} else if len(rt.Perms) > 1 {
				chain = append(chain, middleware.PermissionAnyMiddleware(rt.Perms...))
			}
		}
		chain = append(chain, rt.Handler)
		api.Handle(rt.Method, rt.Path, chain...)
	}

	// 阻止已安装系统访问 init 页面
	r.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/init" || path == "/init/" {
			if config.IsInstalled() {
				c.JSON(http.StatusNotFound, gin.H{
					"code":    404,
					"message": "安装页面已禁用，系统已安装",
				})
				c.Abort()
				return
			}
		}
		c.Next()
	})

	// 移动端 UA 自动跳转到移动端页面（/mobile/...）
	r.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		// 仅处理电脑端页面请求：排除 API、移动端路径、静态资源文件
		if strings.HasPrefix(path, "/api") ||
			path == "/mobile" || strings.HasPrefix(path, "/mobile/") ||
			strings.HasPrefix(path, "/assets/") ||
			filepath.Ext(path) != "" ||
			!isMobileUA(c.GetHeader("User-Agent")) {
			c.Next()
			return
		}
		target := "/mobile"
		if path != "/" {
			target = "/mobile" + path
		}
		if q := c.Request.URL.RawQuery; q != "" {
			target += "?" + q
		}
		c.Redirect(http.StatusFound, target)
		c.Abort()
	})

	// 开发模式：go run 直接运行（或 SCHEDULE_DEV=1）且存在 dev 配置时，
	// 将页面请求反向代理到前端开发服务器（vite），/api 仍由后端直接处理。
	devCfg, devErr := config.LoadDevConfig(config.DevConfigFilePath)
	if devErr != nil {
		log.Printf("加载开发配置 %s 失败: %v", config.DevConfigFilePath, devErr)
	}
	if config.IsDevRun() && devCfg != nil {
		setupDevProxy(r, devCfg)
		return r
	}

	// 静态文件服务（生产模式：直接托管构建产物）
	distPath := getDistPath()
	mobileDistPath := getMobileDistPath()
	assetsPath := filepath.Join(distPath, "assets")
	indexPath := filepath.Join(distPath, "index.html")
	mobileIndexPath := filepath.Join(mobileDistPath, "index.html")

	// 移动端前端路由 - /mobile 及其子路径
	r.GET("/mobile", func(c *gin.Context) {
		if _, err := os.Stat(mobileIndexPath); err == nil {
			c.File(mobileIndexPath)
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "移动端前端未构建",
			})
		}
	})
	r.GET("/mobile/*action", func(c *gin.Context) {
		// 检查是否是静态资源请求
		action := c.Param("action")
		if strings.HasPrefix(action, "/assets/") {
			// 静态资源文件
			filePath := filepath.Join(mobileDistPath, action[1:])
			if _, err := os.Stat(filePath); err == nil {
				c.File(filePath)
				return
			}
		}
		// 其他路径返回 index.html (SPA路由)
		if _, err := os.Stat(mobileIndexPath); err == nil {
			c.File(mobileIndexPath)
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "移动端前端未构建",
			})
		}
	})

	// 电脑端静态资源
	r.Static("/assets", assetsPath)
	r.StaticFile("/", indexPath)

	// NoRoute 兜底：/api 未注册路径默认拒绝 403；其余回退到 SPA index.html
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "接口不存在或无权限访问",
			})
			return
		}
		if _, err := os.Stat(indexPath); err == nil {
			c.File(indexPath)
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "排班系统后端运行中，前端未构建",
			})
		}
	})

	return r
}

// setupDevProxy 开发模式路由：把页面请求代理到 vite 开发服务器。
// 移动端 vite 的 base 为 /mobile/，路径原样转发即可；其余非 /api 请求走电脑端。
func setupDevProxy(r *gin.Engine, cfg *config.DevConfig) {
	mobileProxy := newDevReverseProxy(cfg.Mobile)
	frontendProxy := newDevReverseProxy(cfg.Frontend)

	log.Printf("开发模式（代理前端开发服务器）: 电脑端 -> %s, 移动端 -> %s", cfg.Frontend, cfg.Mobile)

	// 移动端：/mobile 及其子路径（含 /mobile/assets、HMR 相关请求）
	r.Any("/mobile", mobileProxy)
	r.Any("/mobile/*action", mobileProxy)

	// 其余非 API 请求（页面、/@vite/client、/src、/node_modules 等）代理到电脑端
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api") {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "接口不存在或无权限访问"})
			return
		}
		frontendProxy(c)
	})
}

// newDevReverseProxy 创建到前端开发服务器的反向代理处理器。
// httputil.ReverseProxy 支持 WebSocket 升级，vite HMR 可正常工作。
func newDevReverseProxy(target string) gin.HandlerFunc {
	u, err := url.Parse(target)
	if err != nil {
		log.Fatalf("开发模式前端地址无效 %q: %v", target, err)
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, e error) {
		log.Printf("代理到 %s 失败: %v", target, e)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("前端开发服务器未启动或无法连接: " + target + "\n提示：开发模式仅代理页面请求，需先启动前端 vite（对应目录下 npm run dev），或在项目根目录用 .\\dev.ps1 一键启动前后端。"))
	}
	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
