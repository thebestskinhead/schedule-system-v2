/**
 * 查询构建器
 * 提供链式调用的查询接口
 */

export class QueryBuilder {
  constructor(database, table) {
    this.db = database
    this.table = table
    this.conditions = []
    this.sortField = null
    this.sortOrder = 'asc'
    this.limitValue = null
    this.offsetValue = 0
    this.partitionValue = null
  }

  /**
   * 添加WHERE条件
   */
  where(field, operator, value) {
    // 支持两种调用方式:
    // 1. where('name', 'equals', 'John')
    // 2. where({ name: { equals: 'John' }, age: { gt: 18 } })
    if (typeof field === 'object') {
      Object.entries(field).forEach(([f, condition]) => {
        if (typeof condition === 'object') {
          const [op, val] = Object.entries(condition)[0]
          this.conditions.push({ field: f, operator: op, value: val })
        } else {
          this.conditions.push({ field: f, operator: 'equals', value: condition })
        }
      })
    } else {
      this.conditions.push({ field, operator, value })
    }
    return this
  }

  /**
   * 设置排序
   */
  orderBy(field, order = 'asc') {
    this.sortField = field
    this.sortOrder = order
    return this
  }

  /**
   * 设置限制数量
   */
  limit(n) {
    this.limitValue = n
    return this
  }

  /**
   * 设置偏移量
   */
  offset(n) {
    this.offsetValue = n
    return this
  }

  /**
   * 设置分区
   */
  partition(p) {
    this.partitionValue = p
    return this
  }

  /**
   * 执行查询
   */
  async execute() {
    // 构建查询对象
    const query = {
      partition: this.partitionValue
    }

    if (this.conditions.length > 0) {
      query.where = this.conditions.reduce((acc, cond) => {
        acc[cond.field] = { [cond.operator]: cond.value }
        return acc
      }, {})
    }

    if (this.sortField) {
      query.orderBy = [`${this.sortField}:${this.sortOrder}`]
    }

    if (this.offsetValue > 0) {
      query.offset = this.offsetValue
    }

    if (this.limitValue !== null) {
      query.limit = this.limitValue
    }

    return this.db.findMany(this.table, query)
  }

  /**
   * 获取第一条数据
   */
  async first() {
    this.limit(1)
    const results = await this.execute()
    return results[0] || null
  }

  /**
   * 获取数量
   */
  async count() {
    const results = await this.execute()
    return results.length
  }

  /**
   * 检查是否存在
   */
  async exists() {
    const count = await this.count()
    return count > 0
  }

  /**
   * 获取去重后的字段值列表
   */
  async distinct(field) {
    const results = await this.execute()
    const values = results.map(item => item[field])
    return [...new Set(values)]
  }

  /**
   * 分页查询
   */
  async paginate(page, pageSize = 10) {
    this.offset((page - 1) * pageSize)
    this.limit(pageSize)
    
    const [data, total] = await Promise.all([
      this.execute(),
      this.db.findMany(this.table, { partition: this.partitionValue }).then(r => r.length)
    ])

    return {
      list: data,
      total,
      page,
      page_size: pageSize,
      total_pages: Math.ceil(total / pageSize)
    }
  }
}

export default QueryBuilder
