<template>
  <div class="main-container">
    <!-- 排班参数 -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">排班参数</span>
        <el-button type="primary" @click="saveSettings" :loading="saving">
          <el-icon><Check /></el-icon> 保存
        </el-button>
      </div>
      <el-form :model="settings" label-width="140px">
        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="当前周次">
              <el-input-number v-model="settings.current_week" :min="1" :max="30" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="每时段人数">
              <el-input-number v-model="settings.need_per_cell" :min="1" :max="10" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="每时段最少">
              <el-input-number v-model="settings.min_per_cell" :min="0" :max="10" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="每人每天最多">
              <el-input-number v-model="settings.max_per_day" :min="1" :max="10" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="每人每周最多">
              <el-input-number v-model="settings.max_per_week" :min="1" :max="30" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
    </div>

    <!-- 本周分工 -->
    <div class="card" v-if="myDeptAssignment">
      <div class="card-header">
        <span class="card-title">第{{ form.week }}周 {{ userStore.user?.department }} 分工安排</span>
        <el-tag v-if="myDeptAssignment.days?.length" type="success">已分配</el-tag>
        <el-tag v-else type="warning">未分配</el-tag>
      </div>
      <div v-if="myDeptAssignment.days?.length" class="assignment-info">
        <p><strong>值班日期：</strong>周{{ myDeptAssignment.days.map(d => ['一','二','三','四','五'][d-1]).join('、') }}</p>
        <p><strong>任务分配：</strong>{{ myDeptAssignment.task || '暂无任务说明' }}</p>
        <el-button type="primary" size="small" @click="applyAssignmentDays">
          <el-icon><Check /></el-icon> 应用分工日期
        </el-button>
      </div>
      <el-empty v-else description="本周暂未分配值班日期，请联系办公室管理员" />
    </div>

    <!-- 生成排班 -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">生成排班</span>
      </div>
      <el-form :model="form" :rules="rules" ref="formRef" label-width="80px" inline>
        <el-form-item label="周次" prop="week">
          <el-select-v2 v-model="form.week" :options="weekOptions" style="width: 120px" @change="onWeekChange" />
        </el-form-item>
        <el-form-item label="部门" prop="department">
          <el-select v-model="form.department" placeholder="选择部门" style="width: 120px" :disabled="isDepartmentLocked">
            <el-option label="办公室" value="办公室" />
            <el-option label="竞赛部" value="竞赛部" />
            <el-option label="项目部" value="项目部" />
            <el-option label="科普部" value="科普部" />
          </el-select>
          <span v-if="isDepartmentLocked" class="form-hint ml-2">已自动选择您的部门</span>
        </el-form-item>
        <el-form-item label="星期" prop="days">
          <el-checkbox-group v-model="form.days">
            <el-checkbox v-for="i in 5" :key="i" :value="i">周{{ ['一','二','三','四','五'][i-1] }}</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="节次" prop="periods">
          <el-radio-group v-model="form.periods">
            <el-radio-button v-for="i in 4" :key="i" :value="i">{{ i }}节</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="generatePreview" :loading="loading">生成预览</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 导出模板配置 -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">导出模板配置</span>
        <el-button type="primary" @click="showTemplateDialog = true; editingTemplate = null; resetTemplateForm()">
          <el-icon><Plus /></el-icon> 新建模板
        </el-button>
      </div>
      
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
          <el-button type="success" @click="previewExportTemplate" :disabled="!previewData">
            <el-icon><View /></el-icon> 预览效果
          </el-button>
        </el-form-item>
      </el-form>

      <!-- 模板列表 -->
      <el-table :data="templates" v-if="templates.length > 0" class="mt-4">
        <el-table-column prop="name" label="模板名称" width="150" />
        <el-table-column prop="description" label="描述" show-overflow-tooltip />
        <el-table-column label="操作" width="150">
          <template #default="{ row }">
            <el-button link type="primary" @click="editTemplate(row)">编辑</el-button>
            <el-button link type="danger" @click="deleteTemplateById(row.id)" v-if="!row.is_default">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-else description="暂无模板，请创建" />
    </div>

    <!-- 排班预览 -->
    <div class="card" v-if="previewData">
      <div class="card-header">
        <span class="card-title">第{{ previewData.week }}周排班预览</span>
        <div class="action-bar">
          <el-button type="success" @click="confirmSchedule" :loading="confirming">确认排班</el-button>
          <el-button type="primary" @click="exportExcel" :loading="exporting">导出Excel</el-button>
        </div>
      </div>

      <el-alert v-if="conflicts.length" :title="`以下${conflicts.length}个时段人数不足`" type="error" :closable="false" class="mb-4" />
      
      <el-table :data="previewList" border>
        <el-table-column label="星期" width="80">
          <template #default="{ row }">周{{ ['一','二','三','四','五'][row.weekday-1] }}</template>
        </el-table-column>
        <el-table-column label="节次" width="80">
          <template #default="{ row }">第{{ row.period }}节</template>
        </el-table-column>
        <el-table-column label="值班人员">
          <template #default="{ row }">
            <el-space wrap>
              <el-tag v-for="u in row.users" :key="u.id" closable @close="removeUser(row, u)" size="small">{{ u.name }}</el-tag>
              <el-button size="small" type="primary" plain @click="showAddDialog(row)">+ 添加</el-button>
            </el-space>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 添加人员 -->
    <el-dialog v-model="addDialogVisible" title="添加值班人员" width="400px">
      <p class="mb-4">{{ addForm.weekdayText }} 第{{ addForm.period }}节</p>
      <el-select-v2 v-model="addForm.users" :options="availableUsers" multiple style="width: 100%" />
      <template #footer>
        <el-button @click="addDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmAdd">添加</el-button>
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
            <el-radio-button value="list">列表格式</el-radio-button>
            <el-radio-button value="schedule">课表格式</el-radio-button>
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
import { ElMessage } from 'element-plus'
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
  deleteTemplate as deleteTemplateAPI
} from '../api/schedule'
import { getUsersForSchedule } from '../api/user'

