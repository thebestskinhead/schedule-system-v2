<template>
  <div class="page-container">
    <!-- 当前周排班概览 -->
    <div class="section-card">
      <div class="card-header">
        <div class="header-left">
          <span class="card-title">第 {{ currentWeek }} 周值班安排</span>
          <van-tag v-if="userStore.isAdmin" type="primary" size="medium">管理员</van-tag>
        </div>
        <van-button 
          v-if="userStore.isAdmin" 
          size="small" 
          type="primary"
          plain
          @click="showWeekDialog = true"
        >
          设置周次
        </van-button>
      </div>

      <van-loading v-if="scheduleLoading" class="loading-center" />

      <van-empty v-else-if="scheduleList.length === 0" description="本周暂无排班安排" />

      <div v-else class="schedule-table-wrapper">
        <div class="schedule-table">
          <!-- 表头 -->
          <div class="table-row header-row">
            <div class="table-cell corner-cell">节次</div>
            <div 
              v-for="day in scheduleDays" 
              :key="day"
              class="table-cell header-cell"
            >
              周{{ weekNames[day - 1] }}
            </div>
          </div>
          <!-- 数据行 -->
          <div 
            v-for="period in 4" 
            :key="period"
            class="table-row"
          >
            <div class="table-cell period-cell">第{{ period }}节</div>
            <div 
              v-for="day in scheduleDays" 
              :key="day"
              class="table-cell data-cell"
              :class="{ 'has-duty': getCellUsers(day, period).length > 0 }"
            >
              <template v-if="getCellUsers(day, period).length > 0">
                <div class="user-names">
                  {{ getCellUsers(day, period).map(u => u.user_name).join('、') }}
                </div>
                <div class="completion-dot" :class="getCellCompletionClass(day, period)"></div>
              </template>
              <span v-else class="empty-text">-</span>
            </div>
          </div>
        </div>
        <!-- 图例 -->
        <div class="table-legend">
          <div class="legend-item">
            <span class="dot completed"></span>
            <span>已完成</span>
          </div>
          <div class="legend-item">
            <span class="dot partial"></span>
            <span>部分完成</span>
          </div>
          <div class="legend-item">
            <span class="dot pending"></span>
            <span>未完成</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 快捷入口 -->
    <div class="section-card quick-section">
      <div class="card-title">快捷入口</div>
      <van-grid :column-num="2" :gutter="12">
        <van-grid-item 
          v-for="item in quickItems" 
          :key="item.path"
          :icon="item.icon"
          :text="item.name"
          @click="$router.push(item.path)"
        >
          <template #icon>
            <div class="quick-icon" :style="{ background: item.color }">
              <van-icon :name="item.icon" size="24" color="#fff" />
            </div>
          </template>
          <template #text>
            <div class="quick-text">
              <div class="quick-name">{{ item.name }}</div>
              <div class="quick-desc">{{ item.desc }}</div>
            </div>
          </template>
        </van-grid-item>
      </van-grid>
    </div>

    <!-- 我的信息 -->
    <div class="section-card">
      <div class="card-title">我的信息</div>
      <van-cell-group inset>
        <van-cell title="姓名" :value="userStore.user?.name" />
        <van-cell title="学号" :value="userStore.user?.student_id" />
        <van-cell title="部门">
          <template #value>
            <van-tag size="medium">{{ userStore.user?.department || '未设置' }}</van-tag>
          </template>
        </van-cell>
        <van-cell title="当前周次" :value="`第 ${currentWeek} 周`" />
        <van-cell title="部门角色">
          <template #value>
            <van-tag :type="userStore.isDeptAdmin ? 'warning' : 'default'" size="medium">
              {{ userStore.isDeptAdmin ? '部门管理员' : '成员' }}
            </van-tag>
          </template>
        </van-cell>
        <van-cell title="系统角色">
          <template #value>
            <van-tag v-if="userStore.isAdmin" type="danger" size="medium">系统管理员</van-tag>
            <van-tag v-else-if="userStore.isOfficeAdmin" type="success" size="medium">办公室管理员</van-tag>
            <van-tag v-else type="default" size="medium">普通用户</van-tag>
          </template>
        </van-cell>
      </van-cell-group>

      <div class="help-link">
        <van-button 
          type="primary" 
          plain 
          block
          @click="$router.push('/readme')"
        >
          查看使用说明
        </van-button>
      </div>
    </div>

    <!-- 设置周次弹窗 -->
    <van-popup 
      v-model:show="showWeekDialog" 
      position="bottom" 
      round
      :style="{ height: '40%' }"
    >
      <div class="popup-container">
        <div class="popup-title">设置当前周次</div>
        
        <van-cell-group inset>
          <van-field label="当前周次">
            <template #input>
              <van-stepper v-model="weekForm.currentWeek" min="1" max="30" />
            </template>
          </van-field>
          <van-cell title="自动递增" label="排班确认后自动进入下一周">
            <template #right-icon>
              <van-switch v-model="weekForm.autoIncrement" size="20" />
            </template>
          </van-cell>
        </van-cell-group>

        <div class="popup-actions">
          <van-button block @click="showWeekDialog = false">取消</van-button>
          <van-button type="primary" block :loading="saving" @click="saveWeek">保存</van-button>
        </div>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { showToast } from 'vant'
