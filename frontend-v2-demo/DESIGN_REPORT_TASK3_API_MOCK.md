# 任务3：模拟API响应 - 详细设计报告

## 一、现状分析

### 1.1 API层全面调研

通过分析 `src/api/` 目录下的所有模块，统计出完整的API接口清单：

```
API Inventory (Total: 38 endpoints):

src/api/user.js (8 endpoints)
├── POST /user/login           - 用户登录
├── POST /user/register        - 用户注册
├── POST /password/reset-request - 忘记密码
├── POST /password/reset       - 重置密码
├── GET  /user/profile         - 获取用户信息
├── POST /user/change-password - 修改密码
├── GET  /users/for-schedule   - 获取可排班用户
├── GET  /admin/users          - 用户列表(管理)
├── GET  /admin/users/by-dept  - 按部门查用户
├── POST /admin/users          - 创建用户
├── PUT  /admin/users/:id      - 更新用户
├── DELETE /admin/users/:id    - 删除用户
└── PUT  /admin/users/:id/role - 修改用户角色

src/api/schedule.js (15 endpoints)
├── GET  /schedule/current-week     - 获取当前周次
├── GET  /schedule                  - 获取排班表
├── POST /admin/schedule/current-week - 更新当前周次
├── POST /admin/schedule/preview    - 预览排班
├── POST /admin/schedule/confirm    - 确认排班
├── GET  /admin/schedule/settings   - 获取排班设置
├── POST /admin/schedule/settings   - 保存排班设置
├── POST /admin/schedule/update     - 更新排班
├── GET  /duty/my                   - 我的值班
├── PUT  /duty/status               - 更新值班状态
├── GET  /admin/templates           - 获取模板
├── POST /admin/templates           - 创建模板
├── PUT  /admin/templates           - 更新模板
├── DELETE /admin/templates/:id     - 删除模板
├── GET  /admin/duty-assignments    - 获取分工配置
├── POST /admin/duty-assignments    - 保存分工配置
├── GET  /duty-assignments/my-dept  - 我的部门分工
├── GET  /admin/schedule/semester-start - 获取学期起始
└── POST /admin/schedule/semester-start - 设置学期起始

src/api/availability.js (5 endpoints)
├── GET  /availability              - 获取我的无课表
├── POST /availability              - 添加无课时间
├── DELETE /availability            - 删除无课时间
├── POST /availability/import/cookie - Cookie导入
└── POST /availability/import/xls   - XLS导入

src/api/system.js (13 endpoints)
├── GET  /system/installed          - 安装状态
├── POST /system/test-db            - 测试数据库连接
├── POST /system/check-db           - 检查数据库
├── POST /system/init-tables        - 初始化表
├── POST /system/create-admin       - 创建管理员
├── GET  /admin/smtp/config         - 获取SMTP配置
├── POST /admin/smtp/config         - 保存SMTP配置
├── POST /admin/smtp/test           - 测试SMTP
├── GET  /admin/site/config         - 获取站点配置
├── POST /admin/site/config         - 保存站点配置
├── GET  /admin/temp-permissions    - 临时权限列表
├── POST /admin/temp-permissions    - 授予临时权限
├── DELETE /admin/temp-permissions/:id - 撤销权限
├── GET  /temp-permissions/my       - 我的临时权限
├── POST /admin/temp-permissions/cleanup - 清理过期
├── GET  /permissions/list          - 权限列表
└── GET  /departments               - 部门列表

src/api/application.js (9 endpoints)
├── GET  /application/types              - 申请类型
├── GET  /application/permissions/available - 可申请权限
├── GET  /applications/my                - 我的申请
├── POST /applications                   - 创建申请
├── GET  /applications/:id               - 申请详情
├── POST /applications/:id/cancel        - 取消申请
├── GET  /applications/pending           - 待审批列表
├── POST /applications/:id/approve       - 处理审批
└── GET  /applications/stats             - 申请统计
```

