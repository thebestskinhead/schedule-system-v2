package dao

import (
	"time"

	"schedule-system-v2/backend/internal/db"
	"schedule-system-v2/backend/internal/model"
)

type InvitationDAO struct{}

func NewInvitationDAO() *InvitationDAO {
	return &InvitationDAO{}
}

// Create 插入一条邀请记录（token 调用方需先 SHA-256 哈希）
func (d *InvitationDAO) Create(inv *model.Invitation) error {
	query := `INSERT INTO user_invitations
		(student_id, name, email, department, dept_role, token, token_expires_at, status, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := db.GetDB().Exec(query,
		inv.StudentID, inv.Name, inv.Email, inv.Department, inv.DeptRole,
		inv.Token, inv.TokenExpiresAt, model.InvitationStatusPending, inv.CreatedBy)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	inv.ID = int(id)
	return nil
}

// GetByID 按主键查询（含全部状态）
func (d *InvitationDAO) GetByID(id int) (*model.Invitation, error) {
	var inv model.Invitation
	query := `SELECT * FROM user_invitations WHERE id = ?`
	err := db.GetDB().Get(&inv, query, id)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// GetByToken 按令牌哈希查询（含全部状态）
func (d *InvitationDAO) GetByToken(tokenHash string) (*model.Invitation, error) {
	var inv model.Invitation
	query := `SELECT * FROM user_invitations WHERE token = ?`
	err := db.GetDB().Get(&inv, query, tokenHash)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// GetByUserID 按激活后关联的用户 ID 查询
func (d *InvitationDAO) GetByUserID(userID int) (*model.Invitation, error) {
	var inv model.Invitation
	query := `SELECT * FROM user_invitations WHERE activated_user_id = ? ORDER BY id DESC LIMIT 1`
	err := db.GetDB().Get(&inv, query, userID)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// GetByStudentID 按学号查询（含全部状态）
func (d *InvitationDAO) GetByStudentID(studentID string) (*model.Invitation, error) {
	var inv model.Invitation
	query := `SELECT * FROM user_invitations WHERE student_id = ? ORDER BY id DESC LIMIT 1`
	err := db.GetDB().Get(&inv, query, studentID)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// GetByEmail 按邮箱查询（含全部状态）
func (d *InvitationDAO) GetByEmail(email string) (*model.Invitation, error) {
	var inv model.Invitation
	query := `SELECT * FROM user_invitations WHERE email = ? ORDER BY id DESC LIMIT 1`
	err := db.GetDB().Get(&inv, query, email)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// ExistsConflict 学号或邮箱在 pending/expired/revoked 状态的邀请中是否已占用
func (d *InvitationDAO) ExistsConflict(studentID, email string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM user_invitations
		WHERE (student_id = ? OR email = ?) AND status IN (?, ?, ?)`
	err := db.GetDB().Get(&count, query,
		studentID, email,
		model.InvitationStatusPending, model.InvitationStatusExpired, model.InvitationStatusRevoked)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// MarkActivated 标记邀请已使用并回填激活用户
func (d *InvitationDAO) MarkActivated(id, userID int) error {
	query := `UPDATE user_invitations
		SET status = ?, activated_at = NOW(), activated_user_id = ?
		WHERE id = ?`
	_, err := db.GetDB().Exec(query, model.InvitationStatusActivated, userID, id)
	return err
}

// MarkRevoked 作废邀请（仅 pending/expired 可作废）
func (d *InvitationDAO) MarkRevoked(id int) error {
	query := `UPDATE user_invitations
		SET status = ?
		WHERE id = ? AND status IN (?, ?)`
	_, err := db.GetDB().Exec(query, model.InvitationStatusRevoked, id,
		model.InvitationStatusPending, model.InvitationStatusExpired)
	return err
}

// Resend 重发邀请：重置令牌与过期时间、状态回 pending、清空激活信息
func (d *InvitationDAO) Resend(id int, tokenHash string, expiresAt time.Time) error {
	query := `UPDATE user_invitations
		SET token = ?, token_expires_at = ?, status = ?, activated_at = NULL, activated_user_id = NULL
		WHERE id = ? AND status IN (?, ?, ?)`
	_, err := db.GetDB().Exec(query, tokenHash, expiresAt, model.InvitationStatusPending, id,
		model.InvitationStatusPending, model.InvitationStatusExpired, model.InvitationStatusRevoked)
	return err
}
