<template>
  <div class="availability-page">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <h1 class="logo">排班系统</h1>
          <div class="nav-menu">
            <el-button @click="goTo('/')">返回首页</el-button>
            <div class="user-info">
              <span>{{ userStore.user?.name }}</span>
              <el-button type="danger" size="small" @click="logout">退出</el-button>
            </div>
          </div>
        </div>
      </el-header>

      <el-main class="main-content">
        <div class="page-container">
          <el-card>
            <template #header>
              <div class="card-header">
                <span>我的无课表</span>
                <div class="import-actions">
                  <el-button type="warning" @click="showCookieDialog = true">
                    <el-icon><Download /></el-icon> Cookie导入
                  </el-button>
                  <el-button type="primary" @click="showXLSUpload = true">
                    <el-icon><Document /></el-icon> XLS导入
                  </el-button>
                  <el-button type="info" @click="goTo('/availability/crawler')">
                    <el-icon><View /></el-icon> 课表预览
                  </el-button>
                </div>
              </div>
            </template>

            <!-- 单表格展示全部周次 -->
            <div class="timetable-wrapper" v-loading="loading">
              <div v-if="hasData" class="full-timetable">
                <table class="timetable">
                  <thead>
                    <tr>
                      <th class="corner-header">周次 \ 节次</th>
                      <th v-for="period in periods" :key="period.value" :colspan="5" class="period-header">
                        {{ period.label }}
                      </th>
                    </tr>
                    <tr>
                      <th class="week-header">星期</th>
                      <template v-for="period in periods" :key="period.value">
                        <th v-for="day in weekdays" :key="day.value" class="day-subheader">
                          {{ day.shortLabel }}
                        </th>
                      </template>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="week in 30" :key="week" :class="{ 'current-week': week === currentWeek }">
                      <td class="week-label">第{{ week }}周</td>
                      <template v-for="period in periods" :key="period.value">
                        <td 
                          v-for="day in weekdays" 
                          :key="day.value"
                          class="time-cell"
                          :class="{ 
                            'is-free': isFree(week, day.value, period.value),
                            'is-occupied': !isFree(week, day.value, period.value)
                          }"
                          @click="handleCellClick(week, day.value, period.value)"
                        >
                          <el-icon v-if="isFree(week, day.value, period.value)"><Check /></el-icon>
                          <el-icon v-else><Close /></el-icon>
                        </td>
                      </template>
                    </tr>
                  </tbody>
                </table>
              </div>
              <el-empty v-if="!loading && !hasData" description="暂无无课时间记录，请点击上方导入按钮" />
            </div>

            <!-- 图例说明 -->
            <div class="legend" v-if="hasData">
              <div class="legend-item">
                <span class="legend-box is-free"></span>
                <span>无课（可点击切换）</span>
              </div>
              <div class="legend-item">
                <span class="legend-box is-occupied"></span>
                <span>有课（可点击切换）</span>
              </div>
              <div class="legend-item">
                <span class="legend-text">点击任意时段可快速编辑该周次</span>
              </div>
            </div>
          </el-card>
        </div>
      </el-main>
    </el-container>

    <!-- 编辑对话框 -->
    <el-dialog v-model="showEditDialog" title="编辑无课时间" width="500px">
      <div class="edit-info">
        <el-descriptions :column="3" border>
          <el-descriptions-item label="星期">{{ editForm.weekdayLabel }}</el-descriptions-item>
          <el-descriptions-item label="节次">{{ editForm.periodLabel }}</el-descriptions-item>
          <el-descriptions-item label="当前周">第{{ editForm.currentWeek }}周</el-descriptions-item>
        </el-descriptions>
      </div>
      <div class="edit-section">
        <div class="section-title">选择无课的周次（可多选）</div>
        <el-checkbox-group v-model="editForm.selectedWeeks" class="week-checkbox-group">
          <el-checkbox-button 
            v-for="week in 30" 
            :key="week" 
            :label="week"
            :class="{ 'is-current': week === editForm.currentWeek }"
          >
            {{ week }}
          </el-checkbox-button>
        </el-checkbox-group>
      </div>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="danger" @click="handleClearAll">全部清空</el-button>
        <el-button type="primary" @click="handleSaveEdit" :loading="editLoading">保存</el-button>
      </template>
    </el-dialog>

    <!-- Cookie导入对话框 -->
    <el-dialog v-model="showCookieDialog" :title="isImporting ? '导入中...' : '从教务系统导入'" width="550px" :close-on-click-modal="!isImporting" :show-close="!isImporting">
      <!-- 导入表单 -->
      <div v-if="!isImporting">
        <el-form :model="cookieForm" :rules="cookieRules" ref="cookieFormRef" label-width="100px">
          <el-form-item label="Cookie" prop="cookies">
            <el-input
              v-model="cookieForm.cookies"
              type="textarea"
              :rows="3"
              placeholder="bzb_jsxsd=xxx; SERVERID=xxx; bzb_njw=xxx"
            />
            <div class="form-tip">
              获取方法：登录教务系统 → F12 → Network → 刷新页面 → 点击任意请求 → Headers → Request Headers → 复制 Cookie 整行内容
            </div>
          </el-form-item>
          <el-form-item label="学期" prop="semester">
            <el-select v-model="cookieForm.semester" placeholder="选择学期" style="width: 100%">
              <el-option label="2025-2026-2" value="2025-2026-2" />
              <el-option label="2025-2026-1" value="2025-2026-1" />
              <el-option label="2024-2025-2" value="2024-2025-2" />
              <el-option label="2024-2025-1" value="2024-2025-1" />
            </el-select>
          </el-form-item>
        </el-form>
      </div>

      <!-- 导入中状态 -->
      <div v-else class="importing-status">
        <el-progress :percentage="taskProgress" :status="taskStatus === 'failed' ? 'exception' : ''" striped striped-flow />
        <div class="task-status-tag">
          <el-tag :type="getStatusType(taskStatus)" size="large" effect="dark">
            {{ getStatusText(taskStatus) }}
          </el-tag>
        </div>
        <p class="task-detail">{{ taskDetail }}</p>
        <div v-if="currentTask" class="task-info">
          <p><strong>任务ID:</strong> {{ currentTask.id }}</p>
          <p><strong>提交时间:</strong> {{ formatTime(currentTask.created_at) }}</p>
        </div>
      </div>

      <template #footer>
        <el-button v-if="!isImporting" @click="showCookieDialog = false">取消</el-button>
        <el-button v-if="!isImporting" type="primary" @click="handleCookieImport" :loading="cookieLoading">开始导入</el-button>
        <el-button v-if="isImporting" @click="cancelImport">取消等待</el-button>
      </template>
    </el-dialog>

    <!-- XLS上传对话框 -->
    <el-dialog v-model="showXLSUpload" title="从XLS文件导入" width="500px">
      <el-alert
        title="XLS文件格式要求"
        type="info"
        description="Excel文件需要包含以下列：课程名、教师、周次、节次、星期、教室。支持 .xls 和 .xlsx 格式。"
        :closable="false"
        class="mb-4"
      />
      <el-upload
        ref="uploadRef"
        action="#"
        :auto-upload="false"
        :on-change="handleFileChange"
        :limit="1"
        accept=".xls,.xlsx"
      >
        <template #trigger>
          <el-button type="primary">选择文件</el-button>
        </template>
        <template #tip>
          <div class="el-upload__tip">
            只能上传 .xls/.xlsx 文件
          </div>
        </template>
      </el-upload>
      <template #footer>
        <el-button @click="showXLSUpload = false">取消</el-button>
        <el-button type="success" @click="handleXLSImport" :loading="xlsLoading">导入</el-button>
      </template>
    </el-dialog>

    <!-- 导入结果 -->
    <el-dialog v-model="showImportResult" title="导入结果" width="400px">
      <el-result
        :icon="importResult.success ? 'success' : 'error'"
        :title="importResult.message"
      >
        <template #sub-title>
          <p v-if="importResult.weeks_parsed">解析周数: {{ importResult.weeks_parsed }} 周</p>
          <p v-if="importResult.total_cells">总时段数: {{ importResult.total_cells }} 个</p>
          <p v-if="importResult.available_cells">无课时段: {{ importResult.available_cells }} 个</p>
          <p>已导入: {{ importResult.imported }} 条记录</p>
        </template>
      </el-result>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '../stores/user'
