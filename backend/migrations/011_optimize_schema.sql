-- 011_optimize_schema.sql
-- 结构优化迁移（依据 docs/数据库.md v2.1 §5 / §11）
-- 执行顺序：备份 → users 扩展 → duty_records 去重+约束 → schedule_settings 单例化
--          → user_permissions_temp → password_reset_tokens → 审批流三表 → availability
--          → 删除 duty_counters
-- 注意：本迁移假定运行在 init_db.sql + 003~010 之后的现有结构之上。

-- ① 备份 duty_counters（§11 前置备份；IF NOT EXISTS 保证可重复执行不报错）
CREATE TABLE IF NOT EXISTS duty_counters_backup_20261013 AS SELECT * FROM duty_counters;

-- ② users 扩展字段：任职状态（取代 is_active）（§5.1 ①）
ALTER TABLE users
    ADD COLUMN employment_status ENUM('pending','active','resigned') NOT NULL DEFAULT 'active'
        COMMENT '任职状态: 待入职/在职/离职(取代is_active)' AFTER dept_role;

-- ③ 存量回填：is_active 语义迁移到 employment_status（§5.1 ②）
UPDATE users SET employment_status = 'active' WHERE is_active = 1;
UPDATE users SET employment_status = 'resigned' WHERE is_active = 0;

-- ④ 删除账号禁用标志（§5.1 ③）。users.joined_at/left_at 在 init_db.sql 与 system.go 中均不存在，故不处理。
ALTER TABLE users DROP INDEX idx_active;
ALTER TABLE users DROP COLUMN is_active;

-- ⑤ duty_records 存量去重（保留最小 id）+ 唯一约束 + 索引 + CHECK（§5.3）
DELETE d1 FROM duty_records d1
JOIN duty_records d2
  ON d1.week = d2.week AND d1.weekday = d2.weekday
 AND d1.period = d2.period AND d1.user_id = d2.user_id AND d1.id > d2.id;

ALTER TABLE duty_records
    ADD UNIQUE KEY uk_duty_cell (week, weekday, period, user_id),
    ADD KEY idx_user_week (user_id, week),
    ADD KEY idx_user_status (user_id, status),
    ADD CONSTRAINT chk_duty_week CHECK (week BETWEEN 1 AND 30),
    ADD CONSTRAINT chk_duty_weekday CHECK (weekday BETWEEN 1 AND 5),
    ADD CONSTRAINT chk_duty_period CHECK (period BETWEEN 1 AND 4);

-- ⑥ schedule_settings 单例化（§5.2 + §11）：多管理员累积多行 → 仅保留 id 最小一行作为单例行
DELETE FROM schedule_settings
WHERE id <> (SELECT min_id FROM (SELECT MIN(id) AS min_id FROM schedule_settings) t);

ALTER TABLE schedule_settings
    ADD COLUMN config_key VARCHAR(20) NOT NULL DEFAULT 'global' COMMENT '单例键' AFTER id;

UPDATE schedule_settings SET config_key = 'global' WHERE config_key <> 'global';

ALTER TABLE schedule_settings ADD UNIQUE KEY uk_config_key (config_key);

-- 插入单例行（幂等：若已被保留行占据 config_key='global' 则跳过；防空表）
INSERT IGNORE INTO schedule_settings (config_key, admin_id, current_week, auto_increment, need_per_cell, min_per_cell, max_per_day, max_per_week, export_title, semester_start_date)
VALUES ('global', 1, 1, 0, 2, 0, 1, 2, '第{week}周排班表', NULL);

-- ⑦ user_permissions_temp：索引 + 外键 + expires_at 统一为 NULL 默认（§5.5）
ALTER TABLE user_permissions_temp
    ADD KEY idx_user_perm_active (user_id, permission, is_active),
    ADD CONSTRAINT fk_tp_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_tp_granted_by FOREIGN KEY (granted_by) REFERENCES users(id);
-- expires_at 改可空（消除 model 非指针 Scan NULL 报错）
ALTER TABLE user_permissions_temp MODIFY expires_at TIMESTAMP NULL DEFAULT NULL;

-- ⑧ password_reset_tokens：唯一索引（防重放）+ 外键（§5.6）
ALTER TABLE password_reset_tokens
    DROP INDEX idx_token,
    ADD UNIQUE KEY uk_token (token),
    ADD CONSTRAINT fk_prt_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- ⑨ 审批流三表：复合索引 + 外键（§5.7）
ALTER TABLE applications
    ADD KEY idx_status_type (status, type_code),
    ADD KEY idx_applicant_status (applicant_id, status),
    ADD CONSTRAINT fk_app_type FOREIGN KEY (type_code) REFERENCES application_types(code),
    ADD CONSTRAINT fk_app_applicant FOREIGN KEY (applicant_id) REFERENCES users(id),
    ADD CONSTRAINT fk_app_approver FOREIGN KEY (approver_id) REFERENCES users(id);

ALTER TABLE application_approvals
    ADD CONSTRAINT fk_aprv_application FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE,
    ADD CONSTRAINT fk_aprv_approver FOREIGN KEY (approver_id) REFERENCES users(id);

ALTER TABLE application_approvers
    ADD CONSTRAINT fk_aprver_specific FOREIGN KEY (specific_user_id) REFERENCES users(id);

-- ⑩ availability 补 CHECK（§5.4）
ALTER TABLE availability
    ADD CONSTRAINT chk_avail_week CHECK (week BETWEEN 1 AND 30),
    ADD CONSTRAINT chk_avail_weekday CHECK (weekday BETWEEN 1 AND 5),
    ADD CONSTRAINT chk_avail_period CHECK (period BETWEEN 1 AND 4);

-- ⑪ 值班次数不再冗余存储于 users 表，改为运行时聚合
--    （SELECT user_id, COUNT(*) FROM duty_records WHERE status <> 'cancelled' GROUP BY user_id），
--    故此处不再回填任何列（§5.8）。

-- ⑫ 删除 duty_counters（历史计数表，已由运行时聚合取代，§5.8）
DROP TABLE IF EXISTS duty_counters;
