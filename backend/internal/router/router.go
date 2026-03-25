package router

import (
	"net/http"
	"os"
	"path/filepath"
	"schedule-system-v2/backend/internal/auth"
	"schedule-system-v2/backend/internal/config"
	"schedule-system-v2/backend/internal/dao"
	"schedule-system-v2/backend/internal/handler"
	"schedule-system-v2/backend/internal/middleware"
	"schedule-system-v2/backend/internal/service"

	"github.com/gin-gonic/gin"
)

func getDistPath() string {
	// 首先查找本目录下的 dist 或 frontend/dist
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
	systemHandler := handler.NewSystemHandler()
	userHandler := handler.NewUserHandler()
	availabilityHandler := handler.NewAvailabilityHandler()
	scheduleHandler := handler.NewScheduleHandler()
	crawlerHandler := handler.NewCrawlerHandler()
	templateHandler := handler.NewTemplateHandler()
	weeklyDutyHandler := handler.NewWeeklyDutyHandler()
	tempPermissionHandler := handler.NewTempPermissionHandler()
	smtpHandler := handler.NewSMTPHandler()

	// 初始化应用系统相关服务
	var applicationHandler *handler.ApplicationHandler
	if config.IsInstalled() {
		userDao := dao.NewUserDAO()
		tempPermissionDao := dao.NewTempPermissionDAO()
		applicationDao := dao.NewApplicationDao()

		appManager := service.NewApplicationManager()
		applicationService := service.NewApplicationService(applicationDao, userDao, appManager)

		// 注册临权申请执行器
		tempPermExecutor := service.NewTempPermissionExecutor(userDao, tempPermissionDao, applicationDao)
		appManager.Register(tempPermExecutor)

		applicationHandler = handler.NewApplicationHandler(applicationService)
	}

	// 公开API（不需要认证）
	api := r.Group("/api/v1")
	{
		// 系统安装相关（公开）
		api.GET("/system/installed", systemHandler.GetInstallStatus)
		api.POST("/system/test-db", systemHandler.TestDBConnection)
		api.POST("/system/check-db", systemHandler.CheckDatabase)
		api.POST("/system/init-tables", systemHandler.InitDatabaseTables)
		api.POST("/system/create-admin", systemHandler.CreateAdmin)

		// SMTP 相关（公开）
		api.GET("/smtp/check", smtpHandler.CheckConfig)
		api.POST("/password/reset-request", smtpHandler.SendPasswordReset)
		api.GET("/password/reset-verify", smtpHandler.VerifyResetToken)
		api.POST("/password/reset", smtpHandler.ResetPassword)

		// 用户相关（公开）
		api.POST("/user/register", userHandler.Register)
		api.POST("/user/login", userHandler.Login)

		// 排班相关（公开）
		api.GET("/schedule/current-week", scheduleHandler.GetCurrentWeek)
		api.GET("/schedule/current", scheduleHandler.GetCurrentWeekSchedule)
	}

	// 需要登录的API
	authGroup := api.Group("/")
	authGroup.Use(middleware.AuthMiddleware())
	{
		// 用户个人信息（普通用户权限）
		authGroup.GET("/user/profile", middleware.PermissionMiddleware(auth.PermUserProfile), userHandler.GetProfile)
		authGroup.PUT("/user/profile", middleware.PermissionMiddleware(auth.PermUserProfile), userHandler.UpdateProfile)
		authGroup.POST("/user/change-password", middleware.PermissionMiddleware(auth.PermUserProfile), userHandler.ChangePassword)

		// 无课表相关（普通用户权限）
		authGroup.POST("/availability", middleware.PermissionMiddleware(auth.PermAvailabilityEdit), availabilityHandler.AddAvailability)
		authGroup.GET("/availability", middleware.PermissionMiddleware(auth.PermAvailabilityView), availabilityHandler.GetMyAvailability)
		authGroup.DELETE("/availability", middleware.PermissionMiddleware(auth.PermAvailabilityEdit), availabilityHandler.DeleteAvailability)
		authGroup.POST("/availability/import/cookie", middleware.PermissionMiddleware(auth.PermAvailabilityImport), availabilityHandler.ImportFromCookie)
		authGroup.POST("/availability/import/xls", middleware.PermissionMiddleware(auth.PermAvailabilityImport), availabilityHandler.ImportFromXLS)
		authGroup.GET("/availability/import/status", middleware.PermissionMiddleware(auth.PermAvailabilityImport), availabilityHandler.GetImportTaskStatus)
		authGroup.GET("/availability/import/tasks", middleware.PermissionMiddleware(auth.PermAvailabilityImport), availabilityHandler.GetImportTaskList)

		// 爬虫相关（普通用户权限）
		authGroup.POST("/crawler/import", middleware.PermissionMiddleware(auth.PermAvailabilityImport), crawlerHandler.CrawlSchedule)
		authGroup.POST("/crawler/preview", middleware.PermissionMiddleware(auth.PermAvailabilityImport), crawlerHandler.PreviewSchedule)

		// 排班查看（普通用户权限）
		authGroup.GET("/schedule", middleware.PermissionMiddleware(auth.PermScheduleView), scheduleHandler.GetSchedule)
		authGroup.GET("/duty/my", middleware.PermissionMiddleware(auth.PermScheduleView), scheduleHandler.GetMyDuties)
		authGroup.PUT("/duty/status", middleware.PermissionMiddleware(auth.PermScheduleView), scheduleHandler.UpdateDutyStatus)
	}

	// 排班相关用户查询（需要排班查看权限即可）
	authGroup.GET("/users/for-schedule", middleware.PermissionMiddleware(auth.PermScheduleViewDept), userHandler.GetUsersForSchedule)

	// 需要管理员权限的API（使用新的权限检查中间件）
	admin := authGroup.Group("/admin")
	{
		// 用户管理（PermUserManage 或 PermUserManageDept 均可访问，部门级限制在 handler 中处理）
		userPermMW := middleware.PermissionAnyMiddleware(auth.PermUserManage, auth.PermUserManageDept)
		admin.GET("/users", userPermMW, userHandler.GetUserList)
		admin.POST("/users", middleware.PermissionMiddleware(auth.PermUserManage), userHandler.CreateUser)
		admin.PUT("/users/:id", userPermMW, userHandler.AdminUpdateUser)
		admin.DELETE("/users/:id", userPermMW, userHandler.DeleteUser)
		admin.PUT("/users/:id/role", middleware.PermissionMiddleware(auth.PermUserSetRole), userHandler.SetUserRole)
		admin.GET("/users/by-dept", middleware.PermissionMiddleware(auth.PermUserManageDept), userHandler.GetUserListByDepartment)
		admin.PUT("/users/:id/department", middleware.PermissionMiddleware(auth.PermUserManage), userHandler.SetUserDepartment)
		admin.PUT("/users/:id/dept-role", middleware.PermissionMiddleware(auth.PermUserManage), userHandler.SetUserDeptRole)
		admin.GET("/users/filter", middleware.PermissionMiddleware(auth.PermUserManage), userHandler.GetUsersByFilter)

		// 查看所有无课表
		admin.GET("/availability/all", middleware.PermissionMiddleware(auth.PermAvailabilityViewAll), availabilityHandler.GetAllAvailability)

		// 排班管理
		admin.POST("/schedule/preview", middleware.PermissionMiddleware(auth.PermSchedulePreview), scheduleHandler.PreviewSchedule)
		admin.POST("/schedule/confirm", middleware.PermissionMiddleware(auth.PermScheduleConfirm), scheduleHandler.ConfirmSchedule)
		admin.GET("/schedule/settings", middleware.PermissionMiddleware(auth.PermScheduleSettings), scheduleHandler.GetScheduleSettings)
		admin.POST("/schedule/settings", middleware.PermissionMiddleware(auth.PermScheduleSettings), scheduleHandler.SaveScheduleSettings)
		admin.POST("/schedule/update", middleware.PermissionMiddleware(auth.PermScheduleEdit), scheduleHandler.UpdateSchedule)
		admin.POST("/schedule/export", middleware.PermissionMiddleware(auth.PermScheduleExport), scheduleHandler.ExportSchedule)
		admin.POST("/schedule/current-week", middleware.PermissionMiddleware(auth.PermScheduleSettings), scheduleHandler.UpdateCurrentWeek)

		// 学期起始日设置
		admin.GET("/schedule/semester-start", middleware.PermissionMiddleware(auth.PermScheduleSettings), scheduleHandler.GetSemesterStartDate)
		admin.POST("/schedule/semester-start", middleware.PermissionMiddleware(auth.PermScheduleSettings), scheduleHandler.UpdateSemesterStartDate)

		// 每周分工管理
		admin.POST("/duty-assignments", middleware.PermissionMiddleware(auth.PermSchedulePublish), weeklyDutyHandler.PublishAssignment)
		admin.GET("/duty-assignments", middleware.PermissionMiddleware(auth.PermScheduleViewAll), weeklyDutyHandler.GetAssignments)
		admin.GET("/duty-assignments/view", middleware.PermissionMiddleware(auth.PermScheduleViewAll), weeklyDutyHandler.GetAssignmentView)
		admin.PUT("/duty-assignments", middleware.PermissionMiddleware(auth.PermSchedulePublish), weeklyDutyHandler.UpdateAssignment)
		admin.DELETE("/duty-assignments/:id", middleware.PermissionMiddleware(auth.PermSchedulePublish), weeklyDutyHandler.DeleteAssignment)

		// 临时权限管理 - 部门管理员及以上可以管理
		admin.POST("/temp-permissions", middleware.PermissionMiddleware(auth.PermUserManageDept), tempPermissionHandler.GrantPermission)
		admin.GET("/temp-permissions", middleware.PermissionMiddleware(auth.PermUserManageDept), tempPermissionHandler.GetAllPermissions)
		admin.DELETE("/temp-permissions/:id", middleware.PermissionMiddleware(auth.PermUserManageDept), tempPermissionHandler.RevokePermission)
		admin.POST("/temp-permissions/cleanup", middleware.PermissionMiddleware(auth.PermSystemAdmin), tempPermissionHandler.CleanupExpired)

		// 模板管理
		admin.GET("/templates", middleware.PermissionMiddleware(auth.PermTemplateView), templateHandler.GetTemplates)
		admin.GET("/templates/:id", middleware.PermissionMiddleware(auth.PermTemplateView), templateHandler.GetTemplate)
		admin.POST("/templates", middleware.PermissionMiddleware(auth.PermTemplateEdit), templateHandler.CreateTemplate)
		admin.PUT("/templates", middleware.PermissionMiddleware(auth.PermTemplateEdit), templateHandler.UpdateTemplate)
		admin.DELETE("/templates/:id", middleware.PermissionMiddleware(auth.PermTemplateEdit), templateHandler.DeleteTemplate)
		admin.GET("/templates/placeholders", middleware.PermissionMiddleware(auth.PermTemplateView), templateHandler.GetPlaceholderHelp)

		// SMTP 管理（系统管理员）
		admin.GET("/smtp/configs", middleware.PermissionMiddleware(auth.PermSystemAdmin), smtpHandler.GetConfigs)
		admin.GET("/smtp/config", middleware.PermissionMiddleware(auth.PermSystemAdmin), smtpHandler.GetActiveConfig)
		admin.POST("/smtp/config", middleware.PermissionMiddleware(auth.PermSystemAdmin), smtpHandler.SaveConfig)
		admin.DELETE("/smtp/config/:id", middleware.PermissionMiddleware(auth.PermSystemAdmin), smtpHandler.DeleteConfig)
		admin.POST("/smtp/test", middleware.PermissionMiddleware(auth.PermSystemAdmin), smtpHandler.TestEmail)

		// 网站配置（系统管理员）
		admin.GET("/site/config", middleware.PermissionMiddleware(auth.PermSystemAdmin), smtpHandler.GetSiteConfig)
		admin.POST("/site/config", middleware.PermissionMiddleware(auth.PermSystemAdmin), smtpHandler.SaveSiteConfig)
	}

	// 普通用户也可访问的分工相关API
	authGroup.GET("/duty-assignments/my-dept", middleware.PermissionMiddleware(auth.PermScheduleViewDept), weeklyDutyHandler.GetMyDeptAssignment)
	authGroup.GET("/temp-permissions/my", middleware.PermissionMiddleware(auth.PermUserProfile), tempPermissionHandler.GetMyPermissions)

	// 公开部门列表
	authGroup.GET("/departments", userHandler.GetDepartments)
	authGroup.GET("/permissions/list", tempPermissionHandler.GetPermissionList)

	// 应用系统（申请/审批）路由
	if applicationHandler != nil {
		// 申请类型
		authGroup.GET("/application/types", applicationHandler.GetTypes)
		authGroup.GET("/application/permissions/available", applicationHandler.GetAvailablePermissions)

		// 我的申请
		authGroup.GET("/applications/my", applicationHandler.GetMyApplications)
		authGroup.POST("/applications", applicationHandler.Create)
		authGroup.GET("/applications/:id", applicationHandler.GetDetail)
		authGroup.POST("/applications/:id/cancel", applicationHandler.Cancel)

		// 待我审批
		authGroup.GET("/applications/pending", applicationHandler.GetPendingApprovals)
		authGroup.POST("/applications/:id/approve", applicationHandler.ProcessApproval)

		// 申请统计
		authGroup.GET("/applications/stats", applicationHandler.GetStats)
	}

	// 静态文件服务
	distPath := getDistPath()
	assetsPath := filepath.Join(distPath, "assets")
	indexPath := filepath.Join(distPath, "index.html")

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

	r.Static("/assets", assetsPath)
	r.StaticFile("/", indexPath)

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if len(path) < 4 || path[:4] != "/api" {
			if _, err := os.Stat(indexPath); err == nil {
				c.File(indexPath)
			} else {
				c.JSON(http.StatusOK, gin.H{
					"code":    200,
					"message": "排班系统后端运行中，前端未构建",
				})
			}
		} else {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "API not found"})
		}
	})

	return r
}
