package service

import (
	"errors"
	"time"
	"schedule-system-v2/backend/internal/auth"
	"schedule-system-v2/backend/internal/dao"
	"schedule-system-v2/backend/internal/model"
)

type TempPermissionService struct {
	dao     *dao.TempPermissionDAO
	userDAO *dao.UserDAO
}

func NewTempPermissionService() *TempPermissionService {
	return &TempPermissionService{
		dao:     dao.NewTempPermissionDAO(),
		userDAO: dao.NewUserDAO(),
	}
}

// GrantPermission 授予临时权限（带权限检查）
func (s *TempPermissionService) GrantPermission(adminID int, checker *auth.Checker, req *model.SingleGrantRequest) error {
	// 验证用户是否存在
	user, err := s.userDAO.GetByID(req.UserID)
	if err != nil {
		return errors.New("用户不存在")
	}
	if user == nil {
		return errors.New("用户不存在")
	}

	// 验证过期时间
	if req.ExpiresAt.Before(time.Now()) {
		return errors.New("过期时间必须在将来")
	}

	// 验证权限代码
	validPerm := false
	for _, info := range model.GetPermissionList() {
		if info.Code == string(req.Permission) {
			validPerm = true
			break
		}
	}
	if !validPerm {
		return errors.New("无效的权限代码")
	}

	// 检查授权人是否有权限授予此权限
	if !s.canGrantPermission(checker, req.Permission, req.ResourceType) {
		return errors.New("您没有权限授予此权限")
	}

	// 创建权限记录
	perm := &model.UserPermissionTemp{
		UserID:       req.UserID,
		Permission:   req.Permission,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		GrantedBy:    adminID,
		Reason:       req.Reason,
		ExpiresAt:    req.ExpiresAt,
		IsActive:     true,
	}

	return s.dao.Create(perm)
}

// canGrantPermission 检查授权人是否可以授予指定权限
func (s *TempPermissionService) canGrantPermission(checker *auth.Checker, perm model.Permission, resourceType string) bool {
	// 系统管理员可以授予任何权限
	if checker.IsAdmin() {
		return true
	}

	// 部门管理员/办公室管理员只能授予自己拥有的权限
	switch perm {
	case auth.PermScheduleManageDept, auth.PermScheduleViewDept:
		// 排班部门管理权限
		return checker.HasPermission(auth.PermScheduleManageDept)
	case auth.PermUserManageDept:
		// 部门用户管理权限
		return checker.HasPermission(auth.PermUserManageDept)
	case auth.PermScheduleViewAll:
		// 全局排班查看权限
		return checker.HasPermission(auth.PermScheduleViewAll)
	case auth.PermUserManageAll:
		// 全局用户管理权限
		return checker.IsOfficeAdmin() || checker.IsAdmin()
	default:
		return false
	}
}

// RevokePermission 撤销临时权限（带权限检查）
func (s *TempPermissionService) RevokePermission(adminID int, checker *auth.Checker, permID int) error {
	// 获取权限记录
	perm, err := s.dao.GetByID(permID)
	if err != nil {
		return errors.New("权限记录不存在")
	}

	// 系统管理员可以撤销任何权限
	if checker.IsAdmin() {
		return s.dao.Revoke(permID)
	}

	// 办公室管理员可以撤销任何权限
	if checker.IsOfficeAdmin() {
		return s.dao.Revoke(permID)
	}

	// 部门管理员只能撤销自己授予的，或本部门的权限
	if perm.GrantedBy == adminID {
		return s.dao.Revoke(permID)
	}

	// 检查被授权人是否在自己部门
	user, err := s.userDAO.GetByID(perm.UserID)
	if err != nil {
		return errors.New("用户不存在")
	}
	if user.Department == checker.GetDepartment() {
		return s.dao.Revoke(permID)
	}

	return errors.New("无权限撤销此授权")
}

// GetPermissionsByDepartment 按部门获取权限
func (s *TempPermissionService) GetPermissionsByDepartment(dept string) ([]model.TempPermissionView, error) {
	// 获取所有权限
	allPerms, err := s.dao.GetAllActive()
	if err != nil {
		return nil, err
	}

	// 筛选出部门相关的权限
	var filtered []model.UserPermissionTemp
	for _, perm := range allPerms {
		// 获取用户信息
		user, _ := s.userDAO.GetByID(perm.UserID)
		if user != nil && user.Department == dept {
			filtered = append(filtered, perm)
		}
	}

	return s.convertToView(filtered)
}

