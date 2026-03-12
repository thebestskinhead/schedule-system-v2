<template>
  <div class="schedule-page">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <h1 class="logo">排班系统 - 排班管理</h1>
          <div class="nav-menu">
            <el-button @click="router.push('/')">返回首页</el-button>
          </div>
        </div>
      </el-header>

      <el-main class="main-content">
        <div class="page-container">
          <!-- 参数配置卡片 -->
          <el-card>
            <template #header>
              <div class="card-header">
                <span>排班参数配置</span>
                <el-button type="primary" size="small" @click="saveSettings" :loading="savingSettings">
                  <el-icon><Check /></el-icon> 保存设置
                </el-button>
              </div>
            </template>

            <el-form :model="settingsForm" ref="settingsRef" label-width="160px" class="settings-form">
              <el-divider>当前周次设置</el-divider>
              <el-row :gutter="20">
                <el-col :span="12">
                  <el-form-item label="当前周次">
                    <el-input-number v-model="settingsForm.currentWeek" :min="1" :max="30" />
                    <span class="form-hint">首页默认显示该周的排班</span>
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="自动递增周次">
                    <el-switch v-model="settingsForm.autoIncrement" />
                    <span class="form-hint">确认排班后自动进入下一周</span>
                  </el-form-item>
                </el-col>
              </el-row>
              
              <el-divider>排班规则</el-divider>
              <el-row :gutter="20">
                <el-col :span="12">
                  <el-form-item label="每时段最大人数">
                    <el-input-number v-model="settingsForm.needPerCell" :min="1" :max="10" />
                    <span class="form-hint">默认2人</span>
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="每时段最小人数">
                    <el-input-number v-model="settingsForm.minPerCell" :min="0" :max="10" />
                    <span class="form-hint">默认0人（无人时警告）</span>
                  </el-form-item>
                </el-col>
              </el-row>
              <el-row :gutter="20">
                <el-col :span="12">
                  <el-form-item label="每人每天最多排班">
                    <el-input-number v-model="settingsForm.maxPerDay" :min="1" :max="10" />
                    <span class="form-hint">默认1次</span>
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="每人每周最多排班">
                    <el-input-number v-model="settingsForm.maxPerWeek" :min="1" :max="30" />
                    <span class="form-hint">默认2次</span>
                  </el-form-item>
                </el-col>
              </el-row>
              <el-form-item label="导出Excel标题模板">
                <el-input v-model="settingsForm.exportTitle" placeholder="如：xx部第{week}周排班表" style="width: 300px" />
                <span class="form-hint">使用 {week} 作为周次占位符</span>
              </el-form-item>
            </el-form>
          </el-card>

          <!-- 排班生成卡片 -->
          <el-card class="mt-4">
            <template #header>
              <div class="card-header">
                <span>生成排班</span>
              </div>
            </template>

            <el-form :model="form" :rules="rules" ref="formRef" label-width="120px" inline>
              <el-form-item label="周次" prop="week">
                <el-select-v2
                  v-model="form.week"
                  :options="weekOptions"
                  placeholder="选择周次"
                  style="width: 150px"
                />
              </el-form-item>

              <el-form-item label="排班部门" prop="department">
                <el-select v-model="form.department" placeholder="选择部门" style="width: 150px">
                  <el-option label="办公室" value="办公室" />
                  <el-option label="竞赛部" value="竞赛部" />
                  <el-option label="项目部" value="项目部" />
                  <el-option label="科普部" value="科普部" />
                </el-select>
              </el-form-item>

              <el-form-item label="排班星期" prop="days">
                <el-checkbox-group v-model="form.days">
                  <el-checkbox v-for="i in 5" :key="i" :label="i">
                    周{{ ['一','二','三','四','五'][i-1] }}
                  </el-checkbox>
                </el-checkbox-group>
              </el-form-item>

              <el-form-item label="每天节次" prop="periods">
                <el-radio-group v-model="form.periods">
                  <el-radio-button :label="1">1节</el-radio-button>
                  <el-radio-button :label="2">2节</el-radio-button>
                  <el-radio-button :label="3">3节</el-radio-button>
                  <el-radio-button :label="4">4节</el-radio-button>
                </el-radio-group>
              </el-form-item>

              <el-form-item>
                <el-button type="primary" @click="generatePreview" :loading="loading">
                  生成排班预览
                </el-button>
              </el-form-item>
            </el-form>
          </el-card>

          <!-- 导出模板配置 -->
          <el-card class="mt-4">
            <template #header>
              <div class="card-header">
                <span>导出模板配置</span>
                <el-button type="primary" size="small" @click="showTemplateDialog = true">
                  <el-icon><Plus /></el-icon> 新建模板
                </el-button>
              </div>
            </template>
            
            <el-form :inline="true">
              <el-form-item label="选择模板">
                <el-select v-model="selectedTemplateId" placeholder="选择导出模板" style="width: 200px">
                  <el-option 
                    v-for="t in templates" 
                    :key="t.id" 
                    :label="t.name + (t.is_default ? ' (默认)' : '')" 
                    :value="t.id" 
                  />
                </el-select>
              </el-form-item>
              <el-form-item label="部门名称">
                <el-input v-model="exportDepartment" placeholder="如：办公室" style="width: 150px" />
              </el-form-item>
              <el-form-item>
                <el-button type="success" @click="previewExportTemplate">
                  <el-icon><View /></el-icon> 预览效果
                </el-button>
              </el-form-item>
            </el-form>

            <!-- 模板列表 -->
            <el-table :data="templates" size="small" v-if="templates.length > 0">
              <el-table-column prop="name" label="模板名称" width="150" />
              <el-table-column prop="description" label="描述" show-overflow-tooltip />
              <el-table-column label="操作" width="150">
                <template #default="{ row }">
                  <el-button link type="primary" size="small" @click="editTemplate(row)">编辑</el-button>
                  <el-button link type="danger" size="small" @click="deleteTemplateById(row.id)" v-if="!row.is_default">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>

          <!-- 排班预览卡片 -->
          <el-card v-if="previewData" class="mt-4">
            <template #header>
              <div class="card-header">
                <span>排班预览 - 第{{ previewData.week }}周</span>
                <div>
                  <el-button type="success" @click="confirmSchedule" :loading="confirming">
                    确认排班
                  </el-button>
                  <el-button type="primary" @click="exportToExcel" :loading="exporting">
                    <el-icon><Download /></el-icon> 导出Excel
                  </el-button>
                </div>
              </div>
            </template>

            <!-- 错误提示（人数不足最小要求） -->
            <el-alert
              v-if="previewData.conflicts && previewData.conflicts.length > 0"
              title="以下时段人数不足最小要求："
              type="error"
              :closable="false"
              class="mb-4"
            >
              <div v-for="conflict in previewData.conflicts" :key="conflict.weekday + '-' + conflict.period">
                周{{ ['一','二','三','四','五'][conflict.weekday-1] }} 第{{ conflict.period }}节 
                (最少需要{{ conflict.need }}人，实际{{ conflict.available }}人)
              </div>
            </el-alert>

            <!-- 警告提示（人数不足最大要求） -->
            <el-alert
              v-if="previewData.warnings && previewData.warnings.length > 0"
              title="以下时段人数不足最大要求（仅警告）："
              type="warning"
              :closable="false"
              class="mb-4"
            >
              <div v-for="warning in previewData.warnings" :key="warning.weekday + '-' + warning.period">
                周{{ ['一','二','三','四','五'][warning.weekday-1] }} 第{{ warning.period }}节 
                (最多{{ warning.need }}人，实际{{ warning.available }}人)
              </div>
            </el-alert>

            <!-- 排班表格 -->
            <el-table :data="flattenGrid" border style="width: 100%">
              <el-table-column prop="weekdayText" label="星期" width="100" />
              <el-table-column prop="period" label="节次" width="100">
                <template #default="{ row }">第{{ row.period }}节</template>
              </el-table-column>
              <el-table-column label="值班人员">
                <template #default="{ row }">
                  <div class="user-tags">
                    <el-tag 
                      v-for="user in (row.users || [])" 
                      :key="user.id"
                      size="small"
                      class="user-tag"
                      closable
                      @close="removeUser(row, user)"
                    >
                      {{ user.name }}
                    </el-tag>
                    <el-button 
                      v-if="(row.users || []).length < settingsForm.needPerCell"
                      size="small" 
                      type="primary" 
                      plain
                      @click="showAddUserDialog(row)"
                    >
                      <el-icon><Plus /></el-icon> 添加
                    </el-button>
                    <span v-if="(row.users || []).length === 0" class="empty-text">无人可选</span>
                  </div>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </div>
      </el-main>
    </el-container>

    <!-- 添加人员对话框 -->
    <el-dialog v-model="showAddDialog" title="添加值班人员" width="500px">
      <el-form :model="addForm" label-width="100px">
        <el-form-item label="时段">
          <span>{{ addForm.weekdayText }} 第{{ addForm.period }}节</span>
        </el-form-item>
        <el-form-item label="可选人员">
          <el-select-v2
            v-model="addForm.selectedUsers"
            :options="availableUserOptions"
            placeholder="选择人员（可多选）"
            multiple
            clearable
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmAddUsers" :loading="addingUsers">添加</el-button>
      </template>
    </el-dialog>

    <!-- 模板编辑对话框 -->
    <el-dialog v-model="showTemplateDialog" :title="editingTemplate ? '编辑模板' : '新建模板'" width="700px">
      <el-form :model="templateForm" label-width="100px">
        <el-form-item label="模板名称">
          <el-input v-model="templateForm.name" placeholder="输入模板名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="templateForm.description" placeholder="输入模板描述" />
        </el-form-item>
        <el-form-item label="导出格式">
          <el-radio-group v-model="templateForm.config.mode">
            <el-radio-button label="list">列表格式</el-radio-button>
            <el-radio-button label="schedule">课表格式</el-radio-button>
          </el-radio-group>
          <div class="form-hint">
            列表格式：每行一个时段，类似清单<br>
            课表格式：矩阵表格，类似课程表
          </div>
        </el-form-item>
        <el-form-item label="表格标题">
          <el-input v-model="templateForm.config.title" placeholder="如：{department}第{week}周排班表" />
          <div class="form-hint">可用占位符：{week}周次, {department}部门</div>
        </el-form-item>
        
        <!-- 列表格式配置 -->
        <template v-if="templateForm.config.mode === 'list'">
          <el-form-item label="表头设置">
            <div v-for="(header, index) in templateForm.config.headers" :key="index" class="header-row">
              <el-input v-model="templateForm.config.headers[index]" placeholder="表头名称" style="width: 150px" />
              <el-button link type="danger" @click="removeHeader(index)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button link type="primary" @click="addHeader">
              <el-icon><Plus /></el-icon> 添加表头
            </el-button>
          </el-form-item>
          <el-form-item label="数据列配置">
            <div v-for="(col, index) in templateForm.config.dataColumns" :key="index" class="column-config">
              <el-select v-model="col.type" placeholder="类型" style="width: 120px">
                <el-option label="星期" value="weekday" />
                <el-option label="节次" value="period" />
                <el-option label="值班人员" value="users" />
                <el-option label="固定文本" value="text" />
              </el-select>
              <el-input v-model="col.format" placeholder="格式模板" style="width: 200px; margin: 0 10px" />
              <el-button link type="danger" @click="removeDataColumn(index)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
            <el-button link type="primary" @click="addDataColumn">
              <el-icon><Plus /></el-icon> 添加列
            </el-button>
            <div class="form-hint">
              格式模板说明：<br>
              - 星期: 周{weekday_cn} 或 星期{weekday}<br>
              - 节次: 第{period}节<br>
              - 人员: {users}（自动填充姓名列表）
            </div>
          </el-form-item>
        </template>

        <!-- 课表格式配置 -->
        <template v-if="templateForm.config.mode === 'schedule'">
          <el-divider>课表格式配置</el-divider>
          <el-form-item label="行标题">
            <el-input v-model="templateForm.config.scheduleConfig.rowHeader" placeholder="如：节次" style="width: 150px" />
          </el-form-item>
          <el-form-item label="列标题">
            <el-input v-model="templateForm.config.scheduleConfig.colHeader" placeholder="如：星期" style="width: 150px" />
          </el-form-item>
          <el-form-item label="行标签">
            <el-input 
              v-model="scheduleRowLabels" 
              placeholder="第1节,第2节,第3节,第4节" 
              style="width: 300px"
            />
            <div class="form-hint">用逗号分隔，对应1-4节</div>
          </el-form-item>
          <el-form-item label="列标签">
            <el-input 
              v-model="scheduleColLabels" 
              placeholder="周一,周二,周三,周四,周五" 
              style="width: 300px"
            />
            <div class="form-hint">用逗号分隔，对应周一到周五</div>
          </el-form-item>
          <el-form-item label="单元格格式">
            <el-input v-model="templateForm.config.scheduleConfig.cellFormat" placeholder="如：{users} 或 {users}({count}人)" style="width: 250px" />
            <div class="form-hint">{users}=人员名单, {count}=人数</div>
          </el-form-item>
          <el-form-item label="空单元格">
            <el-input v-model="templateForm.config.scheduleConfig.emptyCellText" placeholder="如：- 或 暂无" style="width: 150px" />
            <div class="form-hint">无排班时显示的内容</div>
          </el-form-item>
        </template>

        <el-form-item label="设为默认">
          <el-switch v-model="templateForm.is_default" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showTemplateDialog = false">取消</el-button>
        <el-button type="primary" @click="saveTemplate" :loading="savingTemplate">保存</el-button>
      </template>
    </el-dialog>

    <!-- 预览效果对话框 -->
    <el-dialog v-model="showPreviewDialog" title="导出预览" width="800px">
      <div class="preview-content">
        <h3>{{ previewTitle }}</h3>
        
        <!-- 列表模式预览 -->
        <el-table v-if="previewDataList.length > 0" :data="previewDataList" border size="small">
          <el-table-column 
            v-for="(header, index) in previewHeaders" 
            :key="index"
            :prop="'col' + index" 
            :label="header" 
          />
        </el-table>
        
        <!-- 课表模式预览 -->
        <el-table v-if="previewScheduleData.length > 0" :data="previewScheduleData" border size="small">
          <el-table-column prop="label" :label="previewScheduleHeaders[0]" width="100" />
          <el-table-column 
            v-for="(header, index) in previewScheduleHeaders.slice(1)" 
            :key="index"
            :prop="'day' + index" 
            :label="header" 
            min-width="120"
          />
        </el-table>
      </div>
      <template #footer>
        <el-button @click="showPreviewDialog = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { 
  previewSchedule, 
  confirmSchedule as confirmScheduleAPI,
  getScheduleSettings,
  saveScheduleSettings,
  updateSchedule,
  getTemplates,
  createTemplate,
  updateTemplate as updateTemplateAPI,
  deleteTemplate as deleteTemplateAPI,
  getCurrentWeek
} from '../api/schedule'
import { getUsersForSchedule } from '../api/user'
import * as XLSX from 'xlsx'

