-- The server performs this upgrade automatically on startup. Safe to run again manually.
SET @has_week_anchor = (SELECT COUNT(*) FROM information_schema.columns
    WHERE table_schema = DATABASE() AND table_name = 'schedule_settings' AND column_name = 'current_week_set_at');
SET @week_anchor_sql = IF(@has_week_anchor = 0,
    'ALTER TABLE schedule_settings ADD COLUMN current_week_set_at DATE DEFAULT NULL COMMENT ''手动周次的起算日期''',
    'SELECT 1');
PREPARE week_anchor_stmt FROM @week_anchor_sql;
EXECUTE week_anchor_stmt;
DEALLOCATE PREPARE week_anchor_stmt;
ALTER TABLE schedule_settings ALTER COLUMN auto_increment SET DEFAULT 1;
UPDATE schedule_settings
SET current_week_set_at = DATE_SUB(DATE(updated_at), INTERVAL WEEKDAY(updated_at) DAY)
WHERE current_week_set_at IS NULL AND semester_start_date IS NULL;
UPDATE schedule_settings SET auto_increment = 1 WHERE auto_increment = 0;
