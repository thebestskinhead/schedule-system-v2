<template>
  <div class="page-container">
    <van-loading v-if="loading" class="loading-center" />

    <van-empty v-else-if="!hasData" description="暂无无课时间记录，请点击下方导入按钮" />

    <template v-else>
      <!-- 当前周选择 -->
      <div class="week-selector">
        <van-dropdown-menu>
          <van-dropdown-item v-model="viewWeek" :options="weekOptions" />
        </van-dropdown-menu>
      </div>

      <!-- 课表网格 -->
      <div class="schedule-grid">
        <!-- 表头：星期 -->
        <div class="grid-header">
          <div class="grid-cell header-cell period-header">节次</div>
          <div 
            v-for="day in weekdays" 
            :key="day.value" 
            class="grid-cell header-cell"
            :class="{ 'is-today': isToday(day.value) }"
          >
            {{ day.shortLabel }}
          </div>
        </div>
        
        <!-- 课表内容 -->
        <div 
          v-for="period in periods" 
          :key="period.value" 
          class="grid-row"
        >
          <div class="grid-cell period-cell">{{ period.label }}</div>
          <div 
            v-for="day in weekdays" 
            :key="day.value"
            class="grid-cell content-cell"
            :class="{ 
              'is-free': isFree(viewWeek, day.value, period.value),
              'is-today': isToday(day.value)
            }"
            @click="handleCellClick(viewWeek, day.value, period.value)"
          >
            <div class="cell-content">
              <van-icon v-if="isFree(viewWeek, day.value, period.value)" name="success" color="#67c23a" size="20" />
              <van-icon v-else name="cross" color="#f56c6c" size="20" />
              <span class="cell-text">{{ isFree(viewWeek, day.value, period.value) ? '无课' : '有课' }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 图例说明 -->
      <div class="legend">
        <div class="legend-item">
          <span class="legend-dot is-free"></span>
          <span>无课</span>
        </div>
        <div class="legend-item">
          <span class="legend-dot is-occupied"></span>
          <span>有课</span>
        </div>
      </div>
    </template>

    <!-- 底部操作栏 -->
    <div class="bottom-action-bar">
      <van-button v-if="isNative" type="warning" size="small" plain @click="handleNativeImport">
        教务系统导入
      </van-button>
      <van-button type="primary" size="small" plain @click="showXLSUpload = true">
        XLS导入
      </van-button>
    </div>

    <!-- 编辑弹窗 -->
    <van-popup 
      v-model:show="showEditDialog" 
      position="bottom" 
      round
      :style="{ height: '70%' }"
    >
      <div class="popup-container">
        <div class="popup-header">
          <span class="popup-title">编辑无课时间</span>
          <van-icon name="cross" @click="showEditDialog = false" />
        </div>

        <van-cell-group inset>
          <van-cell title="星期" :value="editForm.weekdayLabel" />
          <van-cell title="节次" :value="editForm.periodLabel" />
          <van-cell title="当前周" :value="`第${editForm.currentWeek}周`" />
        </van-cell-group>

        <div class="week-select-section">
          <div class="section-title">选择无课的周次（可多选）</div>
          <div class="week-grid">
            <div 
              v-for="week in 30" 
              :key="week"
              class="week-item"
              :class="{ 
                'is-selected': editForm.selectedWeeks.includes(week),
                'is-current': week === editForm.currentWeek
              }"
              @click="toggleWeek(week)"
            >
              {{ week }}
            </div>
          </div>
        </div>

        <div class="popup-actions">
          <van-button block @click="showEditDialog = false">取消</van-button>
          <van-button type="danger" block @click="handleClearAll">全部清空</van-button>
          <van-button type="primary" block :loading="editLoading" @click="handleSaveEdit">保存</van-button>
        </div>
      </div>
    </van-popup>

    <!-- XLS上传弹窗 -->
    <van-popup 
      v-model:show="showXLSUpload" 
      position="bottom" 
      round
      :style="{ height: '50%' }"
    >
      <div class="popup-container">
        <div class="popup-header">
          <span class="popup-title">从XLS文件导入</span>
          <van-icon name="cross" @click="showXLSUpload = false" />
        </div>

        <van-notice-bar
          left-icon="info-o"
          text="Excel文件需要包含以下列：课程名、教师、周次、节次、星期、教室。支持 .xls 和 .xlsx 格式。"
          wrapable
          class="notice-bar"
        />

        <van-uploader
          v-model="fileList"
          :max-count="1"
          accept=".xls,.xlsx"
          :after-read="handleFileChange"
        />

        <div class="popup-actions">
          <van-button block @click="showXLSUpload = false">取消</van-button>
          <van-button type="primary" block :loading="xlsLoading" @click="handleXLSImport">导入</van-button>
        </div>
      </div>
    </van-popup>

    <!-- 导入结果弹窗 -->
    <van-dialog 
      v-model:show="showImportResult" 
      :title="importResult.success ? '导入成功' : '导入失败'"
      show-confirm-button
    >
      <div class="result-content">
        <van-icon :name="importResult.success ? 'passed' : 'close'" :color="importResult.success ? '#67c23a' : '#f56c6c'" size="48" />
        <p class="result-message">{{ importResult.message }}</p>
        <div v-if="importResult.success" class="result-detail">
          <p v-if="importResult.weeks_parsed">解析周数: {{ importResult.weeks_parsed }} 周</p>
          <p v-if="importResult.total_cells">总时段数: {{ importResult.total_cells }} 个</p>
          <p v-if="importResult.available_cells">无课时段: {{ importResult.available_cells }} 个</p>
          <p>已导入: {{ importResult.imported }} 条记录</p>
        </div>
      </div>
    </van-dialog>

    <!-- 安卓原生导入等待遮罩 -->
    <van-overlay :show="nativeImporting" class="import-overlay">
      <div class="import-loading-wrapper">
        <van-loading type="spinner" size="36px" />
        <p class="import-loading-text">{{ nativeImportText }}</p>
      </div>
    </van-overlay>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { showToast } from 'vant'
