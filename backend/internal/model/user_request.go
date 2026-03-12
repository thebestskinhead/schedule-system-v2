package model

// SetDepartmentRequest 设置部门请求
type SetDepartmentRequest struct {
	UserID     int    `json:"user_id" binding:"required"`
	Department string `json:"department" binding:"required"`
}

// SetDeptRoleRequest 设置部门角色请求
type SetDeptRoleRequest struct {
	UserID   int    `json:"user_id" binding:"required"`
	DeptRole string `json:"dept_role" binding:"required,oneof=dept_admin dept_member"`
}

// UserListByDeptRequest 按部门获取用户请求
type UserListByDeptRequest struct {
	Department string `form:"department" binding:"required"`
}

// UserListFilter 用户列表筛选
type UserListFilter struct {
	Departments []string `form:"departments"` // 部门列表，空表示全部
	Role        string   `form:"role"`        // 系统角色
	DeptRole    string   `form:"dept_role"`   // 部门角色
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Total int        `json:"total"`
	Users []UserInfo `json:"users"`
}

// UpdateUserDepartmentRequest 更新用户部门请求
type UpdateUserDepartmentRequest struct {
	Department string `json:"department" binding:"required"`
}

// UpdateUserDeptRoleRequest 更新用户部门角色请求
type UpdateUserDeptRoleRequest struct {
	DeptRole string `json:"dept_role" binding:"required,oneof=dept_admin dept_member"`
}

// AdminCreateUserRequest 管理员创建用户请求
type AdminCreateUserRequest struct {
	StudentID  string `json:"student_id" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Department string `json:"department" binding:"required"`
	Role       string `json:"role" binding:"required,oneof=user admin"`
	DeptRole   string `json:"dept_role" binding:"required,oneof=dept_admin dept_member"`
	Password   string `json:"password" binding:"omitempty,min=6"`
}

// AdminUpdateUserRequest 管理员更新用户请求
type AdminUpdateUserRequest struct {
	Name       string `json:"name" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Department string `json:"department" binding:"required"`
	Role       string `json:"role" binding:"required,oneof=user admin"`
	DeptRole   string `json:"dept_role" binding:"required,oneof=dept_admin dept_member"`
}
