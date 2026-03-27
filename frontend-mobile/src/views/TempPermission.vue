<template>
  <div class="main-container">
    <Navbar title="临时权限管理" show-back />
    <!-- 统计卡片 -->
    <div class="stats-cards" v-if="stats">
      <div class="stat-card" @click="activeTab = 'myApplications'" :class="{ active: activeTab === 'myApplications' }">
        <div class="stat-value">{{ stats.my_applications?.pending || 0 }}</div>
        <div class="stat-label">待审批</div>
      </div>
      <div class="stat-card" @click="activeTab = 'pendingApprovals'" :class="{ active: activeTab === 'pendingApprovals' }" v-if="canApprove">
        <div class="stat-value stat-warning">{{ stats.pending_approval || 0 }}</div>
        <div class="stat-label">待我审批</div>
      </div>
      <div class="stat-card">
        <div class="stat-value stat-success">{{ stats.my_applications?.approved || 0 }}</div>
        <div class="stat-label">已通过</div>
      </div>
      <div class="stat-card">
        <div class="stat-value stat-danger">{{ stats.my_applications?.rejected || 0 }}</div>
        <div class="stat-label">已拒绝</div>
      </div>
    </div>

    <!-- Tab 切换 -->
    <van-tabs v-model:active="activeTab" sticky swipeable animated>
      <!-- 权限管理 Tab -->
      <van-tab title="权限管理" name="permissions">
        <div class="tab-content">
          <div class="action-bar" v-if="canGrant">
            <van-button type="primary" size="small" icon="plus" @click="openGrantDialog">授予权限</van-button>
          </div>
          <div class="action-bar" v-else>
            <van-button type="primary" size="small" icon="plus" @click="openRequestDialog">申请权限</van-button>
          </div>

          <div v-if="loading" class="loading-wrapper">
            <van-loading type="spinner" color="#1989fa" />
          </div>

          <div v-else-if="permissionList.length === 0" class="empty-wrapper">
            <van-empty description="暂无权限数据" />
          </div>
          <div v-else class="card-list">
            <div class="perm-card" v-for="item in permissionList" :key="item.id">
              <div class="perm-card-header">
                <span class="perm-card-name" v-if="canGrant">{{ item.user_name }}</span>
                <van-tag v-if="canGrant && item.user_department" size="medium" class="perm-card-dept">{{ item.user_department }}</van-tag>
                <van-tag size="medium" :type="getPermissionType(item.permission)">{{ getPermissionText(item.permission) }}</van-tag>
                <van-tag size="medium" :type="isExpired(item.expires_at) ? 'default' : 'success'" plain>{{ isExpired(item.expires_at) ? '已过期' : '有效' }}</van-tag>
              </div>
              <div class="perm-card-body">
                <div class="perm-card-row" v-if="item.granted_by_name">
                  <span class="perm-card-label">授权人</span>
                  <span>{{ item.granted_by_name }}</span>
                </div>
                <div class="perm-card-row">
                  <span class="perm-card-label">有效期</span>
                  <span :class="{ 'text-danger': isExpired(item.expires_at) }">{{ formatTime(item.expires_at) }}</span>
                </div>
              </div>
              <div class="perm-card-actions" v-if="canGrant">
                <van-button size="mini" type="danger" plain @click="revokePermission(item)">撤销</van-button>
              </div>
            </div>
          </div>
        </div>
      </van-tab>

      <!-- 我的申请 Tab -->
      <van-tab title="我的申请" name="myApplications">
        <div class="tab-content">
          <div v-if="loadingApps" class="loading-wrapper">
            <van-loading type="spinner" color="#1989fa" />
          </div>

          <div v-else-if="myApplications.length === 0" class="empty-wrapper">
            <van-empty description="暂无申请记录" />
          </div>

          <div v-else class="card-list">
            <div class="app-card" v-for="item in myApplications" :key="item.id">
              <div class="app-card-header">
                <span class="app-card-no">{{ item.application_no }}</span>
                <van-tag size="medium" :type="getStatusType(item.status)">{{ getStatusText(item.status) }}</van-tag>
              </div>
              <div class="app-card-body">
                <div class="app-card-row">
                  <span class="app-card-label">申请类型</span>
                  <van-tag size="medium" plain>{{ getAppTypeName(item.type_code) }}</van-tag>
                </div>
                <div class="app-card-row" v-if="parseData(item.data).permission">
                  <span class="app-card-label">权限</span>
                  <span>{{ getPermissionText(parseData(item.data).permission) }}</span>
                </div>
                <div class="app-card-row text-gray" v-if="item.content || item.reason">
                  <span>{{ item.content || item.reason }}</span>
                </div>
                <div class="app-card-row">
                  <span class="app-card-label">申请时间</span>
                  <span class="text-gray">{{ formatTime(item.created_at) }}</span>
                </div>
              </div>
              <div class="app-card-actions">
                <van-button size="mini" type="primary" plain @click="viewDetail(item)">详情</van-button>
                <van-button size="mini" type="danger" plain @click="cancelApp(item)" v-if="item.status === 0">撤销</van-button>
              </div>
            </div>
          </div>

          <div class="load-more" v-if="myApplications.length > 0 && appPage * appPageSize < appTotal">
            <van-button size="small" plain block :loading="loadingApps" @click="loadMoreApps">加载更多</van-button>
          </div>
        </div>
      </van-tab>

      <!-- 待我审批 Tab -->
      <van-tab name="pendingApprovals" v-if="canApprove">
        <template #title>
          待我审批
          <van-badge v-if="stats?.pending_approval" :content="stats.pending_approval" class="tab-badge" />
        </template>
        <div class="tab-content">
          <div v-if="loadingPending" class="loading-wrapper">
            <van-loading type="spinner" color="#1989fa" />
          </div>

          <div v-else-if="pendingList.length === 0" class="empty-wrapper">
            <van-empty description="暂无待审批申请" />
          </div>

          <div v-else class="card-list">
            <div class="pending-card" v-for="item in pendingList" :key="item.id">
              <div class="pending-card-header">
                <div class="pending-card-user">
                  <span class="pending-card-name">{{ item.applicant_name }}</span>
                  <van-tag v-if="item.department" size="medium">{{ item.department }}</van-tag>
                </div>
                <van-tag size="medium" plain>{{ getAppTypeName(item.type_code) }}</van-tag>
              </div>
              <div class="pending-card-body">
                <div class="pending-card-row">
                  <span class="pending-card-label">申请编号</span>
                  <span>{{ item.application_no }}</span>
                </div>
                <div class="pending-card-row" v-if="parseData(item.data).permission">
                  <span class="pending-card-label">权限</span>
                  <span>{{ getPermissionText(parseData(item.data).permission) }}</span>
                </div>
                <div class="pending-card-row" v-if="parseData(item.data).expiry_date">
                  <span class="pending-card-label">期望到期</span>
                  <span>{{ formatDate(parseData(item.data).expiry_date) }}</span>
                </div>
                <div class="pending-card-row text-gray" v-if="item.content">
                  <span>{{ item.content }}</span>
                </div>
                <div class="pending-card-row">
                  <span class="pending-card-label">申请时间</span>
                  <span class="text-gray">{{ formatTime(item.created_at) }}</span>
                </div>
              </div>
              <div class="pending-card-actions">
                <van-button size="small" type="success" @click="approveApp(item)">通过</van-button>
                <van-button size="small" type="danger" plain @click="rejectApp(item)">拒绝</van-button>
              </div>
            </div>
          </div>

          <div class="load-more" v-if="pendingList.length > 0 && pendingPage * pendingPageSize < pendingTotal">
            <van-button size="small" plain block :loading="loadingPending" @click="loadMorePending">加载更多</van-button>
          </div>
        </div>
      </van-tab>
    </van-tabs>

    <!-- 权限申请弹窗 -->
    <van-popup v-model:show="requestDialogVisible" round position="bottom" :style="{ maxHeight: '90%' }">
      <div class="popup-content">
        <div class="popup-header">
          <span class="popup-title">申请临时权限</span>
          <van-icon name="cross" size="20" @click="requestDialogVisible = false" />
        </div>

        <van-notice-bar
          left-icon="info-o"
          :text="requestForm.permission && isGlobalPerm(requestForm.permission) ? '全局权限申请将发送给办公室管理员或系统管理员审批' : '请填写权限申请信息，提交后将发送给您的部门管理员审批'"
          :scrollable="false"
          wrapable
          class="mb-4"
        />

        <van-form ref="requestFormRef" @submit="submitRequest">
          <van-cell-group inset>
            <van-field
              v-model="selectedPermText"
              is-link
              readonly
              label="申请权限"
              placeholder="选择需要的权限"
              :rules="[{ required: true, message: '请选择申请权限' }]"
              @click="showPermPicker = true"
            />
            <van-field
              v-model="requestForm.reason"
              label="申请原因"
              type="textarea"
              placeholder="请详细说明申请原因"
              rows="3"
              autosize
              show-word-limit
              maxlength="200"
              :rules="[{ required: true, message: '请填写申请原因' }]"
            />
            <van-field
              v-model="requestForm.expiry_date"
              is-link
              readonly
              label="期望到期日"
              placeholder="选择到期日期时间"
              :rules="[{ required: true, message: '请选择期望到期日' }]"
              @click="showExpiryDatePicker = true"
            />
          </van-cell-group>

          <div class="popup-actions">
            <van-button plain @click="requestDialogVisible = false">取消</van-button>
            <van-button type="primary" native-type="submit" :loading="requesting" loading-text="提交中...">提交申请</van-button>
          </div>
        </van-form>
      </div>
    </van-popup>

    <!-- 权限选择 Picker -->
    <van-popup v-model:show="showPermPicker" round position="bottom">
      <van-picker
        :columns="availablePermColumns"
        @confirm="onPermConfirm"
        @cancel="showPermPicker = false"
      />
    </van-popup>

    <!-- 到期日选择 -->
    <van-popup v-model:show="showExpiryDatePicker" round position="bottom">
      <van-date-picker
        v-model="expiryDateArr"
        title="选择到期日期时间"
        :min-date="minDate"
        @confirm="onExpiryDateConfirm"
        @cancel="showExpiryDatePicker = false"
      />
    </van-popup>

    <!-- 授予权限弹窗 -->
    <van-popup v-model:show="grantDialogVisible" round position="bottom" :style="{ maxHeight: '90%' }">
      <div class="popup-content">
        <div class="popup-header">
          <span class="popup-title">授予临时权限</span>
          <van-icon name="cross" size="20" @click="grantDialogVisible = false" />
        </div>

        <van-form ref="grantFormRef" @submit="submitGrant">
          <van-cell-group inset>
            <van-field
              v-model="grantUserText"
              is-link
              readonly
              label="选择用户"
              placeholder="搜索并选择用户（可多选）"
              :rules="[{ required: true, message: '请至少选择一个用户', validator: () => grantForm.user_ids.length > 0 }]"
              @click="showUserPicker = true"
            />
            <van-field
              v-model="selectedGrantPermText"
              is-link
              readonly
              label="权限类型"
              placeholder="选择权限"
              :rules="[{ required: true, message: '请选择权限类型' }]"
              @click="showGrantPermPicker = true"
            />
            <van-field
              v-model="grantExpiryText"
              is-link
              readonly
              label="到期日期"
              placeholder="选择到期日期时间"
              :rules="[{ required: true, message: '请选择到期日期' }]"
              @click="showGrantExpiryPicker = true"
            />
            <van-field
              v-model="grantForm.reason"
              label="授权原因"
              type="textarea"
              placeholder="可选：填写授权原因"
              rows="2"
              autosize
              show-word-limit
              maxlength="200"
            />
          </van-cell-group>

          <div class="popup-actions">
            <van-button plain @click="grantDialogVisible = false">取消</van-button>
            <van-button type="primary" native-type="submit" :loading="granting" loading-text="授权中...">确定授权</van-button>
          </div>
        </van-form>
      </div>
    </van-popup>

    <!-- 用户选择弹窗（多选） -->
    <van-popup v-model:show="showUserPicker" round position="bottom" :style="{ maxHeight: '60%' }">
      <div class="user-picker-popup">
        <div class="popup-header">
          <span class="popup-title">选择用户</span>
          <van-icon name="cross" size="20" @click="showUserPicker = false" />
        </div>
        <van-search v-model="userSearchQuery" placeholder="按姓名或学号搜索" shape="round" />
        <div class="user-list">
          <van-checkbox-group v-model="grantForm.user_ids">
            <van-cell-group>
              <van-cell
                v-for="u in filteredUserOptions"
                :key="u.value"
                :title="u.label"
                clickable
                @click="toggleUser(u.value)"
              >
                <template #right-icon>
                  <van-checkbox :name="u.value" />
                </template>
              </van-cell>
            </van-cell-group>
          </van-checkbox-group>
          <van-empty v-if="filteredUserOptions.length === 0" description="未找到用户" />
        </div>
        <div class="user-picker-footer">
          <van-button type="primary" block round @click="confirmUserSelect">
            确认选择 ({{ grantForm.user_ids.length }})
          </van-button>
        </div>
      </div>
    </van-popup>

    <!-- 授权权限选择 Picker -->
    <van-popup v-model:show="showGrantPermPicker" round position="bottom">
      <van-picker
        :columns="grantablePermColumns"
        @confirm="onGrantPermConfirm"
        @cancel="showGrantPermPicker = false"
      />
    </van-popup>

    <!-- 授权到期日期选择 -->
    <van-popup v-model:show="showGrantExpiryPicker" round position="bottom">
      <van-date-picker
        v-model="grantExpiryDateArr"
        title="选择到期日期"
        :min-date="minDate"
        @confirm="onGrantExpiryConfirm"
        @cancel="showGrantExpiryPicker = false"
      />
    </van-popup>

    <!-- 申请详情弹窗 -->
    <van-popup v-model:show="detailDialogVisible" round position="bottom" :style="{ maxHeight: '80%' }">
      <div class="popup-content" v-if="selectedApp">
        <div class="popup-header">
          <span class="popup-title">申请详情</span>
          <van-icon name="cross" size="20" @click="detailDialogVisible = false" />
        </div>
        <van-cell-group inset>
          <van-cell title="申请编号" :value="selectedApp.application_no" />
          <van-cell title="申请类型" :value="getAppTypeName(selectedApp.type_code)" />
          <van-cell title="申请人" :value="selectedApp.applicant_name" />
          <van-cell title="申请时间" :value="formatTime(selectedApp.created_at)" />
          <van-cell title="当前状态">
            <template #value>
              <van-tag :type="getStatusType(selectedApp.status)">{{ getStatusText(selectedApp.status) }}</van-tag>
            </template>
          </van-cell>
          <van-cell title="申请内容" v-if="selectedApp.data">
            <template #value>
              <div class="detail-content">
                <div v-if="parseData(selectedApp.data).permission">
                  权限: {{ getPermissionText(parseData(selectedApp.data).permission) }}
                </div>
                <div v-if="parseData(selectedApp.data).expiry_date">
                  期望到期: {{ formatDate(parseData(selectedApp.data).expiry_date) }}
                </div>
              </div>
            </template>
          </van-cell>
          <van-cell title="申请原因" :value="selectedApp.content || selectedApp.reason || '-'" />
        </van-cell-group>
      </div>
    </van-popup>

    <!-- 审批弹窗 -->
    <van-popup v-model:show="approvalDialogVisible" round position="bottom">
      <div class="popup-content">
        <div class="popup-header">
          <span class="popup-title">{{ approvalAction === 'approve' ? '通过申请' : '拒绝申请' }}</span>
          <van-icon name="cross" size="20" @click="approvalDialogVisible = false" />
        </div>

        <van-notice-bar
          v-if="approvalAction === 'approve'"
          left-icon="passed"
          color="#07c160"
          background="#e8f7ef"
          text="确认通过该申请"
          :scrollable="false"
        />
        <van-notice-bar
          v-else
          left-icon="warning-o"
          color="#ee0a24"
          background="#ffeef0"
          text="确认拒绝该申请，请填写拒绝原因"
          :scrollable="false"
        />

        <van-form ref="approvalFormRef" @submit="submitApproval">
          <van-cell-group inset>
            <van-field
              v-model="approvalForm.comment"
              label="审批意见"
              type="textarea"
              rows="3"
              autosize
              show-word-limit
              maxlength="200"
              :placeholder="approvalAction === 'approve' ? '可选：填写审批意见' : '请填写拒绝原因'"
              :rules="approvalAction === 'reject' ? [{ required: true, message: '请填写拒绝原因' }] : []"
            />
          </van-cell-group>

          <div class="popup-actions">
            <van-button plain @click="approvalDialogVisible = false">取消</van-button>
            <van-button
              :type="approvalAction === 'approve' ? 'success' : 'danger'"
              native-type="submit"
              :loading="approving"
              loading-text="处理中..."
            >
              {{ approvalAction === 'approve' ? '确认通过' : '确认拒绝' }}
            </van-button>
          </div>
        </van-form>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { showToast, showConfirmDialog } from 'vant'
