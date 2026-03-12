-- 排班系统 V2 数据库初始化脚本

CREATE DATABASE IF NOT EXISTS schedule_system_v2 
DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE schedule_system_v2;

-- 系统配置表
CREATE TABLE IF NOT EXISTS system_config (
    id INT PRIMARY KEY AUTO_INCREMENT,
    config_key VARCHAR(50) NOT NULL UNIQUE,
    config_value TEXT,
    description VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB COMMENT='系统配置表';

-- 用户表
CREATE TABLE IF NOT EXISTS users (
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
) ENGINE=InnoDB COMMENT='用户表';

-- 无课时间表 (倒排索引设计)
CREATE TABLE IF NOT EXISTS availability (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '用户ID',
    week TINYINT NOT NULL COMMENT '第几周 1-30',
    weekday TINYINT NOT NULL COMMENT '星期几 1-5',
    period TINYINT NOT NULL COMMENT '节次 1-4',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_avail (user_id, week, weekday, period),
    INDEX idx_time_user (week, weekday, period, user_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB COMMENT='无课时间表';

-- 值班记录表
CREATE TABLE IF NOT EXISTS duty_records (
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
) ENGINE=InnoDB COMMENT='值班记录表';

-- 值班次数统计表
CREATE TABLE IF NOT EXISTS duty_counters (
    user_id INT PRIMARY KEY COMMENT '用户ID',
    total_count INT DEFAULT 0 COMMENT '总值班次数',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB COMMENT='值班次数统计';

-- 排班设置表
CREATE TABLE IF NOT EXISTS schedule_settings (
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
) ENGINE=InnoDB COMMENT='排班设置表';

-- 导出模板表
CREATE TABLE IF NOT EXISTS export_templates (
    id INT PRIMARY KEY AUTO_INCREMENT,
    admin_id INT NOT NULL COMMENT '创建者ID',
    name VARCHAR(100) NOT NULL COMMENT '模板名称',
    description VARCHAR(255) DEFAULT '' COMMENT '模板描述',
    config JSON NOT NULL COMMENT '模板配置：表头、行列占位符等',
    is_default TINYINT(1) DEFAULT 0 COMMENT '是否为默认模板',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_admin_id (admin_id)
) ENGINE=InnoDB COMMENT='Excel导出模板表';

-- 插入系统配置
INSERT INTO system_config (config_key, config_value, description) VALUES
('system_initialized', 'false', '系统是否已初始化'),
('max_duty_per_week', '2', '每人每周最大值班次数'),
('max_duty_per_day', '1', '每人每天最大值班次数')
ON DUPLICATE KEY UPDATE config_value = VALUES(config_value);

-- 插入默认导出模板 - 课表格式（默认）
INSERT IGNORE INTO export_templates (id, admin_id, name, description, config, is_default) VALUES
(1, 1, '课表格式', '矩阵式课表格式（类似课程表）', '{
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
}', 1);

-- 插入默认导出模板 - 列表格式
INSERT IGNORE INTO export_templates (id, admin_id, name, description, config, is_default) VALUES
(2, 1, '列表格式', '行列表格式排班表', '{
    "title": "{department}第{week}周排班表",
    "mode": "list",
    "headers": ["星期", "节次", "值班人员"],
    "dataColumns": [
        {"type": "weekday", "format": "周{weekday_cn}"},
        {"type": "period", "format": "第{period}节"},
        {"type": "users", "format": "{users}", "separator": "、"}
    ],
    "placeholders": {
        "week": "周次数字",
        "department": "部门名称",
        "weekday": "星期数字(1-5)",
        "weekday_cn": "星期中文(一、二、三...)",
        "period": "节次数字(1-4)",
        "users": "值班人员姓名列表"
    }
}', 0);

-- 每周值班分工表
CREATE TABLE IF NOT EXISTS weekly_duty_assignments (
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='每周值班分工表';

-- 用户临时权限表
CREATE TABLE IF NOT EXISTS user_permissions_temp (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '被授权用户ID',
    permission VARCHAR(50) NOT NULL COMMENT '权限代码',
    resource_type VARCHAR(20) DEFAULT 'all' COMMENT '资源类型(all/dept/user)',
    resource_id INT DEFAULT 0 COMMENT '资源ID(部门ID或用户ID)',
    granted_by INT NOT NULL COMMENT '授权人ID',
    reason VARCHAR(255) COMMENT '授权原因',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL COMMENT '过期时间',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否有效',
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户临时权限表';

-- SMTP配置表
CREATE TABLE IF NOT EXISTS smtp_config (
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SMTP配置表';

-- 密码重置令牌表
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '用户ID',
    email VARCHAR(255) NOT NULL COMMENT '邮箱',
    token VARCHAR(255) NOT NULL COMMENT '重置令牌',
    expires_at TIMESTAMP NOT NULL COMMENT '过期时间',
    is_used BOOLEAN DEFAULT FALSE COMMENT '是否已使用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_token (token),
    INDEX idx_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='密码重置令牌表';
