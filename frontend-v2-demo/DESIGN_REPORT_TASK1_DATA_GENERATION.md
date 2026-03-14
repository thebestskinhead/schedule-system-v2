# 任务1：模拟展示数据生成 - 详细设计报告

## 一、现状分析

### 1.1 现有数据结构调研

通过分析 API 层和视图层，识别出以下核心数据实体：

```
Entity Analysis from API Layer:
├── User (用户)
│   ├── id, student_id, name, email
│   ├── role (admin/user)
│   ├── department (办公室/竞赛部/项目部/科普部)
│   ├── dept_role (dept_admin/dept_member)
│   └── password
│
├── Schedule (排班记录)
│   ├── id, week, weekday (1-5), period (1-4)
│   ├── user_id, user_name
│   ├── status (pending/confirmed/completed/cancelled)
│   └── department
│
├── Availability (无课表)
│   ├── user_id
│   ├── week (1-30), weekday (1-5), period (1-4)
│   └── is_available (boolean)
│
├── ScheduleSettings (排班设置)
│   ├── current_week (1-30)
│   ├── auto_increment (boolean)
│   ├── need_per_cell (每时段人数)
│   ├── min_per_cell (最少人数)
│   ├── max_per_day (每天最多)
│   ├── max_per_week (每周最多)
│   └── export_title
│
├── DutyAssignment (每周分工)
│   ├── week
│   ├── department
│   ├── days[] (值班日期)
│   └── task (任务说明)
│
├── TempPermission (临时权限)
│   ├── id, user_id, permission
│   ├── resource_type (all/dept/user)
│   ├── resource_id
│   ├── expires_at
│   └── granted_by
│
├── Application (权限申请)
│   ├── id, application_no
│   ├── applicant_id, type_code
│   ├── status (pending/approved/rejected)
│   ├── content, data (JSON)
│   └── created_at
│
└── Template (排班模板)
    ├── id, name
    ├── department
    └── config (JSON配置)
```

### 1.2 现有数据生成方式

**当前状态**: 项目完全依赖后端 API 返回数据，前端无数据生成逻辑。

**问题识别**:
1. **零数据自主性** - 前端无法独立运行展示
2. **无演示数据** - 新用户看不到任何内容
3. **依赖外部状态** - 需要后端数据库配合
4. **测试困难** - 无法快速切换不同数据场景

### 1.3 视图层数据依赖分析

| 视图组件 | 主要数据依赖 | 数据复杂度 |
|---------|------------|-----------|
| Home.vue | 当前周次、本周排班、用户信息 | 中等 |
| Schedule.vue | 排班设置、分工安排、用户列表、排班预览 | 高 |
| Availability.vue | 用户无课表数据 (30周×5天×4节=600单元格) | 高 |
| ScheduleResult.vue | 多周次排班数据 | 中等 |
| TempPermission.vue | 权限列表、申请列表、统计数据 | 中等 |
| UserManagement.vue | 用户列表、部门列表 | 低 |
| MyDuty.vue | 个人值班记录 | 低 |

## 二、理想状态设计

### 2.1 数据生成架构

