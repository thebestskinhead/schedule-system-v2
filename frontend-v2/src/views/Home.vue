<template>
  <div class="main-container">
    <!-- 当前周排班概览 -->
    <div class="card">
      <div class="card-header">
        <div class="title-section">
          <span class="card-title">第 {{ currentWeek }} 周值班安排</span>
          <el-tag v-if="userStore.isAdmin" type="info" size="small">管理员</el-tag>
        </div>
        <div class="action-bar" v-if="userStore.isAdmin">
          <el-button size="small" @click="showWeekDialog = true">
            <el-icon><Setting /></el-icon> 设置周次
          </el-button>
        </div>
      </div>
      
      <el-empty v-if="scheduleList.length === 0" description="本周暂无排班安排" />
      
      <el-table v-else :data="scheduleList" class="data-table">
        <el-table-column prop="weekdayText" label="星期" width="100" />
        <el-table-column prop="period" label="节次" width="100">
          <template #default="{ row }">第{{ row.period }}节</template>
        </el-table-column>
        <el-table-column label="值班人员">
          <template #default="{ row }">
            <el-space wrap>
              <el-tag 
                v-for="user in row.users" 
                :key="user.user_id"
                :type="getStatusType(user.status)"
                size="small"
              >
                {{ user.user_name }}
                <el-icon v-if="user.status === 'completed'" :size="10"><Check /></el-icon>
              </el-tag>
            </el-space>
          </template>
        </el-table-column>
        <el-table-column label="完成度" width="120">
          <template #default="{ row }">
            <el-progress 
              :percentage="getCompletionRate(row)" 
              :status="getCompletionRate(row) === 100 ? 'success' : ''"
              :stroke-width="6"
              :show-text="false"
            />
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 快捷入口 -->
    <div class="quick-grid">
      <div 
        v-for="item in quickItems" 
        :key="item.path"
        class="quick-card"
        @click="$router.push(item.path)"
      >
        <div class="quick-icon" :style="{ background: item.color }">
          <el-icon :size="24" color="#fff"><component :is="item.icon" /></el-icon>
        </div>
        <div class="quick-info">
          <div class="quick-title">{{ item.name }}</div>
          <div class="quick-desc">{{ item.desc }}</div>
        </div>
        <el-icon class="quick-arrow"><ArrowRight /></el-icon>
      </div>
    </div>

    <!-- 我的信息 -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">我的信息</span>
      </div>
      <el-descriptions :column="3" border>
        <el-descriptions-item label="姓名">{{ userStore.user?.name }}</el-descriptions-item>
        <el-descriptions-item label="学号">{{ userStore.user?.student_id }}</el-descriptions-item>
        <el-descriptions-item label="部门">
          <el-tag size="small">{{ userStore.user?.department || '未设置' }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="当前周次">第 {{ currentWeek }} 周</el-descriptions-item>
        <el-descriptions-item label="部门角色">
          <el-tag size="small" :type="userStore.isDeptAdmin ? 'warning' : 'info'">
            {{ userStore.isDeptAdmin ? '部门管理员' : '成员' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="系统角色">
          <el-tag v-if="userStore.isAdmin" type="danger" size="small">系统管理员</el-tag>
          <el-tag v-else-if="userStore.isOfficeAdmin" type="success" size="small">办公室管理员</el-tag>
          <el-tag v-else type="info" size="small">普通用户</el-tag>
        </el-descriptions-item>
      </el-descriptions>
      
      <div class="help-link">
        <el-link type="primary" @click="$router.push('/readme')">
          <el-icon><Document /></el-icon> 查看使用说明
        </el-link>
      </div>
    </div>

    <!-- 设置周次对话框 -->
    <el-dialog v-model="showWeekDialog" title="设置当前周次" width="400px">
      <el-form :model="weekForm" label-width="100px">
        <el-form-item label="当前周次">
          <el-input-number v-model="weekForm.currentWeek" :min="1" :max="30" />
        </el-form-item>
        <el-form-item label="自动递增">
          <el-switch v-model="weekForm.autoIncrement" />
          <div class="form-hint">排班确认后自动进入下一周</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showWeekDialog = false">取消</el-button>
        <el-button type="primary" @click="saveWeek" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useUserStore } from '../stores/user'
import { getCurrentWeek, updateCurrentWeek, getSchedule } from '../api/schedule'

const userStore = useUserStore()

const currentWeek = ref(1)
const scheduleList = ref([])
const showWeekDialog = ref(false)
const saving = ref(false)
const weekForm = ref({ currentWeek: 1, autoIncrement: false })

// 快捷入口配置 - 根据用户角色显示不同内容
const quickItems = computed(() => {
  const items = [
    { path: '/availability', name: '录入无课表', desc: '设置你的空闲时间', icon: 'Calendar', color: '#409eff' },
    { path: '/duty/my', name: '我的值班', desc: '查看值班安排', icon: 'Timer', color: '#67c23a' },
    { path: '/schedule/result', name: '排班结果', desc: '查看完整排班表', icon: 'View', color: '#e6a23c' }
  ]
  
  // 部门管理员及以上添加排班管理
  if (userStore.canManageDept) {
    items.push({ path: '/schedule', name: '排班管理', desc: '创建和编辑排班', icon: 'Edit', color: '#f56c6c' })
  }
  
  // 办公室管理员和系统管理员添加分工管理
  if (userStore.canManageAll) {
    items.push({ path: '/admin/duty-assignments', name: '每周分工', desc: '管理部门分工安排', icon: 'List', color: '#909399' })
  }
  
  return items
})

const getStatusType = (status) => {
  const map = { pending: 'info', confirmed: 'primary', completed: 'success', cancelled: 'danger' }
  return map[status] || 'info'
}

const getCompletionRate = (row) => {
  if (!row.users?.length) return 0
  const completed = row.users.filter(u => u.status === 'completed').length
  return Math.round((completed / row.users.length) * 100)
}

const loadSchedule = async () => {
  try {
    const weekRes = await getCurrentWeek()
    // 拦截器已提取 data，直接访问
    currentWeek.value = weekRes?.current_week || 1
    
    // 使用封装的 request，统一响应格式处理
    const data = await getSchedule({ week: currentWeek.value })
    
    if (data && data.length > 0) {
      const groups = {}
      data.forEach(record => {
        const key = `${record.weekday}-${record.period}`
        if (!groups[key]) {
          groups[key] = {
            weekday: record.weekday,
            period: record.period,
            weekdayText: '周' + ['一', '二', '三', '四', '五'][record.weekday - 1],
            users: []
          }
        }
        groups[key].users.push(record)
      })
      scheduleList.value = Object.values(groups).sort((a, b) => {
        if (a.weekday !== b.weekday) return a.weekday - b.weekday
        return a.period - b.period
      })
    } else {
      scheduleList.value = []
    }
  } catch (error) {
    console.error('加载排班失败:', error)
  }
}

const saveWeek = async () => {
  saving.value = true
  try {
    await updateCurrentWeek({
      current_week: weekForm.value.currentWeek,
      auto_increment: weekForm.value.autoIncrement
    })
    currentWeek.value = weekForm.value.currentWeek
    ElMessage.success('设置成功')
    showWeekDialog.value = false
    loadSchedule()
  } catch {
    ElMessage.error('设置失败')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadSchedule()
})
</script>

<style scoped>
.title-section {
  display: flex;
  align-items: center;
  gap: 10px;
}

.quick-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}

.quick-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.quick-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.quick-icon {
  width: 48px;
  height: 48px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.quick-info {
  flex: 1;
}

.quick-title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
  margin-bottom: 4px;
}

.quick-desc {
  font-size: 13px;
  color: #999;
}

.quick-arrow {
  color: #ccc;
}

.form-hint {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

.help-link {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px dashed #e4e7ed;
  text-align: center;
}
</style>
