package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"schedule-system-v2/backend/internal/dao"
	"schedule-system-v2/backend/internal/model"
	"strings"
)

type TemplateService struct {
	templateDAO *dao.TemplateDAO
}

func NewTemplateService() *TemplateService {
	return &TemplateService{
		templateDAO: dao.NewTemplateDAO(),
	}
}

// GetAllTemplates 获取所有模板
func (s *TemplateService) GetAllTemplates(adminID int) ([]model.ExportTemplate, error) {
	return s.templateDAO.GetAllTemplates(adminID)
}

// GetTemplateByID 根据ID获取模板
func (s *TemplateService) GetTemplateByID(id int) (*model.ExportTemplate, error) {
	return s.templateDAO.GetTemplateByID(id)
}

// GetDefaultTemplate 获取默认模板
func (s *TemplateService) GetDefaultTemplate() (*model.ExportTemplate, error) {
	return s.templateDAO.GetDefaultTemplate()
}

// CreateTemplate 创建模板
func (s *TemplateService) CreateTemplate(adminID int, req *model.CreateTemplateRequest) (*model.ExportTemplate, error) {
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return nil, err
	}

	template := &model.ExportTemplate{
		AdminID:     adminID,
		Name:        req.Name,
		Description: req.Description,
		Config:      configJSON,
		IsDefault:   req.IsDefault,
	}

	// 如果设置为默认，清除其他默认标记
	if req.IsDefault {
		s.templateDAO.ClearDefaultTemplates()
	}

	if err := s.templateDAO.CreateTemplate(template); err != nil {
		return nil, err
	}
	return template, nil
}

// UpdateTemplate 更新模板
func (s *TemplateService) UpdateTemplate(adminID int, req *model.UpdateTemplateRequest) (*model.ExportTemplate, error) {
	template, err := s.templateDAO.GetTemplateByID(req.ID)
	if err != nil {
		return nil, errors.New("模板不存在")
	}

	// 检查权限（只能修改自己创建的模板，但默认模板允许任何管理员修改）
	if template.AdminID != adminID && !template.IsDefault {
		return nil, errors.New("无权修改此模板")
	}

	if req.Name != "" {
		template.Name = req.Name
	}
	if req.Description != "" {
		template.Description = req.Description
	}
	if req.Config != nil {
		configJSON, err := json.Marshal(req.Config)
		if err != nil {
			return nil, err
		}
		template.Config = configJSON
	}
	if req.IsDefault != nil {
		if *req.IsDefault {
			s.templateDAO.ClearDefaultTemplates()
		}
		template.IsDefault = *req.IsDefault
	}

	if err := s.templateDAO.UpdateTemplate(template); err != nil {
		return nil, err
	}
	return template, nil
}

// DeleteTemplate 删除模板
func (s *TemplateService) DeleteTemplate(adminID int, id int) error {
	template, err := s.templateDAO.GetTemplateByID(id)
	if err != nil {
		return errors.New("模板不存在")
	}
	if template.IsDefault {
		return errors.New("不能删除默认模板")
	}
	if template.AdminID != adminID {
		return errors.New("无权删除此模板")
	}
	return s.templateDAO.DeleteTemplate(id)
}

// ExportData 根据模板导出数据
func (s *TemplateService) ExportData(week int, templateID int, department string) (*ExportResult, error) {
	var template *model.ExportTemplate
	var err error

	if templateID > 0 {
		template, err = s.templateDAO.GetTemplateByID(templateID)
	} else {
		template, err = s.templateDAO.GetDefaultTemplate()
	}

	if err != nil {
		return nil, errors.New("获取模板失败")
	}

	config, err := dao.ParseTemplateConfig(template)
	if err != nil {
		return nil, errors.New("解析模板配置失败")
	}

	// 设置默认模式
	if config.Mode == "" {
		config.Mode = "list"
	}

	return &ExportResult{
		Config:     config,
		Week:       week,
		Department: department,
	}, nil
}

// ExportResult 导出结果
type ExportResult struct {
	Config     *model.ExportTemplateConfig
	Week       int
	Department string
}

// BuildExcelData 构建Excel数据
func (r *ExportResult) BuildExcelData(grid [][]model.Cell) [][]interface{} {
	// 课表模式
	if r.Config.Mode == "schedule" {
		return r.buildScheduleTableData(grid)
	}
	// 列表模式（默认）
	return r.buildListTableData(grid)
}

// buildListTableData 构建列表格式数据
func (r *ExportResult) buildListTableData(grid [][]model.Cell) [][]interface{} {
	var data [][]interface{}

	// 添加标题行
	if r.Config.Title != "" {
		title := r.replacePlaceholders(r.Config.Title, 0, 0, nil)
		data = append(data, []interface{}{title})
		data = append(data, []interface{}{}) // 空行
	}

	// 添加表头
	if len(r.Config.Headers) > 0 {
		headers := make([]interface{}, len(r.Config.Headers))
		for i, h := range r.Config.Headers {
			headers[i] = h
		}
		data = append(data, headers)
	}

	// 添加数据行
	weekdayNames := []string{"一", "二", "三", "四", "五"}
	for day := 0; day < 5; day++ {
		for period := 0; period < 4; period++ {
			cell := grid[day][period]
			if len(cell.Users) == 0 {
				continue
			}

			row := make([]interface{}, len(r.Config.DataColumns))
			for i, col := range r.Config.DataColumns {
				row[i] = r.formatCellValue(col, day+1, period+1, cell.Users, weekdayNames)
			}
			data = append(data, row)
		}
	}

	return data
}

