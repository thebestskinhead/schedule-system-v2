<template>
  <div class="page-container">
    <Navbar title="我的值班" show-back />

    <!-- 加载状态 -->
    <van-loading v-if="loading" class="loading-center" />

    <!-- 空状态 -->
    <van-empty v-else-if="dutyList.length === 0" description="暂无值班安排" />

    <!-- 值班列表 -->
    <div v-else class="duty-list">
      <div 
        v-for="item in dutyList" 
        :key="`${item.week}-${item.weekday}-${item.period}`"
        class="duty-card"
      >
        <div class="card-header">
          <div class="time-info">
            <van-tag type="primary" size="medium">第{{ item.week }}周</van-tag>
            <span class="weekday">周{{ weekNames[item.weekday - 1] }}</span>
          </div>
          <van-tag :type="getStatusType(item.status)" size="medium">
            {{ getStatusText(item.status) }}
          </van-tag>
        </div>
        <div class="card-content">
          <div class="period-info">
            <van-icon name="clock-o" />
            <span>第{{ item.period }}节</span>
          </div>
          <van-button 
            type="default" 
            size="small"
            @click="showDeveloping"
          >
            操作
          </van-button>
        </div>
      </div>
    </div>

    <!-- 状态说明 -->
    <div class="status-section">
      <div class="section-title">状态说明</div>
      <van-cell-group inset>
        <van-cell title="待确认" label="刚安排的值班，等待您确认">
          <template #icon>
            <van-tag type="default" class="status-tag">待确认</van-tag>
          </template>
        </van-cell>
        <van-cell title="已确认" label="已确认参加值班">
          <template #icon>
            <van-tag type="success" class="status-tag">已确认</van-tag>
          </template>
        </van-cell>
        <van-cell title="已完成" label="值班已完成">
          <template #icon>
            <van-tag type="primary" class="status-tag">已完成</van-tag>
          </template>
        </van-cell>
        <van-cell title="已取消" label="请假或取消值班">
          <template #icon>
            <van-tag type="danger" class="status-tag">已取消</van-tag>
          </template>
        </van-cell>
      </van-cell-group>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { showToast, showDialog } from 'vant'
import { getMyDuties } from '../api/schedule'
import Navbar from '../components/Navbar.vue'

const loading = ref(false)
const dutyList = ref([])

const weekNames = ['一', '二', '三', '四', '五']

const showDeveloping = () => {
  showDialog({
    title: '功能开发中',
    message: '确认值班、请假等功能正在开发中，敬请期待。',
    confirmButtonText: '我知道了'
  })
}

const getStatusType = (status) => {
  const map = { pending: 'default', confirmed: 'success', completed: 'primary', cancelled: 'danger' }
  return map[status] || 'default'
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
.page-container {
  min-height: 100%;
  background: #f7f8fa;
  padding-top: env(safe-area-inset-top);
  padding-bottom: calc(20px + env(safe-area-inset-bottom));
}

.loading-center {
  display: flex;
  justify-content: center;
  padding: 40px;
}

.duty-list {
  padding: 12px;
}

.duty-card {
  background: #fff;
  border-radius: 12px;
  margin-bottom: 12px;
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

.time-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.weekday {
  font-size: 16px;
  font-weight: 500;
  color: #323233;
}

.card-content {
  padding: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.period-info {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #646566;
  font-size: 14px;
}

.status-section {
  padding: 0 12px;
  margin-top: 8px;
}

.section-title {
  font-size: 14px;
  color: #969799;
  padding: 16px 4px 8px;
}

.status-tag {
  margin-right: 12px;
}
</style>
