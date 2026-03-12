import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '../stores/user'
import { getInstallStatus } from '../api/system'

const routes = [
  {
    path: '/',
    component: () => import('../components/Layout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Home',
        component: () => import('../views/Home.vue')
      },
      {
        path: 'availability',
        name: 'Availability',
        component: () => import('../views/Availability.vue')
      },
      {
        path: 'availability/crawler',
        name: 'CrawlerImport',
        component: () => import('../views/CrawlerImport.vue')
      },
      {
        path: 'duty/my',
        name: 'MyDuty',
        component: () => import('../views/MyDuty.vue')
      },
      {
        path: 'schedule/result',
        name: 'ScheduleResult',
        component: () => import('../views/ScheduleResult.vue')
      },
      {
        path: 'my-permissions',
        name: 'MyPermissions',
        component: () => import('../views/MyPermissions.vue')
      },
      {
        path: 'readme',
        name: 'Readme',
        component: () => import('../views/Readme.vue')
      },
      // 管理功能
      {
        path: 'schedule',
        name: 'Schedule',
        component: () => import('../views/Schedule.vue'),
        meta: { requiresManageDept: true }
      },
      {
        path: 'admin/duty-assignments',
        name: 'DutyAssignment',
        component: () => import('../views/DutyAssignment.vue'),
        meta: { requiresManageAll: true }
      },
      {
        path: 'admin/users',
        name: 'UserManagement',
        component: () => import('../views/UserManagement.vue'),
        meta: { requiresManageAll: true }
      },
      {
        path: 'admin/temp-permissions',
        name: 'TempPermission',
        component: () => import('../views/TempPermission.vue'),
        meta: { requiresSysAdmin: true }
      },
      {
        path: 'admin/smtp',
        name: 'SMTPConfig',
        component: () => import('../views/SMTPConfig.vue'),
        meta: { requiresSysAdmin: true }
      },
      {
        path: 'admin/semester',
        name: 'SemesterSettings',
        component: () => import('../views/SemesterSettings.vue'),
        meta: { requiresManageAll: true }
      }
    ]
  },
  // 无布局页面
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
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore()

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

  if (to.meta.requiresAuth && !isAuthenticated) {
    next('/login')
    return
  }

  if (to.meta.requiresSysAdmin && !isAdmin) {
    next('/')
    return
  }

  if (to.meta.requiresManageAll && !canManageAll) {
    next('/')
    return
  }

  if (to.meta.requiresManageDept && !canManageDept) {
    next('/')
    return
  }

  if (to.meta.guest && isAuthenticated) {
    next('/')
    return
  }

  next()
})

export default router
