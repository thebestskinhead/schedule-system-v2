# 排班系统纯前端 Demo 方案

> 基于 UI 树分析、权限结构分析和 API 文档的综合方案

## 一、项目分析总结

### 1.1 UI 组件树结构

```
App.vue (根组件)
├── 无布局页面 (guest 路由)
│   ├── Login.vue          登录页
│   ├── Register.vue       注册页
│   ├── ForgotPassword.vue 找回密码
│   ├── ResetPassword.vue  重置密码
│   └── Init.vue           系统初始化
│
└── Layout.vue (主布局)
    ├── Header (顶部导航)
    │   ├── Brand (系统标识)
    │   ├── MainNav (主导航)
    │   │   ├── 首页
    │   │   ├── 无课表
    │   │   ├── 我的值班
    │   │   ├── 排班结果
    │   │   ├── 权限申请
    │   │   └── 使用说明
    │   ├── ManageDropdown (管理菜单) - 权限控制
    │   │   ├── 排班管理 (canManageDept)
    │   │   ├── 每周分工 (canManageAll)
    │   │   ├── 用户管理 (canManageAll)
    │   │   ├── 学期设置 (canManageAll)
    │   │   └── SMTP配置 (isAdmin)
    │   └── UserDropdown (用户菜单)
    │       ├── 我的权限
    │       └── 退出登录
    │
    └── Main Content (主内容区)
        ├── Home.vue              首页概览
        ├── Availability.vue      无课表录入
        ├── CrawlerImport.vue     爬虫导入
        ├── MyDuty.vue            我的值班
        ├── ScheduleResult.vue    排班结果查看
        ├── MyPermissions.vue     我的权限
        ├── Readme.vue            使用说明
        ├── Schedule.vue          排班管理(权限)
        ├── DutyAssignment.vue    每周分工(权限)
        ├── UserManagement.vue    用户管理(权限)
        ├── TempPermission.vue    权限申请/管理
        ├── SMTPConfig.vue        SMTP配置(权限)
        └── SemesterSettings.vue  学期设置(权限)
```

### 1.2 权限校验结构

#### 权限模型
- **认证方式**: JWT Token (Bearer Token)
- **权限模型**: RBAC + 临时权限混合模型
- **角色层级**:
  1. 系统管理员 (`role=admin`) - 最高权限
  2. 办公室管理员 (`department="办公室"` + `dept_role="dept_admin"`)
  3. 部门管理员 (`dept_role="dept_admin"`)
  4. 部门成员 (`dept_role="dept_member"`)

#### 权限计算逻辑 (stores/user.js)
```javascript
isAdmin: user.value?.role === 'admin'
isOfficeAdmin: user.value?.department === '办公室' && user.value?.dept_role === 'dept_admin'
isDeptAdmin: user.value?.dept_role === 'dept_admin'

canManageAll: isAdmin || isOfficeAdmin || hasTempPermission('user:manage:all') || hasTempPermission('schedule:view:all')

canManageDept: isAdmin || isOfficeAdmin || isDeptAdmin || hasTempPermission('schedule:manage:dept') || hasTempPermission('user:manage:dept')
```

#### 路由守卫权限检查 (router/index.js)
- `requiresAuth`: 需要登录
- `requiresSysAdmin`: 需要系统管理员权限
- `requiresManageAll`: 需要办公室管理员及以上权限
- `requiresManageDept`: 需要部门管理员及以上权限
- `guest`: 仅未登录用户可访问

#### 权限组体系
```
schedule:manage:all
├── schedule:view
├── schedule:preview
├── schedule:confirm
├── schedule:edit
├── schedule:publish
└── schedule:manage:dept
    ├── schedule:view
    ├── schedule:view:dept
    ├── schedule:preview
    ├── schedule:confirm
    └── schedule:edit

user:manage:all
├── user:manage
└── user:manage:dept
    ├── user:view
    └── user:edit
```

### 1.3 API 接口清单

#### 认证相关
| 接口 | 方法 | 权限 | 用途 |
|------|------|------|------|
| /user/login | POST | 公开 | 登录 |
| /user/register | POST | 公开 | 注册 |
| /user/profile | GET | 需登录 | 获取用户信息 |

#### 排班管理
| 接口 | 方法 | 权限 | 用途 |
|------|------|------|------|
| /schedule | GET | 需登录 | 查看排班 |
| /schedule/current-week | GET | 公开 | 获取当前周次 |
| /admin/schedule/preview | POST | canManageDept | 预览排班 |
| /admin/schedule/confirm | POST | canManageDept | 确认排班 |
| /admin/schedule/settings | GET/POST | canManageDept | 排班设置 |

