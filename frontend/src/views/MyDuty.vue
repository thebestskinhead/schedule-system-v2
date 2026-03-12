<template>
  <div class="my-duty-page">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <h1 class="logo">排班系统 - 我的值班</h1>
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
                <span>我的值班安排</span>
              </div>
            </template>

            <el-table :data="dutyList" v-loading="loading" style="width: 100%">
              <el-table-column prop="week" label="周次" width="100" sortable />
              <el-table-column prop="weekday" label="星期" width="100">
                <template #default="{ row }">
                  周{{ ['一', '二', '三', '四', '五'][row.weekday - 1] }}
                </template>
              </el-table-column>
              <el-table-column prop="period" label="节次" width="100">
                <template #default="{ row }">
                  第{{ row.period }}节
                </template>
              </el-table-column>
              <el-table-column prop="status" label="状态" width="120">
                <template #default="{ row }">
                  <el-tag :type="getStatusType(row.status)">
                    {{ getStatusText(row.status) }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="200">
                <template #default="{ row }">
                  <el-button 
                    v-if="row.status === 'pending'" 
                    type="success" 
                    size="small"
                    @click="updateStatus(row, 'confirmed')"
                  >
                    确认
                  </el-button>
                  <el-button 
                    v-if="row.status === 'confirmed'" 
                    type="primary" 
                    size="small"
                    @click="updateStatus(row, 'completed')"
                  >
                    标记完成
                  </el-button>
                  <el-button 
                    v-if="row.status !== 'cancelled'" 
                    type="danger" 
                    size="small"
                    @click="updateStatus(row, 'cancelled')"
                  >
                    请假
                  </el-button>
                </template>
              </el-table-column>
            </el-table>

            <el-empty v-if="!loading && dutyList.length === 0" description="暂无值班安排" />
          </el-card>

          <el-card class="mt-4">
            <template #header>
              <div class="card-header">
                <span>状态说明</span>
              </div>
            </template>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="待确认">
                <el-tag type="info">pending</el-tag>
                刚安排的值班，等待确认
              </el-descriptions-item>
              <el-descriptions-item label="已确认">
                <el-tag type="success">confirmed</el-tag>
                已确认参加值班
              </el-descriptions-item>
              <el-descriptions-item label="已完成">
                <el-tag type="primary">completed</el-tag>
                值班已完成
              </el-descriptions-item>
              <el-descriptions-item label="已取消">
                <el-tag type="danger">cancelled</el-tag>
                请假或取消
              </el-descriptions-item>
            </el-descriptions>
          </el-card>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getMyDuties, updateDutyStatus } from '../api/schedule'

const router = useRouter()
const loading = ref(false)
const dutyList = ref([])

const getStatusType = (status) => {
  const map = {
    'pending': 'info',
    'confirmed': 'success',
    'completed': 'primary',
    'cancelled': 'danger'
  }
  return map[status] || 'info'
}

const getStatusText = (status) => {
  const map = {
    'pending': '待确认',
    'confirmed': '已确认',
    'completed': '已完成',
    'cancelled': '已取消'
  }
  return map[status] || status
}

const fetchData = async () => {
  loading.value = true
  try {
    const data = await getMyDuties()
    dutyList.value = data || []
  } catch (error) {
    dutyList.value = []
  } finally {
    loading.value = false
  }
}

const updateStatus = async (row, status) => {
  try {
    await updateDutyStatus({
      duty_id: row.id,
      status: status
    })
    ElMessage.success('状态更新成功')
    fetchData()
  } catch (error) {
    // 错误已在拦截器处理
  }
}

onMounted(() => {
  fetchData()
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

.main-content {
  min-height: calc(100vh - 60px);
  padding: 20px;
}
</style>
