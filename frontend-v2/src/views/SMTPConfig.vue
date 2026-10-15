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

    <!-- 系统更新（热更新） -->
    <div class="card" style="margin-top: 20px;">
      <div class="card-header">
        <span class="card-title">
          <el-icon><Refresh /></el-icon> 系统更新
        </span>
      </div>

      <el-form label-width="120px">
        <el-form-item label="当前版本">
          <span class="ver-text">{{ versionInfo.version || '未知' }}</span>
          <span class="form-hint" style="margin-left: 12px;">
            {{ versionInfo.platform }}{{ versionInfo.updaterPresent ? '' : ' · 未检测到更新程序' }}
          </span>
        </el-form-item>
        <el-form-item label="最新版本">
          <template v-if="updateInfo">
            <span class="ver-text" :class="{ 'ver-new': updateInfo.hasUpdate && updateInfo.sameMajor }">
              {{ updateInfo.target }}
            </span>
            <el-tag v-if="canHotUpdate" type="success" size="small" style="margin-left: 8px;">
              可热更新{{ updateInfo.assetSize ? '（' + formatSize(updateInfo.assetSize) + '）' : '' }}
            </el-tag>
            <el-tag v-else-if="isCrossMajor" type="warning" size="small" style="margin-left: 8px;">跨大版本</el-tag>
            <el-tag v-else type="info" size="small" style="margin-left: 8px;">已是最新</el-tag>
          </template>
          <span v-else class="form-hint">点击「检查更新」获取最新版本信息</span>
        </el-form-item>
        <el-form-item label="更新说明" v-if="updateInfo && updateInfo.releaseNotes">
          <div class="release-notes">{{ updateInfo.releaseNotes }}</div>
        </el-form-item>
        <el-form-item label="上次更新" v-if="lastResult && lastResult.hasResult">
          <el-tag :type="lastResult.phase === 'done' ? 'success' : (lastResult.phase === 'rolled-back' ? 'info' : 'danger')" size="small">
            {{ lastResult.phase === 'done' ? '成功' : (lastResult.phase === 'rolled-back' ? '已终止' : '失败') }}
          </el-tag>
          <span class="form-hint" style="margin-left: 8px;">{{ lastResult.message }}（{{ lastResult.at }}）</span>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="doCheckUpdate" :loading="checking">检查更新</el-button>
          <el-button
            v-if="canHotUpdate"
            type="success"
            @click="doStartUpdate"
            :loading="updating"
          >立即更新</el-button>
          <el-button v-else-if="isCrossMajor" @click="openReleasePage">前往下载整包</el-button>
        </el-form-item>
        <div v-if="isCrossMajor" class="update-warning">
          新版本 {{ updateInfo.target }} 与当前版本（{{ updateInfo.current }}）跨大版本，可能包含数据库结构变更，无法热更新。
          请手动下载整包，并按发布说明迁移数据库后再升级。
        </div>
        <div class="form-hint">
          热更新仅覆盖服务端程序与前后端构建产物，不影响数据库与配置；更新期间服务将短暂中断（约 1 分钟），可在更新进度页监控或随时终止恢复旧版本。
        </div>
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
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getSMTPConfig, saveSMTPConfig, testSMTPConfig, getSiteConfig, saveSiteConfig,
  getVersion, checkUpdate, startUpdate, getUpdateStatus
} from '../api/system'

const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const showTestDialog = ref(false)

