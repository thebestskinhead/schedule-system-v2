package model

import "time"

// SMTPConfig SMTP配置
type SMTPConfig struct {
	ID       int    `json:"id" db:"id"`
	Host     string `json:"host" db:"host" binding:"required"`     // SMTP服务器地址
	Port     int    `json:"port" db:"port" binding:"required,min=1,max=65535"` // 端口
	Username string `json:"username" db:"username" binding:"required"` // 用户名/邮箱
	Password string `json:"password" db:"password" binding:"required"` // 密码/授权码
	From     string `json:"from" db:"from_name" binding:"required"`         // 发件人显示名称
	FromEmail string `json:"from_email" db:"from_email" binding:"required"` // 发件人邮箱
	UseTLS   bool   `json:"use_tls" db:"use_tls"`                      // 是否使用TLS
	UseSSL   bool   `json:"use_ssl" db:"use_ssl"`                      // 是否使用SSL
	IsActive bool   `json:"is_active" db:"is_active"`                  // 是否启用
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// SMTPTestRequest SMTP测试请求
type SMTPTestRequest struct {
	To string `json:"to" binding:"required,email"` // 测试收件人邮箱
}

// SMTPTestResponse SMTP测试响应
type SMTPTestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// PasswordResetRequest 密码重置请求
type PasswordResetRequest struct {
	Email     string `json:"email" binding:"omitempty,email"`
	StudentID string `json:"student_id" binding:"omitempty"`
}

// PasswordResetConfirm 密码重置确认
type PasswordResetConfirm struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// PasswordResetToken 密码重置令牌
type PasswordResetToken struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Email     string    `json:"email" db:"email"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	IsUsed    bool      `json:"is_used" db:"is_used"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
