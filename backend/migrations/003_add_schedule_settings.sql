-- 添加排班设置表（全局唯一配置）
CREATE TABLE IF NOT EXISTS schedule_settings (
    id INT PRIMARY KEY AUTO_INCREMENT,
    admin_id INT DEFAULT 1 COMMENT '最后修改者ID',
    current_week INT DEFAULT 1 COMMENT '当前周次',
    auto_increment TINYINT(1) DEFAULT 0 COMMENT '是否自动递增周次',
    need_per_cell INT NOT NULL DEFAULT 2 COMMENT '每时段最大人数',
    min_per_cell INT NOT NULL DEFAULT 0 COMMENT '每时段最小人数',
    max_per_day INT NOT NULL DEFAULT 1 COMMENT '每人每天最多排班次数',
    max_per_week INT NOT NULL DEFAULT 2 COMMENT '每人每周最多排班次数',
    export_title VARCHAR(255) DEFAULT '第{week}周排班表' COMMENT '导出Excel标题模板',
    semester_start_date DATE DEFAULT NULL COMMENT '学期起始日',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_admin_id (admin_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='排班设置表';
