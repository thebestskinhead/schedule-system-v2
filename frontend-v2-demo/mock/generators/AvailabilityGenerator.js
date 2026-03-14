/**
 * 无课表数据生成器
 */

export class AvailabilityGenerator {
  /**
   * 生成用户的无课表数据
   * @param {number} userId - 用户ID
   * @param {number} startWeek - 开始周次
   * @param {number} endWeek - 结束周次
   * @param {number} availabilityRate - 有空概率 (0-1)
   */
  static generateForUser(userId, startWeek = 1, endWeek = 30, availabilityRate = 0.4) {
    const availability = []
    
    for (let week = startWeek; week <= endWeek; week++) {
      for (let weekday = 1; weekday <= 5; weekday++) {
        for (let period = 1; period <= 4; period++) {
          availability.push({
            id: `av_${userId}_${week}_${weekday}_${period}`,
            user_id: userId,
            week,
            weekday,
            period,
            is_available: Math.random() < availabilityRate
          })
        }
      }
    }
    
    return availability
  }

  /**
   * 生成更真实的无课表（考虑课程分布规律）
   */
  static generateRealisticForUser(userId, startWeek = 1, endWeek = 30) {
    const availability = []
    
    for (let week = startWeek; week <= endWeek; week++) {
      // 每天课程分布不同
      const dailyPatterns = [
        [0.3, 0.4, 0.6, 0.7], // 周一：上午课多
        [0.4, 0.5, 0.5, 0.6], // 周二
        [0.3, 0.4, 0.4, 0.5], // 周三：课最多
        [0.5, 0.6, 0.7, 0.8], // 周四：课较少
        [0.6, 0.7, 0.8, 0.9]  // 周五：课最少
      ]
      
      for (let weekday = 1; weekday <= 5; weekday++) {
        const patterns = dailyPatterns[weekday - 1]
        
        for (let period = 1; period <= 4; period++) {
          const availableRate = patterns[period - 1]
          
          availability.push({
            id: `av_${userId}_${week}_${weekday}_${period}`,
            user_id: userId,
            week,
            weekday,
            period,
            is_available: Math.random() < availableRate
          })
        }
      }
    }
    
    return availability
  }

  /**
   * 批量生成多个用户的无课表
   */
  static generateForUsers(userIds, startWeek = 1, endWeek = 30) {
    const allAvailability = []
    
    for (const userId of userIds) {
      const userAvailability = this.generateRealisticForUser(userId, startWeek, endWeek)
      allAvailability.push(...userAvailability)
    }
    
    return allAvailability
  }

  /**
   * 生成完全空闲的无课表（用于测试）
   */
  static generateAllAvailable(userId, startWeek = 1, endWeek = 30) {
    const availability = []
    
    for (let week = startWeek; week <= endWeek; week++) {
      for (let weekday = 1; weekday <= 5; weekday++) {
        for (let period = 1; period <= 4; period++) {
          availability.push({
            id: `av_${userId}_${week}_${weekday}_${period}`,
            user_id: userId,
            week,
            weekday,
            period,
            is_available: true
          })
        }
      }
    }
    
    return availability
  }

  /**
   * 生成完全无空的无课表（用于测试冲突情况）
   */
  static generateAllBusy(userId, startWeek = 1, endWeek = 30) {
    const availability = []
    
    for (let week = startWeek; week <= endWeek; week++) {
      for (let weekday = 1; weekday <= 5; weekday++) {
        for (let period = 1; period <= 4; period++) {
          availability.push({
            id: `av_${userId}_${week}_${weekday}_${period}`,
            user_id: userId,
            week,
            weekday,
            period,
            is_available: false
          })
        }
      }
    }
    
    return availability
  }
}

export default AvailabilityGenerator
