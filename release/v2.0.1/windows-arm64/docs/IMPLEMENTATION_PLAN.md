# 排班系统v2.1 权限管理升级实施计划

## 概述

本文档详细描述了排班系统从v2.0到v2.1的升级计划，重点是完善基于部门的权限管理系统，新增每周值班分工管理、临时授权和精细化用户管理功能。

---

## 当前状态分析

### 已完成功能 ✅

| 功能 | 状态 | 说明 |
|------|------|------|
| JWT认证 | ✅ | 基础token认证已实现 |
| 基础权限检查 | ✅ | 简单的admin/user角色区分 |
| 部门和部门角色字段 | ✅ | user表已添加department和dept_role字段 |
| 鉴权模块重构 | ✅ | Handler不再直接读取JWT上下文，统一通过auth模块访问 |
| JWT包含部门信息 | ✅ | token现在包含department和dept_role |

### 部分实现 ⚠️

| 功能 | 状态 | 说明 |
|------|------|------|
| 部门数据隔离 | ⚠️ | DAO层查询未添加department过滤条件 |
| 部门权限检查 | ⚠️ | auth模块已提供方法，但业务逻辑未使用 |

### 未实现 ❌

| 功能 | 状态 | 说明 |
|------|------|------|
| 每周值班分工管理 | ❌ | 需要新表和完整CRUD |
| 临时授权系统 | ❌ | user_permissions_temp表未创建 |
| 精细化用户管理 | ❌ | 需要按部门筛选和部门角色设置 |
| 办公室管理员角色 | ❌ | 需要特殊识别"办公室"部门的管理员 |

---

## 权限层级设计

### 角色定义

```
┌─────────────────────────────────────────────────────────────┐
│                      系统管理员                               │
│                   (role = admin)                              │
│                      全部权限                                 │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    办公室管理员                               │
│     (role = user, department = 办公室, dept_role = admin)    │
│          发布每周分工、查看所有部门排班、管理所有用户          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     部门管理员                                │
│           (dept_role = admin, 部门不限)                       │
│              管理部门内用户、管理本部门排班                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      部门成员                                 │
│           (dept_role = member, 部门不限)                      │
│              查看本部门排班、编辑个人无课表、更新值班状态       │
└─────────────────────────────────────────────────────────────┘
```

### 权限矩阵

| 权限 | 系统管理员 | 办公室管理员 | 部门管理员 | 部门成员 |
|------|-----------|-------------|-----------|---------|
| **用户管理** |||||
| 查看所有用户 | ✅ | ✅ | ❌ | ❌ |
| 查看本部门用户 | ✅ | ✅ | ✅ | ❌ |
| 编辑所有用户信息 | ✅ | ✅ | ❌ | ❌ |
| 编辑本部门用户信息 | ✅ | ✅ | ✅ | ❌ |
| 设置用户角色 | ✅ | ❌ | ❌ | ❌ |
| 设置用户部门角色 | ✅ | ✅ | ✅(本部门) | ❌ |
| **每周分工** |||||
| 发布每周分工 | ✅ | ✅ | ❌ | ❌ |
| 查看全部分工 | ✅ | ✅ | ✅(有排班权限的部门) | ❌ |
| 查看本部门分工 | ✅ | ✅ | ✅ | ✅ |
| **部门排班** |||||
| 预览排班 | ✅ | ✅ | ✅(本部门) | ❌ |
| 确认排班 | ✅ | ✅ | ✅(本部门) | ❌ |
| 编辑排班 | ✅ | ❌ | ✅(本部门) | ❌ |
| 查看排班 | ✅ | ✅ | ✅ | ✅ |
| **无课表** |||||
| 编辑自己的无课表 | ✅ | ✅ | ✅ | ✅ |
| 查看自己的无课表 | ✅ | ✅ | ✅ | ✅ |
| 查看所有人的无课表 | ✅ | ✅ | ❌ | ❌ |
| 导入无课表 | ✅ | ✅ | ✅ | ✅ |
| **值班** |||||
| 查看值班安排 | ✅ | ✅ | ✅ | ✅ |
| 更新值班状态 | ✅ | ✅ | ✅ | ✅ |
| **模板** |||||
| 查看模板 | ✅ | ✅ | ✅ | ✅ |
| 编辑模板 | ✅ | ✅ | ❌ | ❌ |

---

## 数据库变更计划

