package middleware

import (
	"schedule-system-v2/backend/internal/auth"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT认证中间件（实现见 auth 包）
func AuthMiddleware() gin.HandlerFunc {
	return auth.AuthMiddleware()
}

// AdminMiddleware 管理员权限中间件（使用auth包）
func AdminMiddleware() gin.HandlerFunc {
	return auth.AdminMiddleware()
}

// PermissionMiddleware 指定权限检查中间件
func PermissionMiddleware(perm auth.Permission) gin.HandlerFunc {
	return auth.Middleware(perm)
}

// PermissionOnMiddleware 带资源范围的权限检查中间件
func PermissionOnMiddleware(perm auth.Permission, resourceType string, resourceID int) gin.HandlerFunc {
	return auth.MiddlewareOn(perm, resourceType, resourceID)
}

// PermissionAnyMiddleware 多权限"任一满足"中间件
func PermissionAnyMiddleware(perms ...auth.Permission) gin.HandlerFunc {
	return auth.AnyPermissionMiddleware(perms...)
}
