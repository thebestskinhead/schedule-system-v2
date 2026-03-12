package model

import "time"

// Permission 权限常量定义
type Permission string

const (
	// 系统级权限
	PermSystemAdmin Permission = "system:admin" // 系统管理

	// 用户管理权限
	PermUserProfile    Permission = "user:profile"    // 查看/修改个人信息
	PermUserManage     Permission = "user:manage"     // 用户管理（管理员）
	PermUserManageAll  Permission = "user:manage:all" // 全部用户管理
	PermUserManageDept Permission = "user:manage:dept" // 部门用户管理
	PermUserSetRole    Permission = "user:set_role"   // 设置用户角色
	PermUserView       Permission = "user:view"       // 查看用户信息
	PermUserEdit       Permission = "user:edit"       // 编辑自己信息

	// 排班相关权限
	PermScheduleView       Permission = "schedule:view"        // 查看排班
	PermSchedulePreview    Permission = "schedule:preview"     // 预览排班
	PermScheduleConfirm    Permission = "schedule:confirm"     // 确认排班
	PermScheduleEdit       Permission = "schedule:edit"        // 编辑排班
	PermSchedulePublish    Permission = "schedule:publish"     // 发布每周分工
	PermScheduleSettings   Permission = "schedule:settings"    // 排班设置
	PermScheduleExport     Permission = "schedule:export"      // 导出排班
	PermScheduleViewAll    Permission = "schedule:view:all"    // 查看全部分工
	PermScheduleViewDept   Permission = "schedule:view:dept"   // 查看部门排班
	PermScheduleManageDept Permission = "schedule:manage:dept" // 部门排班管理
	PermScheduleManageAll  Permission = "schedule:manage:all"  // 全部排班管理（简化权限组）

	// 无课表权限
	PermAvailabilityEdit    Permission = "availability:edit"      // 编辑自己无课表
	PermAvailabilityView    Permission = "availability:view"      // 查看无课表
	PermAvailabilityImport  Permission = "availability:import"    // 导入无课表
	PermAvailabilityViewAll Permission = "availability:view_all"  // 查看所有无课表（管理员）

	// 值班相关权限
	PermDutyView   Permission = "duty:view"   // 查看值班
	PermDutyUpdate Permission = "duty:update" // 更新值班状态

	// 模板管理权限
	PermTemplateManage Permission = "template:manage" // 模板管理
	PermTemplateView   Permission = "template:view"   // 查看模板
	PermTemplateEdit   Permission = "template:edit"   // 编辑模板
)

// PermissionGroup 权限组
type PermissionGroup struct {
	Code        string       `json:"code" db:"code"`
	Name        string       `json:"name" db:"name"`
	Description string       `json:"description" db:"description"`
	Permissions []Permission `json:"permissions"`
}

// Role 角色定义
type Role struct {
	ID          int       `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	RoleType    string    `json:"role_type" db:"role_type"` // system, dept, custom
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// RolePermission 角色权限关联
type RolePermission struct {
	RoleID       int        `json:"role_id" db:"role_id"`
	Permission   Permission `json:"permission" db:"permission"`
	ResourceType string     `json:"resource_type,omitempty" db:"resource_type"` // all, dept, user
	ResourceID   int        `json:"resource_id,omitempty" db:"resource_id"`     // 部门ID或用户ID
}

// UserRole 用户角色关联
type UserRole struct {
	ID        int        `json:"id" db:"id"`
	UserID    int        `json:"user_id" db:"user_id"`
	RoleID    int        `json:"role_id" db:"role_id"`
	DeptID    *int       `json:"dept_id,omitempty" db:"dept_id"`       // 部门角色时指定部门
	IsActive  bool       `json:"is_active" db:"is_active"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" db:"expires_at"` // 临时授权过期时间
}

