# 排班系统移动端 - 禁区文档

> ⚠️ **警告**：本文档列出的代码区域与后端接口、权限系统强耦合，**禁止修改**或**只能谨慎调整**。任何改动都可能导致系统功能异常。

---

## 1. 核心请求模块 (CRITICAL)

### 1.1 文件：`src/api/request.js`

**危险等级**：🔴 CRITICAL - 禁止修改

**禁区内容**：
```javascript
// 1. 请求基础配置
const request = axios.create({
  baseURL: '/api/v1',    // ⚠️ 必须与后端 API 路径一致
  timeout: 30000         // ⚠️ 超时时间影响用户体验
})

// 2. 请求拦截器 - Token 注入
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')   // ⚠️ Token 存储键名
    if (token) {
      config.headers.Authorization = `Bearer ${token}`  // ⚠️ 认证头格式
    }
    return config
  }
)

// 3. 响应拦截器 - 统一响应处理
request.interceptors.response.use(
  (response) => {
    const { code, message, data } = response.data
    if (code !== 200) {           // ⚠️ 后端状态码约定
      ElMessage.error(message || '请求失败')
      return Promise.reject(new Error(message))
    }
    return data    // ⚠️ 直接返回 data，影响所有 API 调用
  },
  (error) => {
    if (error.response?.status === 401) {   // ⚠️ 401 处理逻辑
      localStorage.removeItem('token')
      window.location.href = '/login'         // ⚠️ 跳转逻辑
    }
    return Promise.reject(error)
  }
)
```

**为什么重要**：
- 所有 API 请求都依赖此模块
- Token 认证格式必须与后端 JWT 实现匹配
- 响应拦截器统一处理 `code: 200` 的约定
- 401 处理直接影响登录态判断

**允许调整**：
- ✅ 提示组件从 `ElMessage` 改为 `Toast`（Vant）
- ✅ 添加请求/响应日志（开发环境）

**禁止修改**：
- ❌ `baseURL` 路径
- ❌ `Authorization` 头部格式
- ❌ `code !== 200` 的判断逻辑
- ❌ `localStorage` 存储的键名 `token`
- ❌ 401 状态的统一处理逻辑

---

## 2. 权限存储模块 (CRITICAL)

### 2.1 文件：`src/stores/user.js`

**危险等级**：🔴 CRITICAL - 核心逻辑禁止修改

#### 禁区 2.1.1：权限计算属性
```javascript
// ⚠️ 角色判断逻辑 - 与后端 role 字段强耦合
const isAdmin = computed(() => user.value?.role === 'admin')

// ⚠️ 办公室管理员判断 - 与后端 dept_role 字段强耦合
const isOfficeAdmin = computed(() => 
  user.value?.department === '办公室' && user.value?.dept_role === 'dept_admin'
)

const isDeptAdmin = computed(() => user.value?.dept_role === 'dept_admin')
```

**字段映射**：
| 字段名 | 来源 | 可能值 | 说明 |
|--------|------|--------|------|
| `role` | 后端 User.role | `admin`, `user` | 系统角色 |
| `dept_role` | 后端 User.dept_role | `dept_admin`, `member` | 部门角色 |
| `department` | 后端 User.department | `办公室`, `竞赛部`, etc. | 部门名称 |

#### 禁区 2.1.2：权限组合逻辑
```javascript
// ⚠️ 全局管理权限 - 影响多个页面访问
const canManageAll = computed(() => {
  return isAdmin.value || 
         isOfficeAdmin.value ||
         hasTempPermission('user:manage:all') ||      // ⚠️ 临时权限码
         hasTempPermission('schedule:view:all')       // ⚠️ 临时权限码
})

// ⚠️ 部门管理权限 - 影响排班管理页面
const canManageDept = computed(() => {
  return isAdmin.value || 
         isOfficeAdmin.value || 
         isDeptAdmin.value ||
         hasTempPermission('schedule:manage:dept') ||  // ⚠️ 临时权限码
         hasTempPermission('user:manage:dept')         // ⚠️ 临时权限码
})
```

