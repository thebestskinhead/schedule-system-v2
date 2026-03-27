package dao

import (
	"schedule-system-v2/backend/internal/db"
	"schedule-system-v2/backend/internal/model"
)

type DutyDAO struct{}

func NewDutyDAO() *DutyDAO {
	return &DutyDAO{}
}

func (d *DutyDAO) Create(record *model.DutyRecord) error {
	query := `INSERT INTO duty_records (week, weekday, period, user_id, assigned_by, status) 
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
	query := `INSERT INTO duty_records (week, weekday, period, user_id, assigned_by, status) 
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
		WHERE dr.week = ? AND u.is_active = 1
		ORDER BY dr.weekday, dr.period, u.id`
	err := db.GetDB().Select(&list, query, week)
	return list, err
}

func (d *DutyDAO) GetByWeekAndDepartment(week int, department string) ([]model.DutyRecordWithUser, error) {
	var list []model.DutyRecordWithUser
	query := `SELECT dr.id, dr.week, dr.weekday, dr.period, dr.user_id, dr.assigned_by, dr.status, dr.remark, dr.created_at, dr.updated_at, u.student_id, u.name as user_name 
		FROM duty_records dr 
		JOIN users u ON dr.user_id = u.id 
		WHERE dr.week = ? AND u.department = ? AND u.is_active = 1
		ORDER BY dr.weekday, dr.period, u.id`
	err := db.GetDB().Select(&list, query, week, department)
	return list, err
}

func (d *DutyDAO) GetByUserID(userID int) ([]model.DutyRecordWithUser, error) {
	var list []model.DutyRecordWithUser
	query := `SELECT dr.id, dr.week, dr.weekday, dr.period, dr.user_id, dr.assigned_by, dr.status, dr.remark, dr.created_at, dr.updated_at, u.student_id, u.name as user_name 
		FROM duty_records dr 
		JOIN users u ON dr.user_id = u.id 
		WHERE dr.user_id = ? AND u.is_active = 1
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

// DeleteByWeekdayPeriodAndUser 删除指定时段的特定用户
func (d *DutyDAO) DeleteByWeekdayPeriodAndUser(week, weekday, period, userID int) error {
	query := `DELETE FROM duty_records WHERE week = ? AND weekday = ? AND period = ? AND user_id = ?`
	_, err := db.GetDB().Exec(query, week, weekday, period, userID)
	return err
}

// GetScheduleSettings 获取排班设置
func (d *DutyDAO) GetScheduleSettings(adminID int) (*model.ScheduleSettings, error) {
	var settings model.ScheduleSettings
	query := `SELECT * FROM schedule_settings WHERE admin_id = ? ORDER BY id DESC LIMIT 1`
	err := db.GetDB().Get(&settings, query, adminID)
	return &settings, err
}

// SaveScheduleSettings 保存排班设置
func (d *DutyDAO) SaveScheduleSettings(settings *model.ScheduleSettings) error {
	query := `INSERT INTO schedule_settings (admin_id, current_week, auto_increment, need_per_cell, min_per_cell, max_per_day, max_per_week, export_title, semester_start_date, created_at, updated_at) 
		VALUES (:admin_id, :current_week, :auto_increment, :need_per_cell, :min_per_cell, :max_per_day, :max_per_week, :export_title, :semester_start_date, NOW(), NOW())
		ON DUPLICATE KEY UPDATE 
		current_week = :current_week, auto_increment = :auto_increment, need_per_cell = :need_per_cell, 
		min_per_cell = :min_per_cell, max_per_day = :max_per_day, max_per_week = :max_per_week, 
		export_title = :export_title, semester_start_date = :semester_start_date, updated_at = NOW()`
	_, err := db.GetDB().NamedExec(query, settings)
	return err
}

// GetCurrentWeek 获取当前周次（取第一个管理员的设置）
func (d *DutyDAO) GetCurrentWeek() (int, bool, error) {
	var result struct {
		CurrentWeek   int `db:"current_week"`
		AutoIncrement int `db:"auto_increment"`
	}
	query := `SELECT current_week, auto_increment FROM schedule_settings ORDER BY id DESC LIMIT 1`
	err := db.GetDB().Get(&result, query)
	if err != nil {
		// 如果没有设置，返回默认值
		return 1, false, nil
	}
	return result.CurrentWeek, result.AutoIncrement == 1, nil
}

// UpdateCurrentWeek 更新当前周次
func (d *DutyDAO) UpdateCurrentWeek(adminID int, currentWeek int, autoIncrement bool) error {
	autoIncVal := 0
	if autoIncrement {
		autoIncVal = 1
	}
	query := `INSERT INTO schedule_settings (admin_id, current_week, auto_increment, need_per_cell, min_per_cell, max_per_day, max_per_week, export_title, semester_start_date, created_at, updated_at) 
		VALUES (?, ?, ?, 2, 0, 1, 2, '第{week}周排班表', NULL, NOW(), NOW())
		ON DUPLICATE KEY UPDATE 
		current_week = ?, auto_increment = ?, updated_at = NOW()`
	_, err := db.GetDB().Exec(query, adminID, currentWeek, autoIncVal, currentWeek, autoIncVal)
	return err
}

// IncrementCurrentWeek 当前周次自动递增
func (d *DutyDAO) IncrementCurrentWeek() error {
	query := `UPDATE schedule_settings SET current_week = current_week + 1, updated_at = NOW() 
		WHERE auto_increment = 1 AND current_week < 30`
	_, err := db.GetDB().Exec(query)
	return err
}

// GetSemesterStartDate 获取学期起始日
func (d *DutyDAO) GetSemesterStartDate() (*string, error) {
	var result struct {
		SemesterStartDate *string `db:"semester_start_date"`
	}
	query := `SELECT semester_start_date FROM schedule_settings ORDER BY id DESC LIMIT 1`
	err := db.GetDB().Get(&result, query)
	if err != nil {
		// 如果没有设置，返回nil
		return nil, nil
	}
	return result.SemesterStartDate, nil
}

// UpdateSemesterStartDate 更新学期起始日
func (d *DutyDAO) UpdateSemesterStartDate(adminID int, startDate string) error {
	query := `INSERT INTO schedule_settings (admin_id, current_week, auto_increment, need_per_cell, min_per_cell, max_per_day, max_per_week, export_title, semester_start_date, created_at, updated_at) 
		VALUES (?, 1, 0, 2, 0, 1, 2, '第{week}周排班表', ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE 
		semester_start_date = ?, updated_at = NOW()`
	_, err := db.GetDB().Exec(query, adminID, startDate, startDate)
	return err
}
