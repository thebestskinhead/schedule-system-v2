# 排班系统 V2 - 项目核心记忆文档

## 项目概述

**技术栈**: Vue 3 + Element Plus + Go + MySQL
**项目路径**: `/workspace/schedule-system-v2/`

---

## 目录结构

```
schedule-system-v2/
├── backend/                    # Go 后端
│   ├── cmd/server/main.go     # 入口
│   ├── configs/               # 配置文件
│   ├── internal/              # 内部代码
│   │   ├── model/            # 数据模型 (已创建)
│   │   ├── dao/              # 数据访问层 (待创建)
│   │   ├── service/          # 业务逻辑层 (待创建)
│   │   ├── handler/          # HTTP处理器 (待创建)
│   │   ├── middleware/       # 中间件 (待创建)
│   │   ├── router/           # 路由 (待创建)
│   │   ├── db/               # 数据库连接 (待创建)
│   │   └── utils/            # 工具函数 (待创建)
│   ├── scripts/              # SQL脚本 (已创建)
│   └── go.mod                # (已创建)
└── frontend/                  # Vue 前端
    ├── src/
    │   ├── main.js           # (已创建)
    │   ├── App.vue           # (待创建)
    │   ├── router/           # 路由 (待创建)
    │   ├── stores/           # Pinia状态 (待创建)
    │   ├── api/              # API接口 (待创建)
    │   ├── views/            # 页面 (待创建)
    │   └── components/       # 组件 (待创建)
    ├── package.json          # (已创建)
    └── vite.config.js        # (已创建)
```

---

## 数据库设计

### 表结构

| 表名 | 用途 | 关键字段 |
|------|------|---------|
| `system_config` | 系统配置 | config_key, config_value |
| `users` | 用户 | student_id, name, email, password, role(admin/user) |
| `availability` | 无课时间 | user_id, week(1-30), weekday(1-5), period(1-4) |
| `duty_records` | 值班记录 | week, weekday, period, user_id, status |
| `duty_counters` | 值班统计 | user_id, total_count |

### 核心索引
- `availability`: UNIQUE(user_id, week, weekday, period), INDEX(week, weekday, period, user_id)
- `duty_records`: INDEX(week), INDEX(user_id), INDEX(week, weekday, period)

---

## 数据模型 (已创建)

### User
```go
type User struct {
    ID, StudentID, Name, Email, Password, Role, IsActive
}
// Role: admin/user
```

### Availability
```go
type Availability struct {
    ID, UserID, Week, Weekday, Period
}
// 倒排索引设计: 时间段 -> 有空的人
```

### DutyRecord
```go
type DutyRecord struct {
    ID, Week, Weekday, Period, UserID, AssignedBy, Status
}
// Status: pending/confirmed/completed/cancelled
```

---

## 核心功能需求

### 1. 系统初始化
- 首次启动检测 `system_initialized` 配置
- 创建第一个管理员账号
- 设置学期等基础配置

### 2. 用户系统
- 注册/登录 (JWT)
- 角色: admin(可排班)/user(普通用户)
- 管理员可指定其他管理员

### 3. 无课表录入
**手动录入**:
- 输入: 星期 + 节次 + 周次列表
- 示例: 周一第1-2节, 周次[1,2,3,5,6]
- 支持多次录入、修改、删除

**自动导入** (可选):
- Cookie + 爬虫抓取教务系统

### 4. 排班功能
**权限**: 仅 admin
**输入**:
- 周次 (1-30)
- 排班星期 [1,2,3,4,5]
- 每格人数 (1-10)
- 每天节次 (1-4)

**算法策略**:
1. 查询该时间段所有有空的人
2. 按已排次数升序排序 (均衡)
3. 同次数随机打乱
4. 约束: 每人每周最多2次，每天最多1次
5. 选择前 N 个人

**输出**: 预览 + 确认保存

### 5. 查看与导出
- 按周查看排班表
- 个人查看自己的值班
- 导出 Excel (前端实现)

---

## API 设计规范

### 基础路径
- 前缀: `/api/v1`
- 认证: Header `Authorization: Bearer <token>`

### 接口列表

#### 系统
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/system/status` | 获取系统状态 |
| POST | `/system/init` | 系统初始化 |

#### 用户
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/user/register` | 注册 |
| POST | `/user/login` | 登录 |
| GET | `/user/profile` | 个人信息 |
| GET | `/users` | 用户列表 (admin) |

