package service

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"schedule-system-v2/backend/internal/config"
	"schedule-system-v2/backend/internal/dao"
	"schedule-system-v2/backend/internal/db"
	"schedule-system-v2/backend/internal/model"
	"schedule-system-v2/backend/internal/utils"
)

type InvitationService struct {
	invitationDAO *dao.InvitationDAO
	userDAO       *dao.UserDAO
	smtpService   *SMTPService
}

func NewInvitationService() *InvitationService {
	return &InvitationService{
		invitationDAO: dao.NewInvitationDAO(),
		userDAO:       dao.NewUserDAO(),
		smtpService:   NewSMTPService(),
	}
}

// generateInviteToken 生成原始令牌（未哈希，用于邮件链接）
func generateInviteToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// hashToken 对原始令牌做 SHA-256，返回 64 位十六进制串（与 user_invitations.token 存储一致）
func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// CreateMember 录入单个正式成员：创建 pending 用户 + 邀请令牌 + 发送激活邮件。
// 录入本身不受邮件发送失败影响（邮件失败仅记日志，录入视为成功）。
func (s *InvitationService) CreateMember(req *model.MemberCreateRequest, operatorID int) (*model.User, error) {
	if err := s.validateAndDedup(req); err != nil {
		return nil, err
	}

	hashedPwd, err := utils.HashPassword("123456")
	if err != nil {
		return nil, fmt.Errorf("生成占位密码失败: %v", err)
	}

	user := &model.User{
		StudentID:  req.StudentID,
		Name:       req.Name,
		Email:      req.Email,
		Password:   hashedPwd,
		Role:       model.RoleUser,
		Department: req.Department,
		DeptRole:   req.DeptRole,
	}
	if err := s.userDAO.CreatePending(user); err != nil {
		// 并发下「先查后插」可能被同批/跨请求穿透，靠唯一键兜底并转成友好提示
		if db.IsDuplicateEntry(err) {
			return nil, errors.New("学号或邮箱已被占用")
		}
		return nil, fmt.Errorf("创建用户失败: %v", err)
	}

	if err := s.createInvitationAndSend(user, operatorID); err != nil {
		// 录入已成功，邮件失败仅告警
		return user, err
	}
	return user, nil
}

// BatchCreateMembers 批量录入：并发逐条校验并汇总成功/失败。邮件异步入队，不影响录入结果。
func (s *InvitationService) BatchCreateMembers(items []model.MemberCreateRequest, operatorID int) *model.BatchMemberResult {
	result := &model.BatchMemberResult{
		Success: []model.BatchMemberItemResult{},
		Failed:  []model.BatchMemberItemFailed{},
	}
	if len(items) == 0 {
		return result
	}

	// 批内预去重：同一批次内学号/邮箱重复的条目，保留首次出现，其余直接判失败。
	// 避免并发下多条同时通过「先查后插」校验而互相穿透。
	seen := make([]bool, len(items))
	seenStudentID := make(map[string]int)
	seenEmail := make(map[string]int)
	for i, item := range items {
		if prev, ok := seenStudentID[item.StudentID]; ok {
			seen[i] = true
			result.Failed = append(result.Failed, model.BatchMemberItemFailed{
				StudentID: item.StudentID,
				Name:      item.Name,
				Email:     item.Email,
				Reason:    fmt.Sprintf("与本批次第 %d 条学号重复", prev+1),
			})
			continue
		}
		if prev, ok := seenEmail[item.Email]; ok {
			seen[i] = true
			result.Failed = append(result.Failed, model.BatchMemberItemFailed{
				StudentID: item.StudentID,
				Name:      item.Name,
				Email:     item.Email,
				Reason:    fmt.Sprintf("与本批次第 %d 条邮箱重复", prev+1),
			})
			continue
		}
		seenStudentID[item.StudentID] = i
		seenEmail[item.Email] = i
	}

	// 收集待并发处理的条目下标
	pending := make([]int, 0, len(items))
	for i := range items {
		if !seen[i] {
			pending = append(pending, i)
		}
	}

	// 并发度：最多 8，避免过度冲击数据库
	concurrency := 8
	if len(pending) < concurrency {
		concurrency = len(pending)
	}

	if concurrency > 0 {
		type outcome struct {
			idx  int
			user *model.User
			item model.MemberCreateRequest
			err  error
		}
		jobs := make(chan int)
		results := make(chan outcome, len(pending))

		var wg sync.WaitGroup
		for w := 0; w < concurrency; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for idx := range jobs {
					item := items[idx]
					user, err := s.CreateMember(&item, operatorID)
					results <- outcome{idx: idx, user: user, item: item, err: err}
				}
			}()
		}

		go func() {
			for _, idx := range pending {
				jobs <- idx
			}
			close(jobs)
		}()

		wg.Wait()
		close(results)

		// 按输入顺序汇总，保证结果稳定可读
		collected := make([]outcome, len(items))
		for o := range results {
			collected[o.idx] = o
		}
		for _, idx := range pending {
			o := collected[idx]
			if o.err != nil {
				result.Failed = append(result.Failed, model.BatchMemberItemFailed{
					StudentID: o.item.StudentID,
					Name:      o.item.Name,
					Email:     o.item.Email,
					Reason:    o.err.Error(),
				})
				continue
			}
			result.Success = append(result.Success, model.BatchMemberItemResult{
				StudentID: o.user.StudentID,
				Name:      o.user.Name,
				Email:     o.user.Email,
			})
		}
	}

	return result
}

