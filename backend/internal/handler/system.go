package handler

import (
	"log"
	"schedule-system-v2/backend/internal/config"
	"schedule-system-v2/backend/internal/db"
	"schedule-system-v2/backend/internal/service"
	"schedule-system-v2/backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct {
	service *service.SystemService
}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{
		service: service.NewSystemService(),
	}
}

// GetInstallStatus 获取系统安装状态
func (h *SystemHandler) GetInstallStatus(c *gin.Context) {
	installed := config.IsInstalled()
	utils.Success(c, gin.H{
		"installed": installed,
	})
}

// TestDBConnection 测试数据库连接
func (h *SystemHandler) TestDBConnection(c *gin.Context) {
	var req config.DatabaseConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	if err := h.service.TestDBConnection(&req); err != nil {
		utils.Error(c, 500, "数据库连接失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "连接成功"})
}

// CheckDatabase 检查数据库状态（空/非空）
func (h *SystemHandler) CheckDatabase(c *gin.Context) {
	var req config.DatabaseConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	isEmpty, tables, err := h.service.CheckDatabaseEmpty(&req)
	if err != nil {
		utils.Error(c, 500, "检查数据库失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"empty":   isEmpty,
		"tables":  tables,
		"message": "检查完成",
	})
}

// InitDatabaseTables 初始化数据库表
func (h *SystemHandler) InitDatabaseTables(c *gin.Context) {
	// 检查是否已安装
	if config.IsInstalled() {
		utils.Error(c, 400, "系统已安装，无法重复初始化")
		return
	}

	var req struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		DBName   string `json:"dbname"`
		Charset  string `json:"charset"`
		Force    bool   `json:"force"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	cfg := &config.DatabaseConfig{
		Host:     req.Host,
		Port:     req.Port,
		User:     req.User,
		Password: req.Password,
		DBName:   req.DBName,
		Charset:  req.Charset,
	}

	if err := h.service.InitDatabaseTables(cfg, req.Force); err != nil {
		utils.Error(c, 500, "初始化表失败: "+err.Error())
		return
	}

	if err := h.service.SaveDBConfig(cfg); err != nil {
		utils.Error(c, 500, "保存配置失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "数据库表初始化成功"})
}

// CreateAdmin 创建管理员账号
func (h *SystemHandler) CreateAdmin(c *gin.Context) {
	// 检查是否已安装
	if config.IsInstalled() {
		utils.Error(c, 400, "系统已安装，无法重复创建管理员")
		return
	}

	var req CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	user, err := h.service.CreateAdmin(req.StudentID, req.Name, req.Email, req.Password, req.Department)
	if err != nil {
		utils.Error(c, 500, "创建管理员失败: "+err.Error())
		return
	}

	if err := h.service.MarkInstalled(); err != nil {
		utils.Error(c, 500, "标记安装状态失败: "+err.Error())
		return
	}

	// 安装完成后立即初始化数据库连接（无需重启服务器）
	cfg, err := config.LoadConfig(config.ConfigFilePath)
	if err == nil {
		if err := db.InitDB(&cfg.Database); err != nil {
			// 记录日志但不影响返回成功，下次启动时会重新初始化
			log.Printf("警告: 安装完成后数据库连接初始化失败: %v", err)
		} else {
			log.Println("安装完成，数据库连接已初始化")
		}
	}

	utils.Success(c, gin.H{
		"message": "管理员创建成功，系统安装完成",
		"user":    user,
	})
}

// CreateAdminRequest 创建管理员请求
type CreateAdminRequest struct {
	StudentID  string `json:"studentId" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	Department string `json:"department" binding:"required"`
}
