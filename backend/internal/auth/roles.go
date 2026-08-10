package auth

import "schedule-system-v2/backend/internal/model"

// RolePermissions 统一角色 → 权限映射（单一事实来源）
// 键为统一角色码（model.RoleAdmin / RoleOfficeAdmin / RoleDeptAdmin / RoleDeptMember / RoleUser）
// 系统管理员（RoleAdmin）由 Checker 短路处理（拥有全部权限），此处仍列出以便展示/参考。
var RolePermissions = map[string][]Permission{
	model.RoleAdmin: {
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
	},
	model.RoleOfficeAdmin: {
		// 用户管理权限组（不含设置系统角色）
		PermUserProfile, PermUserManage, PermUserManageAll, PermUserManageDept, PermUserView, PermUserEdit,
		// 无课表权限
		PermAvailabilityView, PermAvailabilityEdit, PermAvailabilityImport, PermAvailabilityViewAll,
		// 排班权限组（全部）
		PermScheduleView, PermSchedulePreview, PermScheduleConfirm, PermScheduleEdit, PermSchedulePublish,
		PermScheduleSettings, PermScheduleExport, PermScheduleViewAll, PermScheduleViewDept, PermScheduleManageDept, PermScheduleManageAll,
		// 值班权限
		PermDutyView, PermDutyUpdate,
		// 模板权限
		PermTemplateView, PermTemplateEdit, PermTemplateManage,
	},
	model.RoleDeptAdmin: {
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
	},
	model.RoleDeptMember: {
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
	},
	model.RoleUser: {
		// 个人基础权限
		PermUserProfile, PermUserEdit,
		PermAvailabilityEdit, PermAvailabilityView,
		PermDutyView,
	},
}

// deptScopedPermissions 部门级权限集合
// 这些权限在授予部门角色（dept_admin/dept_member）时，自动限定资源范围为该用户所在部门。
var deptScopedPermissions = map[Permission]bool{
	PermScheduleViewDept:   true,
	PermScheduleManageDept: true,
	PermUserManageDept:     true,
	PermUserView:           true,
}

// deptAccessPermissions 部门访问权限集合
// 拥有这些权限且资源范围覆盖目标部门时，视为可以访问该部门数据。
var deptAccessPermissions = map[Permission]bool{
	PermScheduleViewDept:    true,
	PermScheduleManageDept:  true,
	PermScheduleViewAll:     true,
	PermScheduleManageAll:   true,
	PermUserManageDept:      true,
	PermUserView:            true,
	PermUserManage:          true,
	PermUserManageAll:       true,
	PermAvailabilityViewAll: true,
}

// isDeptAccessPermission 判断权限是否属于部门访问权限
func isDeptAccessPermission(p Permission) bool {
	return deptAccessPermissions[p]
}

// ResolveRole 将用户的 (role, department, deptRole) 解析为统一角色码。
// 兼容旧数据模型：role=admin 视为系统管理员，其余由 dept_role + department 推导。
func ResolveRole(role, department, deptRole string) string {
	if role == model.RoleAdmin {
		return model.RoleAdmin
	}
	if department == model.DepartmentOffice && deptRole == model.RoleDeptAdmin {
		return model.RoleOfficeAdmin
	}
	if deptRole == model.RoleDeptAdmin {
		return model.RoleDeptAdmin
	}
	if deptRole == model.RoleDeptMember {
		return model.RoleDeptMember
	}
	return model.RoleUser
}
