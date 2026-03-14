/**
 * 用户管理 Mock Handler
 */

import { db, repositories } from '../database/index.js'
import { UserGenerator } from '../generators/UserGenerator.js'
import { success, error, forbidden } from '../utils/response.js'
import { delay } from '../utils/delay.js'
import { requireAuth, requirePermission } from '../middleware/auth.js'

export const userHandlers = {
  // 获取用户列表
  async getUserList(query, user) {
    await delay(200)
    
    try {
      const options = {
        limit: parseInt(query.limit) || 100,
        offset: parseInt(query.offset) || 0
      }
      
      const users = await repositories.users.findAll(options)
      
      // 过滤敏感信息
      const safeUsers = users.map(u => {
        const { password, ...safe } = u
        return safe
      })
      
      return success(safeUsers)
    } catch (err) {
      return error(err.message)
    }
  },

  // 按部门获取用户
  async getUsersByDepartment(department, user) {
    await delay(150)
    
    try {
      const users = await repositories.users.findByDepartment(department)
      
      const safeUsers = users.map(u => {
        const { password, ...safe } = u
        return safe
      })
      
      return success(safeUsers)
    } catch (err) {
      return error(err.message)
    }
  },

  // 获取可排班用户
  async getUsersForSchedule(user) {
    await delay(150)
    
    try {
      // 如果是办公室管理员，返回所有部门用户
      // 如果是部门管理员，只返回本部门用户
      const department = user.department === '办公室' ? null : user.department
      
      let users
      if (department) {
        users = await repositories.users.findAvailableForSchedule(department)
      } else {
        users = await repositories.users.findAll()
      }
      
      const safeUsers = users.map(u => {
        const { password, ...safe } = u
        return safe
      })
      
      return success(safeUsers)
    } catch (err) {
      return error(err.message)
    }
  },

  // 创建用户
  async createUser(data, user) {
    await delay(400)
    
    try {
      const newUser = await repositories.users.create(data)
      const { password, ...safeUser } = newUser
      return success(safeUser)
    } catch (err) {
      return error(err.message)
    }
  },

  // 更新用户
  async updateUser(id, data, user) {
    await delay(300)
    
    try {
      const updated = await repositories.users.update(parseInt(id), data)
      const { password, ...safeUser } = updated
      return success(safeUser)
    } catch (err) {
      return error(err.message)
    }
  },

  // 删除用户
  async deleteUser(id, user) {
    await delay(300)
    
    try {
      await repositories.users.delete(parseInt(id))
      return success({ success: true })
    } catch (err) {
      return error(err.message)
    }
  },

  // 更新用户角色
  async updateUserRole(id, data, user) {
    await delay(200)
    
    try {
      const { dept_role } = data
      const updated = await repositories.users.updateDeptRole(parseInt(id), dept_role)
      const { password, ...safeUser } = updated
      return success(safeUser)
    } catch (err) {
      return error(err.message)
    }
  },

  // 更新用户部门
  async updateUserDepartment(id, data, user) {
    await delay(200)
    
    try {
      const { department } = data
      const updated = await repositories.users.updateDepartment(parseInt(id), department)
      const { password, ...safeUser } = updated
      return success(safeUser)
    } catch (err) {
      return error(err.message)
    }
  }
}

export default userHandlers
