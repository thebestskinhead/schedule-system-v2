import request from './request'

// 获取导入任务状态
export const getImportStatus = (taskId) => request.get('/availability/import/status', { params: { task_id: taskId } })

// 获取导入任务列表
export const getImportTaskList = () => request.get('/availability/import/tasks')

// 从教务系统导入课表（返回 need_login 表示需扫码登录后继续）
export const importFromEdu = () => request.post('/availability/import/edu')

// 扫码登录教务系统后继续导入
export const resumeEduImport = (taskId) => request.post('/availability/import/edu/resume', { task_id: taskId })

// 取消导入任务
export const cancelImport = (taskId) => request.post('/availability/import/cancel', { task_id: taskId })
