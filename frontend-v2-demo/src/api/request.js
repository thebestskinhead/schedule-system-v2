import axios from 'axios'
import { ElMessage } from 'element-plus'
import { mockAPI } from '../../mock/index.js'

// Mock模式标志
const isMockMode = import.meta.env.VITE_MOCK === 'true' || true

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 30000
})

// 请求拦截器
request.interceptors.request.use(
  async (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    
    // Mock模式拦截
    if (isMockMode) {
      try {
        const mockResult = await mockAPI.handle(config)
        
        // 构造一个模拟的Axios响应
        const mockResponse = {
          data: mockResult,
          status: mockResult.code === 200 ? 200 : mockResult.code,
          statusText: mockResult.code === 200 ? 'OK' : 'Error',
          headers: {},
          config
        }
        
        // 返回一个被拒绝的Promise，但带上Mock响应标记
        return Promise.reject({
          __isMock: true,
          __mockResponse: mockResponse
        })
      } catch (error) {
        // Mock处理出错
        console.error('[Mock] 处理请求失败:', error)
      }
    }
    
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    const { code, message, data } = response.data
    if (code !== 200) {
      ElMessage.error(message || '请求失败')
      return Promise.reject(new Error(message))
    }
    return data
  },
  (error) => {
    // 处理Mock响应
    if (error.__isMock && error.__mockResponse) {
      const response = error.__mockResponse
      const { code, message, data } = response.data
      
      if (code !== 200) {
        ElMessage.error(message || '请求失败')
        return Promise.reject(new Error(message))
      }
      return data
    }
    
    const message = error.response?.data?.message || error.message || '网络错误'
    ElMessage.error(message)
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default request
