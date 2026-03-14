/**
 * 无课表管理 Mock Handler
 */

import { repositories } from '../database/index.js'
import { success, error, forbidden, notFound } from '../utils/response.js'
import { delay } from '../utils/delay.js'

// 导入任务队列（内存存储）
let importTasks = []
let taskIdCounter = 1

export const availabilityHandlers = {
  // 获取我的无课表
  async getMyAvailability(user, query = {}) {
    await delay(200)
    
    try {
      const { week } = query
      const availability = await repositories.availability.findByUser(user.id, week)
      return success(availability)
    } catch (err) {
      return error(err.message)
    }
  },

  // 添加无课时间
  async addAvailability(data, user) {
    await delay(300)
    
    try {
      const { week, weekday, period } = data
      
      // 检查是否已存在
      const existing = await repositories.availability.findByUserAndWeek(user.id, week)
      const found = existing.find(e => e.weekday === weekday && e.period === period)
      
      if (found) {
        // 更新现有记录
        const updated = await repositories.availability.update(
          found.id,
          { is_available: true, updated_at: new Date().toISOString() },
          user.id
        )
        return success(updated)
      } else {
        // 创建新记录
        const created = await repositories.availability.create({
          user_id: user.id,
          week,
          weekday,
          period,
          is_available: true,
          created_at: new Date().toISOString()
        })
        return success(created)
      }
    } catch (err) {
      return error(err.message)
    }
  },

  // 删除无课时间
  async deleteAvailability(data, user) {
    await delay(300)
    
    try {
      const { week, weekday, period } = data
      
      const existing = await repositories.availability.findByUserAndWeek(user.id, week)
      const found = existing.find(e => e.weekday === weekday && e.period === period)
      
      if (found) {
        await repositories.availability.delete(found.id, user.id)
        return success({ success: true })
      } else {
        return notFound('记录不存在')
      }
    } catch (err) {
      return error(err.message)
    }
  },

  // Cookie 导入（模拟）
  async importFromCookie(data, user) {
    await delay(1000)
    
    try {
      const { cookie, week_range } = data
      
      // 创建导入任务
      const taskId = `task_${taskIdCounter++}`
      const task = {
        id: taskId,
        type: 'cookie',
        status: 'processing',
        user_id: user.id,
        progress: 0,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }
      importTasks.push(task)
      
      // 模拟异步处理
      setTimeout(async () => {
        task.status = 'completed'
        task.progress = 100
        task.result = {
          imported_count: Math.floor(Math.random() * 20) + 10,
          failed_count: 0
        }
        task.updated_at = new Date().toISOString()
        
        // 生成模拟的无课表数据
        await generateMockAvailability(user.id, week_range)
      }, 2000)
      
      return success({ task_id: taskId })
    } catch (err) {
      return error(err.message)
    }
  },

  // XLS 导入（模拟）
  async importFromXLS(data, user) {
    await delay(1000)
    
    try {
      const { week_range } = data
      
      // 创建导入任务
      const taskId = `task_${taskIdCounter++}`
      const task = {
        id: taskId,
        type: 'xls',
        status: 'processing',
        user_id: user.id,
        progress: 0,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }
      importTasks.push(task)
      
      // 模拟异步处理
      setTimeout(async () => {
        task.status = 'completed'
        task.progress = 100
        task.result = {
          imported_count: Math.floor(Math.random() * 20) + 10,
          failed_count: 0
        }
        task.updated_at = new Date().toISOString()
        
        // 生成模拟的无课表数据
        await generateMockAvailability(user.id, week_range)
      }, 2000)
      
      return success({ task_id: taskId })
    } catch (err) {
      return error(err.message)
    }
  },

  // 获取导入任务状态
  async getImportStatus(query) {
    await delay(100)
    
    const { task_id } = query
    const task = importTasks.find(t => t.id === task_id)
    
    if (!task) {
      return notFound('任务不存在')
    }
    
    return success({
      task_id: task.id,
      status: task.status,
      progress: task.progress,
      result: task.result
    })
  },

  // 获取导入任务列表
  async getImportTaskList(user) {
    await delay(150)
    
    const tasks = importTasks
      .filter(t => t.user_id === user.id)
      .sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
      .slice(0, 10)
    
    return success(tasks)
  }
}

// 辅助函数：生成模拟无课表数据
async function generateMockAvailability(userId, weekRange) {
  const items = []
  const [startWeek, endWeek] = weekRange || [1, 4]
  
  for (let week = startWeek; week <= endWeek; week++) {
    for (let weekday = 1; weekday <= 5; weekday++) {
      for (let period = 1; period <= 4; period++) {
        // 随机生成可用状态（70%概率可用）
        if (Math.random() > 0.3) {
          items.push({
            user_id: userId,
            week,
            weekday,
            period,
            is_available: true,
            created_at: new Date().toISOString()
          })
        }
      }
    }
  }
  
  if (items.length > 0) {
    await repositories.availability.batchCreate(items)
  }
}

export default availabilityHandlers
