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

// CreatePending 创建「待入职」用户：显式置 employment_status='pending'（占位密码，不可登录）。
// 占位密码由调用方用 bcrypt("123456") 预先生成后写入 user.Password。
func (d *UserDAO) CreatePending(user *model.User) error {
	query := `INSERT INTO users (student_id, name, email, password, role, department, dept_role, employment_status)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'pending')`
	result, err := db.GetDB().Exec(query, user.StudentID, user.Name, user.Email, user.Password, user.Role, user.Department, user.DeptRole)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	user.ID = int(id)
	return nil
}

// GetByID 单点查询：不过滤 employment_status（返回真实行，便于 service 层区分 不存在/待激活/已离职）
func (d *UserDAO) GetByID(id int) (*model.User, error) {
	var user model.User
	query := `SELECT * FROM users WHERE id = ?`
	err := db.GetDB().Get(&user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByStudentID 单点查询：不过滤 employment_status
func (d *UserDAO) GetByStudentID(studentID string) (*model.User, error) {
	var user model.User
	query := `SELECT * FROM users WHERE student_id = ?`
	err := db.GetDB().Get(&user, query, studentID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail 单点查询：不过滤 employment_status
func (d *UserDAO) GetByEmail(email string) (*model.User, error) {
	var user model.User
	query := `SELECT * FROM users WHERE email = ?`
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

// List 列表查询：默认返回全部状态；employmentStatus 非空时按状态筛选（空串 = 不过滤）
func (d *UserDAO) List(employmentStatus string) ([]model.User, error) {
	var users []model.User
	query := `SELECT id, student_id, name, email, role, department, dept_role, employment_status, created_at FROM users`
	args := []interface{}{}
	if employmentStatus != "" {
		query += ` WHERE employment_status = ?`
		args = append(args, employmentStatus)
	}
	query += ` ORDER BY id DESC`
	err := db.GetDB().Select(&users, query, args...)
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

// Delete 将用户标记为离职（调用存储过程 sp_mark_user_resigned），而非物理删除/禁用。
// 存储过程签名：(p_user_id, p_admin_id)。
func (d *UserDAO) Delete(id int, adminID int) error {
	query := `CALL sp_mark_user_resigned(?, ?)`
	_, err := db.GetDB().Exec(query, id, adminID)
	return err
}

// Count 计数：默认全状态；employmentStatus 非空时按状态筛选
func (d *UserDAO) Count(employmentStatus string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM users`
	args := []interface{}{}
	if employmentStatus != "" {
		query += ` WHERE employment_status = ?`
		args = append(args, employmentStatus)
	}
	err := db.GetDB().Get(&count, query, args...)
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

// GetUsersByIDs 按ID批量获取用户：默认全状态；employmentStatus 非空时按状态筛选
func (d *UserDAO) GetUsersByIDs(userIDs []int, employmentStatus string) ([]model.User, error) {
	if len(userIDs) == 0 {
		return []model.User{}, nil
	}
	query := `SELECT id, student_id, name, email, role, department, dept_role, employment_status, created_at FROM users WHERE id IN (?)`
	args := []interface{}{userIDs}
	if employmentStatus != "" {
		query += ` AND employment_status = ?`
		args = append(args, employmentStatus)
	}
	query, args, err := sqlx.In(query, args...)
	if err != nil {
		return nil, err
	}
	query = db.GetDB().Rebind(query)
	var users []model.User
	if err := db.GetDB().Select(&users, query, args...); err != nil {
		return nil, err
	}
	return users, nil
}

// ListByDepartment 按部门获取用户列表：默认全状态；employmentStatus 非空时按状态筛选
func (d *UserDAO) ListByDepartment(dept, employmentStatus string) ([]model.User, error) {
	var users []model.User
	query := `SELECT id, student_id, name, email, role, department, dept_role, employment_status, created_at 
			  FROM users WHERE department = ?`
	args := []interface{}{dept}
	if employmentStatus != "" {
		query += ` AND employment_status = ?`
		args = append(args, employmentStatus)
	}
	query += ` ORDER BY id DESC`
	err := db.GetDB().Select(&users, query, args...)
	return users, err
}

// ListByDepartments 按多个部门获取用户列表：默认全状态；employmentStatus 非空时按状态筛选
func (d *UserDAO) ListByDepartments(depts []string, employmentStatus string) ([]model.User, error) {
	if len(depts) == 0 {
		return d.List(employmentStatus)
	}

	// 使用 sqlx.In 处理 IN 查询
	query := `SELECT id, student_id, name, email, role, department, dept_role, employment_status, created_at 
			  FROM users WHERE department IN (?)`
	args := []interface{}{depts}
	if employmentStatus != "" {
		query += ` AND employment_status = ?`
		args = append(args, employmentStatus)
	}
	query += ` ORDER BY id DESC`
	query, args, err := sqlx.In(query, args...)
	if err != nil {
		return nil, err
	}

	var users []model.User
	err = db.GetDB().Select(&users, query, args...)
	return users, err
}

// ListByFilter 根据筛选条件获取用户列表：默认全状态；employmentStatus 非空时按状态筛选
func (d *UserDAO) ListByFilter(filter model.UserListFilter, employmentStatus string) ([]model.User, error) {
	query := `SELECT id, student_id, name, email, role, department, dept_role, employment_status, created_at 
			  FROM users WHERE 1=1`
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

	if employmentStatus != "" {
		query += ` AND employment_status = ?`
		args = append(args, employmentStatus)
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

// GetByDeptRole 获取指定部门角色的用户：默认全状态；employmentStatus 非空时按状态筛选
func (d *UserDAO) GetByDeptRole(deptRole, employmentStatus string) ([]model.User, error) {
	var users []model.User
	query := `SELECT id, student_id, name, email, role, department, dept_role, employment_status, created_at 
			  FROM users WHERE dept_role = ?`
	args := []interface{}{deptRole}
	if employmentStatus != "" {
		query += ` AND employment_status = ?`
		args = append(args, employmentStatus)
	}
	query += ` ORDER BY id DESC`
	err := db.GetDB().Select(&users, query, args...)
	return users, err
}

// GetByDepartmentAndRole 按部门和角色获取用户：默认全状态；employmentStatus 非空时按状态筛选
func (d *UserDAO) GetByDepartmentAndRole(department, deptRole, employmentStatus string) ([]model.User, error) {
	var users []model.User
	query := `SELECT id, student_id, name, email, role, department, dept_role, employment_status, created_at 
			  FROM users WHERE department = ? AND dept_role = ?`
	args := []interface{}{department, deptRole}
	if employmentStatus != "" {
		query += ` AND employment_status = ?`
		args = append(args, employmentStatus)
	}
	query += ` ORDER BY id DESC`
	err := db.GetDB().Select(&users, query, args...)
	return users, err
}

// GetByRole 按系统角色获取用户：默认全状态；employmentStatus 非空时按状态筛选
func (d *UserDAO) GetByRole(role, employmentStatus string) ([]model.User, error) {
	var users []model.User
	query := `SELECT id, student_id, name, email, role, department, dept_role, employment_status, created_at 
			  FROM users WHERE role = ?`
	args := []interface{}{role}
	if employmentStatus != "" {
		query += ` AND employment_status = ?`
		args = append(args, employmentStatus)
	}
	query += ` ORDER BY id DESC`
	err := db.GetDB().Select(&users, query, args...)
	return users, err
}
