package service

import (
	"database/sql"
	"fmt"
	"log"
	"schedule-system-v2/backend/internal/config"
	"schedule-system-v2/backend/internal/db"
	"schedule-system-v2/backend/internal/model"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type SystemService struct{}

func NewSystemService() *SystemService {
	return &SystemService{}
}

// TestDBConnection 测试数据库连接并自动创建数据库
func (s *SystemService) TestDBConnection(cfg *config.DatabaseConfig) error {
	log.Printf("[DEBUG] TestDBConnection: host=%s, port=%s, user=%s, dbname=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.DBName)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port)

	log.Printf("[DEBUG] DSN: %s", dsn)

	database, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}
	defer database.Close()

	if err := database.Ping(); err != nil {
		return fmt.Errorf("Ping失败: %v", err)
	}

	// 创建数据库（如果不存在）
	_, err = database.Exec(fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		cfg.DBName))
	if err != nil {
		return fmt.Errorf("创建数据库失败: %v", err)
	}

	return nil
}

// SaveDBConfig 保存数据库配置（不标记为已安装）
// 注意：配置文件必须预先存在
func (s *SystemService) SaveDBConfig(cfg *config.DatabaseConfig) error {
	// 确保 charset 有默认值
	if cfg.Charset == "" {
		cfg.Charset = "utf8mb4"
	}

	// 加载现有配置
	conf, err := config.LoadConfig(config.ConfigFilePath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %v", err)
	}

	// 更新数据库配置
	conf.Database = *cfg
	// 确保仍然是未安装状态（只有创建管理员后才标记为已安装）
	conf.Installed = false

	return config.SaveConfig(conf)
}

// MarkInstalled 标记系统已安装完成
func (s *SystemService) MarkInstalled() error {
	// 加载现有配置
	conf, err := config.LoadConfig(config.ConfigFilePath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %v", err)
	}

	// 标记为已安装
	conf.Installed = true

	return config.SaveConfig(conf)
}

// CheckDatabaseEmpty 检查数据库是否为空（没有表）
func (s *SystemService) CheckDatabaseEmpty(cfg *config.DatabaseConfig) (bool, []string, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	database, err := sql.Open("mysql", dsn)
	if err != nil {
		return false, nil, err
	}
	defer database.Close()

	// 查询所有表
	rows, err := database.Query("SHOW TABLES")
	if err != nil {
		return false, nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		tables = append(tables, tableName)
	}

	return len(tables) == 0, tables, nil
}

// InitDatabaseTables 初始化数据库表
func (s *SystemService) InitDatabaseTables(cfg *config.DatabaseConfig, force bool) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&multiStatements=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	// 如果强制覆盖，先删除所有表
	if force {
		database, err := sql.Open("mysql", dsn)
		if err != nil {
			return fmt.Errorf("连接失败: %v", err)
		}
		if err := s.dropAllTables(database); err != nil {
			database.Close()
			return fmt.Errorf("清理旧表失败: %v", err)
		}
		database.Close()
	}

	// 执行建表SQL（使用新的连接）
	database, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}
	defer database.Close()

	if err := s.executeInitSQL(database); err != nil {
		return fmt.Errorf("执行建表SQL失败: %v", err)
	}

	// 关闭临时连接
	database.Close()

	// 重新初始化全局数据库连接池
	if err := db.InitDB(cfg); err != nil {
		return fmt.Errorf("初始化数据库连接池失败: %v", err)
	}

	return nil
}

// dropAllTables 删除所有表（危险操作）
func (s *SystemService) dropAllTables(database *sql.DB) error {
	rows, err := database.Query("SHOW TABLES")
	if err != nil {
		return err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		tables = append(tables, tableName)
	}

	// 禁用外键检查
	database.Exec("SET FOREIGN_KEY_CHECKS = 0")
	for _, table := range tables {
		database.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
	}
	database.Exec("SET FOREIGN_KEY_CHECKS = 1")

	return nil
}

