package dao

import (
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"schedule-system-v2/backend/internal/db"
	"schedule-system-v2/backend/internal/model"
	"time"
)

type DutyDAO struct{}

func NewDutyDAO() *DutyDAO {
	return &DutyDAO{}
}

func (d *DutyDAO) Create(record *model.DutyRecord) error {
	query := `INSERT IGNORE INTO duty_records (week, weekday, period, user_id, assigned_by, status)
		VALUES (?, ?, ?, ?, ?, ?)`
	result, err := db.GetDB().Exec(query, record.Week, record.Weekday, record.Period, record.UserID, record.AssignedBy, record.Status)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	record.ID = int(id)
	return nil
}

func (d *DutyDAO) CreateBatch(records []model.DutyRecord) error {
	if len(records) == 0 {
		return nil
	}
	query := `INSERT IGNORE INTO duty_records (week, weekday, period, user_id, assigned_by, status)
		VALUES (:week, :weekday, :period, :user_id, :assigned_by, :status)`
	_, err := db.GetDB().NamedExec(query, records)
	return err
}

func (d *DutyDAO) GetByID(id int) (*model.DutyRecord, error) {
	var record model.DutyRecord
	query := `SELECT * FROM duty_records WHERE id = ?`
	err := db.GetDB().Get(&record, query, id)
	return &record, err
}

func (d *DutyDAO) GetByWeek(week int) ([]model.DutyRecordWithUser, error) {
	var list []model.DutyRecordWithUser
	query := `SELECT dr.id, dr.week, dr.weekday, dr.period, dr.user_id, dr.assigned_by, dr.status, dr.remark, dr.created_at, dr.updated_at, u.student_id, u.name as user_name
		FROM duty_records dr
		JOIN users u ON dr.user_id = u.id
		WHERE dr.week = ? AND u.employment_status = 'active'
		ORDER BY dr.weekday, dr.period, u.id`
	err := db.GetDB().Select(&list, query, week)
	return list, err
}

func (d *DutyDAO) GetByWeekAndDepartment(week int, department string) ([]model.DutyRecordWithUser, error) {
	var list []model.DutyRecordWithUser
	query := `SELECT dr.id, dr.week, dr.weekday, dr.period, dr.user_id, dr.assigned_by, dr.status, dr.remark, dr.created_at, dr.updated_at, u.student_id, u.name as user_name
		FROM duty_records dr
		JOIN users u ON dr.user_id = u.id
		WHERE dr.week = ? AND u.department = ? AND u.employment_status = 'active'
		ORDER BY dr.weekday, dr.period, u.id`
	err := db.GetDB().Select(&list, query, week, department)
	return list, err
}

func (d *DutyDAO) GetByUserID(userID int) ([]model.DutyRecordWithUser, error) {
	var list []model.DutyRecordWithUser
	query := `SELECT dr.id, dr.week, dr.weekday, dr.period, dr.user_id, dr.assigned_by, dr.status, dr.remark, dr.created_at, dr.updated_at, u.student_id, u.name as user_name
		FROM duty_records dr
		JOIN users u ON dr.user_id = u.id
		WHERE dr.user_id = ? AND u.employment_status = 'active'
		ORDER BY dr.week, dr.weekday, dr.period`
	err := db.GetDB().Select(&list, query, userID)
	return list, err
}

func (d *DutyDAO) GetByUserAndWeek(userID, week int) ([]model.DutyRecord, error) {
	var list []model.DutyRecord
	query := `SELECT * FROM duty_records WHERE user_id = ? AND week = ? ORDER BY weekday, period`
	err := db.GetDB().Select(&list, query, userID, week)
	return list, err
}

func (d *DutyDAO) GetByUserAndDay(userID, week, weekday int) ([]model.DutyRecord, error) {
	var list []model.DutyRecord
	query := `SELECT * FROM duty_records WHERE user_id = ? AND week = ? AND weekday = ?`
	err := db.GetDB().Select(&list, query, userID, week, weekday)
	return list, err
}

func (d *DutyDAO) UpdateStatus(id int, status string) error {
	query := `UPDATE duty_records SET status = ? WHERE id = ?`
	_, err := db.GetDB().Exec(query, status, id)
	return err
}

func (d *DutyDAO) DeleteByWeek(week int) error {
	query := `DELETE FROM duty_records WHERE week = ?`
	_, err := db.GetDB().Exec(query, week)
	return err
}

func (d *DutyDAO) DeleteByID(id int) error {
	query := `DELETE FROM duty_records WHERE id = ?`
	_, err := db.GetDB().Exec(query, id)
	return err
}

func (d *DutyDAO) CountByUser(userID int) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM duty_records WHERE user_id = ?`
	err := db.GetDB().Get(&count, query, userID)
	return count, err
}

func (d *DutyDAO) CountByUserAndWeek(userID, week int) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM duty_records WHERE user_id = ? AND week = ?`
	err := db.GetDB().Get(&count, query, userID, week)
	return count, err
}

