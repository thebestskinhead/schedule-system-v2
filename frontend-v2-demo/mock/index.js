/**
 * Mock API 入口文件
 * 统一管理和导出所有Mock服务
 */

import mockAuth from './auth/MockAuthService.js'
import { db, repositories, initDatabase, resetDatabase } from './database/index.js'
import { seeder } from './seeders/DatabaseSeeder.js'
import { authHandlers } from './handlers/auth.handlers.js'
import { userHandlers } from './handlers/user.handlers.js'
import { scheduleHandlers } from './handlers/schedule.handlers.js'
import { availabilityHandlers } from './handlers/availability.handlers.js'
import { systemHandlers } from './handlers/system.handlers.js'
import { applicationHandlers } from './handlers/application.handlers.js'
import { crawlerHandlers } from './handlers/crawler.handlers.js'

// Mock模式标志
const isMockMode = import.meta.env.VITE_MOCK === 'true' || true

/**
 * Mock路由映射表
 */
export const mockRoutes = {
  // ========== 认证相关 ==========
  'POST /user/login': async (data, config) => {
    return authHandlers.login(data)
  },

  'GET /user/profile': async (data, config) => {
    return authHandlers.getProfile()
  },

  'POST /user/register': async (data, config) => {
    return authHandlers.register(data)
  },

  'POST /password/reset-request': async (data, config) => {
    return authHandlers.forgotPassword(data)
  },

  'POST /password/reset': async (data, config) => {
    return authHandlers.resetPassword(data)
  },

  'POST /user/change-password': async (data, config) => {
    return authHandlers.changePassword(data)
  },

  // ========== 用户管理 ==========
  'GET /admin/users': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return userHandlers.getUserList(data, user)
  },

  'GET /admin/users/by-dept': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return userHandlers.getUsersByDepartment(data.department, user)
  },

  'GET /users/for-schedule': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return userHandlers.getUsersForSchedule(user)
  },

  'POST /admin/users': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return userHandlers.createUser(data, user)
  },

  'PUT /admin/users/:id': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    const id = config.urlParams?.id
    return userHandlers.updateUser(id, data, user)
  },

  'DELETE /admin/users/:id': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    const id = config.urlParams?.id
    return userHandlers.deleteUser(id, user)
  },

  'PUT /admin/users/:id/role': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    const id = config.urlParams?.id
    return userHandlers.updateUserRole(id, data, user)
  },

  'PUT /admin/users/:id/department': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    const id = config.urlParams?.id
    return userHandlers.updateUserDepartment(id, data, user)
  },

  // ========== 排班管理 ==========
  'GET /schedule/current-week': async (data, config) => {
    return scheduleHandlers.getCurrentWeek()
  },

  'POST /admin/schedule/current-week': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return scheduleHandlers.updateCurrentWeek(data)
  },

  'GET /schedule': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return scheduleHandlers.getSchedule(data)
  },

  'POST /admin/schedule/preview': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return scheduleHandlers.previewSchedule(data, user)
  },

  'POST /admin/schedule/confirm': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return scheduleHandlers.confirmSchedule(data, user)
  },

  'GET /admin/schedule/settings': async (data, config) => {
    return scheduleHandlers.getScheduleSettings()
  },

  'POST /admin/schedule/settings': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return scheduleHandlers.saveScheduleSettings(data)
  },

  'POST /admin/schedule/update': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return scheduleHandlers.updateSchedule(data)
  },

  'GET /duty/my': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return scheduleHandlers.getMyDuties(user)
  },

  'PUT /duty/status': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return scheduleHandlers.updateDutyStatus(data, user)
  },

  'GET /admin/templates': async (data, config) => {
    return scheduleHandlers.getTemplates()
  },

  'POST /admin/templates': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return scheduleHandlers.createTemplate(data)
  },

  'PUT /admin/templates': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return scheduleHandlers.updateTemplate(data)
  },

  'DELETE /admin/templates/:id': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    const id = config.urlParams?.id
    return scheduleHandlers.deleteTemplate(id)
  },

  'GET /admin/duty-assignments': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return scheduleHandlers.getDutyAssignments(data)
  },

  'POST /admin/duty-assignments': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return scheduleHandlers.saveDutyAssignments(data)
  },

  'GET /duty-assignments/my-dept': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return scheduleHandlers.getMyDeptAssignment(data, user)
  },

  'GET /admin/schedule/semester-start': async (data, config) => {
    return scheduleHandlers.getSemesterStartDate()
  },

  'POST /admin/schedule/semester-start': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return scheduleHandlers.updateSemesterStartDate(data)
  },

  // ========== 无课表管理 ==========
  'GET /availability': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return availabilityHandlers.getMyAvailability(user, data)
  },

  'POST /availability': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return availabilityHandlers.addAvailability(data, user)
  },

  'DELETE /availability': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return availabilityHandlers.deleteAvailability(data, user)
  },

  'POST /availability/import/cookie': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return availabilityHandlers.importFromCookie(data, user)
  },

  'POST /availability/import/xls': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return availabilityHandlers.importFromXLS(data, user)
  },

  'GET /availability/import/status': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return availabilityHandlers.getImportStatus(data)
  },

  'GET /availability/import/tasks': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return availabilityHandlers.getImportTaskList(user)
  },

  // ========== 爬虫导入 ==========
  'POST /crawler/import': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return crawlerHandlers.importCrawler(data, user)
  },

  'POST /crawler/preview': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return crawlerHandlers.previewCrawler(data, user)
  },

  // ========== 系统设置 ==========
  'GET /system/installed': async (data, config) => {
    return systemHandlers.getInstallStatus()
  },

  'POST /system/test-db': async (data, config) => {
    return systemHandlers.testDBConnection(data)
  },

  'POST /system/check-db': async (data, config) => {
    return systemHandlers.checkDatabase(data)
  },

  'POST /system/init-tables': async (data, config) => {
    return systemHandlers.initDatabaseTables(data)
  },

  'POST /system/create-admin': async (data, config) => {
    return systemHandlers.createAdmin(data)
  },

  'GET /admin/smtp/config': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return systemHandlers.getSMTPConfig()
  },

  'POST /admin/smtp/config': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return systemHandlers.saveSMTPConfig(data)
  },

  'POST /admin/smtp/test': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return systemHandlers.testSMTPConfig(data)
  },

  'GET /admin/site/config': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return systemHandlers.getSiteConfig()
  },

  'POST /admin/site/config': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return systemHandlers.saveSiteConfig(data)
  },

  'GET /admin/temp-permissions': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return systemHandlers.getTempPermissions()
  },

  'POST /admin/temp-permissions': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return systemHandlers.grantTempPermission(data)
  },

  'DELETE /admin/temp-permissions/:id': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    const id = config.urlParams?.id
    return systemHandlers.revokeTempPermission(id)
  },

  'GET /temp-permissions/my': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return systemHandlers.getMyTempPermissions(user)
  },

  'POST /admin/temp-permissions/cleanup': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return systemHandlers.cleanupExpiredPermissions()
  },

  'GET /permissions/list': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return systemHandlers.getPermissionList()
  },

  'GET /departments': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return systemHandlers.getDepartments()
  },

  // ========== 申请审批 ==========
  'GET /application/types': async (data, config) => {
    return applicationHandlers.getApplicationTypes()
  },

  'GET /application/permissions/available': async (data, config) => {
    return applicationHandlers.getAvailablePermissions()
  },

  'GET /applications/my': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return applicationHandlers.getMyApplications(data, user)
  },

  'POST /applications': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return applicationHandlers.createApplication(data, user)
  },

  'GET /applications/:id': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return applicationHandlers.getApplicationDetail(config.urlParams, user)
  },

  'POST /applications/:id/cancel': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return applicationHandlers.cancelApplication(config.urlParams, user)
  },

  'GET /applications/pending': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return applicationHandlers.getPendingApprovals(data, user)
  },

  'POST /applications/:id/approve': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return applicationHandlers.processApproval(config.urlParams, data, user)
  },

  'GET /applications/stats': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    return applicationHandlers.getApplicationStats(user)
  },

  // ========== 演示数据生成 ==========
  'POST /demo/seed-schedules': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    
    if (user.role !== 'admin') {
      return { code: 403, message: '只有管理员可以重新生成排班数据', data: null }
    }
    
    console.log('[Demo] 重新生成排班数据...')
    
    try {
      // 重新运行排班数据填充
      await seeder.seedSchedules()
      
      const stats = await seeder.getStats()
      
      return {
        code: 200,
        message: '排班数据重新生成成功',
        data: {
          schedules: stats.schedules,
          duties: stats.duties,
          weekStats: stats.weekStats
        }
      }
    } catch (err) {
      return { code: 500, message: err.message, data: null }
    }
  },

  'POST /demo/clear-schedules': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    
    if (user.role !== 'admin') {
      return { code: 403, message: '只有管理员可以清空排班数据', data: null }
    }
    
    console.log('[Demo] 清空排班数据...')
    
    // 清空排班数据
    for (let week = 1; week <= 30; week++) {
      try {
        const schedules = await repositories.schedules.findByWeek(week)
        for (const s of schedules) {
          await repositories.schedules.delete(s.id, week)
        }
      } catch (err) {}
    }
    
    return {
      code: 200,
      message: '排班数据已清空',
      data: { success: true }
    }
  },

  'GET /demo/stats': async (data, config) => {
    const user = await mockAuth.getCurrentUser()
    const stats = await seeder.getStats()
    
    return {
      code: 200,
      message: 'success',
      data: stats
    }
  }
}

