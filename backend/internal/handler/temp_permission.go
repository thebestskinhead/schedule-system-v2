package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"schedule-system-v2/backend/internal/auth"
	"schedule-system-v2/backend/internal/model"
	"schedule-system-v2/backend/internal/service"
	"schedule-system-v2/backend/internal/utils"
)

type TempPermissionHandler struct {
	service *service.TempPermissionService
}

func NewTempPermissionHandler() *TempPermissionHandler {
	return &TempPermissionHandler{
		service: service.NewTempPermissionService(),
	}
}

// GrantPermission 授予临时权限（支持批量）
// 注意：权限检查由路由层的 PermUserManageDept 中间件处理
func (h *TempPermissionHandler) GrantPermission(c *gin.Context) {
	checker := auth.GetChecker(c)

	var req model.GrantPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	// 确保过期时间是UTC
	if req.ExpiresAt.Location().String() != "UTC" {
		req.ExpiresAt = req.ExpiresAt.UTC()
	}

	adminID := auth.GetUserIDFromContext(c)
	
	// 批量授权
	var failedUsers []string
	for _, userID := range req.UserIDs {
		singleReq := model.SingleGrantRequest{
			UserID:       userID,
			Permission:   req.Permission,
			ResourceType: req.ResourceType,
			ResourceID:   req.ResourceID,
			ExpiresAt:    req.ExpiresAt,
			Reason:       req.Reason,
		}
		if err := h.service.GrantPermission(adminID, checker, &singleReq); err != nil {
			failedUsers = append(failedUsers, err.Error())
		}
	}

	if len(failedUsers) > 0 && len(failedUsers) == len(req.UserIDs) {
		utils.Error(c, 500, "授权失败: "+failedUsers[0])
		return
	}

	utils.Success(c, gin.H{
		"message": "授权成功",
		"total":   len(req.UserIDs),
		"failed":  len(failedUsers),
	})
}

// GetAllPermissions 获取临时权限列表（支持按部门筛选）
// 注意：权限检查由路由层的 PermUserManageDept 中间件处理
func (h *TempPermissionHandler) GetAllPermissions(c *gin.Context) {
	checker := auth.GetChecker(c)

	var perms []model.TempPermissionView
	var err error

	// 系统管理员/办公室管理员查看所有，部门管理员只查看自己部门的
	if checker.IsAdmin() || checker.IsOfficeAdmin() {
		perms, err = h.service.GetAllActivePermissions()
	} else {
		// 部门管理员只能看到自己部门的权限
		dept := checker.GetDepartment()
		perms, err = h.service.GetPermissionsByDepartment(dept)
	}

	if err != nil {
		utils.Error(c, 500, "获取权限列表失败: "+err.Error())
		return
	}

	utils.Success(c, perms)
}

// RevokePermission 撤销临时权限
// 注意：权限检查由路由层的 PermUserManageDept 中间件处理
func (h *TempPermissionHandler) RevokePermission(c *gin.Context) {
	checker := auth.GetChecker(c)

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error(c, 400, "无效的ID")
		return
	}

	adminID := auth.GetUserIDFromContext(c)
	if err := h.service.RevokePermission(adminID, checker, id); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, nil)
}

// GetMyPermissions 获取我的临时权限
func (h *TempPermissionHandler) GetMyPermissions(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}

	// 使用 GetUserTempPermissions 返回与 GetAllPermissions 一致的格式
	perms, err := h.service.GetUserTempPermissions(userID)
	if err != nil {
		utils.Error(c, 500, "获取权限失败: "+err.Error())
		return
	}

	utils.Success(c, perms)
}

// GetPermissionList 获取可授权的权限列表
func (h *TempPermissionHandler) GetPermissionList(c *gin.Context) {
	// 只有登录用户可以查看
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}

	perms := model.GetPermissionList()
	utils.Success(c, perms)
}

// CleanupExpired 手动触发清理过期权限（管理员）
// 注意：权限检查由路由层的 PermSystemAdmin 中间件处理
func (h *TempPermissionHandler) CleanupExpired(c *gin.Context) {
	if err := h.service.CleanupExpiredPermissions(); err != nil {
		utils.Error(c, 500, "清理失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "过期权限已清理"})
}
