package model

import "time"

// User 用户模型
type User struct {
	ID         int       `json:"id" db:"id"`
	StudentID  string    `json:"student_id" db:"student_id"`
	Name       string    `json:"name" db:"name"`
	Email      string    `json:"email" db:"email"`
	Password   string    `json:"-" db:"password"`
	Role       string    `json:"role" db:"role"`
	Department string    `json:"department" db:"department"`
	DeptRole   string    `json:"dept_role" db:"dept_role"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// UserRole 用户角色常量
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// DeptRole 部门角色常量
const (
	DeptRoleAdmin  = "dept_admin"
	DeptRoleMember = "dept_member"
)

// Departments 可选部门列表
// 部门对应的资源ID：办公室=1, 竞赛部=2, 项目部=3, 科普部=4
var Departments = []string{"办公室", "竞赛部", "项目部", "科普部"}

// RegisterRequest 注册请求
type RegisterRequest struct {
	StudentID  string `json:"student_id" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	Department string `json:"department" binding:"required"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	StudentID string `json:"student_id"`
	Email     string `json:"email"`
	Password  string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string   `json:"token"`
	User  UserInfo `json:"user"`
}

// UserInfo 用户信息（不含密码）
type UserInfo struct {
	ID         int    `json:"id"`
	StudentID  string `json:"student_id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	Department string `json:"department"`
	DeptRole   string `json:"dept_role"`
}

// UpdateUserRequest 更新用户信息请求
type UpdateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

// SetRoleRequest 设置角色请求
type SetRoleRequest struct {
	UserID int    `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required"`
}



// SystemStatus 系统状态
type SystemStatus struct {
	Initialized bool `json:"initialized"`
}
