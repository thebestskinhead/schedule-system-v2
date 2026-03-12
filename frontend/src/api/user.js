import request from './request'

export const register = (data) => {
  return request.post('/user/register', data)
}

export const login = (data) => {
  return request.post('/user/login', data)
}

export const getProfile = () => {
  return request.get('/user/profile')
}

export const getUserList = () => {
  return request.get('/admin/users')
}

// 用于排班的用户列表查询（只需要排班权限）
export const getUsersForSchedule = () => {
  return request.get('/users/for-schedule')
}

export const setUserRole = (data) => {
  return request.post('/admin/users/role', data)
}

export const updateProfile = (data) => {
  return request.put('/user/profile', data)
}

// 统一的 userAPI 对象
export const userAPI = {
  register,
  login,
  getProfile,
  getList: getUserList,
  setUserRole,
  updateProfile,
}
