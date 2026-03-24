package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"schedule-system-v2/backend/internal/auth"
	"schedule-system-v2/backend/internal/model"
	"schedule-system-v2/backend/internal/service"
	"schedule-system-v2/backend/internal/utils"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		service: service.NewUserService(),
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	user, err := h.service.Register(&req)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, model.UserInfo{
		ID:         user.ID,
		StudentID:  user.StudentID,
		Name:       user.Name,
		Email:      user.Email,
		Role:       user.Role,
		Department: user.Department,
		DeptRole:   user.DeptRole,
	})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	resp, err := h.service.Login(&req)
	if err != nil {
		utils.Error(c, 401, err.Error())
		return
	}

	utils.Success(c, resp)
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	// 使用新的权限检查工具获取用户ID
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}

	user, err := h.service.GetUserByID(userID)
	if err != nil {
		utils.Error(c, 404, "用户不存在")
		return
	}

	utils.Success(c, model.UserInfo{
		ID:         user.ID,
		StudentID:  user.StudentID,
		Name:       user.Name,
		Email:      user.Email,
		Role:       user.Role,
		Department: user.Department,
		DeptRole:   user.DeptRole,
	})
}

func (h *UserHandler) GetUserList(c *gin.Context) {
	// 使用新的权限检查工具
	// 方式1: 使用辅助函数
	if !auth.CheckPermission(c, auth.PermUserManage) {
		return // CheckPermission 已经返回403响应
	}

	users, err := h.service.GetUserList()
	if err != nil {
		utils.Error(c, 500, "获取用户列表失败")
		return
	}

	var result []model.UserInfo
	for _, user := range users {
		result = append(result, model.UserInfo{
			ID:         user.ID,
			StudentID:  user.StudentID,
			Name:       user.Name,
			Email:      user.Email,
			Role:       user.Role,
			Department: user.Department,
			DeptRole:   user.DeptRole,
		})
	}
	utils.Success(c, result)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// 使用新的权限检查工具获取用户ID
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if err := h.service.UpdateUser(userID, &req); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}

// ChangePassword 修改密码
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	if err := h.service.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	utils.Success(c, gin.H{"message": "密码修改成功"})
}

func (h *UserHandler) SetUserRole(c *gin.Context) {
	// 使用新的权限检查工具 - 检查是否有设置角色权限
	checker := auth.GetChecker(c)
	if !checker.HasPermission(auth.PermUserSetRole) {
		auth.ResponseForbidden(c, "无权设置用户角色")
		return
	}

	adminID := auth.GetUserIDFromContext(c)

	// 从路径参数获取用户ID
	idStr := c.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error(c, 400, "无效的用户ID")
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	if err := h.service.SetUserRole(adminID, userID, req.Role); err != nil {
		utils.Error(c, 403, err.Error())
		return
	}
	utils.Success(c, nil)
}

// GetUserListByDepartment 按部门获取用户列表
func (h *UserHandler) GetUserListByDepartment(c *gin.Context) {
	// 权限检查 - 需要部门用户管理权限或用户管理权限
	checker := auth.GetChecker(c)
	if !checker.HasPermission(auth.PermUserManageDept) && !checker.HasPermission(auth.PermUserManage) {
		auth.ResponseForbidden(c)
		return
	}

	dept := c.Query("department")
	if dept == "" {
		utils.Error(c, 400, "缺少department参数")
		return
	}

	// 非管理员只能查看自己部门的用户
	if !checker.IsAdmin() && !checker.IsOfficeAdmin() {
		if checker.GetDepartment() != dept {
			auth.ResponseForbidden(c, "只能查看本部门用户")
			return
		}
	}

	users, err := h.service.GetUserListByDepartment(dept)
	if err != nil {
		utils.Error(c, 500, "获取用户列表失败")
		return
	}

	var result []model.UserInfo
	for _, user := range users {
		result = append(result, model.UserInfo{
			ID:         user.ID,
			StudentID:  user.StudentID,
			Name:       user.Name,
			Email:      user.Email,
			Role:       user.Role,
			Department: user.Department,
			DeptRole:   user.DeptRole,
		})
	}
	utils.Success(c, result)
}

// GetUsersForSchedule 获取用于排班的用户列表（需要排班权限即可）
// 排除部门管理员，只返回部门成员
func (h *UserHandler) GetUsersForSchedule(c *gin.Context) {
	checker := auth.GetChecker(c)

	// 系统管理员或办公室管理员：返回所有用户（排除部门管理员）
	if checker.IsAdmin() || checker.IsOfficeAdmin() {
		users, err := h.service.GetUserList()
		if err != nil {
			utils.Error(c, 500, "获取用户列表失败")
			return
		}

		var result []gin.H
		for i := range users {
			// 排除部门管理员
			if users[i].DeptRole == model.DeptRoleAdmin {
				continue
			}
			result = append(result, gin.H{
				"id":         users[i].ID,
				"name":       users[i].Name,
				"student_id": users[i].StudentID,
				"department": users[i].Department,
			})
		}
		utils.Success(c, result)
		return
	}

	// 部门管理员或有排班权限的用户：只返回本部门用户（排除部门管理员）
	dept := checker.GetDepartment()
	if dept == "" {
		auth.ResponseForbidden(c, "未分配部门")
		return
	}

	users, err := h.service.GetUserListByDepartment(dept)
	if err != nil {
		utils.Error(c, 500, "获取用户列表失败")
		return
	}

	// 只返回必要的字段（排班用），排除部门管理员
	var result []gin.H
	for i := range users {
		// 排除部门管理员
		if users[i].DeptRole == model.DeptRoleAdmin {
			continue
		}
		result = append(result, gin.H{
			"id":         users[i].ID,
			"name":       users[i].Name,
			"student_id": users[i].StudentID,
			"department": users[i].Department,
		})
	}
	utils.Success(c, result)
}

