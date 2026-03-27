<template>
  <div class="init-container">
    <van-loading v-if="loading" type="spinner" color="#1989fa" class="loading-overlay" />

    <div class="init-box">
      <!-- 头部 -->
      <div class="init-header">
        <van-icon name="setting-o" size="48" color="#1989fa" class="init-icon" />
        <h2 class="init-title">排班系统安装向导</h2>
        <p class="init-subtitle">欢迎使用排班管理系统，请完成以下步骤进行安装</p>
      </div>

      <!-- 步骤条 -->
      <van-steps :active="currentStep" class="steps">
        <van-step>配置数据库</van-step>
        <van-step>初始化数据</van-step>
        <van-step>创建管理员</van-step>
        <van-step>完成</van-step>
      </van-steps>

      <!-- 步骤1: 配置数据库 -->
      <div v-if="currentStep === 0" class="step-content">
        <van-notice-bar
          left-icon="info-o"
          text="请输入MySQL数据库连接信息。如果数据库不存在，系统会自动创建。"
          class="mb-4"
          :scrollable="false"
          wrapable
        />

        <van-form ref="dbFormRef" @submit="testAndSaveDB">
          <van-cell-group inset>
            <van-field
              v-model="dbConfig.host"
              label="主机"
              placeholder="localhost"
              left-icon="cluster-o"
              :rules="[{ required: true, message: '请输入数据库主机' }]"
            />
            <van-field
              v-model="dbConfig.port"
              label="端口"
              placeholder="3306"
              type="digit"
              :rules="[{ required: true, message: '请输入端口' }]"
            />
            <van-field
              v-model="dbConfig.user"
              label="用户名"
              placeholder="root"
              left-icon="user-o"
              :rules="[{ required: true, message: '请输入用户名' }]"
            />
            <van-field
              v-model="dbConfig.password"
              label="密码"
              placeholder="请输入数据库密码"
              type="password"
              left-icon="lock"
              :rules="[{ required: true, message: '请输入密码' }]"
            />
            <van-field
              v-model="dbConfig.dbname"
              label="数据库名"
              placeholder="schedule_system_v2"
              left-icon="records-o"
              :rules="[{ required: true, message: '请输入数据库名称' }]"
            >
              <template #button>
                <span class="form-hint-inline">自动创建</span>
              </template>
            </van-field>
          </van-cell-group>

          <div class="step-actions">
            <van-button
              round
              block
              type="primary"
              native-type="submit"
              :loading="testingDB"
              loading-text="正在连接..."
            >
              测试并保存
            </van-button>
          </div>
        </van-form>
      </div>

      <!-- 步骤2: 初始化数据表 -->
      <div v-else-if="currentStep === 1" class="step-content">
        <van-notice-bar
          v-if="dbCheckResult && !dbCheckResult.empty && !forceInit"
          left-icon="warning-o"
          color="#ed6a0c"
          background="#fffbe8"
          :text="`数据库非空：已存在 ${dbCheckResult.tables?.length || 0} 个表，请选择操作方式`"
          class="mb-4"
          :scrollable="false"
          wrapable
        />

        <van-notice-bar
          v-else-if="initSuccess"
          left-icon="passed"
          color="#07c160"
          background="#e8f7ef"
          text="数据表创建成功，数据库初始化完成"
          class="mb-4"
          :scrollable="false"
        />

        <div v-else-if="!initStarted" class="init-welcome">
          <van-icon name="passed" size="64" color="#07c160" class="init-welcome-icon" />
          <p class="init-welcome-text">数据库配置已保存</p>
          <p class="init-welcome-sub">点击下方按钮开始初始化数据表</p>
        </div>

        <div v-else-if="initing" class="init-progress">
          <van-circle
            :current-rate="initProgress"
            :rate="initProgress"
            :speed="50"
            :text="initProgress + '%'"
            size="100px"
            :stroke-width="6"
            :color="initProgress === 100 ? '#07c160' : '#1989fa'"
          />
          <p class="progress-text">{{ initStatus }}</p>
        </div>

        <div class="step-actions">
          <van-button plain @click="currentStep = 0">返回修改</van-button>

          <template v-if="dbCheckResult && !dbCheckResult.empty && !forceInit && !initSuccess">
            <van-button type="primary" @click="continueWithoutInit">继续（跳过建表）</van-button>
            <van-button type="danger" @click="enableForceInit">覆盖（重新初始化）</van-button>
          </template>

          <template v-else-if="!initSuccess">
            <van-button
              type="primary"
              @click="initDatabase"
              :loading="initing"
              loading-text="初始化中..."
            >
              {{ forceInit ? '确认覆盖初始化' : '开始初始化' }}
            </van-button>
          </template>

          <van-button v-else type="primary" @click="currentStep = 2">
            下一步
          </van-button>
        </div>
      </div>

      <!-- 步骤3: 创建管理员 -->
      <div v-else-if="currentStep === 2" class="step-content">
        <van-notice-bar
          left-icon="info-o"
          text="请设置系统管理员账号，用于登录和管理系统"
          class="mb-4"
          :scrollable="false"
          wrapable
        />

        <van-form ref="adminFormRef" @submit="createAdmin">
          <van-cell-group inset>
            <van-field
              v-model="adminForm.studentId"
              label="学号/工号"
              placeholder="请输入学号或工号"
              left-icon="user-o"
              :rules="[
                { required: true, message: '请输入学号或工号' },
                { validator: (val) => val.length >= 3 && val.length <= 20, message: '长度在 3 到 20 个字符' }
              ]"
            />
            <van-field
              v-model="adminForm.name"
              label="姓名"
              placeholder="请输入姓名"
              left-icon="contact"
              :rules="[
                { required: true, message: '请输入姓名' },
                { validator: (val) => val.length >= 2 && val.length <= 20, message: '长度在 2 到 20 个字符' }
              ]"
            />
            <van-field
              v-model="adminForm.email"
              label="邮箱"
              placeholder="请输入邮箱地址"
              left-icon="envelop-o"
              :rules="[
                { required: true, message: '请输入邮箱' },
                { pattern: /^[\w.-]+@[\w.-]+\.\w+$/, message: '请输入有效的邮箱地址' }
              ]"
            />
            <van-field
              v-model="adminForm.password"
              label="密码"
              placeholder="请设置登录密码（至少6位）"
              type="password"
              left-icon="lock"
              :rules="[
                { required: true, message: '请输入密码' },
                { validator: (val) => val.length >= 6, message: '密码长度至少6位' }
              ]"
            />
            <van-field
              v-model="adminForm.confirmPassword"
              label="确认密码"
              placeholder="请再次输入密码"
              type="password"
              left-icon="lock"
              :rules="[
                { required: true, message: '请确认密码' },
                { validator: (val) => val === adminForm.password, message: '两次输入的密码不一致' }
              ]"
            />
            <van-field
              v-model="adminForm.department"
              is-link
              readonly
              label="所属部门"
              placeholder="请选择所属部门"
              :rules="[{ required: true, message: '请选择所属部门' }]"
              @click="showDeptPicker = true"
            />
          </van-cell-group>

          <div class="step-actions">
            <van-button plain @click="currentStep = 1">返回</van-button>
            <van-button round type="primary" native-type="submit" :loading="creating" loading-text="创建中...">
              创建管理员
            </van-button>
          </div>
        </van-form>

        <!-- 部门选择弹窗 -->
        <van-popup v-model:show="showDeptPicker" round position="bottom">
          <van-picker
            :columns="deptColumns"
            @confirm="onDeptConfirm"
            @cancel="showDeptPicker = false"
          />
        </van-popup>
      </div>

      <!-- 步骤4: 完成 -->
      <div v-else-if="currentStep === 3" class="step-content">
        <div class="success-result">
          <van-icon name="passed" size="72" color="#07c160" />
          <h3 class="success-title">安装完成</h3>
          <p class="success-subtitle">系统已成功安装，现在可以开始使用了</p>
          <div class="success-info">
            <van-cell-group inset>
              <van-cell title="管理员账号" :value="adminForm.studentId" />
              <van-cell title="访问地址" value="http://localhost:8080" />
            </van-cell-group>
          </div>
          <div class="step-actions">
            <van-button round block type="primary" @click="goToLogin">进入系统</van-button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast, showConfirmDialog } from 'vant'
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
const showDeptPicker = ref(false)

