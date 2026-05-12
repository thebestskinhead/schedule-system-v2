package handler

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"schedule-system-v2/backend/internal/auth"
	"schedule-system-v2/backend/internal/model"
	"schedule-system-v2/backend/internal/service"
	"schedule-system-v2/backend/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type AvailabilityHandler struct {
	service *service.AvailabilityService
}

func NewAvailabilityHandler() *AvailabilityHandler {
	return &AvailabilityHandler{
		service: service.NewAvailabilityService(),
	}
}

func (h *AvailabilityHandler) AddAvailability(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}
	var req model.AddAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	// 提交到任务队列
	queue := service.GetTaskQueue()
	task, err := queue.SubmitManualTask(userID, req.Weekday, req.Period, req.Weeks)
	if err != nil {
		utils.Error(c, http.StatusTooManyRequests, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"task_id":    task.ID,
		"status":     task.Status,
		"message":    "任务已提交，请通过任务ID查询进度",
		"created_at": task.CreatedAt,
	})
}

func (h *AvailabilityHandler) GetMyAvailability(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}
	list, err := h.service.GetMyAvailability(userID)
	if err != nil {
		utils.Error(c, 500, "获取失败")
		return
	}
	utils.Success(c, list)
}

func (h *AvailabilityHandler) GetAllAvailability(c *gin.Context) {
	list, err := h.service.GetAllAvailability()
	if err != nil {
		utils.Error(c, 500, "获取失败")
		return
	}
	utils.Success(c, list)
}

func (h *AvailabilityHandler) DeleteAvailability(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}
	var req struct {
		ID int `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if err := h.service.DeleteAvailability(userID, req.ID); err != nil {
		utils.Error(c, 403, err.Error())
		return
	}
	utils.Success(c, nil)
}

// ImportFromCookie 从Cookie导入课表（异步队列模式）
func (h *AvailabilityHandler) ImportFromCookie(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}

	var req struct {
		Cookies  string `json:"cookies" binding:"required"`
		Semester string `json:"semester" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 提交到任务队列
	queue := service.GetTaskQueue()
	task, err := queue.SubmitTask(userID, req.Cookies, req.Semester)
	if err != nil {
		utils.Error(c, http.StatusTooManyRequests, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"task_id":    task.ID,
		"status":     task.Status,
		"message":    "任务已提交，请通过任务ID查询进度",
		"created_at": task.CreatedAt,
	})
}

// GetImportTaskStatus 获取导入任务状态
func (h *AvailabilityHandler) GetImportTaskStatus(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}
	taskID := c.Query("task_id")

	queue := service.GetTaskQueue()

	var task *service.CookieImportTask
	var found bool

	if taskID != "" {
		task, found = queue.GetTask(taskID)
	} else {
		// 如果没有指定taskID，返回用户最新的任务
		task, found = queue.GetUserTask(userID)
	}

	if !found {
		utils.Error(c, http.StatusNotFound, "任务不存在")
		return
	}

	// 验证用户权限
	if task.UserID != userID {
		utils.Error(c, http.StatusForbidden, "无权查看此任务")
		return
	}

	utils.Success(c, task)
}

// GetImportTaskList 获取用户的导入任务列表
func (h *AvailabilityHandler) GetImportTaskList(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}

	queue := service.GetTaskQueue()
	tasks := queue.GetUserTasks(userID)

	utils.Success(c, gin.H{
		"tasks": tasks,
		"count": len(tasks),
	})
}

// ImportFromXLS 从XLS文件导入课表（异步队列模式）
func (h *AvailabilityHandler) ImportFromXLS(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "请上传XLS文件: "+err.Error())
		return
	}

	// 检查文件扩展名
	ext := filepath.Ext(file.Filename)
	if ext != ".xls" && ext != ".xlsx" {
		utils.Error(c, http.StatusBadRequest, "只支持 .xls 或 .xlsx 文件")
		return
	}

	// 保存临时文件（使用唯一文件名）
	rand.Seed(time.Now().UnixNano())
	tempPath := filepath.Join("/tmp", fmt.Sprintf("schedule_%d_%d%s", userID, rand.Intn(100000), ext))
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		utils.Error(c, http.StatusInternalServerError, "保存文件失败: "+err.Error())
		return
	}

	// 提交到任务队列
	queue := service.GetTaskQueue()
	task, err := queue.SubmitExcelTask(userID, tempPath)
	if err != nil {
		os.Remove(tempPath) // 提交失败则删除临时文件
		utils.Error(c, http.StatusTooManyRequests, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"task_id":    task.ID,
		"status":     task.Status,
		"message":    "任务已提交，请通过任务ID查询进度",
		"created_at": task.CreatedAt,
	})
}

// ImportFromXLSBase64 通过base64编码的XLS文件导入课表（异步队列模式）
func (h *AvailabilityHandler) ImportFromXLSBase64(c *gin.Context) {
	userID := auth.GetUserIDFromContext(c)
	if userID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}

	var req struct {
		Base64   string `json:"base64" binding:"required"`
		FileName string `json:"fileName"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	fileName := req.FileName
	if fileName == "" {
		fileName = "schedule.xlsx"
	}

	ext := filepath.Ext(fileName)
	if ext != ".xls" && ext != ".xlsx" {
		utils.Error(c, http.StatusBadRequest, "只支持 .xls 或 .xlsx 文件")
		return
	}

	// base64 解码
	bytes, err := base64.StdEncoding.DecodeString(req.Base64)
	if err != nil {
		bytes, err = base64.RawStdEncoding.DecodeString(req.Base64)
	}
	if err != nil {
		bytes, err = base64.RawURLEncoding.DecodeString(req.Base64)
	}
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "base64解码失败: "+err.Error())
		return
	}

	// 使用唯一临时文件名防止冲突
	rand.Seed(time.Now().UnixNano())
	tempPath := filepath.Join("/tmp", fmt.Sprintf("schedule_%d_%d%s", userID, rand.Intn(100000), ext))
	if err := os.WriteFile(tempPath, bytes, 0644); err != nil {
		utils.Error(c, http.StatusInternalServerError, "写入临时文件失败: "+err.Error())
		return
	}

	// 提交到任务队列
	queue := service.GetTaskQueue()
	task, err := queue.SubmitExcelTask(userID, tempPath)
	if err != nil {
		os.Remove(tempPath)
		utils.Error(c, http.StatusTooManyRequests, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"task_id":    task.ID,
		"status":     task.Status,
		"message":    "任务已提交，请通过任务ID查询进度",
		"created_at": task.CreatedAt,
	})
}