### 1. 每周值班分工表 (weekly_duty_assignments)

```sql
CREATE TABLE weekly_duty_assignments (
    id INT PRIMARY KEY AUTO_INCREMENT,
    week INT NOT NULL COMMENT '周次',
    department VARCHAR(50) NOT NULL COMMENT '部门',
    weekday INT NOT NULL COMMENT '星期几(1-5)',
    is_assigned BOOLEAN DEFAULT TRUE COMMENT '是否安排值班',
    created_by INT NOT NULL COMMENT '创建人ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_week_dept_weekday (week, department, weekday),
    INDEX idx_week (week),
    INDEX idx_department (department)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='每周值班分工表';
```

### 2. 用户临时权限表 (user_permissions_temp)

```sql
CREATE TABLE user_permissions_temp (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL COMMENT '被授权用户ID',
    permission VARCHAR(50) NOT NULL COMMENT '权限代码',
    resource_type VARCHAR(20) DEFAULT 'all' COMMENT '资源类型(all/dept/user)',
    resource_id INT DEFAULT 0 COMMENT '资源ID(部门ID或用户ID)',
    granted_by INT NOT NULL COMMENT '授权人ID',
    reason VARCHAR(255) COMMENT '授权原因',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL COMMENT '过期时间',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否有效',
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at),
    INDEX idx_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户临时权限表';
```

### 3. 更新现有表

```sql
-- 为用户表添加索引（如果还没有）
ALTER TABLE users ADD INDEX idx_department (department);
ALTER TABLE users ADD INDEX idx_dept_role (dept_role);
ALTER TABLE users ADD INDEX idx_role (role);

-- 为排班记录表添加部门字段（用于数据隔离查询优化）
ALTER TABLE duty_records ADD COLUMN department VARCHAR(50) AFTER user_id;
ALTER TABLE duty_records ADD INDEX idx_department (department);

-- 更新现有记录的部门字段
UPDATE duty_records dr
JOIN users u ON dr.user_id = u.id
SET dr.department = u.department;
```

---

## 后端开发计划

### Phase 1: 数据访问层 (DAO)

#### 1.1 每周分工DAO (weekly_duty_dao.go)

```go
type WeeklyDutyDAO struct {
    db *sqlx.DB
}

// 核心方法
func (d *WeeklyDutyDAO) Create(assignment *model.WeeklyDutyAssignment) error
func (d *WeeklyDutyDAO) GetByWeek(week int) ([]model.WeeklyDutyAssignment, error)
func (d *WeeklyDutyDAO) GetByWeekAndDept(week int, dept string) (*model.WeeklyDutyAssignment, error)
func (d *WeeklyDutyDAO) Update(assignment *model.WeeklyDutyAssignment) error
func (d *WeeklyDutyDAO) Delete(id int) error
func (d *WeeklyDutyDAO) GetByWeekAndDepartments(week int, depts []string) ([]model.WeeklyDutyAssignment, error)
```

#### 1.2 临时权限DAO (temp_permission_dao.go)

```go
type TempPermissionDAO struct {
    db *sqlx.DB
}

// 核心方法
func (d *TempPermissionDAO) Create(perm *model.UserPermissionTemp) error
func (d *TempPermissionDAO) GetActiveByUserID(userID int) ([]model.UserPermissionTemp, error)
func (d *TempPermissionDAO) Revoke(id int) error
func (d *TempPermissionDAO) CleanupExpired() error
func (d *TempPermissionDAO) HasPermission(userID int, perm model.Permission) (bool, error)
```

#### 1.3 用户DAO扩展

```go
// 添加部门筛选方法
func (d *UserDAO) ListByDepartment(dept string) ([]model.User, error)
func (d *UserDAO) ListByDepartments(depts []string) ([]model.User, error)
func (d *UserDAO) SetDeptRole(userID int, deptRole string) error
func (d *UserDAO) SetDepartment(userID int, dept string) error
```

### Phase 2: 业务逻辑层 (Service)

#### 2.1 每周分工服务 (weekly_duty_service.go)

```go
type WeeklyDutyService struct {
    dao *dao.WeeklyDutyDAO
}

// 核心方法
func (s *WeeklyDutyService) PublishAssignment(adminID int, req *model.PublishAssignmentRequest) error
func (s *WeeklyDutyService) GetWeekAssignments(week int, userID int, userDept string, isAdmin bool) ([]model.WeeklyDutyAssignment, error)
func (s *WeeklyDutyService) GetDeptAssignment(week int, dept string) (*model.WeeklyDutyAssignment, error)
func (s *WeeklyDutyService) UpdateAssignment(adminID int, req *model.UpdateAssignmentRequest) error
func (s *WeeklyDutyService) DeleteAssignment(adminID int, id int) error
```

