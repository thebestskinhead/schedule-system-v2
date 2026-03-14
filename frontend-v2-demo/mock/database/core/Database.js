/**
 * Mock Database 核心类
 * 提供类SQL的数据库操作接口
 */

import { LocalStorageAdapter } from '../adapters/LocalStorageAdapter.js'
import { QueryBuilder } from './QueryBuilder.js'

export class MockDatabase {
  constructor(options = {}) {
    this.adapter = options.adapter || new LocalStorageAdapter()
    this.indexes = new Map()
    this.cache = new Map()
    this.cacheEnabled = options.cache !== false
    this.cacheSize = options.cacheSize || 100
    
    // 注册表结构
    this.registerTable('users', { primaryKey: 'id', indexes: ['student_id', 'email', 'department'] })
    this.registerTable('schedules', { primaryKey: 'id', indexes: ['week', 'user_id', 'department'], partitioned: true, partitionKey: 'week' })
    this.registerTable('availability', { primaryKey: 'id', indexes: ['user_id', 'week'], partitioned: true, partitionKey: 'user_id' })
    this.registerTable('settings', { primaryKey: 'id' })
    this.registerTable('applications', { primaryKey: 'id', indexes: ['applicant_id', 'status'] })
    this.registerTable('temp_permissions', { primaryKey: 'id', indexes: ['user_id'] })
    this.registerTable('duty_assignments', { primaryKey: 'id', indexes: ['week', 'department'] })
  }

  /**
   * 注册表
   */
  registerTable(table, options = {}) {
    this.indexes.set(table, {
      primary: options.primaryKey || 'id',
      indexes: options.indexes || [],
      partitioned: options.partitioned || false,
      partitionKey: options.partitionKey || null
    })
  }

  /**
   * 获取分区键值
   */
  getPartitionKey(table, data) {
    const tableInfo = this.indexes.get(table)
    if (!tableInfo || !tableInfo.partitioned) return null
    
    const key = tableInfo.partitionKey
    if (typeof data === 'object') {
      return data[key] || null
    }
    return data
  }

  // ==================== 查询操作 ====================

  /**
   * 查询多条数据
   */
  async findMany(table, query = {}) {
    const cacheKey = this.getCacheKey(table, query)
    
    if (this.cacheEnabled && this.cache.has(cacheKey)) {
      return [...this.cache.get(cacheKey)] // 返回副本
    }

    let data = []
    
    if (query.partition) {
      data = await this.adapter.get(table, query.partition) || []
    } else {
      const tableInfo = this.indexes.get(table)
      if (tableInfo && tableInfo.partitioned) {
        const partitions = await this.adapter.getPartitions(table)
        for (const partition of partitions) {
          const partitionData = await this.adapter.get(table, partition) || []
          data = data.concat(partitionData)
        }
      } else {
        data = await this.adapter.get(table) || []
      }
    }

    // 应用过滤条件
    if (query.where) {
      data = this.applyWhere(data, query.where)
    }

    // 应用排序
    if (query.orderBy) {
      data = this.applyOrderBy(data, query.orderBy)
    }

    // 应用分页
    if (query.offset || query.limit) {
      const start = query.offset || 0
      const end = query.limit ? start + query.limit : undefined
      data = data.slice(start, end)
    }

    // 缓存结果
    if (this.cacheEnabled) {
      this.setCache(cacheKey, data)
    }

    return data
  }

  /**
   * 查询单条数据
   */
  async findOne(table, query = {}) {
    const results = await this.findMany(table, { ...query, limit: 1 })
    return results[0] || null
  }

  /**
   * 根据ID查询
   */
  async findById(table, id, partition = null) {
    const data = await this.adapter.get(table, partition) || []
    return data.find(item => item.id === id) || null
  }

  /**
   * 获取查询构建器
   */
  query(table) {
    return new QueryBuilder(this, table)
  }

  // ==================== 写入操作 ====================

  /**
   * 创建数据
   */
  async create(table, data, partition = null) {
    const partitionKey = partition || this.getPartitionKey(table, data)
    const items = await this.adapter.get(table, partitionKey) || []
    
    const newItem = {
      ...data,
      id: data.id || this.generateId(),
      created_at: data.created_at || new Date().toISOString(),
      updated_at: new Date().toISOString()
    }
    
    items.push(newItem)
    await this.adapter.set(table, items, partitionKey)
    
    // 清除相关缓存
    this.clearTableCache(table)
    
    return newItem
  }

