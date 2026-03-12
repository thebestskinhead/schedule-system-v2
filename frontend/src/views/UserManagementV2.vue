<template>
  <div class="user-management-v2">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>用户管理 V2</span>
          <el-tag type="info">共 {{ users.length }} 位用户</el-tag>
        </div>
      </template>

      <div class="filters">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="部门筛选">
              <el-checkbox-group v-model="selectedDepts" @change="loadUsers">
                <el-checkbox
                  v-for="dept in departments"
                  :key="dept"
                  :label="dept"
                >
                  {{ dept }}
                </el-checkbox>
              </el-checkbox-group>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-row :gutter="10">
              <el-col :span="10">
                <el-form-item label="系统角色">
                  <el-select v-model="filters.role" @change="loadUsers" clearable placeholder="全部">
                    <el-option label="全部" value="" />
                    <el-option label="管理员" value="admin" />
                    <el-option label="普通用户" value="user" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="10">
                <el-form-item label="部门角色">
                  <el-select v-model="filters.dept_role" @change="loadUsers" clearable placeholder="全部">
                    <el-option label="全部" value="" />
                    <el-option label="部门管理员" value="dept_admin" />
                    <el-option label="部门成员" value="dept_member" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="4">
                <el-button @click="resetFilters" style="margin-top: 32px">
                  <el-icon><Refresh /></el-icon>
                  重置
                </el-button>
              </el-col>
            </el-row>
          </el-col>
        </el-row>
      </div>

      <el-table :data="users" border v-loading="loading" stripe>
        <el-table-column prop="student_id" label="学号" width="120" />
        <el-table-column prop="name" label="姓名" width="100" />
        <el-table-column prop="email" label="邮箱" show-overflow-tooltip />
        <el-table-column label="部门" width="120">
          <template #default="{ row }">
            <el-select
              v-model="row.department"
              @change="updateDepartment(row.id, row.department)"
              size="small"
            >
              <el-option
                v-for="dept in departments"
                :key="dept"
                :label="dept"
                :value="dept"
              />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="部门角色" width="130">
          <template #default="{ row }">
            <el-select
              v-model="row.dept_role"
              @change="updateDeptRole(row.id, row.dept_role)"
              size="small"
            >
              <el-option label="部门管理员" value="dept_admin" />
              <el-option label="部门成员" value="dept_member" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="系统角色" width="100">
          <template #default="{ row }">
            <el-tag :type="row.role === 'admin' ? 'danger' : 'success'" size="small">
              {{ row.role === 'admin' ? '管理员' : '用户' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="showDetail(row)">
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 用户详情对话框 -->
    <el-dialog v-model="detailVisible" title="用户详情" width="500px">
      <el-descriptions :column="1" border v-if="selectedUser">
        <el-descriptions-item label="姓名">{{ selectedUser.name }}</el-descriptions-item>
        <el-descriptions-item label="学号">{{ selectedUser.student_id }}</el-descriptions-item>
        <el-descriptions-item label="邮箱">{{ selectedUser.email }}</el-descriptions-item>
        <el-descriptions-item label="部门">
          <el-tag>{{ selectedUser.department }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="部门角色">
          <el-tag :type="selectedUser.dept_role === 'dept_admin' ? 'warning' : 'info'">
            {{ selectedUser.dept_role === 'dept_admin' ? '部门管理员' : '部门成员' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="系统角色">
          <el-tag :type="selectedUser.role === 'admin' ? 'danger' : 'success'">
            {{ selectedUser.role === 'admin' ? '管理员' : '普通用户' }}
          </el-tag>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { userAdminAPI } from '../api/userAdmin.js'

const departments = ref([])
const users = ref([])
const selectedDepts = ref([])
const loading = ref(false)
const detailVisible = ref(false)
const selectedUser = ref(null)

const filters = reactive({
  role: '',
  dept_role: ''
})

const loadDepartments = async () => {
  try {
    const res = await userAdminAPI.getDepartments()
    departments.value = res?.departments || []
  } catch (err) {
    console.error('加载部门失败:', err)
  }
}

const loadUsers = async () => {
  loading.value = true
  try {
    const params = {}
    if (selectedDepts.value.length > 0) {
      params.departments = selectedDepts.value.join(',')
    }
    if (filters.role) {
      params.role = filters.role
    }
    if (filters.dept_role) {
      params.dept_role = filters.dept_role
    }

    const res = await userAdminAPI.listByFilter(params)
    users.value = res?.users || []
  } catch (err) {
    console.error('加载用户失败:', err)
    ElMessage.error('加载用户失败')
  } finally {
    loading.value = false
  }
}

const resetFilters = () => {
  selectedDepts.value = []
  filters.role = ''
  filters.dept_role = ''
  loadUsers()
}

const updateDepartment = async (userId, department) => {
  try {
    await userAdminAPI.setDepartment(userId, department)
    ElMessage.success('部门更新成功！')
  } catch (err) {
    ElMessage.error('更新失败: ' + (err.response?.data?.message || err.message))
    loadUsers()
  }
}

const updateDeptRole = async (userId, deptRole) => {
  try {
    await userAdminAPI.setDeptRole(userId, deptRole)
    ElMessage.success('部门角色更新成功！')
  } catch (err) {
    ElMessage.error('更新失败: ' + (err.response?.data?.message || err.message))
    loadUsers()
  }
}

const showDetail = (user) => {
  selectedUser.value = user
  detailVisible.value = true
}

onMounted(() => {
  loadDepartments()
  loadUsers()
})
</script>

<style scoped>
.user-management-v2 {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 18px;
  font-weight: bold;
}

.filters {
  margin-bottom: 20px;
  padding: 15px;
  background: #f5f7fa;
  border-radius: 4px;
}
</style>
