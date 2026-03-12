<template>
  <div class="crawler-import-page">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <h1 class="logo">排班系统 - 自动导入无课表</h1>
          <div class="nav-menu">
            <el-button @click="router.push('/availability')">返回无课表</el-button>
            <el-button @click="router.push('/')">返回首页</el-button>
          </div>
        </div>
      </el-header>

      <el-main class="main-content">
        <div class="page-container">
          <!-- 步骤说明 -->
          <el-card class="mb-4">
            <template #header>
              <div class="card-header">
                <span>📖 使用步骤</span>
              </div>
            </template>
            <el-steps :active="currentStep" simple>
              <el-step title="登录教务系统" description="访问kdjw.hnust.edu.cn" />
              <el-step title="获取Cookie" description="F12复制bzb_jsxsd的值" />
              <el-step title="粘贴导入" description="输入Cookie自动获取" />
              <el-step title="等待完成" description="系统自动导入无课表" />
            </el-steps>
          </el-card>

          <!-- 输入表单 -->
          <el-card v-if="!isImporting && !importResult">
            <template #header>
              <div class="card-header">
                <span>📝 输入教务系统Cookie</span>
              </div>
            </template>

            <el-form :model="form" :rules="rules" ref="formRef" label-width="120px">
              <el-form-item label="Cookie" prop="cookies">
                <el-input
                  v-model="form.cookies"
                  type="textarea"
                  :rows="3"
                  placeholder="bzb_jsxsd=xxx; SERVERID=xxx; bzb_njw=xxx"
                />
                <div class="form-tip">
                  获取方法：登录教务系统 → F12 → Network → 刷新页面 → 点击任意请求 → Headers → Request Headers → 复制 Cookie 整行内容
                </div>
              </el-form-item>

              <el-form-item label="学期" prop="semester">
                <el-select v-model="form.semester" placeholder="选择学期" style="width: 100%">
                  <el-option label="2025-2026-2" value="2025-2026-2" />
                  <el-option label="2025-2026-1" value="2025-2026-1" />
                  <el-option label="2024-2025-2" value="2024-2025-2" />
                  <el-option label="2024-2025-1" value="2024-2025-1" />
                </el-select>
              </el-form-item>

              <el-form-item>
                <el-button type="primary" @click="previewSchedule" :loading="previewLoading">
                  <el-icon><View /></el-icon> 预览课表
                </el-button>
                <el-button type="success" @click="importSchedule" :loading="importLoading">
                  <el-icon><Download /></el-icon> 导入无课表
                </el-button>
              </el-form-item>
            </el-form>
          </el-card>

          <!-- 预览结果 -->
          <el-card v-if="previewData.length > 0 && !isImporting" class="mt-4">
            <template #header>
              <div class="card-header">
                <span>👀 课表预览（共 {{ previewData.length }} 周）</span>
              </div>
            </template>

            <el-tabs type="border-card">
              <el-tab-pane 
                v-for="week in previewData" 
                :key="week.week" 
                :label="'第' + week.week + '周'"
              >
                <div class="timetable-wrapper">
                  <table class="timetable">
                    <thead>
                      <tr>
                        <th class="time-header">节次 / 星期</th>
                        <th v-for="day in week.days" :key="day.day" class="day-header">
                          {{ day.day }}
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="periodIdx in 4" :key="periodIdx">
                        <th class="period-header">第{{ periodIdx }}节</th>
                        <td 
                          v-for="day in week.days" 
                          :key="day.day"
                          class="time-slot"
                          :class="{ 'has-class': day.periods[periodIdx-1].hasClass, 'free-time': !day.periods[periodIdx-1].hasClass }"
                        >
                          <el-tag :type="day.periods[periodIdx-1].hasClass ? 'danger' : 'success'" size="small">
                            {{ day.periods[periodIdx-1].hasClass ? '有课' : '无课' }}
                          </el-tag>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </el-tab-pane>
            </el-tabs>
          </el-card>

          <!-- 导入中状态 -->
          <el-card v-if="isImporting" class="mt-4 importing-card">
            <template #header>
              <div class="card-header">
                <span>⏳ 正在导入无课表</span>
              </div>
            </template>
            
            <div class="importing-content">
              <el-progress 
                :percentage="taskProgress" 
                :status="taskStatus === 'failed' ? 'exception' : ''"
                :stroke-width="20"
                striped
                striped-flow
              />
              
              <div class="task-status">
                <el-tag :type="getStatusType(taskStatus)" size="large" effect="dark">
                  {{ getStatusText(taskStatus) }}
                </el-tag>
              </div>
              
              <div class="task-detail" v-if="taskDetail">
                <p>{{ taskDetail }}</p>
              </div>
              
              <div class="task-info" v-if="currentTask">
                <p><strong>任务ID:</strong> {{ currentTask.id }}</p>
                <p><strong>提交时间:</strong> {{ formatTime(currentTask.created_at) }}</p>
                <p v-if="currentTask.started_at"><strong>开始时间:</strong> {{ formatTime(currentTask.started_at) }}</p>
                <p v-if="currentTask.retry_count > 0"><strong>重试次数:</strong> {{ currentTask.retry_count }}</p>
              </div>

              <div class="importing-actions">
                <el-button @click="cancelImport">取消等待</el-button>
              </div>
            </div>
          </el-card>

          <!-- 导入结果 -->
          <el-card v-if="importResult" class="mt-4">
            <template #header>
              <div class="card-header">
                <span>✅ 导入结果</span>
              </div>
            </template>
            <el-result
              :icon="importResult.success ? 'success' : 'error'"
              :title="importResult.message"
            >
              <template #sub-title>
                <p>解析周数: {{ importResult.weeks_parsed }} 周</p>
                <p>总时段数: {{ importResult.total_cells }} 个</p>
                <p>无课时段: {{ importResult.available_cells }} 个</p>
                <p>已导入: {{ importResult.imported }} 条记录</p>
              </template>
              <template #extra>
                <el-button type="primary" @click="goToAvailability">查看我的无课表</el-button>
                <el-button @click="resetImport">继续导入</el-button>
              </template>
            </el-result>
          </el-card>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { ref, reactive, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { previewCrawler, importCrawler, getImportStatus } from '../api/crawler'

const router = useRouter()
const formRef = ref()
const previewLoading = ref(false)
const importLoading = ref(false)
const previewData = ref([])
const importResult = ref(null)
const isImporting = ref(false)
const currentTask = ref(null)
const taskStatus = ref('')
const taskProgress = ref(0)
const taskDetail = ref('')
const currentStep = ref(1)

let statusPollTimer = null

const form = reactive({
  cookies: '',
  semester: '2025-2026-2',
  startWeek: 1,
  endWeek: 30
})

const rules = {
  cookies: [{ required: true, message: '请输入Cookie', trigger: 'blur' }],
  semester: [{ required: true, message: '请选择学期', trigger: 'change' }]
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

// 预览课表
const previewSchedule = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  previewLoading.value = true
  try {
    const data = await previewCrawler({
      cookies: form.cookies,
      semester: form.semester,
      start_week: 1,
      end_week: 30
    })
    previewData.value = data.preview || []
    ElMessage.success('预览成功')
  } catch (error) {
    // 错误已在拦截器处理
  } finally {
    previewLoading.value = false
  }
}

// 导入无课表（异步模式）
const importSchedule = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  importLoading.value = true
  try {
    // 提交异步任务
    const data = await importCrawler({
      cookies: form.cookies,
      semester: form.semester
    })

    // 先设置导入中状态（在关闭 loading 之前）
    isImporting.value = true
    currentStep.value = 3

    // 确保 UI 更新
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
    isImporting.value = false
  } finally {
    importLoading.value = false
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
    // 不停止轮询，继续尝试
  }
}

