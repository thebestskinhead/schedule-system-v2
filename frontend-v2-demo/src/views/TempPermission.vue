<template>
  <div class="main-container">
    <!-- 统计卡片 -->
    <div class="stats-cards" v-if="stats">
      <div class="stat-card" @click="activeTab = 'myApplications'" :class="{ active: activeTab === 'myApplications' }">
        <div class="stat-value">{{ stats.my_applications?.pending || 0 }}</div>
        <div class="stat-label">我的待审批</div>
      </div>
      <div class="stat-card" @click="activeTab = 'pendingApprovals'" :class="{ active: activeTab === 'pendingApprovals' }" v-if="canApprove">
        <div class="stat-value stat-warning">{{ stats.pending_approval || 0 }}</div>
        <div class="stat-label">待我审批</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ stats.my_applications?.approved || 0 }}</div>
        <div class="stat-label">已通过</div>
      </div>
    </div>

    <!-- 操作按钮区 -->
    <div class="action-bar">
      <el-radio-group v-model="activeTab" size="large">
        <el-radio-button value="permissions">权限管理</el-radio-button>
        <el-radio-button value="myApplications">我的申请</el-radio-button>
        <el-radio-button value="pendingApprovals" v-if="canApprove">
          待我审批
          <el-badge v-if="stats?.pending_approval" :value="stats.pending_approval" class="tab-badge" />
        </el-radio-button>
      </el-radio-group>

      <el-button type="primary" @click="openGrantDialog" v-if="activeTab === 'permissions' && canGrant">
        <el-icon><Plus /></el-icon> 授予权限
      </el-button>
      <el-button type="primary" @click="openRequestDialog" v-if="activeTab === 'permissions' && !canGrant">
        <el-icon><Plus /></el-icon> 申请权限
      </el-button>
    </div>

    <!-- 权限管理 Tab -->
    <div v-show="activeTab === 'permissions'" class="card">
      <div class="card-header">
        <span class="card-title">{{ canGrant ? '临时权限管理' : '我的临时权限' }}</span>
      </div>
      <el-table :data="permissionList" v-loading="loading" class="data-table">
        <el-table-column prop="user_name" label="用户" width="120" v-if="canGrant" />
        <el-table-column prop="user_department" label="部门" width="100" v-if="canGrant">
          <template #default="{ row }">
            <el-tag size="small">{{ row.user_department || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="权限类型" width="180">
          <template #default="{ row }">
            <el-tag size="small" :type="getPermissionType(row.permission)">{{ getPermissionText(row.permission) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="granted_by_name" label="授权人" width="120" />
        <el-table-column label="有效期" width="180">
          <template #default="{ row }">
            <span :class="{ 'text-danger': isExpired(row.expires_at) }">
              {{ formatTime(row.expires_at) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="isExpired(row.expires_at) ? 'info' : 'success'">
              {{ isExpired(row.expires_at) ? '已过期' : '有效' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" v-if="canGrant">
          <template #default="{ row }">
            <el-button link type="danger" @click="revokePermission(row)">撤销</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 我的申请 Tab -->
    <div v-show="activeTab === 'myApplications'" class="card">
      <div class="card-header">
        <span class="card-title">我的申请记录</span>
      </div>
      <el-table :data="myApplications" v-loading="loadingApps" class="data-table">
        <el-table-column prop="application_no" label="申请编号" width="150" />
        <el-table-column label="申请类型" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ getAppTypeName(row.type_code) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="申请内容" min-width="200">
          <template #default="{ row }">
            <div v-if="row.data">
              <div v-if="parseData(row.data).permission">
                权限: {{ getPermissionText(parseData(row.data).permission) }}
              </div>
              <div class="text-gray">{{ row.content || row.reason }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="申请时间" width="160">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button link type="primary" @click="viewDetail(row)">详情</el-button>
            <el-button link type="danger" @click="cancelApp(row)" v-if="row.status === 0">撤销</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="appPage"
          v-model:page-size="appPageSize"
          :total="appTotal"
          layout="prev, pager, next"
          @current-change="loadMyApplications"
        />
      </div>
    </div>

    <!-- 待我审批 Tab -->
    <div v-show="activeTab === 'pendingApprovals'" class="card" v-if="canApprove">
      <div class="card-header">
        <span class="card-title">待我审批的申请</span>
      </div>
      <el-empty v-if="pendingList.length === 0 && !loadingPending" description="暂无待审批申请" />
      <el-table :data="pendingList" v-loading="loadingPending" class="data-table" v-else>
        <el-table-column prop="application_no" label="申请编号" width="150" />
        <el-table-column prop="applicant_name" label="申请人" width="120" />
        <el-table-column label="部门" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ row.department || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="申请类型" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ getAppTypeName(row.type_code) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="申请内容" min-width="200">
          <template #default="{ row }">
            <div v-if="row.data">
              <div v-if="parseData(row.data).permission">
                权限: {{ getPermissionText(parseData(row.data).permission) }}
              </div>
              <div v-if="parseData(row.data).expiry_date">
                期望到期: {{ formatDate(parseData(row.data).expiry_date) }}
              </div>
            </div>
            <div class="text-gray">{{ row.content }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="申请时间" width="160">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="success" @click="approveApp(row)">通过</el-button>
            <el-button size="small" type="danger" @click="rejectApp(row)">拒绝</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pendingPage"
          v-model:page-size="pendingPageSize"
          :total="pendingTotal"
          layout="prev, pager, next"
          @current-change="loadPendingApprovals"
        />
      </div>
    </div>

    <!-- 权限申请对话框（普通成员使用） -->
    <el-dialog v-model="requestDialogVisible" title="申请临时权限" width="450px">
      <el-alert
        type="info"
        :closable="false"
        description="请填写权限申请信息，提交后将发送给您的部门管理员审批"
        style="margin-bottom: 16px"
      />
      <el-form :model="requestForm" :rules="requestRules" ref="requestFormRef" label-width="100px">
        <el-form-item label="申请权限" prop="permission">
          <el-select v-model="requestForm.permission" placeholder="选择需要的权限" style="width: 100%">
            <el-option 
              v-for="perm in availablePermissions" 
              :key="perm.key" 
              :label="perm.name" 
              :value="perm.key" 
            />
          </el-select>
        </el-form-item>
        <el-form-item label="申请原因" prop="reason">
          <el-input 
            v-model="requestForm.reason" 
            type="textarea" 
            rows="3"
            placeholder="请详细说明申请原因（如：因XX原因需要临时管理权限）"
          />
        </el-form-item>
        <el-form-item label="期望到期日" prop="expiry_date">
          <el-date-picker
            v-model="requestForm.expiry_date"
            type="datetime"
            placeholder="选择到期日期时间"
            style="width: 100%"
            :disabled-date="disabledDate"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="requestDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitRequest" :loading="requesting">提交申请</el-button>
      </template>
    </el-dialog>

    <!-- 授予权限对话框（管理员使用） -->
    <el-dialog v-model="grantDialogVisible" title="授予临时权限" width="500px">
      <el-form :model="grantForm" :rules="grantRules" ref="grantFormRef" label-width="100px">
        <el-form-item label="选择用户" prop="user_ids">
          <el-select-v2
            v-model="grantForm.user_ids"
            :options="userOptions"
            placeholder="搜索并选择用户（可多选）"
            style="width: 100%"
            multiple
            filterable
            clearable
            :filter-method="filterUsers"
          />
          <div class="form-hint">支持按姓名或学号搜索</div>
        </el-form-item>
        
        <el-form-item label="权限类型" prop="permission">
          <el-select v-model="grantForm.permission" placeholder="选择权限" style="width: 100%">
            <el-option 
              v-for="perm in grantablePermissions" 
              :key="perm.key" 
              :label="perm.name" 
              :value="perm.key" 
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="到期日期" prop="expires_at">
          <el-date-picker
            v-model="grantForm.expires_at"
            type="datetime"
            placeholder="选择到期日期时间"
            style="width: 100%"
            :disabled-date="disabledDate"
          />
        </el-form-item>
        
        <el-form-item label="授权原因" prop="reason">
          <el-input 
            v-model="grantForm.reason" 
            type="textarea" 
            rows="2"
            placeholder="可选：填写授权原因"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="grantDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitGrant" :loading="granting">确定授权</el-button>
      </template>
    </el-dialog>

    <!-- 申请详情对话框 -->
    <el-dialog v-model="detailDialogVisible" title="申请详情" width="500px">
      <el-descriptions :column="1" border v-if="selectedApp">
        <el-descriptions-item label="申请编号">{{ selectedApp.application_no }}</el-descriptions-item>
        <el-descriptions-item label="申请类型">{{ getAppTypeName(selectedApp.type_code) }}</el-descriptions-item>
        <el-descriptions-item label="申请人">{{ selectedApp.applicant_name }}</el-descriptions-item>
        <el-descriptions-item label="申请时间">{{ formatTime(selectedApp.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="当前状态">
          <el-tag :type="getStatusType(selectedApp.status)">{{ getStatusText(selectedApp.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="申请内容" v-if="selectedApp.data">
          <div v-if="parseData(selectedApp.data).permission">
            <div>权限: {{ getPermissionText(parseData(selectedApp.data).permission) }}</div>
          </div>
          <div v-if="parseData(selectedApp.data).expiry_date">
            期望到期: {{ formatDate(parseData(selectedApp.data).expiry_date) }}
          </div>
        </el-descriptions-item>
        <el-descriptions-item label="申请原因">{{ selectedApp.content || selectedApp.reason || '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <!-- 审批对话框 -->
    <el-dialog v-model="approvalDialogVisible" title="审批申请" width="450px">
      <el-alert
        :type="approvalAction === 'approve' ? 'success' : 'error'"
        :closable="false"
        :title="approvalAction === 'approve' ? '通过申请' : '拒绝申请'"
        style="margin-bottom: 16px"
      />
      <el-form :model="approvalForm" ref="approvalFormRef" label-width="80px">
        <el-form-item label="审批意见">
          <el-input 
            v-model="approvalForm.comment" 
            type="textarea" 
            rows="3"
            :placeholder="approvalAction === 'approve' ? '可选：填写审批意见' : '请填写拒绝原因'"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="approvalDialogVisible = false">取消</el-button>
        <el-button 
          :type="approvalAction === 'approve' ? 'success' : 'danger'" 
          @click="submitApproval" 
          :loading="approving"
        >
          {{ approvalAction === 'approve' ? '确认通过' : '确认拒绝' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '../stores/user'
import { getTempPermissions, getMyTempPermissions, grantTempPermission, revokeTempPermission } from '../api/system'
import { getUserList, getUserListByDepartment } from '../api/user'
import { 
  getMyApplications, 
  getPendingApprovals, 
  createApplication, 
  processApproval,
  cancelApplication,
  getApplicationStats,
  getAvailablePermissions
} from '../api/application'

const userStore = useUserStore()

// 标签页
const activeTab = ref('permissions')

// 加载状态
const loading = ref(false)
const loadingApps = ref(false)
const loadingPending = ref(false)

// 权限列表
const permissionList = ref([])
const userOptions = ref([])
const allUsers = ref([])

// 申请列表
const myApplications = ref([])
const pendingList = ref([])
const appPage = ref(1)
const appPageSize = ref(10)
const appTotal = ref(0)
const pendingPage = ref(1)
const pendingPageSize = ref(10)
const pendingTotal = ref(0)

// 统计数据
const stats = ref(null)

// 可申请的权限
const availablePermissions = ref([])

// 对话框状态
const requestDialogVisible = ref(false)
const grantDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const approvalDialogVisible = ref(false)

// 加载状态
const requesting = ref(false)
const granting = ref(false)
const approving = ref(false)

// 表单引用
const requestFormRef = ref()
const grantFormRef = ref()
const approvalFormRef = ref()

// 当前选中的申请
const selectedApp = ref(null)
const approvalAction = ref('approve')

// 权限判断
const canGrant = computed(() => userStore.canManageDept || userStore.canManageAll)
const canApprove = computed(() => userStore.canManageDept || userStore.canManageAll)

// 表单数据
const requestForm = reactive({
  permission: '',
  reason: '',
  expiry_date: null
})

const grantForm = reactive({
  user_ids: [],
  permission: '',
  expires_at: null,
  reason: ''
})

const approvalForm = reactive({
  comment: ''
})

// 表单规则
const requestRules = {
  permission: [{ required: true, message: '请选择申请权限' }],
  reason: [{ required: true, message: '请填写申请原因' }],
  expiry_date: [{ required: true, message: '请选择期望到期日' }]
}

const grantRules = {
  user_ids: [{ required: true, message: '请至少选择一个用户', type: 'array', min: 1 }],
  permission: [{ required: true, message: '请选择权限类型' }],
  expires_at: [{ required: true, message: '请选择到期日期' }]
}

// 权限映射
const permissionMap = {
  'duty_manage': '值班管理',
  'user_manage': '用户管理',
  'schedule_manage': '排班管理',
  'crawler_manage': '爬虫管理',
  'system_manage': '系统管理',
  'temp_permission_grant': '授权管理',
  'schedule:manage:dept': '部门排班管理',
  'user:manage:dept': '部门用户管理',
  'schedule:view:all': '全局排班查看',
  'user:manage:all': '全局用户管理',
  'schedule:publish': '设置每周分工'
}

const appTypeMap = {
  'temp_permission': '权限申请',
  'leave': '请假申请'
}

const statusMap = {
  0: { text: '待审批', type: 'warning' },
  1: { text: '审批中', type: 'primary' },
  2: { text: '已通过', type: 'success' },
  3: { text: '已拒绝', type: 'danger' },
  4: { text: '已撤回', type: 'info' }
}

// 管理员可授予的权限
const grantablePermissions = computed(() => {
  const perms = [
    { key: 'duty_manage', name: '值班管理' },
    { key: 'user_manage', name: '用户管理' },
    { key: 'schedule_manage', name: '排班管理' },
    { key: 'crawler_manage', name: '爬虫管理' },
    { key: 'temp_permission_grant', name: '授权管理' }
  ]
  
  if (userStore.canManageAll) {
    perms.push({ key: 'system_manage', name: '系统管理' })
  }
  
  return perms
})

// 方法
const getPermissionText = (perm) => permissionMap[perm] || perm
const getAppTypeName = (type) => appTypeMap[type] || type
const getStatusText = (status) => statusMap[status]?.text || '未知'
const getStatusType = (status) => statusMap[status]?.type || 'info'

const isExpired = (time) => new Date(time) < new Date()

const formatTime = (time) => {
  if (!time) return '-'
  const date = new Date(time)
  return date.toLocaleString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

const formatDate = (time) => {
  if (!time) return '-'
  const date = new Date(time)
  return date.toLocaleDateString('zh-CN')
}

const parseData = (data) => {
  if (!data) return {}
  if (typeof data === 'string') {
    try {
      return JSON.parse(data)
    } catch {
      return {}
    }
  }
  return data
}

const disabledDate = (time) => {
  return time.getTime() < Date.now() - 8.64e7
}

const getPermissionType = (perm) => {
  if (perm.includes('system') || perm.includes('all')) return 'danger'
  if (perm.includes('manage')) return 'warning'
  return 'info'
}

// 加载数据
const loadPermissions = async () => {
  loading.value = true
  try {
    // 管理员加载所有权限，普通用户只加载自己的权限
    if (canGrant.value) {
      const data = await getTempPermissions()
      permissionList.value = data || []
    } else {
      // 普通用户调用 /temp-permissions/my 接口
      const data = await getMyTempPermissions()
      permissionList.value = data || []
    }
  } finally {
    loading.value = false
  }
}

const loadUsers = async () => {
  try {
    let data
    if (userStore.canManageAll) {
      // 系统管理员获取所有用户
      data = await getUserList()
    } else if (userStore.canManageDept) {
      // 部门管理员获取本部门用户
      data = await getUserListByDepartment(userStore.user?.department)
    } else {
      // 普通用户不需要加载用户列表
      return
    }
    allUsers.value = data || []
    userOptions.value = data.map(u => ({ value: u.id, label: `${u.name} (${u.student_id})` }))
  } catch {}
}

const loadMyApplications = async () => {
  loadingApps.value = true
  try {
    const res = await getMyApplications({
      page: appPage.value,
      page_size: appPageSize.value
    })
    // res 直接是 data: { list, total }
    myApplications.value = res?.list || []
    appTotal.value = res?.total || 0
  } catch (err) {
    console.error('加载申请列表失败:', err)
  } finally {
    loadingApps.value = false
  }
}

const loadPendingApprovals = async () => {
  loadingPending.value = true
  try {
    const res = await getPendingApprovals({
      page: pendingPage.value,
      page_size: pendingPageSize.value
    })
    // res 直接是 data: { list, total }
    pendingList.value = res?.list || []
    pendingTotal.value = res?.total || 0
  } catch (err) {
    console.error('加载待审批列表失败:', err)
  } finally {
    loadingPending.value = false
  }
}

const loadStats = async () => {
  try {
    const res = await getApplicationStats()
    // res 直接是 data
    stats.value = res || null
  } catch (err) {
    console.error('加载统计失败:', err)
  }
}

const loadAvailablePermissions = async () => {
  try {
    const res = await getAvailablePermissions()
    // request.js 拦截器已提取 data，res 直接是数据
    availablePermissions.value = res || []
  } catch (err) {
    console.error('加载可用权限失败:', err)
    availablePermissions.value = []
  }
}

// 搜索用户
const filterUsers = (query) => {
  if (!query) {
    userOptions.value = allUsers.value.map(u => ({ value: u.id, label: `${u.name} (${u.student_id})` }))
    return
  }
  const lowerQuery = query.toLowerCase()
  userOptions.value = allUsers.value
    .filter(u => 
      u.name.toLowerCase().includes(lowerQuery) || 
      u.student_id.toLowerCase().includes(lowerQuery)
    )
    .map(u => ({ value: u.id, label: `${u.name} (${u.student_id})` }))
}

// 打开对话框
const openRequestDialog = () => {
  requestDialogVisible.value = true
  requestForm.permission = ''
  requestForm.reason = ''
  requestForm.expiry_date = null
  loadAvailablePermissions()
}

const openGrantDialog = () => {
  grantDialogVisible.value = true
  grantForm.user_ids = []
  grantForm.permission = ''
  grantForm.expires_at = null
  grantForm.reason = ''
}

const viewDetail = (row) => {
  selectedApp.value = row
  detailDialogVisible.value = true
}

const approveApp = (row) => {
  selectedApp.value = row
  approvalAction.value = 'approve'
  approvalForm.comment = ''
  approvalDialogVisible.value = true
}

const rejectApp = (row) => {
  selectedApp.value = row
  approvalAction.value = 'reject'
  approvalForm.comment = ''
  approvalDialogVisible.value = true
}

// 提交操作
const submitRequest = async () => {
  const valid = await requestFormRef.value.validate().catch(() => false)
  if (!valid) return

  requesting.value = true
  try {
    const data = {
      type: 'temp_permission',
      data: {
        permission: requestForm.permission,
        expiry_date: requestForm.expiry_date,
        reason: requestForm.reason
      },
      reason: requestForm.reason
    }
    
    await createApplication(data)
    ElMessage.success('申请已提交，请等待管理员审批')
    requestDialogVisible.value = false
    loadMyApplications()
    loadStats()
  } catch (err) {
    ElMessage.error(err.message || '申请提交失败')
  } finally {
    requesting.value = false
  }
}

const submitGrant = async () => {
  const valid = await grantFormRef.value.validate().catch(() => false)
  if (!valid) return

  granting.value = true
  try {
    const expiresAt = new Date(grantForm.expires_at)
    
    await grantTempPermission({
      user_ids: grantForm.user_ids,
      permission: grantForm.permission,
      resource_type: 'all',
      resource_id: 0,
      expires_at: expiresAt.toISOString(),
      reason: grantForm.reason
    })
    ElMessage.success(`成功授权 ${grantForm.user_ids.length} 位用户`)
    grantDialogVisible.value = false
    loadPermissions()
  } catch (err) {
    ElMessage.error(err.message || '授权失败')
  } finally {
    granting.value = false
  }
}

const submitApproval = async () => {
  if (approvalAction.value === 'reject' && !approvalForm.comment) {
    ElMessage.warning('请填写拒绝原因')
    return
  }

  approving.value = true
  try {
    await processApproval(selectedApp.value.id, {
      action: approvalAction.value,
      comment: approvalForm.comment
    })
    ElMessage.success(approvalAction.value === 'approve' ? '已通过申请' : '已拒绝申请')
    approvalDialogVisible.value = false
    loadPendingApprovals()
    loadStats()
  } catch (err) {
    ElMessage.error(err.message || '审批处理失败')
  } finally {
    approving.value = false
  }
}

const cancelApp = async (row) => {
  try {
    await ElMessageBox.confirm('确定撤销此申请吗？', '确认', { type: 'warning' })
    await cancelApplication(row.id)
    ElMessage.success('申请已撤销')
    loadMyApplications()
    loadStats()
  } catch {}
}

const revokePermission = async (row) => {
  try {
    await ElMessageBox.confirm(`确定撤销 ${row.user_name} 的权限吗？`, '确认', { type: 'warning' })
    await revokeTempPermission(row.id)
    ElMessage.success('撤销成功')
    loadPermissions()
  } catch {}
}

onMounted(() => {
  // 所有用户都加载自己的权限列表
  loadPermissions()
  
  // 管理员额外加载用户列表（用于授权）
  if (canGrant.value) {
    loadUsers()
  }
  
  // 所有用户都加载申请相关数据
  loadMyApplications()
  if (canApprove.value) {
    loadPendingApprovals()
  }
  loadStats()
})
</script>

<style scoped>
.stats-cards {
  display: flex;
  gap: 16px;
  margin-bottom: 20px;
}

.stat-card {
  flex: 1;
  background: white;
  border-radius: 8px;
  padding: 20px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s;
  border: 2px solid transparent;
}

.stat-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.stat-card.active {
  border-color: #409eff;
  background: #f5f7fa;
}

.stat-value {
  font-size: 32px;
  font-weight: bold;
  color: #409eff;
}

.stat-value.stat-warning {
  color: #e6a23c;
}

.stat-label {
  margin-top: 8px;
  color: #606266;
  font-size: 14px;
}

.action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.tab-badge :deep(.el-badge__content) {
  position: relative;
  top: -2px;
  margin-left: 4px;
}

.text-danger {
  color: #f56c6c;
}

.text-gray {
  color: #909399;
  font-size: 13px;
}

.form-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}
</style>