import { getMyAvailability, addAvailability, deleteAvailability, importFromXLS, importFromXLSBase64 } from '../api/availability'
import { getCurrentWeek } from '../api/schedule'
import { isNativeAvailable, fetchScheduleFromSchool, setLoginCallbacks } from '../utils/native'

const isNative = isNativeAvailable()

// 安卓原生导入相关
const nativeImporting = ref(false)
const nativeImportText = ref('正在获取课表...')
const waitingForLogin = ref(false)

const loading = ref(false)
const availabilityList = ref([])
const viewWeek = ref(1)

// XLS上传相关
const showXLSUpload = ref(false)
const xlsLoading = ref(false)
const fileList = ref([])
const selectedFile = ref(null)

// 导入结果
const showImportResult = ref(false)
const importResult = reactive({
  success: false,
  message: '',
  weeks_parsed: 0,
  total_cells: 0,
  available_cells: 0,
  imported: 0
})

// 编辑相关
const showEditDialog = ref(false)
const editLoading = ref(false)
const editForm = reactive({
  weekday: 1,
  weekdayLabel: '',
  period: 1,
  periodLabel: '',
  currentWeek: 1,
  selectedWeeks: []
})

const weekdays = [
  { value: 1, label: '星期一', shortLabel: '周一' },
  { value: 2, label: '星期二', shortLabel: '周二' },
  { value: 3, label: '星期三', shortLabel: '周三' },
  { value: 4, label: '星期四', shortLabel: '周四' },
  { value: 5, label: '星期五', shortLabel: '周五' }
]

const periods = [
  { value: 1, label: '第一二节' },
  { value: 2, label: '第三四节' },
  { value: 3, label: '第五六节' },
  { value: 4, label: '第七八节' }
]

const weekOptions = computed(() => {
  return Array.from({ length: 30 }, (_, i) => ({
    text: `第${i + 1}周`,
    value: i + 1
  }))
})

const hasData = computed(() => availabilityList.value.length > 0)

const availabilitySet = computed(() => {
  const set = new Set()
  for (const item of availabilityList.value) {
    const key = `${item.week}-${item.weekday}-${item.period}`
    set.add(key)
  }
  return set
})

const isFree = (week, weekday, period) => {
  const key = `${week}-${weekday}-${period}`
  return availabilitySet.value.has(key)
}

const isToday = (weekday) => {
  const today = new Date().getDay()
  // 0=周日, 1=周一, ..., 6=周六
  return today === weekday
}

const handleCellClick = (week, weekday, period) => {
  const weekdayObj = weekdays.find(d => d.value === weekday)
  const periodObj = periods.find(p => p.value === period)
  
  editForm.weekday = weekday
  editForm.weekdayLabel = weekdayObj?.label || ''
  editForm.period = period
  editForm.periodLabel = periodObj?.label || ''
  editForm.currentWeek = week
  
  editForm.selectedWeeks = []
  for (let w = 1; w <= 30; w++) {
    if (isFree(w, weekday, period)) {
      editForm.selectedWeeks.push(w)
    }
  }
  
  showEditDialog.value = true
}

const toggleWeek = (week) => {
  const idx = editForm.selectedWeeks.indexOf(week)
  if (idx > -1) {
    editForm.selectedWeeks.splice(idx, 1)
  } else {
    editForm.selectedWeeks.push(week)
  }
}

