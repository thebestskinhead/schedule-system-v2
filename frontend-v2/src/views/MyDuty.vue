<template>
  <div class="main-container">
    <div class="card">
      <div class="card-header">
        <span class="card-title">我的值班安排</span>
      </div>
      
      <el-empty v-if="!loading && dutyList.length === 0" description="暂无值班安排" />
      
      <el-table v-else :data="dutyList" v-loading="loading" class="data-table">
        <el-table-column prop="week" label="周次" width="100" sortable />
        <el-table-column label="星期" width="100">
          <template #default="{ row }">周{{ ['一', '二', '三', '四', '五'][row.weekday - 1] }}</template>
        </el-table-column>
        <el-table-column label="节次" width="100">
          <template #default="{ row }">第{{ row.period }}节</template>
        </el-table-column>
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">{{ getStatusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default>
            <el-button 
              type="info" 
              size="small"
              plain
              @click="showDeveloping"
            >操作</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="card">
      <div class="card-header">
        <span class="card-title">状态说明</span>
      </div>
      <div class="status-grid">
        <div class="status-item">
          <el-tag type="info">待确认</el-tag>
          <span>刚安排的值班，等待您确认</span>
        </div>
        <div class="status-item">
          <el-tag type="success">已确认</el-tag>
          <span>已确认参加值班</span>
        </div>
        <div class="status-item">
          <el-tag type="primary">已完成</el-tag>
          <span>值班已完成</span>
        </div>
        <div class="status-item">
          <el-tag type="danger">已取消</el-tag>
          <span>请假或取消值班</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getMyDuties } from '../api/schedule'

const loading = ref(false)
const dutyList = ref([])

const showDeveloping = () => {
  ElMessageBox.alert(
    '确认值班、请假等功能正在开发中，敬请期待。',
    '功能开发中',
    { type: 'info', confirmButtonText: '我知道了' }
  )
}

const getStatusType = (status) => {
  const map = { pending: 'info', confirmed: 'success', completed: 'primary', cancelled: 'danger' }
  return map[status] || 'info'
}

const getStatusText = (status) => {
  const map = { pending: '待确认', confirmed: '已确认', completed: '已完成', cancelled: '已取消' }
  return map[status] || status
}

const fetchData = async () => {
  loading.value = true
  try {
    const data = await getMyDuties()
    dutyList.value = data || []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.status-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: #fafafa;
  border-radius: 8px;
}

.status-item span {
  color: #666;
  font-size: 14px;
}
</style>
