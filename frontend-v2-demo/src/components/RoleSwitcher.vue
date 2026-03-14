<template>
  <el-dropdown v-if="isMockMode" placement="bottom">
    <el-button type="warning" size="small" class="role-switch-btn">
      <el-icon><Switch /></el-icon>
      <span>切换角色</span>
    </el-button>
    <template #dropdown>
      <el-dropdown-menu class="role-dropdown-menu">
        <el-dropdown-item 
          v-for="user in presetUsers" 
          :key="user.id"
          :class="{ 'is-active': currentUserId === user.id }"
          @click="switchTo(user.id)"
        >
          <div class="role-item">
            <el-avatar :size="28" :src="user.avatar" class="role-avatar" />
            <div class="role-info">
              <span class="role-name">{{ user.name }}</span>
              <span class="role-id">{{ user.student_id }}</span>
            </div>
            <el-tag 
              size="small" 
              :type="getRoleType(user)"
              class="role-tag"
            >
              {{ getRoleLabel(user) }}
            </el-tag>
          </div>
        </el-dropdown-item>
        
        <el-dropdown-item divided @click="handleReset">
          <el-icon><RefreshRight /></el-icon>
          <span>重置数据</span>
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { mockAuth } from '../../mock/index.js'
import { useUserStore } from '../stores/user'

const userStore = useUserStore()

// Mock模式标志 - 始终启用用于演示
const isMockMode = ref(true)

// 当前用户ID
const currentUserId = computed(() => userStore.user?.id)

// 预设用户列表
const presetUsers = computed(() => mockAuth.getPresetUsers())

// 获取角色类型（用于标签颜色）
const getRoleType = (user) => {
  if (user.role === 'admin') return 'danger'
  if (user.department === '办公室' && user.dept_role === 'dept_admin') return 'success'
  if (user.dept_role === 'dept_admin') return 'warning'
  return 'info'
}

// 获取角色标签文字
const getRoleLabel = (user) => {
  if (user.role === 'admin') return '系统管理员'
  if (user.department === '办公室' && user.dept_role === 'dept_admin') return '办公室管理'
  if (user.dept_role === 'dept_admin') return '部门管理'
  return '普通成员'
}

// 切换到指定用户
const switchTo = async (userId) => {
  if (userId === currentUserId.value) {
    ElMessage.info('当前已是该角色')
    return
  }
  
  try {
    const user = presetUsers.value.find(u => u.id === userId)
    await ElMessageBox.confirm(
      `确定要切换到 <strong>${user.name}</strong> 吗？<br>角色: ${getRoleLabel(user)}`,
      '切换角色',
      {
        confirmButtonText: '切换',
        cancelButtonText: '取消',
        type: 'warning',
        dangerouslyUseHTMLString: true
      }
    )
    
    await mockAuth.switchRole(userId)
    
    // 更新store中的用户信息
    await userStore.loadUserInfo()
    
    ElMessage.success(`已切换为 ${user.name}`)
    
    // 如果当前页面有权限限制，可能需要重定向
    setTimeout(() => {
      window.location.reload()
    }, 500)
  } catch (error) {
    if (error !== 'cancel') {
      console.error('切换角色失败:', error)
      ElMessage.error('切换失败: ' + error.message)
    }
  }
}

// 重置Mock数据
const handleReset = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要重置所有Mock数据吗？这将清除您的所有操作记录。',
      '重置数据',
      {
        confirmButtonText: '重置',
        cancelButtonText: '取消',
        type: 'danger'
      }
    )
    
    await mockAuth.logout()
    
    ElMessage.success('数据已重置，请重新登录')
    window.location.href = '/login'
  } catch (error) {
    if (error !== 'cancel') {
      console.error('重置失败:', error)
    }
  }
}

onMounted(() => {
  // 可以在这里添加初始化逻辑
})
</script>

<style scoped>
.role-switch-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-right: 12px;
}

.role-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 4px 0;
  min-width: 200px;
}

.role-avatar {
  flex-shrink: 0;
}

.role-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.role-name {
  font-size: 14px;
  font-weight: 500;
  color: #333;
}

.role-id {
  font-size: 12px;
  color: #999;
}

.role-tag {
  flex-shrink: 0;
}
</style>

<style>
/* 下拉菜单样式 */
.role-dropdown-menu .el-dropdown-menu__item {
  padding: 8px 16px;
}

.role-dropdown-menu .el-dropdown-menu__item.is-active {
  background-color: #ecf5ff;
  color: #409eff;
}

.role-dropdown-menu .el-dropdown-menu__item.is-active .role-name {
  color: #409eff;
}
</style>