import { useUserStore } from '../stores/user'
import { getCurrentWeek, updateCurrentWeek, getSchedule } from '../api/schedule'

const userStore = useUserStore()

const currentWeek = ref(1)
const scheduleList = ref([])
const scheduleLoading = ref(false)
const showWeekDialog = ref(false)
const saving = ref(false)
const weekForm = ref({ currentWeek: 1, autoIncrement: false })

const weekNames = ['一', '二', '三', '四', '五']

// 课表数据：转换为网格结构
const scheduleGrid = computed(() => {
  const grid = {}
  scheduleList.value.forEach(item => {
    const key = `${item.weekday}-${item.period}`
    grid[key] = item.users
  })
  return grid
})

// 有排班的星期列表（只显示有排班的那几天）
const scheduleDays = computed(() => {
  const days = new Set()
  scheduleList.value.forEach(item => {
    days.add(item.weekday)
  })
  return Array.from(days).sort((a, b) => a - b)
})

// 获取某个单元格的用户列表
const getCellUsers = (day, period) => {
  const key = `${day}-${period}`
  return scheduleGrid.value[key] || []
}

// 获取单元格完成状态类名
const getCellCompletionClass = (day, period) => {
  const users = getCellUsers(day, period)
  if (users.length === 0) return ''
  const completed = users.filter(u => u.status === 'completed').length
  if (completed === users.length) return 'completed'
  if (completed > 0) return 'partial'
  return 'pending'
}

// 快捷入口配置 - 根据用户角色显示不同内容
const quickItems = computed(() => {
  const items = [
    { path: '/availability', name: '录入无课表', desc: '设置你的空闲时间', icon: 'calendar-o', color: '#409eff' },
    { path: '/duty/my', name: '我的值班', desc: '查看值班安排', icon: 'clock-o', color: '#67c23a' },
    { path: '/schedule/result', name: '排班结果', desc: '查看完整排班表', icon: 'eye-o', color: '#e6a23c' },
    { path: '/admin/temp-permissions', name: '权限申请', desc: '申请临时权限', icon: 'user-o', color: '#9c27b0' }
  ]
  
  // 部门管理员及以上添加排班管理
  if (userStore.canManageDept) {
    items.push({ path: '/schedule', name: '排班管理', desc: '创建和编辑排班', icon: 'edit', color: '#f56c6c' })
  }
  
  // 办公室管理员和系统管理员添加分工管理
  if (userStore.canManageAll) {
    items.push({ path: '/admin/duty-assignments', name: '每周分工', desc: '管理部门分工安排', icon: 'orders-o', color: '#909399' })
  }
  
  return items
})