import { getMyAvailability, addAvailability, deleteAvailability } from '../api/availability'
import { importFromCookie, importFromXLS } from '../api/availability'
import { getImportStatus } from '../api/crawler'

const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)
const availabilityList = ref([])

// Cookie导入相关
const showCookieDialog = ref(false)
const cookieLoading = ref(false)
const cookieFormRef = ref()
const cookieForm = reactive({
  cookies: '',
  semester: '2025-2026-2'
})
const cookieRules = {
  cookies: [{ required: true, message: '请输入Cookie', trigger: 'blur' }],
  semester: [{ required: true, message: '请选择学期', trigger: 'change' }]
}

// XLS上传相关
const showXLSUpload = ref(false)
const xlsLoading = ref(false)
const uploadRef = ref()
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

// 任务导入相关（Cookie导入使用异步任务模式）
const isImporting = ref(false)
const currentTask = ref(null)
const taskStatus = ref('')
const taskProgress = ref(0)
const taskDetail = ref('')
let statusPollTimer = null

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

// 当前周次（用于高亮显示）
const currentWeek = computed(() => {
  // 可以根据实际日期计算当前周次
  return 1
})

const hasData = computed(() => availabilityList.value.length > 0)

// 将数据转换为 Set 便于快速查询
const availabilitySet = computed(() => {
  const set = new Set()
  for (const item of availabilityList.value) {
    const key = `${item.week}-${item.weekday}-${item.period}`
    set.add(key)
  }
  return set
})