const handleSaveEdit = async () => {
  editLoading.value = true
  try {
    const recordsToDelete = availabilityList.value.filter(
      item => item.weekday === editForm.weekday && item.period === editForm.period
    )
    
    for (const record of recordsToDelete) {
      await deleteAvailability({ id: record.id })
    }
    
    if (editForm.selectedWeeks.length > 0) {
      await addAvailability({
        weekday: editForm.weekday,
        period: editForm.period,
        weeks: editForm.selectedWeeks
      })
    }
    
    showToast({ message: '保存成功', type: 'success' })
    showEditDialog.value = false
    fetchData()
  } catch (error) {
    showToast({ message: '保存失败: ' + (error.message || '未知错误'), type: 'fail' })
  } finally {
    editLoading.value = false
  }
}

const handleClearAll = () => {
  editForm.selectedWeeks = []
}

const fetchData = async () => {
  loading.value = true
  try {
    const data = await getMyAvailability()
    availabilityList.value = data || []
  } catch (error) {
    // 错误已在拦截器处理
  } finally {
    loading.value = false
  }
}

const fetchCurrentWeek = async () => {
  try {
    const res = await getCurrentWeek()
    const week = res?.current_week
    if (week) {
      viewWeek.value = week
    }
  } catch (e) {
    console.error('获取当前周次失败:', e)
  }
}

// XLS文件处理
const handleFileChange = (file) => {
  selectedFile.value = file.file
}

const handleXLSImport = async () => {
  if (!selectedFile.value) {
    showToast({ message: '请先选择文件', type: 'fail' })
    return
  }

  xlsLoading.value = true
  try {
    const data = await importFromXLS(selectedFile.value)
    Object.assign(importResult, {
      success: true,
      message: data.message || '导入成功',
      imported: data.imported || 0
    })
    showXLSUpload.value = false
    showImportResult.value = true
    fileList.value = []
    selectedFile.value = null
    fetchData()
    showToast({ message: '导入成功', type: 'success' })
  } catch (error) {
    Object.assign(importResult, {
      success: false,
      message: error.message || '导入失败',
      imported: 0
    })
    showImportResult.value = true
  } finally {
    xlsLoading.value = false
  }
}

// 安卓原生导入
const handleNativeImport = async () => {
  nativeImporting.value = true
  nativeImportText.value = '正在从教务系统获取课表...'

  try {
    // 1. 调用原生方法获取 base64 编码的课表
    const result = await fetchScheduleFromSchool()
    
    // 情况1：需要登录
    if (result.needLogin) {
      nativeImportText.value = '请在安卓端完成登录...'
      waitingForLogin.value = true
      showToast({ message: '请先在安卓端完成教务系统登录', type: 'fail' })
      // 不关闭遮罩，等待登录回调
      return
    }
    
    // 情况2：成功获取课表
    nativeImportText.value = '正在导入课表...'
    await doImport(result.base64, result.fileName)
  } catch (error) {
    Object.assign(importResult, {
      success: false,
      message: error.message || '导入失败',
      imported: 0
    })
    showImportResult.value = true
    nativeImporting.value = false
  }
}

// 登录完成后继续导入
const onLoginComplete = async () => {
  console.log('[Availability] 登录完成，继续导入')
  waitingForLogin.value = false
  nativeImportText.value = '正在获取课表...'
  
  try {
    const result = await fetchScheduleFromSchool()
    if (result.needLogin) {
      throw new Error('登录后仍无法获取课表')
    }
    nativeImportText.value = '正在导入课表...'
    await doImport(result.base64, result.fileName)
  } catch (error) {
    Object.assign(importResult, {
      success: false,
      message: error.message || '导入失败',
      imported: 0
    })
    showImportResult.value = true
    nativeImporting.value = false
  }
}

// 登录失败处理
const onLoginFailed = (errorMessage) => {
  console.log('[Availability] 登录失败:', errorMessage)
  waitingForLogin.value = false
  nativeImporting.value = false
  showToast({ message: '登录失败: ' + errorMessage, type: 'fail' })
}

// 执行导入
const doImport = async (base64, fileName) => {
  try {
    const data = await importFromXLSBase64(base64, fileName)
    Object.assign(importResult, {
      success: true,
      message: data.message || '导入成功',
      imported: data.imported || 0
    })
    showImportResult.value = true
    fetchData()
    showToast({ message: `导入完成，共导入 ${data.imported || 0} 条数据`, type: 'success' })
  } catch (error) {
    Object.assign(importResult, {
      success: false,
      message: error.message || '导入失败',
      imported: 0
    })
    showImportResult.value = true
  } finally {
    nativeImporting.value = false
  }
}

