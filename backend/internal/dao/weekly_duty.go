package dao

import (
	"github.com/jmoiron/sqlx"
	"schedule-system-v2/backend/internal/db"
	"schedule-system-v2/backend/internal/model"
)

type WeeklyDutyDAO struct{}

func NewWeeklyDutyDAO() *WeeklyDutyDAO {
	return &WeeklyDutyDAO{}
}

// Create 创建每周分工记录
func (d *WeeklyDutyDAO) Create(assignment *model.WeeklyDutyAssignment) error {
	query := `INSERT INTO weekly_duty_assignments (week, department, weekday, is_assigned, created_by) 
			  VALUES (?, ?, ?, ?, ?)`
	result, err := db.GetDB().Exec(query, assignment.Week, assignment.Department, assignment.Weekday, assignment.IsAssigned, assignment.CreatedBy)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	assignment.ID = int(id)
	return nil
}

// CreateBatch 批量创建分工记录
func (d *WeeklyDutyDAO) CreateBatch(assignments []*model.WeeklyDutyAssignment) error {
	if len(assignments) == 0 {
		return nil
	}

	tx, err := db.GetDB().Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO weekly_duty_assignments (week, department, weekday, is_assigned, created_by) 
			  VALUES (?, ?, ?, ?, ?)`

	for _, assignment := range assignments {
		result, err := tx.Exec(query, assignment.Week, assignment.Department, assignment.Weekday, assignment.IsAssigned, assignment.CreatedBy)
		if err != nil {
			return err
		}
		id, _ := result.LastInsertId()
		assignment.ID = int(id)
	}

	return tx.Commit()
}

// GetByID 根据ID获取分工记录
func (d *WeeklyDutyDAO) GetByID(id int) (*model.WeeklyDutyAssignment, error) {
	var assignment model.WeeklyDutyAssignment
	query := `SELECT * FROM weekly_duty_assignments WHERE id = ?`
	err := db.GetDB().Get(&assignment, query, id)
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

// GetByWeek 获取指定周次的所有分工
func (d *WeeklyDutyDAO) GetByWeek(week int) ([]model.WeeklyDutyAssignment, error) {
	var assignments []model.WeeklyDutyAssignment
	query := `SELECT * FROM weekly_duty_assignments WHERE week = ? ORDER BY department, weekday`
	err := db.GetDB().Select(&assignments, query, week)
	return assignments, err
}

// GetByWeekAndDept 获取指定周次和部门的分工
func (d *WeeklyDutyDAO) GetByWeekAndDept(week int, dept string) ([]model.WeeklyDutyAssignment, error) {
	var assignments []model.WeeklyDutyAssignment
	query := `SELECT * FROM weekly_duty_assignments WHERE week = ? AND department = ? ORDER BY weekday`
	err := db.GetDB().Select(&assignments, query, week, dept)
	return assignments, err
}

// GetByWeekAndDepartments 获取指定周次和多个部门的分工
func (d *WeeklyDutyDAO) GetByWeekAndDepartments(week int, depts []string) ([]model.WeeklyDutyAssignment, error) {
	if len(depts) == 0 {
		return []model.WeeklyDutyAssignment{}, nil
	}

	query := `SELECT * FROM weekly_duty_assignments WHERE week = ? AND department IN (?) ORDER BY department, weekday`
	query, args, err := sqlx.In(query, append([]interface{}{week}, toInterfaceSlice(depts)...)...)
	if err != nil {
		return nil, err
	}

	var assignments []model.WeeklyDutyAssignment
	err = db.GetDB().Select(&assignments, query, args...)
	return assignments, err
}

// toInterfaceSlice 将字符串切片转换为接口切片
func toInterfaceSlice(strs []string) []interface{} {
	result := make([]interface{}, len(strs))
	for i, s := range strs {
		result[i] = s
	}
	return result
}

// GetByWeekDeptWeekday 获取指定周次、部门和星期的分工
func (d *WeeklyDutyDAO) GetByWeekDeptWeekday(week int, dept string, weekday int) (*model.WeeklyDutyAssignment, error) {
	var assignment model.WeeklyDutyAssignment
	query := `SELECT * FROM weekly_duty_assignments WHERE week = ? AND department = ? AND weekday = ?`
	err := db.GetDB().Get(&assignment, query, week, dept, weekday)
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

// Update 更新分工记录
func (d *WeeklyDutyDAO) Update(assignment *model.WeeklyDutyAssignment) error {
	query := `UPDATE weekly_duty_assignments SET is_assigned = ? WHERE id = ?`
	_, err := db.GetDB().Exec(query, assignment.IsAssigned, assignment.ID)
	return err
}

// Delete 删除分工记录
func (d *WeeklyDutyDAO) Delete(id int) error {
	query := `DELETE FROM weekly_duty_assignments WHERE id = ?`
	_, err := db.GetDB().Exec(query, id)
	return err
}

// DeleteByWeek 删除指定周次的所有分工
func (d *WeeklyDutyDAO) DeleteByWeek(week int) error {
	query := `DELETE FROM weekly_duty_assignments WHERE week = ?`
	_, err := db.GetDB().Exec(query, week)
	return err
}

// Exists 检查是否存在
func (d *WeeklyDutyDAO) Exists(week int, dept string, weekday int) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM weekly_duty_assignments WHERE week = ? AND department = ? AND weekday = ?`
	err := db.GetDB().Get(&count, query, week, dept, weekday)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Upsert 如果不存在则创建，存在则更新
func (d *WeeklyDutyDAO) Upsert(assignment *model.WeeklyDutyAssignment) error {
	exists, err := d.Exists(assignment.Week, assignment.Department, assignment.Weekday)
	if err != nil {
		return err
	}

	if exists {
		// 获取现有记录ID
		existing, err := d.GetByWeekDeptWeekday(assignment.Week, assignment.Department, assignment.Weekday)
		if err != nil {
			return err
		}
		assignment.ID = existing.ID
		return d.Update(assignment)
	}

	return d.Create(assignment)
}
