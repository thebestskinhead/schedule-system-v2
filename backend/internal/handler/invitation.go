package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"schedule-system-v2/backend/internal/auth"
	"schedule-system-v2/backend/internal/model"
	"schedule-system-v2/backend/internal/service"
	"schedule-system-v2/backend/internal/utils"
)

type InvitationHandler struct {
	service *service.InvitationService
}

func NewInvitationHandler() *InvitationHandler {
	return &InvitationHandler{
		service: service.NewInvitationService(),
	}
}

// CreateMember 录入单个正式成员（需 user:manage）
func (h *InvitationHandler) CreateMember(c *gin.Context) {
	checker := auth.GetChecker(c)
	if !checker.HasPermission(auth.PermUserManage) {
		auth.ResponseForbidden(c, "无权录入正式成员")
		return
	}

	var req model.MemberCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	operatorID := auth.GetUserIDFromContext(c)
	user, err := h.service.CreateMember(&req, operatorID)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, model.UserInfo{
		ID:               user.ID,
		StudentID:        user.StudentID,
		Name:             user.Name,
		Email:            user.Email,
		Role:             user.Role,
		Department:       user.Department,
		DeptRole:         user.DeptRole,
		EmploymentStatus: user.EmploymentStatus,
	})
}

// BatchCreateMembers 批量录入正式成员（需 user:manage）
func (h *InvitationHandler) BatchCreateMembers(c *gin.Context) {
	checker := auth.GetChecker(c)
	if !checker.HasPermission(auth.PermUserManage) {
		auth.ResponseForbidden(c, "无权录入正式成员")
		return
	}

	var req model.BatchMemberCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	operatorID := auth.GetUserIDFromContext(c)
	result := h.service.BatchCreateMembers(req.Items, operatorID)
	utils.Success(c, result)
}

// ResendInvite 重发邀请（需 user:manage）。路径参数为用户 ID。
func (h *InvitationHandler) ResendInvite(c *gin.Context) {
	checker := auth.GetChecker(c)
	if !checker.HasPermission(auth.PermUserManage) {
		auth.ResponseForbidden(c, "无权重发邀请")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Error(c, 400, "无效的用户ID")
		return
	}

	if err := h.service.ResendInvite(id, auth.GetUserIDFromContext(c)); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	utils.Success(c, gin.H{"message": "邀请已重新发送"})
}

// RevokeInvite 作废邀请（需 user:manage）。路径参数为用户 ID。
func (h *InvitationHandler) RevokeInvite(c *gin.Context) {
	checker := auth.GetChecker(c)
	if !checker.HasPermission(auth.PermUserManage) {
		auth.ResponseForbidden(c, "无权作废邀请")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Error(c, 400, "无效的用户ID")
		return
	}

	if err := h.service.RevokeInvite(id); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	utils.Success(c, gin.H{"message": "邀请已作废"})
}

// ActivateInvitation 受邀人自助激活（公开接口）
func (h *InvitationHandler) ActivateInvitation(c *gin.Context) {
	var req model.ActivateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	if err := h.service.ActivateInvitation(req.Token, req.Password); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	utils.Success(c, gin.H{"message": "账号激活成功，请使用学号登录"})
}
