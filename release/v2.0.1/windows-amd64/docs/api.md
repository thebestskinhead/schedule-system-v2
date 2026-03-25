# API接口文档

> 本文档详细描述排班管理系统的所有API接口，包括请求参数、响应格式和错误码说明。

## 基础信息

- **基础URL**: `http://localhost:8080/api/v1`
- **认证方式**: JWT Token (Bearer Token)
- **Content-Type**: `application/json`（文件上传除外）
- **统一响应格式**:

```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

分页响应格式:

```json
{
  "code": 200,
  "message": "success",
  "data": { "list": [...], "total": 100 }
}
```

错误响应:

```json
{
  "code": 400,
  "message": "错误信息",
  "data": null
}
```

## 权限代码说明

系统使用冒号格式的权限代码 (`category:action:scope`)：

| 权限代码 | 名称 | 说明 | 适用范围 |
|----------|------|------|----------|
| `schedule:publish` | 设置每周分工 | 设置各部门本周值班日期 | 全局 |
| `schedule:manage:all` | 排班管理（全部）| 管理所有部门排班 | 全部部门 |
| `user:manage:all` | 用户管理（全部）| 管理所有用户 | 全部部门 |
| `schedule:manage:dept` | 排班管理（部门）| 管理本部门排班 | 本部门 |
| `user:manage:dept` | 用户管理（部门）| 管理本部门成员 | 本部门 |

## 错误码说明

| 错误码 | 说明 | 处理建议 |
|--------|------|----------|
| 200 | 成功 | - |
| 400 | 请求参数错误 | 检查请求参数格式 |
| 401 | 未授权 | Token无效或过期，重新登录 |
| 403 | 无权限 | 没有操作权限，联系管理员 |
| 404 | 资源不存在 | 检查请求的资源ID |
| 500 | 服务器内部错误 | 联系系统管理员 |

---

## 一、系统安装（公开，无需认证）

### 1. 获取安装状态

```http
GET /system/installed
```

**响应示例:**

```json
{
  "code": 200,
  "data": { "installed": true }
}
```

### 2. 测试数据库连接

```http
POST /system/test-db
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| host | string | 是 | 数据库主机 |
| port | int | 是 | 数据库端口 |
| user | string | 是 | 用户名 |
| password | string | 是 | 密码 |
| dbname | string | 是 | 数据库名 |
| charset | string | 否 | 字符集，默认utf8mb4 |

### 3. 检查数据库状态

```http
POST /system/check-db
```

**请求参数:** 同上

**响应示例:**

```json
{
  "code": 200,
  "data": { "empty": true, "tables": [], "message": "数据库为空" }
}
```

### 4. 初始化数据库表

```http
POST /system/init-tables
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| host | string | 是 | 数据库主机 |
| port | int | 是 | 数据库端口 |
| user | string | 是 | 用户名 |
| password | string | 是 | 密码 |
| dbname | string | 是 | 数据库名 |
| charset | string | 否 | 字符集 |
| force | bool | 否 | 是否强制重建 |

### 5. 创建管理员

```http
POST /system/create-admin
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| studentId | string | 是 | 管理员学号 |
| name | string | 是 | 姓名 |
| email | string | 是 | 邮箱 |
| password | string | 是 | 密码(最少6位) |
| department | string | 是 | 部门 |

**响应示例:**

```json
{
  "code": 200,
  "data": {
    "message": "管理员创建成功",
    "user": { "id": 1, "student_id": "admin", "name": "管理员" }
  }
}
```

---

## 二、SMTP 与密码重置（公开）

### 1. 检查SMTP配置

```http
GET /smtp/check
```

**响应示例:**

```json
{
  "code": 200,
  "data": { "configured": true }
}
```

### 2. 请求密码重置

```http
POST /password/reset-request
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 否 | 注册邮箱（与student_id至少填一个） |
| student_id | string | 否 | 学号（与email至少填一个） |

### 3. 验证重置Token

```http
GET /password/reset-verify?token=xxx
```

**响应示例:**

```json
{
  "code": 200,
  "data": { "email": "user@example.com" }
}
```

### 4. 重置密码

```http
POST /password/reset
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| token | string | 是 | 重置令牌 |
| password | string | 是 | 新密码(最少6位) |

