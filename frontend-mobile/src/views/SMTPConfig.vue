<template>
  <div class="main-container">
    <Navbar title="SMTP 邮件配置" show-back />
    <div class="card">
      <div class="card-header">
        <span class="card-title">
          <van-icon name="envelop-o" /> SMTP 邮件配置
        </span>
      </div>

      <van-form @submit="saveConfig">
        <van-cell-group inset>
          <van-field
            v-model="form.host"
            label="SMTP服务器"
            placeholder="如: smtp.qq.com"
            :rules="[{ required: true, message: '请输入SMTP服务器' }]"
          />
          <van-field
            v-model="form.port"
            label="端口"
            type="digit"
            placeholder="如: 587"
            :rules="[{ required: true, message: '请输入端口' }]"
          />
          <van-field label="加密方式" input-align="right" readonly is-link @click="showEncryptionPicker = true">
            <template #input>
              <span>{{ encryptionLabel }}</span>
            </template>
          </van-field>
          <van-field
            v-model="form.username"
            label="用户名/邮箱"
            placeholder="SMTP用户名或邮箱"
          />
          <van-field
            v-model="form.password"
            label="授权码/密码"
            type="password"
            placeholder="邮箱授权码或密码"
          />
          <van-field
            v-model="form.from"
            label="发件人名称"
            placeholder="显示的发件人名称"
          />
          <van-field
            v-model="form.from_email"
            label="发件人邮箱"
            placeholder="发件人邮箱地址"
          />
        </van-cell-group>

        <div style="padding: 16px;">
          <van-button type="primary" block :loading="saving" native-type="submit">保存配置</van-button>
          <van-button block plain style="margin-top: 10px;" @click="showTestDialog = true">发送测试邮件</van-button>
        </div>
      </van-form>
    </div>

    <!-- 加密方式选择弹窗 -->
    <van-popup v-model:show="showEncryptionPicker" position="bottom" round>
      <van-picker
        :columns="encryptionColumns"
        @confirm="onEncryptionConfirm"
        @cancel="showEncryptionPicker = false"
        :default-index="encryptionColumns.findIndex(c => c.value === encryptionType)"
      />
    </van-popup>

    <!-- 网站域名配置 -->
    <div class="card" style="margin-top: 20px;">
      <div class="card-header">
        <span class="card-title">
          <van-icon name="link-o" /> 网站域名配置
        </span>
      </div>
      <van-form @submit="saveSite">
        <van-cell-group inset>
          <van-field
            v-model="siteForm.domain"
            label="网站域名"
            placeholder="如: https://schedule.example.com"
          />
        </van-cell-group>
        <div class="form-hint" style="padding: 0 16px;">用于生成密码重置邮件中的链接</div>
        <div style="padding: 16px;">
          <van-button type="primary" block :loading="savingSite" native-type="submit">保存域名</van-button>
        </div>
      </van-form>
    </div>

    <!-- 测试邮件 -->
    <van-popup v-model:show="showTestDialog" position="bottom" round :style="{ maxHeight: '50%' }">
      <div class="test-dialog-content">
        <div class="test-dialog-title">发送测试邮件</div>
        <van-field
          v-model="testForm.to"
          label="收件邮箱"
          placeholder="输入测试收件邮箱"
        />
        <div style="padding: 16px;">
          <van-button type="primary" block @click="sendTest" :loading="testing">发送</van-button>
        </div>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { showToast } from 'vant'
import Navbar from '../components/Navbar.vue'
import { getSMTPConfig, saveSMTPConfig, testSMTPConfig, getSiteConfig, saveSiteConfig } from '../api/system'

const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const showTestDialog = ref(false)
const showEncryptionPicker = ref(false)

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

const encryptionColumns = [
  { text: 'TLS', value: 'tls' },
  { text: 'SSL', value: 'ssl' },
  { text: '无', value: 'none' }
]

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

const encryptionLabel = computed(() => {
  const map = { tls: 'TLS', ssl: 'SSL', none: '无' }
  return map[encryptionType.value]
})

const onEncryptionConfirm = ({ selectedValues }) => {
  encryptionType.value = selectedValues[0]
  showEncryptionPicker.value = false
}

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
    showToast({ message: '保存成功', type: 'success' })
  } finally {
    saving.value = false
  }
}

const sendTest = async () => {
  if (!testForm.to) {
    showToast({ message: '请输入收件邮箱', type: 'fail' })
    return
  }
  testing.value = true
  try {
    await testSMTPConfig({ to: testForm.to })
    showToast({ message: '测试邮件已发送', type: 'success' })
    showTestDialog.value = false
  } finally {
    testing.value = false
  }
}

const saveSite = async () => {
  if (!siteForm.domain) {
    showToast({ message: '请输入网站域名', type: 'fail' })
    return
  }
  savingSite.value = true
  try {
    await saveSiteConfig(siteForm)
    showToast({ message: '域名配置已保存', type: 'success' })
  } finally {
    savingSite.value = false
  }
}

onMounted(() => {
  loadConfig()
})
</script>

<style scoped>
.main-container {
  min-height: 100%;
  background: #f7f8fa;
  padding-top: env(safe-area-inset-top);
  padding-bottom: calc(20px + env(safe-area-inset-bottom));
}

.card {
  margin: 12px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #323233;
  display: flex;
  align-items: center;
  gap: 8px;
}

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

.test-dialog-content {
  padding-top: 16px;
}

.test-dialog-title {
  text-align: center;
  font-size: 16px;
  font-weight: 600;
  padding: 16px;
  padding-top: 0;
}
</style>
