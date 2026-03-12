package handler

import (
	"net/http"
	"schedule-system-v2/backend/internal/auth"
	"schedule-system-v2/backend/internal/service"
	"schedule-system-v2/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type CrawlerHandler struct {
	availabilityService *service.AvailabilityService
}

func NewCrawlerHandler() *CrawlerHandler {
	return &CrawlerHandler{
		availabilityService: service.NewAvailabilityService(),
	}
}

// CrawlSchedule 爬取课表并导入无课时间
func (h *CrawlerHandler) CrawlSchedule(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}

	var req service.CrawlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if req.StartWeek == 0 {
		req.StartWeek = 1
	}
	if req.EndWeek == 0 {
		req.EndWeek = 30
	}

	// 创建爬虫
	crawler := service.NewEduCrawler(req.Cookies)

	// 爬取课表
	scheduleMap, err := crawler.CrawlSchedule(req.Semester)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "爬取课表失败: "+err.Error())
		return
	}

	// 转换为无课时间
	availabilities := service.ConvertToAvailability(userID, scheduleMap)

	// 保存到数据库
	if len(availabilities) > 0 {
		// 先删除该用户原有的无课时间记录，避免重复键错误
		if err := h.availabilityService.DeleteByUserID(userID); err != nil {
			utils.Error(c, http.StatusInternalServerError, "删除旧记录失败: "+err.Error())
			return
		}
		
		// 批量插入新的无课时间
		if err := h.availabilityService.CreateBatch(userID, availabilities); err != nil {
			utils.Error(c, http.StatusInternalServerError, "保存无课时间失败: "+err.Error())
			return
		}
	}

	// 统计结果
	totalCells := len(scheduleMap) * 5 * 4 // 周数 × 5天 × 4节
	availableCells := len(availabilities)

	utils.Success(c, gin.H{
		"weeks_parsed":    len(scheduleMap),
		"total_cells":     totalCells,
		"available_cells": availableCells,
		"imported":        len(availabilities),
		"message":         "课表导入成功",
	})
}

// PreviewSchedule 预览课表（不保存）
func (h *CrawlerHandler) PreviewSchedule(c *gin.Context) {
	var req service.CrawlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 设置默认值
	if req.StartWeek == 0 {
		req.StartWeek = 1
	}
	if req.EndWeek == 0 {
		req.EndWeek = 2 // 预览只取前2周，避免太慢
	}

	// 创建爬虫
	crawler := service.NewEduCrawler(req.Cookies)

	// 爬取课表
	scheduleMap, err := crawler.CrawlSchedule(req.Semester)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "爬取课表失败: "+err.Error())
		return
	}

	// 转换为可视化格式
	var preview []map[string]interface{}
	for week := req.StartWeek; week <= req.EndWeek && week <= len(scheduleMap); week++ {
		schedule := scheduleMap[week]
		weekData := map[string]interface{}{
			"week": week,
			"days": []map[string]interface{}{},
		}
		
		days := []string{"周一", "周二", "周三", "周四", "周五"}
		periods := []string{"第一二节", "第三四节", "第五六节", "第七八节"}
		
		for dayIdx := 0; dayIdx < 5; dayIdx++ {
			dayData := map[string]interface{}{
				"day":     days[dayIdx],
				"periods": []map[string]interface{}{},
			}
			for periodIdx := 0; periodIdx < 4; periodIdx++ {
				dayData["periods"] = append(dayData["periods"].([]map[string]interface{}), map[string]interface{}{
					"period":   periods[periodIdx],
					"hasClass": schedule[dayIdx][periodIdx],
					"free":     !schedule[dayIdx][periodIdx],
				})
			}
			weekData["days"] = append(weekData["days"].([]map[string]interface{}), dayData)
		}
		preview = append(preview, weekData)
	}

	utils.Success(c, gin.H{
		"preview": preview,
		"message": "预览模式（仅显示前2周）",
	})
}
