-- 012_add_hr_tables.sql
-- 新增 HR 相关表（依据 docs/数据库.md v2.1 §5.9 / §5.10）

-- 请假表（复用审批流）
CREATE TABLE IF NOT EXISTS leave_requests (
    id INT PRIMARY KEY AUTO_INCREMENT,
    application_id INT NOT NULL COMMENT '关联审批单',
    user_id INT NOT NULL COMMENT '请假人',
    leave_type ENUM('sick','personal','annual') NOT NULL COMMENT '请假类型',
    start_date DATE NOT NULL COMMENT '开始日期',
    end_date DATE NOT NULL COMMENT '结束日期',
    status ENUM('pending','approved','rejected') NOT NULL DEFAULT 'pending' COMMENT '状态',
    reason VARCHAR(255) COMMENT '请假原因',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_application (application_id),
    KEY idx_user_date (user_id, start_date, end_date),
    CONSTRAINT fk_leave_application FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE,
    CONSTRAINT fk_leave_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT chk_leave_date CHECK (end_date >= start_date)
) ENGINE=InnoDB COMMENT='请假表';

-- 预注册待激活账号表
CREATE TABLE IF NOT EXISTS user_invitations (
    id INT PRIMARY KEY AUTO_INCREMENT,
    student_id VARCHAR(20) NOT NULL COMMENT '学号',
    name VARCHAR(50) NOT NULL COMMENT '姓名',
    email VARCHAR(100) NOT NULL COMMENT '邮箱',
    department VARCHAR(50) NOT NULL DEFAULT '科普部' COMMENT '部门',
    dept_role ENUM('dept_admin','dept_member') NOT NULL DEFAULT 'dept_member' COMMENT '部门角色',
    token VARCHAR(64) NOT NULL COMMENT '激活令牌(SHA-256哈希)',
    token_expires_at TIMESTAMP NOT NULL COMMENT '过期时间',
    status ENUM('pending','activated','expired','revoked') NOT NULL DEFAULT 'pending' COMMENT '状态',
    created_by INT NOT NULL COMMENT '录入人',
    activated_at TIMESTAMP NULL COMMENT '激活时间',
    activated_user_id INT NULL COMMENT '激活后关联的正式用户',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_token (token),
    UNIQUE KEY uk_student_id (student_id),
    KEY idx_status (status, token_expires_at),
    CONSTRAINT fk_inv_created_by FOREIGN KEY (created_by) REFERENCES users(id),
    CONSTRAINT fk_inv_activated FOREIGN KEY (activated_user_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB COMMENT='预注册待激活账号表';
