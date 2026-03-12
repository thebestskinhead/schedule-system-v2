<template>
  <div class="init-container">
    <el-card class="init-box" v-loading="loading">
      <template #header>
        <div class="init-header">
          <el-icon class="init-icon"><Setting /></el-icon>
          <h2 class="init-title">排班系统安装向导</h2>
          <p class="init-subtitle">欢迎使用排班管理系统，请完成以下步骤进行安装</p>
        </div>
      </template>

      <el-steps :active="currentStep" finish-status="success" class="steps" align-center>
        <el-step title="配置数据库" description="设置数据库连接" />
        <el-step title="初始化数据" description="创建数据表" />
        <el-step title="创建管理员" description="设置管理员账号" />
        <el-step title="完成" description="开始使用" />
      </el-steps>

      <!-- 步骤1: 配置数据库 -->
      <div v-if="currentStep === 0" class="step-content">
        <el-alert
          title="数据库配置"
          description="请输入MySQL数据库连接信息。如果数据库不存在，系统会自动创建。"
          type="info"
          :closable="false"
          show-icon
          class="mb-4"
        />
        
        <el-form :model="dbConfig" :rules="dbRules" ref="dbFormRef" label-position="top">
          <el-row :gutter="20">
            <el-col :span="16">
              <el-form-item label="数据库主机" prop="host">
                <el-input v-model="dbConfig.host" placeholder="localhost">
                  <template #prefix>
                    <el-icon><Connection /></el-icon>
                  </template>
                </el-input>
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="端口" prop="port">
                <el-input v-model="dbConfig.port" placeholder="3306" />
              </el-form-item>
            </el-col>
          </el-row>
          
          <el-form-item label="数据库用户名" prop="user">
            <el-input v-model="dbConfig.user" placeholder="root">
              <template #prefix>
                <el-icon><User /></el-icon>
              </template>
            </el-input>
          </el-form-item>
          
          <el-form-item label="数据库密码" prop="password">
            <el-input v-model="dbConfig.password" type="password" placeholder="请输入数据库密码" show-password>
              <template #prefix>
                <el-icon><Lock /></el-icon>
              </template>
            </el-input>
          </el-form-item>
          
          <el-form-item label="数据库名称" prop="dbname">
            <el-input v-model="dbConfig.dbname" placeholder="schedule_system_v2">
              <template #prefix>
                <el-icon><Collection /></el-icon>
              </template>
            </el-input>
            <div class="form-hint">如果数据库不存在，系统会自动创建</div>
          </el-form-item>
          
          <div class="step-actions">
            <el-button type="primary" size="large" @click="testAndSaveDB" :loading="testingDB">
              {{ testingDB ? '正在连接...' : '测试并保存' }}
            </el-button>
          </div>
        </el-form>
      </div>

      <!-- 步骤2: 初始化数据表 -->
      <div v-else-if="currentStep === 1" class="step-content">
        <el-alert
          v-if="dbCheckResult && !dbCheckResult.empty && !forceInit"
          title="数据库非空"
          :description="`检测到数据库中已存在 ${dbCheckResult.tables?.length || 0} 个表，请选择操作方式`"
          type="warning"
          :closable="false"
          show-icon
          class="mb-4"
        />
        
        <el-alert
          v-else-if="initSuccess"
          title="数据表创建成功"
          description="数据库初始化完成"
          type="success"
          :closable="false"
          show-icon
          class="mb-4"
        />

        <div v-else-if="!initStarted" class="init-welcome">
          <el-icon class="init-welcome-icon"><CircleCheck /></el-icon>
          <p class="init-welcome-text">数据库配置已保存</p>
          <p class="init-welcome-sub">点击下方按钮开始初始化数据表</p>
        </div>

        <div v-else-if="initing" class="init-progress">
          <el-progress :percentage="initProgress" :status="initProgress === 100 ? 'success' : ''" />
          <p class="progress-text">{{ initStatus }}</p>
        </div>
        
        <div class="step-actions">
          <el-button size="large" @click="currentStep = 0">返回修改</el-button>
          
          <!-- 数据库非空时的选项 -->
          <template v-if="dbCheckResult && !dbCheckResult.empty && !forceInit && !initSuccess">
            <el-button type="primary" size="large" @click="continueWithoutInit">
              继续（跳过建表）
            </el-button>
            <el-button type="danger" size="large" @click="enableForceInit">
              覆盖（重新初始化）
            </el-button>
          </template>
          
          <template v-else-if="!initSuccess">
            <el-button type="primary" size="large" @click="initDatabase" :loading="initing">
              {{ forceInit ? '确认覆盖初始化' : '开始初始化' }}
            </el-button>
          </template>
          
          <el-button v-else type="primary" size="large" @click="currentStep = 2">
            下一步
          </el-button>
        </div>
      </div>

      <!-- 步骤3: 创建管理员 -->
      <div v-else-if="currentStep === 2" class="step-content">
        <el-alert
          title="创建管理员账号"
          description="请设置系统管理员账号，用于登录和管理系统"
          type="info"
          :closable="false"
          show-icon
          class="mb-4"
        />

        <el-form :model="adminForm" :rules="adminRules" ref="adminFormRef" label-position="top">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="学号/工号" prop="studentId">
                <el-input v-model="adminForm.studentId" placeholder="请输入学号或工号">
                  <template #prefix><el-icon><User /></el-icon></template>
                </el-input>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="姓名" prop="name">
                <el-input v-model="adminForm.name" placeholder="请输入姓名">
                  <template #prefix><el-icon><Avatar /></el-icon></template>
                </el-input>
              </el-form-item>
            </el-col>
          </el-row>
          
          <el-form-item label="邮箱" prop="email">
            <el-input v-model="adminForm.email" placeholder="请输入邮箱地址">
              <template #prefix><el-icon><Message /></el-icon></template>
            </el-input>
          </el-form-item>
          
          <el-form-item label="密码" prop="password">
            <el-input v-model="adminForm.password" type="password" placeholder="请设置登录密码（至少6位）" show-password>
              <template #prefix><el-icon><Lock /></el-icon></template>
            </el-input>
          </el-form-item>
          
          <el-form-item label="确认密码" prop="confirmPassword">
            <el-input v-model="adminForm.confirmPassword" type="password" placeholder="请再次输入密码" show-password>
              <template #prefix><el-icon><Lock /></el-icon></template>
            </el-input>
          </el-form-item>

          <el-form-item label="所属部门" prop="department">
            <el-select v-model="adminForm.department" placeholder="请选择所属部门" style="width: 100%">
              <el-option label="办公室" value="办公室" />
              <el-option label="竞赛部" value="竞赛部" />
              <el-option label="项目部" value="项目部" />
              <el-option label="科普部" value="科普部" />
            </el-select>
          </el-form-item>

          <div class="step-actions">
            <el-button size="large" @click="currentStep = 1">返回</el-button>
            <el-button type="primary" size="large" @click="createAdmin" :loading="creating">
              创建管理员
            </el-button>
          </div>
        </el-form>
      </div>

      <!-- 步骤4: 完成 -->
      <div v-else-if="currentStep === 3" class="step-content">
        <el-result
          icon="success"
          title="安装完成"
          sub-title="系统已成功安装，现在可以开始使用了"
        >
          <template #extra>
            <div class="success-info">
              <p><strong>管理员账号：</strong>{{ adminForm.studentId }}</p>
              <p><strong>访问地址：</strong>http://localhost:8080</p>
            </div>
            <el-button type="primary" size="large" @click="goToLogin">
              进入系统
            </el-button>
          </template>
        </el-result>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  getInstallStatus, 
  testDBConnection, 
  checkDatabase, 
  initDatabaseTables, 
  createAdmin as apiCreateAdmin 
} from '../api/system'

