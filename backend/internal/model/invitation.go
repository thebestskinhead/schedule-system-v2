package model

import "time"

// Invitation 预注册待激活账号（对应 user_invitations 表，字段严格照 migrations/012）
type Invitation struct {
	ID             int       `db:"id" json:"id"`
	StudentID      string    `db:"student_id" json:"student_id"`
	Name           string    `db:"name" json:"name"`
	Email          string    `db:"email" json:"email"`
	Department     string    `db:"department" json:"department"`
	DeptRole       string    `db:"dept_role" json:"dept_role"`
	Token          string    `db:"token" json:"-"` // 仅存 SHA-256 哈希，不对外暴露
	TokenExpiresAt time.Time `db:"token_expires_at" json:"token_expires_at"`
	Status         string    `db:"status" json:"status"` // pending/activated/expired/revoked
	CreatedBy      int       `db:"created_by" json:"created_by"`
	ActivatedAt    *time.Time `db:"activated_at" json:"activated_at"`
	ActivatedUserID *int     `db:"activated_user_id" json:"activated_user_id"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

// 邀请状态常量（user_invitations.status）
const (
	InvitationStatusPending   = "pending"
	InvitationStatusActivated = "activated"
	InvitationStatusExpired  = "expired"
	InvitationStatusRevoked  = "revoked"
)

// Invitation 链接有效期：7 天
const InvitationTokenTTL = 7 * 24 * time.Hour

// MemberCreateRequest 录入正式成员（单个）请求
type MemberCreateRequest struct {
	Name       string `json:"name" binding:"required"`
	StudentID  string `json:"student_id" binding:"required"`
	Department string `json:"department" binding:"required"`
	DeptRole   string `json:"dept_role" binding:"required,oneof=dept_admin dept_member"`
	Email      string `json:"email" binding:"required,email"`
}

// BatchMemberCreateRequest 批量录入正式成员请求
type BatchMemberCreateRequest struct {
	Items []MemberCreateRequest `json:"items" binding:"required,min=1"`
}

// BatchMemberItemResult 批量录入成功项
type BatchMemberItemResult struct {
	StudentID string `json:"student_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
}

// BatchMemberItemFailed 批量录入失败项
type BatchMemberItemFailed struct {
	StudentID string `json:"student_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Reason    string `json:"reason"`
}

// BatchMemberResult 批量录入汇总
type BatchMemberResult struct {
	Success []BatchMemberItemResult `json:"success"`
	Failed  []BatchMemberItemFailed `json:"failed"`
}

// ActivateInvitationRequest 受邀人自助激活请求
type ActivateInvitationRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}