const dbFormRef = ref()
const adminFormRef = ref()

const deptColumns = [
  { text: '办公室', value: '办公室' },
  { text: '竞赛部', value: '竞赛部' },
  { text: '项目部', value: '项目部' },
  { text: '科普部', value: '科普部' }
]

const dbConfig = reactive({
  host: 'localhost',
  port: '3306',
  user: 'root',
  password: '',
  dbname: 'schedule_system_v2'
})

const adminForm = reactive({
  studentId: '',
  name: '',
  email: '',
  password: '',
  confirmPassword: '',
  department: '办公室'
})

const onDeptConfirm = ({ selectedValues }) => {
  adminForm.department = selectedValues[0]
  showDeptPicker.value = false
}

const testAndSaveDB = async () => {
  testingDB.value = true
  try {
    await testDBConnection(dbConfig)
    showToast({ message: '数据库连接成功', type: 'success' })
    currentStep.value = 1
    await checkDBStatus()
  } catch (error) {
    showToast({ message: error?.response?.data?.message || '数据库连接失败，请检查配置', type: 'fail' })
  } finally {
    testingDB.value = false
  }
}

const checkDBStatus = async () => {
  try {
    const data = await checkDatabase(dbConfig)
    dbCheckResult.value = data
  } catch (error) {
    console.error('检查数据库状态失败:', error)
  }
}

