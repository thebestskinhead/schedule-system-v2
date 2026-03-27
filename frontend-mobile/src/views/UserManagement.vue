<template>
  <div class="main-container">
    <Navbar :title="isGlobalAdmin ? '用户管理' : '部门用户管理'" show-back />
    <div class="card">
      <div class="card-header">
        <span class="card-title">
          {{ isGlobalAdmin ? '用户管理' : '部门用户管理' }}
          <van-tag v-if="!isGlobalAdmin" size="medium" type="default" style="margin-left: 8px">{{ userStore.user?.department }}</van-tag>
        </span>
        <van-button v-if="isGlobalAdmin" type="primary" size="small" @click="openCreateDialog">新增用户</van-button>
      </div>

      <!-- 用户卡片列表 -->
      <div class="user-list">
        <van-loading v-if="loading" style="text-align: center; padding: 30px;" />
        <van-empty v-else-if="userList.length === 0" description="暂无用户" />
        <template v-else>
          <div v-for="row in userList" :key="row.id" class="user-card">
            <div class="user-card-top">
              <div class="user-info">
                <div class="user-name">
                  {{ row.name }}
                  <van-tag v-if="row.role === 'admin'" type="danger" size="medium" style="margin-left: 6px">管理员</van-tag>
                  <van-tag v-if="row.dept_role === 'dept_admin'" type="warning" size="medium" style="margin-left: 6px">部门管理员</van-tag>
                </div>
                <div class="user-detail">学号：{{ row.student_id }}</div>
                <div class="user-detail">邮箱：{{ row.email }}</div>
                <div class="user-detail">
                  <van-tag v-if="row.department" plain size="medium">{{ row.department }}</van-tag>
                </div>
                <div v-if="row.department === '办公室' && row.dept_role === 'dept_admin'" class="office-admin-badge">
                  <van-tag type="success" size="medium">办公室管理员</van-tag>
                </div>
              </div>
            </div>
            <div class="user-card-actions" v-if="!isSelf(row)">
              <van-button size="small" type="primary" plain @click="editUser(row)">编辑</van-button>
              <van-button v-if="canDeleteUser(row)" size="small" type="danger" plain @click="deleteUserById(row)">删除</van-button>
            </div>
          </div>
        </template>
      </div>
    </div>

    <!-- 编辑/新增用户弹窗 -->
    <van-popup
      v-model:show="dialogVisible"
      position="bottom"
      round
      :style="{ maxHeight: '85%' }"
    >
      <div class="popup-content">
        <div class="popup-header">
          <span class="popup-title">{{ isEdit ? '编辑用户' : '新增用户' }}</span>
          <van-icon name="cross" size="20" @click="dialogVisible = false" />
        </div>

        <van-form @submit="submitForm">
          <van-cell-group inset>
            <van-field
              v-model="form.student_id"
              label="学号"
              placeholder="请输入学号"
              :disabled="isEdit"
              :rules="[{ required: true, message: '请输入学号' }]"
            />
            <van-field
              v-model="form.name"
              label="姓名"
              placeholder="请输入姓名"
              :rules="[{ required: true, message: '请输入姓名' }]"
            />
            <van-field
              v-model="form.email"
              label="邮箱"
              placeholder="请输入邮箱"
              :rules="[
                { required: true, message: '请输入邮箱' },
                { pattern: /^[\w.-]+@[\w.-]+\.\w+$/, message: '邮箱格式不正确' }
              ]"
            />
            <van-field
              v-if="isGlobalAdmin"
              v-model="form.department"
              is-link
              readonly
              label="部门"
              placeholder="选择部门"
              @click="showDeptPicker = true"
            />
            <van-cell v-if="isGlobalAdmin" title="系统角色">
              <template #value>
                <van-radio-group v-model="form.role" direction="horizontal">
                  <van-radio name="user">普通用户</van-radio>
                  <van-radio name="admin">管理员</van-radio>
                </van-radio-group>
              </template>
            </van-cell>
            <van-cell v-if="isGlobalAdmin" title="部门管理员">
              <template #value>
                <van-switch v-model="form.isDeptAdmin" />
              </template>
            </van-cell>
            <van-field
              v-if="!isEdit && isGlobalAdmin"
              v-model="form.password"
              type="password"
              label="初始密码"
              placeholder="默认 123456"
            />
          </van-cell-group>
          <div class="role-hint" style="padding: 0 16px;" v-if="isGlobalAdmin">提示：办公室管理员请选择部门为"办公室"并开启部门管理员</div>
          <div style="padding: 24px 16px;">
            <van-button type="primary" block :loading="submitting" native-type="submit">确定</van-button>
            <van-button block plain style="margin-top: 10px;" @click="dialogVisible = false">取消</van-button>
          </div>
        </van-form>
      </div>
    </van-popup>

    <!-- 部门选择弹窗 -->
    <van-popup v-model:show="showDeptPicker" position="bottom" round>
      <van-picker
        :columns="deptColumns"
        @confirm="onDeptConfirm"
        @cancel="showDeptPicker = false"
      />
    </van-popup>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { showToast, showConfirmDialog } from 'vant'
