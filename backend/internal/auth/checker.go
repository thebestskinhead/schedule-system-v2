// Package auth 提供handler层面的权限检查工具
package auth

import (
	"net/http"
	"schedule-system-v2/backend/internal/model"

	"github.com/gin-gonic/gin"
)

// ContextKey 上下文键类型
type ContextKey string

const (
	// ContextKeyUserID 用户ID上下文键
	ContextKeyUserID ContextKey = "userID"
	// ContextKeyStudentID 学号上下文键
	ContextKeyStudentID ContextKey = "studentID"
	// ContextKeyRole 角色上下文键
	ContextKeyRole ContextKey = "role"
	// ContextKeyDepartment 部门上下文键
	ContextKeyDepartment ContextKey = "department"
	// ContextKeyDeptRole 部门角色上下文键
	ContextKeyDeptRole ContextKey = "deptRole"
)

// CurrentUser 当前登录用户信息
type CurrentUser struct {
	UserID     int
	StudentID  string
	Role       string
	Department string
	DeptRole   string
}

// GetCurrentUser 从上下文获取当前完整用户信息
func GetCurrentUser(c *gin.Context) *CurrentUser {
	return &CurrentUser{
		UserID:     GetUserIDFromContext(c),
		StudentID:  GetStudentIDFromContext(c),
		Role:       GetRoleFromContext(c),
		Department: GetDepartmentFromContext(c),
		DeptRole:   GetDeptRoleFromContext(c),
	}
}

// GetUserIDFromContext 从上下文获取用户ID
func GetUserIDFromContext(c *gin.Context) int {
	userID, exists := c.Get("userID")
	if !exists {
		return 0
	}
	uid, ok := userID.(int)
	if !ok {
		// 尝试float64（JSON数字默认类型）
		if fuid, ok := userID.(float64); ok {
			return int(fuid)
		}
		return 0
	}
	return uid
}

// GetStudentIDFromContext 从上下文获取学号
func GetStudentIDFromContext(c *gin.Context) string {
	studentID, exists := c.Get("studentID")
	if !exists {
		return ""
	}
	sid, ok := studentID.(string)
	if !ok {
		return ""
	}
	return sid
}

// GetRoleFromContext 从上下文获取角色
func GetRoleFromContext(c *gin.Context) string {
	role, exists := c.Get("role")
	if !exists {
		return ""
	}
	r, ok := role.(string)
	if !ok {
		return ""
	}
	return r
}

// GetDepartmentFromContext 从上下文获取部门
func GetDepartmentFromContext(c *gin.Context) string {
	dept, exists := c.Get("department")
	if !exists {
		return ""
	}
	d, ok := dept.(string)
	if !ok {
		return ""
	}
	return d
}

// GetDeptRoleFromContext 从上下文获取部门角色
func GetDeptRoleFromContext(c *gin.Context) string {
	deptRole, exists := c.Get("deptRole")
	if !exists {
		return ""
	}
	dr, ok := deptRole.(string)
	if !ok {
		return ""
	}
	return dr
}

// CheckPermission 检查当前用户是否有指定权限
// 在handler中使用: if !auth.CheckPermission(c, auth.PermScheduleEdit) { return }
func CheckPermission(c *gin.Context, perm Permission) bool {
	checker := GetChecker(c)
	if !checker.HasPermission(perm) {
		c.JSON(http.StatusForbidden, model.Response{
			Code:    403,
			Message: "无权限执行此操作",
		})
		return false
	}
	return true
}

// CheckDeptPermission 检查当前用户是否有指定部门的指定权限
// 在handler中使用: if !auth.CheckDeptPermission(c, auth.PermScheduleManageDept, "办公室") { return }
func CheckDeptPermission(c *gin.Context, perm Permission, dept string) bool {
	checker := GetChecker(c)
	if !checker.HasDeptPermission(perm, dept) {
		c.JSON(http.StatusForbidden, model.Response{
			Code:    403,
			Message: "无权访问该部门资源",
		})
		return false
	}
	return true
}

// CheckOwnerOrAdmin 检查是否是资源所有者或管理员
// 在handler中使用: if !auth.CheckOwnerOrAdmin(c, targetStudentID) { return }
func CheckOwnerOrAdmin(c *gin.Context, targetStudentID string) bool {
	checker := GetChecker(c)
	if !checker.IsAdmin() && checker.GetStudentID() != targetStudentID {
		c.JSON(http.StatusForbidden, model.Response{
			Code:    403,
			Message: "无权访问此资源",
		})
		return false
	}
	return true
}

// RequireAdmin 要求系统管理员权限
func RequireAdmin(c *gin.Context) bool {
	return CheckPermission(c, PermSystemAdmin)
}

// RequireOfficeAdmin 要求办公室管理员权限
func RequireOfficeAdmin(c *gin.Context) bool {
	checker := GetChecker(c)
	if !checker.IsOfficeAdmin() && !checker.IsAdmin() {
		c.JSON(http.StatusForbidden, model.Response{
			Code:    403,
			Message: "需要办公室管理员权限",
		})
		return false
	}
	return true
}

// GetChecker 从上下文获取权限检查器
func GetChecker(c *gin.Context) *Checker {
	userID := GetUserIDFromContext(c)
	studentID := GetStudentIDFromContext(c)
	role := GetRoleFromContext(c)
	department := GetDepartmentFromContext(c)
	deptRole := GetDeptRoleFromContext(c)
	return NewChecker(userID, studentID, role, department, deptRole)
}

// CheckAPIAccess 检查API访问权限（传入学号和接口路径）
// 返回布尔值表示是否允许访问
func CheckAPIAccess(studentID string, apiPath string) bool {
	return Check(studentID, apiPath)
}

// ResponseForbidden 返回权限不足响应
func ResponseForbidden(c *gin.Context, message ...string) {
	msg := "无权限执行此操作"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	c.JSON(http.StatusForbidden, model.Response{
		Code:    403,
		Message: msg,
	})
}

// ResponseUnauthorized 返回未认证响应
func ResponseUnauthorized(c *gin.Context, message ...string) {
	msg := "未登录或登录已过期"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	c.JSON(http.StatusUnauthorized, model.Response{
		Code:    401,
		Message: msg,
	})
}