/**
 * Mock API 统一入口
 */
export const mockAPI = {
  // 认证服务
  auth: mockAuth,
  
  // 数据库
  db,
  repositories,
  
  // 路由
  routes: mockRoutes,
  
  // 检查是否处于Mock模式
  get isMockMode() {
    return isMockMode
  },

  /**
   * 初始化Mock服务
   */
  async init() {
    if (!isMockMode) {
      console.log('[Mock] Mock模式已禁用')
      return false
    }
    
    console.log('[Mock] ================================')
    console.log('[Mock] Mock服务初始化中...')
    console.log('[Mock] ================================')
    
    // 初始化数据库
    const dbInit = await initDatabase()
    if (!dbInit.success) {
      console.error('[Mock] 数据库初始化失败:', dbInit.error)
      return false
    }
    
    // 填充演示数据
    const seedResult = await seeder.run()
    if (!seedResult.success) {
      console.error('[Mock] 数据填充失败:', seedResult.error)
      // 不阻断初始化，继续运行
    }

    console.log('[Mock] 可用预设用户:')
    console.log('[Mock] ------------------------')
    mockAuth.getPresetUsers().forEach(u => {
      const roleLabel = u.role === 'admin' ? '系统管理员' :
                       u.department === '办公室' && u.dept_role === 'dept_admin' ? '办公室管理' :
                       u.dept_role === 'dept_admin' ? '部门管理' : '普通成员'
      console.log(`  👤 ${u.name} (${u.student_id})`)
      console.log(`     角色: ${roleLabel}`)
      console.log(`     密码: 123456`)
    })
    console.log('[Mock] ------------------------')
    console.log('[Mock] 初始化完成！')
    console.log('[Mock] ================================')
    
    return true
  },

  /**
   * 重置所有Mock数据
   */
  async reset() {
    if (!isMockMode) return
    
    // 清除认证信息
    await mockAuth.logout()
    
    // 重置数据库
    await resetDatabase()
    
    console.log('[Mock] 所有Mock数据已重置')
  },

  /**
   * 处理Mock请求
   */
  async handle(config) {
    const method = config.method.toUpperCase()
    
    // 处理完整 URL（如 http://localhost:8081/api/v1/xxx）
    let url = config.url
    if (url.includes('/api/v1')) {
      url = url.split('/api/v1')[1] || url
    }
    // 确保以 / 开头
    if (!url.startsWith('/')) {
      url = '/' + url
    }
    
    // 移除查询参数用于路由匹配
    const urlWithoutQuery = url.split('?')[0]
    
    // 构建路由键
    const routeKey = `${method} ${urlWithoutQuery}`
    
    // 调试日志
    console.log('[Mock] 原始URL:', config.url)
    console.log('[Mock] 处理后:', urlWithoutQuery)
    console.log('[Mock] 尝试匹配:', routeKey)
    
    // 尝试直接匹配
    let handler = mockRoutes[routeKey]
    
    if (!handler) {
      console.log('[Mock] 未找到处理器，尝试模式匹配...')
    }
    
    // 如果没有直接匹配，尝试模式匹配（如 /admin/users/123）
    if (!handler) {
      const patterns = Object.keys(mockRoutes).filter(key => key.includes(':'))
      for (const pattern of patterns) {
        const regex = new RegExp('^' + pattern.replace(/:[^/]+/g, '([^/]+)') + '$')
        const match = routeKey.match(regex)
        if (match) {
          handler = mockRoutes[pattern]
          // 提取URL参数
          const paramNames = pattern.match(/:[^/]+/g) || []
          config.urlParams = {}
          paramNames.forEach((name, i) => {
            config.urlParams[name.slice(1)] = match[i + 1]
          })
          break
        }
      }
    }
    
    if (handler) {
      console.log('[Mock] ✓ 匹配成功:', routeKey)
      
      // 解析查询参数
      const queryParams = {}
      const queryIndex = url.indexOf('?')
      if (queryIndex !== -1) {
        const searchParams = new URLSearchParams(url.slice(queryIndex))
        searchParams.forEach((value, key) => {
          queryParams[key] = value
        })
      }
      
      // 合并 body 数据和 query 参数
      const data = { ...config.data, ...queryParams }
      
      return handler(data, config)
    }
    
    console.warn('[Mock] ✗ 未找到处理器:', routeKey)
    
    // 未找到处理程序
    console.warn(`[Mock] 未找到路由处理器: ${routeKey}`)
    return {
      code: 404,
      message: 'API未实现（Mock模式）',
      data: null
    }
  }
}

export { mockAuth, db, repositories }
export default mockAPI