import Navbar from '../components/Navbar.vue'
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

const activeTab = ref('permissions')
const loading = ref(false)
const loadingApps = ref(false)
const loadingPending = ref(false)

const permissionList = ref([])
const userOptions = ref([])
const allUsers = ref([])

const myApplications = ref([])
const pendingList = ref([])
const appPage = ref(1)
const appPageSize = ref(10)
const appTotal = ref(0)
const pendingPage = ref(1)
const pendingPageSize = ref(10)
const pendingTotal = ref(0)

const stats = ref(null)
const availablePermissions = ref([])

const requestDialogVisible = ref(false)
const grantDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const approvalDialogVisible = ref(false)

const showPermPicker = ref(false)
const showExpiryDatePicker = ref(false)
const showGrantPermPicker = ref(false)
const showGrantExpiryPicker = ref(false)
const showUserPicker = ref(false)
const userSearchQuery = ref('')

const requesting = ref(false)
const granting = ref(false)
const approving = ref(false)

const requestFormRef = ref()
const grantFormRef = ref()
const approvalFormRef = ref()

const selectedApp = ref(null)
const approvalAction = ref('approve')

const canGrant = computed(() => userStore.isAdmin || userStore.isOfficeAdmin || userStore.isDeptAdmin)
const canApprove = computed(() => userStore.isAdmin || userStore.isOfficeAdmin || userStore.isDeptAdmin)

