<template>
  <div class="forgot-password-container">
    <el-card class="forgot-password-box">
      <template #header>
        <h2 class="title">找回密码</h2>
      </template>

      <el-steps :active="step" finish-status="success" class="steps">
        <el-step title="输入邮箱" />
        <el-step title="发送邮件" />
      </el-steps>

      <!-- 步骤1: 输入邮箱 -->
      <div v-if="step === 0">
        <el-form :model="form" :rules="rules" ref="formRef" label-position="top">
          <el-form-item label="注册邮箱" prop="email">
            <el-input
              v-model="form.email"
              placeholder="请输入注册时使用的邮箱"
              prefix-icon="Message"
              size="large"
            />
          </el-form-item>

          <el-form-item>
            <el-button
              type="primary"
              size="large"
              @click="submitRequest"
              :loading="loading"
              style="width: 100%"
            >
              发送重置邮件
            </el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 步骤2: 发送成功 -->
      <div v-if="step === 1" class="success-message">
        <el-result
          icon="success"
          title="邮件已发送"
          :sub-title="`重置密码邮件已发送至 ${form.email}，请查收邮件并按照提示操作。`"
        >
          <template #extra>
            <el-button type="primary" @click="goToLogin">返回登录</el-button>
          </template>
        </el-result>
      </div>

      <div class="footer">
        <router-link to="/login">返回登录</router-link>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { requestPasswordReset } from '../api/smtp'

const router = useRouter()
const formRef = ref()
const loading = ref(false)
const step = ref(0)

const form = reactive({
  email: ''
})

const rules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' }
  ]
}

const submitRequest = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await requestPasswordReset(form.email)
    step.value = 1
    ElMessage.success('重置邮件已发送')
  } catch (error) {
    // 错误已在拦截器处理
  } finally {
    loading.value = false
  }
}

const goToLogin = () => {
  router.push('/login')
}
</script>

<style scoped>
.forgot-password-container {
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.forgot-password-box {
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

.success-message {
  padding: 20px 0;
}

.footer {
  text-align: center;
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #e4e7ed;
}

.footer a {
  color: #667eea;
  text-decoration: none;
}

.footer a:hover {
  text-decoration: underline;
}
</style>
