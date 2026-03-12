package model

import "time"

// WeeklyDutyAssignment 每周值班分工
type WeeklyDutyAssignment struct {
	ID         int       `json:"id" db:"id"`
	Week       int       `json:"week" db:"week"`
	Department string    `json:"department" db:"department"`
	Weekday    int       `json:"weekday" db:"weekday"`
	IsAssigned bool      `json:"is_assigned" db:"is_assigned"`
	CreatedBy  int       `json:"created_by" db:"created_by"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// PublishAssignmentRequest 发布分工请求
type PublishAssignmentRequest struct {
	Week        int                `json:"week" binding:"required,min=1,max=30"`
	Assignments []AssignmentDetail `json:"assignments" binding:"required"`
}

// AssignmentDetail 分工详情
type AssignmentDetail struct {
	Department string `json:"department" binding:"required"`
	Weekday    int    `json:"weekday" binding:"required,min=1,max=5"`
	IsAssigned bool   `json:"is_assigned"`
}

// UpdateAssignmentRequest 更新分工请求
type UpdateAssignmentRequest struct {
	ID         int  `json:"id" binding:"required"`
	IsAssigned bool `json:"is_assigned"`
}

// WeekAssignmentView 周分工视图（按部门聚合）
type WeekAssignmentView struct {
	Week        int                       `json:"week"`
	Departments []DeptAssignmentView      `json:"departments"`
}

// DeptAssignmentView 部门分工视图
type DeptAssignmentView struct {
	Department string          `json:"department"`
	Weekdays   [5]WeekdayInfo  `json:"weekdays"` // 周一到周五
}

// WeekdayInfo 星期信息
type WeekdayInfo struct {
	Weekday    int  `json:"weekday"`
	IsAssigned bool `json:"is_assigned"`
	AssignmentID int `json:"assignment_id,omitempty"`
}

// MyDeptAssignment 我的部门分工
type MyDeptAssignment struct {
	Week       int            `json:"week"`
	Department string         `json:"department"`
	Weekdays   []WeekdayInfo  `json:"weekdays"`
}