const router = useRouter()
const formRef = ref()
const settingsRef = ref()
const loading = ref(false)
const confirming = ref(false)
const savingSettings = ref(false)
const exporting = ref(false)
const previewData = ref(null)
const allUsers = ref([])

// 模板相关
const templates = ref([])
const selectedTemplateId = ref(0)
const exportDepartment = ref('')
const showTemplateDialog = ref(false)
const savingTemplate = ref(false)
const editingTemplate = ref(null)
const showPreviewDialog = ref(false)
const previewTitle = ref('')
const previewHeaders = ref([])
const previewDataList = ref([])
const previewScheduleHeaders = ref([])
const previewScheduleData = ref([])

const templateForm = reactive({
  name: '',
  description: '',
  is_default: false,
  config: {
    title: '第{week}周排班表',
    mode: 'schedule',
    headers: ['星期', '节次', '值班人员'],
    dataColumns: [
      { type: 'weekday', format: '周{weekday_cn}' },
      { type: 'period', format: '第{period}节' },
      { type: 'users', format: '{users}', separator: '、' }
    ],
    scheduleConfig: {
      rowHeader: '节次',
      colHeader: '星期',
      rowLabels: ['第1节', '第2节', '第3节', '第4节'],
      colLabels: ['周一', '周二', '周三', '周四', '周五'],
      cellFormat: '{users}',
      emptyCellText: '-'
    }
  }
})

