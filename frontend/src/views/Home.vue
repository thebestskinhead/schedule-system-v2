<template>
  <div class="home">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <h1 class="logo">排班系统</h1>
          <div class="nav-menu">
            <el-menu mode="horizontal" :router="true" :default-active="$route.path">
              <el-menu-item index="/">首页</el-menu-item>
              <el-menu-item index="/availability">我的无课表</el-menu-item>
              <el-menu-item index="/duty/my">我的值班</el-menu-item>
              <el-menu-item index="/schedule/result">排班结果</el-menu-item>
              <el-sub-menu index="/admin" v-if="userStore.canManageDept">
                <template #title>管理</template>
                <el-menu-item index="/schedule" v-if="userStore.canManageDept">排班管理</el-menu-item>
                <el-menu-item index="/admin/duty-assignments" v-if="userStore.canManageAll">每周分工</el-menu-item>
                <el-menu-item index="/admin/users/v2" v-if="userStore.canManageAll">用户管理</el-menu-item>
                <el-menu-item index="/admin/temp-permissions" v-if="userStore.isAdmin">临时权限</el-menu-item>
                <el-menu-item index="/admin/smtp" v-if="userStore.isAdmin">SMTP配置</el-menu-item>
              </el-sub-menu>
              <el-menu-item index="/my-permissions">我的权限</el-menu-item>
            </el-menu>
            <div class="user-info">
              <span>{{ userStore.user?.name }}</span>
              <el-button type="danger" size="small" @click="logout">退出</el-button>
            </div>
          </div>
        </div>
      </el-header>
      
      <el-main class="main-content">
        <div class="page-container">
          <!-- 当前周排班情况 -->
          <el-card class="current-week-card">
            <template #header>
              <div class="card-header">
                <div class="header-title">
                  <span>第 {{ currentWeek }} 周值班安排</span>
                  <el-tag v-if="userStore.isAdmin" type="primary" size="small" class="admin-tag">
                    管理员可设置当前周次
                  </el-tag>
                </div>
                <div class="header-actions">
                  <el-button v-if="userStore.isAdmin" type="primary" size="small" @click="showWeekDialog = true">
                    <el-icon><Setting /></el-icon>
                    设置当前周
                  </el-button>
                  <el-button type="success" size="small" @click="goTo('/schedule/result')">
                    查看完整排班
                  </el-button>
                </div>
              </div>
            </template>
            
            <div v-if="currentSchedule.length === 0" class="empty-schedule">
              <el-empty description="暂无排班数据">
                <template #description>
                  <p>第 {{ currentWeek }} 周暂无排班安排</p>
                  <p v-if="userStore.isAdmin" class="hint">请前往排班管理页面生成排班</p>
                </template>
              </el-empty>
            </div>
            
            <el-table v-else :data="groupedSchedule" border stripe>
              <el-table-column prop="weekdayText" label="星期" width="100" />
              <el-table-column prop="period" label="节次" width="80">
                <template #default="{ row }">第{{ row.period }}节</template>
              </el-table-column>
              <el-table-column label="值班人员">
                <template #default="{ row }">
                  <el-tag 
                    v-for="user in row.users" 
                    :key="user.user_id"
                    size="small"
                    :type="getDutyTagType(user.status)"
                    class="user-tag"
                  >
                    {{ user.user_name }}
                    <span v-if="user.status === 'completed'" class="status-icon">✓</span>
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="状态" width="120">
                <template #default="{ row }">
                  <el-progress 
                    :percentage="getCompletionRate(row)" 
                    :status="getCompletionRate(row) === 100 ? 'success' : ''"
                    :stroke-width="8"
                  />
                </template>
              </el-table-column>
            </el-table>
          </el-card>

          <el-row :gutter="20" class="mt-4">
            <el-col :span="8">
              <el-card>
                <template #header>
                  <div class="card-header">
                    <span>我的信息</span>
                  </div>
                </template>
                <div class="welcome-content">
                  <p><strong>姓名：</strong>{{ userStore.user?.name }}</p>
                  <p><strong>学号：</strong>{{ userStore.user?.student_id }}</p>
                  <p>
                    <strong>部门：</strong>
                    <el-tag size="small">{{ userStore.user?.department || '未设置' }}</el-tag>
                  </p>
                  <p>
                    <strong>部门角色：</strong>
                    <el-tag size="small" :type="userStore.isDeptAdmin ? 'warning' : 'info'">
                      {{ userStore.isDeptAdmin ? '部门管理员' : '部门成员' }}
                    </el-tag>
                  </p>
                  <p>
                    <strong>系统角色：</strong>
                    <el-tag size="small" :type="userStore.isAdmin ? 'danger' : 'success'">
                      {{ userStore.isAdmin ? '系统管理员' : '普通用户' }}
                    </el-tag>
                  </p>
                  <p><strong>当前周次：</strong>第 {{ currentWeek }} 周</p>
                </div>
              </el-card>
            </el-col>
            <el-col :span="16">
              <el-card>
                <template #header>
                  <div class="card-header">
                    <span>快速操作</span>
                  </div>
                </template>
                <div class="quick-actions">
                  <el-button type="primary" @click="goTo('/availability')">
                    <el-icon><Calendar /></el-icon>
                    录入无课表
                  </el-button>
                  <el-button type="success" @click="goTo('/duty/my')">
                    <el-icon><Timer /></el-icon>
                    我的值班
                  </el-button>
                  <el-button type="warning" @click="goTo('/schedule/result')">
                    <el-icon><View /></el-icon>
                    排班结果
                  </el-button>
                  <el-button type="info" @click="goTo('/my-permissions')">
                    <el-icon><UserFilled /></el-icon>
                    我的权限
                  </el-button>
                  <el-button v-if="userStore.canManageDept" type="danger" @click="goTo('/schedule')">
                    <el-icon><Setting /></el-icon>
                    排班管理
                  </el-button>
                </div>
              </el-card>
            </el-col>
          </el-row>
        </div>
      </el-main>
    </el-container>

    <!-- 设置当前周次对话框 -->
    <el-dialog v-model="showWeekDialog" title="设置当前周次" width="400px">
      <el-form :model="weekForm" label-width="120px">
        <el-form-item label="当前周次">
          <el-input-number v-model="weekForm.currentWeek" :min="1" :max="30" />
        </el-form-item>
        <el-form-item label="自动递增">
          <el-switch v-model="weekForm.autoIncrement" />
          <div class="form-hint">开启后，排班确认时自动进入下一周</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showWeekDialog = false">取消</el-button>
        <el-button type="primary" @click="saveCurrentWeek" :loading="savingWeek">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Calendar, Timer, View, Setting, UserFilled } from '@element-plus/icons-vue'
