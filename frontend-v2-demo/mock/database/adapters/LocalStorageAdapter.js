/**
 * LocalStorage 适配器
 * 封装localStorage操作，提供统一的存储接口
 */

export class LocalStorageAdapter {
  constructor(options = {}) {
    this.prefix = options.prefix || 'mock_db_'
    this.storage = window.localStorage
  }

  /**
   * 编码存储键名
   */
  encodeKey(table, partition = null) {
    return partition 
      ? `${this.prefix}${table}_${partition}`
      : `${this.prefix}${table}`
  }

  /**
   * 获取数据
   */
  async get(table, partition = null) {
    const key = this.encodeKey(table, partition)
    try {
      const data = this.storage.getItem(key)
      return data ? JSON.parse(data) : null
    } catch (error) {
      console.error(`[LocalStorageAdapter] 读取失败: ${key}`, error)
      return null
    }
  }

  /**
   * 设置数据
   */
  async set(table, data, partition = null) {
    const key = this.encodeKey(table, partition)
    try {
      const serialized = JSON.stringify(data)
      
      // 检查大小限制（约5MB）
      if (serialized.length > 4.5 * 1024 * 1024) {
        throw new Error('数据大小超过localStorage限制（约5MB）')
      }
      
      this.storage.setItem(key, serialized)
      return true
    } catch (error) {
      console.error(`[LocalStorageAdapter] 写入失败: ${key}`, error)
      throw error
    }
  }

  /**
   * 删除数据
   */
  async delete(table, partition = null) {
    const key = this.encodeKey(table, partition)
    try {
      this.storage.removeItem(key)
      return true
    } catch (error) {
      console.error(`[LocalStorageAdapter] 删除失败: ${key}`, error)
      return false
    }
  }

  /**
   * 获取所有分区键
   */
  async getPartitions(table) {
    const partitions = []
    const prefix = this.encodeKey(table, '')
    
    for (let i = 0; i < this.storage.length; i++) {
      const key = this.storage.key(i)
      if (key && key.startsWith(prefix)) {
        const partition = key.slice(prefix.length)
        if (partition) {
          partitions.push(partition)
        }
      }
    }
    
    return partitions
  }

  /**
   * 清空所有Mock数据
   */
  async clear() {
    const keysToRemove = []
    for (let i = 0; i < this.storage.length; i++) {
      const key = this.storage.key(i)
      if (key && key.startsWith(this.prefix)) {
        keysToRemove.push(key)
      }
    }
    keysToRemove.forEach(key => this.storage.removeItem(key))
    return true
  }

  /**
   * 导出所有数据
   */
  async export() {
    const data = {}
    for (let i = 0; i < this.storage.length; i++) {
      const key = this.storage.key(i)
      if (key && key.startsWith(this.prefix)) {
        const shortKey = key.slice(this.prefix.length)
        try {
          data[shortKey] = JSON.parse(this.storage.getItem(key))
        } catch {
          data[shortKey] = this.storage.getItem(key)
        }
      }
    }
    return data
  }

  /**
   * 导入数据
   */
  async import(data) {
    for (const [key, value] of Object.entries(data)) {
      this.storage.setItem(this.prefix + key, JSON.stringify(value))
    }
    return true
  }

  /**
   * 获取存储统计信息
   */
  async getStats() {
    const stats = {
      totalKeys: 0,
      totalSize: 0,
      tables: {}
    }
    
    for (let i = 0; i < this.storage.length; i++) {
      const key = this.storage.key(i)
      if (key && key.startsWith(this.prefix)) {
        const value = this.storage.getItem(key)
        const size = (key.length + value.length) * 2 // UTF-16编码，每个字符2字节
        
        stats.totalKeys++
        stats.totalSize += size
        
        // 解析表名
        const shortKey = key.slice(this.prefix.length)
        const tableName = shortKey.split('_')[0]
        
        if (!stats.tables[tableName]) {
          stats.tables[tableName] = { keys: 0, size: 0 }
        }
        stats.tables[tableName].keys++
        stats.tables[tableName].size += size
      }
    }
    
    return stats
  }
}

export default LocalStorageAdapter
