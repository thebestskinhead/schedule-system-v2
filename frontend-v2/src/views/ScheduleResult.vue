<template>
  <div class="main-container">
    <div class="card">
      <div class="card-header">
        <span class="card-title">排班结果</span>
        <div class="action-bar">
          <el-select v-model="selectedWeek" placeholder="选择周次" style="width: 120px">
            <el-option v-for="w in 30" :key="w" :label="`第${w}周`" :value="w" />
          </el-select>
          <el-button type="primary" @click="loadSchedule">
            <el-icon><Refresh /></el-icon> 查询
          </el-button>
        </div>
      </div>
      
      <el-empty v-if="!loading && scheduleData.length === 0" description="该周暂无排班数据" />
      
      <el-table v-else :data="scheduleData" v-loading="loading" class="data-table">
        <el-table-column label="星期" width="100">
          <template #default="{ row }">周{{ ['一', '二', '三', '四', '五'][row.weekday - 1] }}</template>
        </el-table-column>
        <el-table-column label="节次" width="100">
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
              </el-tag>
            </el-space>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getSchedule, getCurrentWeek } from '../api/schedule'

const loading = ref(false)
const selectedWeek = ref(1)
const scheduleData = ref([])

const getStatusType = (status) => {
  const map = { pending: 'info', confirmed: 'success', completed: 'primary', cancelled: 'danger' }
  return map[status] || 'info'
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
    // 拦截器已提取 data，直接访问
    selectedWeek.value = res?.current_week || 1
    loadSchedule()
  } catch {
    loadSchedule()
  }
})
</script>
