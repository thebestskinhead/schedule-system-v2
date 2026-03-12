package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"schedule-system-v2/backend/internal/dao"
	"schedule-system-v2/backend/internal/model"
)

// TempPermissionApplicationData 临权申请数据
type TempPermissionApplicationData struct {
	Permission string    `json:"permission"`   // 申请的权限标识
	ExpiryDate time.Time `json:"expiry_date"`  // 到期日期
	Reason     string    `json:"reason"`       // 申请理由
}

// TempPermissionExecutor 临时权限申请执行器
type TempPermissionExecutor struct {
	userDao            *dao.UserDAO
	tempPermissionDao  *dao.TempPermissionDAO
	applicationDao     *dao.ApplicationDao
}

// NewTempPermissionExecutor 创建临权执行器
func NewTempPermissionExecutor(userDao *dao.UserDAO, tempPermissionDao *dao.TempPermissionDAO, applicationDao *dao.ApplicationDao) *TempPermissionExecutor {
	return &TempPermissionExecutor{
		userDao:           userDao,
		tempPermissionDao: tempPermissionDao,
		applicationDao:    applicationDao,
	}
}

// GetType 返回应用类型标识
func (e *TempPermissionExecutor) GetType() string {
	return "temp_permission"
}

// Validate 验证申请数据
func (e *TempPermissionExecutor) Validate(data json.RawMessage) error {
	var appData TempPermissionApplicationData
	if err := json.Unmarshal(data, &appData); err != nil {
		return fmt.Errorf("数据格式错误: %v", err)
	}

	if appData.Permission == "" {
		return fmt.Errorf("权限标识不能为空")
	}

	// 检查权限是否有效
	validPermissions := []string{
		"duty_manage",
		"user_manage",
		"schedule_manage",
		"crawler_manage",
		"system_manage",
		"temp_permission_grant",
	}
	found := false
	for _, perm := range validPermissions {
		if perm == appData.Permission {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("无效的权限标识: %s", appData.Permission)
	}

	if appData.ExpiryDate.Before(time.Now()) {
		return fmt.Errorf("到期日期不能早于当前时间")
	}

	// 最大期限限制（例如90天）
	maxExpiry := time.Now().AddDate(0, 3, 0)
	if appData.ExpiryDate.After(maxExpiry) {
		return fmt.Errorf("授权期限不能超过90天")
	}

	return nil
}

// OnCreated 申请创建后的回调
func (e *TempPermissionExecutor) OnCreated(ctx context.Context, app *model.Application) error {
	// 可以在这里发送通知给申请人
	return nil
}

// OnApproved 申请批准后的回调 - 授予临时权限
func (e *TempPermissionExecutor) OnApproved(ctx context.Context, app *model.Application, approval *model.ApplicationApproval) error {
	var data TempPermissionApplicationData
	if err := json.Unmarshal(app.Data, &data); err != nil {
		return fmt.Errorf("解析申请数据失败: %v", err)
	}

	// 创建临时权限记录
	tempPerm := &model.UserPermissionTemp{
		UserID:     app.ApplicantID,
		Permission: model.Permission(data.Permission),
		GrantedBy:  approval.ApproverID,
		CreatedAt:  time.Now(),
		ExpiresAt:  data.ExpiryDate,
		IsActive:   true,
	}

	if err := e.tempPermissionDao.Create(tempPerm); err != nil {
		return fmt.Errorf("创建临时权限失败: %v", err)
	}

	return nil
}

// OnRejected 申请拒绝后的回调
func (e *TempPermissionExecutor) OnRejected(ctx context.Context, app *model.Application, approval *model.ApplicationApproval) error {
	// 可以发送通知给申请人
	return nil
}

// GetRequiredApprovers 获取所需的审批人列表
func (e *TempPermissionExecutor) GetRequiredApprovers(ctx context.Context, app *model.Application) ([]int, error) {
	// 获取申请人信息
	applicant, err := e.userDao.GetByID(app.ApplicantID)
	if err != nil {
		return nil, err
	}

	var approverIDs []int

	// 查找申请人的部门管理员
	deptAdmins, err := e.userDao.GetByDepartmentAndRole(applicant.Department, "dept_admin")
	if err != nil {
		return nil, err
	}
	for _, admin := range deptAdmins {
		if admin.ID != app.ApplicantID { // 不能审批自己的申请
			approverIDs = append(approverIDs, admin.ID)
		}
	}

	// 如果没有部门管理员，查找系统管理员
	if len(approverIDs) == 0 {
		admins, err := e.userDao.GetByRole("admin")
		if err != nil {
			return nil, err
		}
		for _, admin := range admins {
			if admin.ID != app.ApplicantID {
				approverIDs = append(approverIDs, admin.ID)
			}
		}
	}

	return approverIDs, nil
}

// CanApply 检查用户是否可以提交申请
func (e *TempPermissionExecutor) CanApply(ctx context.Context, userID int, data json.RawMessage) (bool, string) {
	user, err := e.userDao.GetByID(userID)
	if err != nil {
		return false, "获取用户信息失败"
	}

	// 系统管理员不需要申请临时权限
	if user.Role == "admin" {
		return false, "系统管理员无需申请临时权限"
	}

	// 检查是否已有相同权限的待审批申请
	apps, _, err := e.applicationDao.GetByApplicant(ctx, userID, string(model.ApplicationStatusPending), 1, 100)
	if err != nil {
		return false, "查询申请记录失败"
	}

	var appData TempPermissionApplicationData
	if err := json.Unmarshal(data, &appData); err != nil {
		return false, "数据格式错误"
	}

	for _, app := range apps {
		if app.TypeCode == "temp_permission" && app.Status == model.ApplicationStatusPending {
			var existingData TempPermissionApplicationData
			if err := json.Unmarshal(app.Data, &existingData); err == nil {
				if existingData.Permission == appData.Permission {
					return false, "您已有一个相同权限的待审批申请"
				}
			}
		}
	}

	// 检查是否已有有效的临时权限
	existingPerms, err := e.tempPermissionDao.GetByUserID(ctx, userID)
	if err != nil {
		return false, "查询权限记录失败"
	}
	for _, perm := range existingPerms {
		if string(perm.Permission) == appData.Permission && perm.ExpiresAt.After(time.Now()) {
			return false, "您已拥有该权限，无需重复申请"
		}
	}

	return true, ""
}

// GetAvailablePermissions 获取用户可申请的所有权限
func GetAvailablePermissions(user *model.User) []map[string]string {
	allPermissions := []map[string]string{
		{"key": "duty_manage", "name": "值班管理", "description": "管理值班安排和记录"},
		{"key": "user_manage", "name": "用户管理", "description": "管理用户信息"},
		{"key": "schedule_manage", "name": "排班管理", "description": "创建和修改排班"},
		{"key": "crawler_manage", "name": "爬虫管理", "description": "管理课程导入爬虫"},
		{"key": "system_manage", "name": "系统管理", "description": "系统配置管理"},
		{"key": "temp_permission_grant", "name": "授权管理", "description": "授予他人临时权限"},
	}

	// 根据用户角色过滤已拥有的权限
	var available []map[string]string
	for _, perm := range allPermissions {
		hasPerm := false
		if user.Role == "admin" {
			hasPerm = true
		} else if user.DeptRole == "dept_admin" {
			// 部门管理员默认拥有大部分权限
			switch perm["key"] {
			case "duty_manage", "schedule_manage", "crawler_manage", "temp_permission_grant":
				hasPerm = true
			}
		}

		if !hasPerm {
			available = append(available, perm)
		}
	}

	return available
}
