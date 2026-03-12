import request from './request'

export const dutyAssignmentAPI = {
  // 发布每周分工（管理员）
  publish: (data) => request.post('/admin/duty-assignments', data),
  
  // 获取分工列表（管理员）
  list: (week) => request.get('/admin/duty-assignments', { params: { week } }),
  
  // 获取分工视图（管理员）
  getView: (week) => request.get('/admin/duty-assignments/view', { params: { week } }),
  
  // 更新分工（管理员）
  update: (data) => request.put('/admin/duty-assignments', data),
  
  // 删除分工（管理员）
  delete: (id) => request.delete(`/admin/duty-assignments/${id}`),
  
  // 获取本部门分工（所有登录用户）
  getMyDept: (week) => request.get('/duty-assignments/my-dept', { params: { week } }),
}
