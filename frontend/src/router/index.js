import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '../stores/user'
import { getInstallStatus } from '../api/system'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: () => import('../views/Home.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { guest: true }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/Register.vue'),
    meta: { guest: true }
  },
  {
    path: '/forgot-password',
    name: 'ForgotPassword',
    component: () => import('../views/ForgotPassword.vue'),
    meta: { guest: true }
  },
  {
    path: '/reset-password',
    name: 'ResetPassword',
    component: () => import('../views/ResetPassword.vue'),
    meta: { guest: true }
  },
  {
    path: '/init',
    name: 'Init',
    component: () => import('../views/Init.vue'),
    meta: { guest: true }
  },
  {
    path: '/availability',
    name: 'Availability',
    component: () => import('../views/Availability.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/availability/crawler',
    name: 'CrawlerImport',
    component: () => import('../views/CrawlerImport.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/schedule',
    name: 'Schedule',
    component: () => import('../views/Schedule.vue'),
    meta: { requiresAuth: true, requiresManageDept: true }
  },
  {
    path: '/schedule/preview',
    name: 'SchedulePreview',
    component: () => import('../views/SchedulePreview.vue'),
    meta: { requiresAuth: true, requiresManageDept: true }
  },
  {
    path: '/schedule/result',
    name: 'ScheduleResult',
    component: () => import('../views/ScheduleResult.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/duty/my',
    name: 'MyDuty',
    component: () => import('../views/MyDuty.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/admin/users',
    name: 'UserManage',
    component: () => import('../views/UserManage.vue'),
    meta: { requiresAuth: true, requiresManageDept: true }
  },
  {
    path: '/admin/users/v2',
    name: 'UserManagementV2',
    component: () => import('../views/UserManagementV2.vue'),
    meta: { requiresAuth: true, requiresManageAll: true }
  },
  {
    path: '/admin/duty-assignments',
    name: 'DutyAssignment',
    component: () => import('../views/DutyAssignment.vue'),
    meta: { requiresAuth: true, requiresManageAll: true }
  },
  {
    path: '/admin/temp-permissions',
    name: 'TempPermission',
    component: () => import('../views/TempPermission.vue'),
    meta: { requiresAuth: true, requiresSysAdmin: true }
  },
  {
    path: '/admin/smtp',
    name: 'SMTPConfig',
    component: () => import('../views/SMTPConfig.vue'),
    meta: { requiresAuth: true, requiresSysAdmin: true }
  },
  {
    path: '/my-permissions',
    name: 'MyPermissions',
    component: () => import('../views/MyPermissions.vue'),
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore()

  // 检查登录状态（会同时加载临时权限）
  if (!userStore.checked) {
    await userStore.checkAuth()
  }

  const isAuthenticated = userStore.isAuthenticated
  const isAdmin = userStore.isAdmin
  const canManageAll = userStore.canManageAll
  const canManageDept = userStore.canManageDept

  // 阻止已安装系统访问 init 页面
  if (to.path === '/init') {
    try {
      const res = await getInstallStatus()
      if (res.data?.data?.installed) {
        next('/login')
        return
      }
    } catch (error) {
      // 请求失败，继续访问 init
    }
  }

  // 需要登录的页面
  if (to.meta.requiresAuth && !isAuthenticated) {
    next('/login')
    return
  }

  // 需要系统管理员权限（临时权限不能替代系统管理员权限）
  if (to.meta.requiresSysAdmin && !isAdmin) {
    next('/')
    return
  }

  // 需要全局管理权限（系统管理员、办公室管理员或拥有user:manage:all/schedul:view:all临时权限）
  if (to.meta.requiresManageAll && !canManageAll) {
    next('/')
    return
  }

  // 需要部门管理权限（系统管理员、办公室管理员、部门管理员或拥有schedule:manage:dept/user:manage:dept临时权限）
  if (to.meta.requiresManageDept && !canManageDept) {
    next('/')
    return
  }

  // 游客页面（登录页、注册页）已登录时跳转
  if (to.meta.guest && isAuthenticated) {
    next('/')
    return
  }

  next()
})

export default router