const minDate = new Date()

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

const expiryDateArr = ref([])
const grantExpiryDateArr = ref([])

const selectedPermText = ref('')
const selectedGrantPermText = ref('')
const grantUserText = ref('')
const grantExpiryText = ref('')

const permissionMap = {
  'schedule:publish': '设置每周分工',
  'schedule:manage:all': '排班管理（全部）',
  'user:manage:all': '用户管理（全部）',
  'schedule:manage:dept': '排班管理（部门）',
  'user:manage:dept': '用户管理（部门）'
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
  4: { text: '已撤回', type: 'default' }
}

const grantablePermissions = computed(() => {
  const perms = [
    { key: 'schedule:publish', name: '设置每周分工' },
    { key: 'schedule:manage:dept', name: '排班管理（部门）' },
    { key: 'user:manage:dept', name: '用户管理（部门）' }
  ]
  if (userStore.canManageAll) {
    perms.unshift(
      { key: 'schedule:manage:all', name: '排班管理（全部）' },
      { key: 'user:manage:all', name: '用户管理（全部）' }
    )
  }
  return perms
})

const availablePermColumns = computed(() =>
  (availablePermissions.value || []).map(p => ({ text: p.name, value: p.key }))
)

const grantablePermColumns = computed(() =>
  grantablePermissions.value.map(p => ({ text: p.name, value: p.key }))
)

