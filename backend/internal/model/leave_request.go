package model

import "time"

// LeaveRequest 请假申请（复用审批流，关联 applications 与 leave_requests）
// 表结构见 migrations/012_add_hr_tables.sql。
type LeaveRequest struct {
	ID            int       `json:"id" db:"id"`
	ApplicationID int       `json:"application_id" db:"application_id"`
	UserID        int       `json:"user_id" db:"user_id"`
	LeaveType     string    `json:"leave_type" db:"leave_type"` // sick/personal/annual
	StartDate     time.Time `json:"start_date" db:"start_date"`
	EndDate       time.Time `json:"end_date" db:"end_date"`
	Status        string    `json:"status" db:"status"` // pending/approved/rejected
	Reason        string    `json:"reason" db:"reason"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// 请假类型
const (
	LeaveTypeSick     = "sick"
	LeaveTypePersonal = "personal"
	LeaveTypeAnnual   = "annual"
)

// 请假状态（与 leave_requests.status 保持一致）
const (
	LeaveStatusPending  = "pending"
	LeaveStatusApproved = "approved"
	LeaveStatusRejected = "rejected"
)