---

## 三、用户认证（公开）

### 1. 用户注册

```http
POST /user/register
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| student_id | string | 是 | 学号 |
| name | string | 是 | 姓名 |
| email | string | 是 | 邮箱 |
| password | string | 是 | 密码(最少6位) |
| department | string | 是 | 部门 |

**响应示例:**

```json
{
  "code": 200,
  "message": "注册成功",
  "data": {
    "id": 1,
    "student_id": "2024001",
    "name": "张三",
    "email": "zhangsan@example.com",
    "department": "竞赛部",
    "role": "user",
    "dept_role": "dept_member"
  }
}
```

### 2. 用户登录

```http
POST /user/login
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| student_id | string | 否 | 学号（与email至少填一个） |
| email | string | 否 | 邮箱（与student_id至少填一个） |
| password | string | 是 | 密码 |

**响应示例:**

```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": 1,
      "student_id": "2024001",
      "name": "张三",
      "email": "zhangsan@example.com",
      "department": "办公室",
      "dept_role": "dept_admin",
      "role": "user"
    }
  }
}
```

---

## 四、排班公开查询（公开）

### 1. 获取当前周次

```http
GET /schedule/current-week
```

**响应示例:**

```json
{
  "code": 200,
  "data": { "current_week": 5, "auto_increment": true }
}
```

### 2. 获取当前周排班

```http
GET /schedule/current
```

**响应示例:**

```json
{
  "code": 200,
  "data": {
    "week": 5,
    "records": [
      {
        "id": 1,
        "weekday": 1,
        "period": 1,
        "user_id": 1,
        "user_name": "张三",
        "department": "竞赛部",
        "status": "confirmed"
      }
    ]
  }
}
```

---

## 五、用户个人信息（需认证）

> 以下接口均需在请求头中携带 `Authorization: Bearer <token>`

### 1. 获取个人信息

```http
GET /user/profile
```

**响应示例:**

```json
{
  "code": 200,
  "data": {
    "id": 1,
    "student_id": "2024001",
    "name": "张三",
    "email": "zhangsan@example.com",
    "department": "办公室",
    "dept_role": "dept_admin",
    "role": "user"
  }
}
```

### 2. 更新个人信息

```http
PUT /user/profile
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 否 | 姓名 |
| email | string | 否 | 邮箱 |

### 3. 修改密码

```http
POST /user/change-password
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| old_password | string | 是 | 旧密码 |
| new_password | string | 是 | 新密码(最少6位) |

---

## 六、无课表管理（需认证）

### 1. 获取我的无课表

```http
GET /availability
```

**响应示例:**

```json
{
  "code": 200,
  "data": [
    { "id": 1, "week": 1, "weekday": 1, "period": 1 },
    { "id": 2, "week": 2, "weekday": 1, "period": 1 },
    { "id": 3, "week": 3, "weekday": 1, "period": 1 }
  ]
}
```

### 2. 添加无课时间

```http
POST /availability
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| weekday | int | 是 | 星期(1-5) |
| period | int | 是 | 节次(1-4) |
| weeks | []int | 是 | 无课的周次列表 |

**请求示例:**

```json
{
  "weekday": 1,
  "period": 1,
  "weeks": [1, 2, 3, 5, 6]
}
```

### 3. 删除无课记录

```http
DELETE /availability
```

**请求参数 (body):**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 无课记录ID |

### 4. Cookie导入无课表

```http
POST /availability/import/cookie
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| cookies | string | 是 | 教务系统Cookie |
| semester | string | 是 | 学期标识 |

**响应示例:**

```json
{
  "code": 200,
  "data": {
    "task_id": "task_xxx",
    "status": "pending",
    "message": "导入任务已创建",
    "created_at": "2026-03-25T10:00:00Z"
  }
}
```

### 5. Excel导入无课表

```http
POST /availability/import/xls
Content-Type: multipart/form-data
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | 是 | Excel文件(.xls/.xlsx) |

**响应示例:**

```json
{
  "code": 200,
  "data": { "imported": 120, "message": "成功导入120条记录" }
}
```

### 6. 获取导入任务状态

```http
GET /availability/import/status?task_id=xxx
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| task_id | string | 否 | 任务ID（不传返回最新任务） |