const filteredUserOptions = computed(() => {
  if (!userSearchQuery.value) return userOptions.value
  const q = userSearchQuery.value.toLowerCase()
  return userOptions.value.filter(u => u.label.toLowerCase().includes(q))
})

const getPermissionText = (perm) => permissionMap[perm] || perm
const getAppTypeName = (type) => appTypeMap[type] || type
const getStatusText = (status) => statusMap[status]?.text || '未知'
const getStatusType = (status) => statusMap[status]?.type || 'default'
const isExpired = (time) => new Date(time) < new Date()

const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

const formatDate = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleDateString('zh-CN')
}

const parseData = (data) => {
  if (!data) return {}
  if (typeof data === 'string') {
    try { return JSON.parse(data) } catch { return {} }
  }
  return data
}

const isGlobalPerm = (perm) => perm === 'schedule:manage:all' || perm === 'user:manage:all'

const getPermissionType = (perm) => {
  if (perm.includes('system') || perm.includes('all')) return 'danger'
  if (perm.includes('manage')) return 'warning'
  return 'primary'
}

const onPermConfirm = ({ selectedValues }) => {
  requestForm.permission = selectedValues[0]
  selectedPermText.value = selectedValues[0] ? getPermissionText(selectedValues[0]) : ''
  showPermPicker.value = false
}

