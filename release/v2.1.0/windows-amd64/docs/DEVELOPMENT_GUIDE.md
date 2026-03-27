# 开发规范指南

> 本文档定义排班系统 V2 的开发规范，所有开发者必须遵守。

## 1. 响应格式规范

### 1.1 强制使用 utils 包

**所有 Handler 必须统一使用 `utils.Success()` 和 `utils.Error()`，禁止使用其他方式返回响应。**

```go
// ✅ 正确做法
import "schedule-system-v2/backend/internal/utils"

func (h *Handler) SomeHandler(c *gin.Context) {
    data, err := h.service.GetData()
    if err != nil {
        utils.Error(c, 500, err.Error())
        return
    }
    utils.Success(c, data)
}

// ❌ 错误做法 - 禁止使用
func (h *Handler) BadHandler(c *gin.Context) {
    c.JSON(200, gin.H{"code": 200, "data": data})  // 禁止
    c.JSON(200, model.Success(data))                 // 禁止
    c.JSON(200, model.Response{Code: 200, ...})     // 禁止
}
```

### 1.2 响应结构

统一响应格式：

```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

错误响应：

```json
{
  "code": 400,
  "message": "错误信息",
  "data": null
}
```

## 2. 权限代码规范

### 2.1 权限代码格式

**所有权限代码必须使用冒号格式 `category:action:scope`**

| 格式 | 示例 | 说明 |
|------|------|------|
| ✅ | `schedule:manage:dept` | 排班管理-部门 |
| ✅ | `user:manage:all` | 用户管理-全部 |
| ✅ | `schedule:publish` | 设置每周分工 |
| ❌ | `duty_manage` | 已废弃的下划线格式 |
| ❌ | `schedule_manage` | 已废弃的下划线格式 |

### 2.2 可授权权限列表

系统目前支持的临时权限：

| 权限代码 | 名称 | 适用范围 |
|----------|------|----------|
| `schedule:publish` | 设置每周分工 | 全局 |
| `schedule:manage:all` | 排班管理（全部）| 全部部门 |
| `user:manage:all` | 用户管理（全部）| 全部部门 |
| `schedule:manage:dept` | 排班管理（部门）| 本部门 |
| `user:manage:dept` | 用户管理（部门）| 本部门 |

### 2.3 前端权限映射

前端必须使用与后端完全一致的权限代码：

```javascript
// ✅ 正确做法
const permissionMap = {
  'schedule:manage:dept': '排班管理（部门）',
  'user:manage:dept': '用户管理（部门）',
  'schedule:manage:all': '排班管理（全部）',
  'user:manage:all': '用户管理（全部）',
  'schedule:publish': '设置每周分工'
}

// ❌ 错误做法 - 禁止使用
const badMap = {
  'duty_manage': '值班管理',        // 已废弃
  'schedule_manage': '排班管理',    // 已废弃
}
```

## 3. 前后端接口规范

### 3.1 请求方式

前端统一使用封装的 `request`：

```javascript
// ✅ 正确做法
import request from './request'

const getData = () => request.get('/api/data')
const postData = (data) => request.post('/api/data', data)

// ❌ 错误做法 - 禁止使用
const badFetch = () => fetch('/api/data')  // 绕过拦截器
```

### 3.2 响应处理

拦截器已自动提取 `data` 字段，直接使用返回数据：

```javascript
// ✅ 正确做法
const data = await getSchedule({ week: 1 })
console.log(data)  // 直接是数据，不是 { code, message, data }

// ❌ 错误做法
const res = await getSchedule({ week: 1 })
console.log(res.data)  // res.data 是 undefined
```

### 3.3 接口路径规范

RESTful 风格：

| 操作 | HTTP方法 | 路径示例 |
|------|----------|----------|
| 获取列表 | GET | `/admin/users` |
| 获取单个 | GET | `/admin/users/:id` |
| 创建 | POST | `/admin/users` |
| 更新 | PUT | `/admin/users/:id` |
| 删除 | DELETE | `/admin/users/:id` |
| 设置角色 | PUT | `/admin/users/:id/role` |

## 4. 代码组织规范

### 4.1 后端目录结构

```
backend/internal/
├── auth/          # 权限认证
├── config/        # 配置读取
├── dao/           # 数据访问层
├── db/            # 数据库连接
├── handler/       # HTTP处理器
├── middleware/    # 中间件
├── model/         # 数据模型
├── router/        # 路由配置
├── service/       # 业务逻辑层
└── utils/         # 工具函数
```

### 4.2 命名规范

| 类型 | 命名规则 | 示例 |
|------|----------|------|
| 文件 | snake_case | `temp_permission.go` |
| 结构体 | PascalCase | `GrantPermissionRequest` |
| 接口 | PascalCase | `ApplicationExecutor` |
| 函数 | PascalCase/CamelCase | `GetPermissionList()` |
| 常量 | UPPER_SNAKE_CASE | `PermScheduleManageDept` |
| 变量 | camelCase | `userList` |

## 5. 数据库规范

### 5.1 字段命名

使用 snake_case：

```sql
-- ✅ 正确
created_at, user_id, dept_role

-- ❌ 错误
createdAt, userId, deptRole
```

### 5.2 表结构

```sql
CREATE TABLE IF NOT EXISTS table_name (
    id INT PRIMARY KEY AUTO_INCREMENT,
    ...
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 6. 安全规范

### 6.1 密码处理

```go
// ✅ 正确做法 - 使用 bcrypt 加密
hashedPassword, err := utils.HashPassword(password)
isValid := utils.CheckPassword(password, hashedHash)
```

### 6.2 JWT 处理

```go
// 从上下文获取用户信息（已由中间件验证）
userID := auth.GetUserIDFromContext(c)
checker := auth.GetChecker(c)
```

### 6.3 配置安全

- 生产配置不得提交到代码库
- 使用环境变量或配置文件模板
- 敏感信息使用 `.env` 文件管理

## 7. 错误处理规范

### 7.1 错误返回

```go
// ✅ 正确做法
if err != nil {
    utils.Error(c, 400, "具体错误信息")
    return
}

// ❌ 错误做法
if err != nil {
    c.JSON(400, gin.H{"error": err.Error()})  // 格式不统一
    return
}
```

### 7.2 错误码规范

| 状态码 | 使用场景 |
|--------|----------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未登录/Token无效 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 8. 注释规范

### 8.1 函数注释

```go
// GrantPermission 授予临时权限（支持批量）
// 注意：权限检查由路由层的 PermUserManageDept 中间件处理
func (h *TempPermissionHandler) GrantPermission(c *gin.Context) {
    // ...
}
```

### 8.2 结构体注释

```go
// TempPermissionView 临时权限视图
type TempPermissionView struct {
    ID           int        `json:"id"`
    UserName     string     `json:"user_name"`
    Permission   Permission `json:"permission"`
    // ...
}
```

## 9. 版本控制规范

### 9.1 提交信息

```
feat: 添加临时权限申请功能
fix: 修复权限检查逻辑错误
docs: 更新API文档
refactor: 重构用户服务代码
```

### 9.2 分支管理

- `main` - 生产分支
- `develop` - 开发分支
- `feature/*` - 功能分支
- `hotfix/*` - 紧急修复分支

## 10. 检查清单

提交代码前请确认：

- [ ] 所有响应使用 `utils.Success/Error`
- [ ] 权限代码使用冒号格式
- [ ] 前后端权限代码一致
- [ ] 无 lint 错误
- [ ] 敏感信息未提交
- [ ] 函数有适当注释

---

**注意**: 违反上述规范可能导致代码被拒绝合并。