### 7. 获取导入任务列表

```http
GET /availability/import/tasks
```

### 8. 查看所有人无课表（管理员）

```http
GET /admin/availability/all
Authorization: Bearer <token>
```

**权限要求:** `availability:view_all`

**响应示例:**

```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "user_id": 1,
      "user_name": "张三",
      "department": "竞赛部",
      "week": 1,
      "weekday": 1,
      "period": 1
    }
  ]
}
```

---

## 七、爬虫导入（需认证）

### 1. 爬虫导入

```http
POST /crawler/import
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| cookies | string | 是 | 教务系统Cookie |
| semester | string | 是 | 学期标识 |
| start_week | int | 否 | 起始周次(默认1) |
| end_week | int | 否 | 结束周次(默认30) |

**响应示例:**

```json
{
  "code": 200,
  "data": {
    "weeks_parsed": 20,
    "total_cells": 400,
    "available_cells": 200,
    "imported": 200,
    "message": "导入完成"
  }
}
```

### 2. 爬虫预览

```http
POST /crawler/preview
Authorization: Bearer <token>
```

**请求参数:** 同上

**响应示例:**

```json
{
  "code": 200,
  "data": {
    "preview": [...],
    "message": "预览模式（仅显示前2周）"
  }
}
```

---

## 八、排班管理（需认证）

### 1. 查看排班

```http
GET /schedule?week=1
Authorization: Bearer <token>
```

**查询参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| week | int | 是 | 周次(1-30) |

**响应示例:**

```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "week": 1,
      "weekday": 1,
      "period": 1,
      "user_id": 1,
      "user_name": "张三",
      "department": "竞赛部",
      "status": "confirmed"
    }
  ]
}
```

### 2. 我的值班

```http
GET /duty/my
Authorization: Bearer <token>
```

**响应示例:**

```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "week": 5,
      "weekday": 1,
      "period": 1,
      "status": "pending",
      "assigned_by": 2
    }
  ]
}
```

### 3. 更新值班状态

```http
PUT /duty/status
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| duty_id | int | 是 | 值班记录ID |
| status | string | 是 | 状态: `confirmed`/`completed`/`cancelled` |

### 4. 获取排班用户列表

```http
GET /users/for-schedule
Authorization: Bearer <token>
```

**权限要求:** `schedule:view:dept`

**响应示例:**

```json
{
  "code": 200,
  "data": [
    { "id": 1, "name": "张三", "student_id": "2024001", "department": "竞赛部" }
  ]
}
```

### 5. 查看本部门分工

```http
GET /duty-assignments/my-dept?week=1
Authorization: Bearer <token>
```

**查询参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| week | int | 是 | 周次(1-30) |

---

## 九、管理员 - 用户管理

> 以下接口均需管理员权限

### 1. 获取用户列表

```http
GET /admin/users
Authorization: Bearer <token>
```

**权限要求:** `user:manage` 或 `user:manage:dept`

**响应示例:**

```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "student_id": "2024001",
      "name": "张三",
      "email": "zhangsan@example.com",
      "department": "办公室",
      "dept_role": "dept_admin",
      "role": "user"
    }
  ]
}
```

### 2. 按部门筛选用户

```http
GET /admin/users/by-dept?department=竞赛部
Authorization: Bearer <token>
```

**权限要求:** `user:manage:dept`

### 3. 多条件筛选用户

```http
GET /admin/users/filter?departments=竞赛部,项目部&role=user&dept_role=dept_member
Authorization: Bearer <token>
```

**查询参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| departments | string | 否 | 部门列表(逗号分隔) |
| role | string | 否 | 系统角色: admin/user |
| dept_role | string | 否 | 部门角色: dept_admin/dept_member |

**权限要求:** `user:manage`

**响应示例:**

```json
{
  "code": 200,
  "data": { "total": 15, "users": [...] }
}
```

### 4. 创建用户

