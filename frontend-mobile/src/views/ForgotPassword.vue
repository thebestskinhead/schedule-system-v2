<template>
  <div class="auth-page">
    <div class="auth-header">
      <van-icon name="lock" size="48" color="#1989FA" />
      <h1 class="auth-title">找回密码</h1>
      <p class="auth-subtitle">输入学号，我们将发送重置链接到您的邮箱</p>
    </div>

    <div class="auth-form">
      <van-form @submit="handleSubmit">
        <van-cell-group inset>
          <van-field
            v-model="form.student_id"
            name="student_id"
            label="学号"
            placeholder="请输入学号"
            :rules="[{ required: true, message: '请输入学号' }]"
          />
        </van-cell-group>

        <div class="auth-actions">
          <van-button round block type="primary" native-type="submit" :loading="loading">
            发送重置链接
          </van-button>
        </div>
      </van-form>

      <div class="auth-links">
        <router-link to="/login" class="link">返回登录</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { showToast } from 'vant'
import { forgotPassword } from '../api/user'

const loading = ref(false)

const form = reactive({
  student_id: ''
})

const handleSubmit = async () => {
  loading.value = true
  try {
    await forgotPassword(form)
    showToast({ message: '重置链接已发送到您的邮箱', type: 'success' })
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
  padding: var(--space-md, 12px) var(--space-md, 12px) 0;
}

.link {
  font-size: var(--font-size-body, 14px);
  color: var(--color-primary, #1989FA);
  text-decoration: none;
}
</style>
