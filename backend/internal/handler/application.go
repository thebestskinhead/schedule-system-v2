package handler

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"schedule-system-v2/backend/internal/auth"
	"schedule-system-v2/backend/internal/model"
	"schedule-system-v2/backend/internal/service"
	"schedule-system-v2/backend/internal/utils"
)

type ApplicationHandler struct {
	applicationService *service.ApplicationService
}

func NewApplicationHandler(applicationService *service.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{
		applicationService: applicationService,
	}
}

// CreateApplicationRequest 创建申请请求
type CreateApplicationRequest struct {
	Type   string          `json:"type" binding:"required"`
	Data   json.RawMessage `json:"data" binding:"required"`
	Reason string          `json:"reason"`
}

// Create 创建申请
func (h *ApplicationHandler) Create(c *gin.Context) {
	var req CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	// 获取当前用户
	checker := auth.GetChecker(c)
	if checker == nil {
		utils.Error(c, 401, "未登录")
		return
	}

	application, err := h.applicationService.CreateApplication(
		c.Request.Context(),
		checker.GetUserID(),
		req.Type,
		req.Data,
		req.Reason,
	)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, application)
}

// GetMyApplications 获取我的申请列表
func (h *ApplicationHandler) GetMyApplications(c *gin.Context) {
	checker := auth.GetChecker(c)
	if checker == nil {
		utils.Error(c, 401, "未登录")
		return
	}

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	applications, total, err := h.applicationService.GetMyApplications(
		c.Request.Context(),
		checker.GetUserID(),
		status,
		page,
		pageSize,
	)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	// 将 total 包含在 data 中，统一响应格式
	utils.Success(c, gin.H{
		"list":  applications,
		"total": total,
	})
}

// GetPendingApprovals 获取待我审批的申请
func (h *ApplicationHandler) GetPendingApprovals(c *gin.Context) {
	checker := auth.GetChecker(c)
	if checker == nil {
		utils.Error(c, 401, "未登录")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	applications, total, err := h.applicationService.GetPendingApprovals(
		c.Request.Context(),
		checker.GetUserID(),
		checker.IsAdmin(),
		checker.GetDepartment(),
		page,
		pageSize,
	)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	// 将 total 包含在 data 中，统一响应格式
	utils.Success(c, gin.H{
		"list":  applications,
		"total": total,
	})
}

// GetDetail 获取申请详情
func (h *ApplicationHandler) GetDetail(c *gin.Context) {
	checker := auth.GetChecker(c)
	if checker == nil {
		utils.Error(c, 401, "未登录")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Error(c, 400, "无效的ID")
		return
	}

	application, err := h.applicationService.GetApplicationDetail(
		c.Request.Context(),
		id,
		checker.GetUserID(),
		checker.IsAdmin(),
	)
	if err != nil {
		utils.Error(c, 403, err.Error())
		return
	}

	utils.Success(c, application)
}

// ProcessApprovalRequest 处理审批请求
type ProcessApprovalRequest struct {
	Action  string `json:"action" binding:"required"` // approve, reject, comment
	Comment string `json:"comment"`
}

// ProcessApproval 处理审批
func (h *ApplicationHandler) ProcessApproval(c *gin.Context) {
	checker := auth.GetChecker(c)
	if checker == nil {
		utils.Error(c, 401, "未登录")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Error(c, 400, "无效的ID")
		return
	}

	var req ProcessApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	// 转换 action
	var action model.ApprovalAction
	switch req.Action {
	case "approve":
		action = model.ApprovalActionApprove
	case "reject":
		action = model.ApprovalActionReject
	case "transfer":
		action = model.ApprovalActionTransfer
	case "comment":
		action = model.ApprovalActionComment
	default:
		utils.Error(c, 400, "无效的操作类型")
		return
	}

	if err := h.applicationService.ProcessApproval(
		c.Request.Context(),
		id,
		checker.GetUserID(),
		action,
		req.Comment,
	); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "审批处理成功"})
}

// Cancel 取消申请
func (h *ApplicationHandler) Cancel(c *gin.Context) {
	checker := auth.GetChecker(c)
	if checker == nil {
		utils.Error(c, 401, "未登录")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Error(c, 400, "无效的ID")
		return
	}

	if err := h.applicationService.CancelApplication(
		c.Request.Context(),
		id,
		checker.GetUserID(),
	); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "申请已取消"})
}

// GetTypes 获取申请类型列表
func (h *ApplicationHandler) GetTypes(c *gin.Context) {
	types, err := h.applicationService.GetApplicationTypes(c.Request.Context())
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, types)
}

// GetStats 获取申请统计
func (h *ApplicationHandler) GetStats(c *gin.Context) {
	checker := auth.GetChecker(c)
	if checker == nil {
		utils.Error(c, 401, "未登录")
		return
	}

	stats, err := h.applicationService.GetApplicationStats(
		c.Request.Context(),
		checker.GetUserID(),
		checker.IsAdmin(),
		checker.GetDepartment(),
	)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, stats)
}

// GetAvailablePermissions 获取可申请的权限列表
func (h *ApplicationHandler) GetAvailablePermissions(c *gin.Context) {
	checker := auth.GetChecker(c)
	if checker == nil {
		utils.Error(c, 401, "未登录")
		return
	}

	// 构建用户信息
	user := &model.User{
		ID:         checker.GetUserID(),
		Role:       checker.GetRole(),
		Department: checker.GetDepartment(),
		DeptRole:   checker.GetDeptRole(),
	}

	permissions := service.GetAvailablePermissions(user)
	utils.Success(c, permissions)
}
