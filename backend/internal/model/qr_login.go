package model

import (
	"net/http/cookiejar"
	"sync"
	"time"
)

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
// 每次前端 poll 时，由后端代理转发到教务网，使用本 session 的 CookieJar
type QRSession struct {
	SessionID string
	UUID      string // 教务网扫码会话 UUID
	CookieJar *cookiejar.Jar
	Status    QRStatus
	Message   string
	StudentID string
	Name      string
	Token     string
	User      UserInfo
	CreatedAt time.Time
	ExpiresAt time.Time
	mu         sync.Mutex // 保护并发读写
	done       bool       // 是否已完成，防止重复处理
	Processing sync.Mutex // 防止并发教务网请求
}

// SetResult 设置最终结果（线程安全）
func (s *QRSession) SetResult(status QRStatus, msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.done {
		return
	}
	s.Status = status
	s.Message = msg
	s.done = true
}

// IsDone 是否已经完成
func (s *QRSession) IsDone() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.done
}

// GetStatus 获取当前状态（线程安全）
func (s *QRSession) GetStatus() (QRStatus, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Status, s.Message
}

// SetStatus 更新状态（线程安全）
func (s *QRSession) SetStatus(status QRStatus, msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = status
	s.Message = msg
}
