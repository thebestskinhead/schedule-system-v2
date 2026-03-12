# 权限系统文档

> 本文档详细介绍排班管理系统的权限体系，包括角色层级、权限组和临时权限机制。

## 概述

系统采用 **RBAC (Role-Based Access Control)** + **临时权限** 的混合权限模型，支持：

- 基于角色的固定权限分配
- 灵活的时间限定临时授权
- 资源范围控制（全局/部门/特定资源）

## 角色层级

### 1. 系统管理员

**标识**: `role=admin`

**权限范围**: 
- 拥有系统的所有权限
- 可以设置用户的系统角色
- 可以配置系统参数（SMTP等）
- 可以进行临时权限授权

**典型用户**: 系统维护人员、超级管理员

### 2. 办公室管理员

**标识**: `department="办公室"` + `dept_role="dept_admin"`

**权限范围**:
- 管理所有部门的排班
- 管理所有用户（不含系统角色设置）
- 发布每周分工
- 进行临时权限授权

**典型用户**: 办公室主任、统筹协调人员

### 3. 部门管理员

**标识**: `dept_role="dept_admin"`

**权限范围**:
- 管理本部门的排班
- 管理本部门成员信息
- 查看本部门无课表

**典型用户**: 各部门部长、副部长

### 4. 部门成员

**标识**: `dept_role="dept_member"`

**权限范围**:
- 查看排班表
- 编辑自己的无课表
- 查看自己的值班
- 确认值班完成

**典型用户**: 普通成员、新加入的干事

## 权限组体系

为了简化权限管理，系统定义了权限组概念。授权时授予权限组，自动包含该组的所有子权限。

### 权限组结构

```
schedule:manage:all (排班管理-全部)
├── schedule:view (查看排班)
├── schedule:preview (预览排班)
├── schedule:confirm (确认排班)
├── schedule:edit (编辑排班)
├── schedule:publish (发布分工)
├── schedule:settings (排班设置)
├── schedule:export (导出排班)
├── schedule:view:all (查看全部)
├── schedule:view:dept (查看部门)
└── schedule:manage:dept (排班管理-部门)
    ├── schedule:view
    ├── schedule:view:dept
    ├── schedule:preview
    ├── schedule:confirm
    └── schedule:edit

user:manage:all (用户管理-全部)
├── user:manage
├── user:manage:dept (用户管理-部门)
│   ├── user:view
│   └── user:edit
├── user:view
└── user:edit
```

### 可授权的权限列表

| 权限代码 | 名称 | 说明 | 适用场景 |
|----------|------|------|----------|
| `schedule:publish` | 设置每周分工 | 设置各部门本周值班日期 | 统筹协调人员 |
| `schedule:manage:all` | 排班管理(全部) | 管理所有部门排班 | 办公室管理员 |
| `user:manage:all` | 用户管理(全部) | 管理所有用户 | 办公室管理员 |
| `schedule:manage:dept` | 排班管理(部门) | 管理本部门排班 | 部门管理员 |
| `user:manage:dept` | 用户管理(部门) | 管理本部门成员 | 部门管理员 |

## 临时权限机制

### 什么是临时权限

临时权限是一种**时间限定**的额外权限授权，用于：

- 临时委派管理职责
- 休假期间的权限代理
- 项目期间的临时授权
- 新管理员培训期间的过渡授权

### 临时权限属性

| 属性 | 说明 | 示例 |
|------|------|------|
| 权限代码 | 授予的权限 | `schedule:manage:dept` |
| 资源类型 | 权限范围 | `all` / `dept` / `user` |
| 资源ID | 具体资源 | 0表示用户当前部门 |
| 过期时间 | 权限失效时间 | 2024-12-31 23:59:59 |
| 授权原因 | 授权说明 | "张三代休期间代理" |

### 资源类型说明

#### 1. 全局权限 (all)

```json
{
  "permission": "schedule:manage:dept",
  "resource_type": "all",
  "resource_id": 0
}
```

**效果**: 可以管理所有部门的排班

#### 2. 部门权限 (dept)

```json
{
  "permission": "schedule:manage:dept",
  "resource_type": "dept",
  "resource_id": 0
}
```

**效果**: 可以管理用户**当前所在部门**的排班

**特点**: 
- resource_id=0 表示动态跟随用户部门
- 用户部门变更时，权限自动适应新部门

#### 3. 特定权限 (user)