const userStore = useUserStore()

const router = useRouter()
const formRef = ref()
const loading = ref(false)
const confirming = ref(false)
const saving = ref(false)
const exporting = ref(false)
const previewData = ref(null)
const allUsers = ref([])
const addDialogVisible = ref(false)

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

// 加载部门分工
const loadMyDeptAssignment = async () => {
  try {
    const data = await getMyDeptAssignment({ week: form.week })
    myDeptAssignment.value = data
    // 如果有分工日期，默认选中
    if (data?.days?.length && form.days.length === 5) {
      // 只在没有手动修改时才自动应用
    }
  } catch (error) {
    console.error('加载部门分工失败:', error)
    myDeptAssignment.value = null
  }
}

// 应用分工日期到排班
const applyAssignmentDays = () => {
  if (myDeptAssignment.value?.days?.length) {
    form.days = [...myDeptAssignment.value.days]
    ElMessage.success('已应用分工日期：周' + form.days.map(d => ['一','二','三','四','五'][d-1]).join('、'))
  }
}

// 周次变化时重新加载分工
const onWeekChange = () => {
  loadMyDeptAssignment()
}

const addForm = reactive({
  weekday: 1,
  weekdayText: '',
  period: 1,
  users: [],
  rowData: null
})

const rules = {
  week: [{ required: true }],
  department: [{ required: true, message: '请选择部门' }],
  days: [{ required: true, type: 'array', min: 1 }],
  periods: [{ required: true }]
}

const weekOptions = Array.from({ length: 30 }, (_, i) => ({ value: i + 1, label: `第${i + 1}周` }))

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

const availableUsers = computed(() => {
  if (!addForm.rowData) return []
  const selected = new Set(addForm.rowData.users.map(u => u.id))
  return allUsers.value.filter(u => !selected.has(u.id)).map(u => ({ value: u.id, label: `${u.name} (${u.student_id})` }))
})

// 部门是否锁定（非管理员自动锁定为自己的部门）
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
      need_per_cell: settings.need_per_cell,
      min_per_cell: settings.min_per_cell,
      max_per_day: settings.max_per_day,
      max_per_week: settings.max_per_week
    })
    previewData.value = data
  } catch {}
  loading.value = false
}

const saveSettings = async () => {
  saving.value = true
  try {
    await saveScheduleSettings(settings)
    ElMessage.success('保存成功')
  } catch {}
  saving.value = false
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
    await confirmAPI({ week: previewData.value.week, cells })
    ElMessage.success('排班确认成功')
    router.push('/schedule/result')
  } catch {}
  confirming.value = false
}

const removeUser = (row, user) => {
  const idx = row.users.findIndex(u => u.id === user.id)
  if (idx > -1) row.users.splice(idx, 1)
}

const showAddDialog = (row) => {
  addForm.weekday = row.weekday
  addForm.weekdayText = '周' + ['一', '二', '三', '四', '五'][row.weekday - 1]
  addForm.period = row.period
  addForm.users = []
  addForm.rowData = row
  addDialogVisible.value = true
}

const confirmAdd = () => {
  addForm.users.forEach(uid => {
    const user = allUsers.value.find(u => u.id === uid)
    if (user && !addForm.rowData.users.find(u => u.id === uid)) {
      addForm.rowData.users.push(user)
    }
  })
  addDialogVisible.value = false
  ElMessage.success('添加成功')
}

const exportExcel = () => {
  if (!previewData.value) return

  exporting.value = true
  try {
    const template = templates.value.find(t => t.id === selectedTemplateId.value)
    if (!template) {
      // 如果没有模板，使用默认导出
      const data = [['星期', '节次', '值班人员']]
      previewList.value.forEach(row => {
        data.push([
          '周' + ['一', '二', '三', '四', '五'][row.weekday - 1],
          `第${row.period}节`,
          row.users.map(u => u.name).join('、')
        ])
      })
      const ws = XLSX.utils.aoa_to_sheet(data)
      const wb = XLSX.utils.book_new()
      XLSX.utils.book_append_sheet(wb, ws, `第${previewData.value.week}周排班`)
      XLSX.writeFile(wb, `排班表_第${previewData.value.week}周.xlsx`)
      ElMessage.success('导出成功')
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

// 重置模板表单
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
  // 非管理员默认选择自己部门
  if (!userStore.isAdmin && !userStore.isOfficeAdmin && userStore.user?.department) {
    form.department = userStore.user.department
  }
  // 获取当前周次并设置
  try {
    const res = await getCurrentWeek()
    console.log('Schedule - getCurrentWeek res:', res)
    // 后端返回的数据在 data 字段中
    const currentWeek = res?.data?.current_week || res?.current_week
    if (currentWeek) {
      form.week = currentWeek
      console.log('Schedule - set form.week to:', form.week)
    }
  } catch (e) {
    console.error('获取当前周次失败:', e)
  }
  // 加载部门分工
  loadMyDeptAssignment()
})
</script>

<style scoped>
.assignment-info {
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
  line-height: 2;
}

.assignment-info p {
  margin: 8px 0;
}

.mb-4 {
  margin-bottom: 16px;
}

.mt-4 {
  margin-top: 16px;
}

.form-hint {
  margin-left: 10px;
  color: #909399;
  font-size: 12px;
  line-height: 1.5;
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

.ml-2 {
  margin-left: 8px;
}
</style>
