package dao

import (
	"schedule-system-v2/backend/internal/db"

	"github.com/jmoiron/sqlx"
)

// ResignDAO 离职（批量）相关只读统计
type ResignDAO struct{}

func NewResignDAO() *ResignDAO {
	return &ResignDAO{}
}

// CountPendingDuties 统计给定用户的 pending 状态排班数（将因离职被置 cancelled）
func (d *ResignDAO) CountPendingDuties(userIDs []int) (int64, error) {
	if len(userIDs) == 0 {
		return 0, nil
	}
	query, args, err := sqlx.In(`SELECT COUNT(*) FROM duty_records WHERE user_id IN (?) AND status = 'pending'`, userIDs)
	if err != nil {
		return 0, err
	}
	query = db.GetDB().Rebind(query)
	var count int64
	if err := db.GetDB().Get(&count, query, args...); err != nil {
		return 0, err
	}
	return count, nil
}

// CountActiveTempPerms 统计给定用户生效中的临时权限数（将因离职被置失效）
func (d *ResignDAO) CountActiveTempPerms(userIDs []int) (int64, error) {
	if len(userIDs) == 0 {
		return 0, nil
	}
	query, args, err := sqlx.In(`SELECT COUNT(*) FROM user_permissions_temp WHERE user_id IN (?) AND is_active = 1`, userIDs)
	if err != nil {
		return 0, err
	}
	query = db.GetDB().Rebind(query)
	var count int64
	if err := db.GetDB().Get(&count, query, args...); err != nil {
		return 0, err
	}
	return count, nil
}