const form = reactive({
  id: null,
  host: '',
  port: 465,
  use_tls: false,
  use_ssl: true,
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

// 端口与加密方式联动：465 为隐式 SSL，587 为 STARTTLS
watch(() => form.port, (port) => {
  if (port === 465) {
    form.use_ssl = true
    form.use_tls = false
  } else if (port === 587) {
    form.use_ssl = false
    form.use_tls = true
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
        id: data.id || null,
        host: data.host || '',
        port: data.port || 465,
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

const EMAIL_RE = /^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$/

const sendTest = async () => {
  const to = (testForm.to || '').trim()
  if (!to) {
    ElMessage.warning('请输入收件邮箱')
    return
  }
  if (!EMAIL_RE.test(to)) {
    ElMessage.warning('邮箱格式不正确，请检查是否包含空格或缺少后缀')
    return
  }
  testForm.to = to
  testing.value = true
  try {
    await testSMTPConfig({ to, ...form })
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

// ---------- 系统更新 ----------
const versionInfo = ref({})
const updateInfo = ref(null)
const lastResult = ref(null)
const checking = ref(false)
const updating = ref(false)

const formatSize = (bytes) => {
  if (!bytes) return ''
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1048576).toFixed(2) + ' MB'
}

// 可热更新：同大版本且存在更高版本
const canHotUpdate = computed(() => {
  const info = updateInfo.value
  return !!info && info.hasUpdate === true && info.sameMajor === true
})

// 跨大版本：有更高大版本（无论是否找到可热更新的同大版本版本）
const isCrossMajor = computed(() => {
  const info = updateInfo.value
  if (!info) return false
  return info.crossMajor === true || (info.hasUpdate === true && info.sameMajor === false)
})

const loadVersion = async () => {
  try {
    versionInfo.value = await getVersion() || {}
  } catch (e) { /* 版本信息加载失败不阻塞页面 */ }
  try {
    lastResult.value = await getUpdateStatus()
  } catch (e) { /* 忽略 */ }
}

const doCheckUpdate = async () => {
  checking.value = true
  try {
    updateInfo.value = await checkUpdate()
  } catch (e) {
    updateInfo.value = null
  } finally {
    checking.value = false
  }
}

const openReleasePage = () => {
  window.open('https://github.com/thebestskinhead/schedule-system-v2/releases', '_blank')
}

const doStartUpdate = async () => {
  try {
    await ElMessageBox.confirm(
      `确定要更新到 ${updateInfo.value.target} 吗？\n\n` +
      '· 更新期间服务将停止约 1 分钟，页面会自动跳转到更新进度页\n' +
      '· 更新只覆盖服务端程序与前后端产物，不影响数据库与配置\n' +
      '· 在进度页中可随时终止更新并恢复当前版本',
      '确认热更新',
      { confirmButtonText: '立即更新', cancelButtonText: '取消', type: 'warning' }
    )
  } catch (e) {
    return // 用户取消
  }
  updating.value = true
  try {
    await startUpdate()
  } catch (e) {
    updating.value = false
    return
  }
  ElMessage.success('更新程序已启动，正在切换到更新进度页…')

  // 轮询 8080：等服务停止 → 进度页上线 → 跳转；
  // 若服务直接恢复（更新失败自动回滚），刷新页面即可继续使用。
  let sawDown = false
  const timer = setInterval(async () => {
    try {
      const res = await fetch('/api/state', { cache: 'no-store' })
      const json = await res.json().catch(() => null)
      if (json && json.phase) {
        clearInterval(timer)
        window.location.href = '/'
      } else if (sawDown) {
        clearInterval(timer)
        ElMessage.info('服务已恢复，正在刷新页面')
        setTimeout(() => window.location.reload(), 1500)
      }
    } catch (e) {
      sawDown = true // 请求失败 = 服务已停止，等待进度页上线
    }
  }, 800)
  setTimeout(() => clearInterval(timer), 180000) // 3 分钟兜底
}

onMounted(() => {
  loadConfig()
  loadVersion()
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

.ver-text {
  font-size: 16px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

.ver-new {
  color: #67c23a;
}

.release-notes {
  max-width: 480px;
  max-height: 120px;
  overflow-y: auto;
  font-size: 12px;
  line-height: 1.6;
  color: #606266;
  background: #fafafa;
  border-radius: 6px;
  padding: 8px 12px;
  white-space: pre-wrap;
  word-break: break-all;
}

.update-warning {
  background: #fdf6ec;
  color: #e6a23c;
  border: 1px solid #faecd8;
  border-radius: 6px;
  padding: 10px 14px;
  font-size: 13px;
  line-height: 1.6;
  margin-bottom: 12px;
}
</style>