```
┌─────────────────────────────────────────────────────────────┐
│                    Data Generation Layer                     │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Generators  │  │  Factories   │  │   Seeders    │      │
│  │  (数据生成器) │  │  (数据工厂)  │  │  (数据填充)  │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
│         │                 │                 │              │
│         └─────────────────┼─────────────────┘              │
│                           ▼                                │
│  ┌─────────────────────────────────────────────────────┐  │
│  │              Mock Database (localStorage)            │  │
│  │  ┌────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ │  │
│  │  │ Users  │ │Schedules │ │Availability│ │Permissions│ │  │
│  │  └────────┘ └──────────┘ └──────────┘ └──────────┘ │  │
│  └─────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 数据生成策略

#### 策略1: 预设演示数据集 (Primary)
- **场景**: 系统首次加载、演示模式
- **数据**: 4个角色用户、30周排班数据、示例无课表
- **优点**: 数据一致、可预测、演示效果好

#### 策略2: 随机数据生成器 (Secondary)
- **场景**: 测试边界情况、性能测试
- **数据**: 根据规则随机生成的数据
- **优点**: 覆盖各种边界条件

#### 策略3: 智能推导生成 (Smart)
- **场景**: 基于用户操作自动生成关联数据
- **示例**: 用户创建排班 → 自动生成对应的值班记录
- **优点**: 数据关联性强、体验真实

### 2.3 核心数据生成规则

```javascript
// 1. 无课表生成规则
const generateAvailability = (userId) => {
  const availability = []
  for (let week = 1; week <= 30; week++) {
    for (let weekday = 1; weekday <= 5; weekday++) {
      for (let period = 1; period <= 4; period++) {
        // 模拟真实课表分布：约60%时段有课
        availability.push({
          user_id: userId,
          week,
          weekday,
          period,
          is_available: Math.random() > 0.6
        })
      }
    }
  }
  return availability
}

// 2. 排班生成规则（基于无课表）
const generateSchedule = (week, settings, availability) => {
  const schedule = []
  settings.days.forEach(weekday => {
    for (let period = 1; period <= settings.periods; period++) {
      // 获取该时段有空的用户
      const availableUsers = availability
        .filter(a => a.week === week && a.weekday === weekday && a.period === period && a.is_available)
        .map(a => a.user_id)
      
      // 随机选择 need_per_cell 个用户
      const selectedUsers = shuffle(availableUsers).slice(0, settings.need_per_cell)
      
      selectedUsers.forEach(userId => {
        schedule.push({
          week,
          weekday,
          period,
          user_id: userId,
          user_name: getUserName(userId),
          status: Math.random() > 0.3 ? 'confirmed' : 'pending',
          department: settings.department
        })
      })
    }
  })
  return schedule
}

// 3. 权限申请生成规则
const generateApplications = () => {
  return [
    {
      id: 1,
      application_no: generateAppNo(),
      applicant_id: 4,  // 普通成员
      type_code: 'temp_permission',
      status: 'pending',
      data: { permission: 'schedule:manage:dept', expiry_date: '2024-12-31' },
      content: '因部门活动需要临时管理权限',
      created_at: new Date().toISOString()
    }
  ]
}
```

## 三、现有结构与理想的差距

### 3.1 架构层面差距

| 维度 | 现有结构 | 理想结构 | 差距等级 |
|-----|---------|---------|---------|
| 数据层 | 无 | 完整Mock数据层 | ⭐⭐⭐⭐⭐ |
| 生成器 | 无 | 多策略生成器 | ⭐⭐⭐⭐⭐ |
| 持久化 | localStorage(token only) | 完整localStorage DB | ⭐⭐⭐⭐ |
| 关联性 | 无 | 数据关系维护 | ⭐⭐⭐⭐ |
| 可配置 | 无 | 生成参数可调 | ⭐⭐⭐ |

### 3.2 具体差距分析

```
Gap Analysis:

1. 数据缺失 (Critical)
   ├── 现状: 页面加载时所有 API 返回空/错误
   ├── 影响: 页面无法正常展示
   └── 解决: 构建完整Mock数据层

2. 关系缺失 (High)
   ├── 现状: 无用户-排班-无课表关联
   ├── 影响: 排班算法无法演示
   └── 解决: 建立数据关系和推导逻辑

3. 状态缺失 (Medium)
   ├── 现状: 无持久化数据存储
   ├── 影响: 刷新页面数据丢失
   └── 解决: localStorage 模拟数据库

4. 多样性缺失 (Low)
   ├── 现状: 单一固定数据
   ├── 影响: 无法演示边界情况
   └── 解决: 随机数据生成器
