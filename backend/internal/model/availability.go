package model

import "time"

// Availability 无课时间模型
type Availability struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Week      int       `json:"week" db:"week"`
	Weekday   int       `json:"weekday" db:"weekday"`
	Period    int       `json:"period" db:"period"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// AvailabilityInput 无课时间输入（手动录入用）
type AvailabilityInput struct {
	Weekday int   `json:"weekday" binding:"required,min=1,max=5"` // 星期几
	Period  int   `json:"period" binding:"required,min=1,max=4"`  // 节次
	Weeks   []int `json:"weeks" binding:"required"`                // 哪些周无课 [1,2,3...]
}

// AvailabilityBatchInput 批量录入
type AvailabilityBatchInput struct {
	Items []AvailabilityInput `json:"items" binding:"required"`
}

// AvailabilityView 无课时间展示（按周查看）
type AvailabilityView struct {
	Week     int                `json:"week"`
	Schedule [5][4]bool         `json:"schedule"` // [weekday][period] true=有空
}

// WeekRange 周次范围
type WeekRange struct {
	MinWeek int `json:"min_week"`
	MaxWeek int `json:"max_week"`
}

// AddAvailabilityRequest 添加无课时间请求
type AddAvailabilityRequest struct {
	Weekday int   `json:"weekday" binding:"required,min=1,max=5"`
	Period  int   `json:"period" binding:"required,min=1,max=4"`
	Weeks   []int `json:"weeks" binding:"required"`
}

// AvailabilityWithUser 带用户信息的无课时间
type AvailabilityWithUser struct {
	Availability
	StudentID string `json:"student_id" db:"student_id"`
	UserName  string `json:"user_name" db:"user_name"`
}
