-- 添加学期起始日字段到排班设置表
ALTER TABLE schedule_settings ADD COLUMN IF NOT EXISTS semester_start_date DATE DEFAULT NULL COMMENT '学期起始日';

-- 更新已有数据，设置默认值为当前年份的9月1日（秋季学期）
UPDATE schedule_settings SET semester_start_date = CONCAT(YEAR(CURDATE()), '-09-01') WHERE semester_start_date IS NULL;