```http
POST /admin/users
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| student_id | string | 是 | 学号 |
| name | string | 是 | 姓名 |
| email | string | 是 | 邮箱 |
| department | string | 是 | 部门 |
| role | string | 是 | 系统角色: admin/user |
| dept_role | string | 是 | 部门角色: dept_admin/dept_member |
| password | string | 否 | 密码(最少6位，不填则使用默认密码) |

**权限要求:** `user:manage`

### 5. 更新用户

```http
PUT /admin/users/:id
Authorization: Bearer <token>
```

**路径参数:** `id` - 用户ID

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 否 | 姓名 |
| email | string | 否 | 邮箱 |
| department | string | 否 | 部门 |
| role | string | 否 | 系统角色 |
| dept_role | string | 否 | 部门角色 |

**权限要求:** `user:manage` 或 `user:manage:dept`

### 6. 删除用户

```http
DELETE /admin/users/:id
Authorization: Bearer <token>
```

**路径参数:** `id` - 用户ID

**权限要求:** `user:manage` 或 `user:manage:dept`

### 7. 设置用户系统角色

```http
PUT /admin/users/:id/role
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| role | string | 是 | 系统角色: admin/user |

**权限要求:** `user:set_role`

### 8. 设置用户部门

```http
PUT /admin/users/:id/department
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| department | string | 是 | 部门名称 |

**权限要求:** `user:manage`（handler内还要求系统管理员或办公室管理员）

### 9. 设置部门角色

```http
PUT /admin/users/:id/dept-role
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| dept_role | string | 是 | 部门角色: `dept_admin` 或 `dept_member` |

**权限要求:** `user:manage`（handler内还要求系统管理员或办公室管理员）

---

## 十、管理员 - 排班管理

### 1. 预览排班

```http
POST /admin/schedule/preview
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| week | int | 是 | 周次(1-30) |
| days | []int | 是 | 值班日期 [1,2,3,4,5] |
| periods | int | 是 | 节次数(1-4) |
| need_per_cell | int | 是 | 每时段需要人数(0-10) |
| min_per_cell | int | 否 | 每时段最少人数(默认0) |
| max_per_day | int | 否 | 每人每天最多排班次数(默认1) |
| max_per_week | int | 否 | 每人每周最多排班次数(默认2) |
| department | string | 是 | 排班部门 |

**响应示例:**

```json
{
  "code": 200,
  "data": {
    "week": 1,
    "grid": [
      [
        {
          "weekday": 1,
          "period": 1,
          "user_ids": [1, 3],
          "users": [
            { "id": 1, "name": "张三", "student_id": "2024001" },
            { "id": 3, "name": "李四", "student_id": "2024003" }
          ]
        }
      ]
    ],
    "conflicts": [],
    "warnings": []
  }
}
```

**权限要求:** `schedule:preview`（handler内还检查部门权限）

### 2. 确认排班

```http
POST /admin/schedule/confirm
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| week | int | 是 | 周次 |
| cells | array | 是 | 排班安排 |

**cells 结构:**

| 字段 | 类型 | 说明 |
|------|------|------|
| weekday | int | 星期(1-5) |
| period | int | 节次(1-4) |
| user_ids | []int | 用户ID列表 |

**请求示例:**

```json
{
  "week": 1,
  "cells": [
    { "weekday": 1, "period": 1, "user_ids": [1, 3] },
    { "weekday": 1, "period": 2, "user_ids": [2, 4] }
  ]
}
```

**权限要求:** `schedule:confirm`

### 3. 手动更新排班

```http
POST /admin/schedule/update
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| week | int | 是 | 周次 |
| weekday | int | 是 | 星期(1-5) |
| period | int | 是 | 节次(1-4) |
| add_user_ids | []int | 否 | 添加的用户ID列表 |
| remove_user_ids | []int | 否 | 移除的用户ID列表 |

**权限要求:** `schedule:edit`

### 4. 获取排班设置

```http
GET /admin/schedule/settings
Authorization: Bearer <token>
```

**权限要求:** `schedule:settings`

**响应示例:**

```json
{
  "code": 200,
  "data": {
    "id": 1,
    "admin_id": 1,
    "current_week": 5,
    "auto_increment": true,
    "need_per_cell": 2,
    "min_per_cell": 0,
    "max_per_day": 1,
    "max_per_week": 2,
    "export_title": "第{week}周排班表",
    "semester_start_date": "2026-02-24"
  }
}
```

### 5. 保存排班设置

```http
POST /admin/schedule/settings
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| current_week | int | 否 | 当前周次 |
| auto_increment | bool | 否 | 自动递增 |
| need_per_cell | int | 否 | 每时段人数 |
| min_per_cell | int | 否 | 最少人数 |
| max_per_day | int | 否 | 每天最多 |
| max_per_week | int | 否 | 每周最多 |
| export_title | string | 否 | 导出标题模板 |
| semester_start_date | string | 否 | 学期起始日期(YYYY-MM-DD) |

