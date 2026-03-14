# 任务4：本地存储模拟数据库 - 详细设计报告

## 一、现状分析

### 1.1 现有数据持久化调研

```
Current Storage Usage Analysis:

src/stores/user.js
└── localStorage usage:
    ├── token: localStorage.getItem('token')
    └── Persistence: Login时写入，Logout时清除

src/api/request.js
└── No direct storage usage
    └── Token passed via Header only

Views (Home.vue, Schedule.vue, etc.)
└── No direct storage usage
    └── All data from API calls

Current State:
├── 使用场景单一: 仅Token存储
├── 数据结构简单: 字符串存储
├── 无查询能力: 直接读写
└── 无关联维护: 数据孤立
```

### 1.2 localStorage现状评估

```javascript
// 当前存储使用情况
localStorage.currentUsage = {
  // 仅存储这些键
  'token': 'eyJhbGciOiJIUzI1NiIs...',  // JWT token
  // 无其他业务数据存储
}

// 存储限制
localStorage.limits = {
  sizePerKey: '~5MB per domain',
  totalKeys: '无明确限制',
  syncAccess: '同步阻塞',
  stringOnly: '仅支持字符串'
}
```

### 1.3 数据关系分析

```
Entity Relationship Diagram (Required for Demo):

Users ||--o{ Availability : has
Users ||--o{ Schedules : assigned_to
Users ||--o{ Duties : has
Users ||--o{ TempPermissions : granted
Users ||--o{ Applications : submits

Departments ||--o{ Users : contains
Departments ||--o{ Schedules : owns
Departments ||--o{ DutyAssignments : assigned

Schedules }o--|| Weeks : belongs_to
Schedules }o--|| Periods : at

Applications }o--|| Users : approved_by
Applications }o--|| Permissions : requests

Current Problem: No relational storage capability
```

## 二、理想状态设计

### 2.1 模拟数据库架构

```
Mock Database Architecture:

┌─────────────────────────────────────────────────────────────────────┐
│                        Mock Database Layer                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                   Storage Adapter Layer                       │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │  │
│  │  │ localStorage │  │  IndexedDB   │  │   Memory     │       │  │
│  │  │   (主要)      │  │   (备用)     │  │   (缓存)      │       │  │
│  │  └──────────────┘  └──────────────┘  └──────────────┘       │  │
│  └───────────────────────────┬──────────────────────────────────┘  │
│                              │                                       │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                   Database Engine                             │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐        │  │
│  │  │  Query   │ │  Index   │ │ Transaction│ │  Schema  │        │  │
│  │  │  Engine  │ │  Manager │ │  Manager │ │ Validator│        │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘        │  │
│  └───────────────────────────┬──────────────────────────────────┘  │
│                              │                                       │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                   ORM / Repository Layer                      │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐        │  │
│  │  │UserRepo  │ │ScheduleRepo│ │AvailRepo │ │PermRepo  │        │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘        │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.2 存储策略设计

#### 策略1: localStorage (Primary)
```javascript
// 存储结构设计
const storageSchema = {
  // 元数据
  'mock_db_meta': {
    version: '1.0.0',
    lastUpdate: timestamp,
    checksum: '...'
  },
  
  // 用户表
  'mock_db_users': JSON.stringify([
    { id: 1, student_id: 'admin', name: '...', ... },
    // ...
  ]),
  
  // 排班表 (按周次分区存储，避免单键过大)
  'mock_db_schedules_w1': JSON.stringify([...]),
  'mock_db_schedules_w2': JSON.stringify([...]),
  // ...
  
  // 无课表 (按用户分区)
  'mock_db_availability_u1': JSON.stringify([...]),
  'mock_db_availability_u2': JSON.stringify([...]),
  
  // 其他表...
  'mock_db_settings': JSON.stringify({...}),
  'mock_db_applications': JSON.stringify([...]),
  'mock_db_temp_permissions': JSON.stringify([...])
}
```

#### 策略2: IndexedDB (Large Data)
```javascript
// 用于存储大数据量（如30周×多用户的完整排班）
const indexedDBSchema = {
  database: 'MockScheduleDB',
  version: 1,
  stores: {
    schedules: { keyPath: 'id', indexes: ['week', 'department'] },
    availability: { keyPath: 'id', indexes: ['user_id', 'week'] },
    logs: { keyPath: 'id', indexes: ['timestamp'] }
  }
}
```

#### 策略3: Memory Cache (Hot Data)
```javascript
// 内存缓存层
const memoryCache = {
  users: new Map(),      // 热点用户数据
  currentWeek: null,     // 当前周次
  userPermissions: new Map(), // 用户权限缓存
  
  // LRU淘汰策略
  maxSize: 100,
  accessOrder: []
}
```

### 2.3 数据库引擎设计

```javascript
// 类SQL查询接口
class MockDatabase {
  constructor() {
    this.storage = new StorageAdapter()
    this.indexes = new Map()
    this.buildIndexes()
  }