  /**
   * 更新数据
   */
  async update(table, id, data, partition = null) {
    const partitionKey = partition || this.getPartitionKey(table, data)
    const items = await this.adapter.get(table, partitionKey) || []
    const index = items.findIndex(item => item.id === id)
    
    if (index === -1) {
      throw new Error(`记录不存在: ${table}.${id}`)
    }
    
    items[index] = {
      ...items[index],
      ...data,
      id,  // 确保ID不变
      updated_at: new Date().toISOString()
    }
    
    await this.adapter.set(table, items, partitionKey)
    this.clearTableCache(table)
    
    return items[index]
  }

  /**
   * 删除数据
   */
  async delete(table, id, partition = null) {
    const items = await this.adapter.get(table, partition) || []
    const filtered = items.filter(item => item.id !== id)
    
    if (filtered.length === items.length) {
      return false  // 未找到
    }
    
    await this.adapter.set(table, filtered, partition)
    this.clearTableCache(table)
    
    return true
  }

  /**
   * 批量创建
   */
  async batchCreate(table, items, partition = null) {
    const partitionKey = partition || this.getPartitionKey(table, {})
    const existing = await this.adapter.get(table, partitionKey) || []
    
    const newItems = items.map(data => ({
      ...data,
      id: data.id || this.generateId(),
      created_at: data.created_at || new Date().toISOString(),
      updated_at: new Date().toISOString()
    }))
    
    await this.adapter.set(table, [...existing, ...newItems], partitionKey)
    this.clearTableCache(table)
    
    return newItems
  }

  // ==================== 事务支持 ====================

  /**
   * 执行事务（简化版）
   */
  async transaction(operations) {
    const backup = await this.adapter.export()
    
    try {
      const results = []
      for (const op of operations) {
        results.push(await op())
      }
      return results
    } catch (error) {
      // 回滚
      await this.adapter.import(backup)
      throw error
    }
  }

  // ==================== 工具方法 ====================

  applyWhere(data, where) {
    return data.filter(item => {
      return Object.entries(where).every(([field, condition]) => {
        const itemValue = item[field]
        
        if (typeof condition === 'object') {
          const [operator, value] = Object.entries(condition)[0]
          switch (operator) {
            case 'equals': return itemValue === value
            case 'not': return itemValue !== value
            case 'gt': return itemValue > value
            case 'gte': return itemValue >= value
            case 'lt': return itemValue < value
            case 'lte': return itemValue <= value
            case 'in': return value.includes(itemValue)
            case 'contains': return String(itemValue).includes(value)
            case 'startsWith': return String(itemValue).startsWith(value)
            case 'endsWith': return String(itemValue).endsWith(value)
            default: return true
          }
        }
        
        return itemValue === condition
      })
    })
  }

  applyOrderBy(data, orderBy) {
    const fields = Array.isArray(orderBy) ? orderBy : [orderBy]
    
    return data.sort((a, b) => {
      for (const field of fields) {
        const [name, order] = field.split(':')
        const aVal = a[name]
        const bVal = b[name]
        
        if (aVal === undefined && bVal === undefined) continue
        if (aVal === undefined) return 1
        if (bVal === undefined) return -1
        
        if (aVal < bVal) return order === 'desc' ? 1 : -1
        if (aVal > bVal) return order === 'desc' ? -1 : 1
      }
      return 0
    })
  }

  generateId() {
    return Date.now() + Math.random().toString(36).substr(2, 9)
  }

  getCacheKey(table, query) {
    return `${table}:${JSON.stringify(query)}`
  }

  setCache(key, data) {
    if (this.cache.size >= this.cacheSize) {
      const firstKey = this.cache.keys().next().value
      this.cache.delete(firstKey)
    }
    this.cache.set(key, data)
  }

  clearTableCache(table) {
    for (const key of this.cache.keys()) {
      if (key.startsWith(`${table}:`)) {
        this.cache.delete(key)
      }
    }
  }

  // ==================== 数据管理 ====================

  async clear() {
    await this.adapter.clear()
    this.cache.clear()
  }

  async export() {
    return this.adapter.export()
  }

  async import(data) {
    await this.adapter.import(data)
    this.cache.clear()
  }

  async getStats() {
    return this.adapter.getStats()
  }
}

// 导出单例
export const db = new MockDatabase()
export default MockDatabase
