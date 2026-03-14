/**
 * 认证相关 Mock Handler
 */

import { mockAuth } from '../auth/MockAuthService.js'
import { success, error, unauthorized } from '../utils/response.js'
import { delay } from '../utils/delay.js'

export const authHandlers = {
  // 登录
  async login(data) {
    await delay(400)
    
    try {
      const result = await mockAuth.login(data)
      return success(result)
    } catch (err) {
      return error(err.message)
    }
  },

  // 获取当前用户信息
  async getProfile() {
    await delay(100)
    
    try {
      const user = await mockAuth.getCurrentUser()
      return success(user)
    } catch (err) {
      return unauthorized(err.message)
    }
  },

  // 注册
  async register(data) {
    await delay(500)
    
    try {
      const result = await mockAuth.register(data)
      return success(result)
    } catch (err) {
      return error(err.message)
    }
  },

  // 忘记密码
  async forgotPassword(data) {
    await delay(500)
    
    try {
      const result = await mockAuth.forgotPassword(data)
      return success(result)
    } catch (err) {
      return error(err.message)
    }
  },

  // 重置密码
  async resetPassword(data) {
    await delay(400)
    
    try {
      const result = await mockAuth.resetPassword(data)
      return success(result)
    } catch (err) {
      return error(err.message)
    }
  },

  // 修改密码
  async changePassword(data) {
    await delay(300)
    
    try {
      const result = await mockAuth.changePassword(data)
      return success(result)
    } catch (err) {
      return error(err.message)
    }
  },

  // 登出
  async logout() {
    await delay(100)
    
    try {
      await mockAuth.logout()
      return success({ success: true })
    } catch (err) {
      return error(err.message)
    }
  }
}

export default authHandlers