```json
{
  "permission": "schedule:manage:dept",
  "resource_type": "user",
  "resource_id": 5
}
```

**效果**: 可以管理指定用户相关的排班（特殊场景）

### 临时权限检查流程

```
用户请求排班操作
      │
      ▼
检查角色权限 ──否──▶ 检查临时权限
      │                 │
     是                  ▼
      │           检查权限组包含
      │                 │
      ▼                 ▼
   允许访问 ◀─────── 检查通过
```

## 权限检查规则

### CanManageDept 检查流程

```go
func CanManageDept(dept string) bool {
    // 1. 系统管理员
    if IsAdmin() return true
    
    // 2. 办公室管理员
    if IsOfficeAdmin() return true
    
    // 3. 部门管理员（本部门）
    if IsDeptAdmin() && department == dept return true
    
    // 4. 临时权限检查
    // 4.1 全局排班权限
    if hasTempPermission(schedule:manage:all) return true
    
    // 4.2 部门排班权限（本部门）
    if HasTempPermissionForDept(schedule:manage:dept, dept) return true
    
    return false
}
```

### 权限优先级

1. **系统管理员** - 最高优先级，无视所有限制
2. **角色权限** - 用户的固定角色权限
3. **临时权限组** - 权限组隐式包含的子权限
4. **临时权限** - 直接授予的临时权限
5. **拒绝访问** - 无任何匹配权限

## 典型授权场景

### 场景1：代理部门管理

**背景**: 部门管理员张三请假一周，需要李四代理

**操作**:
1. 管理员进入"临时权限管理"页面
2. 选择用户：李四
3. 选择权限：`schedule:manage:dept`
4. 资源类型：`部门`
5. 选择部门：`全部部门`（resource_id=0，自动跟随用户部门）
6. 设置过期时间：一周后
7. 填写原因："张三代休期间代理"

**效果**: 李四在接下来的一周内可以管理自己部门的排班

### 场景2：跨部门协调

**背景**: 办公室需要临时协调所有部门排班

**操作**:
1. 系统管理员授予办公室成员 `schedule:manage:all`
2. 资源类型：`全局`
3. 设置合理的过期时间

**效果**: 该成员可以管理所有部门的排班

### 场景3：每周分工设置

**背景**: 需要指定专人负责设置每周各部门值班日期

**操作**:
1. 授予 `schedule:publish` 权限
2. 资源类型：`全局`

**效果**: 该用户可以发布每周分工安排

## 权限继承关系

```
system:admin (系统管理员)
    ├── 所有权限
    └── 可以设置系统角色

office_admin (办公室管理员)
    ├── schedule:manage:all
    ├── user:manage:all
    ├── schedule:publish
    └── 不能设置系统角色

dept_admin (部门管理员)
    ├── schedule:manage:dept
    ├── user:manage:dept
    └── 只能管理本部门

dept_member (部门成员)
    ├── schedule:view
    ├── availability:edit
    └── duty:view
```

## 最佳实践

### 授权原则

1. **最小权限原则** - 只授予必要的权限
2. **限时授权** - 临时权限应设置合理的过期时间
3. **记录原因** - 每次授权都应记录原因，便于审计
4. **定期检查** - 定期清理过期权限

### 安全建议

1. **系统管理员** 账号应严格控制，不轻易授予
2. **临时权限** 过期后应及时续期或撤销
3. **敏感操作**（如删除数据）应保留操作日志
4. **密码安全** - 定期更换密码，使用强密码

## 常见问题

### Q1: 为什么有临时权限还是无法访问？

**可能原因**:
- 临时权限已过期
- 资源类型不匹配（如授权了A部门，但操作B部门）
- 权限组检查失败（确认权限代码正确）

**排查方法**:
1. 查看"我的权限"页面，确认权限有效
2. 检查资源类型和范围是否正确
3. 联系管理员确认授权信息

### Q2: 如何批量授权？

目前系统不支持批量授权，需要逐个用户授权。建议：
- 通过角色分配实现批量权限
- 或使用全局权限（resource_type=all）

### Q3: 临时权限可以转授吗？

不可以。临时权限不能转授，只有系统管理员和办公室管理员可以进行权限授权。

### Q4: 权限过期前会有提醒吗？

目前系统会在"我的权限"页面显示剩余天数，但不会主动推送提醒。建议：
- 定期查看权限状态
- 提前申请续期
