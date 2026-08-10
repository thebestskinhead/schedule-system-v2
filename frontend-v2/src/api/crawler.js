import request from './request'

// 获取导入任务状态
export const getImportStatus = (taskId) => request.get('/availability/import/status', { params: { task_id: taskId } })

// 获取导入任务列表
export const getImportTaskList = () => request.get('/availability/import/tasks')