#### 无课表管理
| 接口 | 方法 | 权限 | 用途 |
|------|------|------|------|
| /availability | GET/PUT | 需登录 | 获取/更新无课表 |
| /crawler/import | POST | 需登录 | 爬虫导入 |

#### 用户管理
| 接口 | 方法 | 权限 | 用途 |
|------|------|------|------|
| /admin/users | GET | canManageAll | 用户列表 |
| /users/for-schedule | GET | canManageDept | 排班用户列表 |
| /admin/users/:id/department | PUT | canManageAll | 设置部门 |
| /admin/users/:id/dept-role | PUT | canManageDept | 设置角色 |

#### 权限系统
| 接口 | 方法 | 权限 | 用途 |
|------|------|------|------|
| /permissions/list | GET | 需登录 | 权限列表 |
| /temp-permissions/my | GET | 需登录 | 我的临时权限 |
| /admin/temp-permissions | GET/POST/DELETE | canManageAll | 临时权限管理 |
| /applications/my | GET | 需登录 | 我的申请 |
| /applications/pending | GET | canManageAll | 待审批列表 |
| /applications/:id/approve | POST | canManageAll | 审批申请 |

---

## 二、纯前端 Demo 方案

### 2.1 核心设计思路

1. **Mock 数据层**: 用本地 JSON 数据替代后端 API
2. **模拟认证**: 本地存储模拟用户登录状态
3. **权限模拟**: 预设多角色用户数据
4. **数据持久化**: 使用 localStorage 模拟数据存储
5. **延迟模拟**: 添加随机延迟模拟网络请求

### 2.2 Mock 数据结构设计

#### 2.2.1 用户数据 (mock/users.js)
```javascript
// 预设多角色用户用于切换演示
const MOCK_USERS = [
  {
    id: 1,
    student_id: 'admin',
    name: '系统管理员',
    email: 'admin@example.com',
    role: 'admin',                    // 系统管理员
    department: '办公室',
    dept_role: 'dept_admin',
    password: '123456'
  },
  {
    id: 2,
    student_id: 'office001',
    name: '办公室管理员',
    email: 'office@example.com',
    role: 'user',
    department: '办公室',
    dept_role: 'dept_admin',          // 办公室管理员
    password: '123456'
  },
  {
    id: 3,
    student_id: 'dept001',
    name: '部门管理员',
    email: 'dept@example.com',
    role: 'user',
    department: '竞赛部',
    dept_role: 'dept_admin',          // 部门管理员
    password: '123456'
  },
  {
    id: 4,
    student_id: 'member001',
    name: '普通成员',
    email: 'member@example.com',
    role: 'user',
    department: '竞赛部',
    dept_role: 'dept_member',         // 部门成员
    password: '123456'
  }
]
```

#### 2.2.2 排班数据 (mock/schedules.js)
```javascript
// 排班表数据结构
const MOCK_SCHEDULES = {
  // 按周次存储排班数据
  1: [
    { id: 1, week: 1, weekday: 1, period: 1, user_id: 4, user_name: '张三', status: 'confirmed' },
    { id: 2, week: 1, weekday: 1, period: 2, user_id: 5, user_name: '李四', status: 'pending' },
    // ...
  ],
  2: [ /* ... */ ],
  // ...
}

// 排班设置
const SCHEDULE_SETTINGS = {
  current_week: 1,
  auto_increment: true,
  need_per_cell: 2,
  min_per_cell: 0,
  max_per_day: 1,
  max_per_week: 2,
  export_title: '第{week}周排班表'
}
```

#### 2.2.3 无课表数据 (mock/availability.js)
```javascript
// 每个用户的无课表 (5天 x 4节 = 20个时段)
const MOCK_AVAILABILITY = {
  4: [  // user_id = 4
    { weekday: 1, period: 1, is_available: true },
    { weekday: 1, period: 2, is_available: false },
    // ...
  ]
}
```

#### 2.2.4 权限申请数据 (mock/permissions.js)
```javascript
const MOCK_APPLICATIONS = [
  {
    id: 1,
    application_no: 'APP202403120001',
    applicant_id: 4,
    applicant_name: '普通成员',
    department: '竞赛部',
    type_code: 'temp_permission',
    status: 'pending',  // pending/approved/rejected
    content: '申请临时排班管理权限',
    created_at: '2024-03-12T10:00:00Z'
  }
]

const MOCK_TEMP_PERMISSIONS = [
  {
    id: 1,
    user_id: 4,
    permission: 'schedule:manage:dept',
    permission_name: '排班管理(部门)',
    resource_type: 'dept',
    expires_at: '2024-12-31T23:59:59Z',
    days_left: 30
  }
]
```

