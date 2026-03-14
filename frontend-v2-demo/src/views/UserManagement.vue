<template>
  <div class="main-container">
    <div class="card">
      <div class="card-header">
        <span class="card-title">用户管理</span>
      </div>

      <el-table :data="userList" v-loading="loading" class="data-table">
        <el-table-column prop="student_id" label="学号" width="120" />
        <el-table-column prop="name" label="姓名" width="100" />
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
        <el-table-column label="操作" width="180">
          <template #default="{ row }">
            <el-button link type="primary" @click="editUser(row)">编辑</el-button>
            <el-button link type="danger" @click="deleteUserById(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 编辑用户 -->
    <el-dialog v-model="dialogVisible" title="编辑用户" width="500px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="学号" prop="student_id">
          <el-input v-model="form.student_id" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="姓名" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item label="部门">
          <el-select v-model="form.department" style="width: 100%">
            <el-option label="办公室" value="办公室" />
            <el-option label="竞赛部" value="竞赛部" />
            <el-option label="项目部" value="项目部" />
            <el-option label="科普部" value="科普部" />
          </el-select>
        </el-form-item>
        <el-form-item label="系统角色">
          <el-radio-group v-model="form.role">
            <el-radio-button value="user">普通用户</el-radio-button>
            <el-radio-button value="admin">系统管理员</el-radio-button>
          </el-radio-group>
          <div class="role-hint">提示：办公室管理员请选择部门为"办公室"并开启部门管理员</div>
        </el-form-item>
        <el-form-item label="部门角色">
          <el-switch v-model="form.isDeptAdmin" active-text="部门管理员" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getUserList, updateUser, deleteUser } from '../api/user'

const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const userList = ref([])
const formRef = ref()

const form = reactive({
  id: null,
  student_id: '',
  name: '',
  email: '',
  department: '',
  role: 'user',
  isDeptAdmin: false
})

const rules = {
  student_id: [{ required: true, message: '请输入学号' }],
  name: [{ required: true, message: '请输入姓名' }],
  email: [{ required: true, message: '请输入邮箱' }, { type: 'email', message: '邮箱格式不正确' }]
}

const fetchUsers = async () => {
  loading.value = true
  try {
    const data = await getUserList()
    userList.value = data || []
  } finally {
    loading.value = false
  }
}

const editUser = (row) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id,
    student_id: row.student_id,
    name: row.name,
    email: row.email,
    department: row.department,
    role: row.role,  // role 只能是 'user' 或 'admin'
    isDeptAdmin: row.dept_role === 'dept_admin',
    password: ''
  })
  dialogVisible.value = true
}

const submitForm = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const data = {
      student_id: form.student_id,
      name: form.name,
      email: form.email,
      department: form.department,
      role: form.role,
      dept_role: form.isDeptAdmin ? 'dept_admin' : 'dept_member'
    }

    await updateUser(form.id, data)
    ElMessage.success('更新成功')
    dialogVisible.value = false
    fetchUsers()
  } finally {
    submitting.value = false
  }
}

const deleteUserById = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除用户 "${row.name}" 吗？`, '确认删除', { type: 'warning' })
    await deleteUser(row.id)
    ElMessage.success('删除成功')
    fetchUsers()
  } catch {}
}

onMounted(() => {
  fetchUsers()
})
</script>

<style scoped>
.role-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 8px;
  line-height: 1.4;
}
.office-admin-badge {
  margin-top: 4px;
}
</style>
