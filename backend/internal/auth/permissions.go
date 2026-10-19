package auth

import "schedule-system-v2/backend/internal/model"

// Permission 权限类型（统一使用 model.Permission）
type Permission = model.Permission

// 权限常量（重新导出 model 包的权限常量，便于使用）
const (
	// 系统级权限
	PermSystemAdmin  Permission = "system:admin"  // 超级管理员（临时权限清理等）
	PermSystemManage Permission = "system:manage" // 系统管理（SMTP 服务设置、系统更新）

	// 用户相关权限
	PermUserProfile    Permission = "user:profile"
	PermUserManage     Permission = "user:manage"
	PermUserManageAll  Permission = "user:manage:all"
	PermUserManageDept Permission = "user:manage:dept"
	PermUserSetRole    Permission = "user:set_role"
	PermUserView       Permission = "user:view"
	PermUserEdit       Permission = "user:edit"

	// 无课表相关权限
	PermAvailabilityView    Permission = "availability:view"
	PermAvailabilityEdit    Permission = "availability:edit"
	PermAvailabilityImport  Permission = "availability:import"
	PermAvailabilityViewAll Permission = "availability:view_all"

	// 排班相关权限
	PermScheduleView       Permission = "schedule:view"
	PermSchedulePreview    Permission = "schedule:preview"
	PermScheduleConfirm    Permission = "schedule:confirm"
	PermScheduleEdit       Permission = "schedule:edit"
	PermSchedulePublish    Permission = "schedule:publish"
	PermScheduleSettings   Permission = "schedule:settings"
	PermScheduleExport     Permission = "schedule:export"
	PermScheduleViewAll    Permission = "schedule:view:all"
	PermScheduleViewDept   Permission = "schedule:view:dept"
	PermScheduleManageDept Permission = "schedule:manage:dept"
	PermScheduleManageAll  Permission = "schedule:manage:all" // 全部排班管理（简化权限组）

	// 值班相关权限
	PermDutyView   Permission = "duty:view"
	PermDutyUpdate Permission = "duty:update"

	// 模板相关权限
	PermTemplateView   Permission = "template:view"
	PermTemplateEdit   Permission = "template:edit"
	PermTemplateManage Permission = "template:manage"
)

// permissionHierarchy 权限层级映射（权限组包含的子权限）
// 注意：层级展开是传递的，例如 schedule:manage:all → schedule:manage:dept → schedule:preview 等
var permissionHierarchy = map[Permission][]Permission{
	// 排班管理（全部）包含所有排班相关权限
	PermScheduleManageAll: {
		PermScheduleView, PermSchedulePreview, PermScheduleConfirm, PermScheduleEdit,
		PermSchedulePublish, PermScheduleSettings, PermScheduleExport,
		PermScheduleViewAll, PermScheduleViewDept, PermScheduleManageDept,
	},
	// 用户管理（全部）包含用户查看、编辑、部门管理（不含设置系统角色）
	PermUserManageAll: {
		PermUserManage, PermUserManageDept, PermUserView, PermUserEdit,
	},
	// 部门排班管理包含部门排班相关权限（包括设置）
	PermScheduleManageDept: {
		PermScheduleView, PermScheduleViewDept, PermSchedulePreview, PermScheduleConfirm, PermScheduleEdit, PermScheduleSettings,
	},
	// 部门用户管理包含部门用户相关权限
	PermUserManageDept: {
		PermUserView, PermUserEdit,
	},
}

// deptScopedGrantPermissions 授出时需限定到被授权人所在部门的权限。
// 其余可授权限（排班/用户管理的全局项）资源范围为 all。
var deptScopedGrantPermissions = map[Permission]bool{
	PermScheduleManageDept: true,
	PermUserManageDept:     true,
}

// GrantResourceScope 计算授予指定权限时应落库的资源范围。
// 部门级权限限定在被授权人所在部门；全局权限为 all。
func GrantResourceScope(perm Permission, targetDepartment string) (string, int) {
	if deptScopedGrantPermissions[perm] {
		return "dept", model.DepartmentID(targetDepartment)
	}
	return "all", 0
}

// permImplies 判断 granted 权限是否隐含 required 权限（含权限组传递展开）
func permImplies(granted, required Permission) bool {
	if granted == required {
		return true
	}
	subs, ok := permissionHierarchy[granted]
	if !ok {
		return false
	}
	for _, sub := range subs {
		if permImplies(sub, required) {
			return true
		}
	}
	return false
}