// 处理导入完成
const handleImportComplete = (data) => {
  clearInterval(statusPollTimer)
  isImporting.value = false
  
  importResult.value = {
    success: true,
    message: '课表导入成功',
    weeks_parsed: data.result?.weeks_parsed || 0,
    total_cells: data.result?.total_cells || 0,
    available_cells: data.result?.available_cells || 0,
    imported: data.result?.imported || 0
  }
  
  ElMessage.success('无课表导入成功！')
}

// 处理导入失败
const handleImportFailed = (data) => {
  clearInterval(statusPollTimer)
  isImporting.value = false
  
  importResult.value = {
    success: false,
    message: data.error || '导入失败',
    weeks_parsed: 0,
    total_cells: 0,
    available_cells: 0,
    imported: 0
  }
  
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

// 重置导入状态
const resetImport = () => {
  importResult.value = null
  previewData.value = []
  currentTask.value = null
  taskStatus.value = ''
  taskProgress.value = 0
  taskDetail.value = ''
  currentStep.value = 1
  form.cookies = ''
}

// 跳转到无课表页面
const goToAvailability = () => {
  router.push('/availability')
}

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

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 8px;
}

.week-preview {
  margin-bottom: 20px;
}

.week-preview h4 {
  margin: 10px 0;
  color: #303133;
}

.period-list {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.period-tag {
  min-width: 80px;
  text-align: center;
}

/* 课表样式 */
.timetable-wrapper {
  overflow-x: auto;
}

.timetable {
  width: 100%;
  border-collapse: collapse;
  border: 1px solid #e5e5e5;
  font-size: 14px;
}

.timetable th,
.timetable td {
  border: 1px solid #e5e5e5;
  text-align: center;
  vertical-align: middle;
}

.time-header {
  width: 100px;
  height: 50px;
  background: #f5f7fa;
  font-weight: bold;
  color: #606266;
}

.day-header {
  height: 50px;
  background: #f5f7fa;
  font-weight: bold;
  color: #606266;
  min-width: 100px;
}

.period-header {
  width: 100px;
  height: 60px;
  background: #f5f7fa;
  font-weight: bold;
  color: #606266;
}

.time-slot {
  height: 60px;
  min-width: 100px;
  padding: 8px;
  background: #fff;
}

.time-slot.has-class {
  background: #fef0f0;
}

.time-slot.free-time {
  background: #f0f9ff;
}

.mb-4 {
  margin-bottom: 16px;
}

.mt-4 {
  margin-top: 16px;
}

/* 导入中状态卡片样式 */
.importing-card {
  min-height: 400px;
}

.importing-content {
  padding: 40px 20px;
  text-align: center;
}

.task-status {
  margin: 30px 0;
}

.task-status .el-tag {
  font-size: 18px;
  padding: 12px 24px;
}

.task-detail {
  margin: 20px 0;
  color: #606266;
  font-size: 14px;
}

.task-info {
  margin: 30px 0;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 8px;
  text-align: left;
}

.task-info p {
  margin: 8px 0;
  color: #606266;
}

.importing-actions {
  margin-top: 30px;
}
</style>