// 课表标签输入（逗号分隔）
const scheduleRowLabels = computed({
  get: () => templateForm.config.scheduleConfig?.rowLabels?.join(',') || '第1节,第2节,第3节,第4节',
  set: (val) => {
    if (!templateForm.config.scheduleConfig) {
      templateForm.config.scheduleConfig = {}
    }
    templateForm.config.scheduleConfig.rowLabels = val.split(',').map(s => s.trim()).filter(s => s)
  }
})

const scheduleColLabels = computed({
  get: () => templateForm.config.scheduleConfig?.colLabels?.join(',') || '周一,周二,周三,周四,周五',
  set: (val) => {
    if (!templateForm.config.scheduleConfig) {
      templateForm.config.scheduleConfig = {}
    }
    templateForm.config.scheduleConfig.colLabels = val.split(',').map(s => s.trim()).filter(s => s)
  }
})

// 本地存储键
const SCHEDULE_FORM_KEY = 'schedule_form_memory'
const SCHEDULE_SETTINGS_KEY = 'schedule_settings_memory'

// 设置表单
const settingsForm = reactive({
  currentWeek: 1,
  autoIncrement: false,
  needPerCell: 2,
  minPerCell: 0,
  maxPerDay: 1,
  maxPerWeek: 2,
  exportTitle: '第{week}周排班表'
})

