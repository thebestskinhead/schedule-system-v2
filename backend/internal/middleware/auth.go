package middleware

import (
	"net/http"
	"schedule-system-v2/backend/internal/auth"
	"schedule-system-v2/backend/internal/model"
	"schedule-system-v2/backend/internal/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

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

		c.Set("userID", claims.UserID)
		c.Set("studentID", claims.StudentID)
		c.Set("role", claims.Role)
		c.Set("department", claims.Department)
		c.Set("deptRole", claims.DeptRole)
		c.Next()
	}
}

// AdminMiddleware 管理员权限中间件（使用auth包）
func AdminMiddleware() gin.HandlerFunc {
	return auth.AdminMiddleware()
}

// PermissionMiddleware 指定权限检查中间件
func PermissionMiddleware(perm auth.Permission) gin.HandlerFunc {
	return auth.Middleware(perm)
}

// RequireAuthAndPermission 认证并检查权限
func RequireAuthAndPermission(perm auth.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行认证
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

		c.Set("userID", claims.UserID)
		c.Set("studentID", claims.StudentID)
		c.Set("role", claims.Role)
		c.Set("department", claims.Department)
		c.Set("deptRole", claims.DeptRole)

		// 再检查权限
		checker := auth.NewChecker(claims.UserID, claims.StudentID, claims.Role, claims.Department, claims.DeptRole)
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
