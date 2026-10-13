<template>
  <div class="main-container">
    <Navbar :title="isGlobalAdmin ? '用户管理' : '部门用户管理'" show-back />
    <div class="card">
      <div class="card-header">
        <span class="card-title">
          {{ isGlobalAdmin ? '用户管理' : '部门用户管理' }}
          <van-tag v-if="!isGlobalAdmin" size="medium" type="default" style="margin-left: 8px">{{ userStore.user?.department }}</van-tag>
        </span>
        <van-button v-if="isGlobalAdmin" type="primary" size="small" @click="openCreateDialog">录入正式成员</van-button>
      </div>

      <!-- 状态筛选（三态） -->
      <div class="filter-bar">
        <span class="filter-chip" :class="{ active: statusFilter === 'all' }" @click="statusFilter = 'all'">全部</span>
        <span class="filter-chip" :class="{ active: statusFilter === 'pending' }" @click="statusFilter = 'pending'">待入职</span>
        <span class="filter-chip" :class="{ active: statusFilter === 'active' }" @click="statusFilter = 'active'">在职</span>
        <span class="filter-chip" :class="{ active: statusFilter === 'resigned' }" @click="statusFilter = 'resigned'">离职</span>
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
                  <van-tag :type="statusTag(row.employment_status).type" size="medium" style="margin-left: 6px">
                    {{ statusTag(row.employment_status).text }}
                  </van-tag>
                </div>
                <div class="user-detail">学号：{{ row.student_id }}</div>
                <div class="user-detail">邮箱：{{ row.email }}</div>
                <div class="user-detail">
                  <van-tag v-if="row.department" plain size="medium">{{ row.department }}</van-tag>
                </div>
                <div v-if="row.department === '办公室' && row.dept_role === 'dept_admin'" class="office-admin-badge">
                  <van-tag type="success" size="medium">办公室管理员</van-tag>
                </div>
                <div v-if="row.employment_status === 'pending'" class="office-admin-badge">
                  <van-tag type="warning" size="medium">待激活</van-tag>
                </div>
              </div>
            </div>
            <div class="user-card-actions" v-if="!isSelf(row)">
              <van-button size="small" type="primary" plain @click="editUser(row)" v-if="canManageDept">编辑</van-button>
              <van-button
                v-if="row.employment_status === 'pending' && isGlobalAdmin"
                size="small" type="primary" plain @click="resendInviteById(row)"
              >重发邀请</van-button>
              <van-button
                v-if="row.employment_status === 'pending' && isGlobalAdmin"
                size="small" type="warning" plain @click="revokeInviteById(row)"
              >作废</van-button>
              <van-button v-if="canResignUser(row)" size="small" type="danger" plain @click="resignUser(row)">离职</van-button>
            </div>
          </div>
        </template>
      </div>
    </div>

    <!-- 录入/编辑弹窗 -->
    <van-popup
      v-model:show="dialogVisible"
      position="bottom"
      round
      :style="{ maxHeight: '85%' }"
    >
      <div class="popup-content">
        <div class="popup-header">
          <span class="popup-title">{{ isEdit ? '编辑用户' : '录入正式成员' }}</span>
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
              placeholder="请输入邮箱（用于接收激活邮件）"
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
            <van-cell v-if="isEdit && isGlobalAdmin" title="系统角色">
              <template #value>
                <van-radio-group v-model="form.role" direction="horizontal">
                  <van-radio name="user">普通用户</van-radio>
                  <van-radio name="admin">管理员</van-radio>
                </van-radio-group>
              </template>
            </van-cell>
            <van-cell v-if="isGlobalAdmin" title="部门角色">
              <template #value>
                <van-radio-group v-model="form.isDeptAdmin" direction="horizontal">
                  <van-radio :name="false">部门成员</van-radio>
                  <van-radio :name="true">部门管理员</van-radio>
                </van-radio-group>
              </template>
            </van-cell>
          </van-cell-group>
          <div class="role-hint" style="padding: 0 16px;" v-if="isGlobalAdmin && !isEdit">
            提示：录入后系统将发送激活邮件，由成员自行设置密码；办公室管理员请选择部门为"办公室"并设为部门管理员
          </div>
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
import {
  getUserList,
  getUserListByDepartment,
  updateUser,
  createMember,
  resendInvite,
  revokeInvite,
  resignMember,
  getDepartments
} from '../api/user'
import { useUserStore } from '../stores/user'

const userStore = useUserStore()

const isGlobalAdmin = computed(() => userStore.canManageAll)
const canManageDept = computed(() => userStore.canManageDept)

const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const statusFilter = ref('all')
const allUserList = ref([])
const departments = ref([])
const showDeptPicker = ref(false)