  // 查询接口
  async find(table, query = {}) {
    const data = await this.storage.get(table)
    return this.applyQuery(data, query)
  }

  async findOne(table, query) {
    const results = await this.find(table, query)
    return results[0] || null
  }

  async findById(table, id) {
    return this.findOne(table, { where: { id: { equals: id } } })
  }

  // 写入接口
  async create(table, data) {
    const items = await this.storage.get(table) || []
    const newItem = { ...data, id: this.generateId() }
    items.push(newItem)
    await this.storage.set(table, items)
    this.updateIndex(table, newItem)
    return newItem
  }

  async update(table, id, data) {
    const items = await this.storage.get(table) || []
    const index = items.findIndex(item => item.id === id)
    if (index === -1) throw new Error('Record not found')
    
    items[index] = { ...items[index], ...data, id }
    await this.storage.set(table, items)
    this.updateIndex(table, items[index])
    return items[index]
  }

  async delete(table, id) {
    const items = await this.storage.get(table) || []
    const filtered = items.filter(item => item.id !== id)
    await this.storage.set(table, filtered)
    this.removeFromIndex(table, id)
    return true
  }

  // 查询构建器
  query(table) {
    return new QueryBuilder(this, table)
  }
}

// 查询构建器
class QueryBuilder {
  constructor(db, table) {
    this.db = db
    this.table = table
    this.conditions = []
    this.sortField = null
    this.sortOrder = 'asc'
    this.limitValue = null
    this.offsetValue = 0
  }

  where(field, operator, value) {
    this.conditions.push({ field, operator, value })
    return this
  }

  orderBy(field, order = 'asc') {
    this.sortField = field
    this.sortOrder = order
    return this
  }

  limit(n) {
    this.limitValue = n
    return this
  }

  offset(n) {
    this.offsetValue = n
    return this
  }

  async execute() {
    let data = await this.db.storage.get(this.table) || []
    
    // 应用条件过滤
    data = data.filter(item => {
      return this.conditions.every(cond => {
        const itemValue = item[cond.field]
        switch(cond.operator) {
          case 'equals': return itemValue === cond.value
          case 'not': return itemValue !== cond.value
          case 'gt': return itemValue > cond.value
          case 'gte': return itemValue >= cond.value
          case 'lt': return itemValue < cond.value
          case 'lte': return itemValue <= cond.value
          case 'in': return cond.value.includes(itemValue)
          case 'contains': return String(itemValue).includes(cond.value)
          default: return true
        }
      })
    })
    
    // 排序
    if (this.sortField) {
      data.sort((a, b) => {
        const aVal = a[this.sortField]
        const bVal = b[this.sortField]
        const compare = aVal < bVal ? -1 : aVal > bVal ? 1 : 0
        return this.sortOrder === 'desc' ? -compare : compare
      })
    }
    
    // 分页
    if (this.offsetValue) {
      data = data.slice(this.offsetValue)
    }
    if (this.limitValue) {
      data = data.slice(0, this.limitValue)
    }
    
    return data
  }
}
```

### 2.4 Repository层设计

```javascript
// 用户Repository
class UserRepository {
  constructor(db) {
    this.db = db
    this.table = 'users'
  }

  async findByStudentId(studentId) {
    return this.db.findOne(this.table, {
      where: { student_id: { equals: studentId } }
    })
  }

  async findByDepartment(department) {
    return this.db.query(this.table)
      .where('department', 'equals', department)
      .execute()
  }