**权限要求:** `schedule:settings`

### 6. 更新当前周次

```http
POST /admin/schedule/current-week
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| current_week | int | 是 | 当前周次(1-30) |
| auto_increment | bool | 是 | 是否自动递增 |

**权限要求:** `schedule:settings`

### 7. 学期起始日

#### 获取学期起始日

```http
GET /admin/schedule/semester-start
Authorization: Bearer <token>
```

**响应示例:**

```json
{
  "code": 200,
  "data": {
    "semester_start_date": "2026-02-24",
    "current_week": 5
  }
}
```

#### 更新学期起始日

```http
POST /admin/schedule/semester-start
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| semester_start_date | string | 是 | 学期起始日期(YYYY-MM-DD) |

**响应示例:**

```json
{
  "code": 200,
  "data": { "current_week": 5 }
}
```

### 8. 导出排班表

```http
POST /admin/schedule/export
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| week | int | 是 | 周次 |
| template_id | int | 否 | 导出模板ID |
| department | string | 否 | 部门 |
| custom_title | string | 否 | 自定义标题 |

**权限要求:** `schedule:export`

**响应:** 直接返回 Excel 文件流 (`application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`)

---

## 十一、管理员 - 每周分工

### 1. 发布分工

```http
POST /admin/duty-assignments
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| week | int | 是 | 周次(1-30) |
| assignments | array | 是 | 分工安排 |

**assignments 结构:**

| 字段 | 类型 | 说明 |
|------|------|------|
| department | string | 部门名称 |
| weekday | int | 星期(1-5) |
| is_assigned | bool | 是否安排值班 |

**请求示例:**

```json
{
  "week": 5,
  "assignments": [
    { "department": "办公室", "weekday": 1, "is_assigned": true },
    { "department": "办公室", "weekday": 2, "is_assigned": false },
    { "department": "竞赛部", "weekday": 1, "is_assigned": false },
    { "department": "竞赛部", "weekday": 2, "is_assigned": true }
  ]
}
```

**权限要求:** `schedule:publish`（handler内还要求系统管理员或办公室管理员）

### 2. 查看分工

```http
GET /admin/duty-assignments?week=1
Authorization: Bearer <token>
```

**权限要求:** `schedule:view:all`

### 3. 查看分工视图

```http
GET /admin/duty-assignments/view?week=1
Authorization: Bearer <token>
```

**响应示例:**

```json
{
  "code": 200,
  "data": {
    "week": 1,
    "departments": [
      { "department": "办公室", "weekdays": [1, 3, 5] },
      { "department": "竞赛部", "weekdays": [2, 4] }
    ]
  }
}
```

### 4. 更新单条分工

```http
PUT /admin/duty-assignments
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 分工记录ID |
| is_assigned | bool | 是 | 是否安排值班 |

**权限要求:** `schedule:publish`（handler内还要求系统管理员或办公室管理员）

### 5. 删除分工

```http
DELETE /admin/duty-assignments/:id
Authorization: Bearer <token>
```

**路径参数:** `id` - 分工记录ID

**权限要求:** `schedule:publish`（handler内还要求系统管理员或办公室管理员）

---

## 十二、管理员 - 临时权限

### 1. 授予临时权限

```http
POST /admin/temp-permissions
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_ids | []int | 是 | 用户ID数组（支持批量） |
| permission | string | 是 | 权限代码 |
| resource_type | string | 是 | `all`/`dept`/`user` |
| resource_id | int | 否 | 资源ID |
| expires_at | datetime | 是 | 过期时间(ISO 8601) |
| reason | string | 否 | 授权原因 |

