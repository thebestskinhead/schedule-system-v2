import request from './request'

// 公开接口
export const getCurrentWeek = () => request.get('/schedule/current-week')
export const getSchedule = (params) => request.get('/schedule', { params })

// 需要权限的接口
export const updateCurrentWeek = (data) => request.post('/admin/schedule/current-week', data)
export const previewSchedule = (data) => request.post('/admin/schedule/preview', data)
export const confirmSchedule = (data) => request.post('/admin/schedule/confirm', data)
export const getScheduleSettings = () => request.get('/admin/schedule/settings')
export const saveScheduleSettings = (data) => request.post('/admin/schedule/settings', data)
export const updateSchedule = (data) => request.post('/admin/schedule/update', data)
export const exportToExcel = (data) => request.post('/admin/schedule/export', data, { responseType: 'blob' })

// 值班相关
export const getMyDuties = () => request.get('/duty/my')
export const updateDutyStatus = (data) => request.put('/duty/status', data)

// 模板相关
export const getTemplates = () => request.get('/admin/templates')
export const createTemplate = (data) => request.post('/admin/templates', data)
export const updateTemplate = (data) => request.put('/admin/templates', data)
export const deleteTemplate = (id) => request.delete(`/admin/templates/${id}`)

// 分工配置
export const getDutyAssignments = (params) => request.get('/admin/duty-assignments', { params })
export const saveDutyAssignments = (data) => request.post('/admin/duty-assignments', data)
export const updateDutyAssignment = (id, data) => request.put('/admin/duty-assignments', { id, ...data })
export const deleteDutyAssignment = (id) => request.delete(`/admin/duty-assignments/${id}`)
export const getMyDeptAssignment = (params) => request.get('/duty-assignments/my-dept', { params })

// 学期起始日设置
export const getSemesterStartDate = () => request.get('/admin/schedule/semester-start')
export const updateSemesterStartDate = (data) => request.post('/admin/schedule/semester-start', data)
