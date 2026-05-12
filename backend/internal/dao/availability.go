package dao

import (
	"schedule-system-v2/backend/internal/db"
	"schedule-system-v2/backend/internal/model"
)

type AvailabilityDAO struct{}

func NewAvailabilityDAO() *AvailabilityDAO {
	return &AvailabilityDAO{}
}

func (d *AvailabilityDAO) Create(availability *model.Availability) error {
	query := `INSERT INTO availability (user_id, week, weekday, period) VALUES (?, ?, ?, ?)`
	result, err := db.GetDB().Exec(query, availability.UserID, availability.Week, availability.Weekday, availability.Period)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	availability.ID = int(id)
	return nil
}

func (d *AvailabilityDAO) CreateBatch(availabilities []model.Availability) error {
	if len(availabilities) == 0 {
		return nil
	}
	query := `INSERT INTO availability (user_id, week, weekday, period) VALUES (:user_id, :week, :weekday, :period)`
	_, err := db.GetDB().NamedExec(query, availabilities)
	return err
}

func (d *AvailabilityDAO) GetByID(id int) (*model.Availability, error) {
	var availability model.Availability
	query := `SELECT * FROM availability WHERE id = ?`
	err := db.GetDB().Get(&availability, query, id)
	return &availability, err
}

func (d *AvailabilityDAO) GetByUserID(userID int) ([]model.Availability, error) {
	var list []model.Availability
	query := `SELECT * FROM availability WHERE user_id = ? ORDER BY week, weekday, period`
	err := db.GetDB().Select(&list, query, userID)
	return list, err
}

func (d *AvailabilityDAO) GetByTime(week, weekday, period int) ([]model.Availability, error) {
	var list []model.Availability
	query := `SELECT a.*, u.student_id, u.name as user_name 
		FROM availability a 
		JOIN users u ON a.user_id = u.id 
		WHERE a.week = ? AND a.weekday = ? AND a.period = ? AND u.is_active = 1`
	err := db.GetDB().Select(&list, query, week, weekday, period)
	return list, err
}

func (d *AvailabilityDAO) GetAvailableUsers(week, weekday, period int) ([]model.User, error) {
	var users []model.User
	query := `SELECT u.id, u.student_id, u.name 
		FROM users u 
		JOIN availability a ON u.id = a.user_id 
		WHERE a.week = ? AND a.weekday = ? AND a.period = ? AND u.is_active = 1
		AND u.dept_role != 'dept_admin'`
	err := db.GetDB().Select(&users, query, week, weekday, period)
	return users, err
}

// GetAvailableUsersForSchedule 获取可用于排班的用户（按部门筛选，排除部门管理员）
func (d *AvailabilityDAO) GetAvailableUsersForSchedule(week, weekday, period int, department string) ([]model.User, error) {
	var users []model.User
	query := `SELECT u.id, u.student_id, u.name 
		FROM users u 
		JOIN availability a ON u.id = a.user_id 
		WHERE a.week = ? AND a.weekday = ? AND a.period = ? 
		AND u.is_active = 1 
		AND u.department = ?
		AND u.dept_role != 'dept_admin'`
	err := db.GetDB().Select(&users, query, week, weekday, period, department)
	return users, err
}

func (d *AvailabilityDAO) GetAllGrouped() ([]model.AvailabilityWithUser, error) {
	var list []model.AvailabilityWithUser
	query := `SELECT a.*, u.student_id, u.name as user_name 
		FROM availability a 
		JOIN users u ON a.user_id = u.id 
		WHERE u.is_active = 1
		ORDER BY a.week, a.weekday, a.period, u.id`
	err := db.GetDB().Select(&list, query)
	return list, err
}

func (d *AvailabilityDAO) Delete(id int) error {
	query := `DELETE FROM availability WHERE id = ?`
	_, err := db.GetDB().Exec(query, id)
	return err
}

func (d *AvailabilityDAO) DeleteByUserIDAndTime(userID, week, weekday, period int) error {
	query := `DELETE FROM availability WHERE user_id = ? AND week = ? AND weekday = ? AND period = ?`
	_, err := db.GetDB().Exec(query, userID, week, weekday, period)
	return err
}

func (d *AvailabilityDAO) DeleteByUserID(userID int) error {
	query := `DELETE FROM availability WHERE user_id = ?`
	_, err := db.GetDB().Exec(query, userID)
	return err
}

func (d *AvailabilityDAO) Exists(userID, week, weekday, period int) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM availability WHERE user_id = ? AND week = ? AND weekday = ? AND period = ?`
	err := db.GetDB().Get(&count, query, userID, week, weekday, period)
	return count > 0, err
}

// GetAvailabilityMatrix 拉取某周某部门的可用性矩阵
func (d *AvailabilityDAO) GetAvailabilityMatrix(week int, department string) ([]model.AvailabilityMatrixItem, error) {
	var items []model.AvailabilityMatrixItem
	query := `SELECT a.user_id, a.weekday, a.period 
		FROM availability a 
		JOIN users u ON a.user_id = u.id 
		WHERE a.week = ? AND u.department = ? AND u.is_active = 1 AND u.dept_role != 'dept_admin'`
	err := db.GetDB().Select(&items, query, week, department)
	return items, err
}
