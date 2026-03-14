# 任务2：模拟登录与认证构建 - 详细设计报告

## 一、现状分析

### 1.1 现有认证流程调研

```
Current Authentication Flow:

┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Login.vue  │────▶│  request.js  │────▶│   Backend    │────▶│  JWT Token   │
└──────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
       │                                                    │
       │                                                    ▼
       │                                             ┌──────────────┐
       │                                             │ localStorage │
       │                                             │   (token)    │
       │                                             └──────────────┘
       │                                                    │
       ▼                                                    ▼
┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  userStore   │◀────│ checkToken() │◀────│ user/profile │◀────│  Bearer Token│
│  (Pinia)     │     │              │     │   (API)      │     │  (Header)    │
└──────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
```

### 1.2 现有认证代码分析

#### 1.2.1 API 层 (src/api/user.js)
```javascript
// 依赖后端API
export const login = (data) => request.post('/user/login', data)
export const checkToken = () => request.get('/user/profile')
```

#### 1.2.2 Store 层 (src/stores/user.js)
```javascript
// State
const token = ref(localStorage.getItem('token') || '')
const user = ref(null)

// Actions
async function checkAuth() {
  if (!token.value) return false
  try {
    await checkToken()      // 后端验证
    await loadUserInfo()    // 获取用户信息
    return true
  } catch {
    clearToken()
    return false
  }
}
```

#### 1.2.3 路由守卫 (src/router/index.js)
```javascript
router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore()
  if (!userStore.checked) {
    await userStore.checkAuth()  // 依赖后端验证
  }
  // 权限检查...
})
```

### 1.3 现有结构问题

| 问题点 | 现状 | 影响 |
|-------|------|------|
| 后端依赖 | 所有认证操作都需后端 | 无法离线运行 |
| 单一角色 | 只能登录一个用户 | 无法演示权限差异 |
| 会话管理 | JWT + 后端校验 | 刷新后需重新验证 |
| 权限计算 | 依赖后端返回的角色 | 无法快速切换 |

## 二、理想状态设计

### 2.1 模拟认证架构

```
Mock Authentication Architecture:

┌─────────────────────────────────────────────────────────────────────────┐
│                         Mock Auth Layer                                  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────────┐   │
│  │  MockAuthService │◀──▶│  MockUserStore   │◀──▶│   RoleSwitcher   │   │
│  │  (认证服务)       │    │  (用户状态管理)   │    │   (角色切换器)    │   │
│  └────────┬─────────┘    └──────────────────┘    └──────────────────┘   │
│           │                                                              │
│           ▼                                                              │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │                    Preset Users (4 Roles)                       │    │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐          │    │
│  │  │  Admin   │ │ Office   │ │  Dept    │ │  Member  │          │    │
│  │  │系统管理员 │ │办公室管理 │ │部门管理  │ │ 普通成员  │          │    │
│  │  │  id: 1   │ │  id: 2   │ │  id: 3   │ │  id: 4   │          │    │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘          │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### 2.2 多角色设计

```javascript
// 预设用户配置
const PRESET_USERS = [
  {
    id: 1,
    student_id: 'admin',
    name: '系统管理员',
    role: 'admin',
    department: '办公室',
    dept_role: 'dept_admin',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=admin',
    permissions: ['*']  // 所有权限
  },
  {
    id: 2,
    student_id: 'office001',
    name: '办公室管理员',
    role: 'user',
    department: '办公室',
    dept_role: 'dept_admin',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=office',
    permissions: [
      'schedule:manage:all',
      'user:manage:all',
      'schedule:publish'
    ]
  },
  {
    id: 3,
    student_id: 'dept001',
    name: '竞赛部部长',
    role: 'user',
    department: '竞赛部',
    dept_role: 'dept_admin',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=dept',
    permissions: [
      'schedule:manage:dept',
      'user:manage:dept'
    ]
  },
  {
    id: 4,
    student_id: 'member001',
    name: '普通成员',
    role: 'user',
    department: '竞赛部',
    dept_role: 'dept_member',
    avatar: 'https://api.dicebear.com/7.x/avataaars/svg?seed=member',
    permissions: [
      'schedule:view',
      'availability:edit',
      'duty:view'
    ]
  }
]
```

### 2.3 认证状态机

```
Auth State Machine:

                    ┌─────────────┐
         ┌─────────▶│  Unauth     │◀────────┐
         │          │  (未认证)    │         │
         │          └──────┬──────┘         │
         │                 │ login()         │ logout()
         │                 ▼                 │
   token │          ┌─────────────┐         │
 expired │    ┌────▶│  Authenticating │     │
         │    │     │  (认证中)     │     │
         │    │     └──────┬──────┘     │
         │    │            │ success    │
         │    │            ▼            │
         │    │     ┌─────────────┐     │
         │    └─────│  Authenticated    │─────┘
         │          │  (已认证)    │
         │          └──────┬──────┘
         │                 │ switchRole()
         │                 ▼
         │          ┌─────────────┐
         └─────────│  RoleSwitching  │
                   │  (切换中)    │
                   └─────────────┘