#### 2.2 临时权限服务 (temp_permission_service.go)

```go
type TempPermissionService struct {
    dao *dao.TempPermissionDAO
}

// 核心方法
func (s *TempPermissionService) GrantPermission(adminID int, req *model.GrantPermissionRequest) error
func (s *TempPermissionService) RevokePermission(adminID int, permID int) error
func (s *TempPermissionService) GetUserTempPermissions(userID int) ([]model.UserPermissionTemp, error)
func (s *TempPermissionService) CheckTempPermission(userID int, perm model.Permission) (bool, error)
func (s *TempPermissionService) CleanupExpiredPermissions() error
```

#### 2.3 用户服务扩展

```go
// 添加部门管理方法
func (s *UserService) GetUserListByDepartment(dept string) ([]model.User, error)
func (s *UserService) SetUserDepartment(adminID, targetUserID int, dept string) error
func (s *UserService) SetUserDeptRole(adminID, targetUserID int, deptRole string) error
func (s *UserService) GetUsersByDepts(depts []string) ([]model.User, error)
```

### Phase 3: 权限检查层扩展

#### 3.1 更新Checker支持临时权限

```go
// Checker结构体添加
type Checker struct {
    studentID  string
    role       string
    department string
    deptRole   string
    userID     int
    tempPerms  []model.Permission // 临时权限列表（懒加载）
}

// 新增方法
func (c *Checker) loadTempPermissions() // 从数据库加载临时权限
func (c *Checker) HasPermissionWithTemp(perm model.Permission) bool // 包含临时权限的检查
func (c *Checker) GetEffectivePermissions() []model.Permission // 获取所有有效权限（含临时）
```

#### 3.2 添加权限守卫函数

```go
// 权限守卫模式
func GuardPermission(c *gin.Context, perm model.Permission) bool
func GuardDeptPermission(c *gin.Context, perm model.Permission, dept string) bool
func GuardAnyPermission(c *gin.Context, perms ...model.Permission) bool
```

### Phase 4: API Handler层

#### 4.1 每周分工Handler (weekly_duty_handler.go)

```go
type WeeklyDutyHandler struct {
    service *service.WeeklyDutyService
}

// 路由映射
POST   /api/v1/admin/duty-assignments          // 发布分工（办公室管理员/系统管理员）
GET    /api/v1/admin/duty-assignments          // 获取分工列表
PUT    /api/v1/admin/duty-assignments/:id      // 更新分工
DELETE /api/v1/admin/duty-assignments/:id      // 删除分工
GET    /api/v1/duty-assignments/my-dept        // 获取本部门分工（所有登录用户）
```

#### 4.2 临时权限Handler (temp_permission_handler.go)

```go
type TempPermissionHandler struct {
    service *service.TempPermissionService
}

// 路由映射
POST   /api/v1/admin/temp-permissions          // 授权（管理员）
GET    /api/v1/admin/temp-permissions          // 获取所有临时权限（管理员）
DELETE /api/v1/admin/temp-permissions/:id      // 撤销权限（管理员）
GET    /api/v1/temp-permissions/my             // 获取自己的临时权限
```

#### 4.3 用户管理Handler扩展

```go
// 新增路由
GET    /api/v1/admin/users/by-dept             // 按部门获取用户
PUT    /api/v1/admin/users/:id/department      // 修改用户部门
PUT    /api/v1/admin/users/:id/dept-role       // 修改用户部门角色
```

---

## 前端开发计划

### 页面清单

| 页面 | 路径 | 权限要求 | 功能描述 |
|------|------|---------|---------|
| 每周分工管理 | /admin/duty-assignments | 办公室管理员/系统管理员 | 发布/编辑每周各部门值班安排 |
| 授权管理 | /admin/temp-permissions | 系统管理员 | 授予临时权限、查看授权记录 |
| 用户管理v2 | /admin/users/v2 | 系统管理员/办公室管理员/部门管理员 | 按部门筛选、设置部门角色 |
| 我的授权 | /my-permissions | 登录用户 | 查看自己获得的临时权限 |

