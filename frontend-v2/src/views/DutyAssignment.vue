<template>
  <div class="main-container">
    <div class="card">
      <div class="card-header">
        <span class="card-title">每周值班分工管理</span>
        <el-button type="primary" @click="saveData" :loading="saving" :disabled="!hasChanges">
          <el-icon><Check /></el-icon> 保存设置
        </el-button>
      </div>

      <div class="week-selector">
        <el-form :inline="true">
          <el-form-item label="选择周次：">
            <el-select v-model="currentWeek" @change="loadData" style="width: 120px">
              <el-option v-for="w in 20" :key="w" :label="`第 ${w} 周`" :value="w" />
            </el-select>
          </el-form-item>
        </el-form>
      </div>

      <el-table :data="tableData" border style="width: 100%; margin-bottom: 20px">
        <el-table-column prop="department" label="部门" width="120" />
        <el-table-column
          v-for="day in weekdays"
          :key="day.value"
          :label="day.label"
          width="100"
          align="center"
        >
          <template #default="{ row }">
            <el-switch
              v-model="row[`day${day.value}`]"
              active-text="值班"
              inactive-text="休息"
              @change="(val) => onSwitchChange(row.department, day.value, val)"
            />
          </template>
        </el-table-column>
      </el-table>

      <div class="preview-section">
        <h3>本周值班预览</h3>
        <el-row :gutter="20">
          <el-col :span="6" v-for="dept in departments" :key="dept">
            <el-card :body-style="{ padding: '10px' }">
              <div class="dept-name">{{ dept }}</div>
              <div class="day-tags">
                <el-tag
                  v-for="day in weekdays"
                  :key="day.value"
                  :type="isAssigned(dept, day.value) ? 'success' : 'info'"
                  size="small"
                  class="day-tag"
                >
                  {{ day.label.slice(0, 2) }}
                </el-tag>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getDutyAssignments, saveDutyAssignments, getCurrentWeek } from '../api/schedule'

const currentWeek = ref(1)
const assignments = ref({})
const originalAssignments = ref({})
const saving = ref(false)
const loading = ref(false)
const departments = ['办公室', '竞赛部', '项目部', '科普部']

const weekdays = [
  { label: '周一', value: 1 },
  { label: '周二', value: 2 },
  { label: '周三', value: 3 },
  { label: '周四', value: 4 },
  { label: '周五', value: 5 },
]

// 表格数据
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
    // 使用 getDutyAssignments 获取当前周的分工
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
    ElMessage.error('加载分工失败')
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
    ElMessage.success('保存成功！')
  } catch (err) {
    ElMessage.error('保存失败: ' + (err.response?.data?.message || err.message))
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  // 先获取当前周次
  try {
    const res = await getCurrentWeek()
    // 拦截器已提取 data，直接访问
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
.week-selector {
  margin-bottom: 20px;
}

.preview-section {
  margin-top: 30px;
}

.preview-section h3 {
  margin-bottom: 16px;
  color: #303133;
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
