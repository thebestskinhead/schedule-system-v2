-- ============================================================
-- 排班系统 V2.1 数据库初始化脚本（完整全新安装）
-- 权威来源：sched_migrate（迁移后目标结构），与迁移后库结构差异为 0
-- 执行：mysql -h<host> -u<user> -p < init_db.sql
-- 末尾附 9 存储过程 + 1 事件（照抄 013_add_db_logic.sql）
-- ============================================================

CREATE DATABASE IF NOT EXISTS schedule_system_v2
DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE schedule_system_v2;

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- --------------------------------------------------
-- 表：system_config
-- --------------------------------------------------

CREATE TABLE `system_config` (
  `id` int NOT NULL AUTO_INCREMENT,
  `config_key` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL,
  `config_value` text COLLATE utf8mb4_unicode_ci,
  `description` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `config_key` (`config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';


-- --------------------------------------------------
-- 表：users
-- --------------------------------------------------

CREATE TABLE `users` (
  `id` int NOT NULL AUTO_INCREMENT,
  `student_id` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '学号/工号',
  `name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '姓名',
  `email` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL,
  `password` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '加密密码',
  `role` enum('admin','user') COLLATE utf8mb4_unicode_ci DEFAULT 'user' COMMENT '系统角色',
  `department` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT '科普部' COMMENT '所属部门',
  `dept_role` enum('dept_admin','dept_member') COLLATE utf8mb4_unicode_ci DEFAULT 'dept_member' COMMENT '部门角色',
  `employment_status` enum('pending','active','resigned') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'active' COMMENT '任职状态: 待入职/在职/离职(取代is_active)',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `student_id` (`student_id`),
  UNIQUE KEY `email` (`email`),
  KEY `idx_role` (`role`),
  KEY `idx_department` (`department`),
  KEY `idx_dept_role` (`dept_role`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';


-- --------------------------------------------------
-- 表：application_types
-- --------------------------------------------------

CREATE TABLE `application_types` (
  `id` int NOT NULL AUTO_INCREMENT,
  `code` varchar(50) NOT NULL COMMENT '类型代码',
  `name` varchar(100) NOT NULL COMMENT '类型名称',
  `description` text COMMENT '类型描述',
  `config` json DEFAULT NULL COMMENT '类型配置JSON（字段定义、审批流程等）',
  `is_active` tinyint(1) DEFAULT '1' COMMENT '是否启用',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='申请类型表';


-- --------------------------------------------------
-- 表：applications
-- --------------------------------------------------

CREATE TABLE `applications` (
  `id` int NOT NULL AUTO_INCREMENT,
  `application_no` varchar(50) NOT NULL COMMENT '申请单号',
  `type_code` varchar(50) NOT NULL COMMENT '申请类型代码',
  `applicant_id` int NOT NULL COMMENT '申请人ID',
  `department` varchar(50) DEFAULT NULL COMMENT '申请人部门',
  `title` varchar(255) NOT NULL COMMENT '申请标题',
  `content` text COMMENT '申请内容描述',
  `data` json DEFAULT NULL COMMENT '申请数据JSON（不同类型的具体数据）',
  `status` tinyint NOT NULL DEFAULT '0' COMMENT '状态: 0待审批 1审批中 2已通过 3已拒绝 4已撤回',
  `current_level` int DEFAULT '1' COMMENT '当前审批层级',
  `total_levels` int DEFAULT '1' COMMENT '总审批层级',
  `approver_id` int DEFAULT NULL COMMENT '当前审批人ID',
  `result` text COMMENT '审批结果/备注',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `application_no` (`application_no`),
  KEY `idx_applicant` (`applicant_id`),
  KEY `idx_type` (`type_code`),
  KEY `idx_status` (`status`),
  KEY `idx_department` (`department`),
  KEY `idx_status_type` (`status`,`type_code`),
  KEY `idx_applicant_status` (`applicant_id`,`status`),
  KEY `fk_app_approver` (`approver_id`),
  CONSTRAINT `fk_app_applicant` FOREIGN KEY (`applicant_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_app_approver` FOREIGN KEY (`approver_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_app_type` FOREIGN KEY (`type_code`) REFERENCES `application_types` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='申请表';


-- --------------------------------------------------
-- 表：application_approvals
-- --------------------------------------------------

CREATE TABLE `application_approvals` (
  `id` int NOT NULL AUTO_INCREMENT,
  `application_id` int NOT NULL COMMENT '申请ID',
  `level` int NOT NULL COMMENT '审批层级',
  `approver_id` int NOT NULL COMMENT '审批人ID',
  `action` tinyint NOT NULL COMMENT '操作: 1通过 2拒绝 3转办 4评论',
  `comment` text COMMENT '审批意见',
  `data` json DEFAULT NULL COMMENT '附加数据',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_application` (`application_id`),
  KEY `idx_approver` (`approver_id`),
  CONSTRAINT `fk_aprv_application` FOREIGN KEY (`application_id`) REFERENCES `applications` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_aprv_approver` FOREIGN KEY (`approver_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='审批历史表';


-- --------------------------------------------------
-- 表：application_approvers
-- --------------------------------------------------

CREATE TABLE `application_approvers` (
  `id` int NOT NULL AUTO_INCREMENT,
  `type_code` varchar(50) NOT NULL COMMENT '申请类型',
  `level` int NOT NULL COMMENT '层级',
  `role_type` enum('admin','dept_admin','office_admin','specific') NOT NULL COMMENT '审批角色类型',
  `specific_user_id` int DEFAULT NULL COMMENT '具体用户ID（role_type=specific时使用）',
  `is_active` tinyint(1) DEFAULT '1',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_type_level` (`type_code`,`level`),
  KEY `idx_type` (`type_code`),
  KEY `fk_aprver_specific` (`specific_user_id`),
  CONSTRAINT `fk_aprver_specific` FOREIGN KEY (`specific_user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='审批人配置表';


-- --------------------------------------------------
-- 表：availability
-- --------------------------------------------------

CREATE TABLE `availability` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` int NOT NULL COMMENT '用户ID',
  `week` tinyint NOT NULL COMMENT '第几周 1-30',
  `weekday` tinyint NOT NULL COMMENT '星期几 1-5',
  `period` tinyint NOT NULL COMMENT '节次 1-4',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_avail` (`user_id`,`week`,`weekday`,`period`),
  KEY `idx_time_user` (`week`,`weekday`,`period`,`user_id`),
  CONSTRAINT `availability_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `chk_avail_period` CHECK ((`period` between 1 and 4)),
  CONSTRAINT `chk_avail_week` CHECK ((`week` between 1 and 30)),
  CONSTRAINT `chk_avail_weekday` CHECK ((`weekday` between 1 and 5))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='无课时间表';


-- --------------------------------------------------
-- 表：duty_records
-- --------------------------------------------------

CREATE TABLE `duty_records` (
  `id` int NOT NULL AUTO_INCREMENT,
  `week` tinyint NOT NULL COMMENT '第几周 1-30',
  `weekday` tinyint NOT NULL COMMENT '星期几 1-5',
  `period` tinyint NOT NULL COMMENT '节次 1-4',
  `user_id` int NOT NULL COMMENT '值班人员ID',
  `department` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '部门',
  `assigned_by` int DEFAULT NULL COMMENT '排班人ID',
  `status` enum('pending','confirmed','completed','cancelled') COLLATE utf8mb4_unicode_ci DEFAULT 'pending',
  `remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_duty_cell` (`week`,`weekday`,`period`,`user_id`),
  KEY `idx_week` (`week`),
  KEY `idx_user` (`user_id`),
  KEY `idx_department` (`department`),
  KEY `idx_time` (`week`,`weekday`,`period`),
  KEY `assigned_by` (`assigned_by`),
  KEY `idx_user_week` (`user_id`,`week`),
  KEY `idx_user_status` (`user_id`,`status`),
  CONSTRAINT `duty_records_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `duty_records_ibfk_2` FOREIGN KEY (`assigned_by`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `chk_duty_period` CHECK ((`period` between 1 and 4)),
  CONSTRAINT `chk_duty_week` CHECK ((`week` between 1 and 30)),
  CONSTRAINT `chk_duty_weekday` CHECK ((`weekday` between 1 and 5))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='值班记录表';


-- --------------------------------------------------
-- 表：schedule_settings
-- --------------------------------------------------

CREATE TABLE `schedule_settings` (
  `id` int NOT NULL AUTO_INCREMENT,
  `config_key` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'global' COMMENT '单例键',
  `admin_id` int DEFAULT '1' COMMENT '最后修改者ID',
  `current_week` int DEFAULT '1' COMMENT '当前周次',
  `auto_increment` tinyint(1) DEFAULT '1' COMMENT '周次始终自动递增',
  `current_week_set_at` date DEFAULT NULL COMMENT '手动周次的起算日期',
  `need_per_cell` int DEFAULT '2',
  `min_per_cell` int DEFAULT '0',
  `max_per_day` int DEFAULT '1',
  `max_per_week` int DEFAULT '2',
  `export_title` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '第{week}周排班表',
  `semester_start_date` date DEFAULT NULL COMMENT '学期起始日',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_config_key` (`config_key`),
  KEY `idx_admin_id` (`admin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='排班设置表';


-- --------------------------------------------------
-- 表：export_templates
-- --------------------------------------------------

CREATE TABLE `export_templates` (
  `id` int NOT NULL AUTO_INCREMENT,
  `admin_id` int NOT NULL COMMENT '创建者ID',
  `name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '模板名称',
  `description` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '模板描述',
  `config` json NOT NULL COMMENT '模板配置：表头、行列占位符等',
  `is_default` tinyint(1) DEFAULT '0' COMMENT '是否为默认模板',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_admin_id` (`admin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Excel导出模板表';


-- --------------------------------------------------
-- 表：weekly_duty_assignments
-- --------------------------------------------------

CREATE TABLE `weekly_duty_assignments` (
  `id` int NOT NULL AUTO_INCREMENT,
  `week` int NOT NULL COMMENT '周次',
  `department` varchar(50) NOT NULL COMMENT '部门',
  `weekday` int NOT NULL COMMENT '星期几(1-5)',
  `is_assigned` tinyint(1) DEFAULT '1' COMMENT '是否安排值班',
  `created_by` int NOT NULL COMMENT '创建人ID',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_week_dept_weekday` (`week`,`department`,`weekday`),
  KEY `idx_week` (`week`),
  KEY `idx_department` (`department`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='每周值班分工表';


-- --------------------------------------------------
-- 表：user_permissions_temp
-- --------------------------------------------------

CREATE TABLE `user_permissions_temp` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` int NOT NULL COMMENT '被授权用户ID',
  `permission` varchar(50) NOT NULL COMMENT '权限代码',
  `resource_type` varchar(20) DEFAULT 'all' COMMENT '资源类型(all/dept/user)',
  `resource_id` int DEFAULT '0' COMMENT '资源ID(部门ID或用户ID)',
  `granted_by` int NOT NULL COMMENT '授权人ID',
  `reason` varchar(255) NOT NULL DEFAULT '' COMMENT '授权原因',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `expires_at` timestamp NULL DEFAULT NULL,
  `is_active` tinyint(1) DEFAULT '1' COMMENT '是否有效',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_expires_at` (`expires_at`),
  KEY `idx_is_active` (`is_active`),
  KEY `idx_user_perm_active` (`user_id`,`permission`,`is_active`),
  KEY `fk_tp_granted_by` (`granted_by`),
  CONSTRAINT `fk_tp_granted_by` FOREIGN KEY (`granted_by`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_tp_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户临时权限表';


-- --------------------------------------------------
-- 表：smtp_config
-- --------------------------------------------------

CREATE TABLE `smtp_config` (
  `id` int NOT NULL AUTO_INCREMENT,
  `host` varchar(255) NOT NULL COMMENT 'SMTP服务器地址',
  `port` int NOT NULL COMMENT '端口',
  `username` varchar(255) NOT NULL COMMENT '用户名/邮箱',
  `password` varchar(255) NOT NULL COMMENT '密码/授权码',
  `from_name` varchar(100) NOT NULL COMMENT '发件人显示名称',
  `from_email` varchar(255) NOT NULL COMMENT '发件人邮箱',
  `use_tls` tinyint(1) DEFAULT '1' COMMENT '是否使用TLS',
  `use_ssl` tinyint(1) DEFAULT '0' COMMENT '是否使用SSL',
  `is_active` tinyint(1) DEFAULT '1' COMMENT '是否启用',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='SMTP配置表';


-- --------------------------------------------------
-- 表：password_reset_tokens
-- --------------------------------------------------

CREATE TABLE `password_reset_tokens` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` int NOT NULL COMMENT '用户ID',
  `email` varchar(255) NOT NULL COMMENT '邮箱',
  `token` varchar(255) NOT NULL COMMENT '重置令牌',
  `expires_at` timestamp NULL DEFAULT NULL COMMENT '过期时间',
  `is_used` tinyint(1) DEFAULT '0' COMMENT '是否已使用',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_token` (`token`),
  KEY `idx_expires_at` (`expires_at`),
  KEY `fk_prt_user` (`user_id`),
  CONSTRAINT `fk_prt_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='密码重置令牌表';


-- --------------------------------------------------
-- 表：leave_requests
-- --------------------------------------------------

CREATE TABLE `leave_requests` (
  `id` int NOT NULL AUTO_INCREMENT,
  `application_id` int NOT NULL COMMENT '关联审批单',
  `user_id` int NOT NULL COMMENT '请假人',
  `leave_type` enum('sick','personal','annual') COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '请假类型',
  `start_date` date NOT NULL COMMENT '开始日期',
  `end_date` date NOT NULL COMMENT '结束日期',
  `status` enum('pending','approved','rejected') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT '状态',
  `reason` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '请假原因',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_application` (`application_id`),
  KEY `idx_user_date` (`user_id`,`start_date`,`end_date`),
  CONSTRAINT `fk_leave_application` FOREIGN KEY (`application_id`) REFERENCES `applications` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_leave_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `chk_leave_date` CHECK ((`end_date` >= `start_date`))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='请假表';


-- --------------------------------------------------
-- 表：user_invitations
-- --------------------------------------------------

CREATE TABLE `user_invitations` (
  `id` int NOT NULL AUTO_INCREMENT,
  `student_id` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '学号',
  `name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '姓名',
  `email` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '邮箱',
  `department` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '科普部' COMMENT '部门',
  `dept_role` enum('dept_admin','dept_member') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'dept_member' COMMENT '部门角色',
  `token` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '激活令牌(SHA-256哈希)',
  `token_expires_at` timestamp NOT NULL COMMENT '过期时间',
  `status` enum('pending','activated','expired','revoked') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pending' COMMENT '状态',
  `created_by` int NOT NULL COMMENT '录入人',
  `activated_at` timestamp NULL DEFAULT NULL COMMENT '激活时间',
  `activated_user_id` int DEFAULT NULL COMMENT '激活后关联的正式用户',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_token` (`token`),
  UNIQUE KEY `uk_student_id` (`student_id`),
  KEY `idx_status` (`status`,`token_expires_at`),
  KEY `fk_inv_created_by` (`created_by`),
  KEY `fk_inv_activated` (`activated_user_id`),
  CONSTRAINT `fk_inv_activated` FOREIGN KEY (`activated_user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_inv_created_by` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='预注册待激活账号表';



SET FOREIGN_KEY_CHECKS = 1;

-- ============================================================
-- 种子数据（最小可用数据，确保外键/单例约束有基础行）
-- ============================================================

-- 系统配置
INSERT INTO system_config (config_key, config_value, description) VALUES
('system_initialized', 'false', '系统是否已初始化'),
('max_duty_per_week', '2', '每人每周最大值班次数'),
('max_duty_per_day', '1', '每人每天最大值班次数')
ON DUPLICATE KEY UPDATE config_value = VALUES(config_value);

-- 默认导出模板 - 课表格式
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

-- 默认导出模板 - 列表格式
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

-- 审批类型（申请流基础数据）
INSERT INTO application_types (code, name, description) VALUES
('temp_permission', '临时权限申请', '申请临时管理权限'),
('leave', '请假申请', '申请休假')
ON DUPLICATE KEY UPDATE name = VALUES(name), description = VALUES(description);

-- 排班设置单例行（config_key='global'）
INSERT IGNORE INTO schedule_settings (config_key, admin_id, current_week, auto_increment, need_per_cell, min_per_cell, max_per_day, max_per_week, export_title, semester_start_date)
VALUES ('global', 1, 1, 1, 2, 0, 1, 2, '第{week}周排班表', NULL);


-- ============================================================

-- 存储过程 + 事件（照抄 013_add_db_logic.sql，保留 DELIMITER；触发器已移除）

-- ============================================================


DELIMITER //

-- ===================================================================
-- §7.1 排班确认（原子替换整周 + 全局互斥锁 + 已确认守卫）
-- ===================================================================

DROP PROCEDURE IF EXISTS sp_confirm_schedule //
CREATE PROCEDURE sp_confirm_schedule(IN p_week INT, IN p_admin_id INT, IN p_cells JSON)
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION BEGIN ROLLBACK; RESIGNAL; END;
    START TRANSACTION;
    -- ① 全局排班互斥锁：锁 schedule_settings 单例行，串行化全部排班写入
    SELECT config_key FROM schedule_settings WHERE config_key = 'global' FOR UPDATE;
    -- ② 守卫：该周存在已确认/已完成记录则拒绝（行锁保证并发可见）
    IF EXISTS(SELECT 1 FROM duty_records
              WHERE week = p_week AND status IN ('confirmed','completed')) THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '该周已有已确认或完成的排班，禁止重排';
    END IF;
    DELETE FROM duty_records WHERE week = p_week;
    INSERT INTO duty_records (week, weekday, period, user_id, department, assigned_by, status)
    SELECT DISTINCT p_week, jt.weekday, jt.period, ju.user_id, u.department, p_admin_id, 'pending'
    FROM JSON_TABLE(p_cells, '$[*]' COLUMNS (
        weekday INT PATH '$.weekday', period INT PATH '$.period',
        user_ids JSON PATH '$.user_ids')) jt
    CROSS JOIN JSON_TABLE(jt.user_ids, '$[*]' COLUMNS (user_id INT PATH '$')) ju
    JOIN users u ON u.id = ju.user_id;
    COMMIT;
END //

-- ===================================================================
-- §7.2 邀请激活（事务 + 行锁防并发；修正原文档 subquery 中 inv 别名作用域问题）
-- ===================================================================

DROP PROCEDURE IF EXISTS sp_activate_invitation //
CREATE PROCEDURE sp_activate_invitation(IN p_token_hash VARCHAR(64), IN p_password_hash VARCHAR(255))
BEGIN
    DECLARE v_student_id VARCHAR(50);
    DECLARE EXIT HANDLER FOR SQLEXCEPTION BEGIN ROLLBACK; RESIGNAL; END;
    START TRANSACTION;
    SELECT id INTO @inv_id FROM user_invitations
        WHERE token = p_token_hash AND status = 'pending' FOR UPDATE;
    IF @inv_id IS NULL THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '令牌无效或已使用';
    END IF;
    IF (SELECT token_expires_at FROM user_invitations WHERE id = @inv_id) < NOW() THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '令牌已过期';
    END IF;
    SELECT student_id INTO v_student_id FROM user_invitations WHERE id = @inv_id;
    UPDATE users u
    JOIN user_invitations inv ON inv.id = @inv_id AND u.student_id = inv.student_id
       SET u.employment_status = 'active',
           u.password = p_password_hash,
           u.email = COALESCE(inv.email, u.email),
           u.updated_at = NOW()
     WHERE u.employment_status = 'pending';
    IF ROW_COUNT() = 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '用户记录不存在或状态异常';
    END IF;
    UPDATE user_invitations
        SET status = 'activated', activated_at = NOW(),
            activated_user_id = (SELECT id FROM users WHERE student_id = v_student_id)
        WHERE id = @inv_id;
    COMMIT;
END //

-- ===================================================================
-- §7.3 离职处理（单用户 + 批量换届）
-- ===================================================================

DROP PROCEDURE IF EXISTS sp_mark_user_resigned //
CREATE PROCEDURE sp_mark_user_resigned(IN p_user_id INT, IN p_admin_id INT)
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION BEGIN ROLLBACK; RESIGNAL; END;
    START TRANSACTION;
    UPDATE users SET employment_status = 'resigned' WHERE id = p_user_id;
    UPDATE duty_records SET status = 'cancelled'
        WHERE user_id = p_user_id AND status = 'pending';
    UPDATE user_permissions_temp SET is_active = 0
        WHERE user_id = p_user_id AND is_active = 1;
    COMMIT;
END //

DROP PROCEDURE IF EXISTS sp_mark_users_resigned //
CREATE PROCEDURE sp_mark_users_resigned(IN p_user_ids JSON, IN p_admin_id INT)
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION BEGIN ROLLBACK; RESIGNAL; END;
    START TRANSACTION;
    UPDATE users u
    JOIN JSON_TABLE(p_user_ids, '$[*]' COLUMNS (user_id INT PATH '$')) jt ON u.id = jt.user_id
       SET u.employment_status = 'resigned', u.updated_at = NOW()
     WHERE u.employment_status = 'active';
    UPDATE duty_records dr
    JOIN JSON_TABLE(p_user_ids, '$[*]' COLUMNS (user_id INT PATH '$')) jt ON dr.user_id = jt.user_id
       SET dr.status = 'cancelled', dr.updated_at = NOW()
     WHERE dr.status = 'pending';
    UPDATE user_permissions_temp pt
    JOIN JSON_TABLE(p_user_ids, '$[*]' COLUMNS (user_id INT PATH '$')) jt ON pt.user_id = jt.user_id
       SET pt.is_active = 0
     WHERE pt.is_active = 1;
    COMMIT;
END //

-- ===================================================================
-- §7.4 请假批准联动
-- ===================================================================

DROP PROCEDURE IF EXISTS sp_approve_leave //
CREATE PROCEDURE sp_approve_leave(IN p_application_id INT, IN p_week_start DATE)
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION BEGIN ROLLBACK; RESIGNAL; END;
    START TRANSACTION;
    UPDATE leave_requests SET status = 'approved' WHERE application_id = p_application_id;
    UPDATE duty_records dr
    JOIN leave_requests lr ON lr.application_id = p_application_id AND lr.user_id = dr.user_id
    SET dr.status = 'cancelled', dr.updated_at = NOW()
    WHERE dr.status IN ('pending', 'confirmed')
      AND DATE_ADD(p_week_start, INTERVAL (dr.week - 1) * 7 + (dr.weekday - 1) DAY)
          BETWEEN lr.start_date AND lr.end_date;
    COMMIT;
END //

-- ===================================================================
-- §12.2.1 单格原子替换（防并发覆盖）
-- ===================================================================

DROP PROCEDURE IF EXISTS sp_update_schedule_cell //
CREATE PROCEDURE sp_update_schedule_cell(
    IN p_week INT, IN p_weekday INT, IN p_period INT,
    IN p_admin_id INT, IN p_add_user_ids JSON, IN p_remove_user_ids JSON)
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION BEGIN ROLLBACK; RESIGNAL; END;
    START TRANSACTION;
    SELECT config_key FROM schedule_settings WHERE config_key = 'global' FOR UPDATE;
    IF EXISTS(SELECT 1 FROM duty_records
              WHERE week = p_week AND weekday = p_weekday AND period = p_period
                AND status IN ('confirmed','completed')) THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '该时段已确认/完成，禁止修改';
    END IF;
    DELETE FROM duty_records
        WHERE week = p_week AND weekday = p_weekday AND period = p_period
          AND user_id IN (SELECT CAST(jt.user_id AS UNSIGNED)
                          FROM JSON_TABLE(p_remove_user_ids, '$[*]' COLUMNS (user_id INT PATH '$')) jt);
    INSERT IGNORE INTO duty_records (week, weekday, period, user_id, department, assigned_by, status)
    SELECT p_week, p_weekday, p_period, ju.user_id, u.department, p_admin_id, 'pending'
    FROM JSON_TABLE(p_add_user_ids, '$[*]' COLUMNS (user_id INT PATH '$')) ju
    JOIN users u ON u.id = ju.user_id
    WHERE u.employment_status = 'active';
    COMMIT;
END //

-- ===================================================================
-- §12.2.2 整周分工原子替换
-- ===================================================================

DROP PROCEDURE IF EXISTS sp_publish_weekly_assignments //
CREATE PROCEDURE sp_publish_weekly_assignments(IN p_week INT, IN p_admin_id INT, IN p_items JSON)
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION BEGIN ROLLBACK; RESIGNAL; END;
    START TRANSACTION;
    DELETE FROM weekly_duty_assignments WHERE week = p_week;
    INSERT INTO weekly_duty_assignments (week, department, weekday, is_assigned, created_by)
    SELECT p_week, jt.department, jt.weekday, IF(jt.is_assigned = 1, TRUE, FALSE), p_admin_id
    FROM JSON_TABLE(p_items, '$[*]' COLUMNS (
        department VARCHAR(50) PATH '$.department',
        weekday INT PATH '$.weekday',
        is_assigned INT PATH '$.is_assigned')) jt;
    COMMIT;
END //

-- ===================================================================
-- §12.2.3 无课表全量原子替换（Excel/手动导入）
-- ===================================================================

DROP PROCEDURE IF EXISTS sp_replace_availability //
CREATE PROCEDURE sp_replace_availability(IN p_user_id INT, IN p_items JSON)
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION BEGIN ROLLBACK; RESIGNAL; END;
    START TRANSACTION;
    DELETE FROM availability WHERE user_id = p_user_id;
    INSERT IGNORE INTO availability (user_id, week, weekday, period)
    SELECT p_user_id, jt.week, jt.weekday, jt.period
    FROM JSON_TABLE(p_items, '$[*]' COLUMNS (
        week INT PATH '$.week', weekday INT PATH '$.weekday', period INT PATH '$.period')) jt;
    COMMIT;
END //

-- ===================================================================
-- §12.2.4 审批记录 + 状态推进原子化（乐观锁）
-- ===================================================================

DROP PROCEDURE IF EXISTS sp_advance_application //
CREATE PROCEDURE sp_advance_application(
    IN p_app_id INT, IN p_approver_id INT, IN p_action TINYINT,
    IN p_comment TEXT, IN p_expected_status TINYINT, IN p_new_status TINYINT)
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION BEGIN ROLLBACK; RESIGNAL; END;
    START TRANSACTION;
    UPDATE applications
       SET status = p_new_status, updated_at = NOW()
     WHERE id = p_app_id AND status = p_expected_status;
    IF ROW_COUNT() = 0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = '申请状态已变更，请刷新后重试';
    END IF;
    INSERT INTO application_approvals (application_id, approver_id, action, comment, level, created_at)
    SELECT p_app_id, p_approver_id, p_action, p_comment,
           current_level, NOW() FROM applications WHERE id = p_app_id;
    COMMIT;
END //

-- ===================================================================
-- §12.4.6 事件调度：周期清理（分批 LIMIT 1000 循环，避免长事务大范围锁）
-- ===================================================================

DROP EVENT IF EXISTS evt_cleanup_expired //
CREATE EVENT evt_cleanup_expired
ON SCHEDULE EVERY 1 HOUR
DO
BEGIN
    DECLARE v_done INT DEFAULT 0;
    -- 分批失效过期临时权限
    SET v_done = 1;
    WHILE v_done > 0 DO
        UPDATE user_permissions_temp SET is_active = 0 WHERE expires_at < NOW() AND is_active = 1 LIMIT 1000;
        SET v_done = ROW_COUNT();
    END WHILE;
    -- 分批标记过期邀请
    SET v_done = 1;
    WHILE v_done > 0 DO
        UPDATE user_invitations SET status = 'expired' WHERE token_expires_at < NOW() AND status = 'pending' LIMIT 1000;
        SET v_done = ROW_COUNT();
    END WHILE;
    -- 分批删除过期/已用令牌
    SET v_done = 1;
    WHILE v_done > 0 DO
        DELETE FROM password_reset_tokens WHERE (expires_at < NOW() OR is_used = 1) LIMIT 1000;
        SET v_done = ROW_COUNT();
    END WHILE;
END //

DELIMITER ;