// 检查某时段是否无课
const isFree = (week, weekday, period) => {
  const key = `${week}-${weekday}-${period}`
  return availabilitySet.value.has(key)
}

// 获取状态标签类型
const getStatusType = (status) => {
  const typeMap = {
    'pending': 'info',
    'running': 'warning',
    'completed': 'success',
    'failed': 'danger'
  }
  return typeMap[status] || 'info'
}

// 获取状态文本
const getStatusText = (status) => {
  const textMap = {
    'pending': '排队中',
    'running': '导入中',
    'completed': '导入完成',
    'failed': '导入失败'
  }
  return textMap[status] || status
}

// 格式化时间
const formatTime = (timeStr) => {
  if (!timeStr) return '-'
  const date = new Date(timeStr)
  return date.toLocaleString('zh-CN')
}

// 点击单元格编辑
const handleCellClick = (week, weekday, period) => {
  const weekdayObj = weekdays.find(d => d.value === weekday)
  const periodObj = periods.find(p => p.value === period)
  
  editForm.weekday = weekday
  editForm.weekdayLabel = weekdayObj?.label || ''
  editForm.period = period
  editForm.periodLabel = periodObj?.label || ''
  editForm.currentWeek = week
  
  // 获取该时段所有无课的周次
  editForm.selectedWeeks = []
  for (let w = 1; w <= 30; w++) {
    if (isFree(w, weekday, period)) {
      editForm.selectedWeeks.push(w)
    }
  }
  
  showEditDialog.value = true
}

// 保存编辑
const handleSaveEdit = async () => {
  editLoading.value = true
  try {
    // 先删除该时段的所有记录
    const recordsToDelete = availabilityList.value.filter(
      item => item.weekday === editForm.weekday && item.period === editForm.period
    )
    
    for (const record of recordsToDelete) {
      await deleteAvailability({ id: record.id })
    }
    
    // 添加新的记录
    if (editForm.selectedWeeks.length > 0) {
      await addAvailability({
        weekday: editForm.weekday,
        period: editForm.period,
        weeks: editForm.selectedWeeks
      })
    }
    
    ElMessage.success('保存成功')
    showEditDialog.value = false
    fetchData()
  } catch (error) {
    ElMessage.error('保存失败: ' + (error.message || '未知错误'))
  } finally {
    editLoading.value = false
  }
}

