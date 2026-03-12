<template>
  <div class="smtp-config-page">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <h1 class="logo">排班系统 - SMTP配置</h1>
          <div class="nav-menu">
            <el-button @click="router.push('/')">返回首页</el-button>
          </div>
        </div>
      </el-header>

      <el-main class="main-content">
        <div class="page-container">
          <el-card>
            <template #header>
              <div class="card-header">
                <span>SMTP服务器配置</span>
                <el-button type="primary" @click="showAddDialog = true">
                  <el-icon><Plus /></el-icon> 添加配置
                </el-button>
              </div>
            </template>

            <el-table :data="configs" v-loading="loading">
              <el-table-column prop="id" label="ID" width="60" />
              <el-table-column prop="host" label="服务器" width="150" />
              <el-table-column prop="port" label="端口" width="80" />
              <el-table-column prop="username" label="用户名" />
              <el-table-column prop="from_email" label="发件邮箱" />
              <el-table-column prop="use_tls" label="TLS" width="80">
                <template #default="{ row }">
                  <el-tag :type="row.use_tls ? 'success' : 'info'">{{ row.use_tls ? '是' : '否' }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="use_ssl" label="SSL" width="80">
                <template #default="{ row }">
                  <el-tag :type="row.use_ssl ? 'success' : 'info'">{{ row.use_ssl ? '是' : '否' }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="is_active" label="状态" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.is_active ? 'success' : 'danger'">
                    {{ row.is_active ? '启用' : '停用' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="200">
                <template #default="{ row }">
                  <el-button link type="primary" @click="editConfig(row)">编辑</el-button>
                  <el-button link type="primary" @click="testConfig(row)">测试</el-button>
                  <el-button link type="danger" @click="deleteConfig(row.id)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>

          <el-card class="mt-4">
            <template #header>
              <div class="card-header">
                <span>网站域名配置</span>
              </div>
            </template>
            <el-form :model="siteForm" inline>
              <el-form-item label="网站域名">
                <el-input v-model="siteForm.domain" placeholder="如: https://schedule.example.com" style="width: 350px">
                  <template #prefix>
                    <el-icon><Link /></el-icon>
                  </template>
                </el-input>
                <div class="form-hint">用于生成密码重置邮件中的链接</div>
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="handleSaveSiteConfig" :loading="savingSite">
                  保存域名
                </el-button>
              </el-form-item>
            </el-form>
          </el-card>

          <el-card class="mt-4">
            <template #header>
              <div class="card-header">
                <span>发送测试邮件</span>
              </div>
            </template>
            <el-form :model="testForm" inline>
              <el-form-item label="收件人邮箱">
                <el-input v-model="testForm.to" placeholder="输入测试收件人邮箱" style="width: 300px" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" @click="sendTestEmail" :loading="testing">
                  发送测试邮件
                </el-button>
              </el-form-item>
            </el-form>
          </el-card>
        </div>
      </el-main>
    </el-container>

    <!-- 添加/编辑配置对话框 -->
    <el-dialog v-model="showAddDialog" :title="editingConfig ? '编辑配置' : '添加配置'" width="600px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="120px">
        <el-form-item label="SMTP服务器" prop="host">
          <el-input v-model="form.host" placeholder="如: smtp.gmail.com, smtp.qq.com" />
        </el-form-item>
        <el-form-item label="端口" prop="port">
          <el-input-number v-model="form.port" :min="1" :max="65535" placeholder="如: 587, 465" />
        </el-form-item>
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" placeholder="邮箱账号" />
        </el-form-item>
        <el-form-item label="密码/授权码" prop="password">
          <el-input v-model="form.password" type="password" placeholder="邮箱密码或授权码" show-password />
        </el-form-item>
        <el-form-item label="发件人名称" prop="from">
          <el-input v-model="form.from" placeholder="显示的发件人名称" />
        </el-form-item>
        <el-form-item label="发件人邮箱" prop="from_email">
          <el-input v-model="form.from_email" placeholder="发件人邮箱地址" />
        </el-form-item>
        <el-form-item label="安全连接">
          <el-radio-group v-model="securityType">
            <el-radio label="tls">TLS</el-radio>
            <el-radio label="ssl">SSL</el-radio>
            <el-radio label="none">无</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="启用配置">
          <el-switch v-model="form.is_active" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="saveConfig" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Link } from '@element-plus/icons-vue'
import { getSMTPConfigs, saveSMTPConfig, deleteSMTPConfig, testSMTP, getSiteConfig, saveSiteConfig } from '../api/smtp'

const router = useRouter()
const loading = ref(false)
const configs = ref([])
const showAddDialog = ref(false)
const saving = ref(false)
const testing = ref(false)
const editingConfig = ref(null)
const formRef = ref()
const securityType = ref('tls')

const form = reactive({
  id: 0,
  host: '',
  port: 587,
  username: '',
  password: '',
  from: '排班系统',
  from_email: '',
  use_tls: true,
  use_ssl: false,
  is_active: true
})

const testForm = reactive({
  to: ''
})

const siteForm = reactive({
  domain: ''
})
const savingSite = ref(false)

const rules = {
  host: [{ required: true, message: '请输入SMTP服务器地址', trigger: 'blur' }],
  port: [{ required: true, message: '请输入端口', trigger: 'blur' }],
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  from: [{ required: true, message: '请输入发件人名称', trigger: 'blur' }],
  from_email: [{ required: true, message: '请输入发件人邮箱', trigger: 'blur' }]
}

const loadConfigs = async () => {
  loading.value = true
  try {
    const data = await getSMTPConfigs()
    configs.value = data || []
    // 加载网站域名配置
    const siteData = await getSiteConfig()
    if (siteData) {
      siteForm.domain = siteData.domain || ''
    }
  } catch (error) {
    console.error('加载配置失败:', error)
  } finally {
    loading.value = false
  }
}

const handleSaveSiteConfig = async () => {
  if (!siteForm.domain) {
    ElMessage.warning('请输入网站域名')
    return
  }
  savingSite.value = true
  try {
    await saveSiteConfig(siteForm)
    ElMessage.success('域名配置已保存')
  } catch (error) {
    ElMessage.error('保存失败')
  } finally {
    savingSite.value = false
  }
}

const editConfig = (config) => {
  editingConfig.value = config
  Object.assign(form, config)
  securityType.value = config.use_ssl ? 'ssl' : (config.use_tls ? 'tls' : 'none')
  showAddDialog.value = true
}

const saveConfig = async () => {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  // 根据安全类型设置标志
  form.use_tls = securityType.value === 'tls'
  form.use_ssl = securityType.value === 'ssl'

  saving.value = true
  try {
    await saveSMTPConfig(form)
    ElMessage.success('保存成功')
    showAddDialog.value = false
    loadConfigs()
  } catch (error) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

const deleteConfig = async (id) => {
  try {
    await ElMessageBox.confirm('确定删除此配置?', '确认', { type: 'warning' })
    await deleteSMTPConfig(id)
    ElMessage.success('删除成功')
    loadConfigs()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const testConfig = async (config) => {
  testing.value = true
  try {
    await testSMTP({ to: config.username })
    ElMessage.success('测试邮件已发送到 ' + config.username)
  } catch (error) {
    ElMessage.error('发送失败')
  } finally {
    testing.value = false
  }
}

const sendTestEmail = async () => {
  if (!testForm.to) {
    ElMessage.warning('请输入收件人邮箱')
    return
  }
  testing.value = true
  try {
    await testSMTP(testForm)
    ElMessage.success('测试邮件已发送')
    testForm.to = ''
  } catch (error) {
    ElMessage.error('发送失败')
  } finally {
    testing.value = false
  }
}

const resetForm = () => {
  editingConfig.value = null
  form.id = 0
  form.host = ''
  form.port = 587
  form.username = ''
  form.password = ''
  form.from = '排班系统'
  form.from_email = ''
  form.use_tls = true
  form.use_ssl = false
  form.is_active = true
  securityType.value = 'tls'
}

onMounted(() => {
  loadConfigs()
})
</script>

<style scoped>
.header {
  background: #fff;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  padding: 0;
}

.header-content {
  max-width: 1400px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
  padding: 0 20px;
}

.logo {
  margin: 0;
  font-size: 20px;
  color: #409eff;
}

.main-content {
  min-height: calc(100vh - 60px);
  padding: 20px;
}

.page-container {
  max-width: 1400px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.mt-4 {
  margin-top: 16px;
}

.form-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
</style>
