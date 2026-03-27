<template>
  <div class="main-container">
    <Navbar title="每周值班分工管理" show-back />
    <div class="card">
      <div class="card-header">
        <span class="card-title">每周值班分工管理</span>
        <van-button type="primary" size="small" @click="saveData" :loading="saving" :disabled="!hasChanges">
          <van-icon name="success" /> 保存
        </van-button>
      </div>

      <!-- 周次选择 -->
      <div class="week-selector">
        <van-field
          v-model="weekLabel"
          is-link
          readonly
          label="选择周次"
          placeholder="选择周次"
          @click="showWeekPicker = true"
        />
      </div>

      <!-- 值班分工表格 - 卡片列表形式 -->
      <div class="duty-list">
        <van-loading v-if="loading" style="text-align: center; padding: 20px;" />
        <template v-else>
          <div v-for="dept in departments" :key="dept" class="duty-card">
            <div class="duty-card-header">{{ dept }}</div>
            <div class="duty-card-body">
              <div v-for="day in weekdays" :key="day.value" class="duty-day-row">
                <span class="day-label">{{ day.label }}</span>
                <van-switch
                  :model-value="isAssigned(dept, day.value)"
                  size="22px"
                  @update:model-value="(val) => onSwitchChange(dept, day.value, val)"
                />
              </div>
            </div>
          </div>
        </template>
      </div>

      <!-- 本周值班预览 -->
      <div class="preview-section">
        <h3>本周值班预览</h3>
        <div class="preview-grid">
          <div v-for="dept in departments" :key="dept" class="preview-card">
            <div class="dept-name">{{ dept }}</div>
            <div class="day-tags">
              <van-tag
                v-for="day in weekdays"
                :key="day.value"
                :type="isAssigned(dept, day.value) ? 'success' : 'default'"
                size="medium"
                class="day-tag"
              >
                {{ day.label.slice(0, 2) }}
              </van-tag>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 周次选择弹窗 -->
    <van-popup v-model:show="showWeekPicker" position="bottom" round>
      <van-picker
        :columns="weekColumns"
        @confirm="onWeekConfirm"
        @cancel="showWeekPicker = false"
      />
    </van-popup>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { showToast } from 'vant'
import Navbar from '../components/Navbar.vue'
import { getDutyAssignments, saveDutyAssignments, getCurrentWeek } from '../api/schedule'

const currentWeek = ref(1)
const assignments = ref({})
const originalAssignments = ref({})
const saving = ref(false)
const loading = ref(false)
const showWeekPicker = ref(false)
const departments = ['办公室', '竞赛部', '项目部', '科普部']

const weekdays = [
  { label: '周一', value: 1 },
  { label: '周二', value: 2 },
  { label: '周三', value: 3 },
  { label: '周四', value: 4 },
  { label: '周五', value: 5 },
]

const weekColumns = Array.from({ length: 20 }, (_, i) => ({
  text: `第 ${i + 1} 周`,
  value: i + 1
}))

const weekLabel = computed(() => `第 ${currentWeek.value} 周`)

const onWeekConfirm = ({ selectedValues }) => {
  currentWeek.value = selectedValues[0]
  showWeekPicker.value = false
  loadData()
}

const tableData = computed(() => {
  return departments.map(dept => {
    const row = { department: dept }
    weekdays.forEach(day => {
      row[`day${day.value}`] = isAssigned(dept, day.value)
    })
    return row
  })
})

const hasChanges = computed(() => {
  return JSON.stringify(assignments.value) !== JSON.stringify(originalAssignments.value)
})

const isAssigned = (dept, weekday) => {
  return assignments.value[`${dept}-${weekday}`] || false
}

const onSwitchChange = (dept, weekday, val) => {
  const key = `${dept}-${weekday}`
  assignments.value[key] = val
}

const loadData = async () => {
  loading.value = true
  try {
    const data = await getDutyAssignments({ week: currentWeek.value })
    if (data && data.assignments) {
      const newAssignments = {}
      data.assignments.forEach(item => {
        newAssignments[`${item.department}-${item.weekday}`] = item.is_assigned
      })
      assignments.value = newAssignments
      originalAssignments.value = { ...newAssignments }
    } else {
      assignments.value = {}
      originalAssignments.value = {}
    }
  } catch (err) {
    console.error('加载分工失败:', err)
    showToast({ message: '加载分工失败', type: 'fail' })
  } finally {
    loading.value = false
  }
}

const saveData = async () => {
  saving.value = true
  try {
    const payload = {
      week: currentWeek.value,
      assignments: []
    }

    departments.forEach(dept => {
      weekdays.forEach(day => {
        payload.assignments.push({
          department: dept,
          weekday: day.value,
          is_assigned: !!assignments.value[`${dept}-${day.value}`]
        })
      })
    })

    await saveDutyAssignments(payload)
    originalAssignments.value = { ...assignments.value }
    showToast({ message: '保存成功！', type: 'success' })
  } catch (err) {
    showToast({ message: '保存失败: ' + (err.response?.data?.message || err.message), type: 'fail' })
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  try {
    const res = await getCurrentWeek()
    const week = res?.current_week
    if (week) {
      currentWeek.value = week
    }
  } catch (e) {
    console.error('获取当前周次失败:', e)
  }
  loadData()
})
</script>

<style scoped>
.main-container {
  min-height: 100%;
  background: #f7f8fa;
  padding-top: env(safe-area-inset-top);
  padding-bottom: calc(20px + env(safe-area-inset-bottom));
}

.card {
  margin: 12px;
  padding: 16px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #323233;
}

.week-selector {
  margin-bottom: 16px;
}

.duty-list {
  margin-bottom: 20px;
}

.duty-card {
  background: #fff;
  border-radius: 8px;
  margin-bottom: 12px;
  overflow: hidden;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
}

.duty-card-header {
  background: #f7f8fa;
  padding: 10px 16px;
  font-weight: 600;
  font-size: 15px;
  color: #323233;
  border-bottom: 1px solid #ebedf0;
}

.duty-card-body {
  padding: 8px 16px;
}

.duty-day-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid #f5f5f5;
}

.duty-day-row:last-child {
  border-bottom: none;
}

.day-label {
  font-size: 14px;
  color: #323233;
}

.preview-section {
  margin-top: 24px;
}

.preview-section h3 {
  margin-bottom: 16px;
  color: #303133;
  font-size: 16px;
}

.preview-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.preview-card {
  background: #fff;
  border-radius: 8px;
  padding: 12px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
}

.dept-name {
  font-weight: bold;
  text-align: center;
  margin-bottom: 10px;
  font-size: 14px;
}

.day-tags {
  display: flex;
  justify-content: center;
  gap: 5px;
  flex-wrap: wrap;
}

.day-tag {
  min-width: 36px;
  text-align: center;
}
</style>