```

## 三、现有结构与理想的差距

### 3.1 差距分析矩阵

| 维度 | 现有结构 | 理想结构 | 差距等级 |
|-----|---------|---------|---------|
| 后端依赖 | 必须 | 零依赖 | ⭐⭐⭐⭐⭐ |
| 角色切换 | 需重新登录 | 一键切换 | ⭐⭐⭐⭐⭐ |
| 会话持久 | 依赖JWT校验 | 本地Token即可 | ⭐⭐⭐⭐ |
| 权限计算 | 后端返回 | 本地计算 | ⭐⭐⭐⭐ |
| 并发模拟 | 单用户 | 多用户切换 | ⭐⭐⭐ |

### 3.2 技术债务分析

```
Technical Debt:

1. API层耦合 (Critical)
   └── 所有认证API直接调用后端
   └── 修改方案: 添加Mock拦截层

2. Store层硬编码 (High)
   └── checkAuth() 直接调用checkToken()
   └── 修改方案: 抽象认证接口

3. 路由守卫依赖 (High)
   └── 依赖后端API获取安装状态
   └── 修改方案: Mock系统状态

4. 视图层直接引用 (Medium)
   └── 多处直接检查user.value?.role
   └── 修改方案: 保持兼容，无需修改
```

## 四、修改位置与实施方案

### 4.1 修改点清单

```
Modification Points:

mock/
├── auth/
│   ├── MockAuthService.js      # 新增: 认证服务
│   ├── MockTokenManager.js     # 新增: Token管理
│   └── RoleSwitcher.js         # 新增: 角色切换器
│
src/
├── api/
│   ├── request.js              # 修改: 添加Mock拦截
│   └── user.js                 # 兼容: 保持接口不变
│
├── stores/
│   └── user.js                 # 可选: 添加调试功能
│
└── components/
    └── Layout.vue              # 修改: 添加角色切换UI
```

### 4.2 核心实现代码

#### 4.2.1 MockAuthService (mock/auth/MockAuthService.js)

```javascript
import { PRESET_USERS } from './preset-users.js'

class MockAuthService {
  constructor() {
    this.currentUser = null
    this.token = null
    this.subscribers = []
  }

  // 模拟登录
  async login(credentials) {
    await this.simulateDelay(300)
    
    const user = PRESET_USERS.find(
      u => u.student_id === credentials.student_id && 
           u.password === credentials.password
    )
    
    if (!user) {
      throw new Error('用户名或密码错误')
    }
    
    this.currentUser = user
    this.token = this.generateToken(user)
    
    // 持久化到localStorage
    localStorage.setItem('mock_auth_token', this.token)
    localStorage.setItem('mock_current_user_id', user.id)
    
    this.notifySubscribers({ type: 'login', user })
    
    return {
      token: this.token,
      user: this.sanitizeUser(user)
    }
  }

  // 快速切换角色（Demo专用）
  async switchRole(userId) {
    const user = PRESET_USERS.find(u => u.id === userId)
    if (!user) throw new Error('用户不存在')
    
    this.currentUser = user
    this.token = this.generateToken(user)
    
    localStorage.setItem('mock_auth_token', this.token)
    localStorage.setItem('mock_current_user_id', user.id)
    
    this.notifySubscribers({ type: 'switch', user })
    
    return { user: this.sanitizeUser(user) }
  }

  // 验证Token
  async checkToken(token) {
    await this.simulateDelay(100)
    
    // 简化验证：检查token是否匹配当前用户
    if (token !== this.token) {
      throw new Error('Token无效')
    }
    
    return this.sanitizeUser(this.currentUser)
  }

  // 登出
  async logout() {
    this.currentUser = null
    this.token = null
    localStorage.removeItem('mock_auth_token')
    localStorage.removeItem('mock_current_user_id')
    this.notifySubscribers({ type: 'logout' })
  }

  // 从本地存储恢复会话
  restoreSession() {
    const token = localStorage.getItem('mock_auth_token')
    const userId = localStorage.getItem('mock_current_user_id')
    
    if (token && userId) {
      const user = PRESET_USERS.find(u => u.id === parseInt(userId))
      if (user) {
        this.currentUser = user
        this.token = token
        return true
      }
    }
    return false
  }

