<template>
  <div class="main-container">
    <!-- 版本与更新 -->
    <div class="card">
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
            <span class="ver-text" :class="{ 'ver-new': canHotUpdate }">
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
            {{ lastResult.phase === 'done' ? '成功' : (lastResult.phase === 'rolled-back' ? '已回滚' : '失败') }}
          </el-tag>
          <span class="form-hint" style="margin-left: 8px;">{{ lastResult.message }}（{{ lastResult.at }}）</span>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="doCheckUpdate" :loading="checking">检查更新</el-button>
          <el-button v-if="isCrossMajor" @click="openReleasePage">前往下载整包</el-button>
        </el-form-item>
        <div v-if="isCrossMajor" class="update-warning">
          新版本 {{ updateInfo.target }} 与当前版本（{{ updateInfo.current }}）跨大版本，可能包含数据库结构变更，无法热更新。
          请手动下载整包，并按发布说明迁移数据库后再升级。
        </div>
        <div class="form-hint">
          更新分两步：先下载并校验更新包（不影响服务），再停机安装（约 10 秒）。更新只覆盖服务端程序与前后端产物，不影响数据库与配置。
        </div>
      </el-form>
    </div>

    <!-- 下载与安装 -->
    <div class="card" style="margin-top: 20px;" v-if="showUpdatePanel">
      <div class="card-header">
        <span class="card-title">
          <el-icon><Download /></el-icon> 下载与安装
        </span>
      </div>

      <!-- 下载进度 -->
      <div v-if="isDownloading" class="download-progress">
        <div class="progress-head">
          <span>{{ progress.message || '正在下载…' }}</span>
          <span v-if="progress.total" class="form-hint">
            {{ formatSize(progress.received) }} / {{ formatSize(progress.total) }}
            <template v-if="progress.mirror"> · 镜像 {{ progress.mirror }}</template>
          </span>
        </div>
        <el-progress :percentage="progressPercent" :stroke-width="10" />
        <div class="progress-actions">
          <el-button size="small" @click="doCancelDownload" :loading="cancelling">取消下载</el-button>
        </div>
      </div>

      <!-- 就绪 -->
      <div v-else-if="isReady" class="ready-box">
        <el-alert type="success" :closable="false" show-icon>
          <template #title>
            更新包已就绪（{{ progress.target }}<span v-if="progress.mirror">，经镜像 {{ progress.mirror }} 下载</span>），可以执行安装。
          </template>
        </el-alert>
        <div class="progress-actions">
          <el-button type="primary" @click="doInstall" :loading="installing">安装更新</el-button>
          <el-button @click="doDownload" :loading="downloading">重新下载</el-button>
        </div>
      </div>

      <!-- 失败 -->
      <div v-else-if="isFailed" class="failed-box">
        <el-alert type="error" :closable="false" show-icon :title="'下载失败：' + (progress.error || '未知错误')" />
        <div class="progress-actions">
          <el-button type="primary" @click="doDownload" :loading="downloading">重新下载</el-button>
        </div>
      </div>

      <!-- 空闲：可发起下载 -->
      <div v-else class="idle-box">
        <el-button type="success" @click="doDownload" :loading="downloading">下载更新包</el-button>
        <span class="form-hint" style="margin-left: 12px;">
          下载期间服务正常运行，可随时取消；下载并校验完成后再执行安装。
        </span>
      </div>
    </div>

    <!-- 镜像加速 -->
    <div class="card" style="margin-top: 20px;">
      <div class="card-header">
        <span class="card-title">
          <el-icon><Connection /></el-icon> 镜像加速
        </span>
      </div>
      <div class="form-hint" style="margin-bottom: 12px;">
        按列表顺序以前缀方式拼接 GitHub 资产地址逐个尝试，全部失败后回退直连；移除条目即停用该镜像。安装包的 sha256 始终从 GitHub 直连校验。
      </div>
      <div v-for="(m, i) in mirrorDraft" :key="i" class="mirror-row">
        <el-input v-model="mirrorDraft[i]" placeholder="如: https://ghfast.top" size="default" />
        <el-button text type="danger" :icon="Delete" @click="removeMirror(i)" />
      </div>
      <div class="mirror-add">
        <el-input v-model="mirrorInput" placeholder="输入镜像地址后添加" @keyup.enter="addMirror" />
        <el-button @click="addMirror">添加</el-button>
        <el-button text @click="fillCommonMirrors">填入常用镜像</el-button>
      </div>
      <div style="margin-top: 12px;">
        <el-button type="primary" @click="saveMirrorList" :loading="savingMirrors">保存镜像配置</el-button>
      </div>
    </div>

    <!-- 更新日志 -->
    <div class="card" style="margin-top: 20px;">
      <div class="card-header">
        <span class="card-title">
          <el-icon><Document /></el-icon> 更新日志
        </span>
        <el-button text @click="toggleLog">{{ logVisible ? '收起' : '查看' }}</el-button>
      </div>
      <pre v-if="logVisible" class="update-log">{{ updateLog || '（暂无日志）' }}</pre>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete } from '@element-plus/icons-vue'
