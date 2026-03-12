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

// GrantPermission 授予临时权限
func (h *TempPermissionHandler) GrantPermission(c *gin.Context) {
	// 权限检查 - 只有系统管理员可以授权
	checker := auth.GetChecker(c)
	if !checker.IsAdmin() {
		auth.ResponseForbidden(c, "只有系统管理员可以授予临时权限")
		return
	}

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
	if err := h.service.GrantPermission(adminID, &req); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, nil)
}

// GetAllPermissions 获取所有临时权限
func (h *TempPermissionHandler) GetAllPermissions(c *gin.Context) {
	// 权限检查 - 只有系统管理员可以查看所有权限
	checker := auth.GetChecker(c)
	if !checker.IsAdmin() {
		auth.ResponseForbidden(c, "只有系统管理员可以查看所有临时权限")
		return
	}

	perms, err := h.service.GetAllActivePermissions()
	if err != nil {
		utils.Error(c, 500, "获取权限列表失败: "+err.Error())
		return
	}

	utils.Success(c, perms)
}

// RevokePermission 撤销临时权限
func (h *TempPermissionHandler) RevokePermission(c *gin.Context) {
	// 权限检查 - 只有系统管理员可以撤销
	checker := auth.GetChecker(c)
	if !checker.IsAdmin() {
		auth.ResponseForbidden(c, "只有系统管理员可以撤销临时权限")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error(c, 400, "无效的ID")
		return
	}

	adminID := auth.GetUserIDFromContext(c)
	if err := h.service.RevokePermission(adminID, id); err != nil {
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

	perms, err := h.service.GetMyPermissions(userID)
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
func (h *TempPermissionHandler) CleanupExpired(c *gin.Context) {
	// 权限检查
	checker := auth.GetChecker(c)
	if !checker.IsAdmin() {
		auth.ResponseForbidden(c)
		return
	}

	if err := h.service.CleanupExpiredPermissions(); err != nil {
		utils.Error(c, 500, "清理失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "过期权限已清理"})
}
