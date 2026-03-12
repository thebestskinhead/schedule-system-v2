import request from './request'

export const getInstallStatus = () => request.get('/system/installed')
export const initSystem = (data) => request.post('/system/create-admin', data)

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
