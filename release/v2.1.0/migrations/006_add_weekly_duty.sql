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