**权限要求:** `user:manage:dept`

### 2. 获取所有临时权限

```http
GET /admin/temp-permissions
Authorization: Bearer <token>
```

**权限要求:** `user:manage:dept`

**响应示例:**

```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "user_id": 2,
      "user_name": "张三",
      "user_department": "竞赛部",
      "permission": "schedule:manage:dept",
      "permission_name": "排班管理(部门)",
      "granted_by_name": "管理员",
      "expires_at": "2026-12-31T23:59:59Z",
      "is_expired": false
    }
  ]
}
```

### 3. 撤销临时权限

```http
DELETE /admin/temp-permissions/:id
Authorization: Bearer <token>
```

**路径参数:** `id` - 权限记录ID

**权限要求:** `user:manage:dept`

### 4. 清理过期权限

```http
POST /admin/temp-permissions/cleanup
Authorization: Bearer <token>
```

**权限要求:** `system:admin`

---

## 十三、管理员 - 模板管理

### 1. 获取模板列表

```http
GET /admin/templates
Authorization: Bearer <token>
```

**权限要求:** `template:view`

### 2. 获取单个模板

```http
GET /admin/templates/:id
Authorization: Bearer <token>
```

**路径参数:** `id` - 模板ID

**权限要求:** `template:view`

### 3. 创建模板

```http
POST /admin/templates
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 模板名称 |
| description | string | 否 | 模板描述 |
| config | object | 是 | 模板配置 |
| is_default | bool | 否 | 是否为默认模板 |

**权限要求:** `template:edit`

### 4. 更新模板

```http
PUT /admin/templates
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 模板ID |
| name | string | 否 | 模板名称 |
| description | string | 否 | 模板描述 |
| config | object | 否 | 模板配置 |
| is_default | bool | 否 | 是否为默认模板 |

**权限要求:** `template:edit`

### 5. 删除模板

```http
DELETE /admin/templates/:id
Authorization: Bearer <token>
```

**路径参数:** `id` - 模板ID

**权限要求:** `template:edit`

### 6. 获取占位符帮助

```http
GET /admin/templates/placeholders
Authorization: Bearer <token>
```

**权限要求:** `template:view`

---

## 十四、管理员 - SMTP 配置

### 1. 获取SMTP配置列表

```http
GET /admin/smtp/configs
Authorization: Bearer <token>
```

**权限要求:** `system:admin`

### 2. 获取当前活跃SMTP配置

```http
GET /admin/smtp/config
Authorization: Bearer <token>
```

**权限要求:** `system:admin`

**响应示例:**

```json
{
  "code": 200,
  "data": {
    "id": 1,
    "host": "smtp.example.com",
    "port": 465,
    "username": "noreply@example.com",
    "password": "***",
    "from": "排班系统",
    "from_email": "noreply@example.com",
    "use_tls": true,
    "use_ssl": true,
    "is_active": true
  }
}
```

### 3. 保存SMTP配置

```http
POST /admin/smtp/config
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| host | string | 是 | SMTP服务器 |
| port | int | 是 | 端口(1-65535) |
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |
| from | string | 是 | 发件人名称 |
| from_email | string | 是 | 发件人邮箱 |
| use_tls | bool | 是 | 是否使用TLS |
| use_ssl | bool | 是 | 是否使用SSL |
| is_active | bool | 否 | 是否设为活跃配置 |

**权限要求:** `system:admin`

### 4. 删除SMTP配置

```http
DELETE /admin/smtp/config/:id
Authorization: Bearer <token>
```

**路径参数:** `id` - 配置ID

**权限要求:** `system:admin`

### 5. 测试SMTP发送

```http
POST /admin/smtp/test
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| to | string | 是 | 测试收件人邮箱 |

**权限要求:** `system:admin`

---

## 十五、管理员 - 站点配置

### 1. 获取站点配置

```http
GET /admin/site/config
Authorization: Bearer <token>
```

**权限要求:** `system:admin`

**响应示例:**

```json
{
  "code": 200,
  "data": { "domain": "https://schedule.example.com" }
}
```

### 2. 保存站点配置

