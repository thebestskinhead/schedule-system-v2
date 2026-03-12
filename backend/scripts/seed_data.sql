-- 排班系统 V2 示例数据
USE schedule_system_v2;

-- 清空现有数据（保留系统配置）
DELETE FROM availability WHERE user_id > 0;
DELETE FROM duty_records WHERE user_id > 0;
DELETE FROM duty_counters WHERE user_id > 0;
DELETE FROM weekly_duty_assignments WHERE id > 0;
DELETE FROM user_permissions_temp WHERE id > 0;
DELETE FROM users WHERE id > 0;

-- 重置自增ID
ALTER TABLE users AUTO_INCREMENT = 1;
ALTER TABLE availability AUTO_INCREMENT = 1;
ALTER TABLE duty_records AUTO_INCREMENT = 1;
ALTER TABLE weekly_duty_assignments AUTO_INCREMENT = 1;
ALTER TABLE user_permissions_temp AUTO_INCREMENT = 1;

-- ============================================
-- 1. 插入示例用户
-- 密码都是: 123456 (bcrypt加密)
-- ============================================

-- 系统管理员
INSERT INTO users (id, student_id, name, email, password, role, department, dept_role, is_active) VALUES
(1, '2024001', '管理员', 'admin@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQz3XJwK9XqZzZzZzZzZzZzZzZz', 'admin', '办公室', 'dept_admin', 1);

-- 办公室用户 (资源ID=1)
INSERT INTO users (student_id, name, email, password, role, department, dept_role, is_active) VALUES
('2024101', '张三', 'zhangsan@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQz3XJwK9XqZzZzZzZzZzZzZzZz', 'user', '办公室', 'dept_member', 1),
('2024102', '李四', 'lisi@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQz3XJwK9XqZzZzZzZzZzZzZzZz', 'user', '办公室', 'dept_admin', 1),
('2024103', '王五', 'wangwu@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQz3XJwK9XqZzZzZzZzZzZzZzZz', 'user', '办公室', 'dept_member', 1);

-- 竞赛部用户 (资源ID=2)
INSERT INTO users (student_id, name, email, password, role, department, dept_role, is_active) VALUES
('2024201', '赵六', 'zhaoliu@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQz3XJwK9XqZzZzZzZzZzZzZzZz', 'user', '竞赛部', 'dept_admin', 1),
('2024202', '钱七', 'qianqi@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQz3XJwK9XqZzZzZzZzZzZzZzZz', 'user', '竞赛部', 'dept_member', 1),
('2024203', '孙八', 'sunba@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQz3XJwK9XqZzZzZzZzZzZzZzZz', 'user', '竞赛部', 'dept_member', 1);

-- 项目部用户 (资源ID=3)
INSERT INTO users (student_id, name, email, password, role, department, dept_role, is_active) VALUES
('2024301', '周九', 'zhoujiu@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQz3XJwK9XqZzZzZzZzZzZzZzZz', 'user', '项目部', 'dept_admin', 1),
('2024302', '吴十', 'wushi@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQz3XJwK9XqZzZzZzZzZzZzZzZz', 'user', '项目部', 'dept_member', 1);

-- 科普部用户 (资源ID=4)
INSERT INTO users (student_id, name, email, password, role, department, dept_role, is_active) VALUES
('2024401', '郑十一', 'zheng11@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQz3XJwK9XqZzZzZzZzZzZzZzZz', 'user', '科普部', 'dept_admin', 1),
('2024402', '陈十二', 'chen12@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQz3XJwK9XqZzZzZzZzZzZzZzZz', 'user', '科普部', 'dept_member', 1);

-- ============================================
-- 2. 插入无课表数据
-- 为用户创建第1-4周的无课时间
-- ============================================

