package service

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"schedule-system-v2/backend/internal/config"
	"schedule-system-v2/backend/internal/dao"
	"schedule-system-v2/backend/internal/model"
	"time"

	"gopkg.in/gomail.v2"
)

type SMTPService struct {
	smtpDAO *dao.SMTPDAO
	userDAO *dao.UserDAO
}

func NewSMTPService() *SMTPService {
	return &SMTPService{
		smtpDAO: dao.NewSMTPDAO(),
		userDAO: dao.NewUserDAO(),
	}
}

// GetConfig 获取当前SMTP配置
func (s *SMTPService) GetConfig() (*model.SMTPConfig, error) {
	return s.smtpDAO.GetConfig()
}

// GetAllConfigs 获取所有配置
func (s *SMTPService) GetAllConfigs() ([]model.SMTPConfig, error) {
	return s.smtpDAO.GetAllConfigs()
}

// SaveConfig 保存配置
func (s *SMTPService) SaveConfig(smtpConfig *model.SMTPConfig) error {
	if smtpConfig.ID == 0 {
		// 新建配置
		if smtpConfig.IsActive {
			// 如果启用新配置，先停用其他配置
			s.smtpDAO.DeactivateAll()
		}
		return s.smtpDAO.CreateConfig(smtpConfig)
	}
	// 更新配置
	if smtpConfig.IsActive {
		s.smtpDAO.DeactivateAll()
	}
	return s.smtpDAO.UpdateConfig(smtpConfig)
}

// DeleteConfig 删除配置
func (s *SMTPService) DeleteConfig(id int) error {
	return s.smtpDAO.DeleteConfig(id)
}

// TestEmail 发送测试邮件
func (s *SMTPService) TestEmail(to string) error {
	smtpConfig, err := s.smtpDAO.GetConfig()
	if err != nil {
		return fmt.Errorf("未配置SMTP或配置无效: %v", err)
	}

	subject := "SMTP测试邮件 - 排班系统"
	body := `
<h2>排班系统 SMTP 测试</h2>
<p>这是一封测试邮件，用于验证 SMTP 配置是否正确。</p>
<p>如果您收到此邮件，说明 SMTP 配置已成功！</p>
<br>
<p>发送时间: ` + time.Now().Format("2006-01-02 15:04:05") + `</p>
`

	return s.sendEmail(smtpConfig, to, subject, body)
}

// SendPasswordResetEmail 发送密码重置邮件
func (s *SMTPService) SendPasswordResetEmail(email, studentID string) error {
	// 查找用户
	var user *model.User
	var err error
	
	if email != "" {
		user, err = s.userDAO.GetByEmail(email)
	} else if studentID != "" {
		user, err = s.userDAO.GetByStudentID(studentID)
	}
	
	if err != nil || user == nil {
		return fmt.Errorf("用户未注册")
	}
	
	// 使用用户邮箱发送邮件
	if email == "" {
		email = user.Email
	}

	// 获取SMTP配置
	smtpConfig, err := s.smtpDAO.GetConfig()
	if err != nil {
		return fmt.Errorf("邮件服务未配置")
	}

	// 生成令牌
	token := generateToken()
	expiresAt := time.Now().Add(30 * time.Minute)

	// 保存令牌
	resetToken := &model.PasswordResetToken{
		UserID:    user.ID,
		Email:     email,
		Token:     token,
		ExpiresAt: expiresAt,
	}
	if err := s.smtpDAO.CreateResetToken(resetToken); err != nil {
		return fmt.Errorf("创建重置令牌失败: %v", err)
	}

	// 获取网站域名
	cfg := config.GetConfig()
	siteDomain := cfg.Site.Domain
	if siteDomain == "" {
		siteDomain = "http://localhost:8080"
	}

	// 发送邮件
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", siteDomain, token)
	subject := "密码重置 - 排班系统"
	body := fmt.Sprintf(`
<h2>密码重置请求</h2>
<p>您好 %s，</p>
<p>您申请了密码重置，请点击以下链接重置密码（30分钟内有效）：</p>
<p><a href="%s" style="padding: 10px 20px; background: #409eff; color: white; text-decoration: none; border-radius: 4px;">重置密码</a></p>
<p>或复制以下链接到浏览器：</p>
<p>%s</p>
<p>如果您没有申请重置密码，请忽略此邮件。</p>
<br>
<p>排班系统</p>
`, user.Name, resetURL, resetURL)

	return s.sendEmail(smtpConfig, email, subject, body)
}

// VerifyResetToken 验证重置令牌
func (s *SMTPService) VerifyResetToken(token string) (*model.PasswordResetToken, error) {
	return s.smtpDAO.GetResetToken(token)
}

// ResetPassword 重置密码
func (s *SMTPService) ResetPassword(token string, newPassword string) error {
	// 验证令牌
	t, err := s.smtpDAO.GetResetToken(token)
	if err != nil {
		return fmt.Errorf("令牌无效或已过期")
	}

	// 更新密码
	if err := s.userDAO.UpdatePassword(t.UserID, newPassword); err != nil {
		return fmt.Errorf("更新密码失败: %v", err)
	}

	// 标记令牌已使用
	return s.smtpDAO.MarkTokenUsed(t.ID)
}

// sendEmail 发送邮件（内部方法）
func (s *SMTPService) sendEmail(smtpConfig *model.SMTPConfig, to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(smtpConfig.FromEmail, smtpConfig.From))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpConfig.Host, smtpConfig.Port, smtpConfig.Username, smtpConfig.Password)
	
	if smtpConfig.UseSSL {
		d.SSL = true
	} else if smtpConfig.UseTLS {
		d.TLSConfig = &tls.Config{ServerName: smtpConfig.Host}
	}

	return d.DialAndSend(m)
}

// generateToken 生成随机令牌
func generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CheckConfigExists 检查是否存在配置
func (s *SMTPService) CheckConfigExists() bool {
	_, err := s.smtpDAO.GetConfig()
	return err == nil
}