  async findAvailableForSchedule(department, week) {
    // 复杂查询：获取某部门用户及其某周的无课情况
    const users = await this.findByDepartment(department)
    const availability = await db.availability.findMany({
      where: { week: { equals: week } }
    })
    
    return users.map(user => ({
      ...user,
      availability: availability.filter(a => a.user_id === user.id)
    }))
  }
}

// 排班Repository
class ScheduleRepository {
  constructor(db) {
    this.db = db
    this.table = 'schedules'
  }

  async findByWeek(week, department = null) {
    const query = this.db.query(this.table)
      .where('week', 'equals', week)
    
    if (department) {
      query.where('department', 'equals', department)
    }
    
    return query.execute()
  }

  async findByUserAndWeek(userId, week) {
    return this.db.query(this.table)
      .where('user_id', 'equals', userId)
      .where('week', 'equals', week)
      .execute()
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
      }, {})
    }
  }
}

// Repository工厂
const repositories = {
  users: (db) => new UserRepository(db),
  schedules: (db) => new ScheduleRepository(db),
  availability: (db) => new AvailabilityRepository(db),
  applications: (db) => new ApplicationRepository(db),
  // ...
}
```

## 三、现有结构与理想的差距

### 3.1 差距分析

```
Gap Analysis:

1. 存储能力差距 (Critical)
   ├── 现状: 仅支持简单键值存储
   ├── 需求: 关系型数据存储与查询
   └── 方案: 封装localStorage，实现类SQL接口

2. 查询能力差距 (Critical)
   ├── 现状: 无查询能力，直接读写
   ├── 需求: 条件查询、排序、分页
   └── 方案: 实现QueryBuilder

3. 数据关联差距 (High)
   ├── 现状: 数据孤立，无关联
   ├── 需求: 表关联查询（用户-排班-无课表）
   └── 方案: Repository层处理关联

4. 事务支持差距 (Medium)
   ├── 现状: 无事务概念
   ├── 需求: 原子操作（如确认排班需同时更新多张表）
   └── 方案: 简单事务模拟

5. 索引性能差距 (Medium)
   ├── 现状: 全表扫描
   ├── 需求: 快速按ID、周次等查询
   └── 方案: 内存索引+分区存储
```

### 3.2 性能对比预估

| 场景 | 当前方式 | 理想方式 | 性能提升 |
|-----|---------|---------|---------|
| 查询用户ByID | O(n)扫描 | O(1)索引 | 100x+ |
| 查询排班By周 | O(n)扫描 | O(1)分区 | 50x+ |
| 复杂关联查询 | 多次API | 单次内存 | 10x+ |
| 数据写入 | API请求 | 本地存储 | 瞬时 |

## 四、修改位置与实施方案

### 4.1 文件结构规划

```
mock/
├── database/
│   ├── index.js                 # 数据库入口
│   ├── adapters/
│   │   ├── localStorage.js      # localStorage适配器
│   │   ├── indexedDB.js         # IndexedDB适配器
│   │   └── memory.js            # 内存适配器
│   ├── core/
│   │   ├── Database.js          # 数据库核心类
│   │   ├── QueryBuilder.js      # 查询构建器
│   │   ├── IndexManager.js      # 索引管理
│   │   └── Transaction.js       # 事务管理
│   ├── repositories/
│   │   ├── index.js             # Repository导出
│   │   ├── UserRepository.js    # 用户仓库
│   │   ├── ScheduleRepository.js # 排班仓库
│   │   ├── AvailabilityRepository.js # 无课表仓库
│   │   ├── ApplicationRepository.js  # 申请仓库
│   │   └── PermissionRepository.js   # 权限仓库
│   └── models/
│       ├── User.js              # 用户模型
│       ├── Schedule.js          # 排班模型
│       └── ...                  # 其他模型
│
└── seeders/
    └── DatabaseSeeder.js        # 数据库填充
```

### 4.2 核心实现代码

#### 4.2.1 StorageAdapter (mock/database/adapters/localStorage.js)

```javascript
class LocalStorageAdapter {
  constructor(options = {}) {
    this.prefix = options.prefix || 'mock_db_'
    this.storage = window.localStorage
  }

  // 键名编码
  encodeKey(table, partition = null) {
    return partition 
      ? `${this.prefix}${table}_${partition}`
      : `${this.prefix}${table}`
  }

