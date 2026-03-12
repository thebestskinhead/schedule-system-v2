package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"schedule-system-v2/backend/internal/auth"
	"schedule-system-v2/backend/internal/model"
	"schedule-system-v2/backend/internal/service"
	"schedule-system-v2/backend/internal/utils"
)

type WeeklyDutyHandler struct {
	service *service.WeeklyDutyService
}

func NewWeeklyDutyHandler() *WeeklyDutyHandler {
	return &WeeklyDutyHandler{
		service: service.NewWeeklyDutyService(),
	}
}

// PublishAssignment 发布每周分工
func (h *WeeklyDutyHandler) PublishAssignment(c *gin.Context) {
	// 权限检查
	checker := auth.GetChecker(c)
	if !h.service.CanPublish(checker) {
		auth.ResponseForbidden(c, "只有系统管理员或办公室管理员可以发布分工")
		return
	}

	var req model.PublishAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	adminID := auth.GetUserIDFromContext(c)
	if err := h.service.PublishAssignment(adminID, &req); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, nil)
}

// GetAssignments 获取分工列表
func (h *WeeklyDutyHandler) GetAssignments(c *gin.Context) {
	weekStr := c.Query("week")
	if weekStr == "" {
		utils.Error(c, 400, "缺少week参数")
		return
	}

	week, err := strconv.Atoi(weekStr)
	if err != nil || week < 1 || week > 30 {
		utils.Error(c, 400, "无效的week参数")
		return
	}

	checker := auth.GetChecker(c)
	userDept := auth.GetDepartmentFromContext(c)

	assignments, err := h.service.GetWeekAssignments(week, userDept, checker)
	if err != nil {
		utils.Error(c, 500, "获取分工失败: "+err.Error())
		return
	}

	utils.Success(c, assignments)
}

// GetAssignmentView 获取分工视图（按部门聚合）
func (h *WeeklyDutyHandler) GetAssignmentView(c *gin.Context) {
	weekStr := c.Query("week")
	if weekStr == "" {
		utils.Error(c, 400, "缺少week参数")
		return
	}

	week, err := strconv.Atoi(weekStr)
	if err != nil || week < 1 || week > 30 {
		utils.Error(c, 400, "无效的week参数")
		return
	}

	checker := auth.GetChecker(c)

	view, err := h.service.GetWeekAssignmentView(week, checker)
	if err != nil {
		utils.Error(c, 500, "获取分工视图失败: "+err.Error())
		return
	}

	utils.Success(c, view)
}

// GetMyDeptAssignment 获取本部门分工
func (h *WeeklyDutyHandler) GetMyDeptAssignment(c *gin.Context) {
	weekStr := c.Query("week")
	if weekStr == "" {
		utils.Error(c, 400, "缺少week参数")
		return
	}

	week, err := strconv.Atoi(weekStr)
	if err != nil || week < 1 || week > 30 {
		utils.Error(c, 400, "无效的week参数")
		return
	}

	userDept := auth.GetDepartmentFromContext(c)
	if userDept == "" {
		utils.Error(c, 400, "用户未设置部门")
		return
	}

	assignment, err := h.service.GetMyDeptAssignment(week, userDept)
	if err != nil {
		utils.Error(c, 500, "获取部门分工失败: "+err.Error())
		return
	}

	utils.Success(c, assignment)
}

// UpdateAssignment 更新分工
func (h *WeeklyDutyHandler) UpdateAssignment(c *gin.Context) {
	// 权限检查
	checker := auth.GetChecker(c)
	if !h.service.CanPublish(checker) {
		auth.ResponseForbidden(c, "只有系统管理员或办公室管理员可以更新分工")
		return
	}

	var req model.UpdateAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	adminID := auth.GetUserIDFromContext(c)
	if err := h.service.UpdateAssignment(adminID, &req); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, nil)
}

// DeleteAssignment 删除分工
func (h *WeeklyDutyHandler) DeleteAssignment(c *gin.Context) {
	// 权限检查
	checker := auth.GetChecker(c)
	if !h.service.CanPublish(checker) {
		auth.ResponseForbidden(c, "只有系统管理员或办公室管理员可以删除分工")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error(c, 400, "无效的ID")
		return
	}

	adminID := auth.GetUserIDFromContext(c)
	if err := h.service.DeleteAssignment(adminID, id); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, nil)
}
