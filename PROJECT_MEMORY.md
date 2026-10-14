# 排班系统 V2 - 项目核心记忆文档

## 项目概述

**技术栈**: Vue 3 + Element Plus + Go + MySQL
**项目路径**: `/workspace/schedule-system-v2/`

---

## 目录结构

```
schedule-system-v2/
├── backend/                        # Go 后端
│   ├── cmd/server/main.go         # 程序入口
│   ├── configs/                    # 配置文件
│   │   ├── config.example.yaml    # 配置示例
│   │   └── config.yaml            # 实际配置（git忽略）
│   ├── internal/                   # 内部代码
│   │   ├── auth/                  # 权限定义
│   │   ├── handler/               # HTTP处理器
│   │   ├── service/               # 业务逻辑
│   │   ├── dao/                   # 数据访问层
│   │   ├── model/                 # 数据模型
│   │   ├── middleware/            # 中间件
│   │   ├── router/                # 路由
│   │   ├── db/                    # 数据库连接
│   │   └── utils/                 # 工具函数
│   ├── scripts/                    # SQL脚本
│   └── go.mod                     # Go模块
├── frontend-v2/                    # Vue 3 前端
│   ├── src/
│   │   ├── main.js               # 入口
│   │   ├── App.vue               # 根组件
│   │   ├── api/                  # API接口（6个模块）
│   │   ├── router/               # 路由
│   │   ├── stores/               # Pinia状态
│   │   ├── views/                # 页面
│   │   └── components/           # 组件
│   ├── package.json
│   └── vite.config.js            # Vite配置
├── docs/                           # 文档
│   ├── api.md                    # API接口文档
│   ├── permission.md             # 权限系统文档
│   ├── user-guide.md             # 用户手册
│   ├── features.md               # 功能介绍
│   ├── DEVELOPMENT_GUIDE.md      # 开发规范
│   └── IMPLEMENTATION_PLAN.md    # 实施计划
├── BUILD.md                        # 构建说明
├── README.md                       # 项目说明
├── Dockerfile                      # Docker构建
└── docker-compose.yml              # Docker编排
```

---

## 数据库设计

### 表结构

| 表名 | 用途 | 关键字段 |
|------|------|---------|
| `system_config` | 系统配置 | config_key, config_value |
| `users` | 用户 | student_id, name, email, password, role(admin/user), department, dept_role(dept_admin/dept_member), employment_status(pending/active/resigned), duty_count |
| `availability` | 无课时间 | user_id, week(1-30), weekday(1-5), period(1-4) |
| `duty_records` | 值班记录 | week, weekday, period, user_id, department, status(pending/confirmed/completed/cancelled) |
| `schedule_settings` | 排班设置（单例，`config_key='global'`） | current_week, auto_increment, need_per_cell, max_per_day, max_per_week, semester_start_date |
| `weekly_duty_assignments` | 每周分工 | week, department, weekday, is_assigned |
| `user_permissions_temp` | 临时权限 | user_id, permission, resource_type, resource_id, expires_at(可空=永久) |
| `export_templates` | 导出模板 | name, config, is_default |
| `smtp_config` | SMTP配置 | host, port, username, password, from_name, use_tls, is_active |
| `password_reset_tokens` | 密码重置令牌 | user_id, email, token, expires_at, is_used |
| `application_types` | 申请类型 | code, name, config |
| `applications` | 申请单 | applicant_id, type_code, status, data, current_level |
| `application_approvals` | 审批历史 | application_id, approver_id, action, level |
| `application_approvers` | 审批人配置 | type_code, level, role_type |
| `leave_requests` | 请假表 | application_id, user_id, leave_type, start_date, end_date, status |
| `user_invitations` | 预注册待激活 | student_id, name, email, department, dept_role, token(SHA-256哈希), token_expires_at, status |

### 核心索引
- `availability`: UNIQUE(user_id, week, weekday, period)
- `duty_records`: UNIQUE uk_duty_cell(week, weekday, period, user_id), INDEX idx_user_week(user_id, week)
- `schedule_settings`: UNIQUE uk_config_key(config_key)
- `user_invitations`: UNIQUE uk_token(token), UNIQUE uk_student_id(student_id)
- `leave_requests`: UNIQUE uk_application(application_id)

---

## 角色与权限体系

