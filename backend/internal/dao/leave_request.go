package dao

import (
	"context"
	"schedule-system-v2/backend/internal/db"
	"schedule-system-v2/backend/internal/model"
)

// LeaveRequestDAO 请假记录数据访问
type LeaveRequestDAO struct{}

func NewLeaveRequestDAO() *LeaveRequestDAO {
	return &LeaveRequestDAO{}
}

// Create 写入请假记录（与 applications 同时创建，见 H.4.2）
func (d *LeaveRequestDAO) Create(ctx context.Context, lr *model.LeaveRequest) error {
	query := `INSERT INTO leave_requests (application_id, user_id, leave_type, start_date, end_date, status, reason)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	result, err := db.GetDB().Exec(query,
		lr.ApplicationID, lr.UserID, lr.LeaveType,
		lr.StartDate.Format("2006-01-02"), lr.EndDate.Format("2006-01-02"),
		lr.Status, lr.Reason)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	lr.ID = int(id)
	return nil
}

// UpdateStatusByApplicationID 按申请单更新请假状态（批准/拒绝联动）
func (d *LeaveRequestDAO) UpdateStatusByApplicationID(ctx context.Context, applicationID int, status string) error {
	query := `UPDATE leave_requests SET status = ?, updated_at = NOW() WHERE application_id = ?`
	_, err := db.GetDB().Exec(query, status, applicationID)
	return err
}