  // 获取数据
  async get(table, partition = null) {
    const key = this.encodeKey(table, partition)
    const data = this.storage.getItem(key)
    return data ? JSON.parse(data) : null
  }

  // 设置数据
  async set(table, data, partition = null) {
    const key = this.encodeKey(table, partition)
    const serialized = JSON.stringify(data)
    
    // 检查大小限制
    if (serialized.length > 5 * 1024 * 1024) {
      throw new Error('Data size exceeds localStorage limit')
    }
    
    this.storage.setItem(key, serialized)
    return true
  }

  // 删除数据
  async delete(table, partition = null) {
    const key = this.encodeKey(table, partition)
    this.storage.removeItem(key)
    return true
  }

  // 获取所有分区
  async getPartitions(table) {
    const partitions = []
    const prefix = this.encodeKey(table, '')
    
    for (let i = 0; i < this.storage.length; i++) {
      const key = this.storage.key(i)
      if (key && key.startsWith(prefix)) {
        const partition = key.slice(prefix.length)
        partitions.push(partition)
      }
    }
    
    return partitions
  }

  // 清空所有数据
  async clear() {
    const keysToRemove = []
    for (let i = 0; i < this.storage.length; i++) {
      const key = this.storage.key(i)
      if (key && key.startsWith(this.prefix)) {
        keysToRemove.push(key)
      }
    }
    keysToRemove.forEach(key => this.storage.removeItem(key))
  }

  // 导出所有数据
  async export() {
    const data = {}
    for (let i = 0; i < this.storage.length; i++) {
      const key = this.storage.key(i)
      if (key && key.startsWith(this.prefix)) {
        const shortKey = key.slice(this.prefix.length)
        data[shortKey] = JSON.parse(this.storage.getItem(key))
      }
    }
    return data
  }

  // 导入数据
  async import(data) {
    for (const [key, value] of Object.entries(data)) {
      this.storage.setItem(this.prefix + key, JSON.stringify(value))
    }
  }
}
```

#### 4.2.2 Database核心类 (mock/database/core/Database.js)

```javascript
import { LocalStorageAdapter } from '../adapters/localStorage'
import { QueryBuilder } from './QueryBuilder'

class MockDatabase {
  constructor(options = {}) {
    this.adapter = options.adapter || new LocalStorageAdapter()
    this.indexes = new Map()
    this.cache = new Map()
    this.cacheEnabled = options.cache !== false
    this.cacheSize = options.cacheSize || 100
  }

  // 表注册
  registerTable(table, options = {}) {
    this.indexes.set(table, {
      primary: options.primaryKey || 'id',
      indexes: options.indexes || []
    })
  }

