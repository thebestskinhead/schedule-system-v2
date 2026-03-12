package handler

import (
	"schedule-system-v2/backend/internal/auth"
	"schedule-system-v2/backend/internal/model"
	"schedule-system-v2/backend/internal/service"
	"schedule-system-v2/backend/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TemplateHandler struct {
	service *service.TemplateService
}

func NewTemplateHandler() *TemplateHandler {
	return &TemplateHandler{
		service: service.NewTemplateService(),
	}
}

// GetTemplates 获取所有模板
func (h *TemplateHandler) GetTemplates(c *gin.Context) {
	adminID := auth.GetUserIDFromContext(c)
	if adminID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}
	templates, err := h.service.GetAllTemplates(adminID)
	if err != nil {
		utils.Error(c, 500, "获取模板列表失败")
		return
	}
	utils.Success(c, templates)
}

// GetTemplate 获取单个模板
func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error(c, 400, "无效模板ID")
		return
	}

	template, err := h.service.GetTemplateByID(id)
	if err != nil {
		utils.Error(c, 404, "模板不存在")
		return
	}
	utils.Success(c, template)
}

// CreateTemplate 创建模板
func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	adminID := auth.GetUserIDFromContext(c)
	if adminID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}
	var req model.CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	template, err := h.service.CreateTemplate(adminID, &req)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, template)
}

// UpdateTemplate 更新模板
func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	adminID := auth.GetUserIDFromContext(c)
	if adminID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}
	var req model.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	template, err := h.service.UpdateTemplate(adminID, &req)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, template)
}

// DeleteTemplate 删除模板
func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	adminID := auth.GetUserIDFromContext(c)
	if adminID == 0 {
		auth.ResponseUnauthorized(c)
		return
	}
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error(c, 400, "无效模板ID")
		return
	}

	if err := h.service.DeleteTemplate(adminID, id); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}

// GetPlaceholderHelp 获取占位符帮助
func (h *TemplateHandler) GetPlaceholderHelp(c *gin.Context) {
	help := h.service.GetPlaceholderHelp()
	utils.Success(c, help)
}