-- 张三 (user_id=2) 的无课时间 - 周一、三、五 全天无课，周二四有课
INSERT INTO availability (user_id, week, weekday, period) VALUES
(2, 1, 1, 1), (2, 1, 1, 2), (2, 1, 1, 3), (2, 1, 1, 4),
(2, 1, 3, 1), (2, 1, 3, 2), (2, 1, 3, 3), (2, 1, 3, 4),
(2, 1, 5, 1), (2, 1, 5, 2), (2, 1, 5, 3), (2, 1, 5, 4),
(2, 2, 1, 1), (2, 2, 1, 2), (2, 2, 1, 3), (2, 2, 1, 4),
(2, 2, 3, 1), (2, 2, 3, 2), (2, 2, 3, 3), (2, 2, 3, 4),
(2, 2, 5, 1), (2, 2, 5, 2), (2, 2, 5, 3), (2, 2, 5, 4),
(2, 3, 1, 1), (2, 3, 1, 2), (2, 3, 1, 3), (2, 3, 1, 4),
(2, 3, 3, 1), (2, 3, 3, 2), (2, 3, 3, 3), (2, 3, 3, 4),
(2, 3, 5, 1), (2, 3, 5, 2), (2, 3, 5, 3), (2, 3, 5, 4),
(2, 4, 1, 1), (2, 4, 1, 2), (2, 4, 1, 3), (2, 4, 1, 4),
(2, 4, 3, 1), (2, 4, 3, 2), (2, 4, 3, 3), (2, 4, 3, 4),
(2, 4, 5, 1), (2, 4, 5, 2), (2, 4, 5, 3), (2, 4, 5, 4);

-- 李四 (user_id=3) 的无课时间 - 周二、四 全天无课
INSERT INTO availability (user_id, week, weekday, period) VALUES
(3, 1, 2, 1), (3, 1, 2, 2), (3, 1, 2, 3), (3, 1, 2, 4),
(3, 1, 4, 1), (3, 1, 4, 2), (3, 1, 4, 3), (3, 1, 4, 4),
(3, 2, 2, 1), (3, 2, 2, 2), (3, 2, 2, 3), (3, 2, 2, 4),
(3, 2, 4, 1), (3, 2, 4, 2), (3, 2, 4, 3), (3, 2, 4, 4),
(3, 3, 2, 1), (3, 3, 2, 2), (3, 3, 2, 3), (3, 3, 2, 4),
(3, 3, 4, 1), (3, 3, 4, 2), (3, 3, 4, 3), (3, 3, 4, 4),
(3, 4, 2, 1), (3, 4, 2, 2), (3, 4, 2, 3), (3, 4, 2, 4),
(3, 4, 4, 1), (3, 4, 4, 2), (3, 4, 4, 3), (3, 4, 4, 4);

-- 王五 (user_id=4) 的无课时间 - 每天只有下午无课(第3、4节)
INSERT INTO availability (user_id, week, weekday, period) VALUES
(4, 1, 1, 3), (4, 1, 1, 4), (4, 1, 2, 3), (4, 1, 2, 4),
(4, 1, 3, 3), (4, 1, 3, 4), (4, 1, 4, 3), (4, 1, 4, 4),
(4, 1, 5, 3), (4, 1, 5, 4),
(4, 2, 1, 3), (4, 2, 1, 4), (4, 2, 2, 3), (4, 2, 2, 4),
(4, 2, 3, 3), (4, 2, 3, 4), (4, 2, 4, 3), (4, 2, 4, 4),
(4, 2, 5, 3), (4, 2, 5, 4),
(4, 3, 1, 3), (4, 3, 1, 4), (4, 3, 2, 3), (4, 3, 2, 4),
(4, 3, 3, 3), (4, 3, 3, 4), (4, 3, 4, 3), (4, 3, 4, 4),
(4, 3, 5, 3), (4, 3, 5, 4),
(4, 4, 1, 3), (4, 4, 1, 4), (4, 4, 2, 3), (4, 4, 2, 4),
(4, 4, 3, 3), (4, 4, 3, 4), (4, 4, 4, 3), (4, 4, 4, 4),
(4, 4, 5, 3), (4, 4, 5, 4);