### 组件设计

#### 1. 每周分工管理组件

```vue
<!-- WeeklyDutyAssignment.vue -->
<template>
  <div>
    <WeekSelector v-model="currentWeek" />
    <DepartmentGrid 
      :week="currentWeek"
      :departments="departments"
      v-model="assignments"
      @save="saveAssignments"
    />
    <AssignmentPreview :assignments="assignments" />
  </div>
</template>
```

**功能**：
- 周次选择器
- 部门网格（显示各部门周一至周五是否值班）
- 快速编辑模式（批量设置）
- 预览和确认

#### 2. 授权管理组件

```vue
<!-- TempPermissionGrant.vue -->
<template>
  <div>
    <UserSelector @select="selectedUser = $event" />
    <PermissionSelector v-model="selectedPermissions" />
    <ResourceSelector 
      v-if="needsResource"
      type="dept|user"
      v-model="resource"
    />
    <ExpirationPicker v-model="expiresAt" />
    <ReasonInput v-model="reason" />
    <Button @click="grant">授权</Button>
    
    <ActivePermissionsTable 
      :permissions="activePermissions"
      @revoke="revokePermission"
    />
  </div>
</template>
```

**功能**：
- 用户搜索和选择
- 权限多选
- 资源范围选择（全部/指定部门/指定用户）
- 过期时间设置
- 授权原因记录
- 活跃权限列表和撤销

#### 3. 精细化用户管理组件

```vue
<!-- UserManagementV2.vue -->
<template>
  <div>
    <DepartmentFilter 
      v-model="selectedDepts"
      :options="availableDepartments"
    />
    <RoleFilter v-model="roleFilter" />
    <UserTable 
      :users="filteredUsers"
      @edit="openEditModal"
      @set-dept-role="setDeptRole"
    />
    <EditUserModal 
      v-model:visible="editModalVisible"
      :user="editingUser"
      @save="saveUser"
    />
  </div>
</template>
```

**功能**：
- 部门多选筛选
- 角色筛选（系统角色+部门角色）
- 批量设置部门角色
- 部门转移功能

### API 封装

```javascript
// api/dutyAssignment.js
export const dutyAssignmentAPI = {
  publish: (data) => api.post('/admin/duty-assignments', data),
  list: (week) => api.get('/admin/duty-assignments', { params: { week } }),
  update: (id, data) => api.put(`/admin/duty-assignments/${id}`, data),
  delete: (id) => api.delete(`/admin/duty-assignments/${id}`),
  getMyDept: () => api.get('/duty-assignments/my-dept'),
}

// api/tempPermission.js
export const tempPermissionAPI = {
  grant: (data) => api.post('/admin/temp-permissions', data),
  list: () => api.get('/admin/temp-permissions'),
  revoke: (id) => api.delete(`/admin/temp-permissions/${id}`),
  getMy: () => api.get('/temp-permissions/my'),
}

// api/user.js 扩展
export const userAPI = {
  // ... 现有方法
  listByDept: (dept) => api.get('/admin/users/by-dept', { params: { dept } }),
  setDepartment: (id, dept) => api.put(`/admin/users/${id}/department`, { department: dept }),
  setDeptRole: (id, role) => api.put(`/admin/users/${id}/dept-role`, { dept_role: role }),
}
```

---

## 测试计划

### 单元测试

```go
// 权限检查测试
func TestChecker_HasDeptPermission(t *testing.T)
func TestChecker_IsOfficeAdmin(t *testing.T)
func TestChecker_CanManageDept(t *testing.T)

// 临时权限测试
func TestTempPermissionService_Grant(t *testing.T)
func TestTempPermissionService_Revoke(t *testing.T)
func TestTempPermissionService_CheckExpired(t *testing.T)

// 每周分工测试
func TestWeeklyDutyService_Publish(t *testing.T)
func TestWeeklyDutyService_GetByWeek(t *testing.T)
```

### 集成测试

```go
// API权限测试
func TestAPI_DutyAssignment_CRUD(t *testing.T)
func TestAPI_TempPermission_Lifecycle(t *testing.T)
func TestAPI_UserManagement_DeptFilter(t *testing.T)

// 数据隔离测试
func TestDataIsolation_DeptAdmin_CanOnlySeeOwnDept(t *testing.T)
func TestDataIsolation_OfficeAdmin_CanSeeAllDepts(t *testing.T)
```

