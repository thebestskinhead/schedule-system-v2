/**
 * Mock认证服务 - 纯前端演示用
 * 模拟后端认证流程，支持多角色切换
 */

import { PRESET_USERS, getUserById, getUserByStudentId, getSanitizedUser } from './preset-users.js'

// 存储键名
const STORAGE_KEY_TOKEN = 'mock_auth_token'
const STORAGE_KEY_USER_ID = 'mock_current_user_id'

class MockAuthService {
  constructor() {
    this.currentUser = null
    this.token = null
    this.subscribers = []
    this.isMockMode = true
    
    // 尝试恢复会话
    this.restoreSession()
  }

  /**
   * 模拟网络延迟
   */
  async delay(ms = 300) {
    const delayTime = ms + Math.random() * 200
    return new Promise(resolve => setTimeout(resolve, delayTime))
  }

  /**
   * 模拟登录
   */
  async login(credentials) {
    await this.delay(400)
    
    const { student_id, password } = credentials
    
    // 查找用户
    const user = getUserByStudentId(student_id)
    
    if (!user) {
      throw new Error('用户名或密码错误')
    }
    
    if (user.password !== password) {
      throw new Error('用户名或密码错误')
    }
    
    // 设置当前用户
    this.currentUser = user
    this.token = this.generateToken(user)
    
    // 持久化到localStorage
    localStorage.setItem(STORAGE_KEY_TOKEN, this.token)
    localStorage.setItem(STORAGE_KEY_USER_ID, String(user.id))
    
    // 通知订阅者
    this.notifySubscribers({ type: 'login', user: getSanitizedUser(user) })
    
    return {
      token: this.token,
      user: getSanitizedUser(user)
    }
  }

  /**
   * 快速切换角色（Demo专用功能）
   */
  async switchRole(userId) {
    await this.delay(200)
    
    const user = getUserById(userId)
    if (!user) {
      throw new Error('用户不存在')
    }
    
    this.currentUser = user
    this.token = this.generateToken(user)
    
    localStorage.setItem(STORAGE_KEY_TOKEN, this.token)
    localStorage.setItem(STORAGE_KEY_USER_ID, String(user.id))
    
    this.notifySubscribers({ type: 'switch', user: getSanitizedUser(user) })
    
    return {
      token: this.token,
      user: getSanitizedUser(user)
    }
  }

  /**
   * 验证Token并获取用户信息
   */
  async checkToken(token) {
    await this.delay(100)
    
    // 如果没有传入token，使用当前token
    const checkToken = token || this.token
    
    if (!checkToken) {
      throw new Error('未提供Token')
    }
    
    // 简化验证：解析token中的用户ID
    try {
      const payload = this.parseToken(checkToken)
      if (!payload || !payload.sub) {
        throw new Error('Token格式错误')
      }
      
      // 检查token是否过期
      if (payload.exp && payload.exp < Date.now()) {
        throw new Error('Token已过期')
      }
      
      const user = getUserById(parseInt(payload.sub))
      if (!user) {
        throw new Error('用户不存在')
      }
      
      return getSanitizedUser(user)
    } catch (error) {
      throw new Error('Token无效: ' + error.message)
    }
  }

  /**
   * 获取当前登录用户信息
   */
  async getCurrentUser() {
    await this.delay(100)
    
    if (!this.currentUser) {
      throw new Error('未登录')
    }
    
    return getSanitizedUser(this.currentUser)
  }

  /**
   * 登出
   */
  async logout() {
    await this.delay(100)
    
    this.currentUser = null
    this.token = null
    
    localStorage.removeItem(STORAGE_KEY_TOKEN)
    localStorage.removeItem(STORAGE_KEY_USER_ID)
    
    this.notifySubscribers({ type: 'logout' })
    
    return { success: true }
  }