// validateAndDedup 字段校验 + 查重（users 全状态 & user_invitations 的 pending/expired/revoked）
func (s *InvitationService) validateAndDedup(req *model.MemberCreateRequest) error {
	if req.DeptRole != model.RoleDeptAdmin && req.DeptRole != model.RoleDeptMember {
		return errors.New("部门角色不合法")
	}

	// users 全状态查重
	if u, err := s.userDAO.GetByStudentID(req.StudentID); err == nil && u != nil {
		return fmt.Errorf("学号 %s 已被占用", req.StudentID)
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("查询学号占用失败: %v", err)
	}
	if u, err := s.userDAO.GetByEmail(req.Email); err == nil && u != nil {
		return fmt.Errorf("邮箱 %s 已被占用", req.Email)
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("查询邮箱占用失败: %v", err)
	}

	// user_invitations 待处理/过期/作废查重
	conflict, err := s.invitationDAO.ExistsConflict(req.StudentID, req.Email)
	if err != nil {
		return fmt.Errorf("查询邀请占用失败: %v", err)
	}
	if conflict {
		return errors.New("该学号或邮箱已存在待处理、过期或已作废的邀请，请先作废或重发")
	}
	return nil
}

// createInvitationAndSend 生成令牌、写库、投递激活邮件（邮件异步入队，不影响录入）
func (s *InvitationService) createInvitationAndSend(user *model.User, operatorID int) error {
	rawToken, err := generateInviteToken()
	if err != nil {
		return fmt.Errorf("生成令牌失败: %v", err)
	}
	inv := &model.Invitation{
		StudentID:      user.StudentID,
		Name:           user.Name,
		Email:          user.Email,
		Department:     user.Department,
		DeptRole:       user.DeptRole,
		Token:          hashToken(rawToken),
		TokenExpiresAt: time.Now().Add(model.InvitationTokenTTL),
		CreatedBy:      operatorID,
	}
	if err := s.invitationDAO.Create(inv); err != nil {
		return fmt.Errorf("创建邀请记录失败: %v", err)
	}

	// 投递激活邮件（异步，立即返回）
	s.sendActivationEmail(user, rawToken)
	return nil
}

