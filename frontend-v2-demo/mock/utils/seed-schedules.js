/**
 * 示例排班数据生成工具
 * 生成更完整的排班和值班数据
 */

import { repositories, db } from '../database/index.js'
import { ScheduleGenerator } from '../generators/ScheduleGenerator.js'
import { PRESET_USERS } from '../auth/preset-users.js'

export class ScheduleSeeder {
  constructor() {
    this.departments = ['办公室', '竞赛部', '项目部', '科普部']
    this.weekdays = [1, 2, 3, 4, 5] // 周一到周五
    this.periods = [1, 2, 3, 4] // 1-4节
  }

  /**
   * 生成完整示例数据
   * @param {Object} options
   * @param {number} options.startWeek - 开始周次 (默认1)
   * @param {number} options.endWeek - 结束周次 (默认16)
   * @param {boolean} options.clearExisting - 是否清空现有数据
   */
  async generate(options = {}) {
    const {
      startWeek = 1,
      endWeek = 16,
      clearExisting = false
    } = options

    console.log(`[ScheduleSeeder] 开始生成排班数据: 第${startWeek}-${endWeek}周`)

    if (clearExisting) {
      console.log('[ScheduleSeeder] 清空现有排班数据...')
      await this.clearSchedules()
    }

    const results = {
      weeks: 0,
      departments: [],
      totalSchedules: 0,
      totalDuties: 0
    }

    // 为每个部门生成排班
    for (const dept of this.departments) {
      console.log(`[ScheduleSeeder] 生成 ${dept} 的排班数据...`)
      
      const deptResult = await this.generateForDepartment(dept, startWeek, endWeek)
      results.departments.push({
        name: dept,
        schedules: deptResult.schedules,
        duties: deptResult.duties
      })
      results.totalSchedules += deptResult.schedules
      results.totalDuties += deptResult.duties
    }

    results.weeks = endWeek - startWeek + 1
    
    console.log('[ScheduleSeeder] 排班数据生成完成!')
    console.log(`  - 覆盖周次: ${results.weeks} 周`)
    console.log(`  - 涉及部门: ${results.departments.map(d => d.name).join(', ')}`)
    console.log(`  - 排班记录: ${results.totalSchedules} 条`)
    console.log(`  - 值班记录: ${results.totalDuties} 条`)

    return results
  }

  /**
   * 为单个部门生成排班
   */
  async generateForDepartment(department, startWeek, endWeek) {
    // 获取部门用户
    const users = await repositories.users.findByDepartment(department)
    if (users.length === 0) {
      console.warn(`[ScheduleSeeder] ${department} 没有用户，跳过`)
      return { schedules: 0, duties: 0 }
    }

    let totalSchedules = 0
    let totalDuties = 0

    // 为每周生成排班
    for (let week = startWeek; week <= endWeek; week++) {
      // 获取用户无课表
      const userIds = users.map(u => u.id)
      const availability = []
      
      for (const userId of userIds) {
        const userAvail = await repositories.availability.findByUserAndWeek(userId, week)
        availability.push(...userAvail)
      }

      // 生成排班
      const scheduleResult = ScheduleGenerator.generateSchedule(
        { 
          week, 
          days: this.weekdays, 
          periods: 4, 
          needPerCell: 2 
        },
        users,
        availability
      )

      // 保存排班记录
      const records = this.convertToScheduleRecords(scheduleResult, department, users)
      
      if (records.length > 0) {
        await repositories.schedules.batchCreate(records)
        totalSchedules += records.length
      }

      // 生成值班记录（基于排班）
      const dutyRecords = this.generateDutyRecordsFromSchedule(records, week, department)
      
      if (dutyRecords.length > 0) {
        await this.saveDutyRecords(dutyRecords)
        totalDuties += dutyRecords.length
      }
    }

    return { schedules: totalSchedules, duties: totalDuties }
  }

  /**
   * 转换排班结果为记录格式
   */
  convertToScheduleRecords(scheduleResult, department, users) {
    const records = []
    const userMap = new Map(users.map(u => [u.id, u]))

    for (const assignment of scheduleResult.assignments) {
      for (const userId of assignment.userIds) {
        const user = userMap.get(userId)
        if (!user) continue

        records.push({
          week: scheduleResult.week,
          weekday: assignment.weekday,
          period: assignment.period,
          user_id: userId,
          user_name: user.name,
          department,
          status: 'confirmed',
          created_at: new Date().toISOString()
        })
      }
    }

    return records
  }

  /**
   * 从排班生成值班记录
   */
  generateDutyRecordsFromSchedule(scheduleRecords, week, department) {
    const duties = []
    const now = new Date()
    const currentWeek = 10 // 假设当前是第10周

    for (const record of scheduleRecords) {
      // 根据周次决定值班状态
      let status = 'pending'
      if (week < currentWeek - 2) {
        status = 'completed' // 过去的值班已完成
      } else if (week < currentWeek) {
        status = Math.random() > 0.3 ? 'completed' : 'pending' // 最近的可能完成也可能未完成
      }

      duties.push({
        week: record.week,
        weekday: record.weekday,
        period: record.period,
        user_id: record.user_id,
        user_name: record.user_name,
        department,
        status,
        duty_status: status,
        check_in_time: status === 'completed' ? new Date(now.getTime() - Math.random() * 7 * 24 * 60 * 60 * 1000).toISOString() : null,
        created_at: new Date().toISOString()
      })
    }

    return duties
  }