onMounted(() => {
  fetchData()
  fetchCurrentWeek()
  
  // 设置登录回调（安卓端登录完成后会调用）
  if (isNative) {
    setLoginCallbacks(onLoginComplete, onLoginFailed)
  }
})
</script>

<style scoped>
.page-container {
  min-height: 100%;
  background: #f7f8fa;
  padding-top: env(safe-area-inset-top);
  padding-bottom: calc(60px + env(safe-area-inset-bottom));
}

.loading-center {
  display: flex;
  justify-content: center;
  padding: 40px;
}

.week-selector {
  background: #fff;
  position: sticky;
  top: env(safe-area-inset-top);
  z-index: 100;
}

/* 课表网格 */
.schedule-grid {
  background: #fff;
  margin: 12px;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.grid-header {
  display: grid;
  grid-template-columns: 80px repeat(5, 1fr);
  background: #f0f2f5;
}

.grid-row {
  display: grid;
  grid-template-columns: 80px repeat(5, 1fr);
  border-top: 1px solid #ebeef0;
}

.grid-cell {
  padding: 12px 8px;
  text-align: center;
  display: flex;
  align-items: center;
  justify-content: center;
  border-right: 1px solid #ebeef0;
}

.grid-cell:last-child {
  border-right: none;
}

.header-cell {
  font-size: 13px;
  font-weight: 600;
  color: #323233;
  padding: 14px 8px;
}

.header-cell.is-today {
  color: #1989fa;
  background: #ecf5ff;
}

.period-header {
  background: #fafafa;
}

.period-cell {
  font-size: 12px;
  font-weight: 500;
  color: #646566;
  background: #fafafa;
  flex-direction: column;
  line-height: 1.4;
}

.content-cell {
  min-height: 60px;
  cursor: pointer;
  transition: all 0.2s;
  background: #fff5f5;
}

.content-cell:active {
  background: #e8e8e8;
}

.content-cell.is-free {
  background: #e8f5e9;
}

.content-cell.is-today {
  background: #f0f9ff;
}

.content-cell.is-free.is-today {
  background: #e8f9e8;
}

.cell-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.cell-text {
  font-size: 11px;
  color: #969799;
}

/* 图例 */
.legend {
  display: flex;
  justify-content: center;
  gap: 24px;
  padding: 16px;
  background: #fff;
  margin: 0 12px;
  border-radius: 12px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #646566;
}

.legend-dot {
  width: 16px;
  height: 16px;
  border-radius: 4px;
}

.legend-dot.is-free {
  background: #67c23a;
}

.legend-dot.is-occupied {
  background: #f56c6c;
}

/* 底部操作栏 */
.bottom-action-bar {
  display: flex;
  gap: 12px;
  padding: 12px 16px;
  margin: 12px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.bottom-action-bar .van-button {
  flex: 1;
}

/* 弹窗 */
.popup-container {
  padding: 16px;
  max-height: 100%;
  overflow-y: auto;
}

.popup-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.popup-title {
  font-size: 18px;
  font-weight: 600;
}

.popup-actions {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-top: 20px;
}

.popup-actions .van-button:first-child {
  grid-column: span 1;
}

/* 周次选择 */
.week-select-section {
  margin-top: 16px;
}

.section-title {
  font-size: 14px;
  font-weight: 500;
  color: #323233;
  margin-bottom: 12px;
}

.week-grid {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 8px;
}

.week-item {
  aspect-ratio: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f5f5;
  border-radius: 8px;
  font-size: 14px;
  color: #646566;
  cursor: pointer;
  transition: all 0.2s;
}

.week-item.is-selected {
  background: #1989fa;
  color: #fff;
}

.week-item.is-current {
  border: 2px solid #1989fa;
}

.notice-bar {
  margin: 12px 0;
}

/* 导入结果 */
.result-content {
  padding: 20px;
  text-align: center;
}

.result-message {
  font-size: 16px;
  margin: 12px 0;
}

.result-detail {
  text-align: left;
  background: #f5f5f5;
  padding: 12px;
  border-radius: 8px;
  margin-top: 12px;
}

.result-detail p {
  margin: 4px 0;
  font-size: 13px;
  color: #646566;
}

/* 原生导入等待遮罩 */
.import-overlay {
  display: flex;
  align-items: center;
  justify-content: center;
}

.import-loading-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 24px 32px;
  background: #fff;
  border-radius: 12px;
}

.import-loading-text {
  margin-top: 16px;
  font-size: 14px;
  color: #646566;
}
</style>
