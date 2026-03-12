import request from './request'

// 获取当前周次（公开接口）
export const getCurrentWeek = () => {
  return request.get('/schedule/current-week')
}

// 更新当前周次（管理员）
export const updateCurrentWeek = (data) => {
  return request.post('/admin/schedule/current-week', data)
}

// 获取当前周排班（公开接口）
export const getCurrentWeekSchedule = () => {
  return request.get('/schedule/current')
}

export const previewSchedule = (data) => {
  return request.post('/admin/schedule/preview', data)
}

export const confirmSchedule = (data) => {
  return request.post('/admin/schedule/confirm', data)
}

export const getSchedule = (week) => {
  return request.get('/schedule', { params: { week } })
}

export const getMyDuties = () => {
  return request.get('/duty/my')
}

export const updateDutyStatus = (data) => {
  return request.put('/duty/status', data)
}

// 排班设置
export const getScheduleSettings = () => {
  return request.get('/admin/schedule/settings')
}

export const saveScheduleSettings = (data) => {
  return request.post('/admin/schedule/settings', data)
}

// 更新排班（添加/删除人员）
export const updateSchedule = (data) => {
  return request.post('/admin/schedule/update', data)
}

// 导出排班表
export const exportSchedule = (data) => {
  return request.post('/admin/schedule/export', data, {
    responseType: 'blob'
  })
}

// 模板管理
export const getTemplates = () => {
  return request.get('/admin/templates')
}

export const getTemplate = (id) => {
  return request.get(`/admin/templates/${id}`)
}

export const createTemplate = (data) => {
  return request.post('/admin/templates', data)
}

export const updateTemplate = (data) => {
  return request.put('/admin/templates', data)
}

export const deleteTemplate = (id) => {
  return request.delete(`/admin/templates/${id}`)
}

export const getPlaceholderHelp = () => {
  return request.get('/admin/templates/placeholders')
}
