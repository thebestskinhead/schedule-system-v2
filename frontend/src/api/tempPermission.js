import request from './request'

export const tempPermissionAPI = {
  // 获取可授权的权限列表
  getPermissionList: () => request.get('/permissions/list'),

  // 授予临时权限（管理员）
  grant: (data) => request.post('/admin/temp-permissions', data),

  // 获取所有临时权限（管理员）
  list: () => request.get('/admin/temp-permissions'),

  // 撤销权限（管理员）
  revoke: (id) => request.delete(`/admin/temp-permissions/${id}`),

  // 获取我的临时权限
  getMy: () => request.get('/temp-permissions/my'),

  // 清理过期权限（管理员）
  cleanup: () => request.post('/admin/temp-permissions/cleanup'),
}
