import request from './request'

// 认证相关
export const login = (data) => request.post('/user/login', data)
export const register = (data) => request.post('/user/register', data)
export const forgotPassword = (data) => request.post('/password/reset-request', data)
export const resetPassword = (data) => request.post('/password/reset', data)
export const checkToken = () => request.get('/user/profile')

// 用户信息
export const getUserInfo = () => request.get('/user/profile')
export const changePassword = (data) => request.post('/user/change-password', data)

// 排班相关用户查询
export const getUsersForSchedule = () => request.get('/users/for-schedule')

// 用户管理
export const getUserList = (params) => request.get('/admin/users', { params })
export const createUser = (data) => request.post('/admin/users', data)
export const updateUser = (id, data) => request.put(`/admin/users/${id}`, data)
export const deleteUser = (id) => request.delete(`/admin/users/${id}`)
export const updateUserRole = (id, data) => request.put(`/admin/users/${id}/role`, data)
