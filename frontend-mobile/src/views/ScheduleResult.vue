<template>
  <div class="page-container">
    <!-- 顶部筛选 -->
    <div class="filter-bar">
      <van-dropdown-menu>
        <van-dropdown-item v-model="selectedWeek" :options="weekOptions" @change="loadSchedule" />
      </van-dropdown-menu>
    </div>

    <!-- 加载状态 -->
    <van-loading v-if="loading" class="loading-center" />

    <!-- 空状态 -->
    <van-empty v-else-if="scheduleData.length === 0" description="该周暂无排班数据" />

    <!-- 排班列表 -->
    <div v-else class="schedule-list">
      <div 
        v-for="item in scheduleData" 
        :key="`${item.weekday}-${item.period}`"
        class="schedule-card"
      >
        <div class="card-header">
          <div class="time-info">
            <van-tag type="primary" size="medium">周{{ weekNames[item.weekday - 1] }}</van-tag>
            <span class="period">第{{ item.period }}节</span>
          </div>
        </div>
        <div class="card-content">
          <div class="users-label">值班人员</div>
          <div class="users-list">
            <van-tag 
              v-for="user in item.users" 
              :key="user.user_id"
              :type="getStatusType(user.status)"
              size="large"
              class="user-tag"
            >
              {{ user.user_name }}
            </van-tag>
            <van-empty v-if="item.users.length === 0" description="暂无人值班" :image-size="60" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getSchedule, getCurrentWeek } from '../api/schedule'

const loading = ref(false)
const selectedWeek = ref(1)
const scheduleData = ref([])

const weekNames = ['一', '二', '三', '四', '五']

const weekOptions = computed(() => {
  return Array.from({ length: 30 }, (_, i) => ({
    text: `第${i + 1}周`,
    value: i + 1
  }))
})

const getStatusType = (status) => {
  const map = { pending: 'default', confirmed: 'success', completed: 'primary', cancelled: 'danger' }
  return map[status] || 'default'
}

const loadSchedule = async () => {
  loading.value = true
  try {
    const data = await getSchedule({ week: selectedWeek.value })
    if (data) {
      // 按时段分组
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
      scheduleData.value = Object.values(groups).sort((a, b) => {
        if (a.weekday !== b.weekday) return a.weekday - b.weekday
        return a.period - b.period
      })
    }
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  try {
    const res = await getCurrentWeek()
    selectedWeek.value = res?.current_week || 1
    loadSchedule()
  } catch {
    loadSchedule()
  }
})
</script>

<style scoped>
.page-container {
  min-height: 100%;
  background: #f7f8fa;
  padding-top: env(safe-area-inset-top);
  padding-bottom: calc(20px + env(safe-area-inset-bottom));
}

.filter-bar {
  position: sticky;
  top: env(safe-area-inset-top);
  z-index: 100;
}

.loading-center {
  display: flex;
  justify-content: center;
  padding: 40px;
}

.schedule-list {
  padding: 12px;
}

.schedule-card {
  background: #fff;
  border-radius: 12px;
  margin-bottom: 12px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.card-header {
  padding: 16px 16px 12px;
  border-bottom: 1px solid #f0f0f0;
}

.time-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.period {
  font-size: 16px;
  font-weight: 500;
  color: #323233;
}

.card-content {
  padding: 16px;
}

.users-label {
  font-size: 13px;
  color: #969799;
  margin-bottom: 12px;
}

.users-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.user-tag {
  padding: 4px 12px;
}
</style>
