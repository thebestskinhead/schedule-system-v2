/**
 * Mock Database 统一入口
 * 提供完整的数据库操作接口
 */

import { MockDatabase } from './core/Database.js'
import { LocalStorageAdapter } from './adapters/LocalStorageAdapter.js'
import { userRepo, scheduleRepo, availabilityRepo } from './repositories/index.js'

// 创建数据库实例
const adapter = new LocalStorageAdapter({ prefix: 'mock_db_' })
export const db = new MockDatabase({ adapter })

// 重新创建Repository以使用新的数据库实例
import { UserRepository } from './repositories/UserRepository.js'
import { ScheduleRepository } from './repositories/ScheduleRepository.js'
import { AvailabilityRepository } from './repositories/AvailabilityRepository.js'

export const repositories = {
  users: new UserRepository(db),
  schedules: new ScheduleRepository(db),
  availability: new AvailabilityRepository(db)
}

/**
 * 初始化数据库
 * 检查并创建必要的表结构
 */
export async function initDatabase() {
  console.log('[MockDB] 初始化数据库...')
  
  try {
    const stats = await db.getStats()
    console.log(`[MockDB] 当前存储: ${stats.totalKeys} 个键, 约 ${(stats.totalSize / 1024).toFixed(2)} KB`)
    
    // 检查是否首次运行
    const isFirstRun = stats.totalKeys === 0
    
    if (isFirstRun) {
      console.log('[MockDB] 首次运行，准备初始化数据...')
    }
    
    return {
      success: true,
      isFirstRun,
      stats
    }
  } catch (error) {
    console.error('[MockDB] 初始化失败:', error)
    return {
      success: false,
      error: error.message
    }
  }
}

/**
 * 重置数据库
 */
export async function resetDatabase() {
  console.log('[MockDB] 重置数据库...')
  
  try {
    await db.clear()
    console.log('[MockDB] 数据库已重置')
    return { success: true }
  } catch (error) {
    console.error('[MockDB] 重置失败:', error)
    return { success: false, error: error.message }
  }
}

/**
 * 导出数据库数据
 */
export async function exportDatabase() {
  return db.export()
}

/**
 * 导入数据库数据
 */
export async function importDatabase(data) {
  return db.import(data)
}

export default {
  db,
  repositories,
  initDatabase,
  resetDatabase,
  exportDatabase,
  importDatabase
}