### 角色层级（4级）

| 角色 | 标识 | 说明 |
|------|------|------|
| 系统管理员 | `role=admin` | 所有权限，可设置系统角色、配置SMTP |
| 办公室管理员 | `department="办公室" + dept_role="dept_admin"` | 管理所有部门排班和用户，发布每周分工 |
| 部门管理员 | `dept_role="dept_admin"` | 管理本部门排班和成员，审批本部门权限申请 |
| 部门成员 | `dept_role="dept_member"` | 查看排班、编辑自己的无课表、申请临时权限 |

### 权限代码格式

使用冒号格式 `category:action:scope`，详见 `docs/permission.md`。

---

## API 概览

### 基础路径
- 前缀: `/api/v1`
- 认证: Header `Authorization: Bearer <token>`
- 统一响应: `{ code, message, data }`

### 接口分类

| 分类 | 数量 | 前缀 | 认证 |
|------|------|------|------|
| 系统安装 | 5 | `/system/` | 无 |
| SMTP/密码重置 | 4 | `/smtp/`, `/password/` | 无 |
| 用户认证 | 2 | `/user/` | 无 |
| 排班公开 | 2 | `/schedule/` | 无 |
| 用户个人 | 3 | `/user/` | JWT |
| 无课表 | 8 | `/availability/` | JWT |
| 爬虫 | 2 | `/crawler/` | JWT |
| 排班/值班 | 5 | `/schedule/`, `/duty/` | JWT |
| 管理员-用户 | 9 | `/admin/users/` | JWT+权限 |
| 管理员-排班 | 9 | `/admin/schedule/` | JWT+权限 |
| 管理员-分工 | 5 | `/admin/duty-assignments/` | JWT+权限 |
| 管理员-权限 | 4 | `/admin/temp-permissions/` | JWT+权限 |
| 管理员-模板 | 6 | `/admin/templates/` | JWT+权限 |
| 管理员-SMTP | 5 | `/admin/smtp/` | JWT+权限 |
| 管理员-站点 | 2 | `/admin/site/` | JWT+权限 |
| 公开数据 | 3 | `/departments`, `/permissions/`, `/temp-permissions/` | JWT |
| 权限申请 | 9 | `/applications/` | JWT |

**API 总计: ~78 个**（详见 `docs/api.md`）

---

## 前端页面

### 路由设计

| 路径 | 页面 | 权限 |
|------|------|------|
| `/` | 首页/仪表盘 | 所有 |
| `/login` | 登录 | 未登录 |
| `/invitations/activate` | 激活账号（邀请链接） | 未登录 |
| `/init` | 系统初始化 | 未初始化 |
| `/forgot-password` | 找回密码 | 未登录 |
| `/reset-password` | 重置密码 | 未登录 |
| `/availability` | 我的无课表 | 登录 |
| `/crawler-import` | 爬虫导入 | 登录 |
| `/schedule` | 排班管理 | 管理员 |
| `/schedule/result` | 排班结果 | 登录 |
| `/duty/my` | 我的值班 | 登录 |
| `/readme` | 使用说明 | 登录 |
| `/admin/users` | 用户管理 | 管理员 |
| `/admin/duty-assignments` | 每周分工 | 管理员 |
| `/admin/temp-permissions` | 权限申请 | 所有 |
| `/admin/smtp` | 邮件配置 | 系统管理员 |
| `/admin/semester` | 学期设置 | 管理员 |

### 前端 API 模块（6个）

| 模块 | 文件 | 函数数 |
|------|------|--------|
| 用户 | `api/user.js` | 15 |
| 排班 | `api/schedule.js` | 21 |
| 系统 | `api/system.js` | 17 |
| 无课表 | `api/availability.js` | 5 |
| 爬虫 | `api/crawler.js` | 4 |
| 申请 | `api/application.js` | 9 |

---

## 核心功能

### 1. 系统初始化
- 首次启动检测 `system_config` 中的 `system_initialized`
- 向导式安装：数据库连接 → 建表 → 创建管理员

### 2. 用户系统
- 登录 (JWT, 7天过期)
- 4级角色 + 临时权限的混合权限模型
- 部门体系：办公室、竞赛部、项目部、科普部

