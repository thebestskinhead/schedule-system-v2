<template>
  <div class="main-container">
    <div class="card">
      <div class="card-header">
        <span class="card-title">
          <el-icon><Message /></el-icon> SMTP 邮件配置
        </span>
      </div>

      <el-form :model="form" label-width="120px">
        <el-form-item label="SMTP服务器">
          <el-input v-model="form.host" placeholder="如: smtp.qq.com" />
        </el-form-item>
        <el-form-item label="端口">
          <el-input-number v-model="form.port" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="加密方式">
          <el-radio-group v-model="encryptionType">
            <el-radio-button value="tls">TLS</el-radio-button>
            <el-radio-button value="ssl">SSL</el-radio-button>
            <el-radio-button value="none">无</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="用户名/邮箱">
          <el-input v-model="form.username" placeholder="SMTP用户名或邮箱" />
        </el-form-item>
        <el-form-item label="授权码/密码">
          <el-input v-model="form.password" type="password" placeholder="邮箱授权码或密码" show-password />
        </el-form-item>
        <el-form-item label="发件人名称">
          <el-input v-model="form.from" placeholder="显示的发件人名称" />
        </el-form-item>
        <el-form-item label="发件人邮箱">
          <el-input v-model="form.from_email" placeholder="发件人邮箱地址" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="saveConfig" :loading="saving">保存配置</el-button>
          <el-button @click="showTestDialog = true">发送测试邮件</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 网站域名配置 -->
    <div class="card" style="margin-top: 20px;">
      <div class="card-header">
        <span class="card-title">
          <el-icon><Link /></el-icon> 网站域名配置
        </span>
      </div>
      <el-form :model="siteForm" label-width="120px">
        <el-form-item label="网站域名">
          <el-input v-model="siteForm.domain" placeholder="如: https://schedule.example.com" style="width: 400px">
            <template #prefix>
              <el-icon><Link /></el-icon>
            </template>
          </el-input>
          <div class="form-hint">用于生成密码重置邮件中的链接</div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="saveSite" :loading="savingSite">保存域名</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 测试邮件 -->
    <el-dialog v-model="showTestDialog" title="发送测试邮件" width="400px">
      <el-form :model="testForm" label-width="80px">
        <el-form-item label="收件邮箱">
          <el-input v-model="testForm.to" placeholder="输入测试收件邮箱" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showTestDialog = false">取消</el-button>
        <el-button type="primary" @click="sendTest" :loading="testing">发送</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { getSMTPConfig, saveSMTPConfig, testSMTPConfig, getSiteConfig, saveSiteConfig } from '../api/system'

const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const showTestDialog = ref(false)

const form = reactive({
  host: '',
  port: 587,
  use_tls: true,
  use_ssl: false,
  username: '',
  password: '',
  from: '排班系统',
  from_email: ''
})

// 加密方式转换
const encryptionType = computed({
  get: () => {
    if (form.use_ssl) return 'ssl'
    if (form.use_tls) return 'tls'
    return 'none'
  },
  set: (val) => {
    form.use_ssl = val === 'ssl'
    form.use_tls = val === 'tls'
  }
})

const testForm = reactive({
  to: ''
})

const siteForm = reactive({
  domain: ''
})
const savingSite = ref(false)

const loadConfig = async () => {
  loading.value = true
  try {
    // 加载SMTP配置
    const data = await getSMTPConfig()
    if (data) {
      Object.assign(form, {
        host: data.host || '',
        port: data.port || 587,
        use_tls: data.use_tls !== undefined ? data.use_tls : true,
        use_ssl: data.use_ssl || false,
        username: data.username || '',
        password: data.password || '',
        from: data.from || '排班系统',
        from_email: data.from_email || ''
      })
    }
    // 加载网站域名配置
    const siteData = await getSiteConfig()
    if (siteData) {
      siteForm.domain = siteData.domain || ''
    }
  } finally {
    loading.value = false
  }
}

const saveConfig = async () => {
  saving.value = true
  try {
    await saveSMTPConfig(form)
    ElMessage.success('保存成功')
  } finally {
    saving.value = false
  }
}

const sendTest = async () => {
  if (!testForm.to) {
    ElMessage.warning('请输入收件邮箱')
    return
  }
  testing.value = true
  try {
    await testSMTPConfig({ to: testForm.to })
    ElMessage.success('测试邮件已发送')
    showTestDialog.value = false
  } finally {
    testing.value = false
  }
}

const saveSite = async () => {
  if (!siteForm.domain) {
    ElMessage.warning('请输入网站域名')
    return
  }
  savingSite.value = true
  try {
    await saveSiteConfig(siteForm)
    ElMessage.success('域名配置已保存')
  } finally {
    savingSite.value = false
  }
}

onMounted(() => {
  loadConfig()
})
</script>

<style scoped>
.form-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
