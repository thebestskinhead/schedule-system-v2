<template>
  <div class="layout">
    <!-- 顶部导航 -->
    <header class="header">
      <div class="header-inner">
        <div class="brand" @click="$router.push('/')">
          <el-icon size="24" color="#409eff"><Calendar /></el-icon>
          <span class="brand-text">排班系统</span>
        </div>
        
        <nav class="nav">
          <router-link 
            v-for="item in mainNavItems" 
            :key="item.path"
            :to="item.path"
            class="nav-item"
            :class="{ active: isActive(item.path) }"
          >
            <el-icon :size="16"><component :is="item.icon" /></el-icon>
            <span>{{ item.name }}</span>
          </router-link>
          
          <!-- 管理菜单 -->
          <el-dropdown v-if="hasManagePermission" placement="bottom">
            <div class="nav-item" :class="{ active: isManageActive }">
              <el-icon :size="16"><Setting /></el-icon>
              <span>管理</span>
              <el-icon :size="12"><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item 
                  v-for="item in manageNavItems" 
                  :key="item.path"
                  @click="$router.push(item.path)"
                >
                  <el-icon :size="14"><component :is="item.icon" /></el-icon>
                  <span>{{ item.name }}</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </nav>

        <div class="user-section">
          <!-- 角色切换器 - 仅在Mock模式显示 -->
          <RoleSwitcher />
          
          <!-- 当前角色标签 -->
          <el-tag v-if="userStore.isAdmin" type="danger" size="small" class="role-badge">系统管理员</el-tag>
          <el-tag v-else-if="userStore.isOfficeAdmin" type="success" size="small" class="role-badge">办公室管理</el-tag>
          <el-tag v-else-if="userStore.isDeptAdmin" type="warning" size="small" class="role-badge">部门管理</el-tag>
          
          <el-dropdown placement="bottom-end">
            <div class="user-trigger">
              <el-avatar :size="32" :icon="UserFilled" />
              <span class="user-name">{{ userStore.user?.name || '用户' }}</span>
              <el-icon><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="$router.push('/my-permissions')">
                  <el-icon><Key /></el-icon> 我的权限
                </el-dropdown-item>
                <el-dropdown-item divided @click="handleLogout">
                  <el-icon><SwitchButton /></el-icon> 退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>
    </header>

    <!-- 主内容区 -->
    <main class="main">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '../stores/user'
import RoleSwitcher from './RoleSwitcher.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

// 主导航 - 常用功能
const mainNavItems = computed(() => [
  { path: '/', name: '首页', icon: 'HomeFilled' },
  { path: '/availability', name: '无课表', icon: 'Calendar' },
  { path: '/duty/my', name: '我的值班', icon: 'Timer' },
  { path: '/schedule/result', name: '排班结果', icon: 'View' },
  { path: '/admin/temp-permissions', name: '权限申请', icon: 'Key' },
  { path: '/readme', name: '使用说明', icon: 'Document' }
])

// 管理导航
const manageNavItems = computed(() => {
  const items = []
  if (userStore.canManageDept) {
    items.push({ path: '/schedule', name: '排班管理', icon: 'Edit' })
  }
  if (userStore.canManageAll) {
    items.push({ path: '/admin/duty-assignments', name: '每周分工', icon: 'List' })
    items.push({ path: '/admin/users', name: '用户管理', icon: 'User' })
    items.push({ path: '/admin/semester', name: '学期设置', icon: 'Date' })
  }
  if (userStore.isAdmin) {
    items.push({ path: '/admin/smtp', name: 'SMTP配置', icon: 'Message' })
  }
  return items
})

const hasManagePermission = computed(() => manageNavItems.value.length > 0)

const isActive = (path) => {
  if (path === '/') {
    return route.path === '/'
  }
  return route.path.startsWith(path)
}

const isManageActive = computed(() => {
  return manageNavItems.value.some(item => route.path.startsWith(item.path))
})

const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    userStore.logout()
    ElMessage.success('已退出登录')
    router.push('/login')
  } catch {
    // 取消
  }
}
</script>

<style scoped>
.layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.header {
  background: #fff;
  border-bottom: 1px solid #e8e8e8;
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-inner {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 24px;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  transition: opacity 0.2s;
}

.brand:hover {
  opacity: 0.8;
}

.brand-text {
  font-size: 20px;
  font-weight: 600;
  color: #333;
}

.nav {
  display: flex;
  align-items: center;
  gap: 8px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 6px;
  color: #666;
  text-decoration: none;
  font-size: 14px;
  transition: all 0.2s;
  cursor: pointer;
}

.nav-item:hover {
  background: #f5f5f5;
  color: #333;
}

.nav-item.active {
  background: #ecf5ff;
  color: #409eff;
  font-weight: 500;
}

.user-section {
  display: flex;
  align-items: center;
  gap: 12px;
}

.role-badge {
  margin-right: 4px;
}

.user-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s;
}

.user-trigger:hover {
  background: #f5f5f5;
}

.user-name {
  font-size: 14px;
  color: #333;
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.main {
  flex: 1;
  background: #f0f2f5;
}

/* 响应式 */
@media (max-width: 768px) {
  .header-inner {
    padding: 0 16px;
  }
  
  .nav-item span {
    display: none;
  }
  
  .nav-item {
    padding: 8px;
  }
  
  .user-name {
    display: none;
  }
}
</style>
