# 排班系统纯前端 Demo - 设计报告总结合并版

> 四个任务的详细设计报告汇总与实施路线图

---

## 目录

1. [项目背景与目标](#一项目背景与目标)
2. [任务1: 模拟展示数据生成](#二任务1-模拟展示数据生成)
3. [任务2: 模拟登录与认证](#三任务2-模拟登录与认证)
4. [任务3: 模拟API响应](#四任务3-模拟api响应)
5. [任务4: 本地存储模拟数据库](#五任务4-本地存储模拟数据库)
6. [整体架构设计](#六整体架构设计)
7. [实施路线图](#七实施路线图)
8. [风险评估与应对](#八风险评估与应对)

---

## 一、项目背景与目标

### 1.1 项目现状

排班系统是一个功能完整的企业级应用，包含：
- **前端**: Vue3 + Element Plus + Pinia
- **后端**: Go + Gin + JWT认证
- **数据库**: 关系型数据库，支持复杂业务逻辑

### 1.2 Demo目标

将全栈应用转化为**纯前端Demo**，实现：
1. ✅ 零后端依赖 - 无需启动任何后端服务
2. ✅ 功能完整 - 保留所有前端功能和交互
3. ✅ 数据真实 - 模拟真实业务数据
4. ✅ 多角色演示 - 支持切换不同权限角色

### 1.3 核心挑战

| 挑战 | 难度 | 关键点 |
|-----|-----|-------|
| API模拟 | ⭐⭐⭐⭐ | 38个接口需要Mock |
| 数据关联 | ⭐⭐⭐⭐⭐ | 多表关联查询 |
| 排班算法 | ⭐⭐⭐⭐ | 复杂业务逻辑前端实现 |
| 权限系统 | ⭐⭐⭐⭐ | 多层级权限校验 |

---

## 二、任务1: 模拟展示数据生成

### 2.1 核心需求

生成真实、关联、可演示的业务数据。

### 2.2 数据实体清单

```
核心实体 (6个主表):
├── Users (用户)         - 4个预设角色用户
├── Schedules (排班)      - 30周 × 5天 × 4节
├── Availability (无课表) - 30周 × 5天 × 4节 × 用户数
├── Applications (申请)   - 权限申请记录
├── TempPermissions (临时权限) - 授权记录
└── Settings (设置)      - 系统配置

数据规模预估:
├── 用户: 4-20人
├── 排班记录: 30周 × 5天 × 4节 × 2人 = 1200条
├── 无课表: 4人 × 30周 × 5天 × 4节 = 2400条
└── 其他: 约500条
```

### 2.3 生成策略

```javascript
// 三层数据体系
Layer 1: 场景层 (Scenario)
├── 新学期场景 - 空数据，待录入
├── 学期中场景 - 完整30周排班
└── 繁忙周场景 - 多部门同时排班

Layer 2: 生成层 (Generator)
├── 关联生成 - 排班依赖无课表
├── 联动更新 - 操作触发关联更新
└── 约束检查 - 保证数据合理性

Layer 3: 存储层 (Storage)
├── localStorage - 主要存储
├── IndexedDB - 大数据备选
└── Memory Cache - 热点缓存
```

### 2.4 实施要点

- **文件位置**: `mock/generators/`, `mock/seeders/`
- **关键技术**: 数据工厂模式、关联生成算法
- **预估工时**: 4-5小时

---

## 三、任务2: 模拟登录与认证

### 3.1 核心需求

实现无需后端的认证系统，支持多角色切换。

### 3.2 预设用户设计

```javascript
const PRESET_USERS = [
  {
    id: 1,
    student_id: 'admin',
    name: '系统管理员',
    role: 'admin',
    department: '办公室',
    dept_role: 'dept_admin',
    permissions: ['*']
  },
  {
    id: 2,
    name: '办公室管理员',
    department: '办公室',
    dept_role: 'dept_admin',
    permissions: ['schedule:manage:all', 'user:manage:all']
  },
  {
    id: 3,
    name: '部门管理员',
    department: '竞赛部',
    dept_role: 'dept_admin',
    permissions: ['schedule:manage:dept']
  },
  {
    id: 4,
    name: '普通成员',
    department: '竞赛部',
    dept_role: 'dept_member',
    permissions: ['schedule:view']
  }
]
```

### 3.3 认证流程

```
登录流程:
Login.vue → MockAuthService.login() → 验证预设用户 
→ 生成Token → localStorage存储 → 更新Store状态

角色切换:
RoleSwitcher → switchRole(userId) → 更新Token 
→ 热刷新状态 → 界面权限更新
```

### 3.4 权限计算

```javascript
// Store层权限计算 (复用现有逻辑)
const canManageAll = computed(() => {
  return isAdmin.value || 
         isOfficeAdmin.value ||
         hasTempPermission('user:manage:all')
})

const canManageDept = computed(() => {
  return isAdmin.value || 
         isOfficeAdmin.value || 
         isDeptAdmin.value ||
         hasTempPermission('schedule:manage:dept')
})
```

### 3.5 实施要点

- **文件位置**: `mock/auth/MockAuthService.js`
- **关键技术**: 模拟JWT、角色切换UI
- **预估工时**: 3-4小时

---

## 四、任务3: 模拟API响应

### 4.1 API清单 (38个接口)

```
分类统计:
├── 认证相关: 5个
├── 用户管理: 8个
├── 排班管理: 15个
├── 无课表: 5个
├── 权限系统: 5个
└── 系统配置: 5个
```

### 4.2 Mock方案

**推荐: MSW (Mock Service Worker)**

```javascript
// MSW Handler示例
rest.get('/api/v1/schedule', async (req, res, ctx) => {
  const week = req.url.searchParams.get('week')
  const schedules = await db.schedules.findByWeek(week)
  return res(ctx.json(success(schedules)))
})

rest.post('/api/v1/admin/schedule/preview', 
  requireAuth(
    requirePermission('schedule:manage:dept', 
      async (req, res, ctx) => {
        const settings = await req.json()
        const assignments = scheduleAlgorithm(settings)
        return res(ctx.json(success({ assignments })))
      }
    )
  )
)
```

### 4.3 响应策略

```
分层响应:
├── 静态缓存 - 部门列表、权限列表
├── 动态生成 - 根据参数实时计算
└── 懒加载 - 大数据量按需生成
```

### 4.4 实施要点

- **文件位置**: `mock/handlers/`, `mock/middleware/`
- **关键技术**: MSW、中间件链、延迟模拟
- **预估工时**: 5-6小时

---

## 五、任务4: 本地存储模拟数据库

### 5.1 存储架构

```
三层存储:
├── Hot (Memory)
│   ├── 当前用户
│   ├── 当前设置
│   └── 热点数据缓存
│
├── Warm (localStorage)
│   ├── 用户基础信息
│   ├── 近期排班
│   └── 配置数据
│
└── Cold (IndexedDB)
    ├── 历史排班
    └── 大量记录
```

### 5.2 数据库设计

```javascript
// 类SQL接口
class MockDatabase {
  // 查询
  async find(table, query)
  async findOne(table, query)
  async findById(table, id)
  
  // 写入
  async create(table, data)
  async update(table, id, data)
  async delete(table, id)
  
  // 高级查询
  query(table)
    .where('field', 'equals', value)
    .where('field', 'in', [values])
    .orderBy('field:asc')
    .limit(10)
    .execute()
}
```

### 5.3 Repository模式

```javascript
// 用户仓库
class UserRepository {
  async findByStudentId(id)
  async findByDepartment(dept)
  async findAvailableForSchedule(dept)
  async create(userData)
  async update(id, data)
}

// 排班仓库
class ScheduleRepository {
  async findByWeek(week, dept)
  async findByUserAndWeek(userId, week)
  async getStats(week, dept)
}
```

### 5.4 实施要点

- **文件位置**: `mock/database/`
- **关键技术**: StorageAdapter、QueryBuilder、Repository模式
- **预估工时**: 4-5小时

---

## 六、整体架构设计

### 6.1 系统架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                          Frontend Demo                               │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                     View Layer                                │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐        │  │
│  │  │  Home    │ │ Schedule │ │ Availability│ │ TempPermission│ │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘        │  │
│  └───────────────────────────┬──────────────────────────────────┘  │
│                              │                                       │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                     Store Layer (Pinia)                       │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐                      │  │
│  │  │userStore │ │scheduleStore│ │appStore  │                      │  │
│  │  └──────────┘ └──────────┘ └──────────┘                      │  │
│  └───────────────────────────┬──────────────────────────────────┘  │
│                              │                                       │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                     API Layer (原有)                          │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐        │  │
│  │  │ user.js  │ │schedule.js│ │availability│ │ system.js │        │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘        │  │
│  └───────────────────────────┬──────────────────────────────────┘  │
│                              │                                       │
│                              ▼                                       │
│  ╔══════════════════════════════════════════════════════════════╗  │
│  ║                     Mock Layer (新增)                         ║  │
│  ╠══════════════════════════════════════════════════════════════╣  │
│  ║  ┌────────────────────────────────────────────────────────┐  ║  │
│  ║  │              MSW Request Interceptor                    │  ║  │
│  ║  │     (拦截API请求，路由到Mock Handler)                    │  ║  │
│  ║  └────────────────────┬───────────────────────────────────┘  ║  │
│  ║                       │                                        ║  │
│  ║       ┌───────────────┼───────────────┐                      ║  │
│  ║       ▼               ▼               ▼                      ║  │
│  ║  ┌──────────┐   ┌──────────┐   ┌──────────┐                 ║  │
│  ║  │  Auth    │   │ Schedule │   │  User    │                 ║  │
│  ║  │ Handlers │   │ Handlers │   │ Handlers │                 ║  │
│  ║  └────┬─────┘   └────┬─────┘   └────┬─────┘                 ║  │
│  ║       │              │              │                        ║  │
│  ║       └──────────────┼──────────────┘                        ║  │
│  ║                      ▼                                        ║  │
│  ║  ┌────────────────────────────────────────────────────────┐  ║  │
│  ║  │              Mock Database (localStorage)               │  ║  │
│  ║  │  ┌────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐    │  ║  │
│  ║  │  │ Users  │ │Schedules │ │Availability│ │Permissions│    │  ║  │
│  ║  │  └────────┘ └──────────┘ └──────────┘ └──────────┘    │  ║  │
│  ║  └────────────────────────────────────────────────────────┘  ║  │
│  ╚══════════════════════════════════════════════════════════════╝  │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 6.2 文件结构

```
frontend-v2-demo/
├── mock/                              # Mock层 (新增)
│   ├── index.js                       # Mock入口
│   ├── browser.js                     # MSW浏览器集成
│   ├── 
│   ├── auth/                          # 认证
│   │   ├── MockAuthService.js
│   │   ├── preset-users.js
│   │   └── RoleSwitcher.vue
│   ├── 
│   ├── database/                      # 模拟数据库
│   │   ├── index.js
│   │   ├── adapters/
│   │   │   └── localStorage.js
│   │   ├── core/
│   │   │   ├── Database.js
│   │   │   └── QueryBuilder.js
│   │   ├── repositories/
│   │   │   ├── UserRepository.js
│   │   │   ├── ScheduleRepository.js
│   │   │   └── ...
│   │   └── models/
│   │       └── ...
│   ├── 
│   ├── handlers/                      # API Handlers
│   │   ├── index.js
│   │   ├── auth.handlers.js
│   │   ├── user.handlers.js
│   │   ├── schedule.handlers.js
│   │   └── ...
│   ├── 
│   ├── middleware/                    # 中间件
│   │   ├── auth.middleware.js
│   │   └── permission.middleware.js
│   ├── 
│   ├── generators/                    # 数据生成器
│   │   ├── user.generator.js
│   │   ├── schedule.generator.js
│   │   └── availability.generator.js
│   ├── 
│   ├── seeders/                       # 数据填充
│   │   └── DatabaseSeeder.js
│   └── 
│   └── utils/                         # 工具函数
│       ├── response.js
│       ├── delay.js
│       └── error.js
│
├── src/                               # 源代码
│   ├── api/                           # API层 (小修改)
│   │   └── request.js                 # 添加MSW初始化
│   ├── 
│   ├── components/                    # 组件
│   │   └── Layout.vue                 # 添加角色切换器
│   ├── 
│   ├── views/                         # 视图 (无需修改)
│   ├── stores/                        # Store (无需修改)
│   └── router/                        # 路由 (无需修改)
│
└── package.json                       # 添加msw依赖
```

---

## 七、实施路线图

### 7.1 阶段规划

```
总工期: 约16-20小时

Phase 1: 基础框架 (4h)
├── [1h] 项目初始化
│   ├── 安装MSW依赖
│   ├── 创建目录结构
│   └── 配置开发环境
├── [2h] Database核心
│   ├── StorageAdapter
│   ├── Database类
│   └── QueryBuilder
└── [1h] MSW集成
    ├── browser.js
    ├── 请求拦截
    └── 路由映射

Phase 2: 数据层 (5h)
├── [1h] 数据生成器
│   ├── UserGenerator
│   ├── ScheduleGenerator
│   └── AvailabilityGenerator
├── [2h] Repository层
│   ├── UserRepository
│   ├── ScheduleRepository
│   ├── AvailabilityRepository
│   └── ApplicationRepository
├── [1h] 数据填充
│   └── DatabaseSeeder
└── [1h] 关联数据生成
    └── 排班算法实现

Phase 3: API层 (5h)
├── [1h] 工具函数
│   ├── response.js
│   ├── delay.js
│   └── error.js
├── [1h] 中间件
│   ├── auth.middleware
│   └── permission.middleware
├── [2h] Handlers
│   ├── auth.handlers
│   ├── user.handlers
│   ├── schedule.handlers
│   └── availability.handlers
└── [1h] 剩余Handlers
    ├── permission.handlers
    ├── application.handlers
    └── system.handlers

Phase 4: 认证层 (3h)
├── [1h] MockAuthService
│   ├── login/logout
│   ├── token管理
│   └── 会话恢复
├── [1h] 多角色支持
│   ├── 预设用户
│   ├── 权限计算
│   └── 角色切换
└── [1h] UI集成
    ├── RoleSwitcher组件
    └── Layout.vue集成

Phase 5: 集成测试 (3h)
├── [1h] 功能测试
│   ├── 登录/切换
│   ├── 排班流程
│   └── 权限控制
├── [1h] 数据测试
│   ├── 数据持久化
│   ├── 关联查询
│   └── 大数据量
└── [1h] 优化完善
    ├── 延迟调整
    ├── 错误处理
    └── 演示数据调优
```

### 7.2 关键里程碑

| 里程碑 | 交付物 | 验收标准 |
|-------|-------|---------|
| M1 | Database层完成 | CRUD正常，查询正确 |
| M2 | API层完成 | 所有接口响应正常 |
| M3 | 认证层完成 | 可登录切换角色 |
| M4 | 完整Demo | 所有功能可用 |

---

## 八、风险评估与应对

### 8.1 风险清单

| 风险 | 概率 | 影响 | 应对策略 |
|-----|-----|-----|---------|
| localStorage容量超限 | 中 | 高 | 数据分区+IndexedDB备选 |
| 排班算法复杂度高 | 中 | 中 | 简化算法，满足Demo即可 |
| 多角色数据隔离复杂 | 低 | 中 | 单用户会话，切换时重置 |
| 浏览器兼容性 | 低 | 低 | 使用标准API，测试覆盖 |

### 8.2 应急预案

```
问题: 排班算法性能差
解决: 预生成排班结果，查询时直接返回

问题: localStorage超出5MB
解决: 启用IndexedDB作为备用存储

问题: 刷新后数据丢失
解决: 检查初始化逻辑，确保数据持久化
```

---

## 九、总结

### 9.1 设计亮点

1. **分层架构** - 清晰的Mock/Database/Repository分层
2. **渐进增强** - 从localStorage开始，按需升级
3. **复用最大化** - 原有业务代码几乎无需修改
4. **场景化数据** - 支持多场景一键切换

### 9.2 关键技术决策

| 决策 | 选择 | 理由 |
|-----|-----|-----|
| Mock框架 | MSW | 业界标准，与浏览器API一致 |
| 存储方案 | localStorage+IndexedDB | 兼容性与容量平衡 |
| 架构模式 | Repository | 清晰分层，易于测试 |
| 认证方案 | 模拟JWT+角色切换 | 零依赖，体验好 |

### 9.3 预期成果

- ✅ 完整的纯前端Demo
- ✅ 零后端依赖
- ✅ 多角色切换演示
- ✅ 数据持久化
- ✅ 功能完整可用

---

**附录**:
- 详细设计报告1: `DESIGN_REPORT_TASK1_DATA_GENERATION.md`
- 详细设计报告2: `DESIGN_REPORT_TASK2_AUTHENTICATION.md`
- 详细设计报告3: `DESIGN_REPORT_TASK3_API_MOCK.md`
- 详细设计报告4: `DESIGN_REPORT_TASK4_LOCAL_DATABASE.md`
