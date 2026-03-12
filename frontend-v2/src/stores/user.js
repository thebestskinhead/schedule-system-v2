import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { getUserInfo, checkToken } from '../api/user'
import { getMyTempPermissions } from '../api/system'

export const useUserStore = defineStore('user', () => {
  // State
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(null)
  const checked = ref(false)
  const tempPermissions = ref([])

  // Getters
  const isAuthenticated = computed(() => !!token.value && !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  // 办公室管理员：部门为办公室且是部门管理员
  const isOfficeAdmin = computed(() => user.value?.department === '办公室' && user.value?.dept_role === 'dept_admin')
  const isDeptAdmin = computed(() => user.value?.dept_role === 'dept_admin')
  
  const canManageAll = computed(() => {
    return isAdmin.value || 
           isOfficeAdmin.value ||
           hasTempPermission('user:manage:all') ||
           hasTempPermission('schedule:view:all')
  })
  
  const canManageDept = computed(() => {
    return isAdmin.value || 
           isOfficeAdmin.value || 
           isDeptAdmin.value ||
           hasTempPermission('schedule:manage:dept') ||
           hasTempPermission('user:manage:dept')
  })

  // Actions
  function hasTempPermission(perm) {
    return tempPermissions.value.some(p => p.permission === perm)
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
      // 加载临时权限
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
    canManageAll,
    canManageDept,
    setToken,
    clearToken,
    loadUserInfo,
    checkAuth,
    logout,
    hasTempPermission
  }
})