### 前端测试

```javascript
// 组件测试
describe('WeeklyDutyAssignment', () => {
  it('should render department grid correctly')
  it('should save assignments when clicking save')
  it('should validate before publishing')
})

describe('TempPermissionGrant', () => {
  it('should show resource selector when permission needs it')
  it('should validate expiration time is in future')
})
```

---

## 部署计划

### 数据库迁移顺序

```bash
# 1. 创建新表
migrate -path ./migrations -database "mysql://..." up

# 2. 更新现有数据
mysql -e "UPDATE duty_records ..."

# 3. 验证数据完整性
```

### 后端部署

```bash
# 1. 编译
cd backend
go build -o server

# 2. 运行迁移
./server migrate

# 3. 启动服务
./server
```

### 前端部署

```bash
cd frontend
npm install
npm run build
# 将dist目录复制到backend/dist
```

---

## 时间安排

| 阶段 | 任务 | 预估时间 | 依赖 |
|------|------|---------|------|
| **Week 1** ||||
| Day 1-2 | 数据库表设计、迁移脚本 | 2天 | - |
| Day 3-4 | DAO层开发 | 2天 | 数据库 |
| Day 5 | 基础Service层 | 1天 | DAO |
| **Week 2** ||||
| Day 1-2 | 权限检查层扩展 | 2天 | Service |
| Day 3-4 | API Handler开发 | 2天 | 权限层 |
| Day 5 | 后端集成测试 | 1天 | Handler |
| **Week 3** ||||
| Day 1-3 | 前端页面开发 | 3天 | API |
| Day 4 | 前端API对接 | 1天 | 前端页面 |
| Day 5 | 前后端联调 | 1天 | 完整功能 |
| **Week 4** ||||
| Day 1-2 | 集成测试、Bug修复 | 2天 | 完整功能 |
| Day 3-4 | 文档完善 | 2天 | 稳定版本 |
| Day 5 | 部署上线 | 1天 | 测试通过 |

---

## 风险与应对

| 风险 | 可能性 | 影响 | 应对措施 |
|------|-------|------|---------|
| JWT变更导致现有用户失效 | 高 | 中 | 前端检测401，引导用户重新登录 |
| 数据隔离遗漏导致信息泄露 | 中 | 高 | 代码审查+自动化测试验证 |
| 临时权限过期逻辑错误 | 中 | 中 | 单元测试覆盖边界条件 |
| 性能问题（权限检查频繁查库） | 中 | 中 | 添加缓存层（Redis） |

---

## 附录

### A. 数据库迁移文件清单

```
migrations/
├── 001_init.sql                    # 已存在
├── 002_add_templates.sql           # 已存在
├── 003_add_current_week.sql        # 已存在
├── 004_add_task_queue.sql          # 已存在
├── 005_add_department.sql          # 已存在
├── 006_add_weekly_duty.sql         # 新增
├── 007_add_temp_permissions.sql    # 新增
└── 008_update_duty_records.sql     # 新增
```

### B. API端点清单

| 方法 | 路径 | 权限 | 描述 |
|------|------|------|------|
| POST | /api/v1/admin/duty-assignments | schedule:publish | 发布分工 |
| GET | /api/v1/admin/duty-assignments | schedule:view:all | 获取全部分工 |
| PUT | /api/v1/admin/duty-assignments/:id | schedule:publish | 更新分工 |
| DELETE | /api/v1/admin/duty-assignments/:id | schedule:publish | 删除分工 |
| GET | /api/v1/duty-assignments/my-dept | schedule:view:dept | 获取本部门分工 |
| POST | /api/v1/admin/temp-permissions | system:admin | 授予临时权限 |
| GET | /api/v1/admin/temp-permissions | system:admin | 列出临时权限 |
| DELETE | /api/v1/admin/temp-permissions/:id | system:admin | 撤销权限 |
| GET | /api/v1/temp-permissions/my | 登录用户 | 我的临时权限 |
| GET | /api/v1/admin/users/by-dept | user:manage:all/dept | 按部门筛选用户 |
| PUT | /api/v1/admin/users/:id/department | user:manage:all | 修改用户部门 |
| PUT | /api/v1/admin/users/:id/dept-role | user:manage:all/dept | 修改部门角色 |

---

**文档版本**: v1.0  
**创建日期**: 2026-03-02  
**最后更新**: 2026-03-02  
**作者**: AI Assistant
