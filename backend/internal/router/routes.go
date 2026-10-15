package router

import (
	"schedule-system-v2/backend/internal/auth"
	"schedule-system-v2/backend/internal/handler"

	"github.com/gin-gonic/gin"
)

// RouteDef 路由定义
// 公开接口：Auth=false
// 登录即可：Auth=true, Perms 为空
// 单权限：Auth=true, Perms=[perm]
// 多权限任一满足：Auth=true, Perms=[p1, p2, ...]
type RouteDef struct {
	Method  string
	Path    string
	Handler gin.HandlerFunc
	Auth    bool
	Perms   []auth.Permission
}

// Handlers 聚合所有 handler 实例
type Handlers struct {
	System       *handler.SystemHandler
	Update       *handler.UpdateHandler
	User         *handler.UserHandler
	Availability *handler.AvailabilityHandler
	Schedule     *handler.ScheduleHandler
	Template     *handler.TemplateHandler
	WeeklyDuty   *handler.WeeklyDutyHandler
	TempPerm     *handler.TempPermissionHandler
	SMTP         *handler.SMTPHandler
	QRLogin      *handler.QRLoginHandler
	Invitation   *handler.InvitationHandler
	Application  *handler.ApplicationHandler
}

// buildRoutes 构建全部路由表（单一事实来源：接口 → 权限）
func buildRoutes(h *Handlers) []RouteDef {
	routes := []RouteDef{
		// ========== 公开 API（无需认证） ==========
		{Method: "GET", Path: "/system/installed", Handler: h.System.GetInstallStatus},
		{Method: "GET", Path: "/system/version", Handler: h.Update.GetVersion},
		{Method: "GET", Path: "/system/health", Handler: h.Update.HealthCheck},
		{Method: "POST", Path: "/system/test-db", Handler: h.System.TestDBConnection},
		{Method: "POST", Path: "/system/check-db", Handler: h.System.CheckDatabase},
		{Method: "POST", Path: "/system/init-tables", Handler: h.System.InitDatabaseTables},
		{Method: "POST", Path: "/system/create-admin", Handler: h.System.CreateAdmin},

		{Method: "GET", Path: "/smtp/check", Handler: h.SMTP.CheckConfig},
		{Method: "POST", Path: "/password/reset-request", Handler: h.SMTP.SendPasswordReset},
		{Method: "GET", Path: "/password/reset-verify", Handler: h.SMTP.VerifyResetToken},
			{Method: "POST", Path: "/password/reset", Handler: h.SMTP.ResetPassword},

		{Method: "POST", Path: "/user/login", Handler: h.User.Login},

		// ========== 邀请激活（公开，受邀人自助） ==========
		{Method: "POST", Path: "/invitations/activate", Handler: h.Invitation.ActivateInvitation},

		{Method: "POST", Path: "/qr/login/start", Handler: h.QRLogin.StartQrLogin},
		{Method: "POST", Path: "/qr/login/poll", Handler: h.QRLogin.PollQrLogin},

		{Method: "GET", Path: "/schedule/current-week", Handler: h.Schedule.GetCurrentWeek},
		{Method: "GET", Path: "/schedule/current", Handler: h.Schedule.GetCurrentWeekSchedule},

		// ========== 用户个人信息 ==========
		{Method: "GET", Path: "/user/profile", Auth: true, Perms: []auth.Permission{auth.PermUserProfile}, Handler: h.User.GetProfile},
		{Method: "PUT", Path: "/user/profile", Auth: true, Perms: []auth.Permission{auth.PermUserProfile}, Handler: h.User.UpdateProfile},
		{Method: "POST", Path: "/user/change-password", Auth: true, Perms: []auth.Permission{auth.PermUserProfile}, Handler: h.User.ChangePassword},

		// ========== 无课表 ==========
		{Method: "POST", Path: "/availability", Auth: true, Perms: []auth.Permission{auth.PermAvailabilityEdit}, Handler: h.Availability.AddAvailability},
		{Method: "GET", Path: "/availability", Auth: true, Perms: []auth.Permission{auth.PermAvailabilityView}, Handler: h.Availability.GetMyAvailability},
		{Method: "DELETE", Path: "/availability", Auth: true, Perms: []auth.Permission{auth.PermAvailabilityEdit}, Handler: h.Availability.DeleteAvailability},
		{Method: "POST", Path: "/availability/import/xls", Auth: true, Perms: []auth.Permission{auth.PermAvailabilityImport}, Handler: h.Availability.ImportFromXLS},
		{Method: "POST", Path: "/availability/import/xls-base64", Auth: true, Perms: []auth.Permission{auth.PermAvailabilityImport}, Handler: h.Availability.ImportFromXLSBase64},
		{Method: "GET", Path: "/availability/import/status", Auth: true, Perms: []auth.Permission{auth.PermAvailabilityImport}, Handler: h.Availability.GetImportTaskStatus},
		{Method: "GET", Path: "/availability/import/tasks", Auth: true, Perms: []auth.Permission{auth.PermAvailabilityImport}, Handler: h.Availability.GetImportTaskList},
		{Method: "POST", Path: "/availability/import/edu", Auth: true, Perms: []auth.Permission{auth.PermAvailabilityImport}, Handler: h.Availability.ImportFromEdu},
		{Method: "POST", Path: "/availability/import/edu/resume", Auth: true, Perms: []auth.Permission{auth.PermAvailabilityImport}, Handler: h.Availability.ResumeEduImport},

		// ========== 排班查看 ==========
		{Method: "GET", Path: "/schedule", Auth: true, Perms: []auth.Permission{auth.PermScheduleView}, Handler: h.Schedule.GetSchedule},
		{Method: "GET", Path: "/duty/my", Auth: true, Perms: []auth.Permission{auth.PermScheduleView}, Handler: h.Schedule.GetMyDuties},
		{Method: "PUT", Path: "/duty/status", Auth: true, Perms: []auth.Permission{auth.PermScheduleView}, Handler: h.Schedule.UpdateDutyStatus},

		// ========== 排班相关用户查询 ==========
		{Method: "GET", Path: "/users/for-schedule", Auth: true, Perms: []auth.Permission{auth.PermScheduleViewDept}, Handler: h.User.GetUsersForSchedule},

		// ========== 分工（部门查看） ==========
		{Method: "GET", Path: "/duty-assignments/my-dept", Auth: true, Perms: []auth.Permission{auth.PermScheduleViewDept}, Handler: h.WeeklyDuty.GetMyDeptAssignment},

		// ========== 我的临时权限 ==========
		{Method: "GET", Path: "/temp-permissions/my", Auth: true, Perms: []auth.Permission{auth.PermUserProfile}, Handler: h.TempPerm.GetMyPermissions},

		// ========== 公共数据（登录即可） ==========
		{Method: "GET", Path: "/departments", Auth: true, Handler: h.User.GetDepartments},
		{Method: "GET", Path: "/permissions/list", Auth: true, Handler: h.TempPerm.GetPermissionList},

		// ========== 用户管理 ==========
		{Method: "GET", Path: "/admin/users", Auth: true, Perms: []auth.Permission{auth.PermUserManage, auth.PermUserManageDept}, Handler: h.User.GetUserList},
		{Method: "POST", Path: "/admin/users", Auth: true, Perms: []auth.Permission{auth.PermUserManage}, Handler: h.User.CreateUser},
		{Method: "PUT", Path: "/admin/users/:id", Auth: true, Perms: []auth.Permission{auth.PermUserManage, auth.PermUserManageDept}, Handler: h.User.AdminUpdateUser},
		{Method: "DELETE", Path: "/admin/users/:id", Auth: true, Perms: []auth.Permission{auth.PermUserManage, auth.PermUserManageDept}, Handler: h.User.DeleteUser},
		{Method: "PUT", Path: "/admin/users/:id/role", Auth: true, Perms: []auth.Permission{auth.PermUserSetRole}, Handler: h.User.SetUserRole},
		{Method: "GET", Path: "/admin/users/by-dept", Auth: true, Perms: []auth.Permission{auth.PermUserManageDept}, Handler: h.User.GetUserListByDepartment},
		{Method: "PUT", Path: "/admin/users/:id/department", Auth: true, Perms: []auth.Permission{auth.PermUserManage}, Handler: h.User.SetUserDepartment},
		{Method: "PUT", Path: "/admin/users/:id/dept-role", Auth: true, Perms: []auth.Permission{auth.PermUserManage}, Handler: h.User.SetUserDeptRole},
		{Method: "GET", Path: "/admin/users/filter", Auth: true, Perms: []auth.Permission{auth.PermUserManage}, Handler: h.User.GetUsersByFilter},

		// ========== 录入正式成员（邀请激活链路） ==========
		{Method: "POST", Path: "/admin/members", Auth: true, Perms: []auth.Permission{auth.PermUserManage}, Handler: h.Invitation.CreateMember},
		{Method: "POST", Path: "/admin/members/batch", Auth: true, Perms: []auth.Permission{auth.PermUserManage}, Handler: h.Invitation.BatchCreateMembers},
		{Method: "POST", Path: "/admin/members/:id/resend-invite", Auth: true, Perms: []auth.Permission{auth.PermUserManage}, Handler: h.Invitation.ResendInvite},
		{Method: "POST", Path: "/admin/members/:id/revoke", Auth: true, Perms: []auth.Permission{auth.PermUserManage}, Handler: h.Invitation.RevokeInvite},

		// ========== 离职（单/批，换届制 H.4.1） ==========
		{Method: "POST", Path: "/admin/members/:id/resign", Auth: true, Perms: []auth.Permission{auth.PermUserManage}, Handler: h.User.ResignUser},
		{Method: "POST", Path: "/admin/members/batch-resign/preview", Auth: true, Perms: []auth.Permission{auth.PermUserManage}, Handler: h.User.BatchResignPreview},
		{Method: "POST", Path: "/admin/members/batch-resign", Auth: true, Perms: []auth.Permission{auth.PermUserManage}, Handler: h.User.BatchResign},

		// ========== 无课表（管理员） ==========
		{Method: "GET", Path: "/admin/availability/all", Auth: true, Perms: []auth.Permission{auth.PermAvailabilityViewAll}, Handler: h.Availability.GetAllAvailability},

		// ========== 排班管理 ==========
		{Method: "POST", Path: "/admin/schedule/preview", Auth: true, Perms: []auth.Permission{auth.PermSchedulePreview}, Handler: h.Schedule.PreviewSchedule},
		{Method: "POST", Path: "/admin/schedule/confirm", Auth: true, Perms: []auth.Permission{auth.PermScheduleConfirm}, Handler: h.Schedule.ConfirmSchedule},
		{Method: "GET", Path: "/admin/schedule/settings", Auth: true, Perms: []auth.Permission{auth.PermScheduleSettings}, Handler: h.Schedule.GetScheduleSettings},
		{Method: "POST", Path: "/admin/schedule/settings", Auth: true, Perms: []auth.Permission{auth.PermScheduleSettings}, Handler: h.Schedule.SaveScheduleSettings},
		{Method: "POST", Path: "/admin/schedule/update", Auth: true, Perms: []auth.Permission{auth.PermScheduleEdit}, Handler: h.Schedule.UpdateSchedule},
		{Method: "POST", Path: "/admin/schedule/export", Auth: true, Perms: []auth.Permission{auth.PermScheduleExport}, Handler: h.Schedule.ExportSchedule},
		{Method: "POST", Path: "/admin/schedule/current-week", Auth: true, Perms: []auth.Permission{auth.PermScheduleSettings}, Handler: h.Schedule.UpdateCurrentWeek},
		{Method: "GET", Path: "/admin/schedule/semester-start", Auth: true, Perms: []auth.Permission{auth.PermScheduleSettings}, Handler: h.Schedule.GetSemesterStartDate},
		{Method: "POST", Path: "/admin/schedule/semester-start", Auth: true, Perms: []auth.Permission{auth.PermScheduleSettings}, Handler: h.Schedule.UpdateSemesterStartDate},

		// ========== 每周分工管理 ==========
		{Method: "POST", Path: "/admin/duty-assignments", Auth: true, Perms: []auth.Permission{auth.PermSchedulePublish}, Handler: h.WeeklyDuty.PublishAssignment},
		{Method: "GET", Path: "/admin/duty-assignments", Auth: true, Perms: []auth.Permission{auth.PermScheduleViewAll}, Handler: h.WeeklyDuty.GetAssignments},
		{Method: "GET", Path: "/admin/duty-assignments/view", Auth: true, Perms: []auth.Permission{auth.PermScheduleViewAll}, Handler: h.WeeklyDuty.GetAssignmentView},
		{Method: "PUT", Path: "/admin/duty-assignments", Auth: true, Perms: []auth.Permission{auth.PermSchedulePublish}, Handler: h.WeeklyDuty.UpdateAssignment},
		{Method: "DELETE", Path: "/admin/duty-assignments/:id", Auth: true, Perms: []auth.Permission{auth.PermSchedulePublish}, Handler: h.WeeklyDuty.DeleteAssignment},

		// ========== 临时权限管理 ==========
		{Method: "POST", Path: "/admin/temp-permissions", Auth: true, Perms: []auth.Permission{auth.PermUserManageDept}, Handler: h.TempPerm.GrantPermission},
		{Method: "GET", Path: "/admin/temp-permissions", Auth: true, Perms: []auth.Permission{auth.PermUserManageDept}, Handler: h.TempPerm.GetAllPermissions},
		{Method: "DELETE", Path: "/admin/temp-permissions/:id", Auth: true, Perms: []auth.Permission{auth.PermUserManageDept}, Handler: h.TempPerm.RevokePermission},
		{Method: "POST", Path: "/admin/temp-permissions/cleanup", Auth: true, Perms: []auth.Permission{auth.PermSystemAdmin}, Handler: h.TempPerm.CleanupExpired},

		// ========== 模板管理 ==========
		{Method: "GET", Path: "/admin/templates", Auth: true, Perms: []auth.Permission{auth.PermTemplateView}, Handler: h.Template.GetTemplates},
		{Method: "GET", Path: "/admin/templates/:id", Auth: true, Perms: []auth.Permission{auth.PermTemplateView}, Handler: h.Template.GetTemplate},
		{Method: "POST", Path: "/admin/templates", Auth: true, Perms: []auth.Permission{auth.PermTemplateEdit}, Handler: h.Template.CreateTemplate},
		{Method: "PUT", Path: "/admin/templates", Auth: true, Perms: []auth.Permission{auth.PermTemplateEdit}, Handler: h.Template.UpdateTemplate},
		{Method: "DELETE", Path: "/admin/templates/:id", Auth: true, Perms: []auth.Permission{auth.PermTemplateEdit}, Handler: h.Template.DeleteTemplate},
		{Method: "GET", Path: "/admin/templates/placeholders", Auth: true, Perms: []auth.Permission{auth.PermTemplateView}, Handler: h.Template.GetPlaceholderHelp},

		// ========== SMTP / 网站配置（系统管理员） ==========
		{Method: "GET", Path: "/admin/smtp/configs", Auth: true, Perms: []auth.Permission{auth.PermSystemAdmin}, Handler: h.SMTP.GetConfigs},
		{Method: "GET", Path: "/admin/smtp/config", Auth: true, Perms: []auth.Permission{auth.PermSystemAdmin}, Handler: h.SMTP.GetActiveConfig},
		{Method: "POST", Path: "/admin/smtp/config", Auth: true, Perms: []auth.Permission{auth.PermSystemAdmin}, Handler: h.SMTP.SaveConfig},
		{Method: "DELETE", Path: "/admin/smtp/config/:id", Auth: true, Perms: []auth.Permission{auth.PermSystemAdmin}, Handler: h.SMTP.DeleteConfig},
		{Method: "POST", Path: "/admin/smtp/test", Auth: true, Perms: []auth.Permission{auth.PermSystemAdmin}, Handler: h.SMTP.TestEmail},
		{Method: "GET", Path: "/admin/site/config", Auth: true, Perms: []auth.Permission{auth.PermSystemAdmin}, Handler: h.SMTP.GetSiteConfig},
		{Method: "POST", Path: "/admin/site/config", Auth: true, Perms: []auth.Permission{auth.PermSystemAdmin}, Handler: h.SMTP.SaveSiteConfig},

		// ========== 热更新（系统管理员） ==========
		{Method: "GET", Path: "/admin/update/check", Auth: true, Perms: []auth.Permission{auth.PermSystemAdmin}, Handler: h.Update.CheckUpdate},
		{Method: "POST", Path: "/admin/update/start", Auth: true, Perms: []auth.Permission{auth.PermSystemAdmin}, Handler: h.Update.StartUpdate},
		{Method: "GET", Path: "/admin/update/status", Auth: true, Perms: []auth.Permission{auth.PermSystemAdmin}, Handler: h.Update.GetUpdateStatus},
		{Method: "GET", Path: "/admin/update/log", Auth: true, Perms: []auth.Permission{auth.PermSystemAdmin}, Handler: h.Update.UpdateLogTail},
		}

	// ========== 应用系统（申请/审批） ==========
	routes = append(routes,
		// 申请类型
		RouteDef{Method: "GET", Path: "/application/types", Auth: true, Handler: h.Application.GetTypes},
		RouteDef{Method: "GET", Path: "/application/permissions/available", Auth: true, Handler: h.Application.GetAvailablePermissions},

		// 我的申请
		RouteDef{Method: "GET", Path: "/applications/my", Auth: true, Handler: h.Application.GetMyApplications},
		RouteDef{Method: "POST", Path: "/applications", Auth: true, Handler: h.Application.Create},
		RouteDef{Method: "GET", Path: "/applications/:id", Auth: true, Handler: h.Application.GetDetail},
		RouteDef{Method: "POST", Path: "/applications/:id/cancel", Auth: true, Handler: h.Application.Cancel},

		// 待我审批
		RouteDef{Method: "GET", Path: "/applications/pending", Auth: true, Handler: h.Application.GetPendingApprovals},
		RouteDef{Method: "POST", Path: "/applications/:id/approve", Auth: true, Handler: h.Application.ProcessApproval},

		// 申请统计
		RouteDef{Method: "GET", Path: "/applications/stats", Auth: true, Handler: h.Application.GetStats},
	)

	return routes
}