const getStatusType = (status) => {
  const map = { pending: 'default', confirmed: 'primary', completed: 'success', cancelled: 'danger' }
  return map[status] || 'default'
}

const loadSchedule = async () => {
  scheduleLoading.value = true
  try {
    const weekRes = await getCurrentWeek()
    currentWeek.value = weekRes?.current_week || 1
    
    const data = await getSchedule({ week: currentWeek.value })
    
    if (data && data.length > 0) {
      const groups = {}
      data.forEach(record => {
        const key = `${record.weekday}-${record.period}`
        if (!groups[key]) {
          groups[key] = {
            weekday: record.weekday,
            period: record.period,
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
  } finally {
    scheduleLoading.value = false
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
    showToast({ message: '设置成功', type: 'success' })
    showWeekDialog.value = false
    loadSchedule()
  } catch {
    showToast({ message: '设置失败', type: 'fail' })
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadSchedule()
})
</script>

<style scoped>
.page-container {
  min-height: 100%;
  background: #f7f8fa;
  padding-top: env(safe-area-inset-top);
  padding-bottom: calc(20px + env(safe-area-inset-bottom));
}

.section-card {
  background: #fff;
  margin: 12px;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.card-header {
  padding: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #f0f0f0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #323233;
  padding: 16px 16px 0;
}

.card-header .card-title {
  padding: 0;
}

.loading-center {
  display: flex;
  justify-content: center;
  padding: 40px;
}

/* 课表样式 */
.schedule-table-wrapper {
  padding: 12px;
}

.schedule-table {
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid #e8e8e8;
}

.table-row {
  display: flex;
}

.table-row.header-row {
  background: #f5f7fa;
}

.table-cell {
  flex: 1;
  min-height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 8px 4px;
  border-right: 1px solid #e8e8e8;
  border-bottom: 1px solid #e8e8e8;
  font-size: 13px;
  text-align: center;
}

.table-cell:last-child {
  border-right: none;
}

.table-row:last-child .table-cell {
  border-bottom: none;
}

.corner-cell {
  background: #f5f7fa;
  font-weight: 500;
  color: #646566;
  width: 50px;
  flex: none;
}

.header-cell {
  font-weight: 500;
  color: #323233;
  min-width: 60px;
}

.period-cell {
  background: #f5f7fa;
  font-weight: 500;
  color: #646566;
  width: 50px;
  flex: none;
}

.data-cell {
  flex-direction: column;
  gap: 4px;
  position: relative;
  background: #fafafa;
}

.data-cell.has-duty {
  background: #e8f4ff;
}

.user-names {
  font-size: 12px;
  color: #323233;
  line-height: 1.4;
  word-break: break-all;
}

.empty-text {
  color: #c8c9cc;
  font-size: 14px;
}

.completion-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  position: absolute;
  top: 4px;
  right: 4px;
}

.completion-dot.completed {
  background: #67c23a;
}

.completion-dot.partial {
  background: #e6a23c;
}

.completion-dot.pending {
  background: #909399;
}

/* 图例 */
.table-legend {
  display: flex;
  justify-content: center;
  gap: 16px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #969799;
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.dot.completed {
  background: #67c23a;
}

.dot.partial {
  background: #e6a23c;
}

.dot.pending {
  background: #909399;
}

/* 快捷入口 */
.quick-section .card-title {
  padding-bottom: 8px;
}

.quick-icon {
  width: 48px;
  height: 48px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.quick-text {
  text-align: center;
  margin-top: 8px;
}

.quick-name {
  font-size: 14px;
  font-weight: 500;
  color: #323233;
}

.quick-desc {
  font-size: 12px;
  color: #969799;
  margin-top: 2px;
}

.help-link {
  padding: 16px;
}

/* 弹窗 */
.popup-container {
  padding: 20px;
}

.popup-title {
  font-size: 18px;
  font-weight: 600;
  text-align: center;
  margin-bottom: 20px;
}

.popup-actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-top: 24px;
}
</style>
