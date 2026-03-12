import request from './request'

export const userAdminAPI = {
  // 获取部门列表
  getDepartments: () => request.get('/departments'),
  
  // 按部门获取用户列表
  listByDept: (department) => request.get('/admin/users/by-dept', { params: { department } }),
  
  // 根据筛选条件获取用户
  listByFilter: (params) => request.get('/admin/users/filter', { params }),
  
  // 设置用户部门
  setDepartment: (id, department) => request.put(`/admin/users/${id}/department`, { department }),
  
  // 设置用户部门角色
  setDeptRole: (id, deptRole) => request.put(`/admin/users/${id}/dept-role`, { dept_role: deptRole }),
}
