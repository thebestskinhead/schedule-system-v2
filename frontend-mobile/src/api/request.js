import axios from 'axios'
import { showToast } from 'vant'

// API 基础 URL
const API_BASE_URL = '/api/v1'

const request = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000
})

// 获取完整的 API 基础 URL（包含域名，用于原生下载）
export const getApiBaseUrl = () => {
  // 在浏览器中，使用当前域名 + API 路径
  if (typeof window !== 'undefined') {
    const origin = window.location.origin
    return `${origin}${API_BASE_URL}`
  }
  return API_BASE_URL
}

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
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
    // blob 类型（文件下载）直接返回，不做 JSON 解析
    if (response.config.responseType === 'blob') {
      return response.data
    }
    const { code, message, data } = response.data
    if (code !== 200) {
      showToast({ message: message || '请求失败', type: 'fail' })
      return Promise.reject(new Error(message))
    }
    return data
  },
  (error) => {
    const message = error.response?.data?.message || error.message || '网络错误'
    showToast({ message, type: 'fail' })
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default request
