<template>
  <div class="auth-page">
    <div class="auth-header">
      <van-icon name="shield-o" size="48" color="#1989FA" />
      <h1 class="auth-title">重置密码</h1>
      <p class="auth-subtitle">请输入您的新密码</p>
    </div>

    <div class="auth-form">
      <van-form @submit="handleSubmit">
        <van-cell-group inset>
          <van-field
            v-model="form.password"
            type="password"
            name="password"
            label="新密码"
            placeholder="请输入新密码"
            :rules="[
              { required: true, message: '请输入新密码' },
              { validator: val => val.length >= 6, message: '密码至少6位' }
            ]"
          />
          <van-field
            v-model="form.confirmPassword"
            type="password"
            name="confirmPassword"
            label="确认密码"
            placeholder="请再次输入新密码"
            :rules="[
              { required: true, message: '请确认新密码' },
              { validator: validateConfirmPassword, message: '两次输入密码不一致' }
            ]"
          />
        </van-cell-group>

        <div class="auth-actions">
          <van-button round block type="primary" native-type="submit" :loading="loading">
            重置密码
          </van-button>
        </div>
      </van-form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast } from 'vant'
import { resetPassword } from '../api/user'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const token = ref('')

const form = reactive({
  password: '',
  confirmPassword: ''
})

const validateConfirmPassword = (val) => {
  return val === form.password
}

onMounted(() => {
  token.value = route.query.token
  if (!token.value) {
    showToast({ message: '无效的链接', type: 'fail' })
    router.push('/login')
  }
})

const handleSubmit = async () => {
  loading.value = true
  try {
    await resetPassword({
      token: token.value,
      password: form.password
    })
    showToast({ message: '密码重置成功，请登录', type: 'success' })
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
</style>
