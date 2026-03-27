<template>
  <div class="page-container">
    <Navbar title="排班管理" show-back />

    <!-- 生成排班 -->
    <div class="section-card">
      <div class="card-header">
        <span class="card-title">生成排班</span>
        <van-tag v-if="assignmentDays.length" type="success" size="medium">
          分工：周{{ assignmentDays.map(d => weekNames[d-1]).join('、') }}
        </van-tag>
      </div>
      <van-cell-group inset>
        <van-cell title="周次" is-link @click="showWeekPicker = true">
          <template #value>第{{ form.week }}周</template>
        </van-cell>
        <van-cell title="部门" is-link @click="showDeptPicker = true" :disabled="isDepartmentLocked">
          <template #value>{{ form.department || '请选择' }}</template>
        </van-cell>
        <van-cell title="星期">
          <template #value>
            <div class="day-checkboxes">
              <van-tag 
                v-for="i in 5" 
                :key="i"
                :type="form.days.includes(i) ? 'primary' : 'default'"
                size="medium"
                class="day-tag"
                @click="toggleDay(i)"
              >
                周{{ weekNames[i-1] }}
              </van-tag>
            </div>
          </template>
        </van-cell>
        <van-cell title="节次">
          <template #value>
            <van-stepper v-model="form.periods" min="1" max="4" />
          </template>
        </van-cell>
      </van-cell-group>
      <div class="action-buttons">
        <van-button type="primary" :loading="loading && !isLoadingExisting" @click="generateNewSchedule">
          生成新排班
        </van-button>
        <van-button type="default" :loading="loading && isLoadingExisting" @click="loadExistingSchedule">
          查看已有排班
        </van-button>
      </div>
    </div>

    <!-- 排班预览 -->
    <div class="section-card" v-if="previewData">
      <div class="card-header">
        <span class="card-title">第{{ previewData.week }}周排班预览</span>
        <van-tag v-if="isEditing" type="warning" size="medium">编辑中</van-tag>
      </div>

      <van-notice-bar v-if="conflicts.length" :text="`以下${conflicts.length}个时段人数不足`" color="#f56c6c" background="#fff1f0" />

      <div class="preview-list">
        <div 
          v-for="row in previewList" 
          :key="`${row.weekday}-${row.period}`"
          class="preview-item"
        >
          <div class="item-header">
            <van-tag type="primary">周{{ weekNames[row.weekday-1] }}</van-tag>
            <span class="period">第{{ row.period }}节</span>
          </div>
          <div class="item-users">
            <van-tag 
              v-for="u in row.users" 
              :key="u.id" 
              size="large"
              closeable
              @close="removeUser(row, u)"
              class="user-tag"
            >
              {{ u.name }}
            </van-tag>
            <van-button size="small" type="primary" plain @click="showAddDialog(row)">+ 添加</van-button>
          </div>
        </div>
      </div>

      <div class="action-buttons">
        <van-button type="success" block :loading="confirming" @click="confirmSchedule">
          {{ isEditing ? '保存修改' : '确认排班' }}
        </van-button>
        <van-button type="primary" block :loading="exporting" @click="exportExcel">导出Excel</van-button>
      </div>
    </div>

    <!-- 折叠面板：排班参数 & 模板配置 -->
    <van-collapse v-model="collapseActive" class="collapse-section">
      <!-- 排班参数 -->
      <van-collapse-item title="排班参数" name="settings">
        <van-cell-group inset>
          <van-cell title="当前周次">
            <template #value>
              <van-stepper v-model="settings.current_week" min="1" max="30" />
            </template>
          </van-cell>
          <van-cell title="每时段人数">
            <template #value>
              <van-stepper v-model="settings.need_per_cell" min="1" max="10" />
            </template>
          </van-cell>
          <van-cell title="每时段最少">
            <template #value>
              <van-stepper v-model="settings.min_per_cell" min="0" max="10" />
            </template>
          </van-cell>
          <van-cell title="每人每天最多">
            <template #value>
              <van-stepper v-model="settings.max_per_day" min="1" max="10" />
            </template>
          </van-cell>
          <van-cell title="每人每周最多">
            <template #value>
              <van-stepper v-model="settings.max_per_week" min="1" max="30" />
            </template>
          </van-cell>
        </van-cell-group>
        <div class="collapse-action">
          <van-button type="primary" size="small" :loading="saving" @click="saveSettings">保存参数</van-button>
        </div>
      </van-collapse-item>

      <!-- 导出模板配置 -->
      <van-collapse-item title="导出模板配置" name="template">
        <van-cell-group inset>
          <van-cell title="选择模板" is-link @click="showTemplateSelect = true">
            <template #value>{{ currentTemplateName }}</template>
          </van-cell>
          <van-field v-model="exportDepartment" label="部门名称" placeholder="如：办公室" />
          <van-cell title="">
            <van-button type="success" size="small" :disabled="!previewData" @click="previewExportTemplate">预览效果</van-button>
          </van-cell>
        </van-cell-group>

        <!-- 模板列表 -->
        <div v-if="templates.length > 0" class="template-list">
          <van-swipe-cell v-for="t in templates" :key="t.id">
            <van-cell :title="t.name + (t.is_default ? ' (默认)' : '')" :label="t.description">
              <template #value>
                <van-button type="primary" size="small" plain @click="editTemplate(t)">编辑</van-button>
              </template>
            </van-cell>
            <template #right>
              <van-button v-if="!t.is_default" square type="danger" text="删除" @click="deleteTemplateById(t.id)" />
            </template>
          </van-swipe-cell>
        </div>
        <van-empty v-else description="暂无模板，请创建" />
        <div class="collapse-action">
          <van-button type="primary" size="small" plain @click="showTemplateDialog = true; editingTemplate = null; resetTemplateForm()">新建模板</van-button>
        </div>
      </van-collapse-item>
    </van-collapse>

    <!-- 周次选择器 -->
    <van-popup v-model:show="showWeekPicker" position="bottom" round>
      <van-picker
        :columns="weekPickerOptions"
        @confirm="onWeekConfirm"
        @cancel="showWeekPicker = false"
      />
    </van-popup>

    <!-- 部门选择器 -->
    <van-popup v-model:show="showDeptPicker" position="bottom" round>
      <van-picker
        :columns="deptOptions"
        @confirm="onDeptConfirm"
        @cancel="showDeptPicker = false"
      />
    </van-popup>

    <!-- 模板选择器 -->
    <van-popup v-model:show="showTemplateSelect" position="bottom" round>
      <van-picker
        :columns="templateOptions"
        @confirm="onTemplateSelect"
        @cancel="showTemplateSelect = false"
      />
    </van-popup>

    <!-- 添加人员 -->
    <van-popup v-model:show="addDialogVisible" position="bottom" round :style="{ height: '50%' }">
      <div class="popup-container">
        <div class="popup-title">{{ addForm.weekdayText }} 第{{ addForm.period }}节</div>
        <van-checkbox-group v-model="addForm.users">
          <van-cell-group inset>
            <van-cell 
              v-for="user in availableUsersList" 
              :key="user.value"
              clickable
              @click="toggleAddUser(user.value)"
            >
              <template #title>{{ user.label }}</template>
              <template #right-icon>
                <van-checkbox :name="user.value" :ref="el => {}" />
              </template>
            </van-cell>
          </van-cell-group>
        </van-checkbox-group>
        <div class="popup-actions">
          <van-button block @click="addDialogVisible = false">取消</van-button>
          <van-button type="primary" block @click="confirmAdd">添加</van-button>
        </div>
      </div>
    </van-popup>

    <!-- 模板编辑弹窗 -->
    <van-popup v-model:show="showTemplateDialog" position="bottom" round :style="{ height: '80%' }">
      <div class="popup-container template-popup">
        <div class="popup-header">
          <span class="popup-title">{{ editingTemplate ? '编辑模板' : '新建模板' }}</span>
          <van-icon name="cross" @click="showTemplateDialog = false" />
        </div>

        <van-cell-group inset>
          <van-field v-model="templateForm.name" label="模板名称" placeholder="输入模板名称" />
          <van-field v-model="templateForm.description" label="描述" placeholder="输入模板描述" />
          <van-cell title="导出格式">
            <template #value>
              <van-radio-group v-model="templateForm.config.mode" direction="horizontal">
                <van-radio name="list">列表</van-radio>
                <van-radio name="schedule">课表</van-radio>
              </van-radio-group>
            </template>
          </van-cell>
          <van-field v-model="templateForm.config.title" label="表格标题" placeholder="如：{department}第{week}周排班表" />
        </van-cell-group>

        <!-- 列表格式配置 -->
        <template v-if="templateForm.config.mode === 'list'">
          <div class="config-section">
            <div class="section-title">表头设置</div>
            <div class="header-list">
              <van-field 
                v-for="(header, index) in templateForm.config.headers" 
                :key="index"
                v-model="templateForm.config.headers[index]"
                placeholder="表头名称"
              >
                <template #button>
                  <van-button size="small" type="danger" plain @click="removeHeader(index)">删除</van-button>
                </template>
              </van-field>
            </div>
            <van-button size="small" type="primary" plain block @click="addHeader">添加表头</van-button>
          </div>
        </template>

        <!-- 课表格式配置 -->
        <template v-if="templateForm.config.mode === 'schedule'">
          <van-cell-group inset>
            <van-field v-model="scheduleRowLabels" label="行标签" placeholder="第1节,第2节,第3节,第4节" />
            <van-field v-model="scheduleColLabels" label="列标签" placeholder="周一,周二,周三,周四,周五" />
            <van-field v-model="templateForm.config.scheduleConfig.cellFormat" label="单元格格式" placeholder="{users}" />
            <van-field v-model="templateForm.config.scheduleConfig.emptyCellText" label="空单元格" placeholder="-" />
          </van-cell-group>
        </template>

        <van-cell-group inset>
          <van-cell title="设为默认">
            <template #right-icon>
              <van-switch v-model="templateForm.is_default" size="20" />
            </template>
          </van-cell>
        </van-cell-group>

        <div class="popup-actions">
          <van-button block @click="showTemplateDialog = false">取消</van-button>
          <van-button type="primary" block :loading="savingTemplate" @click="saveTemplate">保存</van-button>
        </div>
      </div>
    </van-popup>

    <!-- 预览效果弹窗 -->
    <van-popup v-model:show="showPreviewDialog" position="bottom" round :style="{ height: '70%' }">
      <div class="popup-container">
        <div class="popup-header">
          <span class="popup-title">导出预览</span>
          <van-icon name="cross" @click="showPreviewDialog = false" />
        </div>

        <h3 class="preview-title">{{ previewTitle }}</h3>
        
        <!-- 列表模式预览 -->
        <div v-if="previewDataList.length > 0" class="preview-table">
          <div class="preview-row header-row">
            <div v-for="(header, index) in previewHeaders" :key="index" class="preview-cell">{{ header }}</div>
          </div>
          <div v-for="(row, rowIndex) in previewDataList" :key="rowIndex" class="preview-row">
            <div v-for="(header, index) in previewHeaders" :key="index" class="preview-cell">{{ row['col' + index] }}</div>
          </div>
        </div>
        
        <!-- 课表模式预览 -->
        <div v-if="previewScheduleData.length > 0" class="preview-table">
          <div class="preview-row header-row">
            <div v-for="(header, index) in previewScheduleHeaders" :key="index" class="preview-cell">{{ header }}</div>
          </div>
          <div v-for="(row, rowIndex) in previewScheduleData" :key="rowIndex" class="preview-row">
            <div class="preview-cell">{{ row.label }}</div>
            <div v-for="(_, dayIndex) in previewScheduleHeaders.slice(1)" :key="dayIndex" class="preview-cell">
              {{ row['day' + dayIndex] }}
            </div>
          </div>
        </div>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { showToast, showSuccessToast, showFailToast } from 'vant'
