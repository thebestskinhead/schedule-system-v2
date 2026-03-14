import request from './request'

// 爬虫导入（Cookie方式）- 用于CrawlerImport页面
export const importCrawler = (data) => request.post('/crawler/import', data)

// 预览课表
export const previewCrawler = (data) => request.post('/crawler/preview', data)

// 获取导入任务状态
export const getImportStatus = (taskId) => request.get('/availability/import/status', { params: { task_id: taskId } })

// 获取导入任务列表
export const getImportTaskList = () => request.get('/availability/import/tasks')