// 排班生成表单
const form = reactive({
  week: 1,
  days: [1, 2, 3, 4, 5],
  periods: 4,
  department: ''
})

// 从localStorage加载记忆设置
const loadMemorySettings = () => {
  try {
    // 加载排班表单设置
    const savedForm = localStorage.getItem(SCHEDULE_FORM_KEY)
    if (savedForm) {
      const parsed = JSON.parse(savedForm)
      if (parsed.week) form.week = parsed.week
      if (parsed.days && parsed.days.length > 0) form.days = parsed.days
      if (parsed.periods) form.periods = parsed.periods
      if (parsed.department) form.department = parsed.department
    }
    
    // 加载排班参数设置
    const savedSettings = localStorage.getItem(SCHEDULE_SETTINGS_KEY)
    if (savedSettings) {
      const parsed = JSON.parse(savedSettings)
      if (parsed.currentWeek !== undefined) settingsForm.currentWeek = parsed.currentWeek
      if (parsed.autoIncrement !== undefined) settingsForm.autoIncrement = parsed.autoIncrement
      if (parsed.needPerCell !== undefined) settingsForm.needPerCell = parsed.needPerCell
      if (parsed.minPerCell !== undefined) settingsForm.minPerCell = parsed.minPerCell
      if (parsed.maxPerDay !== undefined) settingsForm.maxPerDay = parsed.maxPerDay
      if (parsed.maxPerWeek !== undefined) settingsForm.maxPerWeek = parsed.maxPerWeek
      if (parsed.exportTitle) settingsForm.exportTitle = parsed.exportTitle
    }
  } catch (error) {
    console.error('加载记忆设置失败:', error)
  }
}