import { useRouter } from 'vue-router'
import * as XLSX from 'xlsx'
import { useUserStore } from '../stores/user'
import {
  previewSchedule,
  confirmSchedule as confirmAPI,
  getScheduleSettings,
  saveScheduleSettings,
  getCurrentWeek,
  getMyDeptAssignment,
  getTemplates,
  createTemplate,
  updateTemplate as updateTemplateAPI,
  deleteTemplate as deleteTemplateAPI,
  exportToExcel,
  getSchedule
} from '../api/schedule'
import { getUsersForSchedule } from '../api/user'
import { shareSchedule, isNativeAvailable } from '../utils/native'
import Navbar from '../components/Navbar.vue'

const userStore = useUserStore()

const router = useRouter()
const loading = ref(false)
const confirming = ref(false)
const saving = ref(false)
const exporting = ref(false)
const previewData = ref(null)
const allUsers = ref([])
const addDialogVisible = ref(false)
const collapseActive = ref([])
const isEditing = ref(false) // 是否在编辑已有排班
const isLoadingExisting = ref(false) // 是否正在加载已有排班

const weekNames = ['一', '二', '三', '四', '五']

// 选择器
const showWeekPicker = ref(false)
const showDeptPicker = ref(false)
const showTemplateSelect = ref(false)

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

