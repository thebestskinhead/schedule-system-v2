package handler

import (
	"bytes"
	"fmt"
	"schedule-system-v2/backend/internal/auth"
	"schedule-system-v2/backend/internal/model"
	"schedule-system-v2/backend/internal/service"
	"schedule-system-v2/backend/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type ScheduleHandler struct {
	service         *service.ScheduleService
	templateService *service.TemplateService
}

func NewScheduleHandler() *ScheduleHandler {
	return &ScheduleHandler{
		service:         service.NewScheduleService(),
		templateService: service.NewTemplateService(),
	}
}

func (h *ScheduleHandler) PreviewSchedule(c *gin.Context) {
	var req model.ScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	// 检查是否有权限为该部门排班
	if !auth.CheckDeptPermission(c, auth.PermScheduleManageDept, req.Department) {
		return
	}

	preview, err := h.service.PreviewSchedule(&req)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, preview)
}

func (h *ScheduleHandler) ConfirmSchedule(c *gin.Context) {
	adminID := auth.GetUserIDFromContext(c)
	if adminID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}
	var req model.ConfirmScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if err := h.service.ConfirmSchedule(adminID, &req); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}

func (h *ScheduleHandler) GetSchedule(c *gin.Context) {
	weekStr := c.Query("week")
	week, err := strconv.Atoi(weekStr)
	if err != nil || week == 0 {
		utils.Error(c, 400, "请提供周次")
		return
	}

	records, err := h.service.GetScheduleByWeek(week)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, records)
}

func (h *ScheduleHandler) GetMyDuties(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}
	records, err := h.service.GetMyDuties(userID)
	if err != nil {
		utils.Error(c, 500, "获取失败")
		return
	}
	utils.Success(c, records)
}

func (h *ScheduleHandler) UpdateDutyStatus(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}
	var req model.UpdateDutyStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if err := h.service.UpdateDutyStatus(userID, req.DutyID, req.Status); err != nil {
		utils.Error(c, 403, err.Error())
		return
	}
	utils.Success(c, nil)
}

// GetScheduleSettings 获取排班设置
func (h *ScheduleHandler) GetScheduleSettings(c *gin.Context) {
	adminID := auth.GetUserIDFromContext(c)
	if adminID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}
	settings, err := h.service.GetScheduleSettings(adminID)
	if err != nil {
		// 返回默认设置
		utils.Success(c, model.ScheduleSettings{
			NeedPerCell: 2,
			MinPerCell:  0,
			MaxPerDay:   1,
			MaxPerWeek:  2,
			ExportTitle: "第{week}周排班表",
		})
		return
	}
	utils.Success(c, settings)
}

// SaveScheduleSettings 保存排班设置
func (h *ScheduleHandler) SaveScheduleSettings(c *gin.Context) {
	adminID := auth.GetUserIDFromContext(c)
	if adminID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}
	var req model.ScheduleSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if err := h.service.SaveScheduleSettings(adminID, &req); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}

// UpdateSchedule 更新排班（添加/删除人员）
func (h *ScheduleHandler) UpdateSchedule(c *gin.Context) {
	adminID := auth.GetUserIDFromContext(c)
	if adminID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}
	var req model.UpdateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if err := h.service.UpdateSchedule(adminID, &req); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}

