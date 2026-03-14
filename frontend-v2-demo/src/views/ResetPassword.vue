<template>
  <div class="auth-page">
    <div class="auth-box">
      <div class="auth-header">
        <el-icon size="48" color="#409eff"><Lock /></el-icon>
        <h1>重置密码</h1>
        <p>请输入您的新密码</p>
      </div>
      
      <el-form :model="form" :rules="rules" ref="formRef" class="auth-form">
        <el-form-item prop="password">
          <el-input 
            v-model="form.password" 
            type="password"
            placeholder="新密码"
            size="large"
            :prefix-icon="Lock"
          />
        </el-form-item>
        
        <el-form-item prop="confirmPassword">
          <el-input 
            v-model="form.confirmPassword" 
            type="password"
            placeholder="确认新密码"
            size="large"
            :prefix-icon="Lock"
          />
        </el-form-item>
        
        <el-form-item>
          <el-button 
            type="primary" 
            size="large" 
            @click="handleSubmit"
            :loading="loading"
            style="width: 100%"
          >
            重置密码
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { resetPassword } from '../api/user'

const route = useRoute()
const router = useRouter()
const formRef = ref()
const loading = ref(false)
const token = ref('')

const form = reactive({
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
  password: [{ required: true, message: '请输入新密码', trigger: 'blur' }, { min: 6, message: '密码至少6位', trigger: 'blur' }],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    { validator: validatePass2, trigger: 'blur' }
  ]
}

onMounted(() => {
  token.value = route.query.token
  if (!token.value) {
    ElMessage.error('无效的链接')
    router.push('/login')
  }
})

const handleSubmit = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await resetPassword({
      token: token.value,
      password: form.password
    })
    ElMessage.success('密码重置成功，请登录')
    router.push('/login')
  } catch {
    // 错误已在拦截器处理
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
  max-width: 400px;
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
