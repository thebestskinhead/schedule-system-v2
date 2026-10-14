-- 014_fix_temp_perm_reason.sql
-- user_permissions_temp.reason 收紧为 NOT NULL DEFAULT ''：
-- 历史上该列可空，而模型 Reason 为非指针字符串，SELECT * 遇 NULL 记录时扫描失败，
-- 会导致该用户全部临时权限加载失败且 /temp-permissions/my 返回 500。
-- 先清洗存量 NULL，再收紧列约束。

UPDATE user_permissions_temp SET reason = '' WHERE reason IS NULL;

ALTER TABLE user_permissions_temp
    MODIFY COLUMN reason VARCHAR(255) NOT NULL DEFAULT '' COMMENT '授权原因';

-- 修正存量坏数据：部门级权限 resource_id=0 会被判定层视为全部门通配（越权）。
-- 按被授权人 users 表部门回填正确 ID；部门无效的记录直接置为失效。
UPDATE user_permissions_temp t
JOIN users u ON t.user_id = u.id
SET t.resource_id = CASE u.department
        WHEN '办公室' THEN 1
        WHEN '竞赛部' THEN 2
        WHEN '项目部'  THEN 3
        WHEN '科普部'  THEN 4
        ELSE -1
    END
WHERE t.resource_type = 'dept' AND t.resource_id = 0;

UPDATE user_permissions_temp SET is_active = FALSE
WHERE resource_type = 'dept' AND resource_id = -1;