// ExportSchedule 导出排班表
func (h *ScheduleHandler) ExportSchedule(c *gin.Context) {
	var req model.ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 获取排班预览数据
	previewReq := model.ScheduleRequest{
		Week:        req.Week,
		Days:        []int{1, 2, 3, 4, 5},
		Periods:     4,
		NeedPerCell: 2,
	}

	preview, err := h.service.PreviewSchedule(&previewReq)
	if err != nil {
		// 如果没有预览数据，尝试获取已确认的排班
		records, err := h.service.GetScheduleByWeek(req.Week)
		if err != nil || len(records) == 0 {
			utils.Error(c, 404, "没有找到排班数据")
			return
		}
		// 转换为grid格式
		preview = h.convertRecordsToPreview(req.Week, records)
	}

	// 获取导出配置
	exportResult, err := h.templateService.ExportData(req.Week, req.TemplateID, req.Department)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	// 构建Excel数据
	excelData := exportResult.BuildExcelData(preview.Grid)

	// 创建Excel文件
	f := excelize.NewFile()
	sheetName := fmt.Sprintf("第%d周排班", req.Week)
	f.SetSheetName("Sheet1", sheetName)

	// 写入数据
	for rowIdx, row := range excelData {
		for colIdx, val := range row {
			cell := fmt.Sprintf("%s%d", string(rune('A'+colIdx)), rowIdx+1)
			f.SetCellValue(sheetName, cell, val)
		}
	}

	// 设置列宽
	if exportResult.Config.Mode == "schedule" && exportResult.Config.ScheduleConfig != nil {
		// 课表模式：第一列较窄，其他列较宽
		f.SetColWidth(sheetName, "A", "A", 12)
		for i := 1; i <= len(exportResult.Config.ScheduleConfig.ColLabels); i++ {
			col := string(rune('A' + i))
			f.SetColWidth(sheetName, col, col, 20)
		}
	} else if len(exportResult.Config.Headers) > 0 {
		// 列表模式
		for i := range exportResult.Config.Headers {
			col := string(rune('A' + i))
			f.SetColWidth(sheetName, col, col, 20)
		}
	}

	// 写入buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		utils.Error(c, 500, "生成Excel失败")
		return
	}

	// 返回文件
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=排班表_第%d周.xlsx", req.Week))
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}

// 将值班记录转换为预览格式
func (h *ScheduleHandler) convertRecordsToPreview(week int, records []model.DutyRecordWithUser) *model.SchedulePreview {
	grid := make([][]model.Cell, 5)
	for i := range grid {
		grid[i] = make([]model.Cell, 4)
	}

	for _, r := range records {
		if r.Weekday < 1 || r.Weekday > 5 || r.Period < 1 || r.Period > 4 {
			continue
		}
		cell := &grid[r.Weekday-1][r.Period-1]
		cell.Weekday = r.Weekday
		cell.Period = r.Period
		cell.Users = append(cell.Users, model.User{
			ID:   r.UserID,
			Name: r.UserName,
		})
	}

	return &model.SchedulePreview{
		Week: week,
		Grid: grid,
	}
}

// GetCurrentWeek 获取当前周次（公开接口）
func (h *ScheduleHandler) GetCurrentWeek(c *gin.Context) {
	week, autoIncrement, err := h.service.GetCurrentWeek()
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, gin.H{
		"current_week":   week,
		"auto_increment": autoIncrement,
	})
}

// UpdateCurrentWeek 更新当前周次（仅管理员）
func (h *ScheduleHandler) UpdateCurrentWeek(c *gin.Context) {
	adminID := auth.GetUserIDFromContext(c)
	if adminID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}
	var req model.UpdateCurrentWeekRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	if err := h.service.UpdateCurrentWeek(adminID, req.CurrentWeek, req.AutoIncrement); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}

// GetCurrentWeekSchedule 获取当前周的排班（公开接口）
func (h *ScheduleHandler) GetCurrentWeekSchedule(c *gin.Context) {
	week, _, err := h.service.GetCurrentWeek()
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	records, err := h.service.GetScheduleByWeek(week)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, gin.H{
		"week":    week,
		"records": records,
	})
}

// GetSemesterStartDate 获取学期起始日（公开接口）
func (h *ScheduleHandler) GetSemesterStartDate(c *gin.Context) {
	startDate, err := h.service.GetSemesterStartDate()
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	// 获取当前周次
	currentWeek := 1
	if startDate != nil && *startDate != "" {
		week, err := h.service.CalculateCurrentWeek(*startDate)
		if err == nil {
			currentWeek = week
		}
	}

	response := model.SemesterStartDateResponse{
		CurrentWeek: currentWeek,
	}
	if startDate != nil {
		response.SemesterStartDate = *startDate
	}

	utils.Success(c, response)
}

// UpdateSemesterStartDate 更新学期起始日（仅管理员）
func (h *ScheduleHandler) UpdateSemesterStartDate(c *gin.Context) {
	checker := auth.GetChecker(c)
	if !checker.HasPermission(auth.PermScheduleSettings) {
		auth.ResponseForbidden(c, "无权修改学期设置")
		return
	}

	var req model.UpdateSemesterStartDateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	adminID := auth.GetUserIDFromContext(c)
	if err := h.service.UpdateSemesterStartDate(adminID, req.SemesterStartDate); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	// 计算新的当前周次
	currentWeek, _ := h.service.CalculateCurrentWeek(req.SemesterStartDate)

	utils.Success(c, gin.H{
		"current_week": currentWeek,
	})
}