  /**
   * 保存值班记录
   */
  async saveDutyRecords(duties) {
    // 使用排班表存储值班记录（简化处理）
    // 实际项目中可能有单独的duty表
    for (const duty of duties) {
      try {
        await db.create('duties', duty, `week_${duty.week}`)
      } catch (err) {
        // 忽略重复错误
      }
    }
  }

  /**
   * 清空排班数据
   */
  async clearSchedules() {
    // 清空所有排班分区
    for (let week = 1; week <= 30; week++) {
      try {
        const schedules = await repositories.schedules.findByWeek(week)
        for (const s of schedules) {
          await repositories.schedules.delete(s.id, week)
        }
      } catch (err) {
        // 忽略错误
      }
    }
  }

  /**
   * 生成当前周排班预览数据
   * 用于演示排班预览功能
   */
  async generatePreviewData(week = 10, department = '竞赛部') {
    const users = await repositories.users.findByDepartment(department)
    const userIds = users.map(u => u.id)
    
    // 获取无课表
    const availability = []
    for (const userId of userIds) {
      const userAvail = await repositories.availability.findByUserAndWeek(userId, week)
      availability.push(...userAvail)
    }

    // 生成排班
    const result = ScheduleGenerator.generateSchedule(
      { week, days: [1, 2, 3, 4, 5], periods: 4, needPerCell: 2 },
      users,
      availability
    )

    // 转换为前端需要的格式
    const grid = []
    for (let weekday = 1; weekday <= 5; weekday++) {
      const row = []
      for (let period = 1; period <= 4; period++) {
        const assignment = result.assignments.find(
          a => a.weekday === weekday && a.period === period
        )
        
        const users_in_cell = assignment ? 
          assignment.userIds.map(id => {
            const user = users.find(u => u.id === id)
            return {
              id,
              name: user?.name || '未知',
              student_id: user?.student_id || ''
            }
          }) : []

        row.push({
          weekday,
          period,
          users: users_in_cell,
          is_complete: users_in_cell.length >= 2
        })
      }
      grid.push(row)
    }

    return {
      week,
      department,
      grid,
      stats: result.stats,
      conflicts: result.conflicts
    }
  }

  /**
   * 生成示例个人值班数据
   * 用于"我的值班"页面
   */
  async generateMyDuties(userId, userName, weeks = [8, 9, 10, 11, 12]) {
    const duties = []
    
    for (const week of weeks) {
      // 每周随机1-2次值班
      const count = 1 + Math.floor(Math.random() * 2)
      const usedSlots = new Set()

      for (let i = 0; i < count; i++) {
        let weekday, period, key
        
        // 确保不重复
        do {
          weekday = 1 + Math.floor(Math.random() * 5)
          period = 1 + Math.floor(Math.random() * 4)
          key = `${weekday}_${period}`
        } while (usedSlots.has(key))
        
        usedSlots.add(key)

        // 根据周次决定状态
        let status = 'pending'
        let check_in_time = null
        
        if (week < 10) {
          status = 'completed'
          check_in_time = new Date(Date.now() - (10 - week) * 7 * 24 * 60 * 60 * 1000).toISOString()
        } else if (week === 10) {
          status = Math.random() > 0.5 ? 'completed' : 'confirmed'
          if (status === 'completed') {
            check_in_time = new Date().toISOString()
          }
        } else {
          status = 'confirmed'
        }

        duties.push({
          id: `duty_${userId}_${week}_${weekday}_${period}`,
          week,
          weekday,
          period,
          user_id: userId,
          user_name: userName,
          status: 'confirmed',
          duty_status: status,
          check_in_time,
          created_at: new Date().toISOString()
        })
      }
    }

    return duties.sort((a, b) => {
      if (a.week !== b.week) return a.week - b.week
      if (a.weekday !== b.weekday) return a.weekday - b.weekday
      return a.period - b.period
    })
  }

  /**
   * 获取统计数据
   */
  async getStats() {
    const stats = {
      totalSchedules: 0,
      totalDuties: 0,
      byWeek: {},
      byDepartment: {}
    }

    for (let week = 1; week <= 20; week++) {
      const weekSchedules = await repositories.schedules.findByWeek(week)
      stats.totalSchedules += weekSchedules.length
      
      if (weekSchedules.length > 0) {
        stats.byWeek[week] = weekSchedules.length
      }

      for (const s of weekSchedules) {
        if (!stats.byDepartment[s.department]) {
          stats.byDepartment[s.department] = 0
        }
        stats.byDepartment[s.department]++
      }
    }

    return stats
  }
}

// 导出单例
export const scheduleSeeder = new ScheduleSeeder()
export default scheduleSeeder