const onGrantPermConfirm = ({ selectedValues }) => {
  grantForm.permission = selectedValues[0]
  selectedGrantPermText.value = selectedValues[0] ? getPermissionText(selectedValues[0]) : ''
  showGrantPermPicker.value = false
}

const onExpiryDateConfirm = ({ selectedValues }) => {
  const dateStr = selectedValues.join('-')
  requestForm.expiry_date = new Date(dateStr).toISOString()
  showExpiryDatePicker.value = false
}

const onGrantExpiryConfirm = ({ selectedValues }) => {
  const dateStr = selectedValues.join('-')
  grantForm.expires_at = new Date(dateStr).toISOString()
  grantExpiryText.value = dateStr
  showGrantExpiryPicker.value = false
}

const toggleUser = (id) => {
  const idx = grantForm.user_ids.indexOf(id)
  if (idx > -1) {
    grantForm.user_ids.splice(idx, 1)
  } else {
    grantForm.user_ids.push(id)
  }
}

const confirmUserSelect = () => {
  if (grantForm.user_ids.length === 0) {
    showToast({ message: '请至少选择一个用户', type: 'fail' })
    return
  }
  const names = userOptions.value
    .filter(u => grantForm.user_ids.includes(u.value))
    .map(u => u.label.split(' (')[0])
  grantUserText.value = names.join('、')
  showUserPicker.value = false
}