### 1.2 响应格式标准化分析

```javascript
// 标准成功响应 (from api.md)
{
  "code": 200,
  "message": "success",
  "data": { ... }
}

// 标准错误响应
{
  "code": 400|401|403|404|500,
  "message": "具体错误信息",
  "data": null
}

// 分页响应
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [ ... ],
    "total": 100,
    "page": 1,
    "page_size": 10
  }
}
```

### 1.3 现有请求处理流程

```
Current Request Flow:

View ──▶ API Function ──▶ request.js ──▶ Axios ──▶ Backend
  │           │              │
  │           │              └── Interceptors:
  │           │                   ├── Request: Add Token
  │           │                   └── Response: Handle Errors
  │           │
  │           └── Direct API calls:
  │                ├── GET /schedule?week=1
  │                ├── POST /admin/schedule/preview
  │                └── ...
  │
  └── Components handle responses directly

Problems:
1. 紧密耦合后端 - 每个API调用都需后端支持
2. 错误处理分散 - 每个组件需处理错误
3. 无离线能力 - 网络断开则完全不可用
4. 测试困难 - 需要完整后端环境
```

## 二、理想状态设计

### 2.1 Mock API架构

```
Mock API Architecture:

┌─────────────────────────────────────────────────────────────────────────┐
│                         Mock API Layer                                   │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    Route Matcher                                 │   │
│  │  (路由匹配器: URL + Method → Handler)                            │   │
│  └───────────────────────────┬─────────────────────────────────────┘   │
│                              │                                          │
│          ┌───────────────────┼───────────────────┐                     │
│          ▼                   ▼                   ▼                     │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐               │
│  │ Auth Handlers│   │Schedule Handlers│ │ System Handlers│              │
│  │   (认证)      │   │   (排班)        │   │   (系统)       │              │
│  └──────────────┘   └──────────────┘   └──────────────┘               │
│                                                          │              │
│          ┌───────────────────┬───────────────────┐                     │
│          ▼                   ▼                   ▼                     │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐               │
│  │ User Handlers│   │ Availability │   │ Application  │               │
│  │   (用户)      │   │   Handlers   │   │   Handlers   │               │
│  └──────────────┘   └──────────────┘   └──────────────┘               │
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    Response Builder                              │   │
│  │  (响应构建器: 统一包装成功/错误响应)                               │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### 2.2 Handler设计规范

```javascript
// 统一Handler接口
interface MockHandler {
  // 处理请求
  handle(request: MockRequest): Promise<MockResponse>
  
  // 验证权限
  validate(request: MockRequest): boolean
  
  // 模拟延迟
  delay?: number
}

// 请求上下文
interface MockRequest {
  method: 'GET' | 'POST' | 'PUT' | 'DELETE'
  url: string
  params: Record<string, string>
  body: any
  headers: Record<string, string>
  user?: User  // 当前登录用户
}