func (d *DutyDAO) CountByUserAndDay(userID, week, weekday int) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM duty_records WHERE user_id = ? AND week = ? AND weekday = ?`
	err := db.GetDB().Get(&count, query, userID, week, weekday)
	return count, err
}

// BatchCountByUsers 批量查询全历史有效排班次数（实时聚合 duty_records，口径 status <> 'cancelled'，走索引 idx_user_status）
func (d *DutyDAO) BatchCountByUsers(userIDs []int) (map[int]int64, error) {
	result := make(map[int]int64)
	if len(userIDs) == 0 {
		return result, nil
	}
	query := `SELECT user_id, COUNT(*) AS cnt FROM duty_records WHERE user_id IN (?) AND status <> 'cancelled' GROUP BY user_id`
	args := []interface{}{userIDs}
	query, args, err := sqlx.In(query, args...)
	if err != nil {
		return result, err
	}
	query = db.GetDB().Rebind(query)
	rows, err := db.GetDB().Queryx(query, args...)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var uid int
		var cnt int64
		if err := rows.Scan(&uid, &cnt); err != nil {
			continue
		}
		result[uid] = cnt
	}
	return result, nil
}

// DeleteByWeekdayPeriodAndUser 删除指定时段的特定用户
func (d *DutyDAO) DeleteByWeekdayPeriodAndUser(week, weekday, period, userID int) error {
	query := `DELETE FROM duty_records WHERE week = ? AND weekday = ? AND period = ? AND user_id = ?`
	_, err := db.GetDB().Exec(query, week, weekday, period, userID)
	return err
}

// GetScheduleSettings 获取排班设置（全局唯一记录）
func (d *DutyDAO) GetScheduleSettings() (*model.ScheduleSettings, error) {
	var settings model.ScheduleSettings
	query := `SELECT * FROM schedule_settings WHERE config_key = 'global'`
	err := db.GetDB().Get(&settings, query)
	return &settings, err
}

// SaveScheduleSettings 保存排班设置（全局唯一记录）
func (d *DutyDAO) SaveScheduleSettings(settings *model.ScheduleSettings) error {
	query := `INSERT INTO schedule_settings (config_key, admin_id, current_week, current_week_set_at, auto_increment, need_per_cell, min_per_cell, max_per_day, max_per_week, export_title, created_at, updated_at)
		VALUES ('global', :admin_id, :current_week, :current_week_set_at, 1, :need_per_cell, :min_per_cell, :max_per_day, :max_per_week, :export_title, NOW(), NOW())
		ON DUPLICATE KEY UPDATE
		admin_id = :admin_id, current_week = :current_week, current_week_set_at = :current_week_set_at, auto_increment = 1, need_per_cell = :need_per_cell,
		min_per_cell = :min_per_cell, max_per_day = :max_per_day, max_per_week = :max_per_week,
		export_title = :export_title, updated_at = NOW()`
	var anchor interface{}
	if settings.CurrentWeekSetAt != nil {
		anchor = settings.CurrentWeekSetAt.Format("2006-01-02")
	}
	_, err := db.GetDB().NamedExec(query, map[string]interface{}{
		"admin_id": settings.AdminID, "current_week": settings.CurrentWeek, "current_week_set_at": anchor,
		"need_per_cell": settings.NeedPerCell, "min_per_cell": settings.MinPerCell,
		"max_per_day": settings.MaxPerDay, "max_per_week": settings.MaxPerWeek, "export_title": settings.ExportTitle,
	})
	return err
}

// UpdateCurrentWeek 更新当前周次（全局唯一记录）
func (d *DutyDAO) UpdateCurrentWeek(adminID int, currentWeek int, anchor time.Time) error {
	query := `INSERT INTO schedule_settings (config_key, admin_id, current_week, current_week_set_at, auto_increment, need_per_cell, min_per_cell, max_per_day, max_per_week, export_title, semester_start_date, created_at, updated_at)
		VALUES ('global', ?, ?, ?, 1, 2, 0, 1, 2, '第{week}周排班表', NULL, NOW(), NOW())
		ON DUPLICATE KEY UPDATE
		admin_id = VALUES(admin_id), current_week = VALUES(current_week), current_week_set_at = VALUES(current_week_set_at), auto_increment = 1, updated_at = NOW()`
	_, err := db.GetDB().Exec(query, adminID, currentWeek, anchor.Format("2006-01-02"))
	return err
}

// GetSemesterStartDate 获取学期起始日（全局唯一记录）
func (d *DutyDAO) GetSemesterStartDate() (*string, error) {
	var result struct {
		SemesterStartDate *time.Time `db:"semester_start_date"`
	}
	query := `SELECT semester_start_date FROM schedule_settings WHERE config_key = 'global'`
	err := db.GetDB().Get(&result, query)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && result.SemesterStartDate == nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	date := result.SemesterStartDate.Format("2006-01-02")
	return &date, nil
}

// UpdateSemesterStartDate 更新学期起始日（全局唯一记录）
func (d *DutyDAO) UpdateSemesterStartDate(adminID int, startDate string) error {
	// 默认开启自动计算（auto_increment=1），已存在记录时更新学期起始日和修改者
	query := `INSERT INTO schedule_settings (config_key, admin_id, current_week, auto_increment, need_per_cell, min_per_cell, max_per_day, max_per_week, export_title, semester_start_date, created_at, updated_at)
		VALUES ('global', ?, 1, 1, 2, 0, 1, 2, '第{week}周排班表', ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE
		admin_id = ?, semester_start_date = ?, current_week_set_at = NULL, auto_increment = 1, updated_at = NOW()`
	_, err := db.GetDB().Exec(query, adminID, startDate, adminID, startDate)
	return err
}
