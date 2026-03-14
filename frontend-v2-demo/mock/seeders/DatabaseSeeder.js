/**
 * 数据库填充器
 * 初始化演示数据
 */

import { PRESET_USERS } from '../auth/preset-users.js'
import { UserGenerator } from '../generators/UserGenerator.js'
import { AvailabilityGenerator } from '../generators/AvailabilityGenerator.js'
import { ScheduleGenerator } from '../generators/ScheduleGenerator.js'
import { repositories, db } from '../database/index.js'

export class DatabaseSeeder {
  constructor() {
    this.initialized = false
  }

  /**
   * 检查是否已初始化
   */
  async checkInitialized() {
    const stats = await db.getStats()
    return stats.totalKeys > 0
  }

  /**
   * 运行填充
   */
  async run(options = {}) {
    console.log('[Seeder] 开始填充演示数据...')
    
    const force = options.force || false
    
    // 检查是否已初始化
    if (!force && await this.checkInitialized()) {
      console.log('[Seeder] 数据库已有数据，跳过填充（使用 force: true 强制重新填充）')
      return { success: true, skipped: true }
    }

    try {
      // 1. 填充预设用户
      await this.seedUsers()
      
      // 2. 填充无课表数据
      await this.seedAvailability()
      
      // 3. 填充排班数据
      await this.seedSchedules()
      
      // 4. 填充系统设置
      await this.seedSettings()
      
      console.log('[Seeder] 数据填充完成！')
      
      // 输出统计
      const stats = await this.getStats()
      console.log('[Seeder] 数据概况:')
      console.log(`  - 用户: ${stats.users} 人`)
      console.log(`  - 无课表记录: ${stats.availability} 条`)
      console.log(`  - 排班记录: ${stats.schedules} 条`)
      
      return { success: true, stats }
    } catch (error) {
      console.error('[Seeder] 填充失败:', error)
      return { success: false, error: error.message }
    }
  }

  /**
   * 填充用户数据
   */
  async seedUsers() {
    console.log('[Seeder] 填充用户数据...')
    
    // 添加预设用户
    for (const user of PRESET_USERS) {
      try {
        await repositories.users.create(user)
      } catch (err) {
        // 用户可能已存在，忽略错误
      }
    }
    
    // 添加额外的演示用户
    const demoDepartments = ['竞赛部', '项目部', '科普部']
    
    for (const dept of demoDepartments) {
      // 为每个部门生成5个成员
      const members = UserGenerator.generateDepartmentMembers(dept, 5, true)
      for (const member of members) {
        try {
          await repositories.users.create(member)
        } catch (err) {
          // 忽略重复错误
        }
      }
    }
  }

  /**
   * 填充无课表数据
   */
  async seedAvailability() {
    console.log('[Seeder] 填充无课表数据...')
    
    // 获取所有用户
    const users = await repositories.users.findAll()
    
    // 为每个用户生成无课表
    for (const user of users) {
      const availability = AvailabilityGenerator.generateRealisticForUser(
        user.id,
        1,
        20 // 只生成前20周的数据
      )
      
      // 批量创建
      await repositories.availability.batchCreate(availability)
    }
  }

  /**
   * 填充排班数据
   * 生成1-16周的完整排班数据，包含值班记录
   */
  async seedSchedules() {
    console.log('[Seeder] 填充排班数据...')
    
    const startWeek = 1
    const endWeek = 16
    const departments = ['办公室', '竞赛部', '项目部', '科普部']
    
    // 为每周每个部门生成排班
    for (let week = startWeek; week <= endWeek; week++) {
      for (const dept of departments) {
        // 获取部门用户
        const users = await repositories.users.findByDepartment(dept)
        if (users.length === 0) continue
        
        // 获取这些用户的无课表
        const userIds = users.map(u => u.id)
        const availability = []
        
        for (const userId of userIds) {
          const userAvail = await repositories.availability.findByUserAndWeek(userId, week)
          availability.push(...userAvail)
        }
        
        // 生成排班
        const scheduleResult = ScheduleGenerator.generateSchedule(
          { week, days: [1, 2, 3, 4, 5], periods: 4, needPerCell: 2 },
          users,
          availability
        )
        
        // 转换为带用户名的记录
        const records = this.convertToScheduleRecords(scheduleResult, dept, users)
        if (records.length > 0) {
          await repositories.schedules.batchCreate(records)
        }
        
        // 生成值班记录（基于排班）
        await this.seedDutyRecords(records, week, dept)
      }
    }
    
    console.log(`[Seeder] 已生成第${startWeek}-${endWeek}周排班数据`)
  }
  