// SetUserDepartment 设置用户部门
func (h *UserHandler) SetUserDepartment(c *gin.Context) {
	// 权限检查 - 需要系统管理员或办公室管理员
	checker := auth.GetChecker(c)
	if !checker.IsAdmin() && !checker.IsOfficeAdmin() {
		auth.ResponseForbidden(c, "无权设置用户部门")
		return
	}

	idStr := c.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error(c, 400, "无效的用户ID")
		return
	}

	var req model.UpdateUserDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	adminID := auth.GetUserIDFromContext(c)
	if err := h.service.SetUserDepartment(adminID, userID, req.Department); err != nil {
		utils.Error(c, 403, err.Error())
		return
	}
	utils.Success(c, nil)
}

// SetUserDeptRole 设置用户部门角色
func (h *UserHandler) SetUserDeptRole(c *gin.Context) {
	// 权限检查 - 需要系统管理员或办公室管理员
	checker := auth.GetChecker(c)
	if !checker.IsAdmin() && !checker.IsOfficeAdmin() {
		auth.ResponseForbidden(c, "无权设置用户部门角色")
		return
	}

	idStr := c.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error(c, 400, "无效的用户ID")
		return
	}

	var req model.UpdateUserDeptRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	adminID := auth.GetUserIDFromContext(c)
	if err := h.service.SetUserDeptRole(adminID, userID, req.DeptRole); err != nil {
		utils.Error(c, 403, err.Error())
		return
	}
	utils.Success(c, nil)
}

// GetUsersByFilter 根据筛选条件获取用户
func (h *UserHandler) GetUsersByFilter(c *gin.Context) {
	// 权限检查
	checker := auth.GetChecker(c)
	if !checker.HasPermission(auth.PermUserManage) {
		auth.ResponseForbidden(c)
		return
	}

	var filter model.UserListFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 非管理员只能查看自己部门的用户
	if !checker.IsAdmin() && !checker.IsOfficeAdmin() {
		filter.Departments = []string{checker.GetDepartment()}
	}

	users, err := h.service.GetUsersByFilter(filter)
	if err != nil {
		utils.Error(c, 500, "获取用户列表失败")
		return
	}

	var result []model.UserInfo
	for _, user := range users {
		result = append(result, model.UserInfo{
			ID:         user.ID,
			StudentID:  user.StudentID,
			Name:       user.Name,
			Email:      user.Email,
			Role:       user.Role,
			Department: user.Department,
			DeptRole:   user.DeptRole,
		})
	}
	utils.Success(c, model.UserListResponse{
		Total: len(result),
		Users: result,
	})
}

// GetDepartments 获取部门列表
func (h *UserHandler) GetDepartments(c *gin.Context) {
	// 登录即可访问
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}

	utils.Success(c, gin.H{
		"departments": model.Departments,
	})
}

// CreateUser 管理员创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	// 权限检查
	checker := auth.GetChecker(c)
	if !checker.HasPermission(auth.PermUserManage) {
		auth.ResponseForbidden(c, "无权创建用户")
		return
	}

	var req model.AdminCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	adminID := auth.GetUserIDFromContext(c)
	user, err := h.service.AdminCreateUser(adminID, &req)
	if err != nil {
		utils.Error(c, 403, err.Error())
		return
	}

	utils.Success(c, model.UserInfo{
		ID:         user.ID,
		StudentID:  user.StudentID,
		Name:       user.Name,
		Email:      user.Email,
		Role:       user.Role,
		Department: user.Department,
		DeptRole:   user.DeptRole,
	})
}

// AdminUpdateUser 管理员更新用户
func (h *UserHandler) AdminUpdateUser(c *gin.Context) {
	// 权限检查
	checker := auth.GetChecker(c)
	if !checker.HasPermission(auth.PermUserManage) {
		auth.ResponseForbidden(c, "无权更新用户")
		return
	}

	idStr := c.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error(c, 400, "无效的用户ID")
		return
	}

	var req model.AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	adminID := auth.GetUserIDFromContext(c)
	if err := h.service.AdminUpdateUser(adminID, userID, &req); err != nil {
		utils.Error(c, 403, err.Error())
		return
	}

	utils.Success(c, nil)
}

// DeleteUser 管理员删除用户（实际为禁用用户）
//
// 注意：此操作将用户标记为禁用（is_active = 0），而非物理删除。
// 保留用户数据是为了维护历史值班记录、审批记录等数据的完整性，
// 确保可以追溯"谁排的班"、"谁值的班"、"谁审批的"等操作历史。
// 已禁用的用户无法登录系统，也不会出现在用户列表中。
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// 权限检查
	checker := auth.GetChecker(c)
	if !checker.HasPermission(auth.PermUserManage) {
		auth.ResponseForbidden(c, "无权删除用户")
		return
	}

	idStr := c.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error(c, 400, "无效的用户ID")
		return
	}

	adminID := auth.GetUserIDFromContext(c)
	if err := h.service.AdminDeleteUser(adminID, userID); err != nil {
		utils.Error(c, 403, err.Error())
		return
	}

	utils.Success(c, nil)
}
