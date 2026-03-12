import request from './request'

// 申请类型
export const getApplicationTypes = () => request.get('/application/types')

// 可申请的权限列表
export const getAvailablePermissions = () => request.get('/application/permissions/available')

// 我的申请
export const getMyApplications = (params) => request.get('/applications/my', { params })

// 创建申请
export const createApplication = (data) => request.post('/applications', data)

// 申请详情
export const getApplicationDetail = (id) => request.get(`/applications/${id}`)

// 取消申请
export const cancelApplication = (id) => request.post(`/applications/${id}/cancel`)

// 待我审批的申请
export const getPendingApprovals = (params) => request.get('/applications/pending', { params })

// 处理审批
export const processApproval = (id, data) => request.post(`/applications/${id}/approve`, data)

// 申请统计
export const getApplicationStats = () => request.get('/applications/stats')
