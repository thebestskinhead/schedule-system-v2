package auth

import (
	"schedule-system-v2/backend/internal/model"

	"github.com/gin-gonic/gin"
)

// EffectivePerm 有效权限 = 权限 + 资源范围
type EffectivePerm struct {
	Permission   Permission
	ResourceType string // "all" | "dept" | "user"，空 = 不限制
	ResourceID   int
}

// Checker 权限检查器（唯一授权入口）
// 权限来源：角色权限（RolePermissions）+ 临时权限（temp_permissions 表）
type Checker struct {
	userID     int
	studentID  string
	role       string // 原始 role 字段（admin/user）
	department string
	deptRole   string
	tempPerms  []TempPermInfo // 缓存的临时权限（包含资源范围信息）
}

// NewChecker 创建权限检查器
func NewChecker(userID int, studentID, role, department, deptRole string) *Checker {
	c := &Checker{
		userID:     userID,
		studentID:  studentID,
		role:       role,
		department: department,
		deptRole:   deptRole,
		tempPerms:  loadTempPermissions(userID),
	}
	return c
}

// FromContext 从Gin上下文创建权限检查器
func FromContext(c *gin.Context) *Checker {
	return NewChecker(
		GetUserIDFromContext(c),
		GetStudentIDFromContext(c),
		GetRoleFromContext(c),
		GetDepartmentFromContext(c),
		GetDeptRoleFromContext(c),
	)
}

// GetUserID 获取用户ID
func (c *Checker) GetUserID() int {
	return c.userID
}

// GetStudentID 获取学号
func (c *Checker) GetStudentID() string {
	return c.studentID
}

// GetRole 获取原始角色
func (c *Checker) GetRole() string {
	return c.role
}

// GetDepartment 获取部门
func (c *Checker) GetDepartment() string {
	return c.department
}

// GetDeptRole 获取部门角色
func (c *Checker) GetDeptRole() string {
	return c.deptRole
}

// UnifiedRole 获取统一角色码（admin / dept_admin / dept_member / user）
func (c *Checker) UnifiedRole() string {
	return ResolveRole(c.role, c.department, c.deptRole)
}

// EffectivePermissionsForIdentity 计算指定身份（role / department / dept_role）派生出的权限集（不含临时权限）。
// 身份即策略组：同一身份在任何模块中都得到同一权限集，模块只比对权限集、不关心身份从何而来。
func EffectivePermissionsForIdentity(role, department, deptRole string) []EffectivePerm {
	perms := make([]EffectivePerm, 0, 32)

	if role == model.RoleAdmin {
		// 系统管理员：拥有全部权限（all 范围）
		perms = append(perms, EffectivePerm{Permission: PermSystemAdmin, ResourceType: "all"})
		for _, p := range RolePermissions[model.RoleAdmin] {
			perms = append(perms, EffectivePerm{Permission: p, ResourceType: "all"})
		}
		return perms
	}

	// 角色权限（统一角色）；部门级权限自动限定为该身份所在部门。
	unified := ResolveRole(role, department, deptRole)
	deptID := model.DepartmentID(department)
	for _, p := range RolePermissions[unified] {
		rt := "all"
		rid := 0
		if deptScopedPermissions[p] && (unified == model.RoleDeptAdmin || unified == model.RoleDeptMember) {
			rt = "dept"
			rid = deptID
		}
		perms = append(perms, EffectivePerm{Permission: p, ResourceType: rt, ResourceID: rid})
	}

	// 办公室管理员派生身份：department=办公室 && dept_role=dept_admin → 附加全局管理权限集（all 范围）。
	// office_admin 不是存储角色（users.role 仅允许 admin/user），此判定在部门角色权限基础上叠加全局权限。
	if department == model.DepartmentOffice && deptRole == model.RoleDeptAdmin {
		for _, p := range officeAdminPermissions {
			perms = append(perms, EffectivePerm{Permission: p, ResourceType: "all"})
		}
	}

	return perms
}

// EffectivePermissions 返回用户的完整有效权限集合（身份派生权限 ∪ 临时权限，权限组已展开）
func (c *Checker) EffectivePermissions() []EffectivePerm {
	perms := EffectivePermissionsForIdentity(c.role, c.department, c.deptRole)

	// 临时权限（带自身资源范围）
	for _, tp := range c.tempPerms {
		perms = append(perms, EffectivePerm{
			Permission:   tp.Permission,
			ResourceType: tp.ResourceType,
			ResourceID:   tp.ResourceID,
		})
	}

	return perms
}

