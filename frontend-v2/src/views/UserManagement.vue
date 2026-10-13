<template>
  <div class="main-container">
    <div class="card">
      <div class="card-header">
        <span class="card-title">
          {{ isGlobalAdmin ? '用户管理' : '部门用户管理' }}
          <el-tag v-if="!isGlobalAdmin" size="small" type="info" style="margin-left: 8px">{{ userStore.user?.department }}</el-tag>
        </span>
        <div class="header-actions">
          <el-select v-model="statusFilter" placeholder="状态筛选" style="width: 140px" @change="fetchUsers">
            <el-option label="全部" value="" />
            <el-option label="待入职" value="pending" />
            <el-option label="在职" value="active" />
            <el-option label="离职" value="resigned" />
          </el-select>
          <el-button v-if="canManageAll" type="primary" @click="openMemberDialog">录入正式成员</el-button>
          <el-button v-if="canManageAll" type="warning" :disabled="selectedIds.length === 0" @click="openBatchResign">
            批量离职（{{ selectedIds.length }}）
          </el-button>
        </div>
      </div>

      <el-table
        :data="userList"
        v-loading="loading"
        class="data-table"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="50" :selectable="row => row.employment_status !== 'resigned'" />
        <el-table-column prop="student_id" label="学号" width="120" />
        <el-table-column label="姓名" width="120">
          <template #default="{ row }">
            <span>{{ row.name }}</span>
            <el-tag v-if="row.employment_status === 'pending'" size="small" type="info" style="margin-left: 6px">待激活</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="email" label="邮箱" />
        <el-table-column prop="department" label="部门" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ row.department || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="角色" width="220">
          <template #default="{ row }">
            <el-space>
              <el-tag v-if="row.role === 'admin'" type="danger" size="small">系统管理员</el-tag>
              <el-tag v-else type="info" size="small">普通用户</el-tag>
              <el-tag v-if="row.dept_role === 'dept_admin'" type="warning" size="small">部门管理员</el-tag>
            </el-space>
            <div v-if="row.department === '办公室' && row.dept_role === 'dept_admin'" class="office-admin-badge">
              <el-tag type="success" size="small">办公室管理员</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="statusTag(row.employment_status).type">
              {{ statusTag(row.employment_status).text }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240">
          <template #default="{ row }">
            <el-button v-if="!isSelf(row)" link type="primary" @click="editUser(row)">编辑</el-button>
            <template v-if="row.employment_status === 'pending' && canManageAll">
              <el-button link type="warning" @click="resend(row)">重发邀请</el-button>
              <el-button link type="danger" @click="revoke(row)">作废</el-button>
            </template>
            <el-button
              v-if="row.employment_status !== 'resigned' && !isSelf(row) && canManageAll"
              link
              type="danger"
              @click="resignUser(row)"
            >离职</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 录入正式成员 -->
    <el-dialog v-model="memberDialogVisible" title="录入正式成员" width="560px">
      <el-tabs v-model="memberTab">
        <el-tab-pane label="单个录入" name="single">
          <el-form :model="memberForm" :rules="memberRules" ref="memberFormRef" label-width="80px">
            <el-form-item label="姓名" prop="name">
              <el-input v-model="memberForm.name" />
            </el-form-item>
            <el-form-item label="学号" prop="student_id">
              <el-input v-model="memberForm.student_id" />
            </el-form-item>
            <el-form-item label="部门" prop="department">
              <el-select v-model="memberForm.department" style="width: 100%">
                <el-option v-for="dept in departments" :key="dept" :label="dept" :value="dept" />
              </el-select>
            </el-form-item>
            <el-form-item label="部门角色" prop="dept_role">
              <el-radio-group v-model="memberForm.dept_role">
                <el-radio-button value="dept_member">部门成员</el-radio-button>
                <el-radio-button value="dept_admin">部门管理员</el-radio-button>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="邮箱" prop="email">
              <el-input v-model="memberForm.email" placeholder="用于接收激活邮件" />
            </el-form-item>
          </el-form>
          <p class="form-hint">录入后将自动发送激活邮件，由成员自行设置密码，无需在此设置密码。</p>
        </el-tab-pane>
        <el-tab-pane label="批量录入" name="batch">
          <el-input
            v-model="batchText"
            type="textarea"
            :rows="8"
            placeholder="每行一条，格式：学号,姓名,部门,邮箱&#10;例如：&#10;2021001,张三,办公室,zhangsan@example.com&#10;2021002,李四,竞赛部,lisi@example.com"
          />
          <p class="form-hint">支持逗号分隔的文本或 CSV；将逐条录入并汇总成功/失败。</p>
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="memberDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitMember" :loading="submitting">
          {{ memberTab === 'single' ? '录入并发送激活邮件' : '批量录入' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 批量录入结果 -->
    <el-dialog v-model="batchResultVisible" title="批量录入结果" width="560px">
      <el-alert v-if="batchResult.success.length" type="success" :closable="false"
        :title="`成功 ${batchResult.success.length} 条`" style="margin-bottom: 12px" />
      <el-alert v-if="batchResult.failed.length" type="error" :closable="false"
        :title="`失败 ${batchResult.failed.length} 条`" style="margin-bottom: 12px" />
      <el-table v-if="batchResult.failed.length" :data="batchResult.failed" size="small" border>
        <el-table-column prop="student_id" label="学号" width="120" />
        <el-table-column prop="name" label="姓名" width="100" />
        <el-table-column prop="reason" label="失败原因" />
      </el-table>
      <template #footer>
        <el-button type="primary" @click="batchResultVisible = false">确定</el-button>
      </template>
    </el-dialog>

    <!-- 编辑用户 -->
    <el-dialog v-model="editDialogVisible" title="编辑用户" width="500px">
      <el-form :model="editForm" :rules="editRules" ref="editFormRef" label-width="80px">
        <el-form-item label="学号" prop="student_id">
          <el-input v-model="editForm.student_id" :disabled="true" />
        </el-form-item>
        <el-form-item label="姓名" prop="name">
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="editForm.email" />
        </el-form-item>
        <el-form-item v-if="canManageAll" label="部门">
          <el-select v-model="editForm.department" style="width: 100%">
            <el-option v-for="dept in departments" :key="dept" :label="dept" :value="dept" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="canManageAll" label="系统角色">
          <el-radio-group v-model="editForm.role">
            <el-radio-button value="user">普通用户</el-radio-button>
            <el-radio-button value="admin">系统管理员</el-radio-button>
          </el-radio-group>
          <div class="role-hint">提示：办公室管理员请选择部门为"办公室"并开启部门管理员</div>
        </el-form-item>
        <el-form-item v-if="canManageAll" label="部门角色">
          <el-switch v-model="editForm.isDeptAdmin" active-text="部门管理员" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitEdit" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <!-- 批量离职预览 -->
    <el-dialog v-model="batchResignVisible" title="批量离职确认" width="600px">
      <el-alert type="warning" :closable="false" style="margin-bottom: 16px">
        离职将标记成员不可登录、不可再次注册，并联动取消待处理排班、失效临时权限。
      </el-alert>
      <el-descriptions :column="1" border>
        <el-descriptions-item label="将离职用户数">{{ effectiveResignCount }} 人</el-descriptions-item>
        <el-descriptions-item label="将取消的待处理排班数">{{ previewData?.pending_duty_count ?? 0 }}</el-descriptions-item>
        <el-descriptions-item label="将失效的临时权限数">{{ previewData?.active_temp_perm_count ?? 0 }}</el-descriptions-item>
      </el-descriptions>
      <div class="resign-list-title">勾选需离职的成员（取消勾选 = 晋升例外，保留在职）：</div>
      <el-checkbox-group v-model="resignCheckedIds" class="resign-list">
        <el-checkbox v-for="u in previewData?.affected_users || []" :key="u.id" :value="u.id" border>
          {{ u.name }}（{{ u.student_id }} · {{ u.department }}）
        </el-checkbox>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="batchResignVisible = false">取消</el-button>
        <el-button type="danger" @click="confirmBatchResign" :loading="submitting">确认离职</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getUserList,
  getUserListByDepartment,
  updateUser,
  getDepartments,
  createMember,
  batchCreateMembers,
  resendInvite,
  revokeInvite,
  resignMember,
  previewBatchResign,
  batchResignMembers
} from '../api/user'
import { useUserStore } from '../stores/user'

const userStore = useUserStore()

// 是否为全局管理员（系统管理员或办公室管理员），可管理所有用户
const isGlobalAdmin = computed(() => userStore.canManageAll)
// 批量离职按钮：仅管理员可见（部门管理员限本部门由后端权限控制，前端按 canManageAll 暴露）
const canManageAll = computed(() => userStore.canManageAll)

const loading = ref(false)
const userList = ref([])
const departments = ref([])
const statusFilter = ref('')
const selectedIds = ref([])

const statusMeta = {
  pending: { text: '待入职', type: 'warning' },
  active: { text: '在职', type: 'success' },
  resigned: { text: '离职', type: 'info' }
}
// 后端当前未在列表接口返回 employment_status 时（见交付说明），默认按在职展示，避免空列
const statusTag = (s) => statusMeta[s] || { text: '在职', type: 'success' }

// 录入对话框
const memberDialogVisible = ref(false)
const memberTab = ref('single')
const submitting = ref(false)
const memberFormRef = ref()
const memberForm = reactive({
  name: '',
  student_id: '',
  department: '',
  dept_role: 'dept_member',
  email: ''
})
const memberRules = {
  name: [{ required: true, message: '请输入姓名' }],
  student_id: [{ required: true, message: '请输入学号' }],
  department: [{ required: true, message: '请选择部门' }],
  dept_role: [{ required: true, message: '请选择部门角色' }],
  email: [{ required: true, message: '请输入邮箱' }, { type: 'email', message: '邮箱格式不正确' }]
}
const batchText = ref('')

// 批量录入结果
const batchResultVisible = ref(false)
const batchResult = reactive({ success: [], failed: [] })

// 编辑对话框
const editDialogVisible = ref(false)
const editFormRef = ref()
const editForm = reactive({
  id: null,
  student_id: '',
  name: '',
  email: '',
  department: '',
  role: 'user',
  isDeptAdmin: false
})
const editRules = {
  name: [{ required: true, message: '请输入姓名' }],
  email: [{ required: true, message: '请输入邮箱' }, { type: 'email', message: '邮箱格式不正确' }]
}

// 批量离职预览
const batchResignVisible = ref(false)
const previewData = ref(null)
const resignCheckedIds = ref([])
const effectiveResignCount = computed(() => resignCheckedIds.value.length)

const fetchUsers = async () => {
  loading.value = true
  try {
    const params = {}
    if (statusFilter.value) params.status = statusFilter.value
    let data
    if (isGlobalAdmin.value) {
      data = await getUserList(params)
    } else {
      const dept = userStore.user?.department
      data = await getUserListByDepartment(dept)
    }
    userList.value = data || []
  } finally {
    loading.value = false
  }
}

const fetchDepartments = async () => {
  try {
    const data = await getDepartments()
    departments.value = data?.departments || []
  } catch {}
}

const handleSelectionChange = (rows) => {
  selectedIds.value = rows.map(r => r.id)
}

const isSelf = (row) => row.id === userStore.user?.id

const openMemberDialog = () => {
  memberTab.value = 'single'
  batchText.value = ''
  Object.assign(memberForm, {
    name: '',
    student_id: '',
    department: userStore.user?.department || '',
    dept_role: 'dept_member',
    email: ''
  })
  memberDialogVisible.value = true
}

const submitMember = async () => {
  submitting.value = true
  try {
    if (memberTab.value === 'single') {
      const valid = await memberFormRef.value.validate().catch(() => false)
      if (!valid) { submitting.value = false; return }
      await createMember({ ...memberForm })
      ElMessage.success('已录入并发送激活邮件')
      memberDialogVisible.value = false
      fetchUsers()
    } else {
      const items = parseBatchText(batchText.value)
      if (!items.length) {
        ElMessage.warning('请按「学号,姓名,部门,邮箱」格式粘贴数据')
        submitting.value = false
        return
      }
      const res = await batchCreateMembers(items)
      batchResult.success = res?.success || []
      batchResult.failed = res?.failed || []
      batchResultVisible.value = true
      memberDialogVisible.value = false
      fetchUsers()
    }
  } catch (err) {
    ElMessage.error(err.message || '录入失败')
  } finally {
    submitting.value = false
  }
}

// 解析批量录入文本：支持逗号/中文逗号/制表符/空白分隔
const parseBatchText = (text) => {
  const lines = (text || '').split(/\r?\n/).map(l => l.trim()).filter(Boolean)
  const items = []
  for (const line of lines) {
    let parts = line.split(/[,，\t]/).map(s => s.trim()).filter(Boolean)
    if (parts.length < 4) {
      parts = line.split(/\s+/).map(s => s.trim()).filter(Boolean)
    }
    if (parts.length < 4) continue
    const [student_id, name, department, email] = parts
    items.push({ student_id, name, department, email, dept_role: 'dept_member' })
  }
  return items
}

const editUser = (row) => {
  editDialogVisible.value = true
  Object.assign(editForm, {
    id: row.id,
    student_id: row.student_id,
    name: row.name,
    email: row.email,
    department: row.department,
    role: row.role,
    isDeptAdmin: row.dept_role === 'dept_admin'
  })
}

const submitEdit = async () => {
  const valid = await editFormRef.value.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    const data = {
      name: editForm.name,
      email: editForm.email
    }
    if (canManageAll.value) {
      data.department = editForm.department
      data.role = editForm.role
      data.dept_role = editForm.isDeptAdmin ? 'dept_admin' : 'dept_member'
    }
    await updateUser(editForm.id, data)
    ElMessage.success('更新成功')
    editDialogVisible.value = false
    fetchUsers()
  } catch (err) {
    ElMessage.error(err.message || '更新失败')
  } finally {
    submitting.value = false
  }
}

const resend = async (row) => {
  try {
    await ElMessageBox.confirm(`确定向 ${row.name} 重新发送激活邀请吗？`, '重发邀请', { type: 'warning' })
    await resendInvite(row.id)
    ElMessage.success('邀请已重新发送')
  } catch {}
}

const revoke = async (row) => {
  try {
    await ElMessageBox.confirm(`确定作废 ${row.name} 的激活邀请吗？作废后需重新录入。`, '作废邀请', { type: 'warning' })
    await revokeInvite(row.id)
    ElMessage.success('邀请已作废')
    fetchUsers()
  } catch {}
}

const resignUser = async (row) => {
  try {
    await ElMessageBox.confirm(
      `标记离职后该成员不可登录且不可再次注册，确定将 "${row.name}" 标记为离职？`,
      '确认离职',
      { type: 'warning' }
    )
    await resignMember(row.id)
    ElMessage.success('已标记为离职')
    fetchUsers()
  } catch {}
}

const openBatchResign = async () => {
  submitting.value = false
  try {
    const res = await previewBatchResign(selectedIds.value, [])
    previewData.value = res || null
    resignCheckedIds.value = (res?.affected_users || []).map(u => u.id)
    batchResignVisible.value = true
  } catch (err) {
    ElMessage.error(err.message || '获取预览失败')
  }
}

const confirmBatchResign = async () => {
  if (!resignCheckedIds.value.length) {
    ElMessage.warning('请至少选择一名离职成员')
    return
  }
  const allIds = (previewData.value?.affected_users || []).map(u => u.id)
  const excludeIds = allIds.filter(id => !resignCheckedIds.value.includes(id))
  submitting.value = true
  try {
    const res = await batchResignMembers(resignCheckedIds.value, excludeIds)
    ElMessage.success(`已标记 ${res?.affected_user_count ?? resignCheckedIds.value.length} 人离职`)
    batchResignVisible.value = false
    selectedIds.value = []
    fetchUsers()
  } catch (err) {
    ElMessage.error(err.message || '批量离职失败')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  fetchUsers()
  fetchDepartments()
})
</script>

<style scoped>
.header-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}
.role-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 8px;
  line-height: 1.4;
}
.office-admin-badge {
  margin-top: 4px;
}
.form-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 8px;
  line-height: 1.5;
}
.resign-list-title {
  margin: 16px 0 8px;
  font-size: 13px;
  color: #606266;
}
.resign-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 240px;
  overflow-y: auto;
}
</style>
