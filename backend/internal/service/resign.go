package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"schedule-system-v2/backend/internal/db"
	"schedule-system-v2/backend/internal/model"
)

// ResignPreview 批量离职预览（只读，见 H.4.1 预览-确认两阶段）
type ResignPreview struct {
	AffectedUserCount   int          `json:"affected_user_count"`
	PendingDutyCount    int64        `json:"pending_duty_count"`
	ActiveTempPermCount int64        `json:"active_temp_perm_count"`
	AffectedUsers       []model.User `json:"affected_users"`
}

// ResignResult 离职执行结果
type ResignResult struct {
	AffectedUserCount int      `json:"affected_user_count"`
	AffectedUserIDs   []int    `json:"affected_user_ids"`
	OperatorID        int      `json:"operator_id"`
	ExecutedAt        string   `json:"executed_at"`
}

// ResignUser 单个离职：所有「删除/离职」入口的唯一权威实现（H.4.1）。
// 离职 ≠ 删除：仅标记 employment_status='resigned' 并联动取消待处理排班/失效临时权限（见 sp_mark_user_resigned）。
func (s *UserService) ResignUser(operatorID, targetUserID int) error {
	if operatorID == targetUserID {
		return errors.New("不能离职自己")
	}
	target, err := s.userDAO.GetByID(targetUserID)
	if err != nil {
		return errors.New("用户不存在")
	}
	if target.EmploymentStatus == model.EmploymentStatusResigned {
		return errors.New("该用户已离职")
	}
	if target.Role == model.RoleAdmin {
		return errors.New("不能离职系统管理员")
	}

	// I-1 快照：执行前将受影响对象主键 + 原状态落盘应用日志
	logResignSnapshot(operatorID, "single", []model.User{*target})

	if err := s.userDAO.Delete(targetUserID, operatorID); err != nil {
		return err
	}
	return nil
}

// BatchResignPreview 批量离职预览：纯只读，返回受影响清单（用户数、将取消的待处理排班数、将失效的临时权限数）
func (s *UserService) BatchResignPreview(userIDs, excludeIDs []int) (*ResignPreview, error) {
	effective, err := s.resolveEffectiveIDs(userIDs, excludeIDs)
	if err != nil {
		return nil, err
	}
	users, err := s.userDAO.GetUsersByIDs(effective, "")
	if err != nil {
		return nil, err
	}
	pending, err := s.resignDAO.CountPendingDuties(effective)
	if err != nil {
		return nil, err
	}
	activePerms, err := s.resignDAO.CountActiveTempPerms(effective)
	if err != nil {
		return nil, err
	}
	return &ResignPreview{
		AffectedUserCount:   len(effective),
		PendingDutyCount:    pending,
		ActiveTempPermCount: activePerms,
		AffectedUsers:       users,
	}, nil
}

// BatchResign 批量离职执行：先解析有效（在职、去重、排除）用户，落盘快照后调 sp_mark_users_resigned
func (s *UserService) BatchResign(operatorID int, userIDs, excludeIDs []int) (*ResignResult, error) {
	effective, err := s.resolveEffectiveIDs(userIDs, excludeIDs)
	if err != nil {
		return nil, err
	}
	if len(effective) == 0 {
		return nil, errors.New("没有可离职的在职用户")
	}
	// 禁止将操作者自身纳入批量离职
	for _, id := range effective {
		if id == operatorID {
			return nil, errors.New("不可将自身纳入批量离职")
		}
	}

	users, err := s.userDAO.GetUsersByIDs(effective, "")
	if err != nil {
		return nil, err
	}
	// I-1 快照：执行前将受影响对象主键 + 原状态落盘应用日志
	logResignSnapshot(operatorID, "batch", users)

	idsJSON, err := json.Marshal(effective)
	if err != nil {
		return nil, err
	}

	spErr := db.WithRetry(context.Background(), func() error {
		_, ex := db.GetDB().Exec("CALL sp_mark_users_resigned(?, ?)", string(idsJSON), operatorID)
		return ex
	})
	if spErr != nil {
		if biz, ok := db.BusinessErrorFromSQLError(spErr); ok {
			return nil, biz
		}
		return nil, spErr
	}

	return &ResignResult{
		AffectedUserCount: len(effective),
		AffectedUserIDs:   effective,
		OperatorID:        operatorID,
		ExecutedAt:        time.Now().Format(time.RFC3339),
	}, nil
}

// resolveEffectiveIDs 将 user_ids 解析为真正会受影响（在职、去重、排除 exclude_ids）的用户主键列表
func (s *UserService) resolveEffectiveIDs(userIDs, excludeIDs []int) ([]int, error) {
	exSet := map[int]bool{}
	for _, id := range excludeIDs {
		exSet[id] = true
	}
	seen := map[int]bool{}
	var cand []int
	for _, id := range userIDs {
		if exSet[id] || seen[id] {
			continue
		}
		seen[id] = true
		cand = append(cand, id)
	}
	// 仅保留在职用户（离职/待入职不参与，SP 本身也只处理 active）
	activeUsers, err := s.userDAO.GetUsersByIDs(cand, model.EmploymentStatusActive)
	if err != nil {
		return nil, err
	}
	ids := make([]int, 0, len(activeUsers))
	for _, u := range activeUsers {
		ids = append(ids, u.ID)
	}
	return ids, nil
}

// logResignSnapshot I-1 要求：数据移动/删除前输出可恢复快照（受影响主键 + 原状态 + 操作者 + 时间），落应用日志。
func logResignSnapshot(operatorID int, action string, users []model.User) {
	ids := make([]int, 0, len(users))
	statuses := make([]string, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
		statuses = append(statuses, u.EmploymentStatus)
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"event":            "resign_snapshot",
		"operator_id":      operatorID,
		"action":           action,
		"affected_user_ids": ids,
		"affected_status":  statuses,
		"time":             time.Now().Format(time.RFC3339),
	})
	log.Printf("[RESIGN-SNAPSHOT] %s", string(payload))
}