// Covers 判断本检查器的有效权限集是否覆盖「以指定范围持有某权限」。
// 这是权限边界（permission boundary）判定：授予或变更权限时，不得超出自身范围。
// 拥有 all 范围的权限覆盖任意范围；dept 范围仅覆盖同一部门。
func (c *Checker) Covers(perm Permission, resourceType string, resourceID int) bool {
	if c.IsAdmin() {
		return true
	}
	for _, p := range c.EffectivePermissions() {
		if p.Permission != perm && !permImplies(p.Permission, perm) {
			continue
		}
		if p.ResourceType == "all" || p.ResourceType == "" {
			return true
		}
		if p.ResourceType == resourceType && (p.ResourceID == 0 || p.ResourceID == resourceID) {
			return true
		}
	}
	return false
}

// CanGrantIdentity 判断本检查器是否可赋予目标身份（role / department / dept_role）。
// 规则：目标身份派生出的权限集必须被操作者自身有效权限集完全覆盖，即
// 「不得授予超出自身范围的权限」——无论对象是自己还是他人。
func (c *Checker) CanGrantIdentity(role, department, deptRole string) bool {
	if c.IsAdmin() {
		return true
	}
	for _, tp := range EffectivePermissionsForIdentity(role, department, deptRole) {
		if !c.Covers(tp.Permission, tp.ResourceType, tp.ResourceID) {
			return false
		}
	}
	return true
}

// HasPermission 检查是否拥有指定权限（不限资源范围）
func (c *Checker) HasPermission(perm Permission) bool {
	if c.IsAdmin() {
		return true
	}
	if c.role == "" && len(c.tempPerms) == 0 {
		return false
	}
	for _, p := range c.EffectivePermissions() {
		if p.Permission == perm || permImplies(p.Permission, perm) {
			return true
		}
	}
	return false
}

// HasPermissionOn 检查是否拥有指定权限且资源范围匹配
func (c *Checker) HasPermissionOn(perm Permission, resourceType string, resourceID int) bool {
	if c.IsAdmin() {
		return true
	}
	if c.role == "" && len(c.tempPerms) == 0 {
		return false
	}
	for _, p := range c.EffectivePermissions() {
		if p.Permission != perm && !permImplies(p.Permission, perm) {
			continue
		}
		if p.ResourceType == "all" || p.ResourceType == "" {
			return true
		}
		if p.ResourceType == resourceType && (resourceID == 0 || p.ResourceID == resourceID) {
			return true
		}
	}
	return false
}

// HasAnyPermission 检查是否有任意一个权限
func (c *Checker) HasAnyPermission(perms ...Permission) bool {
	for _, perm := range perms {
		if c.HasPermission(perm) {
			return true
		}
	}
	return false
}

// HasAllPermissions 检查是否拥有所有权限
func (c *Checker) HasAllPermissions(perms ...Permission) bool {
	for _, perm := range perms {
		if !c.HasPermission(perm) {
			return false
		}
	}
	return true
}

// IsAdmin 检查是否是系统管理员（唯一内建角色短路）
func (c *Checker) IsAdmin() bool {
	return c.role == model.RoleAdmin
}

// IsOwner 检查是否是资源所有者（通过学号）
func (c *Checker) IsOwner(targetStudentID string) bool {
	return c.studentID == targetStudentID
}

// CanAccessDept 检查是否可以访问指定部门（统一部门访问判定）
// 拥有任意"部门级权限"且资源范围覆盖该部门即可访问。
func (c *Checker) CanAccessDept(dept string) bool {
	if c.IsAdmin() {
		return true
	}
	if c.role == "" && len(c.tempPerms) == 0 {
		return false
	}
	deptID := model.DepartmentID(dept)
	for _, p := range c.EffectivePermissions() {
		if !isDeptAccessPermission(p.Permission) {
			continue
		}
		if p.ResourceType == "all" || p.ResourceType == "" {
			return true
		}
		if p.ResourceType == "dept" && (p.ResourceID == 0 || p.ResourceID == deptID) {
			return true
		}
	}
	return false
}

// GetTempPermissions 获取用户的临时权限列表
func (c *Checker) GetTempPermissions() []TempPermInfo {
	return c.tempPerms
}

// GetTempPermissionNames 获取用户的临时权限名称列表
func (c *Checker) GetTempPermissionNames() []Permission {
	perms := make([]Permission, 0, len(c.tempPerms))
	for _, p := range c.tempPerms {
		perms = append(perms, p.Permission)
	}
	return perms
}