// 响应结构
interface MockResponse {
  code: number
  message: string
  data: any
}
```

### 2.3 路由注册表设计

```javascript
// 路由注册表 (mock/routes/index.js)
const mockRoutes = [
  // 认证路由
  { method: 'POST', path: '/user/login', handler: authHandlers.login },
  { method: 'GET', path: '/user/profile', handler: authHandlers.profile, auth: true },
  
  // 排班路由
  { method: 'GET', path: '/schedule', handler: scheduleHandlers.getSchedule, auth: true },
  { method: 'POST', path: '/admin/schedule/preview', handler: scheduleHandlers.preview, auth: true, permission: 'schedule:manage:dept' },
  { method: 'POST', path: '/admin/schedule/confirm', handler: scheduleHandlers.confirm, auth: true, permission: 'schedule:manage:dept' },
  
  // 无课表路由
  { method: 'GET', path: '/availability', handler: availabilityHandlers.get, auth: true },
  { method: 'PUT', path: '/availability', handler: availabilityHandlers.update, auth: true },
  
  // 用户管理路由
  { method: 'GET', path: '/admin/users', handler: userHandlers.list, auth: true, permission: 'user:manage:all' },
  { method: 'POST', path: '/admin/users', handler: userHandlers.create, auth: true, permission: 'user:manage:all' },
  
  // 权限路由
  { method: 'GET', path: '/temp-permissions/my', handler: permissionHandlers.getMy, auth: true },
  { method: 'GET', path: '/admin/temp-permissions', handler: permissionHandlers.getAll, auth: true, permission: 'user:manage:dept' },
  
  // 申请路由
  { method: 'GET', path: '/applications/my', handler: applicationHandlers.getMy, auth: true },
  { method: 'POST', path: '/applications', handler: applicationHandlers.create, auth: true },
  { method: 'GET', path: '/applications/pending', handler: applicationHandlers.getPending, auth: true, permission: 'user:manage:dept' },
  { method: 'POST', path: '/applications/:id/approve', handler: applicationHandlers.approve, auth: true, permission: 'user:manage:dept' },
  
  // 系统路由
  { method: 'GET', path: '/system/installed', handler: systemHandlers.checkInstalled },
  { method: 'GET', path: '/departments', handler: systemHandlers.getDepartments, auth: true },
  { method: 'GET', path: '/permissions/list', handler: systemHandlers.getPermissions, auth: true },
]
```

## 三、现有结构与理想的差距

### 3.1 差距分析

```
Gap Analysis:

1. 拦截点缺失 (Critical)
   ├── 现状: Axios直接发起HTTP请求
   ├── 需求: 在请求发出前拦截到Mock层
   └── 方案: Axios适配器或请求拦截器

2. 路由管理缺失 (Critical)
   ├── 现状: API函数直接硬编码URL
   ├── 需求: 可配置的Mock路由表
   └── 方案: 集中式路由注册表

3. 权限验证缺失 (High)
   ├── 现状: 后端进行权限校验
   ├── 需求: Mock层模拟权限检查
   └── 方案: 路由元数据+权限检查中间件

4. 响应模拟缺失 (High)
   ├── 现状: 等待后端返回
   ├── 需求: 立即返回Mock响应
   └── 方案: 预设响应+动态生成

5. 错误场景缺失 (Medium)
   ├── 现状: 真实后端错误
   ├── 需求: 可控的错误模拟
   └── 方案: 错误注入机制
```

### 3.2 技术选型对比

| 方案 | 实现复杂度 | 侵入性 | 灵活性 | 推荐度 |
|-----|-----------|-------|-------|-------|
| A: Axios适配器 | 中 | 低 | 高 | ⭐⭐⭐⭐⭐ |
| B: 请求拦截器 | 低 | 中 | 中 | ⭐⭐⭐⭐ |
| C: MSW (Mock Service Worker) | 低 | 低 | 高 | ⭐⭐⭐⭐⭐ |
| D: 重写API层 | 高 | 高 | 高 | ⭐⭐ |

## 四、修改位置与实施方案

### 4.1 推荐方案: MSW + 分层Handler

```
Architecture:

mock/
├── browser.js              # MSW浏览器集成
├── server.js               # MSW Node集成(用于测试)
├── handlers/
│   ├── index.js            # Handler汇总
│   ├── auth.handlers.js    # 认证相关
│   ├── user.handlers.js    # 用户管理
│   ├── schedule.handlers.js # 排班管理
│   ├── availability.handlers.js # 无课表
│   ├── permission.handlers.js   # 权限管理
│   ├── application.handlers.js  # 申请管理
│   └── system.handlers.js       # 系统接口
├── middleware/
│   ├── auth.middleware.js   # 认证中间件
│   └── permission.middleware.js # 权限中间件
└── utils/
    ├── response.js          # 响应构建工具
    ├── delay.js             # 延迟模拟
    └── error.js             # 错误生成