const enableForceInit = () => {
  showConfirmDialog({
    title: '警告',
    message: '这将删除数据库中所有现有数据，此操作不可恢复！',
    confirmButtonText: '我已了解风险，确认覆盖',
    cancelButtonText: '取消',
    confirmButtonColor: '#ee0a24'
  }).then(() => {
    forceInit.value = true
  }).catch(() => {
    // 取消
  })
}

const continueWithoutInit = () => {
  currentStep.value = 2
}

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
    showToast({ message: '数据库表初始化成功', type: 'success' })
  } catch (error) {
    initError.value = error?.response?.data?.message || '初始化失败'
    initProgress.value = 0
  } finally {
    initing.value = false
  }
}

const createAdmin = async () => {
  creating.value = true
  try {
    await apiCreateAdmin({
      studentId: adminForm.studentId,
      name: adminForm.name,
      email: adminForm.email,
      password: adminForm.password,
      department: adminForm.department
    })

    showToast({ message: '管理员创建成功，系统安装完成', type: 'success' })

    setTimeout(() => {
      window.location.href = '/'
    }, 1500)
  } catch (error) {
    showToast({ message: error?.response?.data?.message || '创建失败', type: 'fail' })
  } finally {
    creating.value = false
  }
}

const goToLogin = () => {
  router.replace('/login')
}

onMounted(async () => {
  try {
    const data = await getInstallStatus()
    if (data?.installed) {
      showToast({ message: '系统已安装，跳转到登录页', type: 'success' })
      router.replace('/login')
    }
  } catch (error) {
    console.error('检查安装状态失败:', error)
  }
})
</script>

<style scoped>
.init-container {
  min-height: 100%;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px 16px;
  padding-top: calc(20px + env(safe-area-inset-top));
  padding-bottom: calc(20px + env(safe-area-inset-bottom));
}

.loading-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 999;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.7);
}

.init-box {
  width: 100%;
  max-width: 500px;
  background: #fff;
  border-radius: 16px;
  padding: 24px 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

.init-header {
  text-align: center;
  padding: 16px 0 24px;
}

.init-icon {
  margin-bottom: 12px;
}

.init-title {
  color: #303133;
  margin: 0 0 8px;
  font-size: 22px;
}

.init-subtitle {
  color: #606266;
  margin: 0;
  font-size: 13px;
  line-height: 1.5;
}

.steps {
  margin-bottom: 24px;
}

.step-content {
  padding: 8px 0;
}

.step-actions {
  display: flex;
  justify-content: center;
  gap: 12px;
  margin-top: 24px;
  padding: 20px 0 0;
  flex-wrap: wrap;
}

.init-progress {
  padding: 40px 20px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.init-welcome {
  padding: 40px 20px;
  text-align: center;
}

.init-welcome-icon {
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
  color: #606266;
  font-size: 14px;
}

.success-result {
  text-align: center;
  padding: 20px 0;
}

.success-title {
  color: #303133;
  margin: 16px 0 8px;
  font-size: 20px;
}

.success-subtitle {
  color: #909399;
  font-size: 14px;
  margin: 0 0 24px;
}

.success-info {
  margin-bottom: 24px;
}

.form-hint-inline {
  font-size: 12px;
  color: #909399;
}

.mb-4 {
  margin-bottom: 12px;
}
</style>
