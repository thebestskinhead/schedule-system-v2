package dao

import (
	"context"
	"time"
	"schedule-system-v2/backend/internal/db"
	"schedule-system-v2/backend/internal/model"
)

type TempPermissionDAO struct{}

func NewTempPermissionDAO() *TempPermissionDAO {
	return &TempPermissionDAO{}
}

// Create 创建临时权限
func (d *TempPermissionDAO) Create(perm *model.UserPermissionTemp) error {
	query := `INSERT INTO user_permissions_temp 
			  (user_id, permission, resource_type, resource_id, granted_by, reason, expires_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	result, err := db.GetDB().Exec(query, 
		perm.UserID, perm.Permission, perm.ResourceType, perm.ResourceID, 
		perm.GrantedBy, perm.Reason, perm.ExpiresAt)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	perm.ID = int(id)
	return nil
}

// GetByID 根据ID获取临时权限
func (d *TempPermissionDAO) GetByID(id int) (*model.UserPermissionTemp, error) {
	var perm model.UserPermissionTemp
	query := `SELECT * FROM user_permissions_temp WHERE id = ?`
	err := db.GetDB().Get(&perm, query, id)
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

// GetActiveByUserID 获取用户所有有效的临时权限
func (d *TempPermissionDAO) GetActiveByUserID(userID int) ([]model.UserPermissionTemp, error) {
	var perms []model.UserPermissionTemp
	query := `SELECT * FROM user_permissions_temp 
			  WHERE user_id = ? AND is_active = TRUE AND expires_at > NOW()
			  ORDER BY created_at DESC`
	err := db.GetDB().Select(&perms, query, userID)
	return perms, err
}

// GetByUserID 获取用户的所有临时权限（包括已过期的）
func (d *TempPermissionDAO) GetByUserID(ctx context.Context, userID int) ([]model.UserPermissionTemp, error) {
	var perms []model.UserPermissionTemp
	query := `SELECT * FROM user_permissions_temp 
			  WHERE user_id = ?
			  ORDER BY created_at DESC`
	err := db.GetDB().Select(&perms, query, userID)
	return perms, err
}

// GetAllActive 获取所有有效的临时权限
func (d *TempPermissionDAO) GetAllActive() ([]model.UserPermissionTemp, error) {
	var perms []model.UserPermissionTemp
	query := `SELECT * FROM user_permissions_temp 
			  WHERE is_active = TRUE AND expires_at > NOW()
			  ORDER BY created_at DESC`
	err := db.GetDB().Select(&perms, query)
	return perms, err
}

// GetAll 获取所有临时权限（包含已过期和已撤销）
func (d *TempPermissionDAO) GetAll() ([]model.UserPermissionTemp, error) {
	var perms []model.UserPermissionTemp
	query := `SELECT * FROM user_permissions_temp ORDER BY created_at DESC`
	err := db.GetDB().Select(&perms, query)
	return perms, err
}

// Revoke 撤销权限（设置is_active为false）
func (d *TempPermissionDAO) Revoke(id int) error {
	query := `UPDATE user_permissions_temp SET is_active = FALSE WHERE id = ?`
	_, err := db.GetDB().Exec(query, id)
	return err
}

// Delete 删除权限记录
func (d *TempPermissionDAO) Delete(id int) error {
	query := `DELETE FROM user_permissions_temp WHERE id = ?`
	_, err := db.GetDB().Exec(query, id)
	return err
}

// CleanupExpired 清理已过期的权限（设置is_active为false）
func (d *TempPermissionDAO) CleanupExpired() error {
	query := `UPDATE user_permissions_temp SET is_active = FALSE WHERE expires_at <= NOW() AND is_active = TRUE`
	_, err := db.GetDB().Exec(query)
	return err
}

// HasPermission 检查用户是否有指定权限（包含临时权限）
func (d *TempPermissionDAO) HasPermission(userID int, perm model.Permission) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM user_permissions_temp 
			  WHERE user_id = ? AND permission = ? AND is_active = TRUE AND expires_at > NOW()`
	err := db.GetDB().Get(&count, query, userID, perm)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// HasPermissionWithResource 检查用户是否有指定权限和资源
func (d *TempPermissionDAO) HasPermissionWithResource(userID int, perm model.Permission, resourceType string, resourceID int) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM user_permissions_temp 
			  WHERE user_id = ? AND permission = ? 
			  AND (resource_type = 'all' OR (resource_type = ? AND resource_id = ?))
			  AND is_active = TRUE AND expires_at > NOW()`
	err := db.GetDB().Get(&count, query, userID, perm, resourceType, resourceID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetExpiredPermissions 获取已过期的权限
func (d *TempPermissionDAO) GetExpiredPermissions() ([]model.UserPermissionTemp, error) {
	var perms []model.UserPermissionTemp
	query := `SELECT * FROM user_permissions_temp WHERE expires_at <= NOW() AND is_active = TRUE`
	err := db.GetDB().Select(&perms, query)
	return perms, err
}

// ExtendExpiration 延长权限过期时间
func (d *TempPermissionDAO) ExtendExpiration(id int, newExpiresAt time.Time) error {
	query := `UPDATE user_permissions_temp SET expires_at = ? WHERE id = ?`
	_, err := db.GetDB().Exec(query, newExpiresAt, id)
	return err
}

// GetUserPermissionStats 获取用户权限统计
func (d *TempPermissionDAO) GetUserPermissionStats(userID int) (activeCount, expiredCount int, err error) {
	query := `
		SELECT 
			SUM(CASE WHEN is_active = TRUE AND expires_at > NOW() THEN 1 ELSE 0 END) as active,
			SUM(CASE WHEN is_active = FALSE OR expires_at <= NOW() THEN 1 ELSE 0 END) as expired
		FROM user_permissions_temp 
		WHERE user_id = ?`
	
	var result struct {
		Active  int `db:"active"`
		Expired int `db:"expired"`
	}
	
	err = db.GetDB().Get(&result, query, userID)
	return result.Active, result.Expired, err
}
