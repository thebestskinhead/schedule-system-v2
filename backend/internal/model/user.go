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

// UserRole 角色常量（单一角色体系）
// users.role 字段仅允许 admin/user：admin 为系统管理员（唯一特殊身份，系统引导创建），user 为普通用户。
// users.dept_role 字段仅允许 dept_admin/dept_member。
// "办公室管理员"（office_admin）不是存储角色，而是派生身份：department=办公室 && dept_role=dept_admin，
// 其权限由 auth 包 EffectivePermissions 按该派生规则附加（见 auth/roles.go officeAdminPermissions），
// 禁止把 office_admin 写入 users.role。
const (
	RoleAdmin      = "admin"       // 系统管理员（拥有全部权限；应为系统引导创建的唯一特殊身份）
	RoleDeptAdmin  = "dept_admin"  // 部门管理员
	RoleDeptMember = "dept_member" // 部门成员
	RoleUser       = "user"        // 普通用户
)

// DepartmentOffice 办公室部门名（办公室管理员所在部门）
const DepartmentOffice = "办公室"

// Departments 可选部门列表
// 部门对应的资源ID：办公室=1, 竞赛部=2, 项目部=3, 科普部=4
var Departments = []string{DepartmentOffice, "竞赛部", "项目部", "科普部"}

// DepartmentID 获取部门对应的资源ID（用于资源范围判定）
func DepartmentID(name string) int {
	switch name {
	case DepartmentOffice:
		return 1
	case "竞赛部":
		return 2
	case "项目部":
		return 3
	case "科普部":
		return 4
	default:
		return 0
	}
}

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