// 保存排班表单到localStorage
const saveFormMemory = () => {
  try {
    const toSave = {
      week: form.week,
      days: form.days,
      periods: form.periods,
      department: form.department
    }
    localStorage.setItem(SCHEDULE_FORM_KEY, JSON.stringify(toSave))
  } catch (error) {
    console.error('保存表单记忆失败:', error)
  }
}

// 保存排班参数到localStorage
const saveSettingsMemory = () => {
  try {
    const toSave = {
      currentWeek: settingsForm.currentWeek,
      autoIncrement: settingsForm.autoIncrement,
      needPerCell: settingsForm.needPerCell,
      minPerCell: settingsForm.minPerCell,
      maxPerDay: settingsForm.maxPerDay,
      maxPerWeek: settingsForm.maxPerWeek,
      exportTitle: settingsForm.exportTitle
    }
    localStorage.setItem(SCHEDULE_SETTINGS_KEY, JSON.stringify(toSave))
  } catch (error) {
    console.error('保存设置记忆失败:', error)
  }
}

const weekOptions = computed(() => {
  return Array.from({ length: 30 }, (_, i) => ({
    value: i + 1,
    label: `第${i + 1}周`
  }))
})

const rules = {
  week: [{ required: true, message: '请选择周次', trigger: 'change' }],
  department: [{ required: true, message: '请选择排班部门', trigger: 'change' }],
  days: [{ required: true, message: '请至少选择一天', trigger: 'change', type: 'array', min: 1 }],
  periods: [{ required: true, message: '请选择节次', trigger: 'change' }]
}

// 添加人员相关
const showAddDialog = ref(false)
const addingUsers = ref(false)
const addForm = reactive({
  weekday: 1,
  weekdayText: '',
  period: 1,
  selectedUsers: []
})
const currentEditRow = ref(null)

// 加载设置
const loadSettings = async () => {
  try {
    const data = await getScheduleSettings()
    if (data) {
      Object.assign(settingsForm, data)
    }
  } catch (error) {
    // 使用默认设置
  }
  // 加载本地记忆设置（会覆盖服务器默认值）
  loadMemorySettings()
}

// 保存设置
const saveSettings = async () => {
  savingSettings.value = true
  try {
    await saveScheduleSettings(settingsForm)
    // 同时保存到本地记忆
    saveSettingsMemory()
    ElMessage.success('设置保存成功')
  } catch (error) {
    ElMessage.error('保存失败')
  } finally {
    savingSettings.value = false
  }
}

// 生成预览
const generatePreview = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const data = await previewSchedule({
      week: form.week,
      days: form.days,
      periods: form.periods,
      department: form.department,
      need_per_cell: settingsForm.needPerCell,
      min_per_cell: settingsForm.minPerCell,
      max_per_day: settingsForm.maxPerDay,
      max_per_week: settingsForm.maxPerWeek
    })
    previewData.value = data
    // 保存表单记忆
    saveFormMemory()
  } catch (error) {
    // 错误已在拦截器处理
  } finally {
    loading.value = false
  }
}

// 确认排班
const confirmSchedule = async () => {
  if (!previewData.value) return

  confirming.value = true
  try {
    const cells = []
    for (let day = 0; day < 5; day++) {
      for (let period = 0; period < 4; period++) {
        const cell = previewData.value.grid[day]?.[period]
        if (cell && cell.users && cell.users.length > 0) {
          cells.push({
            weekday: day + 1,
            period: period + 1,
            user_ids: cell.users.map(u => u.id)
          })
        }
      }
    }

    await confirmScheduleAPI({
      week: previewData.value.week,
      cells
    })
    
    // 如果开启了自动递增，周次+1并保存记忆
    const settings = await getScheduleSettings()
    if (settings && settings.auto_increment) {
      const newWeek = form.week + 1
      if (newWeek <= 30) {
        form.week = newWeek
        saveFormMemory()
      }
    }
    
    ElMessage.success('排班确认成功')
    router.push('/schedule/result')
  } catch (error) {
    // 错误已在拦截器处理
  } finally {
    confirming.value = false
  }
}

