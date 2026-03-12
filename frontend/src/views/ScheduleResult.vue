<template>
  <div class="schedule-result-page">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <h1 class="logo">排班系统 - 排班结果</h1>
          <div class="nav-menu">
            <el-button @click="router.push('/')">返回首页</el-button>
          </div>
        </div>
      </el-header>

      <el-main class="main-content">
        <div class="page-container">
          <el-card>
            <template #header>
              <div class="card-header">
                <span>排班表</span>
                <div class="header-actions">
                  <el-select-v2
                    v-model="selectedWeek"
                    :options="weekOptions"
                    placeholder="选择周次"
                    style="width: 150px; margin-right: 12px"
                    @change="fetchSchedule"
                  />
                  <el-button type="success" @click="exportToExcel">
                    <el-icon><Download /></el-icon> 导出Excel
                  </el-button>
                </div>
              </div>
            </template>

            <el-table 
              id="schedule-table"
              :data="scheduleData" 
              border 
              style="width: 100%"
              v-loading="loading"
            >
              <el-table-column prop="period" label="节次/星期" width="100" fixed>
                <template #default="{ row }">第{{ row.period }}节</template>
              </el-table-column>
              <el-table-column 
                v-for="day in 5" 
                :key="day"
                :label="'周' + ['一','二','三','四','五'][day-1]"
                min-width="150"
              >
                <template #default="{ row }">
                  <div class="duty-cell">
                    <el-tag 
                      v-for="user in getUsersForCell(day, row.period)" 
                      :key="user.user_id"
                      :type="getStatusType(user.status)"
                      size="small"
                      class="user-tag"
                    >
                      {{ user.user_name }}
                    </el-tag>
                    <span v-if="getUsersForCell(day, row.period).length === 0" class="empty">-</span>
                  </div>
                </template>
              </el-table-column>
            </el-table>

            <el-empty v-if="!loading && scheduleData.length === 0" description="暂无排班数据" />
          </el-card>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getSchedule } from '../api/schedule'
import * as XLSX from 'xlsx'

const router = useRouter()
const loading = ref(false)
const selectedWeek = ref(1)
const rawData = ref([])

const weekOptions = computed(() => {
  return Array.from({ length: 30 }, (_, i) => ({
    value: i + 1,
    label: `第${i + 1}周`
  }))
})

const scheduleData = computed(() => {
  return [
    { period: 1 },
    { period: 2 },
    { period: 3 },
    { period: 4 }
  ]
})

const getUsersForCell = (weekday, period) => {
  return rawData.value.filter(item => item.weekday === weekday && item.period === period)
}

const getStatusType = (status) => {
  const map = {
    'pending': 'info',
    'confirmed': 'success',
    'completed': 'primary',
    'cancelled': 'danger'
  }
  return map[status] || 'info'
}

const fetchSchedule = async () => {
  if (!selectedWeek.value) return
  
  loading.value = true
  try {
    const data = await getSchedule(selectedWeek.value)
    rawData.value = data || []
  } catch (error) {
    rawData.value = []
  } finally {
    loading.value = false
  }
}

const exportToExcel = () => {
  const table = document.getElementById('schedule-table')
  if (!table) return

  const wb = XLSX.utils.table_to_book(table, { sheet: `第${selectedWeek.value}周排班表` })
  XLSX.writeFile(wb, `排班表_第${selectedWeek.value}周.xlsx`)
  ElMessage.success('导出成功')
}

onMounted(() => {
  fetchSchedule()
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

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  align-items: center;
}

.duty-cell {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  min-height: 24px;
}

.user-tag {
  margin: 2px;
}

.empty {
  color: #c0c4cc;
}

.main-content {
  min-height: calc(100vh - 60px);
  padding: 20px;
}
</style>