-- 赵六 (user_id=5) 的无课时间 - 每天只有上午无课(第1、2节)
INSERT INTO availability (user_id, week, weekday, period) VALUES
(5, 1, 1, 1), (5, 1, 1, 2), (5, 1, 2, 1), (5, 1, 2, 2),
(5, 1, 3, 1), (5, 1, 3, 2), (5, 1, 4, 1), (5, 1, 4, 2),
(5, 1, 5, 1), (5, 1, 5, 2),
(5, 2, 1, 1), (5, 2, 1, 2), (5, 2, 2, 1), (5, 2, 2, 2),
(5, 2, 3, 1), (5, 2, 3, 2), (5, 2, 4, 1), (5, 2, 4, 2),
(5, 2, 5, 1), (5, 2, 5, 2),
(5, 3, 1, 1), (5, 3, 1, 2), (5, 3, 2, 1), (5, 3, 2, 2),
(5, 3, 3, 1), (5, 3, 3, 2), (5, 3, 4, 1), (5, 3, 4, 2),
(5, 3, 5, 1), (5, 3, 5, 2),
(5, 4, 1, 1), (5, 4, 1, 2), (5, 4, 2, 1), (5, 4, 2, 2),
(5, 4, 3, 1), (5, 4, 3, 2), (5, 4, 4, 1), (5, 4, 4, 2),
(5, 4, 5, 1), (5, 4, 5, 2);

-- 钱七 (user_id=6) 的无课时间 - 周一到周五第2、3节
INSERT INTO availability (user_id, week, weekday, period) VALUES
(6, 1, 1, 2), (6, 1, 1, 3), (6, 1, 2, 2), (6, 1, 2, 3),
(6, 1, 3, 2), (6, 1, 3, 3), (6, 1, 4, 2), (6, 1, 4, 3),
(6, 1, 5, 2), (6, 1, 5, 3),
(6, 2, 1, 2), (6, 2, 1, 3), (6, 2, 2, 2), (6, 2, 2, 3),
(6, 2, 3, 2), (6, 2, 3, 3), (6, 2, 4, 2), (6, 2, 4, 3),
(6, 2, 5, 2), (6, 2, 5, 3),
(6, 3, 1, 2), (6, 3, 1, 3), (6, 3, 2, 2), (6, 3, 2, 3),
(6, 3, 3, 2), (6, 3, 3, 3), (6, 3, 4, 2), (6, 3, 4, 3),
(6, 3, 5, 2), (6, 3, 5, 3),
(6, 4, 1, 2), (6, 4, 1, 3), (6, 4, 2, 2), (6, 4, 2, 3),
(6, 4, 3, 2), (6, 4, 3, 3), (6, 4, 4, 2), (6, 4, 4, 3),
(6, 4, 5, 2), (6, 4, 5, 3);

-- 孙八 (user_id=7) 的无课时间 - 只有周三全天无课
INSERT INTO availability (user_id, week, weekday, period) VALUES
(7, 1, 3, 1), (7, 1, 3, 2), (7, 1, 3, 3), (7, 1, 3, 4),
(7, 2, 3, 1), (7, 2, 3, 2), (7, 2, 3, 3), (7, 2, 3, 4),
(7, 3, 3, 1), (7, 3, 3, 2), (7, 3, 3, 3), (7, 3, 3, 4),
(7, 4, 3, 1), (7, 4, 3, 2), (7, 4, 3, 3), (7, 4, 3, 4);