// buildScheduleTableData 构建课表矩阵格式数据
func (r *ExportResult) buildScheduleTableData(grid [][]model.Cell) [][]interface{} {
	var data [][]interface{}
	cfg := r.Config.ScheduleConfig
	if cfg == nil {
		cfg = &model.ScheduleTableConfig{
			RowHeader:     "节次",
			ColHeader:     "星期",
			RowLabels:     []string{"第1节", "第2节", "第3节", "第4节"},
			ColLabels:     []string{"周一", "周二", "周三", "周四", "周五"},
			CellFormat:    "{users}",
			EmptyCellText: "-",
		}
	}

	// 添加标题行
	if r.Config.Title != "" {
		title := r.replacePlaceholders(r.Config.Title, 0, 0, nil)
		data = append(data, []interface{}{title})
		data = append(data, []interface{}{}) // 空行
	}

	// 构建课表表头（第一列是节次标签，后面是星期）
	headerRow := []interface{}{cfg.RowHeader + "\\" + cfg.ColHeader}
	for _, colLabel := range cfg.ColLabels {
		headerRow = append(headerRow, colLabel)
	}
	data = append(data, headerRow)

	// 构建数据行
	for periodIdx, rowLabel := range cfg.RowLabels {
		row := []interface{}{rowLabel}
		for dayIdx := 0; dayIdx < len(cfg.ColLabels); dayIdx++ {
			if dayIdx < 5 && periodIdx < 4 {
				cell := grid[dayIdx][periodIdx]
				cellText := r.formatScheduleCell(cell.Users, cfg)
				row = append(row, cellText)
			} else {
				row = append(row, cfg.EmptyCellText)
			}
		}
		data = append(data, row)
	}

	return data
}

// formatScheduleCell 格式化课表单元格
func (r *ExportResult) formatScheduleCell(users []model.User, cfg *model.ScheduleTableConfig) string {
	if len(users) == 0 {
		return cfg.EmptyCellText
	}

	userNames := make([]string, len(users))
	for i, u := range users {
		userNames[i] = u.Name
	}

	format := cfg.CellFormat
	if format == "" {
		format = "{users}"
	}

	result := format
	result = strings.ReplaceAll(result, "{users}", strings.Join(userNames, "、"))
	result = strings.ReplaceAll(result, "{count}", fmt.Sprintf("%d", len(users)))
	return result
}

// 格式化单元格值
func (r *ExportResult) formatCellValue(col model.ExportTemplateColumn, weekday, period int, users []model.User, weekdayNames []string) string {
	userNames := make([]string, len(users))
	for i, u := range users {
		userNames[i] = u.Name
	}

	switch col.Type {
	case "weekday":
		value := fmt.Sprintf("%d", weekday)
		result := strings.ReplaceAll(col.Format, "{weekday}", value)
		result = strings.ReplaceAll(result, "{weekday_cn}", weekdayNames[weekday-1])
		return result
	case "period":
		value := fmt.Sprintf("%d", period)
		result := strings.ReplaceAll(col.Format, "{period}", value)
		return result
	case "users":
		separator := col.Separator
		if separator == "" {
			separator = "、"
		}
		usersStr := strings.Join(userNames, separator)
		result := strings.ReplaceAll(col.Format, "{users}", usersStr)
		return result
	case "text":
		return r.replacePlaceholders(col.Value, weekday, period, users)
	default:
		return ""
	}
}

// 替换占位符
func (r *ExportResult) replacePlaceholders(template string, weekday, period int, users []model.User) string {
	result := template
	result = strings.ReplaceAll(result, "{week}", fmt.Sprintf("%d", r.Week))
	result = strings.ReplaceAll(result, "{department}", r.Department)
	result = strings.ReplaceAll(result, "{weekday}", fmt.Sprintf("%d", weekday))
	result = strings.ReplaceAll(result, "{period}", fmt.Sprintf("%d", period))
	
	if users != nil {
		userNames := make([]string, len(users))
		for i, u := range users {
			userNames[i] = u.Name
		}
		result = strings.ReplaceAll(result, "{users}", strings.Join(userNames, "、"))
	}
	
	return result
}

// GetPlaceholderHelp 获取占位符帮助信息
func (s *TemplateService) GetPlaceholderHelp() map[string]string {
	return map[string]string{
		"{week}":       "周次数字，如：1, 2, 3...",
		"{department}": "部门名称，导出时传入",
		"{weekday}":    "星期数字(1-5)",
		"{weekday_cn}": "星期中文(一、二、三、四、五)",
		"{period}":     "节次数字(1-4)",
		"{users}":      "值班人员姓名列表",
		"{date}":       "当前日期",
	}
}