  // 生成模拟JWT Token
  generateToken(user) {
    const header = btoa(JSON.stringify({ alg: 'mock', typ: 'JWT' }))
    const payload = btoa(JSON.stringify({
      sub: user.id,
      name: user.name,
      role: user.role,
      iat: Date.now(),
      exp: Date.now() + 24 * 60 * 60 * 1000  // 24小时
    }))
    const signature = btoa('mock-signature')
    return `${header}.${payload}.${signature}`
  }

  // 辅助方法
  sanitizeUser(user) {
    const { password, ...safe } = user
    return safe
  }

  simulateDelay(ms) {
    return new Promise(resolve => setTimeout(resolve, ms))
  }

  subscribe(callback) {
    this.subscribers.push(callback)
  }

  notifySubscribers(event) {
    this.subscribers.forEach(cb => cb(event))
  }
}

export const mockAuth = new MockAuthService()
```

#### 4.2.2 请求拦截集成 (src/api/request.js)

```javascript
import axios from 'axios'
import { ElMessage } from 'element-plus'
import { mockAuth } from '../../mock/auth/MockAuthService.js'

const isMockMode = import.meta.env.VITE_MOCK === 'true'

// Mock路由映射
const mockRoutes = {
  'POST /user/login': mockAuth.login.bind(mockAuth),
  'GET /user/profile': mockAuth.checkToken.bind(mockAuth),
  // ... 其他Mock路由
}

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 30000
})

// Mock拦截器
if (isMockMode) {
  request.interceptors.request.use(
    async (config) => {
      const routeKey = `${config.method.toUpperCase()} ${config.url}`
      const mockHandler = mockRoutes[routeKey]
      
      if (mockHandler) {
        // 构造Mock响应
        const mockResponse = {
          data: { code: 200, message: 'success', data: null },
          status: 200,
          statusText: 'OK',
          headers: {},
          config
        }
        
        try {
          const data = await mockHandler(config.data || {})
          mockResponse.data.data = data
          return Promise.reject({ 
            __isMock: true, 
            response: mockResponse 
          })
        } catch (error) {
          mockResponse.data.code = 400
          mockResponse.data.message = error.message
          return Promise.reject({
            __isMock: true,
            response: mockResponse
          })
        }
      }
      
      return config
    }
  )
  
  // 处理Mock响应
  request.interceptors.response.use(
    (response) => response,
    (error) => {
      if (error.__isMock) {
        return Promise.resolve(error.response)
      }
      return Promise.reject(error)
    }
  )
}

// 原有拦截器...
export default request
```

#### 4.2.3 角色切换UI组件 (src/components/RoleSwitcher.vue)

```vue
<template>
  <el-dropdown v-if="isMockMode" placement="bottom">
    <el-button type="warning" size="small">
      <el-icon><Switch /></el-icon>
      切换角色
    </el-button>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item 
          v-for="user in presetUsers" 
          :key="user.id"
          :class="{ active: currentUserId === user.id }"
          @click="switchTo(user.id)"
        >
          <el-avatar :size="24" :src="user.avatar" />
          <span>{{ user.name }}</span>
          <el-tag size="small" :type="getRoleType(user)">
            {{ getRoleLabel(user) }}
          </el-tag>
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup>
import { mockAuth } from '../../mock/auth/MockAuthService.js'

const isMockMode = import.meta.env.VITE_MOCK === 'true'
const presetUsers = mockAuth.getPresetUsers()

const switchTo = async (userId) => {
  await mockAuth.switchRole(userId)
  window.location.reload()  // 刷新以更新所有状态
}
</script>
```

### 4.3 实施步骤

```
Phase 1: 基础认证服务 (2h)
├── 1.1 创建 MockAuthService 类
├── 1.2 实现 login/logout/checkToken
├── 1.3 配置预设用户数据
└── 1.4 实现 Token 生成与验证

Phase 2: 请求拦截集成 (1.5h)
├── 2.1 修改 request.js 添加 Mock 拦截
├── 2.2 创建路由映射表
├── 2.3 处理 Mock 错误响应
└── 2.4 验证认证流程

Phase 3: 角色切换功能 (1.5h)
├── 3.1 实现 switchRole 方法
├── 3.2 创建 RoleSwitcher 组件
├── 3.3 集成到 Layout.vue
└── 3.4 添加角色切换提示

