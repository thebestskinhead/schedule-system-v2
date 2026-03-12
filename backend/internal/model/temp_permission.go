package model

import "time"

// GrantPermissionRequest 批量授权请求
type GrantPermissionRequest struct {
	UserIDs      []int      `json:"user_ids" binding:"required,min=1"`
	Permission   Permission `json:"permission" binding:"required"`
	ResourceType string     `json:"resource_type" binding:"required,oneof=all dept user"`
	ResourceID   int        `json:"resource_id"`
	ExpiresAt    time.Time  `json:"expires_at" binding:"required"`
	Reason       string     `json:"reason"`
}

// SingleGrantRequest 单个授权请求（内部使用）
type SingleGrantRequest struct {
	UserID       int
	Permission   Permission
	ResourceType string
	ResourceID   int
	ExpiresAt    time.Time
	Reason       string
}

// RevokePermissionRequest 撤销权限请求
type RevokePermissionRequest struct {
	PermissionID int `json:"permission_id" binding:"required"`
}

// TempPermissionView 临时权限视图
type TempPermissionView struct {
	ID           int        `json:"id"`
	UserID       int        `json:"user_id"`
	UserName     string     `json:"user_name"`
	Permission   Permission `json:"permission"`
	PermissionName string   `json:"permission_name"`
	ResourceType string     `json:"resource_type"`
	ResourceID   int        `json:"resource_id"`
	ResourceName string     `json:"resource_name"`
	GrantedBy    int        `json:"granted_by"`
	GrantedByName string    `json:"granted_by_name"`
	Reason       string     `json:"reason"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    time.Time  `json:"expires_at"`
	IsActive     bool       `json:"is_active"`
	IsExpired    bool       `json:"is_expired"`
}

// MyPermissionView 我的权限视图
type MyPermissionView struct {
	ID           int        `json:"id"`
	Permission   Permission `json:"permission"`
	PermissionName string   `json:"permission_name"`
	ResourceType string     `json:"resource_type"`
	ResourceName string     `json:"resource_name"`
	Reason       string     `json:"reason"`
	ExpiresAt    time.Time  `json:"expires_at"`
	DaysLeft     int        `json:"days_left"`
}

// PermissionInfo 权限信息
type PermissionInfo struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetPermissionList 获取可授权的权限列表（简化版 - 只包含可被临时授权的核心权限）
// 说明：个人权限（如个人资料、编辑自己的无课表等）属于用户本人，不可被授权
func GetPermissionList() []PermissionInfo {
	return []PermissionInfo{
		// 全局权限 - 设置各部门本周值班分工
		{
			Code:        string(PermSchedulePublish),
			Name:        "设置每周分工",
			Description: "设置各部门在本周值班的日期安排（全局权限）",
		},
		// 全部权限组 - 排班管理全部权限
		{
			Code:        string(PermScheduleManageAll),
			Name:        "排班管理（全部）",
			Description: "管理所有部门的排班（预览、确认、编辑、设置、导出、查看全部）",
		},
		// 全部权限组 - 用户管理全部权限（不含设置系统角色，系统角色只能由系统管理员设置）
		{
			Code:        string(PermUserManageAll),
			Name:        "用户管理（全部）",
			Description: "管理所有用户（查看、编辑、设置部门，不含设置系统角色）",
		},
		// 部门权限组 - 部门排班管理权限
		{
			Code:        string(PermScheduleManageDept),
			Name:        "排班管理（部门）",
			Description: "管理部门内排班（预览、确认、编辑、查看部门排班）",
		},
		// 部门权限组 - 部门用户管理权限
		{
			Code:        string(PermUserManageDept),
			Name:        "用户管理（部门）",
			Description: "管理部门内用户（查看、编辑部门成员信息）",
		},
	}
}

// GetPermissionName 获取权限名称
func GetPermissionName(perm Permission) string {
	// 首先检查 GetPermissionList 中的权限
	for _, info := range GetPermissionList() {
		if info.Code == string(perm) {
			return info.Name
		}
	}
	
	// 支持下划线格式的权限名称（用于临时权限申请）
	switch string(perm) {
	case "duty_manage":
		return "值班管理"
	case "user_manage":
		return "用户管理"
	case "schedule_manage":
		return "排班管理"
	case "crawler_manage":
		return "爬虫管理"
	case "system_manage":
		return "系统管理"
	case "temp_permission_grant":
		return "授权管理"
	}
	
	return string(perm)
}
