import request from './request'

// 系统安装状态
export const getInstallStatus = () => request.get('/system/installed')

// 测试数据库连接
export const testDBConnection = (data) => request.post('/system/test-db', data)

// 检查数据库状态（空/非空）
export const checkDatabase = (data) => request.post('/system/check-db', data)

// 初始化数据库表
export const initDatabaseTables = (data) => request.post('/system/init-tables', data)

// 创建管理员账号
export const createAdmin = (data) => request.post('/system/create-admin', data)

// SMTP配置
export const getSMTPConfig = () => request.get('/admin/smtp/config')
export const saveSMTPConfig = (data) => request.post('/admin/smtp/config', data)
export const testSMTPConfig = (data) => request.post('/admin/smtp/test', data)

// 网站域名配置
export const getSiteConfig = () => request.get('/admin/site/config')
export const saveSiteConfig = (data) => request.post('/admin/site/config', data)

// 临时权限
export const getTempPermissions = () => request.get('/admin/temp-permissions')
export const grantTempPermission = (data) => request.post('/admin/temp-permissions', data)
export const revokeTempPermission = (id) => request.delete(`/admin/temp-permissions/${id}`)
export const getMyTempPermissions = () => request.get('/temp-permissions/my')
export const cleanupExpiredPermissions = () => request.post('/admin/temp-permissions/cleanup')

// 权限列表
export const getPermissionList = () => request.get('/permissions/list')

// 部门列表
export const getDepartments = () => request.get('/departments')
