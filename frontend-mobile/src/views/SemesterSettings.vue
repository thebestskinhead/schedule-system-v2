<template>
  <div class="main-container">
    <Navbar title="学期设置" show-back />
    <div class="card">
      <div class="card-header">
        <span class="card-title">学期设置</span>
      </div>

      <van-form @submit="saveSettings">
        <van-cell-group inset>
          <van-field
            v-model="form.semesterStartDate"
            is-link
            readonly
            label="学期起始日"
            placeholder="选择学期起始日"
            @click="showDatePicker = true"
            :rules="[{ required: true, message: '请选择学期起始日' }]"
          />
          <van-field
            v-model="form.currentWeek"
            label="当前周次"
            disabled
          />
        </van-cell-group>
        <div class="form-hint" style="padding: 0 16px;">根据学期起始日自动计算</div>
        <div style="padding: 16px;">
          <van-button type="primary" block :loading="saving" native-type="submit">保存设置</van-button>
          <van-button block plain style="margin-top: 10px;" @click="calculateCurrentWeek">重新计算周次</van-button>
        </div>
      </van-form>
    </div>

    <!-- 日期选择器 -->
    <van-popup v-model:show="showDatePicker" position="bottom" round>
      <van-date-picker
        v-model="datePickerValue"
        title="选择学期起始日"
        :min-date="minDate"
        :max-date="maxDate"
        @confirm="onDateConfirm"
        @cancel="showDatePicker = false"
      />
    </van-popup>

    <div class="card">
      <div class="card-header">
        <span class="card-title">使用说明</span>
      </div>
      <div class="help-content">
        <p><strong>学期起始日：</strong>设置本学期第一周的周一日期。</p>
        <p><strong>当前周次：</strong>系统根据学期起始日和当前日期自动计算。</p>
        <p><strong>示例：</strong>如果学期从2024年9月2日开始，今天是9月16日，则当前为第3周。</p>
        <van-divider />
        <p><strong>快捷设置：</strong></p>
        <div style="display: flex; gap: 10px;">
          <van-button size="small" @click="setThisYearFall">今年秋季学期</van-button>
          <van-button size="small" @click="setThisYearSpring">今年春季学期</van-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { showToast } from 'vant'
import Navbar from '../components/Navbar.vue'
import { getSemesterStartDate, updateSemesterStartDate } from '../api/schedule'

const loading = ref(false)
const saving = ref(false)
const showDatePicker = ref(false)

const form = reactive({
  semesterStartDate: '',
  currentWeek: 1
})

const datePickerValue = ref([])
const minDate = new Date(2020, 0, 1)
const maxDate = new Date(2030, 11, 31)

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

const formatDate = (date) => {
  if (!date) return ''
  if (typeof date === 'string' && date.length === 10) return date
  const d = new Date(date)
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const onDateConfirm = ({ selectedValues }) => {
  const [year, month, day] = selectedValues
  form.semesterStartDate = `${year}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`
  showDatePicker.value = false
}

const saveSettings = async () => {
  if (!form.semesterStartDate) {
    showToast({ message: '请选择学期起始日', type: 'fail' })
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
    showToast({ message: '保存成功', type: 'success' })
  } catch (error) {
    showToast({ message: '保存失败', type: 'fail' })
  } finally {
    saving.value = false
  }
}

const calculateCurrentWeek = () => {
  if (!form.semesterStartDate) {
    showToast({ message: '请先选择学期起始日', type: 'fail' })
    return
  }

  const start = new Date(form.semesterStartDate)
  const now = new Date()
  const diffTime = now - start
  const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24))

  if (diffDays < 0) {
    form.currentWeek = 1
    showToast({ message: '学期尚未开始', type: 'fail' })
  } else {
    form.currentWeek = Math.floor(diffDays / 7) + 1
    if (form.currentWeek > 30) {
      form.currentWeek = 30
    }
    showToast(`当前为第${form.currentWeek}周`)
  }
}

const setThisYearFall = () => {
  const year = new Date().getFullYear()
  let date = new Date(year, 8, 1)
  while (date.getDay() !== 1) {
    date.setDate(date.getDate() + 1)
  }
  form.semesterStartDate = formatDate(date)
  calculateCurrentWeek()
}

const setThisYearSpring = () => {
  const year = new Date().getFullYear()
  let date = new Date(year, 2, 1)
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
.main-container {
  min-height: 100%;
  background: #f7f8fa;
  padding-top: env(safe-area-inset-top);
  padding-bottom: calc(20px + env(safe-area-inset-bottom));
}

.card {
  margin: 12px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #323233;
}

.form-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  margin-bottom: 12px;
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
