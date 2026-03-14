/**
 * 系统设置 Mock Handler
 */

import { repositories } from '../database/index.js'
import { success, error, notFound } from '../utils/response.js'
import { delay } from '../utils/delay.js'

// 系统设置缓存
let systemSettings = {
  installed: true,
  db_configured: true,
  smtp_configured: false
}

// SMTP配置缓存
let smtpConfig = {
  host: '',
  port: 587,
  username: '',
  password: '',
  sender: '',
  use_tls: true
}

// 站点配置缓存
let siteConfig = {
  site_url: 'http://localhost:8080',
  site_name: '排班系统'
}

// 部门列表
const departments = [
  { id: 1, name: '办公室', code: 'office' },
  { id: 2, name: '竞赛部', code: 'competition' },
  { id: 3, name: '宣传部', code: 'publicity' },
  { id: 4, name: '技术部', code: 'tech' },
  { id: 5, name: '外联部', code: 'liaison' }
]

// 权限列表
const permissions = [
  { code: 'schedule:manage:all', name: '排班管理（全局）', description: '管理所有部门排班' },
  { code: 'schedule:manage:dept', name: '排班管理（部门）', description: '管理部门内排班' },
  { code: 'schedule:view:all', name: '查看排班（全局）', description: '查看所有部门排班' },
  { code: 'user:manage:all', name: '用户管理（全局）', description: '管理所有用户' },
  { code: 'user:manage:dept', name: '用户管理（部门）', description: '管理部门内用户' },
  { code: 'system:config', name: '系统配置', description: '配置系统参数' }
]

// 临时权限缓存
let tempPermissions = []
let tempPermissionIdCounter = 1

export const systemHandlers = {
  // ========== 系统安装相关 ==========
  
  // 获取安装状态
  async getInstallStatus() {
    await delay(100)
    return success({
      installed: systemSettings.installed,
      db_configured: systemSettings.db_configured
    })
  },

  // 测试数据库连接
  async testDBConnection(data) {
    await delay(800)
    
    // 模拟连接测试
    const { host, port, username } = data
    
    if (!host || !port || !username) {
      return error('数据库配置信息不完整')
    }
    
    // 模拟成功率90%
    if (Math.random() > 0.1) {
      return success({ 
        connected: true, 
        message: '连接成功',
        server_version: 'MySQL 8.0.32'
      })
    } else {
      return error('连接失败：无法连接到数据库服务器')
    }
  },

  // 检查数据库
  async checkDatabase(data) {
    await delay(500)
    
    return success({
      exists: false,
      has_tables: false,
      message: '数据库为空，需要初始化'
    })
  },

  // 初始化数据库表
  async initDatabaseTables(data) {
    await delay(1500)
    
    systemSettings.db_configured = true
    return success({ 
      initialized: true,
      message: '数据库表初始化成功'
    })
  },

  // 创建管理员账号
  async createAdmin(data) {
    await delay(500)
    
    try {
      const { student_id, name, password, email } = data
      
      const admin = await repositories.users.create({
        student_id,
        name,
        password,
        email,
        role: 'admin',
        department: '办公室',
        dept_role: 'dept_admin'
      })
      
      systemSettings.installed = true
      
      const { password: _, ...safeAdmin } = admin
      return success(safeAdmin)
    } catch (err) {
      return error(err.message)
    }
  },

  // ========== SMTP配置 ==========
  
  // 获取SMTP配置
  async getSMTPConfig() {
    await delay(100)
    
    // 返回配置但隐藏密码
    return success({
      host: smtpConfig.host,
      port: smtpConfig.port,
      username: smtpConfig.username,
      password: smtpConfig.password ? '********' : '',
      sender: smtpConfig.sender,
      use_tls: smtpConfig.use_tls,
      configured: systemSettings.smtp_configured
    })
  },

  // 保存SMTP配置
  async saveSMTPConfig(data) {
    await delay(300)
    
    smtpConfig = { ...smtpConfig, ...data }
    systemSettings.smtp_configured = true
    
    return success({
      saved: true,
      message: 'SMTP配置保存成功'
    })
  },

  // 测试SMTP配置
  async testSMTPConfig(data) {
    await delay(1500)
    
    // 模拟测试邮件发送
    const { test_email } = data
    
    if (!test_email) {
      return error('请提供测试邮箱地址')
    }
    
    // 模拟90%成功率
    if (Math.random() > 0.1) {
      return success({
        sent: true,
        message: `测试邮件已发送至 ${test_email}`
      })
    } else {
      return error('邮件发送失败：SMTP服务器连接超时')
    }
  },

  // ========== 站点配置 ==========
  
  // 获取站点配置
  async getSiteConfig() {
    await delay(100)
    return success(siteConfig)
  },

  // 保存站点配置
  async saveSiteConfig(data) {
    await delay(200)
    siteConfig = { ...siteConfig, ...data }
    return success(siteConfig)
  },

  // ========== 临时权限 ==========
  
  // 获取临时权限列表
  async getTempPermissions() {
    await delay(150)
    
    // 添加用户信息和权限名称
    const enriched = tempPermissions.map(tp => {
      const permission = permissions.find(p => p.code === tp.permission)
      return {
        ...tp,
        permission_name: permission?.name || tp.permission,
        days_left: calculateDaysLeft(tp.expires_at)
      }
    })
    
    return success(enriched)
  },

  // 授予临时权限
  async grantTempPermission(data) {
    await delay(300)
    
    const { user_id, permission, expires_at, reason } = data
    
    const newPermission = {
      id: tempPermissionIdCounter++,
      user_id,
      permission,
      expires_at,
      reason,
      granted_by: 'admin', // 应该从当前用户获取
      granted_at: new Date().toISOString(),
      is_active: true
    }
    
    tempPermissions.push(newPermission)
    
    return success(newPermission)
  },

  // 撤销临时权限
  async revokeTempPermission(id) {
    await delay(200)
    
    const index = tempPermissions.findIndex(tp => tp.id === parseInt(id))
    if (index === -1) {
      return notFound('权限记录不存在')
    }
    
    tempPermissions.splice(index, 1)
    return success({ success: true })
  },

  // 获取我的临时权限
  async getMyTempPermissions(user) {
    await delay(150)
    
    const myPermissions = tempPermissions.filter(tp => 
      tp.user_id === user.id && 
      tp.is_active &&
      new Date(tp.expires_at) > new Date()
    )
    
    const enriched = myPermissions.map(tp => {
      const permission = permissions.find(p => p.code === tp.permission)
      return {
        ...tp,
        permission_name: permission?.name || tp.permission,
        days_left: calculateDaysLeft(tp.expires_at)
      }
    })
    
    return success(enriched)
  },

  // 清理过期权限
  async cleanupExpiredPermissions() {
    await delay(500)
    
    const now = new Date()
    const beforeCount = tempPermissions.length
    
    tempPermissions = tempPermissions.filter(tp => 
      new Date(tp.expires_at) > now && tp.is_active
    )
    
    const cleaned = beforeCount - tempPermissions.length
    
    return success({
      cleaned,
      message: `已清理 ${cleaned} 条过期权限`
    })
  },

  // ========== 基础数据 ==========
  
  // 获取权限列表
  async getPermissionList() {
    await delay(100)
    return success(permissions)
  },

  // 获取部门列表
  async getDepartments() {
    await delay(100)
    return success(departments)
  }
}

// 辅助函数：计算剩余天数
function calculateDaysLeft(expiresAt) {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = Math.ceil((expires - now) / (1000 * 60 * 60 * 24))
  return Math.max(0, diff)
}

export default systemHandlers
