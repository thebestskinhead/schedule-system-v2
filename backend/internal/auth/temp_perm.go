package auth

import (
	"schedule-system-v2/backend/internal/dao"
)

// TempPermInfo 临时权限完整信息
type TempPermInfo struct {
	Permission   Permission
	ResourceType string // all, dept, user
	ResourceID   int
}

// loadTempPermissions 从数据库加载用户的有效临时权限（临时权限来源）
func loadTempPermissions(userID int) []TempPermInfo {
	if userID == 0 {
		return []TempPermInfo{}
	}

	tempPermDAO := dao.NewTempPermissionDAO()
	perms, err := tempPermDAO.GetActiveByUserID(userID)
	if err != nil {
		return []TempPermInfo{}
	}

	result := make([]TempPermInfo, 0, len(perms))
	for _, perm := range perms {
		result = append(result, TempPermInfo{
			Permission:   perm.Permission,
			ResourceType: perm.ResourceType,
			ResourceID:   perm.ResourceID,
		})
	}
	return result
}
