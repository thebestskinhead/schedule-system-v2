<template>
  <div class="auth-page">
    <div class="auth-header">
      <van-icon name="calendar-o" size="48" color="#1989FA" />
      <h1 class="auth-title">注册账号</h1>
      <p class="auth-subtitle">创建您的排班系统账户</p>
    </div>

    <div class="auth-form">
      <van-form @submit="handleRegister">
        <van-cell-group inset>
          <van-field
            v-model="form.student_id"
            name="student_id"
            label="学号"
            placeholder="请输入学号"
            :rules="[{ required: true, message: '请输入学号' }]"
          />
          <van-field
            v-model="form.name"
            name="name"
            label="姓名"
            placeholder="请输入姓名"
            :rules="[{ required: true, message: '请输入姓名' }]"
          />
          <van-field
            v-model="form.email"
            name="email"
            label="邮箱"
            placeholder="请输入邮箱"
            :rules="[
              { required: true, message: '请输入邮箱' },
              { pattern: /^[\w.-]+@[\w.-]+\.\w+$/, message: '邮箱格式不正确' }
            ]"
          />
          <van-field
            v-model="form.password"
            type="password"
            name="password"
            label="密码"
            placeholder="请输入密码"
            :rules="[
              { required: true, message: '请输入密码' },
              { validator: val => val.length >= 6, message: '密码至少6位' }
            ]"
          />
          <van-field
            v-model="form.confirmPassword"
            type="password"
            name="confirmPassword"
            label="确认密码"
            placeholder="请再次输入密码"
            :rules="[
              { required: true, message: '请确认密码' },
              { validator: validateConfirmPassword, message: '两次输入密码不一致' }
            ]"
          />
          <van-field
            v-model="departmentText"
            is-link
            readonly
            name="department"
            label="部门"
            placeholder="请选择所属部门"
            :rules="[{ required: true, message: '请选择所属部门' }]"
            @click="showDepartmentPicker = true"
          />
        </van-cell-group>

        <div class="auth-actions">
          <van-button round block type="primary" native-type="submit" :loading="loading">
            注册
          </van-button>
        </div>
      </van-form>

      <div class="auth-links">
        <span class="link-text">已有账号？</span>
        <router-link to="/login" class="link">立即登录</router-link>
      </div>
    </div>

    <!-- 部门选择弹出层 -->
    <van-popup v-model:show="showDepartmentPicker" position="bottom" round>
      <van-picker
        :columns="departmentColumns"
        @confirm="onDepartmentConfirm"
        @cancel="showDepartmentPicker = false"
      />
    </van-popup>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { register } from '../api/user'

const router = useRouter()
const loading = ref(false)
const showDepartmentPicker = ref(false)

const departmentColumns = [
  { text: '办公室', value: '办公室' },
  { text: '竞赛部', value: '竞赛部' },
  { text: '项目部', value: '项目部' },
  { text: '科普部', value: '科普部' }
]

const form = reactive({
  student_id: '',
  name: '',
  email: '',
  password: '',
  confirmPassword: '',
  department: ''
})

const departmentText = ref('')

const validateConfirmPassword = (val) => {
  return val === form.password
}

const onDepartmentConfirm = ({ selectedValues }) => {
  form.department = selectedValues[0]
  departmentText.value = form.department
  showDepartmentPicker.value = false
}

const handleRegister = async () => {
  loading.value = true
  try {
    await register({
      student_id: form.student_id,
      name: form.name,
      email: form.email,
      password: form.password,
      department: form.department
    })
    showToast({ message: '注册成功，请登录', type: 'success' })
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
  min-height: 100%;
  display: flex;
  flex-direction: column;
  padding-top: env(safe-area-inset-top);
  padding-bottom: env(safe-area-inset-bottom);
  background: var(--color-bg, #F7F8FA);
}

.auth-header {
  text-align: center;
  padding: 60px var(--space, 16px) 40px;
}

.auth-title {
  font-size: 24px;
  font-weight: var(--font-weight-semibold, 600);
  color: var(--color-text-primary, #323233);
  margin: var(--space-md, 12px) 0 var(--space-xs, 4px);
}

.auth-subtitle {
  font-size: var(--font-size-small, 12px);
  color: var(--color-text-tertiary, #969799);
}

.auth-form {
  flex: 1;
  padding: 0 var(--space, 16px);
}

.auth-actions {
  margin-top: var(--space-xl, 24px);
  padding: 0 var(--space-md, 12px);
}

.auth-links {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: var(--space-md, 12px) var(--space-md, 12px) 0;
}

.link-text {
  font-size: var(--font-size-body, 14px);
  color: var(--color-text-secondary, #969799);
}

.link {
  font-size: var(--font-size-body, 14px);
  color: var(--color-primary, #1989FA);
  text-decoration: none;
  margin-left: 4px;
}
</style>
