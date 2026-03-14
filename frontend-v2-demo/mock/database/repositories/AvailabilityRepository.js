/**
 * 无课表Repository
 * 封装无课表相关的数据操作
 */

export class AvailabilityRepository {
  constructor(database) {
    this.db = database
    this.table = 'availability'
  }

  async findByUser(userId, week = null) {
    const query = this.db.query(this.table)
      .partition(String(userId))
      .where('user_id', 'equals', userId)
    
    if (week) {
      query.where('week', 'equals', week)
    }
    
    return query.execute()
  }

  async findByUserAndWeek(userId, week) {
    return this.db.query(this.table)
      .partition(String(userId))
      .where('user_id', 'equals', userId)
      .where('week', 'equals', week)
      .execute()
  }

  async findAvailableUsers(week, weekday, period, userIds = null) {
    let query = this.db.query(this.table)
      .where('week', 'equals', week)
      .where('weekday', 'equals', weekday)
      .where('period', 'equals', period)
      .where('is_available', 'equals', true)
    
    if (userIds && userIds.length > 0) {
      query = query.where('user_id', 'in', userIds)
    }
    
    return query.execute()
  }

  async create(data) {
    const userId = data.user_id
    return this.db.create(this.table, data, String(userId))
  }

  async update(id, data, userId) {
    return this.db.update(this.table, id, data, String(userId))
  }

  async delete(id, userId) {
    return this.db.delete(this.table, id, String(userId))
  }

  async batchCreate(items) {
    const byUser = {}
    items.forEach(item => {
      const userId = String(item.user_id)
      if (!byUser[userId]) byUser[userId] = []
      byUser[userId].push(item)
    })

    const results = []
    for (const [userId, userItems] of Object.entries(byUser)) {
      const created = await this.db.batchCreate(this.table, userItems, userId)
      results.push(...created)
    }
    return results
  }

  async batchUpdate(userId, week, updates) {
    const existing = await this.findByUserAndWeek(userId, week)
    const results = []

    for (const update of updates) {
      const item = existing.find(e => 
        e.weekday === update.weekday && e.period === update.period
      )
      
      if (item) {
        results.push(await this.update(item.id, { is_available: update.is_available }, userId))
      } else {
        results.push(await this.create({
          user_id: userId,
          week,
          weekday: update.weekday,
          period: update.period,
          is_available: update.is_available
        }))
      }
    }

    return results
  }

  async getUserAvailabilityStats(userId, startWeek = 1, endWeek = 30) {
    const allAvailability = []
    for (let week = startWeek; week <= endWeek; week++) {
      const availability = await this.findByUserAndWeek(userId, week)
      allAvailability.push(...availability)
    }

    const available = allAvailability.filter(a => a.is_available).length
    const total = allAvailability.length

    return {
      total,
      available,
      unavailable: total - available,
      availabilityRate: total > 0 ? (available / total * 100).toFixed(2) + '%' : '0%',
      byWeekday: allAvailability.reduce((acc, a) => {
        if (!acc[a.weekday]) acc[a.weekday] = { available: 0, total: 0 }
        acc[a.weekday].total++
        if (a.is_available) acc[a.weekday].available++
        return acc
      }, {})
    }
  }

  async clearUserAvailability(userId) {
    const items = await this.findByUser(userId)
    for (const item of items) {
      await this.delete(item.id, userId)
    }
    return items.length
  }
}

export default AvailabilityRepository