-- 周九 (user_id=8) 的无课时间 - 周一、二全天无课
INSERT INTO availability (user_id, week, weekday, period) VALUES
(8, 1, 1, 1), (8, 1, 1, 2), (8, 1, 1, 3), (8, 1, 1, 4),
(8, 1, 2, 1), (8, 1, 2, 2), (8, 1, 2, 3), (8, 1, 2, 4),
(8, 2, 1, 1), (8, 2, 1, 2), (8, 2, 1, 3), (8, 2, 1, 4),
(8, 2, 2, 1), (8, 2, 2, 2), (8, 2, 2, 3), (8, 2, 2, 4),
(8, 3, 1, 1), (8, 3, 1, 2), (8, 3, 1, 3), (8, 3, 1, 4),
(8, 3, 2, 1), (8, 3, 2, 2), (8, 3, 2, 3), (8, 3, 2, 4),
(8, 4, 1, 1), (8, 4, 1, 2), (8, 4, 1, 3), (8, 4, 1, 4),
(8, 4, 2, 1), (8, 4, 2, 2), (8, 4, 2, 3), (8, 4, 2, 4);

-- 吴十 (user_id=9) 的无课时间 - 周四、五全天无课
INSERT INTO availability (user_id, week, weekday, period) VALUES
(9, 1, 4, 1), (9, 1, 4, 2), (9, 1, 4, 3), (9, 1, 4, 4),
(9, 1, 5, 1), (9, 1, 5, 2), (9, 1, 5, 3), (9, 1, 5, 4),
(9, 2, 4, 1), (9, 2, 4, 2), (9, 2, 4, 3), (9, 2, 4, 4),
(9, 2, 5, 1), (9, 2, 5, 2), (9, 2, 5, 3), (9, 2, 5, 4),
(9, 3, 4, 1), (9, 3, 4, 2), (9, 3, 4, 3), (9, 3, 4, 4),
(9, 3, 5, 1), (9, 3, 5, 2), (9, 3, 5, 3), (9, 3, 5, 4),
(9, 4, 4, 1), (9, 4, 4, 2), (9, 4, 4, 3), (9, 4, 4, 4),
(9, 4, 5, 1), (9, 4, 5, 2), (9, 4, 5, 3), (9, 4, 5, 4);

-- 郑十一 (user_id=10) 的无课时间 - 每天第1节
INSERT INTO availability (user_id, week, weekday, period) VALUES
(10, 1, 1, 1), (10, 1, 2, 1), (10, 1, 3, 1), (10, 1, 4, 1), (10, 1, 5, 1),
(10, 2, 1, 1), (10, 2, 2, 1), (10, 2, 3, 1), (10, 2, 4, 1), (10, 2, 5, 1),
(10, 3, 1, 1), (10, 3, 2, 1), (10, 3, 3, 1), (10, 3, 4, 1), (10, 3, 5, 1),
(10, 4, 1, 1), (10, 4, 2, 1), (10, 4, 3, 1), (10, 4, 4, 1), (10, 4, 5, 1);

-- 陈十二 (user_id=11) 的无课时间 - 每天第4节
INSERT INTO availability (user_id, week, weekday, period) VALUES
(11, 1, 1, 4), (11, 1, 2, 4), (11, 1, 3, 4), (11, 1, 4, 4), (11, 1, 5, 4),
(11, 2, 1, 4), (11, 2, 2, 4), (11, 2, 3, 4), (11, 2, 4, 4), (11, 2, 5, 4),
(11, 3, 1, 4), (11, 3, 2, 4), (11, 3, 3, 4), (11, 3, 4, 4), (11, 3, 5, 4),
(11, 4, 1, 4), (11, 4, 2, 4), (11, 4, 3, 4), (11, 4, 4, 4), (11, 4, 5, 4);

-- 林十三 (user_id=12) 的无课时间 - 周一周五第2节，周三全天
INSERT INTO availability (user_id, week, weekday, period) VALUES
(12, 1, 1, 2), (12, 1, 3, 1), (12, 1, 3, 2), (12, 1, 3, 3), (12, 1, 3, 4), (12, 1, 5, 2),
(12, 2, 1, 2), (12, 2, 3, 1), (12, 2, 3, 2), (12, 2, 3, 3), (12, 2, 3, 4), (12, 2, 5, 2),
(12, 3, 1, 2), (12, 3, 3, 1), (12, 3, 3, 2), (12, 3, 3, 3), (12, 3, 3, 4), (12, 3, 5, 2),
(12, 4, 1, 2), (12, 4, 3, 1), (12, 4, 3, 2), (12, 4, 3, 3), (12, 4, 3, 4), (12, 4, 5, 2);

