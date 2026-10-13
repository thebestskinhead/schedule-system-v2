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
// status 可选值：''（全部）/ 'pending'（待入职）/ 'active'（在职）/ 'resigned'（离职）
export const getUserList = (params) => request.get('/admin/users', { params })
export const getUserListByDepartment = (dept) => request.get('/admin/users/by-dept', { params: { department: dept } })
export const createUser = (data) => request.post('/admin/users', data)
export const updateUser = (id, data) => request.put(`/admin/users/${id}`, data)
export const deleteUser = (id) => request.delete(`/admin/users/${id}`)
export const updateUserRole = (id, data) => request.put(`/admin/users/${id}/role`, data)

// 录入正式成员（邀请激活链路，H.2 / H.3）
export const createMember = (data) => request.post('/admin/members', data)
export const batchCreateMembers = (items) => request.post('/admin/members/batch', { items })
export const resendInvite = (id) => request.post(`/admin/members/${id}/resend-invite`)
export const revokeInvite = (id) => request.post(`/admin/members/${id}/revoke`)
export const resignMember = (id) => request.post(`/admin/members/${id}/resign`)
export const previewBatchResign = (user_ids, exclude_ids = []) =>
  request.post('/admin/members/batch-resign/preview', { user_ids, exclude_ids })
export const batchResignMembers = (user_ids, exclude_ids = []) =>
  request.post('/admin/members/batch-resign', { user_ids, exclude_ids })

// 激活邀请（公开，受邀人自助）
export const activateInvitation = (token, password) =>
  request.post('/invitations/activate', { token, password })

// 扫码登录相关
export const qrLoginStart = () => request.post('/qr/login/start', {})
export const qrLoginPoll = (data) => request.post('/qr/login/poll', data)