// executeInitSQL 执行初始化SQL
func (s *SystemService) executeInitSQL(database *sql.DB) error {
	// 系统配置表
	_, err := database.Exec(`CREATE TABLE IF NOT EXISTS system_config (
		id INT PRIMARY KEY AUTO_INCREMENT,
		config_key VARCHAR(50) NOT NULL UNIQUE,
		config_value TEXT,
		description VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB COMMENT='系统配置表'`)
	if err != nil {
		return fmt.Errorf("创建system_config表失败: %v", err)
	}

	// 用户表
	_, err = database.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INT PRIMARY KEY AUTO_INCREMENT,
		student_id VARCHAR(20) NOT NULL UNIQUE COMMENT '学号/工号',
		name VARCHAR(50) NOT NULL COMMENT '姓名',
		email VARCHAR(100) NOT NULL UNIQUE,
		password VARCHAR(255) NOT NULL COMMENT '加密密码',
		role ENUM('admin', 'user') DEFAULT 'user' COMMENT '系统角色',
		department VARCHAR(50) DEFAULT '科普部' COMMENT '所属部门',
		dept_role ENUM('dept_admin', 'dept_member') DEFAULT 'dept_member' COMMENT '部门角色',
		is_active TINYINT DEFAULT 1 COMMENT '是否启用',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_role (role),
		INDEX idx_department (department),
		INDEX idx_dept_role (dept_role),
		INDEX idx_active (is_active)
	) ENGINE=InnoDB COMMENT='用户表'`)
	if err != nil {
		return fmt.Errorf("创建users表失败: %v", err)
	}

	// 无课时间表
	_, err = database.Exec(`CREATE TABLE IF NOT EXISTS availability (
		id INT PRIMARY KEY AUTO_INCREMENT,
		user_id INT NOT NULL COMMENT '用户ID',
		week TINYINT NOT NULL COMMENT '第几周 1-30',
		weekday TINYINT NOT NULL COMMENT '星期几 1-5',
		period TINYINT NOT NULL COMMENT '节次 1-4',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE KEY uk_avail (user_id, week, weekday, period),
		INDEX idx_time_user (week, weekday, period, user_id),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	) ENGINE=InnoDB COMMENT='无课时间表'`)
	if err != nil {
		return fmt.Errorf("创建availability表失败: %v", err)
	}

	// 值班记录表
	_, err = database.Exec(`CREATE TABLE IF NOT EXISTS duty_records (
		id INT PRIMARY KEY AUTO_INCREMENT,
		week TINYINT NOT NULL COMMENT '第几周 1-30',
		weekday TINYINT NOT NULL COMMENT '星期几 1-5',
		period TINYINT NOT NULL COMMENT '节次 1-4',
		user_id INT NOT NULL COMMENT '值班人员ID',
		department VARCHAR(50) COMMENT '部门',
		assigned_by INT COMMENT '排班人ID',
		status ENUM('pending', 'confirmed', 'completed', 'cancelled') DEFAULT 'pending',
		remark VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_week (week),
		INDEX idx_user (user_id),
		INDEX idx_department (department),
		INDEX idx_time (week, weekday, period),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (assigned_by) REFERENCES users(id) ON DELETE SET NULL
	) ENGINE=InnoDB COMMENT='值班记录表'`)
	if err != nil {
		return fmt.Errorf("创建duty_records表失败: %v", err)
	}

	// 值班次数统计表
	_, err = database.Exec(`CREATE TABLE IF NOT EXISTS duty_counters (
		user_id INT PRIMARY KEY COMMENT '用户ID',
		total_count INT DEFAULT 0 COMMENT '总值班次数',
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	) ENGINE=InnoDB COMMENT='值班次数统计'`)
	if err != nil {
		return fmt.Errorf("创建duty_counters表失败: %v", err)
	}

	// 排班设置表
	_, err = database.Exec(`CREATE TABLE IF NOT EXISTS schedule_settings (
		id INT PRIMARY KEY AUTO_INCREMENT,
		admin_id INT NOT NULL UNIQUE,
		current_week INT DEFAULT 1 COMMENT '当前周次',
		auto_increment TINYINT(1) DEFAULT 0 COMMENT '是否自动递增周次',
		need_per_cell INT DEFAULT 2,
		min_per_cell INT DEFAULT 0,
		max_per_day INT DEFAULT 1,
		max_per_week INT DEFAULT 2,
		export_title VARCHAR(255) DEFAULT '第{week}周排班表',
		semester_start_date DATE DEFAULT NULL COMMENT '学期起始日',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB COMMENT='排班设置表'`)
	if err != nil {
		return fmt.Errorf("创建schedule_settings表失败: %v", err)
	}

	// 导出模板表
	_, err = database.Exec(`CREATE TABLE IF NOT EXISTS export_templates (
		id INT PRIMARY KEY AUTO_INCREMENT,
		admin_id INT NOT NULL COMMENT '创建者ID',
		name VARCHAR(100) NOT NULL COMMENT '模板名称',
		description VARCHAR(255) DEFAULT '' COMMENT '模板描述',
		config JSON NOT NULL COMMENT '模板配置：表头、行列占位符等',
		is_default TINYINT(1) DEFAULT 0 COMMENT '是否为默认模板',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_admin_id (admin_id)
	) ENGINE=InnoDB COMMENT='Excel导出模板表'`)
	if err != nil {
		return fmt.Errorf("创建export_templates表失败: %v", err)
	}

	// 每周值班分工表
	_, err = database.Exec(`CREATE TABLE IF NOT EXISTS weekly_duty_assignments (
		id INT PRIMARY KEY AUTO_INCREMENT,
		week INT NOT NULL COMMENT '周次',
		department VARCHAR(50) NOT NULL COMMENT '部门',
		weekday INT NOT NULL COMMENT '星期几(1-5)',
		is_assigned BOOLEAN DEFAULT TRUE COMMENT '是否安排值班',
		created_by INT NOT NULL COMMENT '创建人ID',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE KEY uk_week_dept_weekday (week, department, weekday),
		INDEX idx_week (week),
		INDEX idx_department (department)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='每周值班分工表'`)
	if err != nil {
		return fmt.Errorf("创建weekly_duty_assignments表失败: %v", err)
	}

	// 用户临时权限表
	_, err = database.Exec(`CREATE TABLE IF NOT EXISTS user_permissions_temp (
		id INT PRIMARY KEY AUTO_INCREMENT,
		user_id INT NOT NULL COMMENT '被授权用户ID',
		permission VARCHAR(50) NOT NULL COMMENT '权限代码',
		resource_type VARCHAR(20) DEFAULT 'all' COMMENT '资源类型(all/dept/user)',
		resource_id INT DEFAULT 0 COMMENT '资源ID(部门ID或用户ID)',
		granted_by INT NOT NULL COMMENT '授权人ID',
		reason VARCHAR(255) COMMENT '授权原因',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP NULL DEFAULT NULL COMMENT '过期时间',
		is_active BOOLEAN DEFAULT TRUE COMMENT '是否有效',
		INDEX idx_user_id (user_id),
		INDEX idx_expires_at (expires_at),
		INDEX idx_is_active (is_active)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户临时权限表'`)
	if err != nil {
		return fmt.Errorf("创建user_permissions_temp表失败: %v", err)
	}

	// SMTP配置表
	_, err = database.Exec(`CREATE TABLE IF NOT EXISTS smtp_config (
		id INT PRIMARY KEY AUTO_INCREMENT,
		host VARCHAR(255) NOT NULL COMMENT 'SMTP服务器地址',
		port INT NOT NULL COMMENT '端口',
		username VARCHAR(255) NOT NULL COMMENT '用户名/邮箱',
		password VARCHAR(255) NOT NULL COMMENT '密码/授权码',
		from_name VARCHAR(100) NOT NULL COMMENT '发件人显示名称',
		from_email VARCHAR(255) NOT NULL COMMENT '发件人邮箱',
		use_tls BOOLEAN DEFAULT TRUE COMMENT '是否使用TLS',
		use_ssl BOOLEAN DEFAULT FALSE COMMENT '是否使用SSL',
		is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SMTP配置表'`)
	if err != nil {
		return fmt.Errorf("创建smtp_config表失败: %v", err)
	}

	// 密码重置令牌表
	_, err = database.Exec(`CREATE TABLE IF NOT EXISTS password_reset_tokens (
		id INT PRIMARY KEY AUTO_INCREMENT,
		user_id INT NOT NULL COMMENT '用户ID',
		email VARCHAR(255) NOT NULL COMMENT '邮箱',
		token VARCHAR(255) NOT NULL COMMENT '重置令牌',
		expires_at TIMESTAMP NULL DEFAULT NULL COMMENT '过期时间',
		is_used BOOLEAN DEFAULT FALSE COMMENT '是否已使用',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_token (token),
		INDEX idx_expires_at (expires_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='密码重置令牌表'`)
	if err != nil {
		return fmt.Errorf("创建password_reset_tokens表失败: %v", err)
	}

	// ========== 申请审批系统表 ==========

	// 申请类型表
	_, err = database.Exec(`CREATE TABLE IF NOT EXISTS application_types (
		id INT PRIMARY KEY AUTO_INCREMENT,
		code VARCHAR(50) NOT NULL UNIQUE COMMENT '类型代码',
		name VARCHAR(100) NOT NULL COMMENT '类型名称',
		description TEXT COMMENT '类型描述',
		config JSON COMMENT '类型配置JSON（字段定义、审批流程等）',
		is_active TINYINT(1) DEFAULT 1 COMMENT '是否启用',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='申请类型表'`)
	if err != nil {
		return fmt.Errorf("创建application_types表失败: %v", err)
	}

	// 申请表
	_, err = database.Exec(`CREATE TABLE IF NOT EXISTS applications (
		id INT PRIMARY KEY AUTO_INCREMENT,
		application_no VARCHAR(50) NOT NULL UNIQUE COMMENT '申请单号',
		type_code VARCHAR(50) NOT NULL COMMENT '申请类型代码',
		applicant_id INT NOT NULL COMMENT '申请人ID',
		department VARCHAR(50) COMMENT '申请人部门',
		title VARCHAR(255) NOT NULL COMMENT '申请标题',
		content TEXT COMMENT '申请内容描述',
		data JSON COMMENT '申请数据JSON（不同类型的具体数据）',
		status TINYINT NOT NULL DEFAULT 0 COMMENT '状态: 0待审批 1审批中 2已通过 3已拒绝 4已撤回',
		current_level INT DEFAULT 1 COMMENT '当前审批层级',
		total_levels INT DEFAULT 1 COMMENT '总审批层级',
		approver_id INT COMMENT '当前审批人ID',
		result TEXT COMMENT '审批结果/备注',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_applicant (applicant_id),
		INDEX idx_type (type_code),
		INDEX idx_status (status),
		INDEX idx_department (department)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='申请表'`)
	if err != nil {
		return fmt.Errorf("创建applications表失败: %v", err)
	}

	// 审批历史记录表
	_, err = database.Exec(`CREATE TABLE IF NOT EXISTS application_approvals (
		id INT PRIMARY KEY AUTO_INCREMENT,
		application_id INT NOT NULL COMMENT '申请ID',
		level INT NOT NULL COMMENT '审批层级',
		approver_id INT NOT NULL COMMENT '审批人ID',
		action TINYINT NOT NULL COMMENT '操作: 1通过 2拒绝 3转办 4评论',
		comment TEXT COMMENT '审批意见',
		data JSON COMMENT '附加数据',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_application (application_id),
		INDEX idx_approver (approver_id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='审批历史表'`)
	if err != nil {
		return fmt.Errorf("创建application_approvals表失败: %v", err)
	}

	// 审批人配置表
	_, err = database.Exec(`CREATE TABLE IF NOT EXISTS application_approvers (
		id INT PRIMARY KEY AUTO_INCREMENT,
		type_code VARCHAR(50) NOT NULL COMMENT '申请类型',
		level INT NOT NULL COMMENT '层级',
		role_type ENUM('admin', 'dept_admin', 'office_admin', 'specific') NOT NULL COMMENT '审批角色类型',
		specific_user_id INT COMMENT '具体用户ID（role_type=specific时使用）',
		is_active TINYINT(1) DEFAULT 1,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE KEY uk_type_level (type_code, level),
		INDEX idx_type (type_code)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='审批人配置表'`)
	if err != nil {
		return fmt.Errorf("创建application_approvers表失败: %v", err)
	}

	// 插入默认申请类型
	_, err = database.Exec(`INSERT INTO application_types (code, name, description, config) VALUES
		('temp_permission', '临时权限申请', '申请临时管理权限', '{
			"fields": [
				{"name": "permission", "label": "申请权限", "type": "select", "required": true, "options": [
					{"value": "schedule:manage:dept", "label": "部门排班管理"},
					{"value": "user:manage:dept", "label": "部门用户管理"},
					{"value": "schedule:view:all", "label": "全局排班查看"}
				]},
				{"name": "expires_at", "label": "期望到期日", "type": "date", "required": true},
				{"name": "reason", "label": "申请原因", "type": "textarea", "required": true}
			],
			"flow": [
				{"level": 1, "role": "dept_admin", "label": "部门管理员审批"},
				{"level": 2, "role": "office_admin", "label": "办公室管理员审批", "optional": true}
			],
			"auto_execute": true
		}'),
		('leave', '请假申请', '申请休假', '{
			"fields": [
				{"name": "leave_type", "label": "请假类型", "type": "select", "required": true, "options": [
					{"value": "sick", "label": "病假"},
					{"value": "personal", "label": "事假"},
					{"value": "annual", "label": "年假"}
				]},
				{"name": "start_date", "label": "开始日期", "type": "date", "required": true},
				{"name": "end_date", "label": "结束日期", "type": "date", "required": true},
				{"name": "reason", "label": "请假原因", "type": "textarea", "required": true}
			],
			"flow": [
				{"level": 1, "role": "dept_admin", "label": "部门管理员审批"}
			]
		}')
		ON DUPLICATE KEY UPDATE updated_at = CURRENT_TIMESTAMP`)
	if err != nil {
		return fmt.Errorf("插入默认申请类型失败: %v", err)
	}

	// 插入系统配置
	_, err = database.Exec(`INSERT INTO system_config (config_key, config_value, description) VALUES
		('system_initialized', 'true', '系统是否已初始化'),
		('max_duty_per_week', '2', '每人每周最大值班次数'),
		('max_duty_per_day', '1', '每人每天最大值班次数')
		ON DUPLICATE KEY UPDATE config_value = VALUES(config_value)`)
	if err != nil {
		return fmt.Errorf("插入系统配置失败: %v", err)
	}

	// 插入默认导出模板 - 课表格式
	_, err = database.Exec(`INSERT IGNORE INTO export_templates (id, admin_id, name, description, config, is_default) VALUES
		(1, 1, '课表格式', '矩阵式课表格式（类似课程表）', '{"title": "{department}第{week}周值班表", "mode": "schedule", "scheduleConfig": {"rowHeader": "节次", "colHeader": "星期", "rowLabels": ["第1节", "第2节", "第3节", "第4节"], "colLabels": ["周一", "周二", "周三", "周四", "周五"], "cellFormat": "{users}", "emptyCellText": "-"}, "placeholders": {"week": "周次数字", "department": "部门名称", "users": "值班人员姓名列表", "count": "值班人数"}}', 1)`)
	if err != nil {
		return fmt.Errorf("插入默认模板失败: %v", err)
	}

	// 插入默认导出模板 - 列表格式
	_, err = database.Exec(`INSERT IGNORE INTO export_templates (id, admin_id, name, description, config, is_default) VALUES
		(2, 1, '列表格式', '行列表格式排班表', '{"title": "{department}第{week}周排班表", "mode": "list", "headers": ["星期", "节次", "值班人员"], "dataColumns": [{"type": "weekday", "format": "周{weekday_cn}"}, {"type": "period", "format": "第{period}节"}, {"type": "users", "format": "{users}", "separator": "、"}], "placeholders": {"week": "周次数字", "department": "部门名称", "weekday": "星期数字(1-5)", "weekday_cn": "星期中文(一、二、三...)", "period": "节次数字(1-4)", "users": "值班人员姓名列表"}}', 0)`)
	if err != nil {
		return fmt.Errorf("插入列表模板失败: %v", err)
	}

	return nil
}

// CreateAdmin 创建管理员账号
func (s *SystemService) CreateAdmin(studentID, name, email, password, department string) (*model.User, error) {
	// 检查数据库连接
	database := db.GetDB()
	if database == nil {
		return nil, fmt.Errorf("数据库连接未初始化，请先完成数据库初始化步骤")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %v", err)
	}

	// 插入用户（默认为部门管理员）
	result, err := database.Exec(
		"INSERT INTO users (student_id, name, email, password, role, department, dept_role, is_active) VALUES (?, ?, ?, ?, 'admin', ?, 'dept_admin', 1)",
		studentID, name, email, string(hashedPassword), department)
	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %v", err)
	}

	userID, _ := result.LastInsertId()

	// 插入默认模板
	if err := s.insertDefaultTemplates(int(userID)); err != nil {
		// 模板插入失败记录日志但不影响主流程
		log.Printf("警告: 插入默认模板失败: %v", err)
	}

	return &model.User{
		ID:         int(userID),
		StudentID:  studentID,
		Name:       name,
		Email:      email,
		Role:       "admin",
		Department: department,
		DeptRole:   "dept_admin",
		IsActive:   true,
	}, nil
}

// insertDefaultTemplates 插入默认模板
func (s *SystemService) insertDefaultTemplates(adminID int) error {
	templates := []struct {
		name        string
		description string
		isDefault   int
		config      string
	}{
		{
			name:        "课表格式",
			description: "矩阵式课表格式（类似课程表）",
			isDefault:   1,
			config: `{
				"title": "{department}第{week}周值班表",
				"mode": "schedule",
				"scheduleConfig": {
					"rowHeader": "节次",
					"colHeader": "星期",
					"rowLabels": ["第1节", "第2节", "第3节", "第4节"],
					"colLabels": ["周一", "周二", "周三", "周四", "周五"],
					"cellFormat": "{users}",
					"emptyCellText": "-"
				},
				"placeholders": {
					"week": "周次数字",
					"department": "部门名称",
					"users": "值班人员姓名列表",
					"count": "值班人数"
				}
			}`,
		},
		{
			name:        "列表格式",
			description: "行列表格式排班表",
			isDefault:   0,
			config: `{
				"title": "{department}第{week}周排班表",
				"mode": "list",
				"headers": ["星期", "节次", "值班人员"],
				"dataColumns": [
					{"type": "weekday", "format": "周{weekday_cn}"},
					{"type": "period", "format": "第{period}节"},
					{"type": "users", "format": "{users}", "separator": "。"}
				],
				"placeholders": {
					"week": "周次数字",
					"department": "部门名称",
					"weekday": "星期数字(1-5)",
					"weekday_cn": "星期中文(一、二、三...)",
					"period": "节次数字(1-4)",
					"users": "值班人员姓名列表"
				}
			}`,
		},
	}

	for _, t := range templates {
		_, err := db.GetDB().Exec(
			"INSERT INTO export_templates (admin_id, name, description, config, is_default) VALUES (?, ?, ?, ?, ?)",
			adminID, t.name, t.description, t.config, t.isDefault)
		if err != nil {
			return fmt.Errorf("插入模板 '%s' 失败: %v", t.name, err)
		}
	}
	return nil
}
