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

func (s *UserService) SetUserRole(adminID, targetUserID int, role string) error {
	admin, _ := s.userDAO.GetByID(adminID)
	if admin == nil || admin.Role != "admin" {
		return errors.New("无权限")
	}

	if role != "admin" && role != "user" {
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

// SetUserDepartment 设置用户部门
func (s *UserService) SetUserDepartment(adminID, targetUserID int, dept string) error {
	// 验证管理员权限
	admin, _ := s.userDAO.GetByID(adminID)
	if admin == nil || (admin.Role != model.RoleAdmin && admin.Department != "办公室") {
		return errors.New("无权限")
	}

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

// SetUserDeptRole 设置用户部门角色
func (s *UserService) SetUserDeptRole(adminID, targetUserID int, deptRole string) error {
	// 验证管理员权限
	admin, _ := s.userDAO.GetByID(adminID)
	if admin == nil {
		return errors.New("管理员不存在")
	}

	// 只有系统管理员或办公室管理员可以设置部门角色
	if admin.Role != model.RoleAdmin && !(admin.Department == "办公室" && admin.DeptRole == model.DeptRoleAdmin) {
		return errors.New("无权限设置部门角色")
	}

	// 验证角色是否有效
	if deptRole != model.DeptRoleAdmin && deptRole != model.DeptRoleMember {
		return errors.New("无效的部门角色")
	}

	// 获取目标用户
	targetUser, _ := s.userDAO.GetByID(targetUserID)
	if targetUser == nil {
		return errors.New("用户不存在")
	}

	// 部门管理员只能设置本部门用户的角色
	if admin.Role != model.RoleAdmin {
		if admin.Department != targetUser.Department {
			return errors.New("只能管理本部门用户")
		}
	}

	return s.userDAO.SetDeptRole(targetUserID, deptRole)
}

// GetUsersByDeptRole 获取指定部门角色的用户
func (s *UserService) GetUsersByDeptRole(deptRole string) ([]model.User, error) {
	return s.userDAO.GetByDeptRole(deptRole)
}

// AdminCreateUser 管理员创建用户
func (s *UserService) AdminCreateUser(adminID int, req *model.AdminCreateUserRequest) (*model.User, error) {
	// 验证管理员权限
	admin, _ := s.userDAO.GetByID(adminID)
	if admin == nil {
		return nil, errors.New("管理员不存在")
	}
	if admin.Role != model.RoleAdmin && !(admin.Department == "办公室" && admin.DeptRole == model.DeptRoleAdmin) {
		return nil, errors.New("无权限创建用户")
	}

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

// AdminUpdateUser 管理员更新用户
func (s *UserService) AdminUpdateUser(adminID, targetUserID int, req *model.AdminUpdateUserRequest) error {
	// 验证管理员权限
	admin, _ := s.userDAO.GetByID(adminID)
	if admin == nil {
		return errors.New("管理员不存在")
	}

	// 获取目标用户
	targetUser, err := s.userDAO.GetByID(targetUserID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 权限检查：系统管理员可以修改任何人，办公室管理员只能修改本部门用户
	if admin.Role != model.RoleAdmin {
		if !(admin.Department == "办公室" && admin.DeptRole == model.DeptRoleAdmin) {
			return errors.New("无权限修改用户")
		}
		// 办公室管理员只能修改本部门用户
		if targetUser.Department != admin.Department {
			return errors.New("只能修改本部门用户")
		}
		// 办公室管理员不能修改系统管理员
		if targetUser.Role == model.RoleAdmin {
			return errors.New("无权限修改系统管理员")
		}
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

// AdminDeleteUser 管理员删除用户
func (s *UserService) AdminDeleteUser(adminID, targetUserID int) error {
	// 验证管理员权限
	admin, _ := s.userDAO.GetByID(adminID)
	if admin == nil {
		return errors.New("管理员不存在")
	}

	// 获取目标用户
	targetUser, err := s.userDAO.GetByID(targetUserID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 不能删除自己
	if adminID == targetUserID {
		return errors.New("不能删除自己")
	}

	// 权限检查
	if admin.Role != model.RoleAdmin {
		if !(admin.Department == "办公室" && admin.DeptRole == model.DeptRoleAdmin) {
			return errors.New("无权限删除用户")
		}
		// 办公室管理员只能删除本部门用户
		if targetUser.Department != admin.Department {
			return errors.New("只能删除本部门用户")
		}
		// 办公室管理员不能删除系统管理员
		if targetUser.Role == model.RoleAdmin {
			return errors.New("无权限删除系统管理员")
		}
	}

	return s.userDAO.Delete(targetUserID)
}
