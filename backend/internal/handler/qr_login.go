package handler

import (
	"github.com/gin-gonic/gin"
	"schedule-system-v2/backend/internal/model"
	"schedule-system-v2/backend/internal/service"
	"schedule-system-v2/backend/internal/utils"
)

type QRLoginHandler struct {
	service *service.QRLoginService
}

func NewQRLoginHandler() *QRLoginHandler {
	return &QRLoginHandler{
		service: service.GetQRLoginService(),
	}
}

// StartQrLogin 开始扫码登录：向教务网获取二维码，返回 session_id
func (h *QRLoginHandler) StartQrLogin(c *gin.Context) {
	resp, err := h.service.CreateSession()
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, resp)
}

// PollQrLogin 前端轮询扫码状态：后端代理请求教务网，返回当前状态
func (h *QRLoginHandler) PollQrLogin(c *gin.Context) {
	var req model.QRPollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	result := h.service.Poll(req.SessionID)

	// success / need_reg 后清理 session
	if result.Status == model.QRStatusSuccess || result.Status == model.QRStatusNeedReg {
		h.service.DeleteSession(req.SessionID)
	}

	utils.Success(c, result)
}