  /**
   * 从本地存储恢复会话
   */
  restoreSession() {
    try {
      const token = localStorage.getItem(STORAGE_KEY_TOKEN)
      const userId = localStorage.getItem(STORAGE_KEY_USER_ID)
      
      if (token && userId) {
        // 验证token有效性
        const payload = this.parseToken(token)
        if (payload && payload.exp && payload.exp > Date.now()) {
          const user = getUserById(parseInt(userId))
          if (user) {
            this.currentUser = user
            this.token = token
            return true
          }
        }
      }
    } catch (error) {
      console.warn('[MockAuth] 恢复会话失败:', error.message)
    }
    
    // 清理无效的存储
    localStorage.removeItem(STORAGE_KEY_TOKEN)
    localStorage.removeItem(STORAGE_KEY_USER_ID)
    return false
  }

  /**
   * 安全的 Base64 编码（支持 Unicode）
   */
  _btoaUnicode(str) {
    return btoa(encodeURIComponent(str).replace(/%([0-9A-F]{2})/g, (_, p1) => String.fromCharCode('0x' + p1)))
  }

  /**
   * 生成模拟JWT Token
   * 格式: header.payload.signature
   */
  generateToken(user) {
    const header = this._btoaUnicode(JSON.stringify({ alg: 'mock', typ: 'JWT' }))
    const payload = this._btoaUnicode(JSON.stringify({
      sub: user.id,
      name: user.name,
      role: user.role,
      department: user.department,
      dept_role: user.dept_role,
      iat: Date.now(),
      exp: Date.now() + 24 * 60 * 60 * 1000  // 24小时过期
    }))
    const signature = this._btoaUnicode('mock-signature-' + Date.now())
    
    return `${header}.${payload}.${signature}`
  }

  /**
   * 解析Token
   */
  parseToken(token) {
    try {
      const parts = token.split('.')
      if (parts.length !== 3) {
        throw new Error('Invalid token format')
      }
      
      const payload = JSON.parse(atob(parts[1]))
      return payload
    } catch (error) {
      throw new Error('Failed to parse token')
    }
  }

  /**
   * 获取当前Token
   */
  getToken() {
    return this.token
  }

  /**
   * 检查是否已登录
   */
  isAuthenticated() {
    return !!this.currentUser && !!this.token
  }

  /**
   * 获取预设用户列表（用于角色切换）
   */
  getPresetUsers() {
    return PRESET_USERS.map(u => getSanitizedUser(u))
  }

  /**
   * 订阅认证状态变化
   */
  subscribe(callback) {
    this.subscribers.push(callback)
    return () => {
      const index = this.subscribers.indexOf(callback)
      if (index > -1) {
        this.subscribers.splice(index, 1)
      }
    }
  }

  /**
   * 通知所有订阅者
   */
  notifySubscribers(event) {
    this.subscribers.forEach(callback => {
      try {
        callback(event)
      } catch (error) {
        console.error('[MockAuth] 订阅者回调错误:', error)
      }
    })
  }

  /**
   * 模拟注册（仅演示，实际添加到本地存储）
   */
  async register(data) {
    await this.delay(500)
    
    // 检查学号是否已存在
    const existing = getUserByStudentId(data.student_id)
    if (existing) {
      throw new Error('该学号已注册')
    }
    
    // 模拟创建用户成功
    return {
      success: true,
      message: '注册成功，请登录'
    }
  }

  /**
   * 模拟密码重置请求
   */
  async forgotPassword(data) {
    await this.delay(500)
    
    const user = getUserByStudentId(data.student_id)
    if (!user) {
      throw new Error('该学号未注册')
    }
    
    return {
      success: true,
      message: '重置链接已发送到您的邮箱'
    }
  }

  /**
   * 模拟密码重置
   */
  async resetPassword(data) {
    await this.delay(400)
    
    // 模拟重置成功
    return {
      success: true,
      message: '密码重置成功，请使用新密码登录'
    }
  }

  /**
   * 模拟修改密码
   */
  async changePassword(data) {
    await this.delay(300)
    
    if (!this.currentUser) {
      throw new Error('未登录')
    }
    
    if (this.currentUser.password !== data.old_password) {
      throw new Error('原密码错误')
    }
    
    return {
      success: true,
      message: '密码修改成功'
    }
  }
}

// 导出单例
export const mockAuth = new MockAuthService()
export default mockAuth
