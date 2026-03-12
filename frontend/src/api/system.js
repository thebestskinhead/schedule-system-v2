import axios from 'axios'

// 获取系统安装状态
export const getInstallStatus = () => {
  return axios.get('/api/v1/system/installed')
}

// 测试数据库连接
export const testDBConnection = (data) => {
  return axios.post('/api/v1/system/test-db', data)
}

// 检查数据库状态（空/非空）
export const checkDatabase = (data) => {
  return axios.post('/api/v1/system/check-db', data)
}

// 初始化数据库表
export const initDatabaseTables = (data) => {
  return axios.post('/api/v1/system/init-tables', data)
}

// 创建管理员账号
export const createAdmin = (data) => {
  return axios.post('/api/v1/system/create-admin', data)
}