### 3. 无课表录入
- 手动录入：选择星期 + 节次 + 周次列表
- Cookie导入：通过教务系统Cookie自动抓取
- Excel导入：通过 .xls/.xlsx 文件批量导入

### 4. 排班功能
- 权限: 需要排班管理权限
- 输入: 周次、值班日期、每格人数、每天/每周上限、部门
- 算法: 倒排索引查询 → 按已排次数升序 → 同次数随机 → 约束过滤
- 输出: 预览(grid+conflicts+warnings) → 手动调整 → 确认保存

### 5. 每周分工
- 办公室管理员设置各部门本周值班日期
- 部门管理员查看本部门分工

### 6. 临时权限申请
- 普通用户提交申请 → 管理员审批 → 自动授权 → 到期失效
- 支持管理员直接授予

### 7. 导出
- 后端导出 Excel (模板系统)
- 前端导出 (xlsx 库)

---

## 排班算法核心逻辑

```go
// 约束条件
// 1. 每人每天最多 max_per_day 次（默认1次）
// 2. 每人每周最多 max_per_week 次（默认2次）
// 3. 按已排次数升序排序（均衡分配）
// 4. 同次数随机打乱

func BuildSchedule(week int, days []int, needPerCell int, maxPerDay int, maxPerWeek int) [][]Cell {
    grid := make([][]Cell, 5) // 5天

    for weekday := 1; weekday <= 5; weekday++ {
        for period := 1; period <= 4; period++ {
            // 1. 获取该时间段有空的人（倒排索引）
            available := availabilityDAO.FindByTime(week, weekday, period)

            // 2. 获取已排次数
            counters := dutyCounterDAO.GetBatch(available)

            // 3. 排序: 按次数升序，同次数随机
            sortByCountThenRandom(available, counters)

            // 4. 应用约束过滤
            filtered := applyConstraints(available, week, weekday, maxPerDay, maxPerWeek)

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

## 运行说明

### 启动开发环境

在项目根目录执行：

```powershell
.\dev.ps1
```

脚本会同时拉起电脑端与移动端的前端开发服务器，并由后端以代理模式转发页面请求。

### 访问
- 电脑端: http://localhost:8080/
- 移动端: http://localhost:8080/mobile/
- 后端 API: http://localhost:8080/api/v1

---

## 关键配置

### MySQL
- 数据库: `schedule_system_v2`
- 字符集: `utf8mb4`

### JWT
- Secret: 配置文件指定
- 过期: 168小时 (7天)

### 部门列表
- 办公室、竞赛部、项目部、科普部

---

## 注意事项

1. **无课表存储**: 使用倒排索引，查询某时间段谁有空时效率高
2. **排班算法**: 注意约束条件（每周最多2次，每天最多1次）
3. **权限控制**: 基于角色的固定权限 + 时间限定的临时权限
4. **Excel导出**: 支持后端模板导出和前端 xlsx 库导出
5. **配置文件**: `backend/configs/config.yaml` 被 gitignore，使用 `config.example.yaml` 作为模板

---

*文档创建时间: 2026-02-27*
*最后更新: 2026-03-25*

---

## 开发环境（本机）

- Go 工具链由 D 盘 usexxx 版本管理器提供：`D:\usexxx\go\versions\1.24.0\go\bin\go.exe`（另有 1.22.5）
- 本机无全局 go / docker；需要 Node 时也可在 `D:\usexxx\node` 下查找
- Git Bash 中直接使用：`export PATH="/d/usexxx/go/versions/1.24.0/go/bin:$PATH"`

*环境记录时间: 2026-08-12*

---

## 用户状态语义

用户任职状态为 `users.employment_status` 三态：

| 状态 | 含义 | 说明 |
|---|---|---|
| 待入职 pending | 已由管理员录入、尚未激活 | 不可登录；扫码登录提示联系管理员录入；经邀请邮件链接激活后转为在职 |
| 在职 active | 正常在职 | 可登录、参与排班 |
| 离职 resigned | 已离职 | 不可登录、不可参与排班，历史记录（值班/审批）保留 |

- 不存在「删除用户」操作：用户移除一律为「标记离职」，由存储过程 `sp_mark_user_resigned` / `sp_mark_users_resigned` 事务化处理，学号与邮箱长期保留、不可再次注册。
- 登录、扫码登录、鉴权中间件均按上述状态拦截。
