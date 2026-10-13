package auth

import (
	"net/http"
	"strings"

	"schedule-system-v2/backend/internal/dao"
	"schedule-system-v2/backend/internal/model"
	"schedule-system-v2/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// ==================== 上下文工具 ====================

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

// GetChecker 从上下文获取权限检查器
func GetChecker(c *gin.Context) *Checker {
	return FromContext(c)
}

// ==================== 中间件 ====================

// AuthMiddleware JWT认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, model.Response{Code: 401, Message: "未提供认证信息"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, model.Response{Code: 401, Message: "认证格式错误"})
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, model.Response{Code: 401, Message: "无效的token"})
			c.Abort()
			return
		}

		// 验签通过后，从数据库重载用户最新状态，而非信任 JWT 中的角色快照。
		// 这样权限变更（降权/调岗）即时生效，禁用账号（is_active=0）即时失效；
		// GetByID 已过滤 is_active=1，被禁用或已删除的用户在此被拒绝。
		user, err := dao.NewUserDAO().GetByID(claims.UserID)
		if err != nil || user == nil {
			c.JSON(http.StatusUnauthorized, model.Response{Code: 401, Message: "用户不存在或已被禁用"})
			c.Abort()
			return
		}

		// 三态账号状态断言：仅 employment_status='active' 可继续。
		// 确保离职/降权即时生效——即使 JWT 仍有效，状态变更后下次请求即被拦截。
		if user.EmploymentStatus != model.EmploymentStatusActive {
			c.JSON(http.StatusUnauthorized, model.Response{Code: 401, Message: "用户不存在或已被禁用"})
			c.Abort()
			return
		}

		c.Set("userID", user.ID)
		c.Set("studentID", user.StudentID)
		c.Set("role", user.Role)
		c.Set("department", user.Department)
		c.Set("deptRole", user.DeptRole)
		c.Next()
	}
}

// Middleware 返回Gin权限检查中间件
func Middleware(perm Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		checker := FromContext(c)
		if !checker.HasPermission(perm) {
			ResponseForbidden(c)
			c.Abort()
			return
		}
		c.Next()
	}
}

// MiddlewareOn 返回带资源范围的权限检查中间件
func MiddlewareOn(perm Permission, resourceType string, resourceID int) gin.HandlerFunc {
	return func(c *gin.Context) {
		checker := FromContext(c)
		if !checker.HasPermissionOn(perm, resourceType, resourceID) {
			ResponseForbidden(c)
			c.Abort()
			return
		}
		c.Next()
	}
}

// AnyPermissionMiddleware 多权限"任一满足"中间件
func AnyPermissionMiddleware(perms ...Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		checker := FromContext(c)
		if !checker.HasAnyPermission(perms...) {
			ResponseForbidden(c)
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

// ==================== handler 层检查工具 ====================

// CheckPermission 检查当前用户是否有指定权限
// 在handler中使用: if !auth.CheckPermission(c, auth.PermScheduleEdit) { return }
func CheckPermission(c *gin.Context, perm Permission) bool {
	checker := GetChecker(c)
	if !checker.HasPermission(perm) {
		ResponseForbidden(c)
		return false
	}
	return true
}

// CheckDeptPermission 检查当前用户是否有指定部门的指定权限
// 在handler中使用: if !auth.CheckDeptPermission(c, auth.PermScheduleManageDept, "办公室") { return }
func CheckDeptPermission(c *gin.Context, perm Permission, dept string) bool {
	checker := GetChecker(c)
	if !checker.HasPermissionOn(perm, "dept", model.DepartmentID(dept)) {
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
		ResponseForbidden(c, "无权访问此资源")
		return false
	}
	return true
}

// RequireAdmin 要求系统管理员权限
func RequireAdmin(c *gin.Context) bool {
	return CheckPermission(c, PermSystemAdmin)
}

// RequireOfficeAdmin 要求办公室管理员或系统管理员权限
func RequireOfficeAdmin(c *gin.Context) bool {
	checker := GetChecker(c)
	if !checker.IsAdmin() && !checker.HasPermission(PermUserManageAll) {
		c.JSON(http.StatusForbidden, model.Response{
			Code:    403,
			Message: "需要办公室管理员权限",
		})
		return false
	}
	return true
}

// OwnerOrAdmin 检查是否是资源所有者或管理员
func OwnerOrAdmin(c *gin.Context, targetStudentID string) bool {
	checker := FromContext(c)
	return checker.IsAdmin() || checker.IsOwner(targetStudentID)
}

// ==================== 响应工具 ====================

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
