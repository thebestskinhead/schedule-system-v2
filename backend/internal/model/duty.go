package model

import "time"

// DutyRecord 值班记录
type DutyRecord struct {
	ID         int        `json:"id" db:"id"`
	Week       int        `json:"week" db:"week"`
	Weekday    int        `json:"weekday" db:"weekday"`
	Period     int        `json:"period" db:"period"`
	UserID     int        `json:"user_id" db:"user_id"`
	UserName   string     `json:"user_name,omitempty" db:"-"`
	AssignedBy int        `json:"assigned_by,omitempty" db:"assigned_by"`
	Status     string     `json:"status" db:"status"`
	Remark     *string    `json:"remark,omitempty" db:"remark"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// DutyStatus 值班状态常量
const (
	DutyStatusPending   = "pending"   // 待确认
	DutyStatusConfirmed = "confirmed" // 已确认
	DutyStatusCompleted = "completed" // 已完成
	DutyStatusCancelled = "cancelled" // 已取消
)

// ScheduleRequest 排班请求
type ScheduleRequest struct {
	Week           int   `json:"week" binding:"required,min=1,max=30"`
	Days           []int `json:"days" binding:"required"`
	NeedPerCell    int   `json:"need_per_cell" binding:"required,min=0,max=10"`
	MinPerCell     int   `json:"min_per_cell" binding:"min=0,max=10"`     // 最小排班人数，默认0
	Periods        int   `json:"periods" binding:"required,min=1,max=4"`
	MaxPerDay      int   `json:"max_per_day" binding:"min=1,max=10"`      // 每人每天最多排几次，默认1
	MaxPerWeek     int   `json:"max_per_week" binding:"min=1,max=30"`     // 每人每周最多排几次，默认2
	Department     string `json:"department" binding:"required"`          // 排班部门
}

// ScheduleSettings 排班设置（用于记忆）
type ScheduleSettings struct {
	ID                int        `json:"id" db:"id"`
	AdminID           int        `json:"admin_id" db:"admin_id"`
	CurrentWeek       int        `json:"current_week" db:"current_week"`             // 当前周次
	AutoIncrement     bool       `json:"auto_increment" db:"auto_increment"`         // 是否自动递增周次
	NeedPerCell       int        `json:"need_per_cell" db:"need_per_cell"`           // 每时段最大人数，默认2
	MinPerCell        int        `json:"min_per_cell" db:"min_per_cell"`             // 每时段最小人数，默认0
	MaxPerDay         int        `json:"max_per_day" db:"max_per_day"`               // 每人每天最多排几次，默认1
	MaxPerWeek        int        `json:"max_per_week" db:"max_per_week"`             // 每人每周最多排几次，默认2
	ExportTitle       string     `json:"export_title" db:"export_title"`             // 导出Excel标题模板
	SemesterStartDate *time.Time `json:"semester_start_date" db:"semester_start_date"` // 学期起始日
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

// UpdateCurrentWeekRequest 更新当前周次请求
type UpdateCurrentWeekRequest struct {
	CurrentWeek   int  `json:"current_week" binding:"required,min=1,max=30"`
	AutoIncrement bool `json:"auto_increment"`
}

// UpdateScheduleRequest 更新排班请求（添加/删除人员）
type UpdateScheduleRequest struct {
	Week    int   `json:"week" binding:"required"`
	Weekday int   `json:"weekday" binding:"required,min=1,max=5"`
	Period  int   `json:"period" binding:"required,min=1,max=4"`
	AddUserIDs    []int `json:"add_user_ids"`    // 要添加的用户ID
	RemoveUserIDs []int `json:"remove_user_ids"` // 要删除的用户ID
}

// ScheduleCell 排班单元格
type ScheduleCell struct {
	Weekday   int    `json:"weekday"`
	Period    int    `json:"period"`
	UserIDs   []int  `json:"user_ids"`
	UserNames []string `json:"user_names,omitempty"`
}

// ScheduleResult 排班结果
type ScheduleResult struct {
	Week  int              `json:"week"`
	Grid  [][]ScheduleCell `json:"grid"` // 5天 × 4节
}

// ScheduleConfirmRequest 确认排班请求
type ScheduleConfirmRequest struct {
	Week int              `json:"week" binding:"required"`
	Grid [][]ScheduleCell `json:"grid" binding:"required"`
}

// DutyCounter 值班统计
type DutyCounter struct {
	UserID     int       `json:"user_id" db:"user_id"`
	UserName   string    `json:"user_name,omitempty" db:"-"`
	TotalCount int       `json:"total_count" db:"total_count"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// DutyQuery 值班查询
type DutyQuery struct {
	Week    int `form:"week"`
	Weekday int `form:"weekday"`
	UserID  int `form:"user_id"`
}

// DutyRecordWithUser 带用户信息的值班记录
type DutyRecordWithUser struct {
	DutyRecord
	StudentID string `json:"student_id" db:"student_id"`
	UserName  string `json:"user_name" db:"user_name"`
}

// Cell 排班单元格
type Cell struct {
	Weekday int    `json:"weekday"`
	Period  int    `json:"period"`
	Users   []User `json:"users"`
}

// ConflictCell 冲突单元格
type ConflictCell struct {
	Weekday   int `json:"weekday"`
	Period    int `json:"period"`
	Need      int `json:"need"`
	Available int `json:"available"`
}

// SchedulePreview 排班预览
type SchedulePreview struct {
	Week      int            `json:"week"`
	Grid      [][]Cell       `json:"grid"`
	Conflicts []ConflictCell `json:"conflicts"`
	Warnings  []ConflictCell `json:"warnings"` // 警告（人数不足但未达最小要求）
}

// ConfirmCell 确认单元格
type ConfirmCell struct {
	Weekday int   `json:"weekday"`
	Period  int   `json:"period"`
	UserIDs []int `json:"user_ids"`
}

// ConfirmScheduleRequest 确认排班请求
type ConfirmScheduleRequest struct {
	Week  int           `json:"week" binding:"required"`
	Cells []ConfirmCell `json:"cells" binding:"required"`
}

// UpdateDutyStatusRequest 更新值班状态请求
type UpdateDutyStatusRequest struct {
	DutyID int    `json:"duty_id" binding:"required"`
	Status string `json:"status" binding:"required"`
}

// UpdateSemesterStartDateRequest 更新学期起始日请求
type UpdateSemesterStartDateRequest struct {
	SemesterStartDate string `json:"semester_start_date" binding:"required"`
}

// SemesterStartDateResponse 学期起始日响应
type SemesterStartDateResponse struct {
	SemesterStartDate string `json:"semester_start_date"`
	CurrentWeek       int    `json:"current_week"`
}
