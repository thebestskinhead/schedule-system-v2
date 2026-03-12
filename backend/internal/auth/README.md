# 权限管理系统

## 概述

本模块提供了统一的权限检查抽象，支持中间件和Handler层面的权限控制。

**核心原则**：Handler层不直接读取JWT上下文，而是通过鉴权模块提供的API获取用户信息和权限判断。

## 架构设计

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Handler层     │────▶│   鉴权模块       │────▶│   JWT上下文     │
│                 │     │ (auth包)        │     │                 │
└─────────────────┘     └─────────────────┘     └─────────────────┘
```

- **JWT上下文**：由 `AuthMiddleware` 解析token后写入gin上下文
- **鉴权模块**：唯一访问JWT上下文的层，提供封装好的权限检查API
- **Handler层**：通过鉴权模块获取用户信息，禁止直接 `c.Get("userID")`

## 用户信息获取（Handler层使用）

```go
// 获取当前用户完整信息
user := auth.GetCurrentUser(c)
// user.UserID, user.StudentID, user.Role, user.Department, user.DeptRole

// 获取单个字段
userID := auth.GetUserIDFromContext(c)
studentID := auth.GetStudentIDFromContext(c)
role := auth.GetRoleFromContext(c)
department := auth.GetDepartmentFromContext(c)
deptRole := auth.GetDeptRoleFromContext(c)

// 注意：所有获取函数在无法获取时返回零值，Handler应检查userID是否为0
if userID == 0 {
    auth.ResponseUnauthorized(c)
    return
}
```

## 权限检查（Handler层使用）

### 1. 检查指定权限

```go
func (h *Handler) SomeHandler(c *gin.Context) {
    // 检查权限，无权限时自动返回403
    if !auth.CheckPermission(c, auth.PermScheduleEdit) {
        return // 已自动返回403
    }
    // 执行业务逻辑...
}
```

### 2. 检查部门权限

```go
func (h *Handler) UpdateDeptSchedule(c *gin.Context) {
    dept := c.Param("department")
    
    // 检查是否有管理该部门的权限
    if !auth.CheckDeptPermission(c, auth.PermScheduleManageDept, dept) {
        return // 已自动返回403
    }
    // 执行业务逻辑...
}
```

### 3. 使用Checker对象（更灵活）

```go
func (h *Handler) SomeHandler(c *gin.Context) {
    checker := auth.GetChecker(c)
    
    // 检查单一权限
    if !checker.HasPermission(auth.PermScheduleEdit) {
        auth.ResponseForbidden(c)
        return
    }
    
    // 检查任意权限
    if !checker.HasAnyPermission(auth.PermScheduleEdit, auth.PermScheduleView) {
        auth.ResponseForbidden(c)
        return
    }
    
    // 检查部门权限
    if !checker.HasDeptPermission(auth.PermScheduleManageDept, "办公室") {
        auth.ResponseForbidden(c)
        return
    }
}
```

### 4. 检查资源所有权

```go
func (h *Handler) GetUserDetail(c *gin.Context) {
    targetStudentID := c.Param("student_id")
    
    // 检查是否是本人或管理员（系统管理员或办公室管理员）
    if !auth.CheckOwnerOrAdmin(c, targetStudentID) {
        return
    }
    // 执行业务逻辑...
}
```

## 角色层级

| 角色 | 标识 | 权限范围 |
|------|------|---------|
| 系统管理员 | `role=admin` | 全部权限 |
| 办公室管理员 | `department=办公室, dept_role=admin` | 发布每周分工、查看所有部门 |
| 部门管理员 | `dept_role=admin` | 管理部门内用户和排班 |
| 部门成员 | `dept_role=member` | 查看本部门信息、管理个人数据 |

## 权限列表

| 权限 | 说明 | 系统管理员 | 办公室管理员 | 部门管理员 | 部门成员 |
|------|------|-----------|-------------|-----------|---------|
| `user:profile` | 查看/修改个人信息 | ✓ | ✓ | ✓ | ✓ |
| `user:manage` | 用户管理（全部） | ✓ | ✓ | ✗ | ✗ |
| `user:set_role` | 设置用户角色 | ✓ | ✗ | ✗ | ✗ |
| `availability:view` | 查看自己的无课表 | ✓ | ✓ | ✓ | ✓ |
| `availability:edit` | 编辑自己的无课表 | ✓ | ✓ | ✓ | ✓ |
| `availability:import` | 导入无课表 | ✓ | ✓ | ✓ | ✓ |
| `availability:view_all` | 查看所有人的无课表 | ✓ | ✓ | ✗ | ✗ |
| `schedule:view` | 查看排班 | ✓ | ✓ | ✓ | ✓ |
| `schedule:preview` | 预览排班 | ✓ | ✓ | ✓ | ✗ |
| `schedule:confirm` | 确认排班 | ✓ | ✓ | ✓ | ✗ |
| `schedule:edit` | 编辑排班 | ✓ | ✗ | ✓(本部门) | ✗ |
| `schedule:settings` | 排班设置 | ✓ | ✓ | ✗ | ✗ |
| `schedule:export` | 导出排班 | ✓ | ✓ | ✓ | ✗ |
| `schedule:publish` | 发布每周分工 | ✓ | ✓ | ✗ | ✗ |
| `template:view` | 查看模板 | ✓ | ✓ | ✓ | ✓ |
| `template:edit` | 编辑模板 | ✓ | ✓ | ✗ | ✗ |
| `system:admin` | 系统管理 | ✓ | ✗ | ✗ | ✗ |

## 在Router中使用中间件

```go
// 使用特定权限中间件
authGroup.GET("/users", middleware.PermissionMiddleware(auth.PermUserManage), userHandler.GetUserList)

