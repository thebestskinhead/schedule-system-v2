import request from './request'

// 认证相关
export const login = (data) => request.post('/user/login', data)
export const forgotPassword = (data) => request.post('/password/reset-request', data)
export const resetPassword = (data) => request.post('/password/reset', data)
export const checkToken = () => request.get('/user/profile')

// 用户信息
export const getUserInfo = () => request.get('/user/profile')
export const changePassword = (data) => request.post('/user/change-password', data)

// 排班相关用户查询
export const getUsersForSchedule = () => request.get('/users/for-schedule')

// 部门列表
export const getDepartments = () => request.get('/departments')

// 用户管理
export const getUserList = (params) => request.get('/admin/users', { params })
export const getUserListByDepartment = (dept) => request.get('/admin/users/by-dept', { params: { department: dept } })
export const createUser = (data) => request.post('/admin/users', data)
export const updateUser = (id, data) => request.put(`/admin/users/${id}`, data)
export const deleteUser = (id) => request.delete(`/admin/users/${id}`)
export const updateUserRole = (id, data) => request.put(`/admin/users/${id}/role`, data)

// 入职链路（录入/邀请/离职，需 user:manage）
export const createMember = (data) => request.post('/admin/members', data)
export const resendInvite = (id) => request.post(`/admin/members/${id}/resend-invite`)
export const revokeInvite = (id) => request.post(`/admin/members/${id}/revoke`)
export const resignMember = (id) => request.post(`/admin/members/${id}/resign`)

// 激活邀请（公开，受邀人自助）
export const activateInvitation = (token, password) =>
  request.post('/invitations/activate', { token, password })

// 扫码登录相关
export const qrLoginStart = () => request.post('/qr/login/start', {})
export const qrLoginPoll = (data) => request.post('/qr/login/poll', data)
