package model

import (
	"encoding/json"
	"time"
)

// ExportTemplate 导出模板
type ExportTemplate struct {
	ID          int             `json:"id" db:"id"`
	AdminID     int             `json:"admin_id" db:"admin_id"`
	Name        string          `json:"name" db:"name"`
	Description string          `json:"description" db:"description"`
	Config      json.RawMessage `json:"config" db:"config"`
	IsDefault   bool            `json:"is_default" db:"is_default"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// ExportTemplateConfig 导出模板配置
type ExportTemplateConfig struct {
	Title       string                     `json:"title"`       // 表格标题，支持占位符
	Headers     []string                   `json:"headers"`     // 表头（列表模式）
	DataColumns []ExportTemplateColumn     `json:"dataColumns"` // 数据列配置（列表模式）
	Placeholders map[string]string          `json:"placeholders"` // 占位符说明
	CustomRows  []ExportTemplateCustomRow  `json:"customRows,omitempty"` // 自定义行（用于复杂模板）
	Mode        string                     `json:"mode"`        // 模式: list(列表) 或 schedule(课表矩阵)
	// 课表模式专用配置
	ScheduleConfig *ScheduleTableConfig    `json:"scheduleConfig,omitempty"`
}

// ScheduleTableConfig 课表模式配置
type ScheduleTableConfig struct {
	RowHeader      string   `json:"rowHeader"`      // 行标题，如"节次"
	ColHeader      string   `json:"colHeader"`      // 列标题，如"星期"
	RowLabels      []string `json:"rowLabels"`      // 行标签，如 ["第1节","第2节"...]
	ColLabels      []string `json:"colLabels"`      // 列标签，如 ["周一","周二"...]
	CellFormat     string   `json:"cellFormat"`     // 单元格内容格式，如"{users}" 或 "{users}\n({count}人)"
	EmptyCellText  string   `json:"emptyCellText"`  // 空单元格显示文本，默认"-"
}

// ExportTemplateColumn 数据列配置
type ExportTemplateColumn struct {
	Type      string `json:"type"`      // 类型: weekday, period, users, text
	Format    string `json:"format"`    // 格式模板，如 "周{weekday_cn}"
	Separator string `json:"separator,omitempty"` // 列表分隔符，如 "、"
	Value     string `json:"value,omitempty"`     // 固定值（当type=text时使用）
}

// ExportTemplateCustomRow 自定义行配置
type ExportTemplateCustomRow struct {
	RowIndex    int                  `json:"rowIndex"`    // 行号
	IsHeader    bool                 `json:"isHeader"`    // 是否为表头行
	Cells       []ExportTemplateCell `json:"cells"`       // 单元格列表
}

// ExportTemplateCell 模板单元格
type ExportTemplateCell struct {
	ColIndex int    `json:"colIndex"` // 列号
	Value    string `json:"value"`    // 值（支持占位符）
	IsMerge  bool   `json:"isMerge"`  // 是否合并单元格
	MergeCol int    `json:"mergeCol"` // 合并到第几列
}

// ExportRequest 导出请求
type ExportRequest struct {
	Week         int    `json:"week" binding:"required"`
	TemplateID   int    `json:"template_id"` // 0表示使用默认模板
	Department   string `json:"department,omitempty"` // 部门名称
	CustomTitle  string `json:"custom_title,omitempty"` // 自定义标题
}

// CreateTemplateRequest 创建模板请求
type CreateTemplateRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Config      ExportTemplateConfig   `json:"config" binding:"required"`
	IsDefault   bool                   `json:"is_default"`
}

// UpdateTemplateRequest 更新模板请求
type UpdateTemplateRequest struct {
	ID          int                    `json:"id" binding:"required"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Config      *ExportTemplateConfig  `json:"config,omitempty"`
	IsDefault   *bool                  `json:"is_default,omitempty"`
}
