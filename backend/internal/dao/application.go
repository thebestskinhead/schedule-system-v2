package dao

import (
	"context"
	"fmt"
	"schedule-system-v2/backend/internal/db"
	"schedule-system-v2/backend/internal/model"
	"strings"
	"time"
)

type ApplicationDao struct{}

func NewApplicationDao() *ApplicationDao {
	return &ApplicationDao{}
}

// ========== 申请类型操作 ==========

// GetApplicationType 根据代码获取申请类型
func (d *ApplicationDao) GetApplicationType(ctx context.Context, code string) (*model.ApplicationType, error) {
	var t model.ApplicationType
	query := `SELECT * FROM application_types WHERE code = ? AND is_active = 1`
	err := db.GetDB().Get(&t, query, code)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// GetAllTypes 获取所有申请类型
func (d *ApplicationDao) GetAllTypes(ctx context.Context) ([]model.ApplicationType, error) {
	var types []model.ApplicationType
	query := `SELECT * FROM application_types WHERE is_active = 1 ORDER BY id`
	err := db.GetDB().Select(&types, query)
	return types, err
}

// ========== 申请单操作 ==========

// Create 创建申请
func (d *ApplicationDao) Create(ctx context.Context, app *model.Application) error {
	query := `INSERT INTO applications 
		(application_no, type_code, applicant_id, department, title, content, data, status, current_level, total_levels, approver_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	result, err := db.GetDB().Exec(query,
		app.ApplicationNo,
		app.TypeCode,
		app.ApplicantID,
		app.Department,
		app.Title,
		app.Content,
		app.Data,
		app.Status,
		app.CurrentLevel,
		app.TotalLevels,
		app.ApproverID,
		app.CreatedAt,
		app.UpdatedAt,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	app.ID = int(id)
	return nil
}

// GetByID 根据ID获取申请
func (d *ApplicationDao) GetByID(ctx context.Context, id int) (*model.Application, error) {
	var app model.Application
	query := `SELECT a.*, u.name as applicant_name
		FROM applications a
		LEFT JOIN users u ON a.applicant_id = u.id
		WHERE a.id = ?`
	err := db.GetDB().Get(&app, query, id)
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// GetByApplicant 获取申请人的申请列表
func (d *ApplicationDao) GetByApplicant(ctx context.Context, applicantID int, status int, page, pageSize int) ([]model.Application, int, error) {
	var conditions []string
	var args []interface{}
	
	conditions = append(conditions, "a.applicant_id = ?")
	args = append(args, applicantID)
	
	if status >= 0 {
		conditions = append(conditions, "a.status = ?")
		args = append(args, status)
	}
	
	whereClause := "WHERE " + strings.Join(conditions, " AND ")
	
	// 查询总数
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM applications a %s", whereClause)
	err := db.GetDB().Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	
	// 查询数据
	query := fmt.Sprintf(`SELECT a.*, u.name as applicant_name
		FROM applications a
		LEFT JOIN users u ON a.applicant_id = u.id
		%s
		ORDER BY a.created_at DESC
		LIMIT ? OFFSET ?`, whereClause)
	
	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)
	
	var apps []model.Application
	err = db.GetDB().Select(&apps, query, args...)
	if err != nil {
		return nil, 0, err
	}
	
	return apps, total, nil
}

// GetPendingByApprover 获取待审批列表
func (d *ApplicationDao) GetPendingByApprover(ctx context.Context, approverID int, isAdmin bool, department string, page, pageSize int) ([]model.Application, int, error) {
	var conditions []string
	var args []interface{}
	
	conditions = append(conditions, "a.status IN (0, 1)")
	
	if !isAdmin && department != "" {
		// 部门管理员只能看到本部门的
		conditions = append(conditions, "a.department = ?")
		args = append(args, department)
		// 排除全局权限申请（type_code=temp_permission 且 data 中包含 :manage:all）
		conditions = append(conditions, "(a.type_code != 'temp_permission' OR a.data NOT LIKE '%manage:all%')")
	}
	
	whereClause := "WHERE " + strings.Join(conditions, " AND ")
	
	// 查询总数
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM applications a %s", whereClause)
	err := db.GetDB().Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	
	// 查询数据
	query := fmt.Sprintf(`SELECT a.*, u.name as applicant_name
		FROM applications a
		LEFT JOIN users u ON a.applicant_id = u.id
		%s
		ORDER BY a.created_at DESC
		LIMIT ? OFFSET ?`, whereClause)
	
	offset := (page - 1) * pageSize
	args = append(args, pageSize, offset)
	
	var apps []model.Application
	err = db.GetDB().Select(&apps, query, args...)
	if err != nil {
		return nil, 0, err
	}
	
	return apps, total, nil
}

// Update 更新申请
func (d *ApplicationDao) Update(ctx context.Context, app *model.Application) error {
	query := `UPDATE applications SET 
		status = ?, current_level = ?, approver_id = ?, updated_at = ?
		WHERE id = ?`
	_, err := db.GetDB().Exec(query, app.Status, app.CurrentLevel, app.ApproverID, app.UpdatedAt, app.ID)
	return err
}

// IsApproverForApplication 检查用户是否可以审批某个申请
func (d *ApplicationDao) IsApproverForApplication(ctx context.Context, appID, userID int) (bool, error) {
	// 获取申请信息
	app, err := d.GetByID(ctx, appID)
	if err != nil {
		return false, err
	}
	
	// 获取审批流程配置
	query := `SELECT COUNT(*) FROM application_approvers 
		WHERE application_id = ? AND approver_id = ? AND level = ?`
	var count int
	err = db.GetDB().Get(&count, query, appID, userID, app.CurrentLevel)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// ========== 审批历史操作 ==========

// CreateApproval 创建审批记录
func (d *ApplicationDao) CreateApproval(ctx context.Context, approval *model.ApplicationApproval) error {
	query := `INSERT INTO application_approvals 
		(application_id, approver_id, action, comment, level, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`
	
	result, err := db.GetDB().Exec(query,
		approval.ApplicationID,
		approval.ApproverID,
		approval.Action,
		approval.Comment,
		approval.Level,
		approval.CreatedAt,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	approval.ID = int(id)
	return nil
}

// GetApprovalHistory 获取审批历史
func (d *ApplicationDao) GetApprovalHistory(applicationID int) ([]model.ApplicationApproval, error) {
	query := `SELECT aa.*, u.name as approver_name
		FROM application_approvals aa
		LEFT JOIN users u ON aa.approver_id = u.id
		WHERE aa.application_id = ?
		ORDER BY aa.created_at ASC`
	
	var history []model.ApplicationApproval
	err := db.GetDB().Select(&history, query, applicationID)
	return history, err
}

// ========== 审批人配置操作 ==========

// GetApproversByType 获取申请类型的审批人配置
func (d *ApplicationDao) GetApproversByType(ctx context.Context, typeCode string) ([]model.ApplicationApprover, error) {
	var approvers []model.ApplicationApprover
	query := `SELECT * FROM application_approvers 
		WHERE type_code = ? AND is_active = 1 
		ORDER BY level ASC`
	err := db.GetDB().Select(&approvers, query, typeCode)
	return approvers, err
}

// ========== 统计操作 ==========

// CountByApplicantAndStatus 统计用户的申请数量
func (d *ApplicationDao) CountByApplicantAndStatus(ctx context.Context, userID int, status int) (int, error) {
	query := `SELECT COUNT(*) FROM applications WHERE applicant_id = ? AND status = ?`
	var count int
	err := db.GetDB().Get(&count, query, userID, status)
	return count, err
}

// CountPendingByApprover 统计待审批数量
func (d *ApplicationDao) CountPendingByApprover(ctx context.Context, userID int, isAdmin bool, department string) (int, error) {
	var query string
	var args []interface{}
	
	if isAdmin {
		query = `SELECT COUNT(*) FROM applications WHERE status IN (0, 1)`
	} else {
		// 部门管理员：本部门 + 排除全局权限申请
		query = `SELECT COUNT(*) FROM applications WHERE status IN (0, 1) AND department = ? AND (type_code != 'temp_permission' OR data NOT LIKE '%manage:all%')`
		args = append(args, department)
	}
	
	var count int
	err := db.GetDB().Get(&count, query, args...)
	return count, err
}

// GenerateApplicationNo 生成申请单号
func (d *ApplicationDao) GenerateApplicationNo(ctx context.Context, appType string) (string, error) {
	now := time.Now()
	prefix := strings.ToUpper(appType)
	dateStr := now.Format("20060102")
	random := now.UnixNano() % 10000
	return fmt.Sprintf("%s%s%04d", prefix, dateStr, random), nil
}
