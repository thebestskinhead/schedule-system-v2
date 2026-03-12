import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as loginApi, getProfile, register as registerApi } from '../api/user'
import { tempPermissionAPI } from '../api/tempPermission'

export const useUserStore = defineStore('user', () => {
  // State
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(null)
  const checked = ref(false)
  const tempPermissions = ref([])

  // Getters
  const isAuthenticated = computed(() => !!token.value && !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isDeptAdmin = computed(() => user.value?.dept_role === 'dept_admin')
  const isOfficeAdmin = computed(() => user.value?.department === '办公室' && user.value?.dept_role === 'dept_admin')
  const department = computed(() => user.value?.department || '')
  
  // 权限组包含的子权限映射（与后端保持一致）
  const permissionHierarchy = {
    // 排班管理（全部）包含所有排班相关权限
    'schedule:manage:all': [
      'schedule:view', 'schedule:preview', 'schedule:confirm', 'schedule:edit',
      'schedule:publish', 'schedule:settings', 'schedule:export',
      'schedule:view:all', 'schedule:view:dept', 'schedule:manage:dept'
    ],
    // 用户管理（全部）包含用户查看、编辑、部门管理
    'user:manage:all': [
      'user:manage', 'user:manage:dept', 'user:view', 'user:edit'
    ],
    // 部门排班管理包含部门排班相关权限
    'schedule:manage:dept': [
      'schedule:view', 'schedule:view:dept', 'schedule:preview', 'schedule:confirm', 'schedule:edit'
    ],
    // 部门用户管理包含部门用户相关权限
    'user:manage:dept': [
      'user:view', 'user:edit'
    ]
  }

  // 检查是否有特定临时权限（支持权限组）
  const hasTempPermission = (perm) => {
    // 直接检查是否拥有该权限
    const hasDirect = tempPermissions.value.some(p => p.permission === perm)
    if (hasDirect) return true

    // 检查是否拥有包含该权限的权限组
    for (const [groupPerm, subPerms] of Object.entries(permissionHierarchy)) {
      if (subPerms.includes(perm)) {
        const hasGroup = tempPermissions.value.some(p => p.permission === groupPerm)
        if (hasGroup) return true
      }
    }
    return false
  }

  // 检查是否有特定临时权限（带资源类型过滤）
  const hasTempPermissionWithResource = (perm, resourceType, resourceID = null) => {
    // 检查直接权限
    const hasDirect = tempPermissions.value.some(p => {
      if (p.permission !== perm) return false
      if (p.resource_type === 'all' || !p.resource_type) return true
      if (p.resource_type === resourceType) return true
      return false
    })
    if (hasDirect) return true

    // 检查权限组
    for (const [groupPerm, subPerms] of Object.entries(permissionHierarchy)) {
      if (subPerms.includes(perm)) {
        const hasGroup = tempPermissions.value.some(p => {
          if (p.permission !== groupPerm) return false
          if (p.resource_type === 'all' || !p.resource_type) return true
          if (p.resource_type === resourceType) return true
          return false
        })
        if (hasGroup) return true
      }
    }
    return false
  }

  // 是否有管理权限（系统管理员或办公室管理员或临时权限）
  const canManageAll = computed(() => {
    if (isAdmin.value || isOfficeAdmin.value) return true
    // 检查临时权限（权限组或其子权限）
    return hasTempPermission('user:manage:all') ||
           hasTempPermission('schedule:manage:all') ||
           hasTempPermission('schedule:view:all')
  })

  // 是否可以管理部门（系统管理员、办公室管理员、部门管理员或临时权限）
  const canManageDept = computed(() => {
    if (isAdmin.value || isOfficeAdmin.value || isDeptAdmin.value) return true
    // 检查临时权限（权限组或其子权限）
    return hasTempPermission('schedule:manage:dept') ||
           hasTempPermission('schedule:manage:all') ||
           hasTempPermission('user:manage:dept') ||
           hasTempPermission('user:manage:all')
  })

  // Actions
  const checkAuth = async () => {
    try {
      if (!token.value) {
        checked.value = true
        return
      }

      const profile = await getProfile()
      user.value = profile
      
      // 加载临时权限
      await loadTempPermissions()
    } catch (error) {
      token.value = ''
      localStorage.removeItem('token')
      tempPermissions.value = []
    } finally {
      checked.value = true
    }
  }
  
  // 加载临时权限
  const loadTempPermissions = async () => {
    try {
      const res = await tempPermissionAPI.getMy()
      tempPermissions.value = res || []
    } catch (error) {
      console.error('加载临时权限失败:', error)
      tempPermissions.value = []
    }
  }

  const login = async (credentials) => {
    const data = await loginApi(credentials)
    token.value = data.token
    user.value = data.user
    localStorage.setItem('token', data.token)
    // 登录成功后加载临时权限
    await loadTempPermissions()
    return data
  }

  const register = async (data) => {
    const result = await registerApi(data)
    return result
  }

  const logout = () => {
    token.value = ''
    user.value = null
    tempPermissions.value = []
    localStorage.removeItem('token')
  }

  const updateUser = (data) => {
    user.value = { ...user.value, ...data }
  }

  return {
    token,
    user,
    checked,
    tempPermissions,
    isAuthenticated,
    isAdmin,
    isDeptAdmin,
    isOfficeAdmin,
    department,
    canManageAll,
    canManageDept,
    hasTempPermission,
    hasTempPermissionWithResource,
    checkAuth,
    login,
    register,
    logout,
    updateUser,
    loadTempPermissions
  }
})