// 清空该时段所有周次
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

const goTo = (path) => {
  router.push(path)
}

const logout = () => {
  userStore.logout()
  ElMessage.success('已退出登录')
  router.push('/login')
}

// Cookie导入（异步任务模式）
const handleCookieImport = async () => {
  const valid = await cookieFormRef.value.validate().catch(() => false)
  if (!valid) return

  cookieLoading.value = true
  try {
    // 提交异步任务
    const data = await importFromCookie({
      cookies: cookieForm.cookies,
      semester: cookieForm.semester
    })

    // 切换到导入中状态
    isImporting.value = true
    await nextTick()

    currentTask.value = {
      id: data.task_id,
      created_at: data.created_at,
      ...data
    }
    taskStatus.value = data.status || 'pending'
    taskProgress.value = 10
    taskDetail.value = '任务已提交，正在排队中...'

    ElMessage.success('任务已提交，正在导入中...')

    // 开始轮询任务状态
    startStatusPolling()
  } catch (error) {
    ElMessage.error(error.message || '提交任务失败')
  } finally {
    cookieLoading.value = false
  }
}

// 开始轮询任务状态
const startStatusPolling = () => {
  // 延迟1秒后查询，给服务器处理时间
  setTimeout(() => {
    checkTaskStatus()
  }, 1000)

  // 每2秒查询一次
  statusPollTimer = setInterval(() => {
    checkTaskStatus()
  }, 2000)
}

// 检查任务状态
const checkTaskStatus = async () => {
  if (!currentTask.value?.id) {
    console.log('无任务ID，停止轮询')
    clearInterval(statusPollTimer)
    return
  }

  try {
    console.log('查询任务状态:', currentTask.value.id)
    const data = await getImportStatus(currentTask.value.id)
    console.log('任务状态响应:', data)

    currentTask.value = data
    taskStatus.value = data.status || 'pending'

    // 更新进度和详情
    if (data.status === 'pending') {
      taskProgress.value = 10
      taskDetail.value = '任务排队中，请稍候...'
    } else if (data.status === 'running') {
      taskProgress.value = 50
      taskDetail.value = '正在爬取课表数据...'
    } else if (data.status === 'completed') {
      taskProgress.value = 100
      taskDetail.value = '导入完成！'
      handleImportComplete(data)
    } else if (data.status === 'failed') {
      taskDetail.value = data.error || '导入失败'
      handleImportFailed(data)
    }
  } catch (error) {
    console.error('查询任务状态失败:', error)
  }
}

// 处理导入完成
const handleImportComplete = (data) => {
  clearInterval(statusPollTimer)
  isImporting.value = false
  showCookieDialog.value = false

  Object.assign(importResult, {
    success: true,
    message: '课表导入成功',
    weeks_parsed: data.result?.weeks_parsed || 0,
    total_cells: data.result?.total_cells || 0,
    available_cells: data.result?.available_cells || 0,
    imported: data.result?.imported || 0
  })

  showImportResult.value = true
  fetchData()
  ElMessage.success('无课表导入成功！')
}

// 处理导入失败
const handleImportFailed = (data) => {
  clearInterval(statusPollTimer)
  isImporting.value = false

  Object.assign(importResult, {
    success: false,
    message: data.error || '导入失败',
    weeks_parsed: 0,
    total_cells: 0,
    available_cells: 0,
    imported: 0
  })

  showCookieDialog.value = false
  showImportResult.value = true
  ElMessage.error(data.error || '导入失败')
}

// 取消导入
const cancelImport = () => {
  clearInterval(statusPollTimer)
  isImporting.value = false
  currentTask.value = null
  taskStatus.value = ''
  taskProgress.value = 0
  taskDetail.value = ''
  ElMessage.info('已取消导入')
}

// XLS文件处理
const handleFileChange = (uploadFile) => {
  selectedFile.value = uploadFile.raw
}