  /**
   * 转换排班结果为记录格式（带用户名）
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
          student_id: user.student_id,
          department,
          status: 'confirmed',
          created_at: new Date().toISOString()
        })
      }
    }
    
    return records
  }
  
  /**
   * 生成值班记录
   */
  async seedDutyRecords(scheduleRecords, week, department) {
    const currentWeek = 10 // 假设当前是第10周
    
    for (const record of scheduleRecords) {
      // 根据周次决定值班状态
      let dutyStatus = 'pending'
      let checkInTime = null
      
      if (week < currentWeek - 2) {
        // 过去的值班都已完成
        dutyStatus = 'completed'
        checkInTime = new Date(Date.now() - (currentWeek - week) * 7 * 24 * 60 * 60 * 1000).toISOString()
      } else if (week < currentWeek) {
        // 最近的值班大部分完成
        dutyStatus = Math.random() > 0.2 ? 'completed' : 'pending'
        if (dutyStatus === 'completed') {
          checkInTime = new Date(Date.now() - Math.random() * 7 * 24 * 60 * 60 * 1000).toISOString()
        }
      } else if (week === currentWeek) {
        // 当前周部分完成
        dutyStatus = Math.random() > 0.7 ? 'completed' : 'confirmed'
        if (dutyStatus === 'completed') {
          checkInTime = new Date().toISOString()
        }
      } else {
        // 未来周次都是已确认
        dutyStatus = 'confirmed'
      }
      
      const dutyRecord = {
        ...record,
        schedule_id: record.id,
        duty_status: dutyStatus,
        check_in_time: checkInTime,
        updated_at: new Date().toISOString()
      }
      
      try {
        await this.saveDutyRecord(dutyRecord)
      } catch (err) {
        // 忽略重复错误
      }
    }
  }
  
  /**
   * 保存值班记录
   */
  async saveDutyRecord(duty) {
    const key = `duties_${duty.week}_${duty.user_id}`
    const existing = await db.findById('duties', key) || []
    
    // 检查是否已存在
    const index = existing.findIndex(d => 
      d.week === duty.week && 
      d.weekday === duty.weekday && 
      d.period === duty.period
    )
    
    if (index >= 0) {
      existing[index] = duty
    } else {
      existing.push(duty)
    }
    
    await db.create('duties', existing, `${duty.week}_${duty.user_id}`)
  }

  /**
   * 填充系统设置
   */
  async seedSettings() {
    console.log('[Seeder] 填充系统设置...')
    
    const settings = {
      id: 'default',
      current_week: 10,
      semester_start: '2024-02-26',
      auto_increment: true,
      need_per_cell: 2,
      min_per_cell: 1,
      max_per_day: 2,
      max_per_week: 4,
      export_title: '第{week}周排班表',
      updated_at: new Date().toISOString()
    }
    
    try {
      await db.create('settings', settings)
    } catch (err) {
      // 设置可能已存在
    }
  }

  /**
   * 获取数据统计
   */
  async getStats() {
    const users = await repositories.users.findAll()
    
    let availability = 0
    for (const user of users) {
      const userAvail = await repositories.availability.findByUser(user.id)
      availability += userAvail.length
    }
    
    let schedules = 0
    let duties = 0
    const weekStats = {}
    
    for (let week = 1; week <= 20; week++) {
      const weekSchedules = await repositories.schedules.findByWeek(week)
      schedules += weekSchedules.length
      
      if (weekSchedules.length > 0) {
        weekStats[week] = weekSchedules.length
      }
      
      // 统计值班记录
      for (const s of weekSchedules) {
        if (s.duty_status) duties++
      }
    }
    
    return {
      users: users.length,
      availability,
      schedules,
      duties,
      weekStats
    }
  }

  /**
   * 重置数据库
   */
  async reset() {
    console.log('[Seeder] 重置数据库...')
    await db.clear()
    return this.run({ force: true })
  }
}

// 导出单例
export const seeder = new DatabaseSeeder()
export default seeder
