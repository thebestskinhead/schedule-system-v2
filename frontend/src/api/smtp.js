import request from './request'

// 获取所有SMTP配置
export const getSMTPConfigs = () => {
  return request.get('/admin/smtp/configs').then(res => res.data?.data || [])
}

// 获取当前启用的配置
export const getActiveSMTPConfig = () => {
  return request.get('/admin/smtp/config').then(res => res.data?.data)
}

// 保存配置
export const saveSMTPConfig = (config) => {
  return request.post('/admin/smtp/config', config)
}

// 删除配置
export const deleteSMTPConfig = (id) => {
  return request.delete(`/admin/smtp/config/${id}`)
}

// 测试发送邮件
export const testSMTP = (data) => {
  return request.post('/admin/smtp/test', data)
}

// 检查SMTP是否配置
export const checkSMTPConfig = () => {
  return request.get('/smtp/check').then(res => res.data?.data?.configured || false)
}

// 获取网站配置
export const getSiteConfig = () => {
  return request.get('/admin/site/config').then(res => res.data?.data)
}

// 保存网站配置
export const saveSiteConfig = (data) => {
  return request.post('/admin/site/config', data)
}

// 请求密码重置
export const requestPasswordReset = (email) => {
  return request.post('/password/reset-request', { email })
}

// 验证重置令牌
export const verifyResetToken = (token) => {
  return request.get('/password/reset-verify', { params: { token } })
}

// 重置密码
export const resetPassword = (token, password) => {
  return request.post('/password/reset', { token, password })
}
