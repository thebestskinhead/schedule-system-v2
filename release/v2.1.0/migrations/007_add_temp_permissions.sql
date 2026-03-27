-- 用户临时权限表
CREATE TABLE IF NOT EXISTS user_permissions_temp (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '被授权用户ID',
    permission VARCHAR(50) NOT NULL COMMENT '权限代码',
    resource_type VARCHAR(20) DEFAULT 'all' COMMENT '资源类型(all/dept/user)',
    resource_id INT DEFAULT 0 COMMENT '资源ID(部门ID或用户ID)',
    granted_by INT NOT NULL COMMENT '授权人ID',
    reason VARCHAR(255) COMMENT '授权原因',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP DEFAULT '2099-12-31 23:59:59' COMMENT '过期时间',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否有效',
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户临时权限表';
