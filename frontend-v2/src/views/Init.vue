<template>
  <div class="auth-page">
    <div class="auth-box" style="max-width: 500px;">
      <div class="auth-header">
        <el-icon size="48" color="#409eff"><Setting /></el-icon>
        <h1>系统初始化</h1>
        <p>创建系统管理员账号完成初始化</p>
      </div>

      <el-form :model="form" :rules="rules" ref="formRef" class="auth-form">
        <el-form-item prop="student_id">
          <el-input v-model="form.student_id" placeholder="学号" size="large" :prefix-icon="User" />
        </el-form-item>
        <el-form-item prop="name">
          <el-input v-model="form.name" placeholder="姓名" size="large" :prefix-icon="UserFilled" />
        </el-form-item>
        <el-form-item prop="email">
          <el-input v-model="form.email" placeholder="邮箱" size="large" :prefix-icon="Message" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="密码" size="large" :prefix-icon="Lock" />
        </el-form-item>
        <el-form-item prop="confirmPassword">
          <el-input v-model="form.confirmPassword" type="password" placeholder="确认密码" size="large" :prefix-icon="Lock" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" size="large" @click="handleInit" :loading="loading" style="width: 100%">
            初始化系统
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { initSystem, getInstallStatus } from '../api/system'

const router = useRouter()
const loading = ref(false)
const formRef = ref()

const form = reactive({
  student_id: '',
  name: '',
  email: '',
  password: '',
  confirmPassword: ''
})

const validatePass2 = (rule, value, callback) => {
  if (value !== form.password) {
    callback(new Error('两次输入密码不一致'))
  } else {
    callback()
  }
}

const rules = {
  student_id: [{ required: true, message: '请输入学号' }],
  name: [{ required: true, message: '请输入姓名' }],
  email: [{ required: true, message: '请输入邮箱' }, { type: 'email', message: '邮箱格式不正确' }],
  password: [{ required: true, message: '请输入密码' }, { min: 6, message: '密码至少6位' }],
  confirmPassword: [{ required: true, message: '请确认密码' }, { validator: validatePass2 }]
}

onMounted(async () => {
  try {
    const res = await getInstallStatus()
    if (res.data?.data?.installed) {
      ElMessage.info('系统已初始化')
      router.push('/login')
    }
  } catch {}
})

const handleInit = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await initSystem({
      student_id: form.student_id,
      name: form.name,
      email: form.email,
      password: form.password
    })
    ElMessage.success('系统初始化成功，请登录')
    router.push('/login')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.auth-box {
  background: #fff;
  border-radius: 16px;
  padding: 40px;
  width: 100%;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.2);
}

.auth-header {
  text-align: center;
  margin-bottom: 32px;
}

.auth-header h1 {
  font-size: 24px;
  margin: 16px 0 8px;
  color: #333;
}

.auth-header p {
  color: #999;
  font-size: 14px;
}

.auth-form {
  margin-bottom: 24px;
}
</style>