const form = reactive({
  id: null,
  student_id: '',
  name: '',
  email: '',
  department: '',
  role: 'user',
  isDeptAdmin: false
})

const deptColumns = computed(() =>
  departments.value.map(dept => ({ text: dept, value: dept }))
)

const onDeptConfirm = ({ selectedValues }) => {
  form.department = selectedValues[0]
  showDeptPicker.value = false
}

const userList = computed(() => {
  if (statusFilter.value === 'all') return allUserList.value
  return allUserList.value.filter(u => u.employment_status === statusFilter.value)
})

const statusTag = (s) => {
  const map = {
    pending: { text: '待入职', type: 'warning' },
    active: { text: '在职', type: 'success' },
    resigned: { text: '离职', type: 'default' }
  }
  return map[s] || { text: s || '未知', type: 'default' }
}

const isSelf = (row) => row.id === userStore.user?.id
const canResignUser = (row) =>
  isGlobalAdmin.value && !isSelf(row) && row.employment_status !== 'resigned'

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
    allUserList.value = data || []
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

const openCreateDialog = () => {
  isEdit.value = false
  Object.assign(form, {
    id: null,
    student_id: '',
    name: '',
    email: '',
    department: userStore.user?.department || '',
    role: 'user',
    isDeptAdmin: false
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
    isDeptAdmin: row.dept_role === 'dept_admin'
  })
  dialogVisible.value = true
}

const submitForm = async () => {
  submitting.value = true
  try {
    if (isEdit.value) {
      const data = { name: form.name, email: form.email }
      if (isGlobalAdmin.value) {
        data.department = form.department
        data.role = form.role
        data.dept_role = form.isDeptAdmin ? 'dept_admin' : 'dept_member'
      }
      await updateUser(form.id, data)
      showToast({ message: '更新成功', type: 'success' })
    } else {
      await createMember({
        student_id: form.student_id,
        name: form.name,
        email: form.email,
        department: form.department,
        dept_role: form.isDeptAdmin ? 'dept_admin' : 'dept_member'
      })
      showToast({ message: '已录入并发送激活邮件', type: 'success' })
    }
    dialogVisible.value = false
    fetchUsers()
  } catch {
    // 错误已由拦截器处理
  } finally {
    submitting.value = false
  }
}

const resendInviteById = async (row) => {
  try {
    await showConfirmDialog({
      title: '重发邀请',
      message: `确认向 ${row.name}（${row.email}）重新发送激活邮件？`
    })
  } catch {
    return
  }
  try {
    await resendInvite(row.id)
    showToast({ message: '已重新发送邀请', type: 'success' })
  } catch {}
}

const revokeInviteById = async (row) => {
  try {
    await showConfirmDialog({
      title: '作废邀请',
      message: `确认作废 ${row.name} 的邀请？作废后该学号需重新录入。`
    })
  } catch {
    return
  }
  try {
    await revokeInvite(row.id)
    showToast({ message: '邀请已作废', type: 'success' })
    fetchUsers()
  } catch {}
}

const resignUser = async (row) => {
  try {
    await showConfirmDialog({
      title: '离职',
      message: '标记离职后该成员不可登录且不可再次注册，确定？'
    })
  } catch {
    return
  }
  try {
    await resignMember(row.id)
    showToast({ message: '已标记离职（批量离职请在电脑端操作）', type: 'success' })
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
  padding: env(safe-area-inset-top) 0 calc(var(--space-xl, 24px) + env(safe-area-inset-bottom));
  background: var(--color-bg, #F7F8FA);
  min-height: 100%;
}

.card {
  margin: 12px;
  background: #fff;
  border-radius: 12px;
  padding: 16px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #323233;
}

.filter-bar {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.filter-chip {
  padding: 5px 14px;
  border-radius: 16px;
  font-size: 13px;
  background: #f2f3f5;
  color: #646566;
  cursor: pointer;
  transition: all 0.2s;
}

.filter-chip.active {
  background: #1989fa;
  color: #fff;
}

.user-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.user-card {
  border: 1px solid #f0f0f0;
  border-radius: 10px;
  padding: 12px;
}

.user-card-top {
  display: flex;
  justify-content: space-between;
}

.user-info { flex: 1; }

.user-name {
  font-size: 15px;
  font-weight: 600;
  color: #323233;
  margin-bottom: 6px;
}

.user-detail {
  font-size: 12px;
  color: #969799;
  margin-top: 2px;
}

.office-admin-badge {
  margin-top: 6px;
}

.user-card-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-end;
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid #f5f5f5;
}

.role-hint {
  font-size: 12px;
  color: #969799;
  line-height: 1.5;
}
</style>
