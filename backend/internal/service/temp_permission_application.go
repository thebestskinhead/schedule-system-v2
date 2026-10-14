package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"schedule-system-v2/backend/internal/auth"
	"schedule-system-v2/backend/internal/dao"
	"schedule-system-v2/backend/internal/model"
)

// TempPermissionApplicationData 临权申请数据
// ExpiryDate 为指针：nil 表示永久有效。
type TempPermissionApplicationData struct {
	Permission string     `json:"permission"`  // 申请的权限标识
	ExpiryDate *time.Time `json:"expiry_date"` // 到期日期，空 = 永久
	Reason     string     `json:"reason"`      // 申请理由
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

	// 检查权限是否有效 - 使用冒号格式与系统权限保持一致
	validPermissions := []string{
		"schedule:publish",
		"schedule:manage:all",
		"user:manage:all",
		"schedule:manage:dept",
		"user:manage:dept",
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

	// 到期日期：空表示永久；非空需在当前时间之后，且不超过 90 天
	if appData.ExpiryDate != nil {
		if appData.ExpiryDate.Before(time.Now()) {
			return fmt.Errorf("到期日期不能早于当前时间")
		}
		maxExpiry := time.Now().AddDate(0, 3, 0)
		if appData.ExpiryDate.After(maxExpiry) {
			return fmt.Errorf("授权期限不能超过90天")
		}
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

	// 资源范围按申请人身份与权限语义推导：部门级权限限定在申请人所在部门，全局权限为 all
	resourceType, resourceID := auth.GrantResourceScope(model.Permission(data.Permission), app.Department)

	// 创建临时权限记录
	tempPerm := &model.UserPermissionTemp{
		UserID:       app.ApplicantID,
		Permission:   model.Permission(data.Permission),
		ResourceType: resourceType,
		ResourceID:   resourceID,
		GrantedBy:    approval.ApproverID,
		CreatedAt:    time.Now(),
		ExpiresAt:    data.ExpiryDate,
		IsActive:     true,
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

	// 解析申请的权限
	var data struct {
		Permission string `json:"permission"`
	}
	if err := json.Unmarshal(app.Data, &data); err != nil || data.Permission == "" {
		return nil, fmt.Errorf("无法解析申请的权限信息")
	}

	var approverIDs []int

	if isGlobalPermission(data.Permission) {
		// 全局权限（如 user:manage:all, schedule:manage:all）→ 发给办公室管理员或系统管理员
		officeAdmins, err := e.userDao.GetByDepartmentAndRole(model.DepartmentOffice, model.RoleDeptAdmin, model.EmploymentStatusActive)
		if err == nil {
			for _, admin := range officeAdmins {
				if admin.ID != app.ApplicantID {
					approverIDs = append(approverIDs, admin.ID)
				}
			}
		}
		if len(approverIDs) == 0 {
			admins, err := e.userDao.GetByRole(model.RoleAdmin, model.EmploymentStatusActive)
			if err != nil {
				return nil, err
			}
			for _, admin := range admins {
				if admin.ID != app.ApplicantID {
					approverIDs = append(approverIDs, admin.ID)
				}
			}
		}
	} else {
		// 部门级权限 → 发给部门管理员，无则回退到办公室管理员或系统管理员
		deptAdmins, err := e.userDao.GetByDepartmentAndRole(applicant.Department, model.RoleDeptAdmin, model.EmploymentStatusActive)
		if err != nil {
			return nil, err
		}
		for _, admin := range deptAdmins {
			if admin.ID != app.ApplicantID {
				approverIDs = append(approverIDs, admin.ID)
			}
		}
		if len(approverIDs) == 0 {
			officeAdmins, err := e.userDao.GetByDepartmentAndRole(model.DepartmentOffice, model.RoleDeptAdmin, model.EmploymentStatusActive)
			if err == nil {
				for _, admin := range officeAdmins {
					if admin.ID != app.ApplicantID {
						approverIDs = append(approverIDs, admin.ID)
					}
				}
			}
		}
		if len(approverIDs) == 0 {
			admins, err := e.userDao.GetByRole(model.RoleAdmin, model.EmploymentStatusActive)
			if err != nil {
				return nil, err
			}
			for _, admin := range admins {
				if admin.ID != app.ApplicantID {
					approverIDs = append(approverIDs, admin.ID)
				}
			}
		}
	}

	return approverIDs, nil
}

// isGlobalPermission 判断是否为全局权限
func isGlobalPermission(perm string) bool {
	return perm == "schedule:manage:all" || perm == "user:manage:all"
}

// CanApply 检查用户是否可以提交申请
func (e *TempPermissionExecutor) CanApply(ctx context.Context, userID int, data json.RawMessage) (bool, string) {
	user, err := e.userDao.GetByID(userID)
	if err != nil {
		return false, "获取用户信息失败"
	}

	// 系统管理员不需要申请临时权限
	if user.Role == model.RoleAdmin {
		return false, "系统管理员无需申请临时权限"
	}

	// 检查是否已有相同权限的待审批申请
	apps, _, err := e.applicationDao.GetByApplicant(ctx, userID, int(model.ApplicationStatusPending), 1, 100)
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
		if string(perm.Permission) == appData.Permission && (perm.ExpiresAt == nil || perm.ExpiresAt.After(time.Now())) {
			return false, "您已拥有该权限，无需重复申请"
		}
	}

	return true, ""
}

// GetAvailablePermissions 获取用户可申请的所有权限
// 基于用户的权限集合（角色权限 ∪ 临时权限）过滤已拥有的权限，不再依赖角色字符串判断。
func GetAvailablePermissions(checker *auth.Checker) []map[string]string {
	// 统一使用冒号格式与系统权限保持一致
	allPermissions := []map[string]string{
		{"key": "schedule:publish", "name": "设置每周分工", "description": "设置各部门在本周值班的日期安排"},
		{"key": "schedule:manage:all", "name": "排班管理（全部）", "description": "管理所有部门的排班"},
		{"key": "user:manage:all", "name": "用户管理（全部）", "description": "管理所有用户"},
		{"key": "schedule:manage:dept", "name": "排班管理（部门）", "description": "管理部门内排班"},
		{"key": "user:manage:dept", "name": "用户管理（部门）", "description": "管理部门内用户"},
	}

	// 根据权限集合过滤已拥有的权限
	permMap := map[string]auth.Permission{
		"schedule:publish":     auth.PermSchedulePublish,
		"schedule:manage:all":  auth.PermScheduleManageAll,
		"user:manage:all":      auth.PermUserManageAll,
		"schedule:manage:dept": auth.PermScheduleManageDept,
		"user:manage:dept":     auth.PermUserManageDept,
	}

	var available []map[string]string
	for _, perm := range allPermissions {
		if checker.HasPermission(permMap[perm["key"]]) {
			continue
		}
		available = append(available, perm)
	}

	return available
}
