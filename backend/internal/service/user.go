package service

import (
	"errors"
	"schedule-system-v2/backend/internal/dao"
	"schedule-system-v2/backend/internal/model"
	"schedule-system-v2/backend/internal/utils"
)

type UserService struct {
	userDAO    *dao.UserDAO
	systemDAO  *dao.SystemDAO
}

func NewUserService() *UserService {
	return &UserService{
		userDAO:   dao.NewUserDAO(),
		systemDAO: dao.NewSystemDAO(),
	}
}

func (s *UserService) Register(req *model.RegisterRequest) (*model.User, error) {
	existing, _ := s.userDAO.GetByStudentID(req.StudentID)
	if existing != nil {
		return nil, errors.New("学号已注册")
	}

	existingEmail, _ := s.userDAO.GetByEmail(req.Email)
	if existingEmail != nil {
		return nil, errors.New("邮箱已被使用")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	role := "user"
	deptRole := model.DeptRoleMember
	initialized, _ := s.systemDAO.IsInitialized()
	if !initialized {
		role = "admin"
		deptRole = model.DeptRoleAdmin
	}

	user := &model.User{
		StudentID:  req.StudentID,
		Name:       req.Name,
		Email:      req.Email,
		Password:   hashedPassword,
		Role:       role,
		Department: req.Department,
		DeptRole:   deptRole,
		IsActive:   true,
	}

	if err := s.userDAO.Create(user); err != nil {
		return nil, err
	}

	if !initialized {
		s.systemDAO.SetInitialized()
	}

	return user, nil
}

func (s *UserService) Login(req *model.LoginRequest) (*model.LoginResponse, error) {
	user, err := s.userDAO.GetByStudentID(req.StudentID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, errors.New("密码错误")
	}

	token, err := utils.GenerateToken(user.ID, user.StudentID, user.Role, user.Department, user.DeptRole)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		Token: token,
		User: model.UserInfo{
			ID:         user.ID,
			StudentID:  user.StudentID,
			Name:       user.Name,
			Email:      user.Email,
			Role:       user.Role,
			Department: user.Department,
			DeptRole:   user.DeptRole,
		},
	}, nil
}

func (s *UserService) GetUserByID(id int) (*model.User, error) {
	return s.userDAO.GetByID(id)
}

func (s *UserService) GetUserList() ([]model.User, error) {
	return s.userDAO.List()
}

func (s *UserService) UpdateUser(userID int, req *model.UpdateUserRequest) error {
	user, err := s.userDAO.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	user.Name = req.Name
	user.Email = req.Email
	return s.userDAO.Update(user)
}

func (s *UserService) SetUserRole(targetUserID int, role string) error {
	if role != model.RoleAdmin && role != model.RoleUser {
		return errors.New("无效的角色")
	}

	return s.userDAO.SetRole(targetUserID, role)
}

func (s *UserService) DeleteUser(id int) error {
	return s.userDAO.Delete(id)
}

// GetUserListByDepartment 按部门获取用户列表
func (s *UserService) GetUserListByDepartment(dept string) ([]model.User, error) {
	return s.userDAO.ListByDepartment(dept)
}

// GetUsersByDepts 按多个部门获取用户
func (s *UserService) GetUsersByDepts(depts []string) ([]model.User, error) {
	return s.userDAO.ListByDepartments(depts)
}

// GetUsersByFilter 根据筛选条件获取用户
func (s *UserService) GetUsersByFilter(filter model.UserListFilter) ([]model.User, error) {
	return s.userDAO.ListByFilter(filter)
}

// SetUserDepartment 设置用户部门（权限由 handler 层校验）
func (s *UserService) SetUserDepartment(targetUserID int, dept string) error {
	// 验证部门是否有效
	validDept := false
	for _, d := range model.Departments {
		if d == dept {
			validDept = true
			break
		}
	}
	if !validDept {
		return errors.New("无效的部门")
	}

	return s.userDAO.SetDepartment(targetUserID, dept)
}

