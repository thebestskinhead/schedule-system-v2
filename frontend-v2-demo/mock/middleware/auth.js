/**
 * 认证中间件
 */

import { mockAuth } from '../auth/MockAuthService.js'
import { unauthorized, forbidden } from '../utils/response.js'

/**
 * 从请求中获取Token
 */
export function getTokenFromRequest(config) {
  if (!config || !config.headers) return null
  
  const authHeader = config.headers.Authorization || config.headers.authorization
  if (!authHeader) return null
  
  return authHeader.replace('Bearer ', '')
}

/**
 * 获取当前用户
 */
export async function getCurrentUser(config) {
  const token = getTokenFromRequest(config)
  if (!token) return null
  
  try {
    return await mockAuth.checkToken(token)
  } catch {
    return null
  }
}

/**
 * 需要登录的中间件
 */
export function requireAuth(handler) {
  return async (data, config) => {
    const user = await getCurrentUser(config)
    
    if (!user) {
      return unauthorized()
    }
    
    return handler(data, user, config)
  }
}

/**
 * 需要权限的中间件
 */
export function requirePermission(permission, handler) {
  return async (data, config) => {
    const user = await getCurrentUser(config)
    
    if (!user) {
      return unauthorized()
    }
    
    // 系统管理员拥有所有权限
    if (user.role === 'admin') {
      return handler(data, user, config)
    }
    
    // 办公室管理员拥有大部分管理权限
    if (user.department === '办公室' && user.dept_role === 'dept_admin') {
      const officeAdminPermissions = [
        'user:manage:all',
        'schedule:manage:all',
        'schedule:view:all'
      ]
      if (officeAdminPermissions.includes(permission) || 
          officeAdminPermissions.some(p => permission.startsWith(p.replace(':all', '')))) {
        return handler(data, user, config)
      }
    }
    
    // 部门管理员
    if (user.dept_role === 'dept_admin') {
      const deptAdminPermissions = [
        'user:manage:dept',
        'schedule:manage:dept'
      ]
      if (deptAdminPermissions.includes(permission) ||
          deptAdminPermissions.some(p => permission.startsWith(p.replace(':dept', '')))) {
        return handler(data, user, config)
      }
    }
    
    // TODO: 检查临时权限
    
    return forbidden(`需要权限: ${permission}`)
  }
}

/**
 * 检查是否有指定权限
 */
export function hasPermission(user, permission) {
  if (!user) return false
  
  // 系统管理员
  if (user.role === 'admin') return true
  
  // 办公室管理员
  if (user.department === '办公室' && user.dept_role === 'dept_admin') {
    return true
  }
  
  // 部门管理员
  if (user.dept_role === 'dept_admin') {
    const deptPermissions = ['schedule:manage:dept', 'user:manage:dept']
    if (deptPermissions.includes(permission)) return true
  }
  
  return false
}

export default {
  getTokenFromRequest,
  getCurrentUser,
  requireAuth,
  requirePermission,
  hasPermission
}