**临时权限码清单**：
```
user:manage:all          - 全局用户管理
schedule:view:all        - 全局排班查看
schedule:manage:dept     - 部门排班管理
user:manage:dept         - 部门用户管理
```

#### 禁区 2.1.3：临时权限处理
```javascript
// ⚠️ 临时权限数据结构依赖后端
const tempPermissions = ref([])

function hasTempPermission(perm) {
  // ⚠️ 依赖后端返回的 permission 字段
  return tempPermissions.value.some(p => p.permission === perm)
}

async function loadTempPermissions() {
  const perms = await getMyTempPermissions()  // ⚠️ API 端点
  tempPermissions.value = perms || []
}
```

**为什么重要**：
- 权限计算直接影响路由守卫的判断
- 临时权限系统与后端权限表结构强耦合
- 错误的权限判断会导致安全问题或功能异常

**允许调整**：
- ✅ 添加新的权限计算属性（扩展）
- ✅ 修改状态管理语法（如 options 转 setup）
- ✅ 添加辅助方法

**禁止修改**：
- ❌ `isAdmin`, `isOfficeAdmin`, `isDeptAdmin` 的判断逻辑
- ❌ `canManageAll`, `canManageDept` 的权限组合逻辑
- ❌ 临时权限码字符串值
- ❌ `hasTempPermission` 的判断逻辑
- ❌ `loadUserInfo` 和 `checkAuth` 的调用顺序

---

## 3. 路由守卫模块 (CRITICAL)

### 3.1 文件：`src/router/index.js`

**危险等级**：🔴 CRITICAL - 权限控制核心

#### 禁区 3.1.1：路由元信息
```javascript
// ⚠️ 权限控制标记 - 被 beforeEach 使用
meta: { requiresAuth: true }         // 需要登录
meta: { requiresSysAdmin: true }     // 需要系统管理员
meta: { requiresManageAll: true }    // 需要全局管理权限
meta: { requiresManageDept: true }   // 需要部门管理权限
meta: { guest: true }                // 仅游客可访问
```

#### 禁区 3.1.2：路由守卫逻辑
```javascript
router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore()

  // ⚠️ 认证检查 - 影响所有页面访问
  if (!userStore.checked) {
    await userStore.checkAuth()
  }

  const isAuthenticated = userStore.isAuthenticated
  const isAdmin = userStore.isAdmin
  const canManageAll = userStore.canManageAll
  const canManageDept = userStore.canManageDept

  // ⚠️ init 页面特殊处理 - 系统安装检查
  if (to.path === '/init') {
    const data = await getInstallStatus()   // ⚠️ API 端点
    if (data?.installed) {
      next('/login')
      return
    }
  }

  // ⚠️ 登录检查
  if (to.meta.requiresAuth && !isAuthenticated) {
    next('/login')
    return
  }

  // ⚠️ 系统管理员权限检查
  if (to.meta.requiresSysAdmin && !isAdmin) {
    next('/')
    return
  }

  // ⚠️ 全局管理权限检查
  if (to.meta.requiresManageAll && !canManageAll) {
    next('/')
    return
  }

  // ⚠️ 部门管理权限检查
  if (to.meta.requiresManageDept && !canManageDept) {
    next('/')
    return
  }

  // ⚠️ 游客页面检查
  if (to.meta.guest && isAuthenticated) {
    next('/')
    return
  }

  next()
})
```

**为什么重要**：
- 是整个应用的前门守卫
- 错误的权限判断会导致未授权访问或功能无法使用
- 与 `user.js` 的权限计算强耦合

**允许调整**：
- ✅ 添加新的路由
- ✅ 修改路由路径（需同步修改导航组件）
- ✅ 添加新的 meta 字段和对应的权限检查

**禁止修改**：
- ❌ 现有的 `meta` 标记和对应的检查逻辑
- ❌ 权限检查的调用顺序
- ❌ `next()` 的调用方式
- ❌ `userStore.checkAuth()` 调用

---

## 4. API 接口定义 (HIGH)

### 4.1 文件：`src/api/*.js`

**危险等级**：🟠 HIGH - 端点路径禁止修改

