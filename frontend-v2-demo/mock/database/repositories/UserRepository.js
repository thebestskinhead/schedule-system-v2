/**
 * 用户Repository
 * 封装用户相关的数据操作
 */

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
      .first()
  }

  async findByEmail(email) {
    return this.db.query(this.table)
      .where('email', 'equals', email)
      .first()
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

export default UserRepository