```

### 4.2 核心实现

#### 4.2.1 MSW集成 (mock/browser.js)

```javascript
import { setupWorker, rest } from 'msw'
import { handlers } from './handlers'

export const worker = setupWorker(...handlers)

// 启动Mock服务
export async function enableMocking() {
  if (import.meta.env.VITE_MOCK === 'true') {
    await worker.start({
      onUnhandledRequest: 'bypass', // 未处理的请求直接放行
    })
    console.log('[MSW] Mock服务已启动')
  }
}
```

#### 4.2.2 Handler示例 (mock/handlers/auth.handlers.js)

```javascript
import { rest } from 'msw'
import { db } from '../database'
import { success, error, unauthorized } from '../utils/response'
import { delay } from '../utils/delay'

export const authHandlers = [
  // 登录
  rest.post('/api/v1/user/login', async (req, res, ctx) => {
    await delay(300)
    
    const { student_id, password } = await req.json()
    
    const user = db.users.findFirst({
      where: { student_id: { equals: student_id } }
    })
    
    if (!user || user.password !== password) {
      return res(ctx.status(400), ctx.json(error('用户名或密码错误')))
    }
    
    const token = generateToken(user)
    
    return res(ctx.json(success({
      token,
      user: sanitizeUser(user)
    })))
  }),
  
  // 获取用户信息
  rest.get('/api/v1/user/profile', async (req, res, ctx) => {
    await delay(100)
    
    const user = getCurrentUser(req)
    if (!user) {
      return res(ctx.status(401), ctx.json(unauthorized()))
    }
    
    return res(ctx.json(success(sanitizeUser(user))))
  }),
]

// 辅助函数
function generateToken(user) {
  return `mock_token_${user.id}_${Date.now()}`
}

function getCurrentUser(req) {
  const authHeader = req.headers.get('Authorization')
  if (!authHeader) return null
  
  const token = authHeader.replace('Bearer ', '')
  const userId = parseInt(token.split('_')[2])
  
  return db.users.findFirst({
    where: { id: { equals: userId } }
  })
}

function sanitizeUser(user) {
  const { password, ...rest } = user
  return rest
}
```

#### 4.2.3 排班Handler (mock/handlers/schedule.handlers.js)

```javascript
import { rest } from 'msw'
import { db } from '../database'
import { success, error, forbidden } from '../utils/response'
import { scheduleAlgorithm } from '../algorithm/schedule'
import { requireAuth, requirePermission } from '../middleware'

