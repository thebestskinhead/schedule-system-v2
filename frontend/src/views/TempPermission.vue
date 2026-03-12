<template>
  <div class="temp-permission-page">
    <el-row :gutter="20">
      <el-col :span="8">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>授予临时权限</span>
            </div>
          </template>

          <el-form :model="grantForm" label-position="top">
            <el-form-item label="选择用户">
              <el-select v-model="grantForm.user_id" filterable placeholder="请选择用户" style="width: 100%">
                <el-option
                  v-for="user in users"
                  :key="user.id"
                  :label="`${user.name} (${user.student_id})`"
                  :value="user.id"
                />
              </el-select>
            </el-form-item>

            <el-form-item label="权限">
              <el-select v-model="grantForm.permission" placeholder="请选择权限" style="width: 100%">
                <el-option
                  v-for="perm in permissions"
                  :key="perm.code"
                  :label="perm.name"
                  :value="perm.code"
                />
              </el-select>
            </el-form-item>

            <el-form-item label="资源类型">
              <el-radio-group v-model="grantForm.resource_type" @change="onResourceTypeChange">
                <el-radio label="all">全部</el-radio>
                <el-radio label="dept">部门</el-radio>
                <el-radio label="user">用户</el-radio>
              </el-radio-group>
            </el-form-item>

            <!-- 部门选择 -->
            <el-form-item label="选择部门" v-if="grantForm.resource_type === 'dept'">
              <el-select v-model="selectedDept" placeholder="请选择部门" style="width: 100%" @change="onDeptChange">
                <el-option label="全部部门" value="" />
                <el-option v-for="dept in departments" :key="dept" :label="dept" :value="dept" />
              </el-select>
              <span class="form-hint">选择"全部部门"表示用户所在部门（用户部门变更时自动跟随）</span>
            </el-form-item>

            <!-- 用户ID输入（保留但通常不需要） -->
            <el-form-item label="资源ID" v-if="grantForm.resource_type === 'user'">
              <el-input-number v-model="grantForm.resource_id" :min="1" style="width: 100%" placeholder="用户ID" />
            </el-form-item>

            <el-form-item label="过期时间">
              <el-date-picker
                v-model="grantForm.expires_at"
                type="datetime"
                placeholder="选择过期时间"
                style="width: 100%"
              />
            </el-form-item>

            <el-form-item label="授权原因">
              <el-input
                v-model="grantForm.reason"
                type="textarea"
                rows="2"
                placeholder="请输入授权原因"
              />
            </el-form-item>

            <el-form-item>
              <el-button type="primary" @click="grantPermission" :loading="granting" style="width: 100%">
                <el-icon><Plus /></el-icon>
                授予权限
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>

      <el-col :span="16">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>当前临时权限</span>
              <el-button type="danger" size="small" @click="cleanupExpired">
                <el-icon><Delete /></el-icon>
                清理过期
              </el-button>
            </div>
          </template>

          <el-table :data="activePermissions" border v-loading="loading">
            <el-table-column prop="user_name" label="用户" width="100" />
            <el-table-column prop="permission_name" label="权限" width="120" />
            <el-table-column prop="resource_name" label="资源" width="100" />
            <el-table-column prop="granted_by_name" label="授权人" width="100" />
            <el-table-column label="过期时间" width="160">
              <template #default="{ row }">
                {{ formatDate(row.expires_at) }}
              </template>
            </el-table-column>
            <el-table-column prop="reason" label="原因" show-overflow-tooltip />
            <el-table-column label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.is_expired ? 'info' : 'success'" size="small">
                  {{ row.is_expired ? '已过期' : '有效' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="80">
              <template #default="{ row }">
                <el-button type="danger" size="small" @click="revokePermission(row.id)">
                  撤销
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import { tempPermissionAPI } from '../api/tempPermission.js'
import { userAPI } from '../api/user.js'

const users = ref([])
const permissions = ref([])
const activePermissions = ref([])
const loading = ref(false)
const granting = ref(false)

const grantForm = ref({
  user_id: '',
  permission: '',
  resource_type: 'all',
  resource_id: 0,
  expires_at: '',
  reason: ''
})

// 部门列表
const departments = ref(['办公室', '竞赛部', '项目部', '科普部'])
const selectedDept = ref('')

// 资源类型变化时重置
const onResourceTypeChange = (type) => {
  if (type === 'dept') {
    selectedDept.value = ''
    grantForm.value.resource_id = 0
  } else if (type === 'all') {
    selectedDept.value = ''
    grantForm.value.resource_id = 0
  }
}

// 部门选择变化
const onDeptChange = (dept) => {
  // 部门权限：resource_id = 0 表示用户当前部门（动态）
  // 这样当用户部门变更时，权限仍然有效
  grantForm.value.resource_id = 0
}

const loadUsers = async () => {
  try {
    const res = await userAPI.getList()
    users.value = res || []
  } catch (err) {
    console.error('加载用户失败:', err)
  }
}

const loadPermissions = async () => {
  try {
    const res = await tempPermissionAPI.getPermissionList()
    permissions.value = res || []
  } catch (err) {
    console.error('加载权限列表失败:', err)
  }
}

const loadActivePermissions = async () => {
  loading.value = true
  try {
    const res = await tempPermissionAPI.list()
    activePermissions.value = res || []
  } catch (err) {
    console.error('加载权限列表失败:', err)
  } finally {
    loading.value = false
  }
}

const grantPermission = async () => {
  if (!grantForm.value.user_id || !grantForm.value.permission || !grantForm.value.expires_at) {
    ElMessage.warning('请填写完整信息')
    return
  }

  granting.value = true
  try {
    const payload = {
      user_id: grantForm.value.user_id,
      permission: grantForm.value.permission,
      resource_type: grantForm.value.resource_type,
      resource_id: grantForm.value.resource_type === 'all' ? 0 : (grantForm.value.resource_id || 0),
      expires_at: grantForm.value.expires_at,
      reason: grantForm.value.reason
    }
    await tempPermissionAPI.grant(payload)
    ElMessage.success('授权成功！')
    grantForm.value = {
      user_id: '',
      permission: '',
      resource_type: 'all',
      resource_id: 0,
      expires_at: '',
      reason: ''
    }
    loadActivePermissions()
  } catch (err) {
    ElMessage.error('授权失败: ' + (err.response?.data?.message || err.message))
  } finally {
    granting.value = false
  }
}

const revokePermission = async (id) => {
  try {
    await ElMessageBox.confirm('确定要撤销此权限吗？', '提示', { type: 'warning' })
    await tempPermissionAPI.revoke(id)
    ElMessage.success('撤销成功！')
    loadActivePermissions()
  } catch (err) {
    if (err !== 'cancel') {
      ElMessage.error('撤销失败: ' + (err.response?.data?.message || err.message))
    }
  }
}

const cleanupExpired = async () => {
  try {
    await tempPermissionAPI.cleanup()
    ElMessage.success('清理完成')
    loadActivePermissions()
  } catch (err) {
    ElMessage.error('清理失败')
  }
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

onMounted(() => {
  loadUsers()
  loadPermissions()
  loadActivePermissions()
})
</script>

<style scoped>
.temp-permission-page {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 16px;
  font-weight: bold;
}
</style>
