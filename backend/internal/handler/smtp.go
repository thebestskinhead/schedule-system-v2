package handler

import (
	"schedule-system-v2/backend/internal/config"
	"schedule-system-v2/backend/internal/model"
	"schedule-system-v2/backend/internal/service"
	"schedule-system-v2/backend/internal/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type SMTPHandler struct {
	service *service.SMTPService
}

func NewSMTPHandler() *SMTPHandler {
	return &SMTPHandler{
		service: service.NewSMTPService(),
	}
}

// GetConfigs 获取所有SMTP配置（管理员）
func (h *SMTPHandler) GetConfigs(c *gin.Context) {
	configs, err := h.service.GetAllConfigs()
	if err != nil {
		utils.Success(c, []model.SMTPConfig{})
		return
	}
	utils.Success(c, configs)
}

// GetActiveConfig 获取当前启用的配置（管理员）
func (h *SMTPHandler) GetActiveConfig(c *gin.Context) {
	config, err := h.service.GetConfig()
	if err != nil {
		utils.Error(c, 404, "未配置SMTP")
		return
	}
	utils.Success(c, config)
}

// SaveConfig 保存SMTP配置（管理员）
func (h *SMTPHandler) SaveConfig(c *gin.Context) {
	var req model.SMTPConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "SMTP 配置不完整或端口无效，请检查后重试")
		return
	}

	req.Host = strings.TrimSpace(req.Host)
	req.Username = strings.TrimSpace(req.Username)
	req.From = strings.TrimSpace(req.From)
	req.FromEmail = utils.NormalizeEmail(req.FromEmail)

	if req.Host == "" || req.Username == "" || req.From == "" || req.FromEmail == "" {
		utils.Error(c, 400, "SMTP 配置不完整或端口无效，请检查后重试")
		return
	}
	if !utils.IsEmail(req.FromEmail) {
		utils.Error(c, 400, "发件人邮箱格式不正确")
		return
	}

	if err := h.service.SaveConfig(&req); err != nil {
		utils.Error(c, 500, "保存失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "保存成功"})
}

// DeleteConfig 删除配置（管理员）
func (h *SMTPHandler) DeleteConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if err := h.service.DeleteConfig(id); err != nil {
		utils.Error(c, 500, "删除失败")
		return
	}

	utils.Success(c, gin.H{"message": "删除成功"})
}

// TestEmail 发送测试邮件（管理员）
func (h *SMTPHandler) TestEmail(c *gin.Context) {
	var req model.SMTPTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "请填写收件邮箱")
		return
	}

	to := utils.NormalizeEmail(req.To)
	if !utils.IsEmail(to) {
		utils.Error(c, 400, "收件邮箱格式不正确")
		return
	}

	if err := h.service.TestEmail(to); err != nil {
		utils.Error(c, 500, "发送失败: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "测试邮件已发送"})
}

// SendPasswordReset 发送密码重置邮件（公开）
func (h *SMTPHandler) SendPasswordReset(c *gin.Context) {
	var req model.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: 请提供email或student_id")
		return
	}

	// 验证至少提供一个参数
	if req.Email == "" && req.StudentID == "" {
		utils.Error(c, 400, "参数错误: 请提供email或student_id")
		return
	}

	// 无论用户是否存在，都返回成功（防止枚举攻击）
	h.service.SendPasswordResetEmail(req.Email, req.StudentID)

	utils.Success(c, gin.H{"message": "如果该用户已注册，重置邮件已发送"})
}

// VerifyResetToken 验证重置令牌（公开）
func (h *SMTPHandler) VerifyResetToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		utils.Error(c, 400, "缺少令牌")
		return
	}

	t, err := h.service.VerifyResetToken(token)
	if err != nil {
		utils.Error(c, 400, "令牌无效或已过期")
		return
	}

	utils.Success(c, gin.H{"email": t.Email})
}

// ResetPassword 重置密码（公开）
func (h *SMTPHandler) ResetPassword(c *gin.Context) {
	var req model.PasswordResetConfirm
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	if err := h.service.ResetPassword(req.Token, req.Password); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "密码重置成功"})
}

// CheckConfig 检查SMTP是否配置（公开）
func (h *SMTPHandler) CheckConfig(c *gin.Context) {
	exists := h.service.CheckConfigExists()
	utils.Success(c, gin.H{"configured": exists})
}

// GetSiteConfig 获取网站配置（管理员）
func (h *SMTPHandler) GetSiteConfig(c *gin.Context) {
	cfg := config.GetConfig()
	utils.Success(c, gin.H{"domain": cfg.Site.Domain})
}

// SaveSiteConfig 保存网站配置（管理员）
func (h *SMTPHandler) SaveSiteConfig(c *gin.Context) {
	var req struct {
		Domain string `json:"domain" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}

	cfg, err := config.LoadConfig(config.ConfigFilePath)
	if err != nil {
		utils.Error(c, 500, "加载配置失败")
		return
	}

	cfg.Site.Domain = req.Domain
	if err := config.SaveConfig(cfg); err != nil {
		utils.Error(c, 500, "保存配置失败")
		return
	}

	utils.Success(c, gin.H{"message": "保存成功"})
}