const loadPermissions = async () => {
  loading.value = true
  try {
    if (canGrant.value) {
      const data = await getTempPermissions()
      permissionList.value = data || []
    } else {
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
      data = await getUserList()
    } else if (userStore.canManageDept) {
      data = await getUserListByDepartment(userStore.user?.department)
    } else {
      return
    }
    allUsers.value = data || []
    userOptions.value = data.map(u => ({ value: u.id, label: `${u.name} (${u.student_id})` }))
  } catch {}
}

const loadMyApplications = async () => {
  loadingApps.value = true
  try {
    const res = await getMyApplications({ page: appPage.value, page_size: appPageSize.value })
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
    const res = await getPendingApprovals({ page: pendingPage.value, page_size: pendingPageSize.value })
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
    stats.value = res || null
  } catch (err) {
    console.error('加载统计失败:', err)
  }
}

const loadAvailablePermissions = async () => {
  try {
    const res = await getAvailablePermissions()
    availablePermissions.value = res || []
  } catch (err) {
    console.error('加载可用权限失败:', err)
    availablePermissions.value = []
  }
}

const loadMoreApps = () => { appPage.value++; loadMyApplications() }
const loadMorePending = () => { pendingPage.value++; loadPendingApprovals() }

const openRequestDialog = () => {
  requestDialogVisible.value = true
  requestForm.permission = ''
  requestForm.reason = ''
  requestForm.expiry_date = null
  selectedPermText.value = ''
  expiryDateArr.value = []
  loadAvailablePermissions()
}

