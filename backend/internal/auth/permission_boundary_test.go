package auth

import (
	"testing"

	"schedule-system-v2/backend/internal/model"
)

// 权限边界（permission boundary）单测：授权/改身份不得超出自身范围。
// 直接构造 Checker（同包可见未导出字段），避免依赖数据库。

func TestCovers_AdminCoversEverything(t *testing.T) {
	c := &Checker{userID: 1, studentID: "s1", role: model.RoleAdmin, department: model.DepartmentOffice, deptRole: model.RoleDeptAdmin}
	if !c.Covers(PermUserManageAll, "all", 0) {
		t.Fatal("系统管理员应覆盖任意权限")
	}
	if !c.Covers(PermScheduleManageDept, "dept", model.DepartmentID("竞赛部")) {
		t.Fatal("系统管理员应覆盖任意部门范围")
	}
}

func TestCovers_DepartmentScopedOnlyOwnDept(t *testing.T) {
	// 部门管理员：拥有 schedule:manage:dept（dept 范围，本部门=竞赛部）
	c := &Checker{userID: 2, studentID: "s2", role: model.RoleUser, department: "竞赛部", deptRole: model.RoleDeptAdmin}

	// 覆盖本部门的 schedule:manage:dept
	if !c.Covers(PermScheduleManageDept, "dept", model.DepartmentID("竞赛部")) {
		t.Fatal("部门管理员应覆盖本部门的 schedule:manage:dept")
	}
	// 不覆盖其他部门
	if c.Covers(PermScheduleManageDept, "dept", model.DepartmentID("项目部")) {
		t.Fatal("部门管理员不得覆盖其他部门范围")
	}
	// 不覆盖全局范围
	if c.Covers(PermScheduleManageAll, "all", 0) {
		t.Fatal("部门管理员不得覆盖全局权限")
	}
}

func TestCovers_TempPermIsBounded(t *testing.T) {
	// 普通成员 + 临时 user:manage:all（all 范围）
	c := &Checker{
		userID:     3,
		studentID:  "s3",
		role:       model.RoleUser,
		department: "竞赛部",
		deptRole:   model.RoleDeptMember,
		tempPerms: []TempPermInfo{
			{Permission: PermUserManageAll, ResourceType: "all", ResourceID: 0},
		},
	}
	// 覆盖 user:manage:all
	if !c.Covers(PermUserManageAll, "all", 0) {
		t.Fatal("持临时 user:manage:all 应覆盖该权限")
	}
	// 但不覆盖排班管理
	if c.Covers(PermScheduleManageAll, "all", 0) {
		t.Fatal("持临时 user:manage:all 不得覆盖排班全局权限")
	}
}

func TestCovers_NoPermNothingCovered(t *testing.T) {
	// 普通成员，无任何管理权限
	c := &Checker{
		userID:     4,
		studentID:  "s4",
		role:       model.RoleUser,
		department: "竞赛部",
		deptRole:   model.RoleDeptMember,
	}
	if c.Covers(PermUserManage, "all", 0) {
		t.Fatal("普通成员不得覆盖用户管理权限")
	}
	if c.Covers(PermSchedulePublish, "all", 0) {
		t.Fatal("普通成员不得覆盖排班发布权限")
	}
}

func TestCanGrantIdentity_SelfEscalationBlocked(t *testing.T) {
	// 持有临时 user:manage:all 的普通成员
	c := &Checker{
		userID:     5,
		studentID:  "s5",
		role:       model.RoleUser,
		department: "竞赛部",
		deptRole:   model.RoleDeptMember,
		tempPerms: []TempPermInfo{
			{Permission: PermUserManageAll, ResourceType: "all", ResourceID: 0},
		},
	}

	// 试图把自己改成 办公室 + dept_admin → 派生 officeAdminPermissions（含排班全局），应被拒
	if c.CanGrantIdentity(model.RoleUser, model.DepartmentOffice, model.RoleDeptAdmin) {
		t.Fatal("不得将身份变更为超出自身范围的（办公室 + dept_admin）")
	}
	// 试图把自己改成 系统管理员，应被拒
	if c.CanGrantIdentity(model.RoleAdmin, "竞赛部", model.RoleDeptMember) {
		t.Fatal("不得将身份变更为系统管理员")
	}
	// 保持普通成员身份不变 → 允许（权限集未扩大）
	if !c.CanGrantIdentity(model.RoleUser, "竞赛部", model.RoleDeptMember) {
		t.Fatal("保持原身份应被允许")
	}
}

func TestCanGrantIdentity_SysAdminCanGrantAny(t *testing.T) {
	c := &Checker{userID: 1, studentID: "s1", role: model.RoleAdmin, department: model.DepartmentOffice, deptRole: model.RoleDeptAdmin}
	if !c.CanGrantIdentity(model.RoleUser, model.DepartmentOffice, model.RoleDeptAdmin) {
		t.Fatal("系统管理员应可授予任意身份")
	}
}

func TestCanGrantIdentity_OfficeAdminCannotCreateSysAdmin(t *testing.T) {
	// 办公室管理员：拥有全部管理权限，但无 PermUserSetRole（不含在 officeAdminPermissions）
	c := &Checker{userID: 6, studentID: "s6", role: model.RoleUser, department: model.DepartmentOffice, deptRole: model.RoleDeptAdmin}
	if c.CanGrantIdentity(model.RoleAdmin, "竞赛部", model.RoleDeptMember) {
		t.Fatal("办公室管理员不得授予系统管理员身份")
	}
	// 但可以授予办公室 + dept_admin（与自身同级）
	if !c.CanGrantIdentity(model.RoleUser, model.DepartmentOffice, model.RoleDeptAdmin) {
		t.Fatal("办公室管理员应可授予同级（办公室 + dept_admin）")
	}
}

func TestGrantResourceScope(t *testing.T) {
	if rt, rid := GrantResourceScope(PermScheduleManageDept, "竞赛部"); rt != "dept" || rid != model.DepartmentID("竞赛部") {
		t.Fatalf("部门级权限应落 dept + 部门ID，得到 %s/%d", rt, rid)
	}
	if rt, rid := GrantResourceScope(PermUserManageAll, "竞赛部"); rt != "all" || rid != 0 {
		t.Fatalf("全局权限应落 all/0，得到 %s/%d", rt, rid)
	}
}