### 2.3 API Mock 层设计 (mock/index.js)

#### 2.3.1 核心 Mock 架构
```javascript
// Mock 请求拦截器
class MockAPI {
  constructor() {
    this.storage = window.localStorage
    this.initData()
  }

  // 模拟延迟
  delay(ms = 300) {
    return new Promise(resolve => setTimeout(resolve, ms + Math.random() * 200))
  }

  // 统一响应格式
  success(data) {
    return { code: 200, message: 'success', data }
  }

  error(message, code = 400) {
    return { code, message, data: null }
  }

  // 初始化 Mock 数据到 localStorage
  initData() {
    if (!this.storage.getItem('mock_initialized')) {
      this.storage.setItem('mock_users', JSON.stringify(MOCK_USERS))
      this.storage.setItem('mock_schedules', JSON.stringify(MOCK_SCHEDULES))
      this.storage.setItem('mock_availability', JSON.stringify(MOCK_AVAILABILITY))
      this.storage.setItem('mock_current_week', '1')
      this.storage.setItem('mock_initialized', 'true')
    }
  }

  // 获取当前登录用户
  getCurrentUser() {
    const token = this.storage.getItem('mock_token')
    if (!token) return null
    const users = JSON.parse(this.storage.getItem('mock_users') || '[]')
    return users.find(u => u.id === parseInt(token))
  }
}
```

#### 2.3.2 各模块 Mock API

```javascript
// 用户模块
const userAPI = {
  async login(data) {
    await this.delay()
    const users = JSON.parse(this.storage.getItem('mock_users') || '[]')
    const user = users.find(u => u.student_id === data.student_id && u.password === data.password)
    if (!user) return this.error('用户名或密码错误')
    this.storage.setItem('mock_token', String(user.id))
    return this.success({ token: user.id, user: this.sanitizeUser(user) })
  },

  async getUserInfo() {
    await this.delay()
    const user = this.getCurrentUser()
    if (!user) return this.error('未登录', 401)
    return this.success(this.sanitizeUser(user))
  },

  sanitizeUser(user) {
    const { password, ...rest } = user
    return rest
  }
}

// 排班模块
const scheduleAPI = {
  async getSchedule(params) {
    await this.delay()
    const schedules = JSON.parse(this.storage.getItem('mock_schedules') || '{}')
    return this.success(schedules[params.week] || [])
  },

  async previewSchedule(data) {
    await this.delay()
    // 模拟排班算法
    const users = JSON.parse(this.storage.getItem('mock_availability') || '{}')
    // 根据无课表计算可排班人员
    const assignments = this.calculateSchedule(data, users)
    return this.success({ week: data.week, assignments, stats: { total_slots: 20, filled_slots: 18 } })
  },

  async confirmSchedule(data) {
    await this.delay()
    // 保存排班结果
    const schedules = JSON.parse(this.storage.getItem('mock_schedules') || '{}')
    schedules[data.week] = data.assignments
    this.storage.setItem('mock_schedules', JSON.stringify(schedules))
    return this.success({ week: data.week, total_assigned: data.assignments.length })
  }
}
```

### 2.4 权限模拟方案

#### 2.4.1 多角色切换机制
在 Demo 中增加角色切换功能，方便演示不同权限视图：

```javascript
// 在 Layout.vue 添加角色切换器
const DEMO_USERS = [
  { id: 1, name: '系统管理员', role: 'admin' },
  { id: 2, name: '办公室管理员', role: 'office_admin' },
  { id: 3, name: '部门管理员', role: 'dept_admin' },
  { id: 4, name: '普通成员', role: 'member' }
]

// 快速切换用户
const switchUser = (userId) => {
  localStorage.setItem('mock_token', String(userId))
  window.location.reload()
}
```

#### 2.4.2 权限可视化
在界面显著位置显示当前角色信息：

```vue
<!-- 在 Header 添加角色标识 -->
<el-tag v-if="userStore.isAdmin" type="danger" size="small">系统管理员</el-tag>
<el-tag v-else-if="userStore.isOfficeAdmin" type="success" size="small">办公室管理员</el-tag>
<el-tag v-else-if="userStore.isDeptAdmin" type="warning" size="small">部门管理员</el-tag>
<el-tag v-else type="info" size="small">普通成员</el-tag>
```

### 2.5 数据持久化策略

