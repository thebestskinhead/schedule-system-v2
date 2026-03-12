package dao

import (
	"schedule-system-v2/backend/internal/db"
	"schedule-system-v2/backend/internal/model"
)

type SMTPDAO struct{}

func NewSMTPDAO() *SMTPDAO {
	return &SMTPDAO{}
}

// GetConfig 获取SMTP配置（取第一个启用的配置）
func (d *SMTPDAO) GetConfig() (*model.SMTPConfig, error) {
	var config model.SMTPConfig
	query := `SELECT * FROM smtp_config WHERE is_active = TRUE ORDER BY id DESC LIMIT 1`
	err := db.GetDB().Get(&config, query)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetConfigByID 根据ID获取配置
func (d *SMTPDAO) GetConfigByID(id int) (*model.SMTPConfig, error) {
	var config model.SMTPConfig
	query := `SELECT * FROM smtp_config WHERE id = ?`
	err := db.GetDB().Get(&config, query, id)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetAllConfigs 获取所有配置
func (d *SMTPDAO) GetAllConfigs() ([]model.SMTPConfig, error) {
	var configs []model.SMTPConfig
	query := `SELECT * FROM smtp_config ORDER BY id DESC`
	err := db.GetDB().Select(&configs, query)
	return configs, err
}

// CreateConfig 创建配置
func (d *SMTPDAO) CreateConfig(config *model.SMTPConfig) error {
	query := `INSERT INTO smtp_config 
		(host, port, username, password, from_name, from_email, use_tls, use_ssl, is_active) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := db.GetDB().Exec(query, 
		config.Host, config.Port, config.Username, config.Password,
		config.From, config.FromEmail, config.UseTLS, config.UseSSL, config.IsActive)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	config.ID = int(id)
	return nil
}

// UpdateConfig 更新配置
func (d *SMTPDAO) UpdateConfig(config *model.SMTPConfig) error {
	query := `UPDATE smtp_config SET 
		host = ?, port = ?, username = ?, password = ?, 
		from_name = ?, from_email = ?, use_tls = ?, use_ssl = ?, is_active = ?
		WHERE id = ?`
	_, err := db.GetDB().Exec(query,
		config.Host, config.Port, config.Username, config.Password,
		config.From, config.FromEmail, config.UseTLS, config.UseSSL, config.IsActive, config.ID)
	return err
}

// DeleteConfig 删除配置
func (d *SMTPDAO) DeleteConfig(id int) error {
	query := `DELETE FROM smtp_config WHERE id = ?`
	_, err := db.GetDB().Exec(query, id)
	return err
}

// DeactivateAll 停用所有配置
func (d *SMTPDAO) DeactivateAll() error {
	query := `UPDATE smtp_config SET is_active = FALSE`
	_, err := db.GetDB().Exec(query)
	return err
}

// CreateResetToken 创建密码重置令牌
func (d *SMTPDAO) CreateResetToken(token *model.PasswordResetToken) error {
	query := `INSERT INTO password_reset_tokens (user_id, email, token, expires_at) VALUES (?, ?, ?, ?)`
	result, err := db.GetDB().Exec(query, token.UserID, token.Email, token.Token, token.ExpiresAt)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	token.ID = int(id)
	return nil
}

// GetResetToken 获取有效令牌
func (d *SMTPDAO) GetResetToken(token string) (*model.PasswordResetToken, error) {
	var t model.PasswordResetToken
	query := `SELECT * FROM password_reset_tokens WHERE token = ? AND is_used = FALSE AND expires_at > NOW()`
	err := db.GetDB().Get(&t, query, token)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// MarkTokenUsed 标记令牌已使用
func (d *SMTPDAO) MarkTokenUsed(id int) error {
	query := `UPDATE password_reset_tokens SET is_used = TRUE WHERE id = ?`
	_, err := db.GetDB().Exec(query, id)
	return err
}

// CleanupExpiredTokens 清理过期令牌
func (d *SMTPDAO) CleanupExpiredTokens() error {
	query := `DELETE FROM password_reset_tokens WHERE expires_at <= NOW() OR is_used = TRUE`
	_, err := db.GetDB().Exec(query)
	return err
}
