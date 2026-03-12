<template>
  <div class="auth-page">
    <div class="auth-box">
      <div class="auth-header">
        <el-icon size="48" color="#409eff"><Key /></el-icon>
        <h1>找回密码</h1>
        <p>输入学号，我们将发送重置链接到您的邮箱</p>
      </div>
      
      <el-form :model="form" :rules="rules" ref="formRef" class="auth-form">
        <el-form-item prop="student_id">
          <el-input 
            v-model="form.student_id" 
            placeholder="学号"
            size="large"
            :prefix-icon="User"
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
            发送重置链接
          </el-button>
        </el-form-item>
      </el-form>
      
      <div class="auth-links">
        <router-link to="/login">返回登录</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { forgotPassword } from '../api/user'

const formRef = ref()
const loading = ref(false)

const form = reactive({
  student_id: ''
})

const rules = {
  student_id: [{ required: true, message: '请输入学号', trigger: 'blur' }]
}

const handleSubmit = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await forgotPassword(form)
    ElMessage.success('重置链接已发送到您的邮箱')
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

.auth-links {
  text-align: center;
  font-size: 14px;
}

.auth-links a {
  color: #409eff;
  text-decoration: none;
}
</style>
