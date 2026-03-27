<template>
  <div class="login-page">
    <div class="login-header">
      <van-icon name="calendar-o" size="48" color="#1989FA" />
      <h1 class="login-title">排班系统</h1>
      <p class="login-subtitle">移动版</p>
    </div>

    <div class="login-form">
      <van-form @submit="handleLogin">
        <van-cell-group inset>
          <van-field
            v-model="form.student_id"
            name="student_id"
            label="学号"
            placeholder="请输入学号"
            :rules="[{ required: true, message: '请输入学号' }]"
          />
          <van-field
            v-model="form.password"
            type="password"
            name="password"
            label="密码"
            placeholder="请输入密码"
            :rules="[{ required: true, message: '请输入密码' }]"
          />
        </van-cell-group>

        <div class="login-actions">
          <van-button round block type="primary" native-type="submit" :loading="loading">
            登录
          </van-button>
        </div>
      </van-form>

      <div class="login-links">
        <router-link to="/register" class="link">注册账号</router-link>
        <router-link to="/forgot-password" class="link">忘记密码</router-link>
      </div>


    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import { login } from '../api/user'
import { useUserStore } from '../stores/user'

const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)

const form = reactive({
  student_id: '',
  password: ''
})

const handleLogin = async () => {
  loading.value = true
  try {
    const data = await login(form)
    userStore.setToken(data.token)
    if (data.user) {
      userStore.user = data.user
      userStore.checked = true
    }
    await userStore.checkAuth()
    showToast({ message: '登录成功', type: 'success' })
    router.push('/')
  } catch (error) {
    showToast({ message: error.message || '登录失败', type: 'fail' })
  } finally {
    loading.value = false
  }
}

</script>

<style scoped>
.login-page {
  min-height: 100%;
  display: flex;
  flex-direction: column;
  padding-top: env(safe-area-inset-top);
  padding-bottom: env(safe-area-inset-bottom);
  background: var(--color-bg, #F7F8FA);
}

.login-header {
  text-align: center;
  padding: 60px var(--space, 16px) 40px;
}

.login-title {
  font-size: 24px;
  font-weight: var(--font-weight-semibold, 600);
  color: var(--color-text-primary, #323233);
  margin: var(--space-md, 12px) 0 var(--space-xs, 4px);
}

.login-subtitle {
  font-size: var(--font-size-small, 12px);
  color: var(--color-text-tertiary, #969799);
}

.login-form {
  flex: 1;
  padding: 0 var(--space, 16px);
}

.login-actions {
  margin-top: var(--space-xl, 24px);
  padding: 0 var(--space-md, 12px);
}

.login-links {
  display: flex;
  justify-content: space-between;
  padding: var(--space-md, 12px) var(--space-md, 12px) 0;
}

.link {
  font-size: var(--font-size-body, 14px);
  color: var(--color-primary, #1989FA);
  text-decoration: none;
}

</style>