const router = useRouter()

const currentStep = ref(0)
const loading = ref(false)
const testingDB = ref(false)
const initing = ref(false)
const creating = ref(false)
const initStarted = ref(false)
const initSuccess = ref(false)
const initError = ref('')
const initProgress = ref(0)
const initStatus = ref('准备初始化...')
const dbCheckResult = ref(null)
const forceInit = ref(false)

const dbFormRef = ref()
const adminFormRef = ref()

// 数据库配置
const dbConfig = reactive({
  host: 'localhost',
  port: '3306',
  user: 'root',
  password: '',
  dbname: 'schedule_system_v2'
})

// 管理员表单
const adminForm = reactive({
  studentId: '',
  name: '',
  email: '',
  password: '',
  confirmPassword: '',
  department: '办公室'
})

// 验证确认密码
const validateConfirmPassword = (rule, value, callback) => {
  if (value !== adminForm.password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const dbRules = {
  host: [{ required: true, message: '请输入数据库主机', trigger: 'blur' }],
  port: [{ required: true, message: '请输入端口', trigger: 'blur' }],
  user: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  dbname: [{ required: true, message: '请输入数据库名称', trigger: 'blur' }]
}

const adminRules = {
  studentId: [
    { required: true, message: '请输入学号或工号', trigger: 'blur' },
    { min: 3, max: 20, message: '长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  name: [
    { required: true, message: '请输入姓名', trigger: 'blur' },
    { min: 2, max: 20, message: '长度在 2 到 20 个字符', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ],
  department: [
    { required: true, message: '请选择所属部门', trigger: 'change' }
  ]
}

// 测试数据库连接
const testAndSaveDB = async () => {
  const valid = await dbFormRef.value.validate().catch(() => false)
  if (!valid) return

  testingDB.value = true
  try {
    // 测试连接
    await testDBConnection(dbConfig)
    ElMessage.success('数据库连接成功')
    currentStep.value = 1
    // 检查数据库状态
    checkDBStatus()
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '数据库连接失败，请检查配置')
  } finally {
    testingDB.value = false
  }
}

// 检查数据库状态
const checkDBStatus = async () => {
  try {
    const res = await checkDatabase(dbConfig)
    dbCheckResult.value = res.data
  } catch (error) {
    console.error('检查数据库状态失败:', error)
  }
}

// 启用强制初始化
const enableForceInit = () => {
  ElMessageBox.confirm(
    '这将删除数据库中所有现有数据，此操作不可恢复！',
    '警告',
    {
      confirmButtonText: '我已了解风险，确认覆盖',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => {
    forceInit.value = true
  })
}

// 跳过初始化直接继续
const continueWithoutInit = () => {
  currentStep.value = 2
}

// 初始化数据库表
const initDatabase = async () => {
  initing.value = true
  initStarted.value = true
  initError.value = ''
  initProgress.value = 30
  initStatus.value = '正在创建数据表...'
  
  try {
    await initDatabaseTables({
      ...dbConfig,
      force: forceInit.value
    })
    initProgress.value = 100
    initSuccess.value = true
    initStatus.value = '初始化完成'
    ElMessage.success('数据库表初始化成功')
  } catch (error) {
    initError.value = error.response?.data?.message || '初始化失败'
    ElMessage.error(initError.value)
  } finally {
    initing.value = false
  }
}

// 创建管理员
const createAdmin = async () => {
  const valid = await adminFormRef.value.validate().catch(() => false)
  if (!valid) return

  creating.value = true
  try {
    const res = await apiCreateAdmin({
      studentId: adminForm.studentId,
      name: adminForm.name,
      email: adminForm.email,
      password: adminForm.password,
      department: adminForm.department
    })
    
    ElMessage.success(res.data?.message || '管理员创建成功，系统安装完成')
    
    // 安装完成，刷新页面重新加载应用
    setTimeout(() => {
      window.location.href = '/'
    }, 1500)
  } catch (error) {
    ElMessage.error(error.response?.data?.message || '创建失败')
  } finally {
    creating.value = false
  }
}

// 跳转到登录页
const goToLogin = () => {
  router.replace('/login')
}

onMounted(async () => {
  // 检查系统是否已安装
  try {
    const res = await getInstallStatus()
    // 注意：res.data 是 {code, message, data}，真正的数据在 res.data.data
    if (res.data?.data?.installed) {
      ElMessage.info('系统已安装，跳转到登录页')
      router.replace('/login')
    }
  } catch (error) {
    // 请求失败，继续显示安装页面
    console.error('检查安装状态失败:', error)
  }
})
</script>

<style scoped>
.init-container {
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.init-box {
  width: 600px;
  max-width: 100%;
}

.init-header {
  text-align: center;
  padding: 20px 0;
}

.init-icon {
  font-size: 48px;
  color: #409eff;
  margin-bottom: 16px;
}

.init-title {
  color: #303133;
  margin: 0 0 8px;
  font-size: 24px;
}

.init-subtitle {
  color: #606266;
  margin: 0;
  font-size: 14px;
}

.step-content {
  padding: 20px 0;
}

.step-actions {
  display: flex;
  justify-content: center;
  gap: 16px;
  margin-top: 30px;
  padding-top: 20px;
  border-top: 1px solid #e4e7ed;
  flex-wrap: wrap;
}

.init-progress {
  padding: 40px 20px;
  text-align: center;
}

.init-welcome {
  padding: 40px 20px;
  text-align: center;
}

.init-welcome-icon {
  font-size: 64px;
  color: #67c23a;
  margin-bottom: 16px;
}

.init-welcome-text {
  font-size: 18px;
  color: #303133;
  margin: 0 0 8px;
}

.init-welcome-sub {
  font-size: 14px;
  color: #909399;
  margin: 0;
}

.progress-text {
  margin-top: 16px;
  color: #606266;
}

.success-info {
  text-align: left;
  background: #f5f7fa;
  padding: 16px 24px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.success-info p {
  margin: 8px 0;
}

.form-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.mb-4 {
  margin-bottom: 16px;
}

:deep(.el-steps) {
  margin-bottom: 30px;
}
</style>