// 删除人员
const removeUser = (row, user) => {
  const cell = previewData.value.grid[row.weekday - 1]?.[row.period - 1]
  if (cell) {
    const index = cell.users.findIndex(u => u.id === user.id)
    if (index > -1) {
      cell.users.splice(index, 1)
      ElMessage.success(`已移除 ${user.name}`)
    }
  }
}

// 显示添加人员对话框
const showAddUserDialog = (row) => {
  currentEditRow.value = row
  addForm.weekday = row.weekday
  addForm.weekdayText = row.weekdayText
  addForm.period = row.period
  addForm.selectedUsers = []
  showAddDialog.value = true
}

// 可选人员列表
const availableUserOptions = computed(() => {
  if (!currentEditRow.value) return []
  
  // 获取当前已选人员ID
  const cell = previewData.value.grid[currentEditRow.value.weekday - 1]?.[currentEditRow.value.period - 1]
  const selectedIds = cell?.users.map(u => u.id) || []
  
  // 过滤出可选人员
  return allUsers.value
    .filter(user => !selectedIds.includes(user.id))
    .map(user => ({
      value: user.id,
      label: `${user.name} (${user.student_id})`
    }))
})

// 确认添加人员
const confirmAddUsers = () => {
  if (!currentEditRow.value || addForm.selectedUsers.length === 0) {
    showAddDialog.value = false
    return
  }

  const cell = previewData.value.grid[currentEditRow.value.weekday - 1]?.[currentEditRow.value.period - 1]
  if (cell) {
    addForm.selectedUsers.forEach(userId => {
      const user = allUsers.value.find(u => u.id === userId)
      if (user && !cell.users.find(u => u.id === userId)) {
        cell.users.push(user)
      }
    })
    ElMessage.success('添加成功')
  }
  
  showAddDialog.value = false
}



const flattenGrid = computed(() => {
  if (!previewData.value || !previewData.value.grid) return []
  
  const result = []
  for (let day = 0; day < 5; day++) {
    for (let period = 0; period < 4; period++) {
      const cell = previewData.value.grid[day]?.[period]
      const users = cell?.users || []
      if (cell && (users.length > 0 || isInSelectedDaysAndPeriods(day + 1, period + 1))) {
        result.push({
          weekday: day + 1,
          weekdayText: '周' + ['一','二','三','四','五'][day],
          period: period + 1,
          users: users
        })
      }
    }
  }
  return result
})

const isInSelectedDaysAndPeriods = (weekday, period) => {
  return form.days.includes(weekday) && period <= form.periods
}

// 加载用户列表（用于排班选择）
const loadAllUsers = async () => {
  try {
    const data = await getUsersForSchedule()
    allUsers.value = data.map(user => ({
      id: user.id,
      name: user.name,
      student_id: user.student_id
    }))
  } catch (error) {
    allUsers.value = []
    console.error('加载用户列表失败:', error)
  }
}

// 加载模板列表
const loadTemplates = async () => {
  try {
    const data = await getTemplates()
    templates.value = data || []
    // 选中默认模板
    const defaultTemplate = templates.value.find(t => t.is_default)
    if (defaultTemplate) {
      selectedTemplateId.value = defaultTemplate.id
    } else if (templates.value.length > 0) {
      selectedTemplateId.value = templates.value[0].id
    }
  } catch (error) {
    templates.value = []
  }
}

// 模板操作
const addHeader = () => {
  templateForm.config.headers.push('')
}

const removeHeader = (index) => {
  templateForm.config.headers.splice(index, 1)
}

const addDataColumn = () => {
  templateForm.config.dataColumns.push({ type: 'text', format: '' })
}

const removeDataColumn = (index) => {
  templateForm.config.dataColumns.splice(index, 1)
}

const saveTemplate = async () => {
  savingTemplate.value = true
  try {
    if (editingTemplate.value) {
      await updateTemplateAPI({
        id: editingTemplate.value.id,
        ...templateForm
      })
      ElMessage.success('模板更新成功')
    } else {
      await createTemplate(templateForm)
      ElMessage.success('模板创建成功')
    }
    showTemplateDialog.value = false
    loadTemplates()
  } catch (error) {
    ElMessage.error('保存失败')
  } finally {
    savingTemplate.value = false
  }
}

