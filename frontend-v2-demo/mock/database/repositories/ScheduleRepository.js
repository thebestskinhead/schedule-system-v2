/**
 * 排班Repository
 * 封装排班相关的数据操作
 */

export class ScheduleRepository {
  constructor(database) {
    this.db = database
    this.table = 'schedules'
  }

  async findByWeek(week, department = null) {
    const query = this.db.query(this.table)
      .partition(String(week))
      .where('week', 'equals', week)
    
    if (department) {
      query.where('department', 'equals', department)
    }
    
    return query.execute()
  }

  async findByUser(userId, week = null) {
    const query = this.db.query(this.table)
      .where('user_id', 'equals', userId)
    
    if (week) {
      query.where('week', 'equals', week)
    }
    
    return query.execute()
  }

  async findByUserAndWeek(userId, week) {
    return this.db.query(this.table)
      .where('user_id', 'equals', userId)
      .where('week', 'equals', week)
      .execute()
  }

  async create(scheduleData) {
    const week = scheduleData.week
    return this.db.create(this.table, scheduleData, String(week))
  }

  async update(id, data, week) {
    return this.db.update(this.table, id, data, String(week))
  }

  async delete(id, week) {
    return this.db.delete(this.table, id, String(week))
  }

  async batchCreate(schedules) {
    const byWeek = {}
    schedules.forEach(s => {
      const week = String(s.week)
      if (!byWeek[week]) byWeek[week] = []
      byWeek[week].push(s)
    })

    const results = []
    for (const [week, items] of Object.entries(byWeek)) {
      const created = await this.db.batchCreate(this.table, items, week)
      results.push(...created)
    }
    return results
  }

  async getStats(week, department) {
    const schedules = await this.findByWeek(week, department)
    
    return {
      total: schedules.length,
      byStatus: schedules.reduce((acc, s) => {
        acc[s.status] = (acc[s.status] || 0) + 1
        return acc
      }, {}),
      byWeekday: schedules.reduce((acc, s) => {
        acc[s.weekday] = (acc[s.weekday] || 0) + 1
        return acc
      }, {}),
      byPeriod: schedules.reduce((acc, s) => {
        acc[s.period] = (acc[s.period] || 0) + 1
        return acc
      }, {})
    }
  }

  async getUserStats(userId, startWeek = 1, endWeek = 30) {
    const allSchedules = []
    for (let week = startWeek; week <= endWeek; week++) {
      const schedules = await this.findByUserAndWeek(userId, week)
      allSchedules.push(...schedules)
    }

    return {
      total: allSchedules.length,
      byWeek: allSchedules.reduce((acc, s) => {
        acc[s.week] = (acc[s.week] || 0) + 1
        return acc
      }, {}),
      byStatus: allSchedules.reduce((acc, s) => {
        acc[s.status] = (acc[s.status] || 0) + 1
        return acc
      }, {})
    }
  }

  async clearWeek(week, department = null) {
    const schedules = await this.findByWeek(week, department)
    for (const schedule of schedules) {
      await this.delete(schedule.id, week)
    }
    return schedules.length
  }
}

export default ScheduleRepository