export const scheduleHandlers = [
  // 获取当前周次
  rest.get('/api/v1/schedule/current-week', async (req, res, ctx) => {
    await delay(100)
    const settings = db.settings.findFirst()
    return res(ctx.json(success({ current_week: settings?.current_week || 1 })))
  }),
  
  // 获取排班表
  rest.get('/api/v1/schedule', requireAuth(async (req, res, ctx) => {
    await delay(200)
    
    const week = parseInt(req.url.searchParams.get('week'))
    const department = req.url.searchParams.get('department')
    
    const schedules = db.schedules.findMany({
      where: {
        week: { equals: week },
        ...(department && { department: { equals: department } })
      }
    })
    
    return res(ctx.json(success(schedules)))
  })),
  
  // 预览排班
  rest.post('/api/v1/admin/schedule/preview', 
    requireAuth(
      requirePermission('schedule:manage:dept', async (req, res, ctx) => {
        await delay(500)  // 模拟算法计算时间
        
        const settings = await req.json()
        
        // 获取部门用户
        const users = db.users.findMany({
          where: { department: { equals: settings.department } }
        })
        
        // 获取用户无课表
        const availability = db.availability.findMany({
          where: {
            user_id: { in: users.map(u => u.id) }
          }
        })
        
        // 执行排班算法
        const assignments = scheduleAlgorithm({
          users,
          availability,
          settings
        })
        
        return res(ctx.json(success({
          week: settings.week,
          assignments,
          stats: {
            total_slots: settings.days.length * settings.periods * settings.need_per_cell,
            filled_slots: assignments.reduce((sum, a) => sum + a.users.length, 0),
            conflicts: []
          }
        })))
      })
    )
  ),
  
  // 确认排班
  rest.post('/api/v1/admin/schedule/confirm',
    requireAuth(
      requirePermission('schedule:manage:dept', async (req, res, ctx) => {
        await delay(300)
        
        const { week, assignments, department } = await req.json()
        
        // 保存排班记录
        assignments.forEach(assignment => {
          assignment.users.forEach(user => {
            db.schedules.create({
              week,
              weekday: assignment.weekday,
              period: assignment.period,
              user_id: user.id,
              user_name: user.name,
              department,
              status: 'confirmed'
            })
          })
        })
        
        // 生成值班记录
        assignments.forEach(assignment => {
          assignment.users.forEach(user => {
            db.duties.create({
              week,
              weekday: assignment.weekday,
              period: assignment.period,
              user_id: user.id,
              status: 'pending'
            })
          })
        })
        
        return res(ctx.json(success({
          week,
          total_assigned: assignments.reduce((sum, a) => sum + a.users.length, 0)
        })))
      })
    )
  ),
]
```

#### 4.2.4 中间件实现 (mock/middleware/auth.middleware.js)

```javascript
// 认证中间件
export const requireAuth = (handler) => async (req, res, ctx) => {
  const user = getCurrentUser(req)
  
  if (!user) {
    return res(
      ctx.status(401),
      ctx.json({ code: 401, message: '未登录或Token已过期', data: null })
    )
  }
  
  // 将用户信息附加到请求
  req.user = user
  return handler(req, res, ctx)
}

// 权限中间件
export const requirePermission = (permission, handler) => async (req, res, ctx) => {
  const user = req.user
  
  // 系统管理员拥有所有权限
  if (user.role === 'admin') {
    return handler(req, res, ctx)
  }
  
  // 检查临时权限
  const tempPermissions = db.tempPermissions.findMany({
    where: { user_id: { equals: user.id } }
  })
  
  const hasPermission = checkPermission(user, permission, tempPermissions)
  
  if (!hasPermission) {
    return res(
      ctx.status(403),
      ctx.json({ code: 403, message: '无权限执行此操作', data: null })
    )
  }
  
  return handler(req, res, ctx)
}

function checkPermission(user, requiredPermission, tempPermissions) {
  // 权限组检查逻辑
  const permissionGroups = {
    'schedule:manage:dept': ['schedule:manage:dept', 'schedule:view', 'schedule:preview'],
    'schedule:manage:all': ['schedule:manage:all', 'schedule:manage:dept', 'schedule:view:all'],
    'user:manage:dept': ['user:manage:dept', 'user:view', 'user:edit'],
    'user:manage:all': ['user:manage:all', 'user:manage:dept']
  }
  
  // 检查角色权限
  if (user.dept_role === 'dept_admin') {
    return true  // 简化处理，实际需要更精细检查
  }
  
  // 检查临时权限
  return tempPermissions.some(tp => {
    const group = permissionGroups[tp.permission] || [tp.permission]
    return group.includes(requiredPermission)
  })
}
```

### 4.3 实施步骤

```
Phase 1: MSW基础搭建 (1.5h)
├── 1.1 安装MSW依赖
│   └── npm install msw --save-dev
├── 1.2 初始化MSW
│   └── npx msw init public/
├── 1.3 创建browser.js和server.js
└── 1.4 在main.js中集成

Phase 2: 工具函数开发 (1h)
├── 2.1 响应构建工具 (success/error/unauthorized/forbidden)
├── 2.2 延迟模拟工具
├── 2.3 错误注入工具
└── 2.4 Token解析工具

