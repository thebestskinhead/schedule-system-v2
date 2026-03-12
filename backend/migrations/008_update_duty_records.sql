-- 为用户表添加索引（如果还没有）
ALTER TABLE users ADD INDEX IF NOT EXISTS idx_department (department);
ALTER TABLE users ADD INDEX IF NOT EXISTS idx_dept_role (dept_role);
ALTER TABLE users ADD INDEX IF NOT EXISTS idx_role (role);

-- 为排班记录表添加部门字段（用于数据隔离查询优化）
ALTER TABLE duty_records ADD COLUMN IF NOT EXISTS department VARCHAR(50) AFTER user_id;
ALTER TABLE duty_records ADD INDEX IF NOT EXISTS idx_department (department);

-- 更新现有记录的部门字段
UPDATE duty_records dr
JOIN users u ON dr.user_id = u.id
SET dr.department = u.department
WHERE dr.department IS NULL;