```

## 四、修改位置与实施方案

### 4.1 文件结构规划

```
frontend-v2-demo/
├── mock/                          # 新增: Mock数据层
│   ├── index.js                   # Mock入口，请求拦截
│   ├── database.js                # localStorage数据库封装
│   ├── generators/                # 数据生成器
│   │   ├── user.generator.js
│   │   ├── schedule.generator.js
│   │   ├── availability.generator.js
│   │   └── permission.generator.js
│   ├── factories/                 # 数据工厂
│   │   └── index.js
│   ├── seeders/                   # 数据填充
│   │   └── demo.seeder.js
│   └── data/                      # 预设数据
│       ├── demo-users.js
│       └── demo-schedules.js
│
└── src/
    └── api/
        └── request.js             # 修改: 添加Mock拦截
```

### 4.2 核心修改点

#### 修改点1: 请求拦截器 (src/api/request.js)

```javascript
// 添加环境判断和Mock拦截
import mockAPI from '../../mock/index.js'

const isMockMode = import.meta.env.VITE_MOCK === 'true'

// 在请求拦截器中
request.interceptors.request.use(
  async (config) => {
    if (isMockMode) {
      // 返回Mock响应，阻止真实请求
      return mockAPI.handle(config)
    }
    // ... 原有代码
  }
)
```

#### 修改点2: 数据库存取层 (mock/database.js)

```javascript
class MockDatabase {
  constructor() {
    this.prefix = 'mock_db_'
    this.initIfEmpty()
  }

  // 表定义
  tables = {
    users: [],
    schedules: {},  // 按周次索引
    availability: {}, // 按用户ID索引
    settings: {},
    applications: [],
    temp_permissions: [],
    duty_assignments: []
  }

  // CRUD操作
  get(table, id) { }
  find(table, query) { }
  insert(table, data) { }
  update(table, id, data) { }
  delete(table, id) { }
  
  // 关系查询
  findByUser(table, userId) { }
  findByWeek(table, week) { }
}
```

#### 修改点3: 生成器层 (mock/generators/*.js)

```javascript
// 统一生成器接口
class BaseGenerator {
  generate(count, options = {}) { }
  generateOne(options = {}) { }
}

class UserGenerator extends BaseGenerator {
  generateOne(options) {
    return {
      id: generateId(),
      student_id: generateStudentId(),
      name: generateName(),
      email: generateEmail(),
      department: options.department || randomDepartment(),
      dept_role: options.dept_role || randomRole(),
      role: options.role || 'user',
      password: '123456'
    }
  }
}
```

### 4.3 实施步骤

```
Phase 1: 基础框架 (2h)
├── 1.1 创建 mock/ 目录结构
├── 1.2 实现 Database 基础类
├── 1.3 实现请求拦截机制
└── 1.4 验证基础数据读写

Phase 2: 核心数据生成 (3h)
├── 2.1 实现 UserGenerator
├── 2.2 实现 AvailabilityGenerator (30周数据)
├── 2.3 实现 ScheduleGenerator (含排班算法)
└── 2.4 实现 PermissionGenerator

Phase 3: 数据关联与填充 (2h)
├── 3.1 创建 DemoSeeder
├── 3.2 建立数据关系
├── 3.3 预设演示场景数据
└── 3.4 实现数据重置功能