const openGrantDialog = () => {
  grantDialogVisible.value = true
  grantForm.user_ids = []
  grantForm.permission = ''
  grantForm.expires_at = null
  grantForm.reason = ''
  selectedGrantPermText.value = ''
  grantExpiryText.value = ''
  grantUserText.value = ''
  grantExpiryDateArr.value = []
}

const viewDetail = (row) => { selectedApp.value = row; detailDialogVisible.value = true }

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

const submitRequest = async () => {
  requesting.value = true
  try {
    await createApplication({
      type: 'temp_permission',
      data: {
        permission: requestForm.permission,
        expiry_date: requestForm.expiry_date,
        reason: requestForm.reason
      },
      reason: requestForm.reason
    })
    showToast({ message: '申请已提交，请等待管理员审批', type: 'success' })
    requestDialogVisible.value = false
    loadMyApplications()
    loadStats()
  } catch (err) {
    showToast({ message: err.message || '申请提交失败', type: 'fail' })
  } finally {
    requesting.value = false
  }
}

const submitGrant = async () => {
  if (grantForm.user_ids.length === 0) {
    showToast({ message: '请至少选择一个用户', type: 'fail' })
    return
  }
  granting.value = true
  try {
    await grantTempPermission({
      user_ids: grantForm.user_ids,
      permission: grantForm.permission,
      resource_type: 'all',
      resource_id: 0,
      expires_at: new Date(grantForm.expires_at).toISOString(),
      reason: grantForm.reason
    })
    showToast({ message: `成功授权 ${grantForm.user_ids.length} 位用户`, type: 'success' })
    grantDialogVisible.value = false
    loadPermissions()
  } catch (err) {
    showToast({ message: err.message || '授权失败', type: 'fail' })
  } finally {
    granting.value = false
  }
}

const submitApproval = async () => {
  if (approvalAction.value === 'reject' && !approvalForm.comment) {
    showToast({ message: '请填写拒绝原因', type: 'fail' })
    return
  }
  approving.value = true
  try {
    await processApproval(selectedApp.value.id, {
      action: approvalAction.value,
      comment: approvalForm.comment
    })
    showToast({ message: approvalAction.value === 'approve' ? '已通过申请' : '已拒绝申请', type: 'success' })
    approvalDialogVisible.value = false
    loadPendingApprovals()
    loadStats()
  } catch (err) {
    showToast({ message: err.message || '审批处理失败', type: 'fail' })
  } finally {
    approving.value = false
  }
}

const cancelApp = async (row) => {
  try {
    await showConfirmDialog({ title: '确认', message: '确定撤销此申请吗？' })
    await cancelApplication(row.id)
    showToast({ message: '申请已撤销', type: 'success' })
    loadMyApplications()
    loadStats()
  } catch {}
}

const revokePermission = async (row) => {
  try {
    await showConfirmDialog({ title: '确认', message: `确定撤销 ${row.user_name} 的权限吗？` })
    await revokeTempPermission(row.id)
    showToast({ message: '撤销成功', type: 'success' })
    loadPermissions()
  } catch {}
}

onMounted(() => {
  loadPermissions()
  if (canGrant.value) loadUsers()
  loadMyApplications()
  loadStats()
  if (canApprove.value) loadPendingApprovals()
})
</script>

