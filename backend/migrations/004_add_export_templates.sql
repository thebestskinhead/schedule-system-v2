-- 导出模板表
CREATE TABLE IF NOT EXISTS export_templates (
    id INT PRIMARY KEY AUTO_INCREMENT,
    admin_id INT NOT NULL COMMENT '创建者ID',
    name VARCHAR(100) NOT NULL COMMENT '模板名称',
    description VARCHAR(255) DEFAULT '' COMMENT '模板描述',
    -- 模板配置JSON，定义行列结构
    config JSON NOT NULL COMMENT '模板配置：表头、行列占位符等',
    is_default TINYINT(1) DEFAULT 0 COMMENT '是否为默认模板',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_admin_id (admin_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Excel导出模板表';

-- 插入默认模板 - 列表格式
INSERT INTO export_templates (admin_id, name, description, config, is_default) VALUES
(1, '列表格式', '行列表格式排班表', '{
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

-- 插入课表格式模板
INSERT INTO export_templates (admin_id, name, description, config, is_default) VALUES
(1, '课表格式', '矩阵式课表格式（类似课程表）', '{
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
