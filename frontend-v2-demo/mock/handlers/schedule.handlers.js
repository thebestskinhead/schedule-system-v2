/**
 * 排班管理 Mock Handler
 */

import { repositories } from '../database/index.js'
import { success, error, forbidden, notFound } from '../utils/response.js'
import { delay } from '../utils/delay.js'

// 内存中的设置缓存
let scheduleSettings = {
  current_week: 1,
  auto_increment: true,
  need_per_cell: 2,
  min_per_cell: 1,
  max_per_day: 2,
  max_per_week: 4,
  semester_start_date: new Date().toISOString().split('T')[0]
}

// 模板缓存
let templates = [
  { id: 1, name: '默认模板', need_per_cell: 2, description: '标准排班模板' }
]

// 分工配置缓存
let dutyAssignments = []

export const scheduleHandlers = {
  // 获取当前周次
  async getCurrentWeek() {
    await delay(100)
    return success({ current_week: scheduleSettings.current_week })
  },

  // 更新当前周次
  async updateCurrentWeek(data) {
    await delay(200)
    scheduleSettings.current_week = data.current_week || 1
    return success({ current_week: scheduleSettings.current_week })
  },

  // 获取排班表
  async getSchedule(query) {
    await delay(200)
    
    const { week, department } = query
    if (!week) {
      return error('请指定周次')
    }
    
    try {
      const schedules = await repositories.schedules.findByWeek(parseInt(week), department)
      return success(schedules)
    } catch (err) {
      return error(err.message)
    }
  },

  // 预览排班（简化版算法）
  async previewSchedule(data, user) {
    await delay(600)
    
    try {
      const { week, department, need_per_cell = 2 } = data
      
      // 获取部门用户
      let users
      if (department) {
        users = await repositories.users.findByDepartment(department)
      } else {
        users = await repositories.users.findAll()
      }
      
      // 过滤出有空的用户
      const userIds = users.map(u => u.id)
      
      // 生成排班预览（简化算法：每时段随机选择可用人员）
      const assignments = []
      const weekdays = [1, 2, 3, 4, 5] // 周一到周五
      const periods = [1, 2, 3, 4] // 1-4节
      
      for (const weekday of weekdays) {
        for (const period of periods) {
          // 获取该时段有空的用户
          const availableUsers = await repositories.availability.findAvailableUsers(
            week, weekday, period, userIds
          )
          
          // 随机选择
          const selected = availableUsers
            .sort(() => 0.5 - Math.random())
            .slice(0, need_per_cell)
          
          assignments.push({
            weekday,
            period,
            users: selected.map(u => ({
              id: u.user_id,
              name: users.find(user => user.id === u.user_id)?.name || '未知'
            })),
            status: 'preview'
          })
        }
      }
      
      return success({
        week,
        department,
        assignments,
        stats: {
          total_slots: weekdays.length * periods.length,
          filled_slots: assignments.reduce((sum, a) => sum + a.users.length, 0),
          need_per_cell
        }
      })
    } catch (err) {
      return error(err.message)
    }
  },

  // 确认排班
  async confirmSchedule(data, user) {
    await delay(500)
    
    try {
      const { week, assignments, department } = data
      
      // 先清空该周的排班
      await repositories.schedules.clearWeek(week, department)
      
      // 创建新的排班记录
      const schedules = []
      for (const assignment of assignments) {
        for (const u of assignment.users) {
          schedules.push({
            week,
            weekday: assignment.weekday,
            period: assignment.period,
            user_id: u.id,
            user_name: u.name,
            department: department || user.department,
            status: 'confirmed',
            created_at: new Date().toISOString()
          })
        }
      }
      
      if (schedules.length > 0) {
        await repositories.schedules.batchCreate(schedules)
      }
      
      return success({
        week,
        total_assigned: schedules.length,
        department
      })
    } catch (err) {
      return error(err.message)
    }
  },

  // 获取排班设置
  async getScheduleSettings() {
    await delay(100)
    return success(scheduleSettings)
  },

  // 保存排班设置
  async saveScheduleSettings(data) {
    await delay(200)
    scheduleSettings = { ...scheduleSettings, ...data }
    return success(scheduleSettings)
  },

  // 更新排班（单个修改）
  async updateSchedule(data) {
    await delay(300)
    
    try {
      const { id, ...updateData } = data
      const updated = await repositories.schedules.update(id, updateData)
      return success(updated)
    } catch (err) {
      return error(err.message)
    }
  },

  // 获取我的值班
  async getMyDuties(user) {
    await delay(200)
    
    try {
      const currentWeek = scheduleSettings.current_week
      const schedules = await repositories.schedules.findByUser(user.id, currentWeek)
      
      // 添加状态信息
      const duties = schedules.map(s => ({
        ...s,
        duty_status: s.duty_status || 'pending', // pending, completed, missed
        check_in_time: s.check_in_time || null
      }))
      
      return success(duties)
    } catch (err) {
      return error(err.message)
    }
  },

  // 更新值班状态
  async updateDutyStatus(data, user) {
    await delay(200)
    
    try {
      const { schedule_id, status } = data
      const schedules = await repositories.schedules.findByUser(user.id)
      const schedule = schedules.find(s => s.id === parseInt(schedule_id))
      
      if (!schedule) {
        return notFound('值班记录不存在')
      }
      
      const updated = await repositories.schedules.update(
        parseInt(schedule_id),
        {
          duty_status: status,
          check_in_time: status === 'completed' ? new Date().toISOString() : null
        },
        schedule.week
      )
      
      return success(updated)
    } catch (err) {
      return error(err.message)
    }
  },

  // 获取模板列表
  async getTemplates() {
    await delay(100)
    return success(templates)
  },

  // 创建模板
  async createTemplate(data) {
    await delay(200)
    
    const newTemplate = {
      id: templates.length + 1,
      ...data,
      created_at: new Date().toISOString()
    }
    templates.push(newTemplate)
    return success(newTemplate)
  },

  // 更新模板
  async updateTemplate(data) {
    await delay(200)
    
    const { id, ...updateData } = data
    const index = templates.findIndex(t => t.id === id)
    
    if (index === -1) {
      return notFound('模板不存在')
    }
    
    templates[index] = { ...templates[index], ...updateData }
    return success(templates[index])
  },

  // 删除模板
  async deleteTemplate(id) {
    await delay(200)
    
    const index = templates.findIndex(t => t.id === parseInt(id))
    if (index === -1) {
      return notFound('模板不存在')
    }
    
    templates.splice(index, 1)
    return success({ success: true })
  },

  // 获取分工配置
  async getDutyAssignments(query) {
    await delay(150)
    
    const { department } = query
    let result = dutyAssignments
    
    if (department) {
      result = dutyAssignments.filter(d => d.department === department)
    }
    
    return success(result)
  },

  // 保存分工配置
  async saveDutyAssignments(data) {
    await delay(300)
    
    const { department, assignments } = data
    
    // 删除旧配置
    dutyAssignments = dutyAssignments.filter(d => d.department !== department)
    
    // 添加新配置
    const newAssignments = assignments.map(a => ({
      ...a,
      department,
      id: dutyAssignments.length + 1,
      created_at: new Date().toISOString()
    }))
    
    dutyAssignments.push(...newAssignments)
    return success(newAssignments)
  },

  // 获取我的部门分工
  async getMyDeptAssignment(query, user) {
    await delay(150)
    
    const { week } = query
    const assignments = dutyAssignments.filter(d => 
      d.department === user.department && d.week === parseInt(week)
    )
    
    return success(assignments)
  },

  // 获取学期起始日
  async getSemesterStartDate() {
    await delay(100)
    return success({ 
      semester_start_date: scheduleSettings.semester_start_date 
    })
  },

  // 更新学期起始日
  async updateSemesterStartDate(data) {
    await delay(200)
    scheduleSettings.semester_start_date = data.semester_start_date
    return success({ 
      semester_start_date: scheduleSettings.semester_start_date 
    })
  }
}

export default scheduleHandlers
