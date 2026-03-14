/**
 * 用户数据生成器
 */

import { PRESET_USERS } from '../auth/preset-users.js'

export class UserGenerator {
  static departments = ['办公室', '竞赛部', '项目部', '科普部', '外联部']
  static deptRoles = ['dept_admin', 'dept_member']
  static firstNames = ['伟', '芳', '娜', '敏', '静', '丽', '强', '磊', '军', '洋', '勇', '艳', '杰', '娟', '涛', '明', '超', '秀英', '华', '鹏']
  static lastNames = ['王', '李', '张', '刘', '陈', '杨', '黄', '赵', '吴', '周', '徐', '孙', '马', '朱', '胡', '郭', '何', '林', '罗', '高']

  /**
   * 生成预设用户
   */
  static generatePresetUsers() {
    return PRESET_USERS.map(u => {
      const { password, ...safeUser } = u
      return { ...safeUser, password: '123456' }
    })
  }

  /**
   * 生成随机用户
   */
  static generateRandomUser(options = {}) {
    const id = options.id || Date.now() + Math.random().toString(36).substr(2, 9)
    const studentId = options.student_id || this.generateStudentId()
    const name = options.name || this.generateName()
    
    return {
      id,
      student_id: studentId,
      name,
      email: options.email || `${studentId}@example.com`,
      password: options.password || '123456',
      role: options.role || 'user',
      department: options.department || this.randomDepartment(),
      dept_role: options.dept_role || this.randomDeptRole(),
      avatar: options.avatar || `https://api.dicebear.com/7.x/avataaars/svg?seed=${studentId}`,
      created_at: new Date().toISOString()
    }
  }

  /**
   * 批量生成用户
   */
  static generateUsers(count, options = {}) {
    const users = []
    for (let i = 0; i < count; i++) {
      users.push(this.generateRandomUser({
        department: options.department,
        dept_role: options.dept_role
      }))
    }
    return users
  }

  /**
   * 生成学号
   */
  static generateStudentId() {
    const year = 2021 + Math.floor(Math.random() * 4)
    const dept = ['01', '02', '03', '04', '05'][Math.floor(Math.random() * 5)]
    const num = String(Math.floor(Math.random() * 999)).padStart(3, '0')
    return `${year}${dept}${num}`
  }

  /**
   * 生成姓名
   */
  static generateName() {
    const lastName = this.lastNames[Math.floor(Math.random() * this.lastNames.length)]
    const firstName = this.firstNames[Math.floor(Math.random() * this.firstNames.length)]
    return lastName + firstName
  }

  /**
   * 随机部门
   */
  static randomDepartment() {
    return this.departments[Math.floor(Math.random() * this.departments.length)]
  }

  /**
   * 随机部门角色
   */
  static randomDeptRole() {
    // 10%概率是管理员
    return Math.random() < 0.1 ? 'dept_admin' : 'dept_member'
  }

  /**
   * 为部门生成成员
   */
  static generateDepartmentMembers(department, count, includeAdmin = true) {
    const members = []
    
    if (includeAdmin) {
      members.push(this.generateRandomUser({
        department,
        dept_role: 'dept_admin',
        name: `${department}部长`
      }))
    }
    
    for (let i = 0; i < count - (includeAdmin ? 1 : 0); i++) {
      members.push(this.generateRandomUser({
        department,
        dept_role: 'dept_member'
      }))
    }
    
    return members
  }
}

export default UserGenerator
