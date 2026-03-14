/**
 * 预设用户数据 - 用于Mock认证系统
 * 包含4个不同角色的用户，用于演示不同权限视图
 */

export const PRESET_USERS = [
  {
    id: 1,
    student_id: 'admin',
    name: '系统管理员',
    email: 'admin@example.com',
    password: '123456',
    role: 'admin',
    department: '办公室',
    dept_role: 'dept_admin',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=admin',
    created_at: '2024-01-01T00:00:00Z'
  },
  {
    id: 2,
    student_id: 'office001',
    name: '办公室管理员',
    email: 'office@example.com',
    password: '123456',
    role: 'user',
    department: '办公室',
    dept_role: 'dept_admin',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=office',
    created_at: '2024-01-01T00:00:00Z'
  },
  {
    id: 3,
    student_id: 'dept001',
    name: '竞赛部部长',
    email: 'dept@example.com',
    password: '123456',
    role: 'user',
    department: '竞赛部',
    dept_role: 'dept_admin',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=dept',
    created_at: '2024-01-01T00:00:00Z'
  },
  {
    id: 4,
    student_id: 'member001',
    name: '普通成员',
    email: 'member@example.com',
    password: '123456',
    role: 'user',
    department: '竞赛部',
    dept_role: 'dept_member',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=member',
    created_at: '2024-01-01T00:00:00Z'
  }
]

/**
 * 根据ID获取用户信息（包含密码）
 */
export function getUserById(id) {
  return PRESET_USERS.find(u => u.id === id)
}

/**
 * 根据学号获取用户信息
 */
export function getUserByStudentId(studentId) {
  return PRESET_USERS.find(u => u.student_id === studentId)
}

/**
 * 获取安全的用户信息（去除敏感信息）
 */
export function getSanitizedUser(user) {
  if (!user) return null
  const { password, ...safeUser } = user
  return safeUser
}

/**
 * 获取所有预设用户列表（用于角色切换）
 */
export function getPresetUsersForDisplay() {
  return PRESET_USERS.map(u => getSanitizedUser(u))
}