#### 禁区 4.1.1：API 端点路径
```javascript
// user.js
export const login = (data) => request.post('/user/login', data)
export const getUserInfo = () => request.get('/user/profile')
export const checkToken = () => request.get('/user/profile')  // ⚠️ 复用同一端点

// availability.js
export const getMyAvailability = () => request.get('/availability')

// schedule.js
export const getCurrentWeek = () => request.get('/schedule/current-week')
export const getSchedule = (params) => request.get('/schedule', { params })

// system.js
export const getInstallStatus = () => request.get('/system/installed')
export const getMyTempPermissions = () => request.get('/temp-permissions/my')
```

**完整 API 端点清单**：

| 端点 | 方法 | 用途 | 文件 |
|------|------|------|------|
| `/user/login` | POST | 登录 | user.js |
| `/user/register` | POST | 注册 | user.js |
| `/user/profile` | GET | 获取用户信息 | user.js |
| `/availability` | GET/POST/DELETE | 无课表管理 | availability.js |
| `/availability/import/xls` | POST | XLS导入 | availability.js |
| `/schedule` | GET | 查询排班 | schedule.js |
| `/schedule/current-week` | GET | 获取当前周 | schedule.js |
| `/duty/my` | GET | 我的值班 | schedule.js |
| `/admin/*` | 各种 | 管理接口 | schedule.js/system.js |

#### 禁区 4.1.2：请求/响应数据结构
```javascript
// ⚠️ 文件上传特殊处理
export const importFromXLS = (file) => {
  const formData = new FormData()
  formData.append('file', file)   // ⚠️ 字段名必须是 'file'
  return request.post('/availability/import/xls', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}
```

**为什么重要**：
- 端点路径必须与后端路由完全匹配
- 请求字段名与后端结构体字段绑定
- 文件上传的 `Content-Type` 特殊处理

**允许调整**：
- ✅ 添加新的 API 函数
- ✅ 封装参数处理逻辑
- ✅ 添加 TypeScript 类型定义（如使用 TS）

**禁止修改**：
- ❌ 现有的 API 端点路径
- ❌ 请求字段名
- ❌ 文件上传的 `Content-Type` 设置
- ❌ HTTP 方法（GET/POST/PUT/DELETE）

---

## 5. 用户数据结构 (HIGH)

### 5.1 文件：`src/stores/user.js` 中的 User 对象

**危险等级**：🟠 HIGH - 字段名禁止修改

**后端返回的用户结构**：
```javascript
user.value = {
  id: 1,                    // ⚠️ 用户ID
  username: "zhangsan",     // ⚠️ 用户名
  name: "张三",             // ⚠️ 真实姓名
  student_id: "2021001",    // ⚠️ 学号
  role: "user",             // ⚠️ 系统角色: admin/user
  department: "竞赛部",      // ⚠️ 部门名称
  dept_role: "dept_admin",  // ⚠️ 部门角色: dept_admin/member
  email: "xxx@xxx.com",     // 邮箱
  phone: "13800138000",     // 电话
  created_at: "2024-01-01"  // 创建时间
}
```

**关键字段**：
| 字段 | 类型 | 用途 | 影响范围 |
|------|------|------|----------|
| `role` | string | 系统角色判断 | `isAdmin`, 路由守卫 |
| `department` | string | 部门判断 | `isOfficeAdmin` |
| `dept_role` | string | 部门权限判断 | `isDeptAdmin`, `canManageDept` |
| `name` | string | 显示名称 | 页面显示 |

**为什么重要**：
- 字段名与后端数据库模型强耦合
- 影响权限计算的准确性

**禁止修改**：
- ❌ 字段名映射
- ❌ 权限相关字段的判断值

---

## 6. 后端响应格式约定 (HIGH)

### 6.1 标准响应结构

**危险等级**：🟠 HIGH

```javascript
// 成功响应
{
  code: 200,           // ⚠️ 必须严格等于 200
  message: "success",  // 消息文本
  data: { ... }        // 实际数据
}

// 错误响应
{
  code: 400/401/403/500,  // HTTP 状态码对应
  message: "错误信息",     // 错误提示
  data: null
}
```