<style scoped>
.main-container {
  padding: env(safe-area-inset-top) 0 calc(var(--space-xl, 24px) + env(safe-area-inset-bottom));
  background: var(--color-bg, #F7F8FA);
  min-height: 100%;
}

.stats-cards {
  display: flex;
  gap: 10px;
  padding: 12px 16px;
  background: #fff;
}

.stat-card {
  flex: 1;
  background: #f7f8fa;
  border-radius: 8px;
  padding: 12px 8px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s;
  border: 2px solid transparent;
}

.stat-card.active { border-color: #1989fa; background: #ecf5ff; }
.stat-value { font-size: 24px; font-weight: bold; color: #1989fa; }
.stat-value.stat-warning { color: #ff976a; }
.stat-value.stat-success { color: #07c160; }
.stat-value.stat-danger { color: #ee0a24; }
.stat-label { margin-top: 4px; color: #969799; font-size: 12px; }

.tab-content { padding: 12px 16px; }
.action-bar { display: flex; justify-content: flex-end; margin-bottom: 12px; }
.tab-badge { margin-left: 4px; vertical-align: middle; }

.loading-wrapper { display: flex; justify-content: center; padding: 40px 0; }
.empty-wrapper { padding: 20px 0; }
.card-list { display: flex; flex-direction: column; gap: 12px; }

.perm-card {
  background: #fff; border-radius: 12px; padding: 14px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}
.perm-card-header { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; margin-bottom: 10px; }
.perm-card-name { font-weight: 600; font-size: 15px; color: #323233; }
.perm-card-body { display: flex; flex-direction: column; gap: 6px; margin-bottom: 10px; }
.perm-card-row { display: flex; align-items: center; gap: 8px; font-size: 13px; }
.perm-card-label { color: #969799; min-width: 52px; }
.perm-card-actions { display: flex; justify-content: flex-end; padding-top: 8px; border-top: 1px solid #f5f5f5; }

.app-card {
  background: #fff; border-radius: 12px; padding: 14px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}
.app-card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.app-card-no { font-size: 12px; color: #969799; }
.app-card-body { display: flex; flex-direction: column; gap: 6px; margin-bottom: 10px; }
.app-card-row { display: flex; align-items: center; gap: 8px; font-size: 13px; }
.app-card-label { color: #969799; min-width: 60px; }
.app-card-actions { display: flex; justify-content: flex-end; gap: 8px; padding-top: 8px; border-top: 1px solid #f5f5f5; }

.pending-card {
  background: #fff; border-radius: 12px; padding: 14px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}
.pending-card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.pending-card-user { display: flex; align-items: center; gap: 8px; }
.pending-card-name { font-weight: 600; font-size: 15px; color: #323233; }
.pending-card-body { display: flex; flex-direction: column; gap: 6px; margin-bottom: 10px; }
.pending-card-row { display: flex; align-items: center; gap: 8px; font-size: 13px; }
.pending-card-label { color: #969799; min-width: 60px; }
.pending-card-actions { display: flex; justify-content: flex-end; gap: 8px; padding-top: 8px; border-top: 1px solid #f5f5f5; }

.text-danger { color: #ee0a24; }
.text-gray { color: #969799; font-size: 12px; }
.load-more { padding: 16px 0; }

.popup-content { padding: 16px; }
.popup-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; padding: 0 4px; }
.popup-title { font-size: 16px; font-weight: 600; color: #323233; }
.popup-actions { display: flex; gap: 12px; margin-top: 20px; padding: 0 16px; }
.popup-actions .van-button { flex: 1; }
.mb-4 { margin-bottom: 12px; }
.detail-content { text-align: right; font-size: 13px; line-height: 1.6; }

.user-picker-popup { display: flex; flex-direction: column; height: 60vh; }
.user-picker-popup .popup-header { padding: 16px; border-bottom: 1px solid #f5f5f5; margin-bottom: 0; }
.user-list { flex: 1; overflow-y: auto; }
.user-picker-footer { padding: 12px 16px; border-top: 1px solid #f5f5f5; }
</style>