// UserPermissionTemp 用户临时权限
type UserPermissionTemp struct {
	ID           int       `json:"id" db:"id"`
	UserID       int       `json:"user_id" db:"user_id"`
	Permission   Permission `json:"permission" db:"permission"`
	ResourceType string    `json:"resource_type" db:"resource_type"`
	ResourceID   int       `json:"resource_id" db:"resource_id"`
	GrantedBy    int       `json:"granted_by" db:"granted_by"`
	Reason       string    `json:"reason" db:"reason"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	IsActive     bool      `json:"is_active" db:"is_active"`
}

// PermissionCheck 权限检查请求
type PermissionCheck struct {
	UserID       int        `json:"user_id"`
	Permission   Permission `json:"permission"`
	ResourceType string     `json:"resource_type,omitempty"`
	ResourceID   int        `json:"resource_id,omitempty"`
}

// RoleConstant 系统内置角色常量
const (
	RoleOfficeAdmin = "office_admin" // 办公室管理员
	RoleDeptAdmin   = "dept_admin"   // 部门管理员
	RoleDeptMember  = "dept_member"  // 部门成员
	RoleCustom      = "custom"       // 自定义角色
)

// GetSystemRoles 获取系统预定义角色
func GetSystemRoles() []Role {
	return []Role{
		{
			Code:        RoleOfficeAdmin,
			Name:        "办公室管理员",
			Description: "拥有系统全部权限",
			RoleType:    "system",
			IsActive:    true,
		},
		{
			Code:        RoleDeptAdmin,
			Name:        "部门管理员",
			Description: "管理部门内用户和排班",
			RoleType:    "dept",
			IsActive:    true,
		},
		{
			Code:        RoleDeptMember,
			Name:        "部门成员",
			Description: "部门普通成员",
			RoleType:    "dept",
			IsActive:    true,
		},
	}
}

// GetRoleDefaultPermissions 获取角色默认权限
func GetRoleDefaultPermissions(roleCode string) []Permission {
	switch roleCode {
	case RoleOfficeAdmin:
		return []Permission{
			PermSystemAdmin,
			// 用户管理权限组
			PermUserProfile, PermUserManage, PermUserManageAll, PermUserManageDept, PermUserSetRole, PermUserView, PermUserEdit,
			// 无课表权限
			PermAvailabilityView, PermAvailabilityEdit, PermAvailabilityImport, PermAvailabilityViewAll,
			// 排班权限组（全部）
			PermScheduleView, PermSchedulePreview, PermScheduleConfirm, PermScheduleEdit, PermSchedulePublish,
			PermScheduleSettings, PermScheduleExport, PermScheduleViewAll, PermScheduleViewDept, PermScheduleManageDept, PermScheduleManageAll,
			// 值班权限
			PermDutyView, PermDutyUpdate,
			// 模板权限
			PermTemplateView, PermTemplateEdit, PermTemplateManage,
		}
	case RoleDeptAdmin:
		return []Permission{
			// 个人权限
			PermUserProfile, PermUserEdit,
			// 部门用户管理权限
			PermUserManageDept, PermUserView,
			// 无课表个人权限
			PermAvailabilityView, PermAvailabilityEdit, PermAvailabilityImport,
			// 排班权限组（部门）
			PermScheduleView, PermScheduleViewDept, PermScheduleManageDept,
			// 值班权限
			PermDutyView, PermDutyUpdate,
			// 模板查看
			PermTemplateView,
		}
	case RoleDeptMember:
		return []Permission{
			// 个人权限
			PermUserProfile, PermUserEdit,
			// 无课表个人权限
			PermAvailabilityView, PermAvailabilityEdit, PermAvailabilityImport,
			// 排班查看权限
			PermScheduleView, PermScheduleViewDept,
			// 值班权限
			PermDutyView, PermDutyUpdate,
			// 模板查看
			PermTemplateView,
		}
	default:
		return []Permission{
			// 个人基础权限
			PermUserProfile, PermUserEdit,
			PermAvailabilityEdit, PermAvailabilityView,
			PermDutyView,
		}
	}
}