import {
  getVersion, checkUpdate, startDownload, getDownloadProgress, cancelDownload,
  installUpdate, getUpdateStatus, getUpdateLog, getMirrors, saveMirrors
} from '../api/system'

// ---------- 版本与检查 ----------

const versionInfo = ref({})
const updateInfo = ref(null)
const lastResult = ref(null)
const checking = ref(false)

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

const formatSize = (bytes) => {
  if (!bytes) return ''
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1048576).toFixed(2) + ' MB'
}

const openReleasePage = () => {
  window.open('https://github.com/thebestskinhead/schedule-system-v2/releases', '_blank')
}

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

// ---------- 下载 / 安装状态机 ----------

const progress = ref({ phase: 'idle' })
const downloading = ref(false)
const cancelling = ref(false)
const installing = ref(false)
let progressTimer = null

const isDownloading = computed(() =>
  ['checking', 'downloading', 'verifying'].includes(progress.value.phase))
const isReady = computed(() => progress.value.phase === 'ready')
const isFailed = computed(() => progress.value.phase === 'failed')
const progressPercent = computed(() => {
  const p = progress.value.progress
  return typeof p === 'number' ? p : 0
})

// 有可更新版本、或下载流程已启动时才展示「下载与安装」面板
const showUpdatePanel = computed(() => canHotUpdate.value || progress.value.phase !== 'idle')

const loadProgress = async () => {
  try {
    progress.value = await getDownloadProgress() || { phase: 'idle' }
  } catch (e) {
    return // 拉取失败时保持原状，下次轮询或手动刷新再试
  }
  // 刷新页面 / server 重启后恢复轮询：非终态即视为仍在下载
  if (isDownloading.value) {
    startProgressPolling()
  }
}

const startProgressPolling = () => {
  stopProgressPolling()
  progressTimer = setInterval(async () => {
    try {
      progress.value = await getDownloadProgress() || progress.value
    } catch (e) {
      return
    }
    if (!isDownloading.value) {
      stopProgressPolling()
    }
  }, 1000)
}

const stopProgressPolling = () => {
  if (progressTimer) {
    clearInterval(progressTimer)
    progressTimer = null
  }
}

const doDownload = async () => {
  downloading.value = true
  try {
    await startDownload()
    startProgressPolling()
    await loadProgress()
  } catch (e) {
    // 409（已有任务在进行）等：直接拉一次最新状态继续展示
    await loadProgress()
    if (isDownloading.value) {
      startProgressPolling()
    }
  } finally {
    downloading.value = false
  }
}

const doCancelDownload = async () => {
  try {
    await ElMessageBox.confirm('确定要取消本次下载吗？\n已下载的部分将被丢弃，可随时重新下载。', '取消下载', {
      confirmButtonText: '取消下载', cancelButtonText: '返回', type: 'warning'
    })
  } catch (e) {
    return
  }
  cancelling.value = true
  try {
    await cancelDownload()
    stopProgressPolling()
    await loadProgress()
  } catch (e) {
    await loadProgress()
  } finally {
    cancelling.value = false
  }
}

