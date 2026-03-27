// Package auth 提供权限检查和鉴权功能
package auth

import (
	"net/http"
	"schedule-system-v2/backend/internal/dao"
	"schedule-system-v2/backend/internal/model"

	"github.com/gin-gonic/gin"
)

// Permission 权限类型（统一使用 model.Permission）
type Permission = model.Permission

// 重新导出 model 包的权限常量，便于使用
const (
	// 系统级权限
	PermSystemAdmin Permission = "system:admin"

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

// RolePermissions 角色权限映射（基于 role + dept_role）
// 简化后的权限体系：
// - 系统管理员：拥有所有权限
// - 办公室管理员：拥有除系统管理外的所有权限（通过部门="办公室" + dept_role="admin"识别）
// - 部门管理员：拥有部门排班管理 + 部门用户管理权限
// - 普通成员：拥有个人权限（查看自己的排班、编辑自己的无课表等）
var RolePermissions = map[string]map[string][]Permission{
	// 系统管理员
	model.RoleAdmin: {
		model.DeptRoleAdmin: {
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
		model.DeptRoleMember: {
			PermSystemAdmin,
			PermUserProfile, PermUserManage, PermUserManageAll, PermUserManageDept, PermUserSetRole, PermUserView, PermUserEdit,
			PermAvailabilityView, PermAvailabilityEdit, PermAvailabilityImport, PermAvailabilityViewAll,
			PermScheduleView, PermSchedulePreview, PermScheduleConfirm, PermScheduleEdit, PermSchedulePublish,
			PermScheduleSettings, PermScheduleExport, PermScheduleViewAll, PermScheduleViewDept, PermScheduleManageDept, PermScheduleManageAll,
			PermDutyView, PermDutyUpdate,
			PermTemplateView, PermTemplateEdit, PermTemplateManage,
		},
	},
	// 普通用户
	model.RoleUser: {
		// 部门管理员（拥有部门管理权限）
		model.DeptRoleAdmin: {
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
		// 部门成员（只有个人权限）
		model.DeptRoleMember: {
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
	},
}

// PathPermissionMap API路径到权限的映射
var PathPermissionMap = map[string]Permission{
	// 用户相关
	"/api/v1/user/profile":     PermUserProfile,
	"/api/v1/admin/users":      PermUserManage,
	"/api/v1/admin/users/role": PermUserSetRole,

	// 无课表相关
	"/api/v1/availability":               PermAvailabilityEdit,
	"/api/v1/availability/import/cookie": PermAvailabilityImport,
	"/api/v1/availability/import/xls":          PermAvailabilityImport,
	"/api/v1/availability/import/xls-base64":   PermAvailabilityImport,
	"/api/v1/admin/availability/all":     PermAvailabilityViewAll,

	// 排班相关
	"/api/v1/schedule":               PermScheduleView,
	"/api/v1/admin/schedule/preview": PermSchedulePreview,
	"/api/v1/admin/schedule/confirm": PermScheduleConfirm,
	"/api/v1/admin/schedule/update":  PermScheduleEdit,
	"/api/v1/admin/schedule/settings": PermScheduleSettings,
	"/api/v1/admin/schedule/export":  PermScheduleExport,

	// 模板相关
	"/api/v1/admin/templates":              PermTemplateEdit,
	"/api/v1/admin/templates/placeholders": PermTemplateView,
}

// TempPermInfo 临时权限完整信息
type TempPermInfo struct {
	Permission   Permission
	ResourceType string // all, dept, user
	ResourceID   int
}

// Checker 权限检查器
type Checker struct {
	userID     int
	studentID  string
	role       string
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
		tempPerms:  make([]TempPermInfo, 0),
	}
	// 加载临时权限
	c.loadTempPermissions()
	return c
}

// FromContext 从Gin上下文创建权限检查器
func FromContext(c *gin.Context) *Checker {
	userID, _ := c.Get("userID")
	studentID, _ := c.Get("studentID")
	role, _ := c.Get("role")
	department, _ := c.Get("department")
	deptRole, _ := c.Get("deptRole")

	uid, _ := userID.(int)
	if fuid, ok := userID.(float64); ok {
		uid = int(fuid)
	}
	sid, _ := studentID.(string)
	r, _ := role.(string)
	dept, _ := department.(string)
	dr, _ := deptRole.(string)

	return NewChecker(uid, sid, r, dept, dr)
}

// loadTempPermissions 从数据库加载用户的有效临时权限
func (c *Checker) loadTempPermissions() {
	if c.userID == 0 {
		return
	}

	tempPermDAO := dao.NewTempPermissionDAO()
	perms, err := tempPermDAO.GetActiveByUserID(c.userID)
	if err != nil {
		return
	}

	c.tempPerms = make([]TempPermInfo, 0, len(perms))
	for _, perm := range perms {
		c.tempPerms = append(c.tempPerms, TempPermInfo{
			Permission:   perm.Permission,
			ResourceType: perm.ResourceType,
			ResourceID:   perm.ResourceID,
		})
	}
}

// GetUserID 获取用户ID
func (c *Checker) GetUserID() int {
	return c.userID
}

// GetStudentID 获取学号
func (c *Checker) GetStudentID() string {
	return c.studentID
}

// GetRole 获取角色
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

// permissionHierarchy 权限层级映射（权限组包含的子权限）
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

// HasPermission 检查是否有指定权限
func (c *Checker) HasPermission(perm Permission) bool {
	if c.role == "" {
		return false
	}

	// 1. 检查系统管理员（拥有所有权限）
	if c.role == model.RoleAdmin {
		return true
	}

	// 2. 检查角色权限（包括权限组隐式包含的子权限）
	if c.hasRolePermission(perm) {
		return true
	}

	// 3. 检查角色是否拥有包含该权限的权限组
	if c.hasPermissionGroup(perm) {
		return true
	}

	// 4. 检查临时权限
	if c.hasTempPermission(perm) {
		return true
	}

	// 5. 检查临时权限是否拥有包含该权限的权限组
	if c.hasTempPermissionGroup(perm) {
		return true
	}

	return false
}

// hasPermissionGroup 检查角色是否拥有包含指定权限的权限组
func (c *Checker) hasPermissionGroup(perm Permission) bool {
	for groupPerm, subPerms := range permissionHierarchy {
		// 检查角色是否拥有该权限组
		if c.hasRolePermission(groupPerm) {
			// 检查该权限组是否包含目标权限
			for _, subPerm := range subPerms {
				if subPerm == perm {
					return true
				}
			}
		}
	}
	return false
}

// hasTempPermissionGroup 检查临时权限是否拥有包含指定权限的权限组
func (c *Checker) hasTempPermissionGroup(perm Permission) bool {
	for groupPerm, subPerms := range permissionHierarchy {
		// 检查临时权限是否拥有该权限组
		if c.hasTempPermission(groupPerm) {
			// 检查该权限组是否包含目标权限
			for _, subPerm := range subPerms {
				if subPerm == perm {
					return true
				}
			}
		}
	}
	return false
}

// hasRolePermission 检查角色权限
func (c *Checker) hasRolePermission(perm Permission) bool {
	rolePerms, exists := RolePermissions[c.role]
	if !exists {
		return false
	}

	deptPerms, exists := rolePerms[c.deptRole]
	if !exists {
		return false
	}

	for _, p := range deptPerms {
		if p == perm {
			return true
		}
	}
	return false
}

// hasTempPermission 检查临时权限
func (c *Checker) hasTempPermission(perm Permission) bool {
	for _, p := range c.tempPerms {
		if p.Permission == perm {
			return true
		}
	}
	return false
}

// HasTempPermissionWithResource 检查临时权限（带资源限定）
// resourceType: "all", "dept", "user"
// resourceID: 部门ID或用户ID（当resourceType为dept或user时使用）
func (c *Checker) HasTempPermissionWithResource(perm Permission, resourceType string, resourceID int) bool {
	for _, p := range c.tempPerms {
		if p.Permission == perm {
			// 资源类型为 "all" 或空，拥有全局权限
			if p.ResourceType == "all" || p.ResourceType == "" {
				return true
			}
			// 检查资源是否匹配
			if p.ResourceType == resourceType && p.ResourceID == resourceID {
				return true
			}
		}
	}
	return false
}

// HasTempPermissionForDept 检查是否有指定部门的临时权限
func (c *Checker) HasTempPermissionForDept(perm Permission, dept string) bool {
	// 先检查是否有全局权限
	for _, p := range c.tempPerms {
		if p.Permission == perm {
			if p.ResourceType == "all" || p.ResourceType == "" {
				return true
			}
		}
	}

	// 检查是否有该部门的权限
	// 方案1: 通过resource_id匹配（数字ID）
	for _, p := range c.tempPerms {
		if p.Permission == perm && p.ResourceType == "dept" {
			// 如果resource_id为0，表示所有部门
			if p.ResourceID == 0 {
				return true
			}
		}
	}

	// 方案2: 通过用户的部门匹配
	// 如果用户有所请求部门的临时权限，且用户的当前部门就是该部门
	if c.department == dept {
		for _, p := range c.tempPerms {
			if p.Permission == perm && p.ResourceType == "dept" {
				return true
			}
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

// IsAdmin 检查是否是系统管理员
func (c *Checker) IsAdmin() bool {
	return c.role == model.RoleAdmin
}

// IsOwner 检查是否是资源所有者（通过学号）
func (c *Checker) IsOwner(targetStudentID string) bool {
	return c.studentID == targetStudentID
}

// IsOfficeAdmin 检查是否是办公室管理员
func (c *Checker) IsOfficeAdmin() bool {
	return c.department == "办公室" && c.deptRole == model.DeptRoleAdmin
}

// IsDeptAdmin 检查是否是部门管理员
func (c *Checker) IsDeptAdmin() bool {
	return c.deptRole == model.DeptRoleAdmin
}

// CanManageDept 检查是否可以管理指定部门
func (c *Checker) CanManageDept(dept string) bool {
	// 系统管理员可以管理所有部门
	if c.IsAdmin() {
		return true
	}
	// 办公室管理员可以管理所有部门
	if c.IsOfficeAdmin() {
		return true
	}
	// 部门管理员只能管理自己部门
	if c.IsDeptAdmin() {
		return c.department == dept
	}
	// 检查临时权限：schedule:manage:all (全局)
	if c.hasTempPermission(PermScheduleManageAll) {
		return true
	}
	// 检查临时权限：schedule:manage:dept (特定部门)
	if c.HasTempPermissionForDept(PermScheduleManageDept, dept) {
		return true
	}
	return false
}

// CanViewDept 检查是否可以查看指定部门
func (c *Checker) CanViewDept(dept string) bool {
	// 系统管理员可以查看所有部门
	if c.IsAdmin() {
		return true
	}
	// 办公室管理员可以查看所有部门
	if c.IsOfficeAdmin() {
		return true
	}
	// 其他用户只能查看自己部门
	return c.department == dept
}

// HasDeptPermission 检查是否有指定部门的指定权限
func (c *Checker) HasDeptPermission(perm Permission, dept string) bool {
	// 首先检查是否有该权限
	if !c.HasPermission(perm) {
		return false
	}
	// 然后检查是否可以访问该部门
	return c.CanManageDept(dept)
}

// CanAccessPath 检查是否可以访问指定API路径
func (c *Checker) CanAccessPath(path string) bool {
	perm, exists := PathPermissionMap[path]
	if !exists {
		// 未配置的默认允许访问（或者可以改为默认拒绝）
		return true
	}
	return c.HasPermission(perm)
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

// CheckResult 返回权限检查结果
type CheckResult struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// Check 执行权限检查
func Check(studentID string, path string) bool {
	role := model.RoleUser
	if studentID == "admin" || len(studentID) >= 5 && studentID[:5] == "admin" {
		role = model.RoleAdmin
	}

	checker := NewChecker(0, studentID, role, "", "")
	return checker.CanAccessPath(path)
}

// CheckWithContext 使用Gin上下文进行权限检查
func CheckWithContext(c *gin.Context, perm Permission) bool {
	checker := FromContext(c)
	return checker.HasPermission(perm)
}

// CheckDeptPermissionWithContext 使用Gin上下文检查部门权限
func CheckDeptPermissionWithContext(c *gin.Context, perm Permission, dept string) bool {
	checker := FromContext(c)
	return checker.HasDeptPermission(perm, dept)
}

// Middleware 返回Gin权限检查中间件
func Middleware(perm Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		checker := FromContext(c)
		if !checker.HasPermission(perm) {
			c.JSON(http.StatusForbidden, model.Response{
				Code:    403,
				Message: "无权限执行此操作",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// AdminMiddleware 管理员权限中间件
func AdminMiddleware() gin.HandlerFunc {
	return Middleware(PermSystemAdmin)
}

// OwnerOrAdmin 检查是否是资源所有者或管理员
func OwnerOrAdmin(c *gin.Context, targetStudentID string) bool {
	checker := FromContext(c)
	if checker.IsAdmin() {
		return true
	}
	if checker.IsOfficeAdmin() {
		return true
	}
	return checker.IsOwner(targetStudentID)
}
