<template>
  <div class="main-container">
    <div class="card">
      <div class="card-header">
        <span class="card-title">学期设置</span>
      </div>

      <el-form :model="form" label-width="120px" class="settings-form">
        <el-form-item label="学期起始日">
          <el-date-picker
            v-model="form.semesterStartDate"
            type="date"
            placeholder="选择学期起始日"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            style="width: 200px"
          />
          <div class="form-hint">设置学期第一天，系统将自动计算当前周次</div>
        </el-form-item>

        <el-form-item label="当前周次">
          <el-input-number v-model="form.currentWeek" :min="1" :max="30" disabled />
          <span class="week-hint">根据学期起始日自动计算</span>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="saveSettings" :loading="saving">
            保存设置
          </el-button>
          <el-button @click="calculateCurrentWeek">重新计算周次</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="card">
      <div class="card-header">
        <span class="card-title">使用说明</span>
      </div>
      <div class="help-content">
        <p><strong>学期起始日：</strong>设置本学期第一周的周一日期。</p>
        <p><strong>当前周次：</strong>系统根据学期起始日和当前日期自动计算。</p>
        <p><strong>示例：</strong>如果学期从2024年9月2日开始，今天是9月16日，则当前为第3周。</p>
        <el-divider />
        <p><strong>快捷设置：</strong></p>
        <el-space>
          <el-button size="small" @click="setThisYearFall">今年秋季学期</el-button>
          <el-button size="small" @click="setThisYearSpring">今年春季学期</el-button>
        </el-space>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getSemesterStartDate, updateSemesterStartDate } from '../api/schedule'

const loading = ref(false)
const saving = ref(false)

const form = reactive({
  semesterStartDate: '',
  currentWeek: 1
})

// 获取学期设置
const fetchSettings = async () => {
  loading.value = true
  try {
    const data = await getSemesterStartDate()
    if (data) {
      form.semesterStartDate = data.semester_start_date || ''
      form.currentWeek = data.current_week || 1
    }
  } catch (error) {
    console.error('获取学期设置失败:', error)
  } finally {
    loading.value = false
  }
}

// 格式化日期为 YYYY-MM-DD
const formatDate = (date) => {
  if (!date) return ''
  if (typeof date === 'string' && date.length === 10) return date
  const d = new Date(date)
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

// 保存设置
const saveSettings = async () => {
  if (!form.semesterStartDate) {
    ElMessage.warning('请选择学期起始日')
    return
  }

  saving.value = true
  try {
    const res = await updateSemesterStartDate({
      semester_start_date: formatDate(form.semesterStartDate)
    })
    if (res && res.current_week) {
      form.currentWeek = res.current_week
    }
    ElMessage.success('保存成功')
  } catch (error) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

// 计算当前周次
const calculateCurrentWeek = () => {
  if (!form.semesterStartDate) {
    ElMessage.warning('请先选择学期起始日')
    return
  }

  const start = new Date(form.semesterStartDate)
  const now = new Date()
  const diffTime = now - start
  const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24))

  if (diffDays < 0) {
    form.currentWeek = 1
    ElMessage.info('学期尚未开始')
  } else {
    form.currentWeek = Math.floor(diffDays / 7) + 1
    if (form.currentWeek > 30) {
      form.currentWeek = 30
    }
    ElMessage.success(`当前为第${form.currentWeek}周`)
  }
}

// 设置今年秋季学期（9月1日）
const setThisYearFall = () => {
  const year = new Date().getFullYear()
  // 找到9月第一个周一
  let date = new Date(year, 8, 1) // 9月1日
  while (date.getDay() !== 1) { // 1表示周一
    date.setDate(date.getDate() + 1)
  }
  form.semesterStartDate = formatDate(date)
  calculateCurrentWeek()
}

// 设置今年春季学期（2月或3月）
const setThisYearSpring = () => {
  const year = new Date().getFullYear()
  // 找到2月最后一个周一或3月第一个周一
  let date = new Date(year, 2, 1) // 3月1日
  while (date.getDay() !== 1) {
    date.setDate(date.getDate() + 1)
  }
  form.semesterStartDate = formatDate(date)
  calculateCurrentWeek()
}

onMounted(() => {
  fetchSettings()
})
</script>

<style scoped>
.settings-form {
  max-width: 500px;
  padding: 20px 0;
}

.form-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 8px;
}

.week-hint {
  margin-left: 12px;
  color: #909399;
  font-size: 13px;
}

.help-content {
  padding: 20px;
  line-height: 2;
  color: #606266;
}

.help-content p {
  margin: 8px 0;
}
</style>
