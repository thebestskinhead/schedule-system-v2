package model

import "time"

// QRStatus 扫码登录状态
type QRStatus string

const (
	QRStatusPending   QRStatus = "pending"
	QRStatusScanned   QRStatus = "scanned"
	QRStatusConfirmed QRStatus = "confirmed"
	QRStatusSuccess   QRStatus = "success"
	QRStatusNeedReg   QRStatus = "need_reg"
	QRStatusExpired   QRStatus = "expired"
	QRStatusError     QRStatus = "error"
)

// QRStartResponse 获取二维码响应
type QRStartResponse struct {
	SessionID string `json:"session_id"`
	QRCode    string `json:"qrcode"`
	ExpiresIn int    `json:"expires_in"`
}

// QRPollRequest 轮询请求
type QRPollRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

// QRPollResponse 轮询响应
type QRPollResponse struct {
	Status    QRStatus `json:"status"`
	Message   string   `json:"message,omitempty"`
	Token     string   `json:"token,omitempty"`
	User      UserInfo `json:"user,omitempty"`
	StudentID string   `json:"student_id,omitempty"`
	Name      string   `json:"name,omitempty"`
}

// EduUserInfo 从教务网获取的用户信息
type EduUserInfo struct {
	StudentID string
	Name      string
}

// QRSession 扫码登录会话（内存存储）
type QRSession struct {
	SessionID    string
	StudentID    string
	Name         string
	Status       QRStatus
	Message      string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	Token        string
	User         UserInfo
	StopCh       chan struct{}
	DoneCh       chan struct{}
}