const editTemplate = (template) => {
  editingTemplate.value = template
  templateForm.name = template.name
  templateForm.description = template.description || ''
  templateForm.is_default = template.is_default
  // 解析config
  if (template.config) {
    const config = typeof template.config === 'string' ? JSON.parse(template.config) : template.config
    templateForm.config.title = config.title || '第{week}周排班表'
    templateForm.config.mode = config.mode || 'list'
    templateForm.config.headers = config.headers || ['星期', '节次', '值班人员']
    templateForm.config.dataColumns = config.dataColumns || [
      { type: 'weekday', format: '周{weekday_cn}' },
      { type: 'period', format: '第{period}节' },
      { type: 'users', format: '{users}', separator: '、' }
    ]
    // 课表配置
    if (config.scheduleConfig) {
      templateForm.config.scheduleConfig = {
        rowHeader: config.scheduleConfig.rowHeader || '节次',
        colHeader: config.scheduleConfig.colHeader || '星期',
        rowLabels: config.scheduleConfig.rowLabels || ['第1节', '第2节', '第3节', '第4节'],
        colLabels: config.scheduleConfig.colLabels || ['周一', '周二', '周三', '周四', '周五'],
        cellFormat: config.scheduleConfig.cellFormat || '{users}',
        emptyCellText: config.scheduleConfig.emptyCellText || '-'
      }
    }
  }
  showTemplateDialog.value = true
}

