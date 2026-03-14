<template>
  <div class="auth-page">
    <div class="auth-box">
      <div class="auth-header">
        <el-icon size="48" color="#409eff"><Calendar /></el-icon>
        <h1>注册账号</h1>
        <p>创建您的排班系统账户</p>
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
        
        <el-form-item prop="name">
          <el-input 
            v-model="form.name" 
            placeholder="姓名"
            size="large"
            :prefix-icon="UserFilled"
          />
        </el-form-item>
        
        <el-form-item prop="email">
          <el-input 
            v-model="form.email" 
            placeholder="邮箱"
            size="large"
            :prefix-icon="Message"
          />
        </el-form-item>
        
        <el-form-item prop="password">
          <el-input 
            v-model="form.password" 
            type="password" 
            placeholder="密码"
            size="large"
            :prefix-icon="Lock"
          />
        </el-form-item>
        
        <el-form-item prop="confirmPassword">
          <el-input 
            v-model="form.confirmPassword" 
            type="password" 
            placeholder="确认密码"
            size="large"
            :prefix-icon="Lock"
          />
        </el-form-item>
        
        <el-form-item prop="department">
          <el-select 
            v-model="form.department" 
            placeholder="请选择所属部门"
            size="large"
            style="width: 100%"
          >
            <el-option label="办公室" value="办公室" />
            <el-option label="竞赛部" value="竞赛部" />
            <el-option label="项目部" value="项目部" />
            <el-option label="科普部" value="科普部" />
          </el-select>
        </el-form-item>
        
        <el-form-item>
          <el-button 
            type="primary" 
            size="large" 
            @click="handleRegister"
            :loading="loading"
            style="width: 100%"
          >
            注册
          </el-button>
        </el-form-item>
      </el-form>
      
      <div class="auth-links">
        <span>已有账号？</span>
        <router-link to="/login">立即登录</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { register } from '../api/user'

const router = useRouter()
const formRef = ref()
const loading = ref(false)

const form = reactive({
  student_id: '',
  name: '',
  email: '',
  password: '',
  confirmPassword: '',
  department: ''
})

const validatePass2 = (rule, value, callback) => {
  if (value !== form.password) {
    callback(new Error('两次输入密码不一致'))
  } else {
    callback()
  }
}

const rules = {
  student_id: [{ required: true, message: '请输入学号', trigger: 'blur' }],
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' }
  ],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }, { min: 6, message: '密码至少6位', trigger: 'blur' }],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validatePass2, trigger: 'blur' }
  ],
  department: [{ required: true, message: '请选择所属部门', trigger: 'change' }]
}

const handleRegister = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await register({
      student_id: form.student_id,
      name: form.name,
      email: form.email,
      password: form.password,
      department: form.department
    })
    ElMessage.success('注册成功，请登录')
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

.auth-links {
  text-align: center;
  font-size: 14px;
  color: #666;
}

.auth-links a {
  color: #409eff;
  text-decoration: none;
  margin-left: 4px;
}
</style>