// SetUserDeptRole 设置用户部门角色（权限由 handler 层校验）
func (s *UserService) SetUserDeptRole(targetUserID int, deptRole string) error {
	// 验证角色是否有效
	if deptRole != model.DeptRoleAdmin && deptRole != model.DeptRoleMember {
		return errors.New("无效的部门角色")
	}

	return s.userDAO.SetDeptRole(targetUserID, deptRole)
}

// GetUsersByDeptRole 获取指定部门角色的用户
func (s *UserService) GetUsersByDeptRole(deptRole string) ([]model.User, error) {
	return s.userDAO.GetByDeptRole(deptRole)
}

// AdminCreateUser 管理员创建用户（权限由 handler 层校验）
func (s *UserService) AdminCreateUser(req *model.AdminCreateUserRequest) (*model.User, error) {
	// 检查学号是否已存在
	existing, _ := s.userDAO.GetByStudentID(req.StudentID)
	if existing != nil {
		return nil, errors.New("学号已注册")
	}

	// 检查邮箱是否已被使用
	existingEmail, _ := s.userDAO.GetByEmail(req.Email)
	if existingEmail != nil {
		return nil, errors.New("邮箱已被使用")
	}

	// 设置默认密码
	password := req.Password
	if password == "" {
		password = "123456"
	}
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		StudentID:  req.StudentID,
		Name:       req.Name,
		Email:      req.Email,
		Password:   hashedPassword,
		Role:       req.Role,
		Department: req.Department,
		DeptRole:   req.DeptRole,
		IsActive:   true,
	}

	if err := s.userDAO.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// AdminUpdateUser 管理员更新用户（权限和范围由 handler 层校验）
func (s *UserService) AdminUpdateUser(targetUserID int, req *model.AdminUpdateUserRequest) error {
	targetUser, err := s.userDAO.GetByID(targetUserID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 如果邮箱变更，检查是否冲突
	if targetUser.Email != req.Email {
		existingEmail, _ := s.userDAO.GetByEmail(req.Email)
		if existingEmail != nil && existingEmail.ID != targetUserID {
			return errors.New("邮箱已被使用")
		}
	}

	// 更新用户信息
	targetUser.Name = req.Name
	targetUser.Email = req.Email
	targetUser.Department = req.Department
	targetUser.Role = req.Role
	targetUser.DeptRole = req.DeptRole

	return s.userDAO.Update(targetUser)
}

// AdminDeleteUser 管理员删除用户（实际为禁用用户）
//
// 设计说明：
// 本系统采用"软删除"策略，即通过设置 is_active = 0 来标记用户为禁用状态，而非物理删除。
// 原因如下：
// 1. 数据完整性：用户可能关联历史值班记录、审批记录等，删除用户会导致这些记录失去关联
// 2. 审计追溯：需要保留"谁值的班"、"谁排的班"、"谁审批的"等操作历史
// 3. 恢复能力：误删后可以恢复账号，无需重新录入用户信息和历史数据
//
// 如果需要物理删除，需要：
// - 归档用户关键信息（姓名、学号）到历史表
// - 处理所有关联数据的外键约束
// - 或者匿名化处理后删除
func (s *UserService) AdminDeleteUser(operatorID, targetUserID int) error {
	// 不能删除自己
	if operatorID == targetUserID {
		return errors.New("不能删除自己")
	}

	return s.userDAO.Delete(targetUserID)
}

// ChangePassword 修改用户密码
func (s *UserService) ChangePassword(userID int, oldPassword, newPassword string) error {
	user, err := s.userDAO.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 验证旧密码
	if !utils.CheckPassword(oldPassword, user.Password) {
		return errors.New("旧密码错误")
	}

	// DAO层会进行bcrypt加密，这里直接传明文
	return s.userDAO.UpdatePassword(userID, newPassword)
}
