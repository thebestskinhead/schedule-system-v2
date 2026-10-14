-- 013_add_db_logic.sql
-- 存储过程 + 事件调度（依据 docs/数据库.md v2.1 §7 / §8 / §12.2 / §12.4.6）
-- 注意：本文件含存储过程/事件，必须使用 DELIMITER 以正确分隔语句。

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
       SET pt.is_active = 0, pt.updated_at = NOW()
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
