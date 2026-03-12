package dao

import (
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	"schedule-system-v2/backend/internal/db"
	"schedule-system-v2/backend/internal/model"
)

type UserDAO struct{}

func NewUserDAO() *UserDAO {
	return &UserDAO{}
}

func (d *UserDAO) Create(user *model.User) error {
	query := `INSERT INTO users (student_id, name, email, password, role, department, dept_role) VALUES (?, ?, ?, ?, ?, ?, ?)`
	result, err := db.GetDB().Exec(query, user.StudentID, user.Name, user.Email, user.Password, user.Role, user.Department, user.DeptRole)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	user.ID = int(id)
	return nil
}

func (d *UserDAO) GetByID(id int) (*model.User, error) {
	var user model.User
	query := `SELECT * FROM users WHERE id = ? AND is_active = 1`
	err := db.GetDB().Get(&user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (d *UserDAO) GetByStudentID(studentID string) (*model.User, error) {
	var user model.User
	query := `SELECT * FROM users WHERE student_id = ? AND is_active = 1`
	err := db.GetDB().Get(&user, query, studentID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (d *UserDAO) GetByEmail(email string) (*model.User, error) {
	var user model.User
	query := `SELECT * FROM users WHERE email = ? AND is_active = 1`
	err := db.GetDB().Get(&user, query, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// IsEmpty 检查数据库是否为空（没有用户数据）
func (d *UserDAO) IsEmpty() (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users`
	err := db.GetDB().Get(&count, query)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (d *UserDAO) List() ([]model.User, error) {
	var users []model.User
	query := `SELECT id, student_id, name, email, role, department, dept_role, is_active, created_at FROM users WHERE is_active = 1 ORDER BY id DESC`
	err := db.GetDB().Select(&users, query)
	return users, err
}

func (d *UserDAO) Update(user *model.User) error {
	query := `UPDATE users SET name = ?, email = ?, role = ?, department = ?, dept_role = ? WHERE id = ?`
	_, err := db.GetDB().Exec(query, user.Name, user.Email, user.Role, user.Department, user.DeptRole, user.ID)
	return err
}

func (d *UserDAO) UpdatePassword(id int, password string) error {
	// 使用 bcrypt 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	query := `UPDATE users SET password = ? WHERE id = ?`
	_, err = db.GetDB().Exec(query, string(hashedPassword), id)
	return err
}

func (d *UserDAO) Delete(id int) error {
	query := `UPDATE users SET is_active = 0 WHERE id = ?`
	_, err := db.GetDB().Exec(query, id)
	return err
}

func (d *UserDAO) Count() (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM users WHERE is_active = 1`
	err := db.GetDB().Get(&count, query)
	return count, err
}

func (d *UserDAO) SetRole(id int, role string) error {
	query := `UPDATE users SET role = ? WHERE id = ?`
	_, err := db.GetDB().Exec(query, role, id)
	return err
}

// SetDeptRole 设置用户部门角色
func (d *UserDAO) SetDeptRole(userID int, deptRole string) error {
	query := `UPDATE users SET dept_role = ? WHERE id = ?`
	_, err := db.GetDB().Exec(query, deptRole, userID)
	return err
}

// SetDepartment 设置用户部门
func (d *UserDAO) SetDepartment(userID int, dept string) error {
	query := `UPDATE users SET department = ? WHERE id = ?`
	_, err := db.GetDB().Exec(query, dept, userID)
	return err
}

// ListByDepartment 按部门获取用户列表
func (d *UserDAO) ListByDepartment(dept string) ([]model.User, error) {
	var users []model.User
	query := `SELECT id, student_id, name, email, role, department, dept_role, is_active, created_at 
			  FROM users WHERE is_active = 1 AND department = ? ORDER BY id DESC`
	err := db.GetDB().Select(&users, query, dept)
	return users, err
}

// ListByDepartments 按多个部门获取用户列表
func (d *UserDAO) ListByDepartments(depts []string) ([]model.User, error) {
	if len(depts) == 0 {
		return d.List()
	}

	// 使用 sqlx.In 处理 IN 查询
	query := `SELECT id, student_id, name, email, role, department, dept_role, is_active, created_at 
			  FROM users WHERE is_active = 1 AND department IN (?) ORDER BY id DESC`
	query, args, err := sqlx.In(query, depts)
	if err != nil {
		return nil, err
	}

	var users []model.User
	err = db.GetDB().Select(&users, query, args...)
	return users, err
}

// ListByFilter 根据筛选条件获取用户列表
func (d *UserDAO) ListByFilter(filter model.UserListFilter) ([]model.User, error) {
	query := `SELECT id, student_id, name, email, role, department, dept_role, is_active, created_at 
			  FROM users WHERE is_active = 1`
	var args []interface{}

	if len(filter.Departments) > 0 {
		query += ` AND department IN (?)`
		args = append(args, filter.Departments)
	}

	if filter.Role != "" {
		query += ` AND role = ?`
		args = append(args, filter.Role)
	}

	if filter.DeptRole != "" {
		query += ` AND dept_role = ?`
		args = append(args, filter.DeptRole)
	}

	query += ` ORDER BY id DESC`

	// 如果有部门筛选，使用 sqlx.In 处理
	if len(filter.Departments) > 0 {
		var err error
		query, args, err = sqlx.In(query, args...)
		if err != nil {
			return nil, err
		}
	}

	var users []model.User
	err := db.GetDB().Select(&users, query, args...)
	return users, err
}

// GetByDeptRole 获取指定部门角色的用户
func (d *UserDAO) GetByDeptRole(deptRole string) ([]model.User, error) {
	var users []model.User
	query := `SELECT id, student_id, name, email, role, department, dept_role, is_active, created_at 
			  FROM users WHERE is_active = 1 AND dept_role = ? ORDER BY id DESC`
	err := db.GetDB().Select(&users, query, deptRole)
	return users, err
}

// GetByDepartmentAndRole 按部门和角色获取用户
func (d *UserDAO) GetByDepartmentAndRole(department, deptRole string) ([]model.User, error) {
	var users []model.User
	query := `SELECT id, student_id, name, email, role, department, dept_role, is_active, created_at 
			  FROM users WHERE is_active = 1 AND department = ? AND dept_role = ? ORDER BY id DESC`
	err := db.GetDB().Select(&users, query, department, deptRole)
	return users, err
}

// GetByRole 按系统角色获取用户
func (d *UserDAO) GetByRole(role string) ([]model.User, error) {
	var users []model.User
	query := `SELECT id, student_id, name, email, role, department, dept_role, is_active, created_at 
			  FROM users WHERE is_active = 1 AND role = ? ORDER BY id DESC`
	err := db.GetDB().Select(&users, query, role)
	return users, err
}