  // 基础CRUD
  async findMany(table, query = {}) {
    const cacheKey = this.getCacheKey(table, query)
    
    if (this.cacheEnabled && this.cache.has(cacheKey)) {
      return this.cache.get(cacheKey)
    }

    // 获取分区数据
    let data = []
    if (query.partition) {
      data = await this.adapter.get(table, query.partition) || []
    } else {
      // 合并所有分区
      const partitions = await this.adapter.getPartitions(table)
      for (const partition of partitions) {
        const partitionData = await this.adapter.get(table, partition) || []
        data = data.concat(partitionData)
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

  async findOne(table, query) {
    const results = await this.findMany(table, { ...query, limit: 1 })
    return results[0] || null
  }

  async findById(table, id, partition = null) {
    // 使用索引优化
    const indexInfo = this.indexes.get(table)
    if (indexInfo && indexInfo.primary === 'id') {
      // 尝试从分区直接获取
      const data = await this.adapter.get(table, partition) || []
      return data.find(item => item.id === id) || null
    }
    
    return this.findOne(table, {
      where: { id: { equals: id } },
      partition
    })
  }

  async create(table, data, partition = null) {
    const items = await this.adapter.get(table, partition) || []
    
    const newItem = {
      ...data,
      id: data.id || this.generateId(),
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    }
    
    items.push(newItem)
    await this.adapter.set(table, items, partition)
    
    // 清除相关缓存
    this.clearTableCache(table)
    
    return newItem
  }

  async update(table, id, data, partition = null) {
    const items = await this.adapter.get(table, partition) || []
    const index = items.findIndex(item => item.id === id)
    
    if (index === -1) {
      throw new Error(`Record with id ${id} not found in ${table}`)
    }
    
    items[index] = {
      ...items[index],
      ...data,
      id,  // 确保ID不变
      updated_at: new Date().toISOString()
    }
    
    await this.adapter.set(table, items, partition)
    this.clearTableCache(table)
    
    return items[index]
  }

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

  // 批量操作
  async batchCreate(table, items, partition = null) {
    const existing = await this.adapter.get(table, partition) || []
    const newItems = items.map(data => ({
      ...data,
      id: data.id || this.generateId(),
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    }))
    
    await this.adapter.set(table, [...existing, ...newItems], partition)
    this.clearTableCache(table)
    
    return newItems
  }

  // 查询构建器
  query(table) {
    return new QueryBuilder(this, table)
  }

  // 事务（简化版）
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

  // 辅助方法
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
      // LRU淘汰
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

  // 数据导入导出
  async export() {
    return this.adapter.export()
  }

  async import(data) {
    await this.adapter.import(data)
    this.cache.clear()
  }

  async clear() {
    await this.adapter.clear()
    this.cache.clear()
  }
}

export const db = new MockDatabase()
export default MockDatabase
```

#### 4.2.3 Repository层 (mock/database/repositories/UserRepository.js)

```javascript
export class UserRepository {
  constructor(database) {
    this.db = database
    this.table = 'users'
  }

  async findAll(options = {}) {
    return this.db.query(this.table)
      .orderBy(options.orderBy || 'id:asc')
      .limit(options.limit)
      .offset(options.offset)
      .execute()
  }

  async findById(id) {
    return this.db.findById(this.table, id)
  }

  async findByStudentId(studentId) {
    return this.db.query(this.table)
      .where('student_id', 'equals', studentId)
      .execute()
      .then(results => results[0] || null)
  }

  async findByEmail(email) {
    return this.db.query(this.table)
      .where('email', 'equals', email)
      .execute()
      .then(results => results[0] || null)
  }

  async findByDepartment(department, options = {}) {
    return this.db.query(this.table)
      .where('department', 'equals', department)
      .orderBy(options.orderBy || 'name:asc')
      .execute()
  }

  async findByRole(role) {
    return this.db.query(this.table)
      .where('role', 'equals', role)
      .execute()
  }

  async findDeptAdmins(department = null) {
    const query = this.db.query(this.table)
      .where('dept_role', 'equals', 'dept_admin')
    
    if (department) {
      query.where('department', 'equals', department)
    }
    
    return query.execute()
  }

  async findAvailableForSchedule(department) {
    return this.db.query(this.table)
      .where('department', 'equals', department)
      .where('dept_role', 'in', ['dept_admin', 'dept_member'])
      .execute()
  }

  async create(userData) {
    // 验证唯一性
    const existing = await this.findByStudentId(userData.student_id)
    if (existing) {
      throw new Error('学号已存在')
    }

    const emailExisting = await this.findByEmail(userData.email)
    if (emailExisting) {
      throw new Error('邮箱已存在')
    }

    return this.db.create(this.table, {
      ...userData,
      created_at: new Date().toISOString()
    })
  }

  async update(id, userData) {
    const user = await this.findById(id)
    if (!user) {
      throw new Error('用户不存在')
    }

    // 如果修改学号，检查唯一性
    if (userData.student_id && userData.student_id !== user.student_id) {
      const existing = await this.findByStudentId(userData.student_id)
      if (existing) {
        throw new Error('学号已存在')
      }
    }

    return this.db.update(this.table, id, userData)
  }

  async delete(id) {
    const user = await this.findById(id)
    if (!user) {
      throw new Error('用户不存在')
    }

    // 检查是否有排班关联
    const schedules = await db.schedules.findByUser(id)
    if (schedules.length > 0) {
      throw new Error('该用户有排班记录，无法删除')
    }

    return this.db.delete(this.table, id)
  }

  async updateDepartment(id, department) {
    return this.update(id, { department })
  }

  async updateDeptRole(id, deptRole) {
    return this.update(id, { dept_role: deptRole })
  }

  async getStats() {
    const users = await this.findAll()
    
    return {
      total: users.length,
      byDepartment: users.reduce((acc, u) => {
        acc[u.department] = (acc[u.department] || 0) + 1
        return acc
      }, {}),
      byRole: users.reduce((acc, u) => {
        acc[u.dept_role] = (acc[u.dept_role] || 0) + 1
        return acc
      }, {})
    }
  }
}
```

### 4.3 实施步骤

```
Phase 1: 基础架构 (2h)
├── 1.1 创建 database/ 目录结构
├── 1.2 实现 LocalStorageAdapter
├── 1.3 实现 Database 核心类
└── 1.4 实现 QueryBuilder

Phase 2: Repository层 (3h)
├── 2.1 UserRepository
├── 2.2 ScheduleRepository
├── 2.3 AvailabilityRepository
├── 2.4 ApplicationRepository
└── 2.5 PermissionRepository

Phase 3: 数据分区优化 (1.5h)
├── 3.1 设计分区策略（按周次、按用户）
├── 3.2 实现分区查询
├── 3.3 添加缓存机制
└── 3.4 性能测试

Phase 4: 事务与工具 (1.5h)
├── 4.1 实现简单事务
├── 4.2 数据导入导出功能
├── 4.3 数据库填充脚本
└── 4.4 集成到 Mock API
```

## 五、再次分析：核心需求与更优解

### 5.1 核心需求重定义

原始需求: "本地存储模拟数据库"
深层需求分析:
1. **数据持久化** - 刷新页面后数据不丢失
2. **查询能力** - 支持复杂条件查询
3. **关系维护** - 数据之间有关联关系
4. **性能可接受** - 操作流畅，无明显卡顿
5. **开发友好** - API简洁，易于使用

### 5.2 存储方案对比

| 方案 | 容量 | 查询能力 | 结构化 | 推荐度 |
|-----|-----|---------|-------|-------|
| localStorage | ~5-10MB | 需自行实现 | 差 | ⭐⭐⭐ |
| IndexedDB | 较大 | 较好 | 好 | ⭐⭐⭐⭐⭐ |
| SQLite (wasm) | 大 | 强 | 强 | ⭐⭐⭐⭐ |
| OPFS | 大 | 需自行实现 | 中 | ⭐⭐⭐ |

### 5.3 更优方案: Hybrid Storage (混合存储)

```
Optimized Storage Strategy:

┌─────────────────────────────────────────────────────────────┐
│                    Hybrid Storage Layer                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                   Hot Data (Memory)                    │  │
│  │  - 当前登录用户                                        │  │
│  │  - 当前周次设置                                        │  │
│  │  - 热点排班数据                                        │  │
│  │  - 用户权限缓存                                        │  │
│  │  LRU淘汰策略                                           │  │
│  └───────────────────────────────────────────────────────┘  │
│                           │                                  │
│                           ▼                                  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │              Warm Data (localStorage)                  │  │
│  │  - 用户基础信息                                        │  │
│  │  - 配置设置                                            │  │
│  │  - 近期排班（当前周±2周）                               │  │
│  │  - 申请和权限记录                                      │  │
│  │  分区存储，快速读写                                     │  │
│  └───────────────────────────────────────────────────────┘  │
│                           │                                  │
│                           ▼                                  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │               Cold Data (IndexedDB)                    │  │
│  │  - 历史排班数据（超过2周）                              │  │
│  │  - 大量无课表记录                                      │  │
│  │  - 操作日志                                            │  │
│  │  大容量，复杂查询                                       │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**优势**:
1. **性能优化** - 热数据内存访问，微秒级响应
2. **容量扩展** - 冷数据存IndexedDB，突破5MB限制
3. **查询能力** - IndexedDB支持索引和范围查询
4. **渐进增强** - 从localStorage开始，按需升级

### 5.4 关键决策

| 决策点 | 选择 | 理由 |
|-------|-----|-----|
| 主要存储 | localStorage | 兼容性好，API简单 |
| 大容量存储 | IndexedDB (可选) | 需要时启用，渐进增强 |
| 缓存策略 | LRU | 简单有效，内存可控 |
| 分区策略 | 按业务+时间 | 平衡查询性能和存储效率 |
| ORM风格 | Repository模式 | 清晰分层，易于测试 |

---

**结论**: 采用"localStorage为主 + IndexedDB为辅 + 内存缓存"的混合存储方案，在保证兼容性的同时优化性能和容量。