const settings = reactive({
  current_week: 1,
  need_per_cell: 2,
  min_per_cell: 0,
  max_per_day: 1,
  max_per_week: 2,
  export_title: '第{week}周排班表'
})

const form = reactive({
  week: 1,
  days: [1, 2, 3, 4, 5],
  periods: 4,
  department: ''
})

// 部门分工数据
const myDeptAssignment = ref(null)

const assignmentDays = computed(() => {
  if (!myDeptAssignment.value?.weekdays) return []
  return myDeptAssignment.value.weekdays
    .filter(w => w.is_assigned)
    .map(w => w.weekday)
})

const weekPickerOptions = computed(() => {
  return Array.from({ length: 30 }, (_, i) => ({
    text: `第${i + 1}周`,
    value: i + 1
  }))
})

const deptOptions = [
  { text: '办公室', value: '办公室' },
  { text: '竞赛部', value: '竞赛部' },
  { text: '项目部', value: '项目部' },
  { text: '科普部', value: '科普部' }
]

const templateOptions = computed(() => {
  return templates.value.map(t => ({
    text: t.name + (t.is_default ? ' (默认)' : ''),
    value: t.id
  }))
})

const currentTemplateName = computed(() => {
  const t = templates.value.find(t => t.id === selectedTemplateId.value)
  return t ? t.name + (t.is_default ? ' (默认)' : '') : '请选择'
})