// XLS导入
const handleXLSImport = async () => {
  if (!selectedFile.value) {
    ElMessage.warning('请先选择文件')
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
    selectedFile.value = null
    uploadRef.value?.clearFiles()
    fetchData()
    ElMessage.success('导入成功')
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

onMounted(() => {
  fetchData()
})

// 组件卸载时清理定时器
onUnmounted(() => {
  if (statusPollTimer) {
    clearInterval(statusPollTimer)
  }
})
</script>

<style scoped>
.header {
  background: #fff;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  padding: 0;
}

.header-content {
  max-width: 1400px;
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

.nav-menu {
  display: flex;
  align-items: center;
  gap: 20px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.import-actions {
  display: flex;
  gap: 10px;
}

.importing-status {
  padding: 20px;
  text-align: center;
}

.task-status-tag {
  margin: 20px 0;
}

.task-status-tag .el-tag {
  font-size: 16px;
  padding: 10px 20px;
}

.task-detail {
  color: #606266;
  margin: 15px 0;
}

.task-info {
  margin-top: 20px;
  padding: 15px;
  background: #f5f7fa;
  border-radius: 8px;
  text-align: left;
}

.task-info p {
  margin: 8px 0;
  color: #606266;
}

.main-content {
  min-height: calc(100vh - 60px);
  padding: 20px;
}

.page-container {
  max-width: 1400px;
  margin: 0 auto;
}

/* 表格样式 */
.timetable-wrapper {
  overflow-x: auto;
  max-height: 70vh;
  overflow-y: auto;
}

.full-timetable {
  min-width: 1000px;
}

.timetable {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.timetable th,
.timetable td {
  border: 1px solid #e0e0e0;
  text-align: center;
  vertical-align: middle;
}

.corner-header {
  width: 80px;
  height: 50px;
  background: #f5f7fa;
  font-weight: bold;
  position: sticky;
  top: 0;
  left: 0;
  z-index: 3;
}

.period-header {
  background: #f5f7fa;
  font-weight: bold;
  height: 30px;
  position: sticky;
  top: 0;
  z-index: 2;
}

.week-header {
  background: #f5f7fa;
  font-weight: bold;
  height: 30px;
  position: sticky;
  left: 0;
  z-index: 2;
}

.day-subheader {
  background: #f5f7fa;
  font-weight: normal;
  font-size: 11px;
  width: 40px;
  position: sticky;
  top: 30px;
  z-index: 2;
}

.week-label {
  background: #f5f7fa;
  font-weight: bold;
  width: 80px;
  height: 32px;
  position: sticky;
  left: 0;
  z-index: 1;
}

.time-cell {
  width: 40px;
  height: 32px;
  cursor: pointer;
  transition: all 0.2s;
}

.time-cell:hover {
  transform: scale(1.1);
  z-index: 10;
  box-shadow: 0 2px 8px rgba(0,0,0,0.15);
}

.time-cell.is-free {
  background: #67c23a;
  color: #fff;
}

.time-cell.is-occupied {
  background: #f56c6c;
  color: #fff;
}

.current-week .week-label {
  background: #409eff;
  color: #fff;
}

/* 图例 */
.legend {
  display: flex;
  gap: 30px;
  margin-top: 20px;
  padding: 15px;
  background: #f5f7fa;
  border-radius: 4px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.legend-box {
  width: 20px;
  height: 20px;
  border-radius: 4px;
}

.legend-box.is-free {
  background: #67c23a;
}

.legend-box.is-occupied {
  background: #f56c6c;
}

.legend-text {
  color: #909399;
  font-size: 13px;
}

/* 编辑对话框样式 */
.edit-info {
  margin-bottom: 20px;
}

.edit-section {
  margin-top: 20px;
}

.section-title {
  font-weight: bold;
  margin-bottom: 15px;
  color: #303133;
}

.week-checkbox-group {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.week-checkbox-group :deep(.el-checkbox-button__inner) {
  width: 50px;
  padding: 8px 0;
}

.week-checkbox-group :deep(.el-checkbox-button.is-current .el-checkbox-button__inner) {
  border-color: #409eff;
  box-shadow: -1px 0 0 0 #409eff;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 8px;
}

.mb-4 {
  margin-bottom: 16px;
}
</style>
