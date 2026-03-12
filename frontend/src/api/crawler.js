import request from './request'

// 预览课表（只爬取前2周）
export const previewCrawler = (data) => {
  return request.post('/crawler/preview', data)
}

// 导入无课表（异步任务模式）
export const importCrawler = (data) => {
  return request.post('/availability/import/cookie', data)
}

// 获取导入任务状态
export const getImportStatus = (taskId) => {
  return request.get('/availability/import/status', {
    params: { task_id: taskId }
  })
}
