<template>
  <div class="reset-password-container">
    <el-card class="reset-password-box">
      <template #header>
        <h2 class="title">重置密码</h2>
      </template>

      <el-steps :active="step" finish-status="success" class="steps">
        <el-step title="验证身份" />
        <el-step title="设置新密码" />
        <el-step title="完成" />
      </el-steps>

      <!-- 步骤1: 验证中 -->
      <div v-if="step === 0" class="loading-state">
        <el-skeleton :rows="3" animated />
        <p class="loading-text">正在验证链接有效性...</p>
      </div>

      <!-- 步骤2: 设置新密码 -->
      <div v-if="step === 1">
        <el-alert
          title="请设置新密码"
          description="密码长度至少6位，建议包含字母和数字"
          type="info"
          :closable="false"
          class="mb-4"
        />

        <el-form :model="form" :rules="rules" ref="formRef" label-position="top">
          <el-form-item label="新密码" prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="请输入新密码"
              prefix-icon="Lock"
              size="large"
              show-password
            />
          </el-form-item>

          <el-form-item label="确认密码" prop="confirmPassword">
            <el-input
              v-model="form.confirmPassword"
              type="password"
              placeholder="请再次输入新密码"
              prefix-icon="Lock"
              size="large"
              show-password
            />
          </el-form-item>

          <el-form-item>
            <el-button
              type="primary"
              size="large"
              @click="submitReset"
              :loading="loading"
              style="width: 100%"
            >
              重置密码
            </el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 步骤3: 完成 -->
      <div v-if="step === 2" class="success-message">
        <el-result
          icon="success"
          title="密码重置成功"
          sub-title="您可以使用新密码登录系统了"
        >
          <template #extra>
            <el-button type="primary" @click="goToLogin">去登录</el-button>
          </template>
        </el-result>
      </div>

      <!-- 错误状态 -->
      <div v-if="step === -1" class="error-message">
        <el-result
          icon="error"
          title="链接无效或已过期"
          sub-title="重置密码链接已过期或无效，请重新申请"
        >
          <template #extra>
            <el-button type="primary" @click="goToForgot">重新申请</el-button>
          </template>
        </el-result>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { verifyResetToken, resetPassword } from '../api/smtp'

const route = useRoute()
const router = useRouter()
const formRef = ref()
const loading = ref(false)
const step = ref(0)
const token = ref('')

const form = reactive({
  password: '',
  confirmPassword: ''
})

const validateConfirmPassword = (rule, value, callback) => {
  if (value !== form.password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules = {
  password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

onMounted(async () => {
  const t = route.query.token
  if (!t) {
    step.value = -1
    return
  }

  token.value = t

  try {
    await verifyResetToken(t)
    step.value = 1
  } catch (error) {
    step.value = -1
  }
})

const submitReset = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await resetPassword(token.value, form.password)
    step.value = 2
    ElMessage.success('密码重置成功')
  } catch (error) {
    // 错误已在拦截器处理
  } finally {
    loading.value = false
  }
}

const goToLogin = () => {
  router.push('/login')
}

const goToForgot = () => {
  router.push('/forgot-password')
}
</script>

<style scoped>
.reset-password-container {
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.reset-password-box {
  width: 450px;
}

.title {
  text-align: center;
  color: #333;
  margin: 0;
}

.steps {
  margin: 30px 0;
}

.loading-state {
  padding: 20px;
}

.loading-text {
  text-align: center;
  color: #909399;
  margin-top: 20px;
}

.success-message,
.error-message {
  padding: 20px 0;
}

.mb-4 {
  margin-bottom: 16px;
}
</style>