-- 黄十四 (user_id=13) 的无课时间 - 周二周四第1、2节，周五全天
INSERT INTO availability (user_id, week, weekday, period) VALUES
(13, 1, 2, 1), (13, 1, 2, 2), (13, 1, 4, 1), (13, 1, 4, 2),
(13, 1, 5, 1), (13, 1, 5, 2), (13, 1, 5, 3), (13, 1, 5, 4),
(13, 2, 2, 1), (13, 2, 2, 2), (13, 2, 4, 1), (13, 2, 4, 2),
(13, 2, 5, 1), (13, 2, 5, 2), (13, 2, 5, 3), (13, 2, 5, 4),
(13, 3, 2, 1), (13, 3, 2, 2), (13, 3, 4, 1), (13, 3, 4, 2),
(13, 3, 5, 1), (13, 3, 5, 2), (13, 3, 5, 3), (13, 3, 5, 4),
(13, 4, 2, 1), (13, 4, 2, 2), (13, 4, 4, 1), (13, 4, 4, 2),
(13, 4, 5, 1), (13, 4, 5, 2), (13, 4, 5, 3), (13, 4, 5, 4);

-- ============================================
-- 3. 插入排班设置
-- ============================================
INSERT INTO schedule_settings (admin_id, need_per_cell, min_per_cell, max_per_day, max_per_week, export_title) VALUES
(1, 2, 1, 1, 2, '第{week}周排班表');

-- ============================================
-- 4. 插入每周值班分工示例（第1周）
-- ============================================
INSERT INTO weekly_duty_assignments (week, department, weekday, is_assigned, created_by) VALUES
-- 办公室第1周值班安排：周一、三、五值班
(1, '办公室', 1, TRUE, 1), (1, '办公室', 3, TRUE, 1), (1, '办公室', 5, TRUE, 1),
-- 竞赛部第1周值班安排：周二、四值班
(1, '竞赛部', 2, TRUE, 1), (1, '竞赛部', 4, TRUE, 1),
-- 项目部第1周值班安排：周一、四值班
(1, '项目部', 1, TRUE, 1), (1, '项目部', 4, TRUE, 1),
-- 科普部第1周值班安排：周三、五值班
(1, '科普部', 3, TRUE, 1), (1, '科普部', 5, TRUE, 1);

-- ============================================
-- 5. 插入值班记录示例（第1周部分排班）
-- ============================================
-- 办公室第1周周一值班
INSERT INTO duty_records (week, weekday, period, user_id, department, assigned_by, status) VALUES
(1, 1, 1, 2, '办公室', 1, 'confirmed'),  -- 张三
(1, 1, 2, 3, '办公室', 1, 'confirmed'),  -- 李四
(1, 1, 3, 2, '办公室', 1, 'confirmed'),  -- 张三
(1, 1, 4, 3, '办公室', 1, 'confirmed');  -- 李四

-- 竞赛部第1周周二下午值班
INSERT INTO duty_records (week, weekday, period, user_id, department, assigned_by, status) VALUES
(1, 2, 3, 6, '竞赛部', 1, 'confirmed'),  -- 钱七
(1, 2, 4, 6, '竞赛部', 1, 'confirmed');  -- 钱七

-- ============================================
-- 6. 更新值班次数统计
-- ============================================
INSERT INTO duty_counters (user_id, total_count) VALUES
(2, 2),  -- 张三
(3, 2),  -- 李四
(6, 2);  -- 钱七

-- ============================================
-- 完成
-- ============================================
SELECT '示例数据导入完成！' AS message;
SELECT CONCAT('用户数量: ', COUNT(*)) AS info FROM users;
SELECT CONCAT('无课表记录: ', COUNT(*)) AS info FROM availability;
SELECT CONCAT('值班记录: ', COUNT(*)) AS info FROM duty_records;
