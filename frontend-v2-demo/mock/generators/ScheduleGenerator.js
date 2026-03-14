/**
 * 排班数据生成器
 * 包含排班算法
 */

export class ScheduleGenerator {
  /**
   * 生成排班
   * @param {Object} options - 排班选项
   * @param {number} options.week - 周次
   * @param {number[]} options.days - 工作日 [1,2,3,4,5]
   * @param {number} options.periods - 每天节数
   * @param {number} options.needPerCell - 每时段需要人数
   * @param {Array} users - 用户列表
   * @param {Array} availability - 无课表数据
   */
  static generateSchedule(options, users, availability) {
    const { week, days = [1, 2, 3, 4, 5], periods = 4, needPerCell = 2 } = options
    const assignments = []
    
    // 按用户ID组织无课表数据
    const availabilityMap = new Map()
    for (const av of availability) {
      if (av.week === week && av.is_available) {
        const key = `${av.user_id}_${av.weekday}_${av.period}`
        if (!availabilityMap.has(key)) {
          availabilityMap.set(key, [])
        }
        availabilityMap.get(key).push(av.user_id)
      }
    }
    
    // 遍历每个时段
    for (const weekday of days) {
      for (let period = 1; period <= periods; period++) {
        const key = `_${weekday}_${period}`
        const availableUserIds = []
        
        // 收集该时段有空的用户
        for (const user of users) {
          const userKey = `${user.id}_${weekday}_${period}`
          if (availabilityMap.has(userKey)) {
            availableUserIds.push(user.id)
          }
        }
        
        // 随机选择用户
        const selectedUsers = this.shuffleArray(availableUserIds).slice(0, needPerCell)
        
        assignments.push({
          weekday,
          period,
          userIds: selectedUsers,
          userCount: selectedUsers.length,
          isComplete: selectedUsers.length >= needPerCell
        })
      }
    }
    
    return {
      week,
      assignments,
      stats: {
        totalSlots: days.length * periods,
        filledSlots: assignments.filter(a => a.userCount > 0).length,
        completeSlots: assignments.filter(a => a.isComplete).length,
        totalAssigned: assignments.reduce((sum, a) => sum + a.userCount, 0),
        conflicts: []
      }
    }
  }

  /**
   * 生成排班记录（用于保存确认后的排班）
   */
  static generateScheduleRecords(scheduleResult, department) {
    const records = []
    const { week, assignments } = scheduleResult
    
    for (const assignment of assignments) {
      for (const userId of assignment.userIds) {
        records.push({
          id: `sch_${week}_${assignment.weekday}_${assignment.period}_${userId}`,
          week,
          weekday: assignment.weekday,
          period: assignment.period,
          user_id: userId,
          department,
          status: 'confirmed',
          created_at: new Date().toISOString()
        })
      }
    }
    
    return records
  }

  /**
   * 生成示例排班数据
   */
  static generateDemoSchedule(week, department, userNames) {
    const records = []
    const weekdays = [1, 2, 3, 4, 5]
    const periods = [1, 2, 3, 4]
    
    for (const weekday of weekdays) {
      for (const period of periods) {
        // 每个时段随机1-2人
        const count = 1 + Math.floor(Math.random() * 2)
        const shuffled = this.shuffleArray([...userNames])
        
        for (let i = 0; i < count && i < shuffled.length; i++) {
          records.push({
            id: `demo_${week}_${weekday}_${period}_${i}`,
            week,
            weekday,
            period,
            user_name: shuffled[i],
            department,
            status: Math.random() > 0.2 ? 'confirmed' : 'pending'
          })
        }
      }
    }
    
    return records
  }

  /**
   * 生成个人值班记录
   */
  static generateDutyRecords(userId, userName, startWeek = 1, endWeek = 30) {
    const records = []
    
    for (let week = startWeek; week <= endWeek; week++) {
      // 每周随机1-3次值班
      const count = 1 + Math.floor(Math.random() * 3)
      const scheduledDays = new Set()
      
      for (let i = 0; i < count; i++) {
        let weekday
        do {
          weekday = 1 + Math.floor(Math.random() * 5)
        } while (scheduledDays.has(weekday))
        scheduledDays.add(weekday)
        
        const period = 1 + Math.floor(Math.random() * 4)
        
        records.push({
          id: `duty_${userId}_${week}_${weekday}`,
          user_id: userId,
          user_name: userName,
          week,
          weekday,
          period,
          status: week < 10 ? 'completed' : week < 15 ? 'confirmed' : 'pending'
        })
      }
    }
    
    return records
  }

  /**
   * Fisher-Yates 洗牌算法
   */
  static shuffleArray(array) {
    const arr = [...array]
    for (let i = arr.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1))
      ;[arr[i], arr[j]] = [arr[j], arr[i]]
    }
    return arr
  }
}

export default ScheduleGenerator
