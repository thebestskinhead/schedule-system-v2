package service

import (
	"context"
	"encoding/json"
	"errors"
	"schedule-system-v2/backend/internal/auth"
	"schedule-system-v2/backend/internal/dao"
	"schedule-system-v2/backend/internal/db"
	"schedule-system-v2/backend/internal/model"
)

type WeeklyDutyService struct {
	dao *dao.WeeklyDutyDAO
}

func NewWeeklyDutyService() *WeeklyDutyService {
	return &WeeklyDutyService{
		dao: dao.NewWeeklyDutyDAO(),
	}
}

// PublishAssignment 发布每周分工
func (s *WeeklyDutyService) PublishAssignment(adminID int, req *model.PublishAssignmentRequest) error {
	if req.Week < 1 || req.Week > 30 {
		return errors.New("周次必须在1-30之间")
	}

	// 序列化与 sp_publish_weekly_assignments 的 JSON_TABLE 解析字段一致：
	// department / weekday / is_assigned(以 0/1 表达)
	type assignmentItem struct {
		Department string `json:"department"`
		Weekday    int    `json:"weekday"`
		IsAssigned int    `json:"is_assigned"`
	}
	items := make([]assignmentItem, 0, len(req.Assignments))
	for _, detail := range req.Assignments {
		if detail.Weekday < 1 || detail.Weekday > 5 {
			continue // 跳过无效的星期
		}
		v := 0
		if detail.IsAssigned {
			v = 1
		}
		items = append(items, assignmentItem{
			Department: detail.Department,
			Weekday:    detail.Weekday,
			IsAssigned: v,
		})
	}

	if len(items) == 0 {
		return errors.New("没有有效的分工数据")
	}

	itemsJSON, err := json.Marshal(items)
	if err != nil {
		return err
	}

	err = db.WithRetry(context.Background(), func() error {
		_, e := db.GetDB().Exec("CALL sp_publish_weekly_assignments(?, ?, ?)", req.Week, adminID, string(itemsJSON))
		return e
	})
	if err != nil {
		if biz, ok := db.BusinessErrorFromSQLError(err); ok {
			return biz
		}
		return err
	}

	return nil
}

// GetWeekAssignments 获取指定周次的分工
// checker: 权限检查器
func (s *WeeklyDutyService) GetWeekAssignments(week int, userDept string, checker *auth.Checker) ([]model.WeeklyDutyAssignment, error) {
	// 拥有全局排班查看权限：返回所有部门分工
	if checker.HasPermission(auth.PermScheduleViewAll) {
		return s.dao.GetByWeek(week)
	}

	// 其余用户只能查看自己部门的分工
	return s.dao.GetByWeekAndDept(week, userDept)
}

// GetDeptAssignment 获取指定周次和部门的分工
func (s *WeeklyDutyService) GetDeptAssignment(week int, dept string) ([]model.WeeklyDutyAssignment, error) {
	return s.dao.GetByWeekAndDept(week, dept)
}

// GetMyDeptAssignment 获取用户所在部门的分工
func (s *WeeklyDutyService) GetMyDeptAssignment(week int, dept string) (*model.MyDeptAssignment, error) {
	assignments, err := s.dao.GetByWeekAndDept(week, dept)
	if err != nil {
		// 没有数据时返回空结果而不是错误
		assignments = []model.WeeklyDutyAssignment{}
	}

	result := &model.MyDeptAssignment{
		Week:       week,
		Department: dept,
		Weekdays:   make([]model.WeekdayInfo, 5),
	}

	// 默认全部不安排值班
	for i := 0; i < 5; i++ {
		result.Weekdays[i] = model.WeekdayInfo{
			Weekday:    i + 1,
			IsAssigned: false,
		}
	}

	// 填充实际数据
	for _, assignment := range assignments {
		if assignment.Weekday >= 1 && assignment.Weekday <= 5 {
			result.Weekdays[assignment.Weekday-1] = model.WeekdayInfo{
				Weekday:      assignment.Weekday,
				IsAssigned:   assignment.IsAssigned,
				AssignmentID: assignment.ID,
			}
		}
	}

	return result, nil
}

// UpdateAssignment 更新分工
func (s *WeeklyDutyService) UpdateAssignment(adminID int, req *model.UpdateAssignmentRequest) error {
	// 获取原记录
	assignment, err := s.dao.GetByID(req.ID)
	if err != nil {
		return errors.New("分工记录不存在")
	}

	assignment.IsAssigned = req.IsAssigned
	return s.dao.Update(assignment)
}

// DeleteAssignment 删除分工
func (s *WeeklyDutyService) DeleteAssignment(adminID int, id int) error {
	return s.dao.Delete(id)
}

// GetWeekAssignmentView 获取周分工视图（按部门聚合）
func (s *WeeklyDutyService) GetWeekAssignmentView(week int, checker *auth.Checker) (*model.WeekAssignmentView, error) {
	view := &model.WeekAssignmentView{
		Week:        week,
		Departments: []model.DeptAssignmentView{},
	}

	// 获取所有部门
	departments := model.Departments
	if !checker.HasPermission(auth.PermScheduleViewAll) {
		// 非全局管理员只能看自己部门
		departments = []string{checker.GetDepartment()}
	}

	for _, dept := range departments {
		deptView := model.DeptAssignmentView{
			Department: dept,
		}

		// 获取该部门的分工
		assignments, err := s.dao.GetByWeekAndDept(week, dept)
		if err != nil {
			continue
		}

		// 填充星期信息（默认不值班）
		for i := 0; i < 5; i++ {
			deptView.Weekdays[i] = model.WeekdayInfo{
				Weekday:    i + 1,
				IsAssigned: false,
			}
		}

		// 填充实际数据
		for _, assignment := range assignments {
			if assignment.Weekday >= 1 && assignment.Weekday <= 5 {
				deptView.Weekdays[assignment.Weekday-1] = model.WeekdayInfo{
					Weekday:      assignment.Weekday,
					IsAssigned:   assignment.IsAssigned,
					AssignmentID: assignment.ID,
				}
			}
		}

		view.Departments = append(view.Departments, deptView)
	}

	return view, nil
}

// CanPublish 检查用户是否可以发布分工
func (s *WeeklyDutyService) CanPublish(checker *auth.Checker) bool {
	// 拥有发布权限即可
	return checker.HasPermission(auth.PermSchedulePublish)
}
