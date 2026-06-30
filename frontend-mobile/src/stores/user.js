import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { getUserInfo, checkToken } from '../api/user'
import { getMyTempPermissions } from '../api/system'

// 权限层级映射（与后端 permissionHierarchy 保持一致）
const permissionHierarchy = {
  'schedule:manage:all': [
    'schedule:view', 'schedule:preview', 'schedule:confirm', 'schedule:edit',
    'schedule:publish', 'schedule:settings', 'schedule:export',
    'schedule:view:all', 'schedule:view:dept', 'schedule:manage:dept'
  ],
  'user:manage:all': [
    'user:manage', 'user:manage:dept', 'user:view', 'user:edit'
  ],
  'schedule:manage:dept': [
    'schedule:view', 'schedule:view:dept', 'schedule:preview',
    'schedule:confirm', 'schedule:edit', 'schedule:settings'
  ],
  'user:manage:dept': [
    'user:view', 'user:edit'
  ]
}

export const useUserStore = defineStore('user', () => {
  // State
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(null)
  const checked = ref(false)
  const tempPermissions = ref([])

  // Getters
  const isAuthenticated = computed(() => !!token.value && !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isOfficeAdmin = computed(() => user.value?.department === '办公室' && user.value?.dept_role === 'dept_admin')
  const isDeptAdmin = computed(() => user.value?.dept_role === 'dept_admin')
  const department = computed(() => user.value?.department || '')

  const canManageAll = computed(() => {
    return isAdmin.value ||
           isOfficeAdmin.value ||
           hasTempPermission('user:manage:all') ||
           hasTempPermission('schedule:manage:all') ||
           hasTempPermission('schedule:view:all')
  })

  const canManageDept = computed(() => {
    return isAdmin.value ||
           isOfficeAdmin.value ||
           isDeptAdmin.value ||
           hasTempPermission('schedule:manage:dept') ||
           hasTempPermission('schedule:manage:all') ||
           hasTempPermission('user:manage:dept') ||
           hasTempPermission('user:manage:all')
  })

  // Actions
  function hasTempPermission(perm) {
    // 先直接匹配
    const directMatch = tempPermissions.value.some(p => p.permission === perm)
    if (directMatch) return true

    // 再通过权限组隐式包含匹配
    for (const [group, perms] of Object.entries(permissionHierarchy)) {
      if (tempPermissions.value.some(p => p.permission === group) && perms.includes(perm)) {
        return true
      }
    }

    return false
  }

  function hasTempPermissionWithResource(perm, resourceType, resourceId) {
    for (const p of tempPermissions.value) {
      if (p.permission === perm) {
        if (!resourceType) return true
        // 如果指定了资源类型，检查是否匹配
        if (p.resource_type === 'all') return true
        if (p.resource_type === resourceType && (!resourceId || p.resource_id === resourceId)) return true
        // 权限组隐式包含
        if (permissionHierarchy[perm]) {
          return true
        }
      }
      // 权限组隐式包含
      if (permissionHierarchy[p.permission]?.includes(perm)) {
        if (!resourceType) return true
        if (p.resource_type === 'all') return true
        if (p.resource_type === resourceType && (!resourceId || p.resource_id === resourceId)) return true
      }
    }
    return false
  }

  // 扫码登录（直接接收 token 和 user）
  async function qrLogin({ token: newToken, user: newUser }) {
    token.value = newToken
    user.value = newUser
    localStorage.setItem('token', newToken)
    await loadTempPermissions()
  }

  function setToken(newToken) {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  function clearToken() {
    token.value = ''
    user.value = null
    tempPermissions.value = []
    localStorage.removeItem('token')
  }

  async function loadUserInfo() {
    try {
      const data = await getUserInfo()
      user.value = data
      await loadTempPermissions()
      return data
    } catch (error) {
      clearToken()
      throw error
    }
  }

  async function loadTempPermissions() {
    try {
      const perms = await getMyTempPermissions()
      tempPermissions.value = perms || []
    } catch (error) {
      console.error('加载临时权限失败:', error)
      tempPermissions.value = []
    }
  }

  async function checkAuth() {
    if (checked.value) return isAuthenticated.value

    if (!token.value) {
      checked.value = true
      return false
    }

    try {
      await checkToken()
      await loadUserInfo()
      checked.value = true
      return true
    } catch (error) {
      clearToken()
      checked.value = true
      return false
    }
  }

  function logout() {
    clearToken()
    user.value = null
    tempPermissions.value = []
    checked.value = false
  }

  return {
    token,
    user,
    checked,
    tempPermissions,
    isAuthenticated,
    isAdmin,
    isOfficeAdmin,
    isDeptAdmin,
    department,
    canManageAll,
    canManageDept,
    qrLogin,
    setToken,
    clearToken,
    loadUserInfo,
    loadTempPermissions,
    checkAuth,
    logout,
    hasTempPermission,
    hasTempPermissionWithResource
  }
})