import { useUserStore } from '../stores/user'
import { getCurrentWeek, updateCurrentWeek } from '../api/schedule'

const router = useRouter()
const userStore = useUserStore()

const currentWeek = ref(1)
const currentSchedule = ref([])
const loading = ref(false)
const showWeekDialog = ref(false)
const savingWeek = ref(false)

const weekForm = ref({
  currentWeek: 1,
  autoIncrement: false
})

// 按星期和节次分组排班数据
const groupedSchedule = computed(() => {
  const groups = {}
  
  currentSchedule.value.forEach(record => {
    const key = `${record.weekday}-${record.period}`
    if (!groups[key]) {
      groups[key] = {
        weekday: record.weekday,
        period: record.period,
        weekdayText: '周' + ['一', '二', '三', '四', '五'][record.weekday - 1],
        users: []
      }
    }
    groups[key].users.push({
      user_id: record.user_id,
      user_name: record.user_name,
      status: record.status
    })
  })
  
  // 按星期和节次排序
  return Object.values(groups).sort((a, b) => {
    if (a.weekday !== b.weekday) return a.weekday - b.weekday
    return a.period - b.period
  })
})

const goTo = (path) => {
  router.push(path)
}

const logout = () => {
  userStore.logout()
  ElMessage.success('已退出登录')
  router.push('/login')
}

const getDutyTagType = (status) => {
  const types = {
    'pending': 'info',
    'confirmed': 'primary',
    'completed': 'success',
    'cancelled': 'danger'
  }
  return types[status] || 'info'
}

const getCompletionRate = (row) => {
  if (row.users.length === 0) return 0
  const completed = row.users.filter(u => u.status === 'completed').length
  return Math.round((completed / row.users.length) * 100)
}

// 加载当前周次和排班
const loadCurrentSchedule = async () => {
  loading.value = true
  try {
    const weekRes = await getCurrentWeek()
    currentWeek.value = weekRes.data.current_week || 1
    
    // 获取当前周的排班数据
    const scheduleRes = await fetch(`/api/v1/schedule?week=${currentWeek.value}`, {
      headers: {
        'Authorization': `Bearer ${userStore.token}`
      }
    }).then(r => r.json())
    
    if (scheduleRes.code === 200) {
      currentSchedule.value = scheduleRes.data || []
    }
  } catch (error) {
    console.error('加载排班失败:', error)
  } finally {
    loading.value = false
  }
}

// 打开设置对话框
const openWeekDialog = async () => {
  try {
    const res = await getCurrentWeek()
    weekForm.value.currentWeek = res.data.current_week || 1
    weekForm.value.autoIncrement = res.data.auto_increment || false
    showWeekDialog.value = true
  } catch (error) {
    ElMessage.error('获取当前周次失败')
  }
}

// 保存当前周次
const saveCurrentWeek = async () => {
  savingWeek.value = true
  try {
    await updateCurrentWeek({
      current_week: weekForm.value.currentWeek,
      auto_increment: weekForm.value.autoIncrement
    })
    ElMessage.success('设置成功')
    showWeekDialog.value = false
    currentWeek.value = weekForm.value.currentWeek
    loadCurrentSchedule()
  } catch (error) {
    ElMessage.error('设置失败')
  } finally {
    savingWeek.value = false
  }
}

onMounted(() => {
  loadCurrentSchedule()
})
</script>

<style scoped>
.header {
  background: #fff;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  padding: 0;
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
  padding: 0 20px;
}

.logo {
  margin: 0;
  font-size: 20px;
  color: #409eff;
}

.nav-menu {
  display: flex;
  align-items: center;
  gap: 20px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.main-content {
  min-height: calc(100vh - 60px);
  padding: 20px;
}

.page-container {
  max-width: 1200px;
  margin: 0 auto;
}

.current-week-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  font-weight: bold;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.admin-tag {
  font-weight: normal;
}

.empty-schedule {
  padding: 40px 0;
}

.hint {
  color: #909399;
  font-size: 12px;
  margin-top: 8px;
}

.user-tag {
  margin-right: 8px;
  margin-bottom: 4px;
}

.status-icon {
  margin-left: 4px;
  color: #67c23a;
}

.welcome-content p {
  margin: 8px 0;
  line-height: 1.6;
}

.quick-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.quick-actions .el-button {
  display: flex;
  align-items: center;
  gap: 4px;
}

.form-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.mt-4 {
  margin-top: 16px;
}
</style>
