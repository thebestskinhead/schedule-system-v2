/**
 * 申请审批 Mock Handler
 */

import { repositories } from '../database/index.js'
import { success, error, notFound } from '../utils/response.js'
import { delay } from '../utils/delay.js'

// 申请类型
const applicationTypes = [
  { code: 'temp_permission', name: '临时权限申请', description: '申请临时排班管理权限' },
  { code: 'dept_transfer', name: '部门调动申请', description: '申请调动到其他部门' },
  { code: 'role_change', name: '角色变更申请', description: '申请变更部门角色' }
]

// 可申请权限
const availablePermissions = [
  { code: 'schedule:manage:dept', name: '排班管理（部门）', description: '临时管理部门排班', max_days: 30 },
  { code: 'schedule:view:all', name: '查看排班（全局）', description: '临时查看所有部门排班', max_days: 7 },
  { code: 'user:manage:dept', name: '用户管理（部门）', description: '临时管理部门用户', max_days: 14 }
]

// 申请数据缓存
let applications = []
let applicationIdCounter = 1

export const applicationHandlers = {
  // 获取申请类型
  async getApplicationTypes() {
    await delay(100)
    return success(applicationTypes)
  },

  // 获取可申请的权限列表
  async getAvailablePermissions() {
    await delay(100)
    return success(availablePermissions)
  },

  // 获取我的申请
  async getMyApplications(query, user) {
    await delay(200)
    
    const { page = 1, page_size = 10, status } = query
    
    let myApps = applications.filter(a => a.applicant_id === user.id)
    
    if (status) {
      myApps = myApps.filter(a => a.status === status)
    }
    
    // 按时间倒序
    myApps.sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
    
    const total = myApps.length
    const start = (page - 1) * page_size
    const end = start + page_size
    const list = myApps.slice(start, end)
    
    // 丰富信息
    const enriched = list.map(app => ({
      ...app,
      type_name: applicationTypes.find(t => t.code === app.type_code)?.name || app.type_code,
      status_text: getStatusText(app.status)
    }))
    
    return success({
      list: enriched,
      total,
      page: parseInt(page),
      page_size: parseInt(page_size)
    })
  },

  // 创建申请
  async createApplication(data, user) {
    await delay(400)
    
    const { type_code, content, target_permission, target_department, duration_days } = data
    
    // 生成申请编号
    const date = new Date()
    const dateStr = date.toISOString().slice(0, 10).replace(/-/g, '')
    const seq = String(applicationIdCounter).padStart(4, '0')
    const applicationNo = `APP${dateStr}${seq}`
    
    const newApp = {
      id: applicationIdCounter++,
      application_no: applicationNo,
      applicant_id: user.id,
      applicant_name: user.name,
      applicant_student_id: user.student_id,
      department: user.department,
      type_code,
      content,
      target_permission,
      target_department,
      duration_days,
      status: 'pending',
      created_at: date.toISOString(),
      updated_at: date.toISOString()
    }
    
    applications.push(newApp)
    
    return success({
      id: newApp.id,
      application_no: newApp.application_no,
      status: newApp.status
    })
  },

  // 获取申请详情
  async getApplicationDetail(params, user) {
    await delay(150)
    
    const { id } = params
    const app = applications.find(a => a.id === parseInt(id))
    
    if (!app) {
      return notFound('申请不存在')
    }
    
    // 检查权限（只能看自己的或待自己审批的）
    if (app.applicant_id !== user.id && !canApprove(user)) {
      return error('无权查看此申请')
    }
    
    return success({
      ...app,
      type_name: applicationTypes.find(t => t.code === app.type_code)?.name || app.type_code,
      status_text: getStatusText(app.status)
    })
  },

  // 取消申请
  async cancelApplication(params, user) {
    await delay(200)
    
    const { id } = params
    const app = applications.find(a => a.id === parseInt(id))
    
    if (!app) {
      return notFound('申请不存在')
    }
    
    if (app.applicant_id !== user.id) {
      return error('只能取消自己的申请')
    }
    
    if (app.status !== 'pending') {
      return error('只能取消待审批的申请')
    }
    
    app.status = 'cancelled'
    app.updated_at = new Date().toISOString()
    
    return success({ success: true })
  },

  // 获取待审批列表
  async getPendingApprovals(query, user) {
    await delay(200)
    
    if (!canApprove(user)) {
      return error('无审批权限')
    }
    
    const { page = 1, page_size = 10 } = query
    
    // 获取待审批的申请
    let pending = applications.filter(a => a.status === 'pending')
    
    // 部门管理员只能看到本部门的申请
    if (user.dept_role === 'dept_admin' && user.department !== '办公室') {
      pending = pending.filter(a => a.department === user.department)
    }
    
    // 按时间正序（先申请的在前）
    pending.sort((a, b) => new Date(a.created_at) - new Date(b.created_at))
    
    const total = pending.length
    const start = (page - 1) * page_size
    const end = start + page_size
    const list = pending.slice(start, end)
    
    // 丰富信息
    const enriched = list.map(app => ({
      ...app,
      type_name: applicationTypes.find(t => t.code === app.type_code)?.name || app.type_code
    }))
    
    return success({
      list: enriched,
      total,
      page: parseInt(page),
      page_size: parseInt(page_size)
    })
  },

  // 处理审批
  async processApproval(params, data, user) {
    await delay(300)
    
    if (!canApprove(user)) {
      return error('无审批权限')
    }
    
    const { id } = params
    const { action, comment } = data
    
    const app = applications.find(a => a.id === parseInt(id))
    
    if (!app) {
      return notFound('申请不存在')
    }
    
    if (app.status !== 'pending') {
      return error('该申请已被处理')
    }
    
    // 部门管理员只能审批本部门
    if (user.dept_role === 'dept_admin' && user.department !== '办公室') {
      if (app.department !== user.department) {
        return error('无权审批其他部门的申请')
      }
    }
    
    const now = new Date().toISOString()
    
    if (action === 'approve') {
      app.status = 'approved'
      
      // 如果是权限申请，自动授予临时权限
      if (app.type_code === 'temp_permission' && app.target_permission) {
        await grantTempPermission(app)
      }
    } else if (action === 'reject') {
      app.status = 'rejected'
    } else {
      return error('无效的审批操作')
    }
    
    app.approved_by = user.id
    app.approved_by_name = user.name
    app.approved_at = now
    app.comment = comment
    app.updated_at = now
    
    return success({
      id: app.id,
      status: app.status,
      approved_at: now
    })
  },

  // 获取申请统计
  async getApplicationStats(user) {
    await delay(150)
    
    const myApps = applications.filter(a => a.applicant_id === user.id)
    
    const stats = {
      total: myApps.length,
      pending: myApps.filter(a => a.status === 'pending').length,
      approved: myApps.filter(a => a.status === 'approved').length,
      rejected: myApps.filter(a => a.status === 'rejected').length,
      cancelled: myApps.filter(a => a.status === 'cancelled').length
    }
    
    // 如果有审批权限，添加待审批统计
    if (canApprove(user)) {
      let pending = applications.filter(a => a.status === 'pending')
      if (user.dept_role === 'dept_admin' && user.department !== '办公室') {
        pending = pending.filter(a => a.department === user.department)
      }
      stats.pending_approval = pending.length
    }
    
    return success(stats)
  }
}

// 辅助函数：检查是否有审批权限
function canApprove(user) {
  return user.role === 'admin' || 
         user.dept_role === 'dept_admin' || 
         user.department === '办公室'
}

// 辅助函数：获取状态文本
function getStatusText(status) {
  const map = {
    pending: '待审批',
    approved: '已通过',
    rejected: '已拒绝',
    cancelled: '已取消'
  }
  return map[status] || status
}

// 辅助函数：授予临时权限
async function grantTempPermission(application) {
  // 计算过期时间
  const expiresAt = new Date()
  expiresAt.setDate(expiresAt.getDate() + (application.duration_days || 7))
  
  // 这里应该调用systemHandlers来授予权限
  // 简化处理：直接记录到application中
  application.granted_permission = {
    permission: application.target_permission,
    expires_at: expiresAt.toISOString()
  }
}

export default applicationHandlers
