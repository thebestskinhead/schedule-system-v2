package model

import (
	"encoding/json"
	"time"
)

// ApplicationStatus 申请状态
type ApplicationStatus int

const (
	ApplicationStatusPending   ApplicationStatus = 0 // 待审批
	ApplicationStatusProcessing ApplicationStatus = 1 // 审批中
	ApplicationStatusApproved  ApplicationStatus = 2 // 已通过
	ApplicationStatusRejected  ApplicationStatus = 3 // 已拒绝
	ApplicationStatusWithdrawn ApplicationStatus = 4 // 已撤回
)

func (s ApplicationStatus) String() string {
	switch s {
	case ApplicationStatusPending:
		return "待审批"
	case ApplicationStatusProcessing:
		return "审批中"
	case ApplicationStatusApproved:
		return "已通过"
	case ApplicationStatusRejected:
		return "已拒绝"
	case ApplicationStatusWithdrawn:
		return "已撤回"
	default:
		return "未知"
	}
}

// ApprovalAction 审批操作
type ApprovalAction int

const (
	ApprovalActionApprove ApprovalAction = 1 // 通过
	ApprovalActionReject  ApprovalAction = 2 // 拒绝
	ApprovalActionTransfer ApprovalAction = 3 // 转办
	ApprovalActionComment ApprovalAction = 4 // 评论
)

// ApplicationType 申请类型
type ApplicationType struct {
	ID          int             `json:"id" db:"id"`
	Code        string          `json:"code" db:"code"`
	Name        string          `json:"name" db:"name"`
	Description string          `json:"description" db:"description"`
	Config      json.RawMessage `json:"config" db:"config"`
	IsActive    bool            `json:"is_active" db:"is_active"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// TypeConfig 类型配置
type TypeConfig struct {
	Fields      []FieldConfig   `json:"fields"`
	Flow        []FlowStep      `json:"flow"`
	AutoExecute bool            `json:"auto_execute"` // 审批通过后是否自动执行
}

// FieldConfig 字段配置
type FieldConfig struct {
	Name     string      `json:"name"`
	Label    string      `json:"label"`
	Type     string      `json:"type"` // text, textarea, select, date, datetime, number
	Required bool        `json:"required"`
	Options  []Option    `json:"options,omitempty"`
}

// Option 选项
type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// FlowStep 审批流程步骤
type FlowStep struct {
	Level    int    `json:"level"`
	Role     string `json:"role"` // admin, dept_admin, office_admin, specific
	Label    string `json:"label"`
	Optional bool   `json:"optional"`
}

// Scan 实现sql.Scanner接口
func (c *TypeConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

// Application 申请表
type Application struct {
	ID            int                `json:"id" db:"id"`
	ApplicationNo string             `json:"application_no" db:"application_no"`
	TypeCode      string             `json:"type_code" db:"type_code"`
	ApplicantID   int                `json:"applicant_id" db:"applicant_id"`
	Department    string             `json:"department" db:"department"`
	Title         string             `json:"title" db:"title"`
	Content       string             `json:"content" db:"content"`
	Data          json.RawMessage    `json:"data" db:"data"`
	Status        ApplicationStatus  `json:"status" db:"status"`
	CurrentLevel  int                `json:"current_level" db:"current_level"`
	TotalLevels   int                `json:"total_levels" db:"total_levels"`
	ApproverID    *int               `json:"approver_id,omitempty" db:"approver_id"`
	Result        *string            `json:"result,omitempty" db:"result"`  // 改为指针类型支持 NULL
	CreatedAt     time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at" db:"updated_at"`
	
	// 关联字段（查询时使用）
	ApplicantName string `json:"applicant_name,omitempty" db:"applicant_name"`
	TypeName      string `json:"type_name,omitempty" db:"type_name"`
	ApproverName  string `json:"approver_name,omitempty" db:"approver_name"`
}

// ApplicationView 申请视图（带扩展信息）
type ApplicationView struct {
	Application
	TypeConfig    *TypeConfig         `json:"type_config,omitempty"`
	History       []ApprovalHistory   `json:"history,omitempty"`
}

// ApprovalHistory 审批历史
type ApplicationApproval struct {
	ID            int             `json:"id" db:"id"`
	ApplicationID int             `json:"application_id" db:"application_id"`
	Level         int             `json:"level" db:"level"`
	ApproverID    int             `json:"approver_id" db:"approver_id"`
	Action        ApprovalAction  `json:"action" db:"action"`
	Comment       string          `json:"comment" db:"comment"`
	Data          json.RawMessage `json:"data,omitempty" db:"data"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
}

// ApprovalHistory 审批历史视图
type ApprovalHistory struct {
	ID            int            `json:"id"`
	Level         int            `json:"level"`
	ApproverID    int            `json:"approver_id"`
	ApproverName  string         `json:"approver_name"`
	Action        string         `json:"action"`
	Comment       string         `json:"comment"`
	CreatedAt     time.Time      `json:"created_at"`
}

// ApplicationApprover 审批人配置
type ApplicationApprover struct {
	ID             int       `json:"id" db:"id"`
	TypeCode       string    `json:"type_code" db:"type_code"`
	Level          int       `json:"level" db:"level"`
	RoleType       string    `json:"role_type" db:"role_type"`
	SpecificUserID *int      `json:"specific_user_id,omitempty" db:"specific_user_id"`
	IsActive       bool      `json:"is_active" db:"is_active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// ========== 请求/响应模型 ==========

// CreateApplicationRequest 创建申请请求
type CreateApplicationRequest struct {
	TypeCode string                 `json:"type_code" binding:"required"`
	Title    string                 `json:"title" binding:"required"`
	Content  string                 `json:"content"`
	Data     map[string]interface{} `json:"data" binding:"required"`
}

// ApproveApplicationRequest 审批请求
type ApproveApplicationRequest struct {
	Action  ApprovalAction `json:"action" binding:"required"`
	Comment string         `json:"comment"`
	Data    map[string]interface{} `json:"data"`
}

// ApplicationListRequest 申请列表查询请求
type ApplicationListRequest struct {
	TypeCode   string            `form:"type_code"`
	Status     ApplicationStatus `form:"status"`
	ApplicantID int              `form:"applicant_id"`
	Department string            `form:"department"`
	Page       int               `form:"page,default=1"`
	PageSize   int               `form:"page_size,default=20"`
}

// ApplicationTypeResponse 申请类型响应
type ApplicationTypeResponse struct {
	Code        string      `json:"code"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Config      *TypeConfig `json:"config"`
}

// PendingCountResponse 待审批数量响应
type PendingCountResponse struct {
	Total      int            `json:"total"`
	ByType     map[string]int `json:"by_type"`
}

// ApplicationExecutor 申请执行器接口（不同类型申请实现此接口）
type ApplicationExecutor interface {
	// GetTypeCode 返回处理的申请类型代码
	GetTypeCode() string
	
	// Validate 验证申请数据
	Validate(data map[string]interface{}) error
	
	// Execute 审批通过后执行操作
	Execute(application *Application) error
	
	// GetTitle 生成申请标题
	GetTitle(data map[string]interface{}) string
}