import Navbar from '../components/Navbar.vue'
import { getUserList, getUserListByDepartment, createUser, updateUser, deleteUser } from '../api/user'
import { getDepartments } from '../api/user'
import { useUserStore } from '../stores/user'

const userStore = useUserStore()

const isGlobalAdmin = computed(() => userStore.canManageAll)

const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const userList = ref([])
const departments = ref([])
const showDeptPicker = ref(false)

const form = reactive({
  id: null,
  student_id: '',
  name: '',
  email: '',
  department: '',
  role: 'user',
  isDeptAdmin: false,
  password: ''
})

const deptColumns = computed(() =>
  departments.value.map(dept => ({ text: dept, value: dept }))
)

const onDeptConfirm = ({ selectedValues }) => {
  form.department = selectedValues[0]
  showDeptPicker.value = false
}

const fetchUsers = async () => {
  loading.value = true
  try {
    let data
    if (isGlobalAdmin.value) {
      data = await getUserList()
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

const canDeleteUser = (row) => {
  if (!isGlobalAdmin.value) {
    if (row.role === 'admin') return false
  }
  return true
}

const isSelf = (row) => {
  return row.id === userStore.user?.id
}

const openCreateDialog = () => {
  isEdit.value = false
  Object.assign(form, {
    id: null,
    student_id: '',
    name: '',
    email: '',
    department: userStore.user?.department || '',
    role: 'user',
    isDeptAdmin: false,
    password: ''
  })
  dialogVisible.value = true
}

const editUser = (row) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id,
    student_id: row.student_id,
    name: row.name,
    email: row.email,
    department: row.department,
    role: row.role,
    isDeptAdmin: row.dept_role === 'dept_admin',
    password: ''
  })
  dialogVisible.value = true
}

const submitForm = async () => {
  submitting.value = true
  try {
    if (isEdit.value) {
      const data = {
        name: form.name,
        email: form.email,
      }
      if (isGlobalAdmin.value) {
        data.department = form.department
        data.role = form.role
        data.dept_role = form.isDeptAdmin ? 'dept_admin' : 'dept_member'
      }
      await updateUser(form.id, data)
      showToast({ message: '更新成功', type: 'success' })
    } else {
      const data = {
        student_id: form.student_id,
        name: form.name,
        email: form.email,
        department: form.department,
        role: form.role,
        dept_role: form.isDeptAdmin ? 'dept_admin' : 'dept_member',
      }
      if (form.password) {
        data.password = form.password
      }
      await createUser(data)
      showToast({ message: '创建成功', type: 'success' })
    }
    dialogVisible.value = false
    fetchUsers()
  } finally {
    submitting.value = false
  }
}

const deleteUserById = async (row) => {
  try {
    await showConfirmDialog({
      title: '确认删除',
      message: `确定删除用户 "${row.name}" 吗？`,
    })
    await deleteUser(row.id)
    showToast({ message: '删除成功', type: 'success' })
    fetchUsers()
  } catch {}
}

onMounted(() => {
  fetchUsers()
  fetchDepartments()
})
</script>

<style scoped>
.main-container {
  min-height: 100%;
  background: #f7f8fa;
  padding-top: env(safe-area-inset-top);
  padding-bottom: calc(20px + env(safe-area-inset-bottom));
}

.card {
  margin: 12px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #323233;
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

.user-list {
  padding: 0;
}

.user-card {
  padding: 12px 16px;
  border-bottom: 1px solid #f5f5f5;
}

.user-card:last-child {
  border-bottom: none;
}

.user-card-top {
  display: flex;
  align-items: flex-start;
}

.user-info {
  flex: 1;
}

.user-name {
  font-size: 15px;
  font-weight: 600;
  color: #323233;
  margin-bottom: 4px;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
}

.user-detail {
  font-size: 13px;
  color: #969799;
  margin: 2px 0;
}

.user-card-actions {
  display: flex;
  gap: 8px;
  margin-top: 10px;
  justify-content: flex-end;
}

.popup-content {
  padding-top: 16px;
}

.popup-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 16px 16px;
}

.popup-title {
  font-size: 16px;
  font-weight: 600;
}
</style>
