-- 添加排班设置表
CREATE TABLE IF NOT EXISTS schedule_settings (
    id INT PRIMARY KEY AUTO_INCREMENT,
    admin_id INT NOT NULL COMMENT '管理员ID',
    need_per_cell INT NOT NULL DEFAULT 2 COMMENT '每时段最大人数',
    min_per_cell INT NOT NULL DEFAULT 0 COMMENT '每时段最小人数',
    max_per_day INT NOT NULL DEFAULT 1 COMMENT '每人每天最多排班次数',
    max_per_week INT NOT NULL DEFAULT 2 COMMENT '每人每周最多排班次数',
    export_title VARCHAR(255) DEFAULT '第{week}周排班表' COMMENT '导出Excel标题模板',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_admin (admin_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='排班设置表';