// 使用管理员中间件（仅系统管理员）
admin.POST("/schedule/confirm", middleware.AdminMiddleware(), scheduleHandler.ConfirmSchedule)

// 使用办公室管理员检查
admin.POST("/schedule/publish", middleware.PermissionMiddleware(auth.PermSchedulePublish), scheduleHandler.PublishSchedule)
```

## Checker方法参考

```go
checker := auth.GetChecker(c)

// 角色检查
checker.IsAdmin()        // 是否是系统管理员
checker.IsOfficeAdmin()  // 是否是办公室管理员
checker.IsDeptAdmin()    // 是否是部门管理员

// 部门访问检查
checker.CanManageDept("办公室")  // 是否可以管理指定部门
checker.CanViewDept("办公室")    // 是否可以查看指定部门

// 权限检查
checker.HasPermission(auth.PermScheduleEdit)
checker.HasAnyPermission(auth.PermScheduleEdit, auth.PermScheduleView)
checker.HasAllPermissions(auth.PermScheduleEdit, auth.PermScheduleView)
checker.HasDeptPermission(auth.PermScheduleManageDept, "办公室")

// 资源所有权
checker.IsOwner(targetStudentID)

// 获取用户信息
checker.GetStudentID()
checker.GetRole()
checker.GetDepartment()
checker.GetDeptRole()
```

## 响应辅助函数

```go
// 返回403权限不足
auth.ResponseForbidden(c)
auth.ResponseForbidden(c, "自定义错误消息")

// 返回401未认证
auth.ResponseUnauthorized(c)
auth.ResponseUnauthorized(c, "自定义错误消息")
```

## 扩展方法

### 添加新权限

1. 在 `permission.go` 中定义新权限常量
2. 在 `RolePermissions` 中为对应角色分配权限
3. 在 `PathPermissionMap` 中映射API路径（可选）

### 添加新角色

```go
RolePermissions["moderator"] = []Permission{
    PermUserProfile,
    PermAvailabilityView,
    PermAvailabilityEdit,
    // ...
}
```

## 注意事项

1. **Handler禁止直接访问JWT上下文**：所有 `c.Get("xxx")` 都应通过 `auth` 包提供的函数
2. **始终检查userID是否为0**：表示未获取到用户信息
3. **部门权限检查**：使用 `CheckDeptPermission` 或 `HasDeptPermission` 进行部门级别的权限控制
4. **JWT包含部门信息**：登录后JWT token包含 `department` 和 `dept_role`，修改后需重新登录生效