// GetUserTempPermissions 获取用户的临时权限
func (s *TempPermissionService) GetUserTempPermissions(userID int) ([]model.TempPermissionView, error) {
	perms, err := s.dao.GetActiveByUserID(userID)
	if err != nil {
		return nil, err
	}

	return s.convertToView(perms)
}

// GetAllActivePermissions 获取所有有效的临时权限
func (s *TempPermissionService) GetAllActivePermissions() ([]model.TempPermissionView, error) {
	perms, err := s.dao.GetAllActive()
	if err != nil {
		return nil, err
	}

	return s.convertToView(perms)
}

// CheckTempPermission 检查用户是否有指定临时权限
func (s *TempPermissionService) CheckTempPermission(userID int, perm model.Permission) (bool, error) {
	return s.dao.HasPermission(userID, perm)
}

// CheckTempPermissionWithResource 检查用户是否有指定临时权限（带资源）
func (s *TempPermissionService) CheckTempPermissionWithResource(userID int, perm model.Permission, resourceType string, resourceID int) (bool, error) {
	return s.dao.HasPermissionWithResource(userID, perm, resourceType, resourceID)
}

// CleanupExpiredPermissions 清理过期权限
func (s *TempPermissionService) CleanupExpiredPermissions() error {
	return s.dao.CleanupExpired()
}

// GetMyPermissions 获取"我的权限"列表
func (s *TempPermissionService) GetMyPermissions(userID int) ([]model.MyPermissionView, error) {
	perms, err := s.dao.GetActiveByUserID(userID)
	if err != nil {
		return nil, err
	}

	var result []model.MyPermissionView
	now := time.Now()

	for _, perm := range perms {
		expiresAt := time.Time(perm.ExpiresAt)
		daysLeft := int(expiresAt.Sub(now).Hours() / 24)
		if daysLeft < 0 {
			daysLeft = 0
		}

		// 获取资源名称
		resourceName := "全部"
		if perm.ResourceType == "dept" {
			resourceName = "部门ID: " + string(perm.ResourceID)
		} else if perm.ResourceType == "user" {
			resourceName = "用户ID: " + string(perm.ResourceID)
		}

		result = append(result, model.MyPermissionView{
			ID:             perm.ID,
			Permission:     perm.Permission,
			PermissionName: model.GetPermissionName(perm.Permission),
			ResourceType:   perm.ResourceType,
			ResourceName:   resourceName,
			Reason:         perm.Reason,
			ExpiresAt:      perm.ExpiresAt,
			DaysLeft:       daysLeft,
		})
	}

	return result, nil
}

// convertToView 转换为视图模型
func (s *TempPermissionService) convertToView(perms []model.UserPermissionTemp) ([]model.TempPermissionView, error) {
	var result []model.TempPermissionView
	now := time.Now()

	for _, perm := range perms {
		// 获取用户信息
		user, _ := s.userDAO.GetByID(perm.UserID)
		userName := ""
		if user != nil {
			userName = user.Name
		}

		// 获取授权人信息
		grantedBy, _ := s.userDAO.GetByID(perm.GrantedBy)
		grantedByName := ""
		if grantedBy != nil {
			grantedByName = grantedBy.Name
		}

		// 获取资源名称
		resourceName := "全部"
		if perm.ResourceType == "dept" {
			resourceName = "部门ID: " + string(perm.ResourceID)
		} else if perm.ResourceType == "user" {
			resourceName = "用户ID: " + string(perm.ResourceID)
		}

		result = append(result, model.TempPermissionView{
			ID:             perm.ID,
			UserID:         perm.UserID,
			UserName:       userName,
			Permission:     perm.Permission,
			PermissionName: model.GetPermissionName(perm.Permission),
			ResourceType:   perm.ResourceType,
			ResourceID:     perm.ResourceID,
			ResourceName:   resourceName,
			GrantedBy:      perm.GrantedBy,
			GrantedByName:  grantedByName,
			Reason:         perm.Reason,
			CreatedAt:      perm.CreatedAt,
			ExpiresAt:      perm.ExpiresAt,
			IsActive:       perm.IsActive,
			IsExpired:      perm.ExpiresAt.Before(now),
		})
	}

	return result, nil
}

// CanGrant 检查用户是否可以授权
func (s *TempPermissionService) CanGrant(checker *auth.Checker) bool {
	// 只有系统管理员可以授权
	return checker.IsAdmin()
}
