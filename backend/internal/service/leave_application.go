package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"schedule-system-v2/backend/internal/dao"
	"schedule-system-v2/backend/internal/db"
	"schedule-system-v2/backend/internal/model"
)

// LeaveApplicationData 请假申请 data（对应 applications.data 与 leave_requests 字段）
type LeaveApplicationData struct {
	LeaveType string `json:"leave_type"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Reason    string `json:"reason"`
}

// LeaveExecutor 请假申请执行器，实现 service.ApplicationExecutor（照既有 temp_permission 范例）。
// 注册类型 "leave"；OnApproved 调 sp_approve_leave。
type LeaveExecutor struct {
	userDao        *dao.UserDAO
	leaveDao       *dao.LeaveRequestDAO
	dutyDao        *dao.DutyDAO
	applicationDao *dao.ApplicationDao
}

func NewLeaveExecutor(userDao *dao.UserDAO, leaveDao *dao.LeaveRequestDAO, dutyDao *dao.DutyDAO, applicationDao *dao.ApplicationDao) *LeaveExecutor {
	return &LeaveExecutor{
		userDao:        userDao,
		leaveDao:       leaveDao,
		dutyDao:        dutyDao,
		applicationDao: applicationDao,
	}
}

func (e *LeaveExecutor) GetType() string {
	return "leave"
}

func (e *LeaveExecutor) Validate(data json.RawMessage) error {
	var d LeaveApplicationData
	if err := json.Unmarshal(data, &d); err != nil {
		return fmt.Errorf("数据格式错误: %v", err)
	}
	if d.LeaveType == "" {
		return errors.New("请假类型不能为空")
	}
	switch d.LeaveType {
	case model.LeaveTypeSick, model.LeaveTypePersonal, model.LeaveTypeAnnual:
	default:
		return fmt.Errorf("无效的请假类型: %s", d.LeaveType)
	}
	if d.StartDate == "" || d.EndDate == "" {
		return errors.New("请假起止日期不能为空")
	}
	start, err := time.Parse("2006-01-02", d.StartDate)
	if err != nil {
		return errors.New("开始日期格式错误")
	}
	end, err := time.Parse("2006-01-02", d.EndDate)
	if err != nil {
		return errors.New("结束日期格式错误")
	}
	if end.Before(start) {
		return errors.New("结束日期不能早于开始日期")
	}
	return nil
}

// OnCreated 申请创建后：写入 leave_requests（status=pending）
func (e *LeaveExecutor) OnCreated(ctx context.Context, app *model.Application) error {
	var d LeaveApplicationData
	if err := json.Unmarshal(app.Data, &d); err != nil {
		return fmt.Errorf("解析请假数据失败: %v", err)
	}
	start, err := time.Parse("2006-01-02", d.StartDate)
	if err != nil {
		return fmt.Errorf("开始日期解析失败: %v", err)
	}
	end, err := time.Parse("2006-01-02", d.EndDate)
	if err != nil {
		return fmt.Errorf("结束日期解析失败: %v", err)
	}
	lr := &model.LeaveRequest{
		ApplicationID: app.ID,
		UserID:        app.ApplicantID,
		LeaveType:     d.LeaveType,
		StartDate:     start,
		EndDate:       end,
		Status:        model.LeaveStatusPending,
		Reason:        d.Reason,
	}
	if err := e.leaveDao.Create(ctx, lr); err != nil {
		return fmt.Errorf("写入请假记录失败: %v", err)
	}
	return nil
}

// OnApproved 审批通过：更新请假状态为 approved，并调 sp_approve_leave 取消覆盖期内排班
func (e *LeaveExecutor) OnApproved(ctx context.Context, app *model.Application, approval *model.ApplicationApproval) error {
	if err := e.leaveDao.UpdateStatusByApplicationID(ctx, app.ID, model.LeaveStatusApproved); err != nil {
		return fmt.Errorf("更新请假状态失败: %v", err)
	}

	semesterStart, err := e.dutyDao.GetSemesterStartDate()
	if err != nil {
		return fmt.Errorf("读取学期起始日失败: %v", err)
	}
	weekStart, err := computeWeekStart(semesterStart)
	if err != nil {
		return err
	}

	spErr := db.WithRetry(ctx, func() error {
		_, ex := db.GetDB().Exec("CALL sp_approve_leave(?, ?)", app.ID, weekStart.Format("2006-01-02"))
		return ex
	})
	if spErr != nil {
		if biz, ok := db.BusinessErrorFromSQLError(spErr); ok {
			return biz
		}
		return spErr
	}
	return nil
}

// OnRejected 审批拒绝：请假状态置 rejected
func (e *LeaveExecutor) OnRejected(ctx context.Context, app *model.Application, approval *model.ApplicationApproval) error {
	if err := e.leaveDao.UpdateStatusByApplicationID(ctx, app.ID, model.LeaveStatusRejected); err != nil {
		return fmt.Errorf("更新请假状态失败: %v", err)
	}
	return nil
}

// GetRequiredApprovers 审批人：本部门部门管理员，无则回退系统管理员（排除申请人自己）
func (e *LeaveExecutor) GetRequiredApprovers(ctx context.Context, app *model.Application) ([]int, error) {
	applicant, err := e.userDao.GetByID(app.ApplicantID)
	if err != nil {
		return nil, err
	}
	deptAdmins, err := e.userDao.GetByDepartmentAndRole(applicant.Department, model.RoleDeptAdmin, model.EmploymentStatusActive)
	if err != nil {
		return nil, err
	}
	var ids []int
	for _, a := range deptAdmins {
		if a.ID != app.ApplicantID {
			ids = append(ids, a.ID)
		}
	}
	if len(ids) == 0 {
		admins, err := e.userDao.GetByRole(model.RoleAdmin, model.EmploymentStatusActive)
		if err != nil {
			return nil, err
		}
		for _, a := range admins {
			if a.ID != app.ApplicantID {
				ids = append(ids, a.ID)
			}
		}
	}
	return ids, nil
}

// CanApply 仅在职成员可申请请假
func (e *LeaveExecutor) CanApply(ctx context.Context, userID int, data json.RawMessage) (bool, string) {
	user, err := e.userDao.GetByID(userID)
	if err != nil {
		return false, "获取用户信息失败"
	}
	if user.EmploymentStatus != model.EmploymentStatusActive {
		return false, "仅在职成员可申请请假"
	}
	return true, ""
}

// computeWeekStart 由学期起始日推导 sp_approve_leave 的 p_week_start 锚点。
// 系统约定：week=1, weekday=1 对应的日历日期 = semester_start_date（学期起始日即第1周周一），
// 与 dao/availability.go 中排班候选排除请假的日期算法完全一致。
// 若需改为“学期起始日所在周的周一”，仅需在此处做 weekday 偏移，SP 与候选查询同步生效。
func computeWeekStart(semesterStart *string) (time.Time, error) {
	if semesterStart == nil || *semesterStart == "" {
		return time.Time{}, errors.New("未设置学期起始日(semester_start_date)，无法计算请假对应排班日期")
	}
	t, err := time.Parse("2006-01-02", *semesterStart)
	if err != nil {
		return time.Time{}, fmt.Errorf("学期起始日格式错误: %w", err)
	}
	return t, nil
}
