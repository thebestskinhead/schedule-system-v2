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

// StartQrLogin 开始扫码登录，返回二维码
func (h *QRLoginHandler) StartQrLogin(c *gin.Context) {
	resp, err := h.service.CreateSession()
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, resp)
}

// PollQrLogin 轮询扫码登录状态
func (h *QRLoginHandler) PollQrLogin(c *gin.Context) {
	var req model.QRPollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	session := h.service.GetSession(req.SessionID)
	if session == nil {
		utils.Success(c, model.QRPollResponse{
			Status:  model.QRStatusExpired,
			Message: "二维码已过期，请重新获取",
		})
		return
	}

	switch session.Status {
	case model.QRStatusSuccess:
		utils.Success(c, model.QRPollResponse{
			Status:  model.QRStatusSuccess,
			Token:   session.Token,
			User:    session.User,
			Message: "登录成功",
		})
		// 登录成功后清理 session
		h.service.DeleteSession(req.SessionID)

	case model.QRStatusNeedReg:
		utils.Success(c, model.QRPollResponse{
			Status:    model.QRStatusNeedReg,
			Message:   "用户未注册，请先注册",
			StudentID: session.StudentID,
			Name:      session.Name,
		})
		// 清理 session
		h.service.DeleteSession(req.SessionID)

	default:
		utils.Success(c, model.QRPollResponse{
			Status:  session.Status,
			Message: session.Message,
		})
	}
}