const onWeekConfirm = ({ selectedOptions }) => {
  form.week = selectedOptions[0].value
  showWeekPicker.value = false
  loadMyDeptAssignment()
}

const onDeptConfirm = ({ selectedOptions }) => {
  form.department = selectedOptions[0].value
  showDeptPicker.value = false
}

const onTemplateSelect = ({ selectedOptions }) => {
  selectedTemplateId.value = selectedOptions[0].value
  showTemplateSelect.value = false
}

const toggleDay = (day) => {
  const idx = form.days.indexOf(day)
  if (idx > -1) {
    form.days.splice(idx, 1)
  } else {
    form.days.push(day)
  }
}

const loadMyDeptAssignment = async () => {
  try {
    const data = await getMyDeptAssignment({ week: form.week })
    myDeptAssignment.value = data
  } catch (error) {
    console.error('加载部门分工失败:', error)
    myDeptAssignment.value = null
  }
}

const addForm = reactive({
  weekday: 1,
  weekdayText: '',
  period: 1,
  users: [],
  rowData: null
})

const previewList = computed(() => {
  if (!previewData.value?.grid) return []
  const list = []
  for (let d = 0; d < 5; d++) {
    for (let p = 0; p < 4; p++) {
      const cell = previewData.value.grid[d]?.[p]
      if (cell && form.days.includes(d + 1) && p < form.periods) {
        list.push({ weekday: d + 1, period: p + 1, users: cell.users || [] })
      }
    }
  }
  return list
})

const conflicts = computed(() => {
  if (!previewData.value?.conflicts) return []
  return previewData.value.conflicts.filter(c => form.days.includes(c.weekday) && c.period <= form.periods)
})

const availableUsersList = computed(() => {
  if (!addForm.rowData) return []
  const selected = new Set(addForm.rowData.users.map(u => u.id))
  return allUsers.value.filter(u => !selected.has(u.id)).map(u => ({ value: u.id, label: `${u.name} (${u.student_id})` }))
})

const isDepartmentLocked = computed(() => {
  return !userStore.isAdmin && !userStore.isOfficeAdmin && !!userStore.user?.department
})

const loadSettings = async () => {
  try {
    const data = await getScheduleSettings()
    if (data) Object.assign(settings, data)
  } catch {}
}