| 数据类型 | 存储方式 | 说明 |
|---------|---------|------|
| 用户信息 | localStorage | 模拟登录状态 |
| 排班数据 | localStorage | 用户操作后的排班结果 |
| 无课表数据 | localStorage | 每个用户的无课表 |
| 权限申请 | localStorage | 申请记录和状态 |
| 临时权限 | localStorage | 授权的临时权限 |
| 系统设置 | localStorage | 当前周次等配置 |

### 2.6 页面适配方案

#### 2.6.1 无需修改的页面
- Login.vue - 只需替换 API 调用
- Register.vue - 只需替换 API 调用
- Home.vue - 只需替换 API 调用
- Availability.vue - 只需替换 API 调用
- MyDuty.vue - 只需替换 API 调用
- ScheduleResult.vue - 只需替换 API 调用
- MyPermissions.vue - 只需替换 API 调用
- Readme.vue - 静态页面无需修改

#### 2.6.2 需要特殊处理的页面
- Schedule.vue - 排班算法需要前端模拟实现
- UserManagement.vue - 用户 CRUD 需要 Mock
- TempPermission.vue - 权限申请流程需要 Mock
- SMTPConfig.vue - 可以保留界面但标记为演示

### 2.7 排班算法模拟

排班系统的核心是排班算法，需要在 Mock 层实现简化版：

```javascript
// 模拟排班算法
function calculateSchedule(settings, availability) {
  const { week, days, periods, need_per_cell, department } = settings
  const assignments = []

  // 遍历每个时段
  days.forEach(weekday => {
    for (let period = 1; period <= periods; period++) {
      // 获取该时段有空的用户
      const availableUsers = getAvailableUsers(weekday, period, availability, department)

      // 随机选择指定数量用户
      const selectedUsers = shuffleArray(availableUsers).slice(0, need_per_cell)

      assignments.push({
        weekday,
        period,
        users: selectedUsers.map(u => ({
          id: u.id,
          name: u.name,
          student_id: u.student_id
        }))
      })
    }
  })

  return assignments
}
```

### 2.8 实施步骤

#### Phase 1: 基础 Mock 层搭建
1. 创建 `mock/` 目录结构
2. 实现 MockAPI 基础类
3. 实现用户认证 Mock
4. 修改 `request.js` 添加 Mock 拦截

#### Phase 2: 核心业务 Mock
1. 实现排班数据 Mock
2. 实现无课表数据 Mock
3. 实现排班算法 Mock
4. 实现用户管理 Mock

#### Phase 3: 权限系统 Mock
1. 实现临时权限 Mock
2. 实现权限申请 Mock
3. 添加多角色切换功能

#### Phase 4: 数据初始化
1. 创建初始化数据脚本
2. 预设演示数据
3. 添加数据重置功能

#### Phase 5: 优化与完善
1. 添加操作提示
2. 模拟网络延迟
3. 添加演示引导

### 2.9 文件结构规划

```
frontend-v2-demo/
├── mock/
│   ├── index.js           # Mock 入口，请求拦截
│   ├── users.js           # 用户数据和方法
│   ├── schedules.js       # 排班数据和方法
│   ├── availability.js    # 无课表数据和方法
│   ├── permissions.js     # 权限相关数据和方法
│   ├── algorithm.js       # 排班算法模拟
│   └── data/
│       ├── initial-data.js # 初始数据
│       └── generators.js   # 数据生成器
├── src/
│   ├── api/
│   │   └── request.js     # 修改以支持 Mock
│   ├── components/
│   ├── views/
│   └── stores/
└── PURE_FRONTEND_DEMO_PLAN.md # 本文档
```

### 2.10 关键技术点

1. **请求拦截**: 在 `request.js` 中判断环境，使用 Mock 替代真实请求
2. **数据关联**: 排班结果依赖于无课表数据，需要维护数据关联
3. **权限同步**: 用户角色变更后，权限计算需要实时更新
4. **状态持久**: 页面刷新后数据不丢失
5. **并发处理**: 模拟异步操作的竞态条件

---

## 三、总结

本方案通过 Mock 数据层、模拟认证、权限模拟和数据持久化四个核心机制，将原本依赖后端的全栈应用转化为纯前端 Demo。关键优势：

1. **零后端依赖**: 无需启动任何后端服务
2. **多角色演示**: 快速切换不同权限角色
3. **数据可持久**: 用户操作的数据会被保存
4. **功能完整**: 保留所有前端功能和交互
5. **易于部署**: 纯静态文件，可部署到任意静态托管服务