const doInstall = async () => {
  try {
    await ElMessageBox.confirm(
      `确定要安装 ${progress.value.target || updateInfo.value?.target || '新版本'} 吗？\n\n` +
      '· 安装期间服务将短暂停机（约 10 秒），页面会自动跳转到安装进度页\n' +
      '· 更新只覆盖服务端程序与前后端产物，不影响数据库与配置\n' +
      '· 安装失败会自动回滚并恢复更新前的版本',
      '确认安装更新',
      { confirmButtonText: '安装', cancelButtonText: '取消', type: 'warning' }
    )
  } catch (e) {
    return
  }
  installing.value = true
  try {
    await installUpdate()
  } catch (e) {
    // 安装程序启动后本服务随即被停止，响应可能中断——无论成功失败都开始监听进度页
  }
  ElMessage.success('安装程序已启动，正在切换到安装进度页…')

  // 轮询 8080：等服务停止 → 安装进度页上线 → 跳转；
  // 若服务直接恢复（安装失败自动回滚），刷新页面即可继续使用。
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
      sawDown = true // 请求失败 = 服务已停止，等待安装进度页上线
    }
  }, 800)
  setTimeout(() => clearInterval(timer), 180000) // 3 分钟兜底
}

// ---------- 镜像管理 ----------

const mirrors = ref([])
const mirrorDraft = ref([])
const mirrorInput = ref('')
const savingMirrors = ref(false)

const loadMirrors = async () => {
  try {
    const data = await getMirrors()
    mirrors.value = data?.mirrors || []
    mirrorDraft.value = [...mirrors.value]
  } catch (e) { /* 忽略 */ }
}

const addMirror = () => {
  const m = (mirrorInput.value || '').trim()
  if (!m) return
  if (!m.startsWith('http://') && !m.startsWith('https://')) {
    ElMessage.warning('镜像地址必须以 http:// 或 https:// 开头')
    return
  }
  if (mirrorDraft.value.includes(m)) {
    ElMessage.warning('该镜像已在列表中')
    return
  }
  mirrorDraft.value.push(m)
  mirrorInput.value = ''
}

const removeMirror = (i) => {
  mirrorDraft.value.splice(i, 1)
}

// 常用公共镜像：默认不预设（随时可能失效），由管理员一键填入后自行取舍
const COMMON_MIRRORS = [
  'https://ghfast.top/',
  'https://ghproxy.net/',
  'https://gh-proxy.com/',
  'https://github.moeyy.xyz/'
]

const fillCommonMirrors = () => {
  let added = 0
  for (const m of COMMON_MIRRORS) {
    if (!mirrorDraft.value.includes(m)) {
      mirrorDraft.value.push(m)
      added++
    }
  }
  if (added === 0) {
    ElMessage.info('常用镜像均已在列表中')
  }
}

const saveMirrorList = async () => {
  savingMirrors.value = true
  try {
    await saveMirrors(mirrorDraft.value)
    mirrors.value = [...mirrorDraft.value]
    ElMessage.success('镜像配置已保存')
  } catch (e) {
    // 错误信息由请求层提示
  } finally {
    savingMirrors.value = false
  }
}

// ---------- 更新日志 ----------

const logVisible = ref(false)
const updateLog = ref('')

const toggleLog = async () => {
  logVisible.value = !logVisible.value
  if (logVisible.value && !updateLog.value) {
    try {
      const data = await getUpdateLog()
      updateLog.value = data?.log || ''
    } catch (e) { /* 忽略 */ }
  }
}

onMounted(() => {
  loadVersion()
  loadProgress()
  loadMirrors()
})

onUnmounted(() => {
  stopProgressPolling()
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

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
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

.download-progress .progress-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  font-size: 14px;
  margin-bottom: 8px;
}

.download-progress .progress-actions,
.ready-box .progress-actions,
.failed-box .progress-actions {
  margin-top: 12px;
}

.ready-box,
.failed-box,
.idle-box {
  padding: 4px 0;
}

.mirror-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.mirror-row .el-input {
  max-width: 420px;
}

.mirror-add {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
}

.mirror-add .el-input {
  max-width: 420px;
}

.update-log {
  background: #1e1e2e;
  color: #a9b1d6;
  border-radius: 8px;
  padding: 12px 14px;
  font-family: Consolas, "Courier New", monospace;
  font-size: 12px;
  line-height: 1.7;
  max-height: 240px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