**状态码约定**：
| code | 含义 | 处理 |
|------|------|------|
| 200 | 成功 | 返回 data |
| 400 | 请求错误 | 显示 message |
| 401 | 未认证 | 跳转登录页 |
| 403 | 无权限 | 显示 message |
| 500 | 服务器错误 | 显示 message |

**为什么重要**：
- `request.js` 中的拦截器依赖 `code === 200`
- 错误处理逻辑依赖 code 值

**禁止修改**：
- ❌ 响应解析逻辑中的 `code !== 200` 判断

---

## 7. 本地存储键名 (MEDIUM)

### 7.1 Token 存储

**危险等级**：🟡 MEDIUM

```javascript
// request.js
const token = localStorage.getItem('token')   // ⚠️ 键名: 'token'
localStorage.setItem('token', newToken)       // ⚠️ 键名: 'token'
localStorage.removeItem('token')              // ⚠️ 键名: 'token'
```

**影响范围**：
- `request.js` - 请求拦截器
- `user.js` - Token 状态管理
- 路由守卫 - 认证判断

**禁止修改**：
- ❌ 存储键名 `'token'`（除非同时修改所有引用）

---

## 禁区修改检查清单

在修改以下文件前，请逐项检查：

### 修改 request.js 前
- [ ] 没有修改 `baseURL`
- [ ] 没有修改 `Authorization` 头部格式
- [ ] 没有修改 `code !== 200` 判断
- [ ] 没有修改 `localStorage` 键名

### 修改 user.js 前
- [ ] 没有修改 `isAdmin`, `isOfficeAdmin`, `isDeptAdmin` 逻辑
- [ ] 没有修改 `canManageAll`, `canManageDept` 逻辑
- [ ] 没有修改临时权限码字符串
- [ ] 没有修改 `hasTempPermission` 逻辑

### 修改 router/index.js 前
- [ ] 没有修改现有的 `meta` 标记逻辑
- [ ] 没有修改权限检查的调用顺序
- [ ] 没有修改 `next()` 调用方式

### 修改 API 文件前
- [ ] 没有修改现有的 API 端点路径
- [ ] 没有修改请求字段名
- [ ] 没有修改 HTTP 方法

---

## 安全修改指南

### ✅ 安全的修改示例

1. **修改提示组件**
```javascript
// request.js - 允许
import { showToast } from 'vant'  // 替换 ElMessage

// 原代码
ElMessage.error(message)
// 修改为
showToast(message)
```

2. **添加新的 API**
```javascript
// user.js - 允许
export const updateProfile = (data) => request.put('/user/profile', data)
```

3. **添加新的权限计算**
```javascript
// user.js - 允许（新增）
const canViewSchedule = computed(() => {
  return isAuthenticated.value && user.value?.department
})
```

4. **修改路由结构（保留 meta）**
```javascript
// router/index.js - 允许
{
  path: '/new-page',           // 新路径
  name: 'NewPage',
  component: () => import('../views/NewPage.vue'),
  meta: { requiresAuth: true }  // 保留原有 meta
}
```

### ❌ 危险的修改示例

```javascript
// 不要这样做！
const isAdmin = computed(() => user.value?.role === 'superadmin')  // 改角色值
const canManageAll = computed(() => isAdmin.value)  // 移除临时权限检查
export const login = (data) => request.post('/auth/login', data)  // 改端点
if (code !== 0) { ... }  // 改状态码判断
```

---

## 紧急情况处理

如果必须修改禁区代码：

1. **记录变更**：在本文档中添加变更记录
2. **同步后端**：确保后端同步修改
3. **全量测试**：测试所有受影响的功能
4. **代码审查**：必须双人审查

**变更记录模板**：
```
日期: 2024-XX-XX
修改文件: src/xxx/xxx.js
修改内容: 简述
影响范围: 列出受影响的功能
审查人: XXX
测试结果: 通过/失败
```

---

**文档版本**: v1.0  
**更新日期**: 2024-03-26  
**关联文档**: API文档、后端接口文档
