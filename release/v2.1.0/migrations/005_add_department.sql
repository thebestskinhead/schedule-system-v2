-- 添加部门和部门角色字段
-- 部门角色: dept_admin(部门管理员), dept_member(部门成员)

-- 1. 添加 department 字段
ALTER TABLE users ADD COLUMN department VARCHAR(50) DEFAULT '科普部' COMMENT '所属部门' AFTER role;

-- 2. 添加 dept_role 字段
ALTER TABLE users ADD COLUMN dept_role ENUM('dept_admin', 'dept_member') DEFAULT 'dept_member' COMMENT '部门角色' AFTER department;

-- 3. 更新现有用户为科普部部门管理员
UPDATE users SET department = '科普部', dept_role = 'dept_admin';

-- 4. 添加部门索引
ALTER TABLE users ADD INDEX idx_department (department);
ALTER TABLE users ADD INDEX idx_dept_role (dept_role);