Phase 4: 集成与优化 (1h)
├── 4.1 集成到 request.js
├── 4.2 添加生成参数配置
├── 4.3 性能优化 (大数据量)
└── 4.4 测试验证
```

## 五、再次分析：核心需求与更优解

### 5.1 核心需求重定义

原始需求: "生成模拟展示数据"
深层需求分析:
1. **演示完整性** - 让用户看到系统的全部功能
2. **体验真实性** - 数据看起来是真实的、有关联的
3. **操作可持久** - 用户的操作可以被保存
4. **场景多样性** - 支持不同角色、不同数据场景

### 5.2 更优解决方案

#### 方案对比

| 方案 | 复杂度 | 真实感 | 灵活性 | 推荐度 |
|-----|-------|-------|-------|-------|
| A: 静态JSON数据 | 低 | 低 | 低 | ⭐⭐ |
| B: 随机生成器 | 中 | 中 | 高 | ⭐⭐⭐ |
| **C: 智能关联生成** | **中** | **高** | **高** | **⭐⭐⭐⭐⭐** |
| D: Service Worker拦截 | 高 | 高 | 中 | ⭐⭐⭐ |

#### 推荐方案 C: 智能关联生成

**核心思想**: 不只是生成孤立数据，而是构建一个"微型数据生态系统"

```javascript
// 智能关联示例
class SmartDataGenerator {
  generateScenario(scenarioType) {
    switch(scenarioType) {
      case 'new_semester':
        // 新学期场景：周次=1，空排班表，待填无课表
        return this.generateNewSemesterData()
      
      case 'mid_semester':
        // 学期中场景：已有多周排班，部分值班完成
        return this.generateMidSemesterData()
      
      case 'busy_week':
        // 繁忙周场景：多部门同时排班
        return this.generateBusyWeekData()
      
      case 'permission_pending':
        // 待审批场景：有待处理的权限申请
        return this.generatePendingApprovalData()
    }
  }

  // 数据变更联动
  onUserCreate(user) {
    // 创建用户 → 生成无课表 → 更新可排班人员
    this.generateAvailability(user.id)
    this.updateScheduleableUsers()
  }

  onScheduleConfirm(schedule) {
    // 确认排班 → 生成值班记录 → 通知用户
    this.generateDutyRecords(schedule)
    this.createNotifications(schedule)
  }
}
```

**优势**:
1. **场景化** - 一键切换不同演示场景
2. **自洽性** - 数据之间有逻辑关联
3. **响应式** - 操作触发数据联动
4. **教育性** - 帮助用户理解系统流程

### 5.3 最终推荐架构

```
┌────────────────────────────────────────────────────────────┐
│                    推荐架构: 三层数据体系                      │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  Layer 3: Scenario Layer (场景层)                          │
│  ┌─────────────────────────────────────────────────────┐  │
│  │  - 新学期场景  - 学期中场景  - 繁忙周场景  - 特殊场景  │  │
│  │  一键切换，重置整个数据状态                              │  │
│  └─────────────────────────────────────────────────────┘  │
│                           │                                │
│                           ▼                                │
│  Layer 2: Smart Generator Layer (智能生成层)                │
│  ┌─────────────────────────────────────────────────────┐  │
│  │  - 关联生成  - 联动更新  - 约束检查  - 一致性维护      │  │
│  │  确保数据之间的逻辑关系正确                              │  │
│  └─────────────────────────────────────────────────────┘  │
│                           │                                │
│                           ▼                                │
│  Layer 1: Storage Layer (存储层)                           │
│  ┌─────────────────────────────────────────────────────┐  │
│  │  - localStorage封装  - 索引管理  - 查询优化           │  │
│  │  高性能的本地数据存取                                    │  │
│  └─────────────────────────────────────────────────────┘  │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

### 5.4 关键决策点

| 决策 | 选择 | 理由 |
|-----|-----|-----|
| 数据生成时机 | 懒加载 + 缓存 | 避免一次性生成大量数据导致卡顿 |
| 数据更新策略 | 乐观更新 + 同步 | 提升用户体验，保证数据一致性 |
| 多用户模拟 | Token切换机制 | 简单高效，无需复杂会话管理 |
| 数据重置 | 场景重置 + 完全重置 | 满足不同演示需求 |
| 数据导出 | 支持JSON导出 | 便于分享和备份演示数据 |

---

**结论**: 采用"智能关联生成"方案，构建三层数据体系，既能满足演示需求，又能提供接近真实的使用体验。