Phase 4: 会话持久化 (1h)
├── 4.1 实现 restoreSession
├── 4.2 页面刷新后自动恢复
├── 4.3 处理 Token 过期
└── 4.4 测试完整流程
```

## 五、再次分析：核心需求与更优解

### 5.1 核心需求重定义

原始需求: "模拟登录与认证"
深层需求分析:
1. **演示便利性** - 快速展示不同权限角色的界面差异
2. **离线可用性** - 不依赖后端即可体验完整功能
3. **状态隔离性** - 不同角色的数据相互独立
4. **开发调试性** - 方便开发者测试各种场景

### 5.2 潜在问题分析

当前方案存在的问题:

```
Problems:

1. Token切换成本
   ├── 问题: 每次切换角色需要刷新页面
   ├── 影响: 用户体验有中断感
   └── 优化: 热切换 + 状态同步

2. 数据隔离复杂
   ├── 问题: 不同角色应看到不同数据
   ├── 影响: 切换后数据可能不匹配
   └── 优化: 角色-数据绑定机制

3. 并发模拟困难
   ├── 问题: 无法模拟多用户同时操作
   ├── 影响: 无法演示协作场景
   └── 优化: 多会话模拟
```

### 5.3 更优解决方案

#### 方案A: 热切换方案 (推荐)

**核心思想**: 不刷新页面，实时切换用户状态

```javascript
class HotRoleSwitcher {
  async switchRole(userId) {
    // 1. 保存当前状态
    const currentState = this.captureState()
    
    // 2. 切换用户
    await mockAuth.switchRole(userId)
    
    // 3. 重新加载用户相关数据
    await userStore.loadUserInfo()
    
    // 4. 更新路由权限
    await router.replace('/')
    
    // 5. 通知所有组件更新
    eventBus.emit('user-changed', { userId })
  }
}
```

**优势**:
- 无页面刷新，体验流畅
- 可保留部分全局状态
- 切换动画可定制

#### 方案B: 多会话模拟

**核心思想**: 同时维护多个用户会话

```javascript
class MultiSessionManager {
  sessions = new Map()
  
  createSession(userId) {
    const session = {
      user: PRESET_USERS.find(u => u.id === userId),
      token: generateToken(),
      data: {},  // 用户特定数据
      socket: null  // 模拟WebSocket
    }
    this.sessions.set(session.token, session)
    return session
  }
  
  // 模拟两个用户对话
  simulateConversation(user1Id, user2Id) {
    const session1 = this.createSession(user1Id)
    const session2 = this.createSession(user2Id)
    
    // 创建关联事件
    // 如：用户1提交申请 → 用户2收到审批通知
  }
}
```

**优势**:
- 可演示协作场景
- 支持实时通知模拟
- 更接近真实多用户环境

### 5.4 最终推荐架构

```
Recommended Architecture:

┌────────────────────────────────────────────────────────────────┐
│                      Auth Layer                                 │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────────┐         ┌──────────────────┐             │
│  │   Hot Switcher   │◀───────▶│  Session Manager │             │
│  │   (热切换器)      │         │   (会话管理器)    │             │
│  └────────┬─────────┘         └────────┬─────────┘             │
│           │                            │                       │
│           ▼                            ▼                       │
│  ┌──────────────────┐         ┌──────────────────┐             │
│  │  MockAuthService │         │  Multi-Session   │             │
│  │  (认证服务)       │         │  (多会话模拟)     │             │
│  └────────┬─────────┘         └──────────────────┘             │
│           │                                                     │
│           ▼                                                     │
│  ┌──────────────────────────────────────────────────────┐      │
│  │              Preset Users (4 Roles)                   │      │
│  │  ┌────────┐  ┌────────┐  ┌────────┐  ┌────────┐     │      │
│  │  │ Admin  │  │ Office │  │  Dept  │  │ Member │     │      │
│  │  │ Session│  │ Session│  │ Session│  │ Session│     │      │
│  │  └────────┘  └────────┘  └────────┘  └────────┘     │      │
│  └──────────────────────────────────────────────────────┘      │
│                                                                 │
└────────────────────────────────────────────────────────────────┘
```

### 5.5 关键决策

| 决策点 | 选择 | 理由 |
|-------|-----|-----|
| 切换方式 | 热切换 | 体验更好，技术可行 |
| 会话模型 | 单会话优先 | 满足主要场景，复杂度可控 |
| Token格式 | 模拟JWT | 兼容原有代码，无需修改解析逻辑 |
| 持久化策略 | localStorage | 简单可靠，支持刷新恢复 |
| 权限计算 | 本地计算 | 快速响应，无需网络请求 |

---

**结论**: 采用"热切换 + 单会话"方案，在保证演示便利性的同时控制复杂度，满足纯前端Demo的核心需求。
