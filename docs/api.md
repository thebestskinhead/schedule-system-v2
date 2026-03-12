# API接口文档

> 本文档详细描述排班管理系统的所有API接口，包括请求参数、响应格式和错误码说明。

## 基础信息

- **基础URL**: `http://localhost:8080/api/v1`
- **认证方式**: JWT Token (Bearer Token)
- **Content-Type**: `application/json`

## 认证相关

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
| password | string | 是 | 密码(6-20位) |

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
    "role": "user"
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
| student_id | string | 是 | 学号 |
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
      "department": "办公室",
      "dept_role": "dept_admin"
    }
  }
}
```

### 3. 获取用户信息

```http
GET /user/profile
Authorization: Bearer <token>
```

**响应示例:**

```json
{
  "code": 200,
  "message": "success",
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

## 排班管理

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
| need_per_cell | int | 是 | 每时段需要人数 |
| min_per_cell | int | 否 | 每时段最少人数(默认0) |
| max_per_day | int | 否 | 每人每天最多排班次数(默认1) |
| max_per_week | int | 否 | 每人每周最多排班次数(默认2) |
| department | string | 是 | 排班部门 |

**响应示例:**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "week": 1,
    "assignments": [
      {
        "weekday": 1,
        "period": 1,
        "users": [
          {"id": 1, "name": "张三", "student_id": "2024001"}
        ]
      }
    ],
    "stats": {
      "total_slots": 20,
      "filled_slots": 18,
      "conflicts": []
    }
  }
}
```

**权限要求:** `schedule:manage:dept` 或 `schedule:manage:all`

### 2. 确认排班

```http
POST /admin/schedule/confirm
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| week | int | 是 | 周次 |
| assignments | array | 是 | 排班安排 |
| department | string | 是 | 部门 |

**响应示例:**

```json
{
  "code": 200,
  "message": "排班确认成功",
  "data": {
    "week": 1,
    "total_assigned": 20
  }
}
```

### 3. 获取排班设置

```http
GET /admin/schedule/settings
Authorization: Bearer <token>
```

**响应示例:**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "current_week": 1,
    "auto_increment": true,
    "need_per_cell": 2,
    "min_per_cell": 0,
    "max_per_day": 1,
    "max_per_week": 2,
    "export_title": "第{week}周排班表"
  }
}
```

### 4. 保存排班设置

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

### 5. 查看排班

```http
GET /schedule?week=1
Authorization: Bearer <token>
```

**查询参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| week | int | 是 | 周次 |
| department | string | 否 | 部门(默认用户所在部门) |

**响应示例:**

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "week": 1,
      "weekday": 1,
      "period": 1,
      "user_id": 1,
      "user_name": "张三",
      "status": "confirmed"
    }
  ]
}
```

## 用户管理

### 1. 获取用户列表(管理权限)

```http
GET /admin/users
Authorization: Bearer <token>
```

**权限要求:** `user:manage` 或 `user:manage:all`

**响应示例:**

```json
{
  "code": 200,
  "message": "success",
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

### 2. 获取排班用户列表

```http
GET /users/for-schedule
Authorization: Bearer <token>
```

**说明:** 用于排班页面选择用户，只需要排班查看权限

**权限要求:** `schedule:view:dept`

### 3. 设置用户部门

```http
PUT /admin/users/:id/department
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| department | string | 是 | 部门名称 |

**权限要求:** 系统管理员或办公室管理员

### 4. 设置部门角色

```http
PUT /admin/users/:id/dept-role
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| dept_role | string | 是 | `dept_admin` 或 `dept_member` |

## 无课表管理

### 1. 获取我的无课表

```http
GET /availability
Authorization: Bearer <token>
```

**响应示例:**

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "weekday": 1,
      "period": 1,
      "is_available": false
    }
  ]
}
```

### 2. 更新无课表

```http
PUT /availability
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| availability | array | 是 | 无课表数据 |

**示例:**

```json
{
  "availability": [
    {"weekday": 1, "period": 1, "is_available": false},
    {"weekday": 1, "period": 2, "is_available": true}
  ]
}
```

### 3. 导入无课表(爬虫)

```http
POST /crawler/import
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| cookie | string | 是 | 教务系统Cookie |
| year | string | 是 | 学年 |
| term | string | 是 | 学期 |

## 值班管理

### 1. 获取我的值班

```http
GET /duty/my?week=1
Authorization: Bearer <token>
```

**响应示例:**

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "week": 1,
      "weekday": 1,
      "period": 1,
      "status": "pending",
      "assigned_by": 2
    }
  ]
}
```

### 2. 更新值班状态

```http
PUT /duty/status
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| duty_id | int | 是 | 值班ID |
| status | string | 是 | `confirmed` 或 `completed` |

## 临时权限管理

### 1. 获取权限列表

```http
GET /permissions/list
Authorization: Bearer <token>
```

**响应示例:**

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "code": "schedule:manage:all",
      "name": "排班管理(全部)",
      "description": "管理所有部门的排班"
    }
  ]
}
```

### 2. 授予临时权限

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
| expires_at | datetime | 是 | 过期时间 |
| reason | string | 否 | 授权原因 |

**权限要求:** `user:manage:dept`

### 3. 获取我的临时权限

```http
GET /temp-permissions/my
Authorization: Bearer <token>
```

**响应示例:**

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "permission": "schedule:manage:dept",
      "permission_name": "排班管理(部门)",
      "resource_type": "dept",
      "resource_name": "全部部门",
      "expires_at": "2024-12-31T23:59:59Z",
      "days_left": 30
    }
  ]
}
```

