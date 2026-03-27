-- 申请类型表
CREATE TABLE IF NOT EXISTS application_types (
    id INT PRIMARY KEY AUTO_INCREMENT,
    code VARCHAR(50) NOT NULL UNIQUE COMMENT '类型代码',
    name VARCHAR(100) NOT NULL COMMENT '类型名称',
    description TEXT COMMENT '类型描述',
    config JSON COMMENT '类型配置JSON（字段定义、审批流程等）',
    is_active TINYINT(1) DEFAULT 1 COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='申请类型表';

-- 申请表
CREATE TABLE IF NOT EXISTS applications (
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='申请表';

-- 审批历史记录表
CREATE TABLE IF NOT EXISTS application_approvals (
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='审批历史表';

-- 审批人配置表
CREATE TABLE IF NOT EXISTS application_approvers (
    id INT PRIMARY KEY AUTO_INCREMENT,
    type_code VARCHAR(50) NOT NULL COMMENT '申请类型',
    level INT NOT NULL COMMENT '层级',
    role_type ENUM('admin', 'dept_admin', 'office_admin', 'specific') NOT NULL COMMENT '审批角色类型',
    specific_user_id INT COMMENT '具体用户ID（role_type=specific时使用）',
    is_active TINYINT(1) DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_type_level (type_code, level),
    INDEX idx_type (type_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='审批人配置表';

-- 插入默认申请类型：权限申请
INSERT INTO application_types (code, name, description, config) VALUES
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
ON DUPLICATE KEY UPDATE updated_at = CURRENT_TIMESTAMP;