const loadUsers = async () => {
  try {
    const data = await getUsersForSchedule()
    allUsers.value = data || []
  } catch {}
}

const loadTemplates = async () => {
  try {
    const data = await getTemplates()
    templates.value = data || []
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

const addHeader = () => {
  templateForm.config.headers.push('')
}

const removeHeader = (index) => {
  templateForm.config.headers.splice(index, 1)
}

const saveTemplate = async () => {
  savingTemplate.value = true
  try {
    if (editingTemplate.value) {
      await updateTemplateAPI({
        id: editingTemplate.value.id,
        ...templateForm
      })
      showSuccessToast('模板更新成功')
    } else {
      await createTemplate(templateForm)
      showSuccessToast('模板创建成功')
    }
    showTemplateDialog.value = false
    loadTemplates()
  } catch (error) {
    showFailToast('保存失败')
  } finally {
    savingTemplate.value = false
  }
}

const editTemplate = (template) => {
  editingTemplate.value = template
  templateForm.name = template.name
  templateForm.description = template.description || ''
  templateForm.is_default = template.is_default
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
    showSuccessToast('删除成功')
    loadTemplates()
  } catch (error) {
    showFailToast('删除失败')
  }
}

const previewExportTemplate = () => {
  if (!previewData.value) {
    showToast({ message: '请先生成排班预览', type: 'fail' })
    return
  }
  
  const template = templates.value.find(t => t.id === selectedTemplateId.value)
  if (!template) {
    showToast({ message: '请选择模板', type: 'fail' })
    return
  }

  const config = typeof template.config === 'string' ? JSON.parse(template.config) : template.config
  
  previewTitle.value = (config.title || '第{week}周排班表')
    .replace('{week}', previewData.value.week)
    .replace('{department}', exportDepartment.value || 'XX部')
  
  if (config.mode === 'schedule') {
    previewScheduleTable(config)
    return
  }
  
  previewListTable(config)
}

const previewListTable = (config) => {
  previewHeaders.value = config.headers || ['星期', '节次', '值班人员']
  previewScheduleData.value = []
  
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

const previewScheduleTable = (config) => {
  const sc = config.scheduleConfig || {
    rowLabels: ['第1节', '第2节', '第3节', '第4节'],
    colLabels: ['周一', '周二', '周三', '周四', '周五'],
    cellFormat: '{users}',
    emptyCellText: '-'
  }
  
  previewHeaders.value = []
  previewDataList.value = []
  previewScheduleHeaders.value = ['节次\\星期', ...sc.colLabels]
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

const buildListExcelData = (config, grid) => {
  const data = []
  if (config.headers) {
    data.push(config.headers)
  }
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

const buildScheduleExcelData = (config, grid) => {
  const data = []
  const sc = config.scheduleConfig || {
    rowLabels: ['第1节', '第2节', '第3节', '第4节'],
    colLabels: ['周一', '周二', '周三', '周四', '周五'],
    cellFormat: '{users}',
    emptyCellText: '-'
  }
  
  data.push([`${sc.rowHeader}\\${sc.colHeader}`, ...sc.colLabels])
  
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

// 生成新排班
const generateNewSchedule = async () => {
  if (!form.week || !form.department || form.days.length === 0) {
    showToast({ message: '请完善排班参数', type: 'fail' })
    return
  }

  loading.value = true
  isEditing.value = false
  isLoadingExisting.value = false

  try {
    const data = await previewSchedule({
      week: form.week,
      days: form.days,
      periods: form.periods,
      department: form.department,
      need_per_cell: settings.need_per_cell,
      min_per_cell: settings.min_per_cell,
      max_per_day: settings.max_per_day,
      max_per_week: settings.max_per_week
    })
    previewData.value = data
    showSuccessToast('已生成排班预览')
  } catch {
    showFailToast('生成预览失败')
  } finally {
    loading.value = false
  }
}

// 加载已有排班
const loadExistingSchedule = async () => {
  if (!form.week || !form.department) {
    showToast({ message: '请选择周次和部门', type: 'fail' })
    return
  }

  loading.value = true
  isLoadingExisting.value = true

  try {
    const records = await getSchedule({ week: form.week, department: form.department })

    if (records && records.length > 0) {
      // 将记录列表转换为 grid 格式
      const grid = []
      for (let d = 0; d < 5; d++) {
        grid[d] = []
        for (let p = 0; p < 4; p++) {
          grid[d][p] = { weekday: d + 1, period: p + 1, users: [] }
        }
      }

      for (const r of records) {
        const d = r.weekday - 1
        const p = r.period - 1
        if (d >= 0 && d < 5 && p >= 0 && p < 4) {
          grid[d][p].users.push({
            id: r.user_id,
            name: r.user_name
          })
        }
      }

      previewData.value = {
        week: form.week,
        department: form.department,
        grid,
        conflicts: []
      }
      isEditing.value = true
      showSuccessToast('已加载已有排班数据')
    } else {
      showToast({ message: '该周暂无排班数据', type: 'fail' })
    }
  } catch {
    showFailToast('加载失败')
  } finally {
    loading.value = false
    isLoadingExisting.value = false
  }
}

const saveSettings = async () => {
  saving.value = true
  try {
    await saveScheduleSettings(settings)
    showSuccessToast('保存成功')
  } catch {
    showFailToast('保存失败')
  } finally {
    saving.value = false
  }
}

const confirmSchedule = async () => {
  confirming.value = true
  try {
    const cells = []
    for (let d = 0; d < 5; d++) {
      for (let p = 0; p < 4; p++) {
        const cell = previewData.value.grid[d]?.[p]
        if (cell?.users?.length) {
          cells.push({ weekday: d + 1, period: p + 1, user_ids: cell.users.map(u => u.id) })
        }
      }
    }

    // confirmAPI 是覆盖式的，先删除整周数据再插入，可用于新建和编辑
    await confirmAPI({ week: previewData.value.week, cells })
    showSuccessToast(isEditing.value ? '排班修改成功' : '排班确认成功')
  } catch {
    showFailToast(isEditing.value ? '修改失败' : '确认失败')
  } finally {
    confirming.value = false
  }
}

const removeUser = (row, user) => {
  const idx = row.users.findIndex(u => u.id === user.id)
  if (idx > -1) row.users.splice(idx, 1)
}

const showAddDialog = (row) => {
  addForm.weekday = row.weekday
  addForm.weekdayText = '周' + weekNames[row.weekday - 1]
  addForm.period = row.period
  addForm.users = []
  addForm.rowData = row
  addDialogVisible.value = true
}

const toggleAddUser = (userId) => {
  const idx = addForm.users.indexOf(userId)
  if (idx > -1) {
    addForm.users.splice(idx, 1)
  } else {
    addForm.users.push(userId)
  }
}

const confirmAdd = () => {
  addForm.users.forEach(uid => {
    const user = allUsers.value.find(u => u.id === uid)
    if (user && !addForm.rowData.users.find(u => u.id === uid)) {
      addForm.rowData.users.push(user)
    }
  })
  addDialogVisible.value = false
  showSuccessToast('添加成功')
}

const exportExcel = async () => {
  if (!previewData.value) return

  exporting.value = true
  const week = previewData.value.week
  const fileName = `排班表_第${week}周.xlsx`

  try {
    // 统一前端生成 Excel
    const template = templates.value.find(t => t.id === selectedTemplateId.value)
    let excelData = []

    if (!template) {
      // 无模板，使用默认格式
      excelData.push(['星期', '节次', '值班人员'])
      previewList.value.forEach(row => {
        excelData.push([
          '周' + weekNames[row.weekday - 1],
          `第${row.period}节`,
          row.users.map(u => u.name).join('、')
        ])
      })
    } else {
      const config = typeof template.config === 'string' ? JSON.parse(template.config) : template.config
      const department = exportDepartment.value || 'XX部'

      if (config.title) {
        const title = config.title.replace('{week}', week).replace('{department}', department)
        excelData.push([title])
        excelData.push([])
      }

      if (config.mode === 'schedule') {
        excelData = excelData.concat(buildScheduleExcelData(config, previewData.value.grid))
      } else {
        excelData = excelData.concat(buildListExcelData(config, previewData.value.grid))
      }
    }

    // 创建工作簿和工作表
    const ws = XLSX.utils.aoa_to_sheet(excelData)
    const templateConfig = template ? (typeof template.config === 'string' ? JSON.parse(template.config) : template.config) : null

    if (!ws['!cols']) ws['!cols'] = []
    if (templateConfig?.mode === 'schedule' && templateConfig.scheduleConfig) {
      ws['!cols'] = [{ wch: 12 }, ...templateConfig.scheduleConfig.colLabels.map(() => ({ wch: 20 }))]
    } else if (templateConfig?.headers) {
      ws['!cols'] = templateConfig.headers.map(() => ({ wch: 20 }))
    } else {
      ws['!cols'] = [{ wch: 10 }, { wch: 10 }, { wch: 30 }]
    }

    const wb = XLSX.utils.book_new()
    XLSX.utils.book_append_sheet(wb, ws, `第${week}周排班`)

    // 根据环境选择输出方式
    if (isNativeAvailable()) {
      // 安卓原生环境：生成 base64 并调用分享
      const wbout = XLSX.write(wb, { bookType: 'xlsx', type: 'base64' })
      await shareSchedule({
        week,
        fileName,
        fileData: wbout
      })
      showSuccessToast('导出成功')
    } else {
      // 浏览器环境：直接下载
      XLSX.writeFile(wb, fileName)
      showSuccessToast('导出成功')
    }
  } catch (error) {
    console.error('导出失败:', error)
    showFailToast(error.message || '导出失败')
  } finally {
    exporting.value = false
  }
}

const resetTemplateForm = () => {
  templateForm.name = ''
  templateForm.description = ''
  templateForm.is_default = false
  templateForm.config = {
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
}

onMounted(async () => {
  loadSettings()
  loadUsers()
  loadTemplates()

  // 设置部门默认值
  form.department = userStore.user.department
  exportDepartment.value = userStore.user.department

  // 获取当前周次，自动设置为下一周
  try {
    const res = await getCurrentWeek()
    const currentWeek = res?.current_week || 1
    form.week = currentWeek + 1
    if (form.week > 30) form.week = 30
  } catch (e) {
    console.error('获取当前周次失败:', e)
    form.week = 1
  }

  // 加载分工数据并自动应用
  await loadMyDeptAssignment()
  if (assignmentDays.value.length) {
    form.days = [...assignmentDays.value]
  }
})
</script>

<style scoped>
.page-container {
  min-height: 100%;
  background: #f7f8fa;
  padding-top: env(safe-area-inset-top);
  padding-bottom: calc(20px + env(safe-area-inset-bottom));
}

.section-card {
  background: #fff;
  margin: 12px;
  border-radius: 12px;
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

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #323233;
}

.day-checkboxes {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.day-tag {
  cursor: pointer;
}

.action-buttons {
  padding: 16px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.action-buttons.single {
  grid-template-columns: 1fr;
}

/* 折叠面板 */
.collapse-section {
  margin: 12px;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.collapse-action {
  padding: 12px 16px;
  display: flex;
  justify-content: flex-end;
}

.template-list {
  margin-top: 8px;
}

.preview-list {
  padding: 12px;
}

.preview-item {
  background: #f5f5f5;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 12px;
}

.item-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.period {
  font-size: 15px;
  font-weight: 500;
  color: #323233;
}

.item-users {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.user-tag {
  margin-right: 4px;
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
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-top: 20px;
}

.template-popup {
  padding-bottom: 80px;
}

.config-section {
  margin: 16px 0;
  padding: 12px;
  background: #f5f5f5;
  border-radius: 8px;
}

.config-section .section-title {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 12px;
  color: #323233;
}

.header-list .van-field {
  margin-bottom: 8px;
}

/* 预览表格 */
.preview-title {
  text-align: center;
  margin-bottom: 16px;
  font-size: 16px;
  font-weight: 600;
}

.preview-table {
  overflow-x: auto;
}

.preview-row {
  display: flex;
  border-bottom: 1px solid #e8e8e8;
}

.preview-row.header-row {
  background: #f5f5f5;
  font-weight: 500;
}

.preview-cell {
  flex: 1;
  min-width: 80px;
  padding: 10px 8px;
  font-size: 13px;
  text-align: center;
  border-right: 1px solid #e8e8e8;
}

.preview-cell:last-child {
  border-right: none;
}
</style>
