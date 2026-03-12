<template>
  <div class="user-manage-page">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <h1 class="logo">排班系统 - 用户管理</h1>
          <div class="nav-menu">
            <el-button @click="router.push('/')">返回首页</el-button>
          </div>
        </div>
      </el-header>

      <el-main class="main-content">
        <div class="page-container">
          <el-card>
            <template #header>
              <div class="card-header">
                <span>用户列表</span>
                <span class="sub-title">共 {{ userList.length }} 位用户</span>
              </div>
            </template>

            <el-table :data="userList" v-loading="loading" style="width: 100%">
              <el-table-column prop="id" label="ID" width="80" />
              <el-table-column prop="studentID" label="学号" width="150" />
              <el-table-column prop="name" label="姓名" width="150" />
              <el-table-column prop="email" label="邮箱" min-width="200" />
              <el-table-column prop="role" label="角色" width="120">
                <template #default="{ row }">
                  <el-tag :type="row.role === 'admin' ? 'danger' : 'info'">
                    {{ row.role === 'admin' ? '管理员' : '普通用户' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="150" fixed="right">
                <template #default="{ row }">
                  <el-button 
                    v-if="row.role !== 'admin'" 
                    type="primary" 
                    size="small"
                    @click="setAdmin(row)"
                  >
                    设为管理员
                  </el-button>
                  <el-button 
                    v-else 
                    type="info" 
                    size="small"
                    @click="setUser(row)"
                  >
                    取消管理员
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getUserList, setUserRole } from '../api/user'

const router = useRouter()
const loading = ref(false)
const userList = ref([])

const fetchData = async () => {
  loading.value = true
  try {
    const data = await getUserList()
    userList.value = data || []
  } catch (error) {
    userList.value = []
  } finally {
    loading.value = false
  }
}

const setAdmin = async (row) => {
  try {
    await ElMessageBox.confirm(`确定将 "${row.name}" 设为管理员吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await setUserRole({
      user_id: row.id,
      role: 'admin'
    })
    
    ElMessage.success('设置成功')
    fetchData()
  } catch (error) {
    if (error !== 'cancel') {
      // 错误已在拦截器处理
    }
  }
}

const setUser = async (row) => {
  try {
    await ElMessageBox.confirm(`确定取消 "${row.name}" 的管理员权限吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await setUserRole({
      user_id: row.id,
      role: 'user'
    })
    
    ElMessage.success('设置成功')
    fetchData()
  } catch (error) {
    if (error !== 'cancel') {
      // 错误已在拦截器处理
    }
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.header {
  background: #fff;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  padding: 0;
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
  padding: 0 20px;
}

.logo {
  margin: 0;
  font-size: 20px;
  color: #409eff;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.sub-title {
  font-size: 14px;
  color: #909399;
}

.main-content {
  min-height: calc(100vh - 60px);
  padding: 20px;
}
</style>