### 4. 获取所有临时权限（管理员）

```http
GET /admin/temp-permissions
Authorization: Bearer <token>
```

**响应示例:**

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "user_id": 2,
      "user_name": "张三",
      "permission": "schedule:manage:dept",
      "permission_name": "排班管理(部门)",
      "granted_by_name": "管理员",
      "expires_at": "2024-12-31T23:59:59Z",
      "is_expired": false
    }
  ]
}
```

**权限要求:** `user:manage:dept`

### 5. 撤销临时权限

```http
DELETE /admin/temp-permissions/:id
Authorization: Bearer <token>
```

**权限要求:** `user:manage:dept`

---

## 权限申请系统

### 1. 获取可申请权限列表

```http
GET /application/permissions/available
Authorization: Bearer <token>
```

**响应示例:**

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "key": "duty_manage",
      "name": "值班管理",
      "description": "管理值班安排和记录"
    },
    {
      "key": "schedule_manage",
      "name": "排班管理",
      "description": "创建和修改排班"
    }
  ]
}
```

### 2. 创建权限申请

```http
POST /applications
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 是 | 申请类型：`temp_permission` |
| data | object | 是 | 申请数据 |
| data.permission | string | 是 | 申请的权限代码 |
| data.expiry_date | datetime | 是 | 期望到期时间 |
| data.reason | string | 是 | 申请原因 |
| reason | string | 是 | 申请说明 |

**请求示例:**

```json
{
  "type": "temp_permission",
  "data": {
    "permission": "schedule_manage",
    "expiry_date": "2024-12-31T23:59:59Z",
    "reason": "因部门活动需要临时管理权限"
  },
  "reason": "申请临时排班管理权限"
}
```

**权限要求:** 登录用户

### 3. 获取我的申请列表

```http
GET /applications/my?page=1&page_size=10
Authorization: Bearer <token>
```

**响应示例:**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "application_no": "APP202403120001",
        "type_code": "temp_permission",
        "status": 0,
        "content": "申请临时排班管理权限",
        "created_at": "2024-03-12T10:00:00Z"
      }
    ],
    "total": 1
  }
}
```

### 4. 获取待审批列表

```http
GET /applications/pending?page=1&page_size=10
Authorization: Bearer <token>
```

**响应示例:**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "application_no": "APP202403120001",
        "applicant_name": "张三",
        "department": "竞赛部",
        "type_code": "temp_permission",
        "content": "申请临时排班管理权限",
        "created_at": "2024-03-12T10:00:00Z"
      }
    ],
    "total": 1
  }
}
```

**权限要求:** `user:manage:dept`

### 5. 处理审批

```http
POST /applications/:id/approve
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| action | string | 是 | `approve` 或 `reject` |
| comment | string | 否 | 审批意见 |

**请求示例:**

```json
{
  "action": "approve",
  "comment": "同意申请，请合理使用权限"
}
```

**权限要求:** `user:manage:dept`

### 6. 取消申请

```http
POST /applications/:id/cancel
Authorization: Bearer <token>
```

**说明:** 只能取消自己提交的待审批申请

### 7. 获取申请统计

```http
GET /applications/stats
Authorization: Bearer <token>
```

**响应示例:**

```json
{
  "code": 200,
  "message": "success",
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

## 每周分工

### 1. 发布分工

```http
POST /admin/duty-assignments
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| week | int | 是 | 周次 |
| assignments | array | 是 | 分工安排 |

**示例:**

```json
{
  "week": 1,
  "assignments": [
    {"department": "办公室", "days": [1, 3, 5]},
    {"department": "竞赛部", "days": [2, 4]}
  ]
}
```

**权限要求:** `schedule:publish`

### 2. 查看分工

```http
GET /admin/duty-assignments?week=1
Authorization: Bearer <token>
```

## SMTP配置

### 1. 获取SMTP配置

```http
GET /admin/smtp/config
Authorization: Bearer <token>
```

**权限要求:** 系统管理员

### 2. 保存SMTP配置

```http
POST /admin/smtp/config
Authorization: Bearer <token>
```

**请求参数:**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| host | string | 是 | SMTP服务器 |
| port | int | 是 | 端口 |
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |
| from | string | 是 | 发件人名称 |
| from_email | string | 是 | 发件人邮箱 |
| use_tls | bool | 是 | 是否使用TLS |

## 错误码说明

| 错误码 | 说明 | 处理建议 |
|--------|------|----------|
| 200 | 成功 | - |
| 400 | 请求参数错误 | 检查请求参数格式 |
| 401 | 未授权 | Token无效或过期，重新登录 |
| 403 | 无权限 | 没有操作权限，联系管理员 |
| 404 | 资源不存在 | 检查请求的资源ID |
| 500 | 服务器内部错误 | 联系系统管理员 |

## 通用响应格式

### 成功响应

```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

### 错误响应

```json
{
  "code": 400,
  "message": "参数错误: xxx",
  "data": null
}
```