Phase 3: Handler开发 (4h)
├── 3.1 auth.handlers.js (登录/注册/信息获取)
├── 3.2 user.handlers.js (用户CRUD)
├── 3.3 schedule.handlers.js (排班算法+CRUD)
├── 3.4 availability.handlers.js (无课表管理)
├── 3.5 permission.handlers.js (权限管理)
├── 3.6 application.handlers.js (申请审批流程)
└── 3.7 system.handlers.js (系统配置)

Phase 4: 中间件开发 (1.5h)
├── 4.1 认证中间件
├── 4.2 权限检查中间件
├── 4.3 日志中间件
└── 4.4 请求验证中间件

Phase 5: 集成测试 (1h)
├── 5.1 验证所有API响应格式
├── 5.2 测试权限控制
├── 5.3 测试错误场景
└── 5.4 性能测试(大数据量)
```

## 五、再次分析：核心需求与更优解

### 5.1 核心需求重定义

原始需求: "模拟API响应"
深层需求分析:
1. **接口一致性** - Mock响应与真实API完全一致
2. **行为真实性** - 业务逻辑与后端一致（如排班算法）
3. **可控性** - 可模拟各种场景（成功/失败/延迟）
4. **可维护性** - API变更时Mock易于更新

### 5.2 潜在问题与优化

#### 问题1: API版本同步
```
问题: 后端API变更后，Mock需要同步更新
解决: 
├── 方案A: 从Swagger/OpenAPI自动生成Mock
├── 方案B: 使用契约测试保证一致性
└── 推荐: 方案A + 定期同步脚本
```

#### 问题2: 复杂业务逻辑
```
问题: 排班算法等复杂逻辑难以在前端完全模拟
解决:
├── 方案A: 简化算法逻辑
├── 方案B: 使用WebAssembly移植后端算法
└── 推荐: 方案A（Demo场景简化可接受）
```

#### 问题3: 大数据量性能
```
问题: 30周×5天×4节×多用户 = 大量数据
解决:
├── 方案A: 分页加载
├── 方案B: 懒生成数据
└── 推荐: 方案B（按需生成，避免初始化卡顿）
```

### 5.3 更优架构: 分层响应策略

```
Optimized Architecture:

┌────────────────────────────────────────────────────────────┐
│                    Response Strategy Layer                  │
├────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Static Response Cache                   │   │
│  │  (静态响应缓存: 常用数据预生成，如部门列表、权限列表)   │   │
│  └─────────────────────────────────────────────────────┘   │
│                           │                                 │
│                           ▼                                 │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Dynamic Response Builder                │   │
│  │  (动态响应构建: 基于请求参数实时计算)                  │   │
│  │  - 根据week参数返回对应周的排班                       │   │
│  │  - 根据department参数过滤用户                         │   │
│  └─────────────────────────────────────────────────────┘   │
│                           │                                 │
│                           ▼                                 │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Lazy Data Generator                     │   │
│  │  (懒数据生成: 首次访问时生成并缓存)                    │   │
│  │  - 用户无课表按需生成                                 │   │
│  │  - 排班结果实时计算                                   │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└────────────────────────────────────────────────────────────┘
```

### 5.4 关键决策

| 决策点 | 选择 | 理由 |
|-------|-----|-----|
| Mock框架 | MSW | 业界标准，与浏览器API一致，支持Service Worker |
| 响应策略 | 分层混合 | 静态+动态+懒加载，平衡性能与灵活性 |
| 权限模拟 | 完整实现 | 路由守卫+权限中间件，与真实系统一致 |
| 算法模拟 | 简化实现 | Demo场景下，简化排班算法可接受 |
| 错误模拟 | 可配置注入 | 便于测试各种错误场景 |

---

**结论**: 采用MSW框架，实现分层响应策略，在保证API一致性的同时优化性能，构建可维护的Mock API层。