// sendActivationEmail 构造激活链接并投递邀请邮件（异步）
func (s *InvitationService) sendActivationEmail(user *model.User, rawToken string) {
	cfg := config.GetConfig()
	domain := cfg.Site.Domain
	if domain == "" {
		domain = "http://localhost:8080"
	}
	activateURL := fmt.Sprintf("%s/invitations/activate?token=%s", domain, rawToken)
	s.smtpService.SendInvitationEmail(user.Email, user.Name, activateURL)
}

// ResendInvite 重发邀请：重置令牌与过期时间并重新发送邮件（仅 pending/expired/revoked 可重发）
func (s *InvitationService) ResendInvite(id int, operatorID int) error {
	inv, err := s.invitationDAO.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("邀请记录不存在")
		}
		return fmt.Errorf("查询邀请失败: %v", err)
	}
	if inv.Status == model.InvitationStatusActivated {
		return errors.New("该邀请已激活，无需重发")
	}

	rawToken, err := generateInviteToken()
	if err != nil {
		return fmt.Errorf("生成令牌失败: %v", err)
	}
	if err := s.invitationDAO.Resend(id, hashToken(rawToken), time.Now().Add(model.InvitationTokenTTL)); err != nil {
		return fmt.Errorf("更新邀请失败: %v", err)
	}

	// 读取最新用户（pending）用于邮件称呼
	user := &model.User{StudentID: inv.StudentID, Name: inv.Name, Email: inv.Email}
	s.sendActivationEmail(user, rawToken)
	return nil
}

// RevokeInvite 作废邀请（仅 pending/expired 可作废）
func (s *InvitationService) RevokeInvite(id int) error {
	inv, err := s.invitationDAO.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("邀请记录不存在")
		}
		return fmt.Errorf("查询邀请失败: %v", err)
	}
	if inv.Status == model.InvitationStatusActivated {
		return errors.New("该邀请已激活，无法作废")
	}
	if err := s.invitationDAO.MarkRevoked(id); err != nil {
		return fmt.Errorf("作废邀请失败: %v", err)
	}
	return nil
}

// ActivateInvitation 受邀人自助激活：对 token 做 SHA-256 后调用 sp_activate_invitation。
// 返回清晰错误：令牌无效 / 已过期 / 已使用。
func (s *InvitationService) ActivateInvitation(token, password string) error {
	tokenHash := hashToken(token)

	// 预检：给出更清晰的分类错误
	inv, err := s.invitationDAO.GetByToken(tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("邀请令牌无效")
		}
		return fmt.Errorf("查询邀请失败: %v", err)
	}
	switch inv.Status {
	case model.InvitationStatusActivated:
		return errors.New("该邀请已使用，请勿重复激活")
	case model.InvitationStatusRevoked:
		return errors.New("该邀请已作废，请联系管理员")
	case model.InvitationStatusExpired:
		return errors.New("邀请令牌已过期，请联系管理员重新发送邀请")
	case model.InvitationStatusPending:
		if time.Now().After(inv.TokenExpiresAt) {
			return errors.New("邀请令牌已过期，请联系管理员重新发送邀请")
		}
	default:
		return errors.New("邀请状态异常")
	}

	// 密码先 bcrypt，再传入存储过程
	hashedPwd, err := utils.HashPassword(password)
	if err != nil {
		return fmt.Errorf("加密密码失败: %v", err)
	}

	// 调用存储过程完成激活（事务 + 行锁，唯一入口）
	if _, err := db.GetDB().Exec("CALL sp_activate_invitation(?, ?)", tokenHash, hashedPwd); err != nil {
		msg := err.Error()
		switch {
		case strings.Contains(msg, "已过期"):
			return errors.New("邀请令牌已过期，请联系管理员重新发送邀请")
		case strings.Contains(msg, "无效或已使用"):
			return errors.New("邀请令牌无效或已使用")
		case strings.Contains(msg, "状态异常"):
			return errors.New("用户记录不存在或状态异常，请联系管理员")
		default:
			return fmt.Errorf("激活失败: %v", err)
		}
	}
	return nil
}