```http
POST /admin/site/config
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| domain | string | 是 | 站点域名 |

**权限要求:** `system:admin`

---

## 十六、公开数据接口（需登录）

### 1. 获取部门列表

```http
GET /departments
Authorization: Bearer <token>
```

**响应示例:**

```json
{
  "code": 200,
  "data": {
    "departments": ["办公室", "竞赛部", "项目部", "科普部"]
  }
}
```

### 2. 获取权限列表

```http
GET /permissions/list
Authorization: Bearer <token>
```

**响应示例:**

```json
{
  "code": 200,
  "data": [
    {
      "code": "schedule:manage:all",
      "name": "排班管理(全部)",
      "description": "管理所有部门的排班"
    }
  ]
}
```

### 3. 获取我的临时权限

```http
GET /temp-permissions/my
Authorization: Bearer <token>
```

**响应示例:**

```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "permission": "schedule:manage:dept",
      "permission_name": "排班管理(部门)",
      "resource_type": "dept",
      "resource_name": "全部部门",
      "expires_at": "2026-12-31T23:59:59Z",
      "days_left": 30
    }
  ]
}
```

---

## 十七、权限申请系统（需登录）

### 1. 获取申请类型

```http
GET /application/types
Authorization: Bearer <token>
```

### 2. 获取可申请权限列表

```http
GET /application/permissions/available
Authorization: Bearer <token>
```

**响应示例:**

```json
{
  "code": 200,
  "data": [
    {
      "key": "schedule:manage:dept",
      "name": "排班管理(部门)",
      "description": "管理本部门排班"
    },
    {
      "key": "user:manage:dept",
      "name": "用户管理(部门)",
      "description": "管理本部门成员"
    }
  ]
}
```

### 3. 创建权限申请

```http
POST /applications
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 是 | 申请类型: `temp_permission` |
| data | object | 是 | 申请数据 |
| data.permission | string | 是 | 申请的权限代码 |
| data.expiry_date | datetime | 是 | 期望到期时间 |
| data.reason | string | 是 | 申请原因 |
| reason | string | 否 | 申请说明 |

**请求示例:**

```json
{
  "type": "temp_permission",
  "data": {
    "permission": "schedule:manage:dept",
    "expiry_date": "2026-12-31T23:59:59Z",
    "reason": "因部门活动需要临时管理权限"
  },
  "reason": "申请临时排班管理权限"
}
```

### 4. 获取我的申请列表

```http
GET /applications/my?page=1&page_size=10
Authorization: Bearer <token>
```

**查询参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| status | string | 否 | 状态过滤 |
| page | int | 否 | 页码(默认1) |
| page_size | int | 否 | 每页数量(默认10，最大100) |

**响应示例:**

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "application_no": "APP202603250001",
        "type_code": "temp_permission",
        "status": "pending",
        "content": "申请临时排班管理权限",
        "applicant_name": "张三",
        "created_at": "2026-03-25T10:00:00Z"
      }
    ],
    "total": 1
  }
}
```

### 5. 获取申请详情

```http
GET /applications/:id
Authorization: Bearer <token>
```

**路径参数:** `id` - 申请ID

### 6. 取消申请

```http
POST /applications/:id/cancel
Authorization: Bearer <token>
```

**路径参数:** `id` - 申请ID（只能取消自己的待审批申请）

### 7. 获取待审批列表

```http
GET /applications/pending?page=1&page_size=10
Authorization: Bearer <token>
```

**查询参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码(默认1) |
| page_size | int | 否 | 每页数量(默认10) |

### 8. 处理审批

```http
POST /applications/:id/approve
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| action | string | 是 | `approve` / `reject` / `transfer` / `comment` |
| comment | string | 否 | 审批意见 |

**请求示例:**

```json
{
  "action": "approve",
  "comment": "同意申请，请合理使用权限"
}
```

### 9. 获取申请统计

```http
GET /applications/stats
Authorization: Bearer <token>
```

**响应示例:**

```json
{
  "code": 200,
  "data": {
    "my_applications": {
      "pending": 2,
      "approved": 5,
      "rejected": 1
    },
    "pending_approval": 3
  }
}
```

---

*文档最后更新: 2026-03-25*