#### 无课时间
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/availability` | 录入无课时间 |
| GET | `/availability` | 查询自己的无课时间 |
| DELETE | `/availability` | 删除某条记录 |
| GET | `/availability/all` | 查看所有人 (admin) |

#### 排班
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/schedule/preview` | 预览排班 |
| POST | `/schedule/confirm` | 确认排班 |
| GET | `/schedule` | 查询排班结果 |
| GET | `/schedule/export` | 导出排班表 |

#### 值班记录
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/duty/my` | 我的值班 |
| PUT | `/duty/:id/status` | 更新状态 |

---

## 前端页面规划

### 路由设计
| 路径 | 页面 | 权限 |
|------|------|------|
| / | 首页/仪表盘 | 所有 |
| /login | 登录 | 未登录 |
| /register | 注册 | 未登录 |
| /init | 系统初始化 | 未初始化 |
| /availability | 我的无课表 | 登录 |
| /availability/input | 录入无课表 | 登录 |
| /schedule | 排班管理 | admin |
| /schedule/preview | 排班预览 | admin |
| /schedule/result | 排班结果 | 登录 |
| /duty/my | 我的值班 | 登录 |
| /admin/users | 用户管理 | admin |

### 组件规划
- `ScheduleGrid`: 排班表格组件 (5天×4节)
- `AvailabilityInput`: 无课时间录入组件
- `WeekSelector`: 周次选择器
- `UserSelect`: 用户选择器 (admin用)

---

## 排班算法核心逻辑

```go
// BuildSchedule 构建排班
func BuildSchedule(week int, days []int, needPerCell int) [][]Cell {
    grid := make([][]Cell, 5) // 5天
    
    for weekday := 1; weekday <= 5; weekday++ {
        for period := 1; period <= 4; period++ {
            // 1. 获取该时间段有空的人
            available := availabilityDAO.FindByTime(week, weekday, period)
            
            // 2. 获取已排次数
            counters := dutyCounterDAO.GetBatch(available)
            
            // 3. 排序: 按次数升序，同次数随机
            sortByCountThenRandom(available, counters)
            
            // 4. 应用约束过滤
            filtered := applyConstraints(available, week, weekday)
            
            // 5. 选择前 needPerCell 个
            selected := filtered[:min(needPerCell, len(filtered))]
            
            grid[weekday-1][period-1] = Cell{
                Weekday: weekday,
                Period: period,
                UserIDs: selected,
            }
        }
    }
    
    return grid
}
```

---

## 开发状态

### 已完成 ✅
- [x] 项目目录结构
- [x] 后端数据模型 (model/)
- [x] 数据库初始化脚本
- [x] 前端基础配置

### 待开发 ⏳
- [x] 后端 DAO 层
- [x] 后端 Service 层
- [x] 后端 Handler 层
- [x] 后端路由
- [x] 前端页面组件
- [x] 前端 API 封装
- [x] 前端状态管理

---

## 运行说明

### 启动后端
```bash
cd /workspace/schedule-system-v2/backend
go mod tidy
go run cmd/server/main.go
```

### 启动前端
```bash
cd /workspace/schedule-system-v2/frontend
npm install
npm run dev
```

### 访问
- 前端: http://localhost:3000
- 后端API: http://localhost:8080/api/v1

---

## 关键配置

### MySQL
- 数据库: `schedule_system_v2`
- 用户: `root`
- 密码: `Schedule@2024`

### JWT
- Secret: `schedule-system-secret-key`
- 过期: 168小时 (7天)

### 默认管理员
- 学号: `admin`
- 密码: `admin123456`

---

## 继续开发指令

当用户说"继续完善"时，按以下顺序开发：

1. **后端基础** (数据库连接、中间件、路由)
2. **后端业务** (DAO、Service、Handler)
3. **前端基础** (路由、API、状态管理)
4. **前端页面** (登录、无课表、排班、导出)

---

## 注意事项

1. **无课表存储**: 使用倒排索引，查询某时间段谁有空时效率高
2. **排班算法**: 注意约束条件（每周最多2次，每天最多1次）
3. **权限控制**: admin 才能排班，普通用户只能查看
4. **Excel导出**: 前端使用 xlsx 库实现，减轻服务器压力

---

*文档创建时间: 2026-02-27*
*最后更新: 2026-02-27*