const deleteTemplateById = async (id) => {
  try {
    await deleteTemplateAPI(id)
    ElMessage.success('删除成功')
    loadTemplates()
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

// 预览导出效果
const previewExportTemplate = () => {
  if (!previewData.value) {
    ElMessage.warning('请先生成排班预览')
    return
  }
  
  const template = templates.value.find(t => t.id === selectedTemplateId.value)
  if (!template) {
    ElMessage.warning('请选择模板')
    return
  }

  const config = typeof template.config === 'string' ? JSON.parse(template.config) : template.config
  
  // 生成预览标题
  previewTitle.value = (config.title || '第{week}周排班表')
    .replace('{week}', previewData.value.week)
    .replace('{department}', exportDepartment.value || 'XX部')
  
  // 课表模式预览
  if (config.mode === 'schedule') {
    previewScheduleTable(config)
    return
  }
  
  // 列表模式预览
  previewListTable(config)
}

// 列表模式预览
const previewListTable = (config) => {
  // 生成表头
  previewHeaders.value = config.headers || ['星期', '节次', '值班人员']
  
  // 生成预览数据
  const weekdayNames = ['一', '二', '三', '四', '五']
  previewDataList.value = []
  
  for (let day = 0; day < 5; day++) {
    for (let period = 0; period < 4; period++) {
      const cell = previewData.value.grid[day]?.[period]
      if (cell && cell.users && cell.users.length > 0) {
        const row = {}
        config.dataColumns?.forEach((col, idx) => {
          const key = 'col' + idx
          const users = cell.users.map(u => u.name).join('、')
          switch (col.type) {
            case 'weekday':
              row[key] = col.format?.replace('{weekday}', day + 1).replace('{weekday_cn}', weekdayNames[day]) || '周' + weekdayNames[day]
              break
            case 'period':
              row[key] = col.format?.replace('{period}', period + 1) || `第${period + 1}节`
              break
            case 'users':
              row[key] = users
              break
            default:
              row[key] = col.format || ''
          }
        })
        previewDataList.value.push(row)
      }
    }
  }
  
  showPreviewDialog.value = true
}

// 课表模式预览
const previewScheduleTable = (config) => {
  const sc = config.scheduleConfig || {
    rowLabels: ['第1节', '第2节', '第3节', '第4节'],
    colLabels: ['周一', '周二', '周三', '周四', '周五'],
    cellFormat: '{users}',
    emptyCellText: '-'
  }
  
  // 表头
  previewScheduleHeaders.value = ['节次\\星期', ...sc.colLabels]
  
  // 数据行
  previewScheduleData.value = []
  for (let period = 0; period < sc.rowLabels.length; period++) {
    const row = {
      label: sc.rowLabels[period]
    }
    for (let day = 0; day < sc.colLabels.length; day++) {
      const cell = previewData.value.grid[day]?.[period]
      const users = cell?.users?.map(u => u.name) || []
      
      if (users.length > 0) {
        let cellText = sc.cellFormat || '{users}'
        cellText = cellText.replace('{users}', users.join('、'))
        cellText = cellText.replace('{count}', users.length)
        row['day' + day] = cellText
      } else {
        row['day' + day] = sc.emptyCellText || '-'
      }
    }
    previewScheduleData.value.push(row)
  }
  
  showPreviewDialog.value = true
}

// 前端生成Excel并下载
const exportToExcel = () => {
  if (!previewData.value) return

  exporting.value = true
  try {
    const template = templates.value.find(t => t.id === selectedTemplateId.value)
    if (!template) {
      ElMessage.warning('请选择模板')
      exporting.value = false
      return
    }

    const config = typeof template.config === 'string' ? JSON.parse(template.config) : template.config
    const week = previewData.value.week
    const department = exportDepartment.value || 'XX部'

    // 根据模板模式生成数据
    let excelData = []
    
    // 标题行
    if (config.title) {
      const title = config.title.replace('{week}', week).replace('{department}', department)
      excelData.push([title])
      excelData.push([]) // 空行
    }

    if (config.mode === 'schedule') {
      // 课表格式
      excelData = excelData.concat(buildScheduleExcelData(config, previewData.value.grid))
    } else {
      // 列表格式
      excelData = excelData.concat(buildListExcelData(config, previewData.value.grid))
    }

    // 创建工作簿
    const ws = XLSX.utils.aoa_to_sheet(excelData)
    
    // 设置列宽
    if (!ws['!cols']) ws['!cols'] = []
    if (config.mode === 'schedule' && config.scheduleConfig) {
      ws['!cols'] = [{ wch: 12 }, ...config.scheduleConfig.colLabels.map(() => ({ wch: 20 }))]
    } else if (config.headers) {
      ws['!cols'] = config.headers.map(() => ({ wch: 20 }))
    }

    const wb = XLSX.utils.book_new()
    XLSX.utils.book_append_sheet(wb, ws, `第${week}周排班`)

    // 下载
    XLSX.writeFile(wb, `排班表_第${week}周.xlsx`)
    
    ElMessage.success('导出成功')
  } catch (error) {
    console.error('导出失败:', error)
    ElMessage.error('导出失败')
  } finally {
    exporting.value = false
  }
}

// 构建列表格式数据
const buildListExcelData = (config, grid) => {
  const data = []
  
  // 表头
  if (config.headers) {
    data.push(config.headers)
  }
  
  // 数据行
  const weekdayNames = ['一', '二', '三', '四', '五']
  for (let day = 0; day < 5; day++) {
    for (let period = 0; period < 4; period++) {
      const cell = grid[day]?.[period]
      if (cell && cell.users && cell.users.length > 0) {
        const users = cell.users.map(u => u.name).join('、')
        const row = config.dataColumns.map(col => {
          switch (col.type) {
            case 'weekday':
              return col.format?.replace('{weekday}', day + 1).replace('{weekday_cn}', weekdayNames[day]) || `周${weekdayNames[day]}`
            case 'period':
              return col.format?.replace('{period}', period + 1) || `第${period + 1}节`
            case 'users':
              return users
            default:
              return col.format || ''
          }
        })
        data.push(row)
      }
    }
  }
  
  return data
}

// 构建课表格式数据
const buildScheduleExcelData = (config, grid) => {
  const data = []
  const sc = config.scheduleConfig || {
    rowLabels: ['第1节', '第2节', '第3节', '第4节'],
    colLabels: ['周一', '周二', '周三', '周四', '周五'],
    cellFormat: '{users}',
    emptyCellText: '-'
  }
  
  // 表头
  data.push([`${sc.rowHeader}\\${sc.colHeader}`, ...sc.colLabels])
  
  // 数据行
  for (let period = 0; period < sc.rowLabels.length; period++) {
    const row = [sc.rowLabels[period]]
    for (let day = 0; day < sc.colLabels.length; day++) {
      const cell = grid[day]?.[period]
      const users = cell?.users?.map(u => u.name) || []
      
      if (users.length > 0) {
        let cellText = sc.cellFormat || '{users}'
        cellText = cellText.replace('{users}', users.join('、'))
        cellText = cellText.replace('{count}', users.length)
        row.push(cellText)
      } else {
        row.push(sc.emptyCellText || '-')
      }
    }
    data.push(row)
  }
  
  return data
}

onMounted(() => {
  loadSettings()
  loadAllUsers()
  loadTemplates()
  // 获取当前周次并设置
  getCurrentWeek().then(res => {
    if (res.data && res.data.current_week && form.week === 1) {
      form.week = res.data.current_week
    }
  }).catch(() => {})
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

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.user-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.user-tag {
  margin-right: 0;
}

.empty-text {
  color: #909399;
  font-size: 12px;
}

.main-content {
  min-height: calc(100vh - 60px);
  padding: 20px;
}

.page-container {
  max-width: 1400px;
  margin: 0 auto;
}

.settings-form {
  background: #f5f7fa;
  padding: 20px;
  border-radius: 4px;
}

.form-hint {
  margin-left: 10px;
  color: #909399;
  font-size: 12px;
}

.mt-4 {
  margin-top: 16px;
}

.header-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.column-config {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}

.preview-content {
  padding: 20px;
}

.preview-content h3 {
  text-align: center;
  margin-bottom: 20px;
}
</style>
