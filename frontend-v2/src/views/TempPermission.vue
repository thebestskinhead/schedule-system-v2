<template>
  <div class="main-container">
    <div class="card">
      <div class="card-header">
        <span class="card-title">临时权限管理</span>
        <el-button type="primary" @click="grantDialogVisible = true">
          <el-icon><Plus /></el-icon> 授予权限
        </el-button>
      </div>

      <el-table :data="permissionList" v-loading="loading" class="data-table">
        <el-table-column prop="user_name" label="用户" width="120" />
        <el-table-column label="权限类型" width="180">
          <template #default="{ row }">
            <el-tag size="small">{{ getPermissionText(row.permission) }}</el-tag>
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
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button link type="danger" @click="revokePermission(row)">撤销</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 授予权限 -->
    <el-dialog v-model="grantDialogVisible" title="授予临时权限" width="450px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="用户" prop="user_id">
          <el-select-v2 v-model="form.user_id" :options="userOptions" placeholder="选择用户" style="width: 100%" />
        </el-form-item>
        <el-form-item label="权限类型" prop="permission">
          <el-select v-model="form.permission" placeholder="选择权限" style="width: 100%">
            <el-option label="部门排班管理" value="schedule:manage:dept" />
            <el-option label="部门用户管理" value="user:manage:dept" />
            <el-option label="全局排班查看" value="schedule:view:all" />
            <el-option label="全局用户管理" value="user:manage:all" />
          </el-select>
        </el-form-item>
        <el-form-item label="资源类型" prop="resource_type">
          <el-select v-model="form.resource_type" placeholder="选择资源类型" style="width: 100%">
            <el-option label="全部" value="all" />
            <el-option label="部门" value="dept" />
            <el-option label="用户" value="user" />
          </el-select>
        </el-form-item>
        <el-form-item label="有效期" prop="expires_hours">
          <el-slider v-model="form.expires_hours" :max="168" :marks="{24:'1天', 72:'3天', 168:'7天'}" />
          <span>{{ form.expires_hours }} 小时</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="grantDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitGrant" :loading="granting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getTempPermissions, grantTempPermission, revokeTempPermission } from '../api/system'
import { getUserList } from '../api/user'

const loading = ref(false)
const grantDialogVisible = ref(false)
const granting = ref(false)
const permissionList = ref([])
const userOptions = ref([])
const formRef = ref()

const form = reactive({
  user_id: null,
  permission: '',
  resource_type: 'all',
  resource_id: 0,
  expires_hours: 24,
  reason: ''
})

const rules = {
  user_id: [{ required: true, message: '请选择用户' }],
  permission: [{ required: true, message: '请选择权限类型' }],
  resource_type: [{ required: true, message: '请选择资源类型' }]
}

const permissionMap = {
  'schedule:manage:dept': '部门排班管理',
  'user:manage:dept': '部门用户管理',
  'schedule:view:all': '全局排班查看',
  'user:manage:all': '全局用户管理'
}

const getPermissionText = (perm) => permissionMap[perm] || perm

const isExpired = (time) => new Date(time) < new Date()

const formatTime = (time) => {
  if (!time) return '-'
  const date = new Date(time)
  return date.toLocaleString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

const loadPermissions = async () => {
  loading.value = true
  try {
    const data = await getTempPermissions()
    permissionList.value = data || []
  } finally {
    loading.value = false
  }
}

const loadUsers = async () => {
  try {
    const data = await getUserList()
    userOptions.value = (data || []).map(u => ({ value: u.id, label: `${u.name} (${u.student_id})` }))
  } catch {}
}

const submitGrant = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  granting.value = true
  try {
    // 计算过期时间
    const expiresAt = new Date(Date.now() + form.expires_hours * 60 * 60 * 1000)
    
    await grantTempPermission({
      user_id: form.user_id,
      permission: form.permission,
      resource_type: form.resource_type,
      resource_id: form.resource_id,
      expires_at: expiresAt.toISOString(),
      reason: form.reason
    })
    ElMessage.success('授权成功')
    grantDialogVisible.value = false
    loadPermissions()
  } finally {
    granting.value = false
  }
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
  loadPermissions()
  loadUsers()
})
</script>

<style scoped>
.text-danger {
  color: #f56c6c;
}
</style>
