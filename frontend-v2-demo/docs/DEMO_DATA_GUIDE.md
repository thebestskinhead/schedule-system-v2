# 示例排班数据使用指南

## 自动生成数据

系统初始化时会自动生成以下演示数据：

| 数据类型 | 数量 | 说明 |
|---------|------|------|
| 用户 | 19人 | 4个预设角色 + 15个演示用户 |
| 无课表 | ~7600条 | 每人20周 × 每周20个时段 |
| 排班记录 | ~2560条 | 第1-16周，4个部门 |
| 值班记录 | ~2560条 | 带完成状态 |

## 浏览器控制台工具

打开浏览器控制台(F12)，可以使用以下命令：

### 1. 生成排班数据

```javascript
// 生成默认数据（第1-16周）
await seedDemoData()

// 生成指定周次
await seedDemoData({ startWeek: 1, endWeek: 20 })

// 生成指定部门
await seedDemoData({ departments: ['竞赛部', '项目部'] })
```

### 2. 生成个人值班数据

```javascript
// 为当前用户生成值班数据
const user = JSON.parse(localStorage.getItem('mock_auth_user'))
await seedMyDuties(user.id, user.name)

// 为指定用户生成（指定周次）
await seedMyDuties(4, '张三', [8, 9, 10, 11, 12])
```

### 3. 查看统计信息

```javascript
await showScheduleStats()
```

输出示例：
```
📊 排班数据统计
==================
📈 总排班记录: 2560

📅 按周次分布:
  第01周: 160 条
  第02周: 160 条
  ...
  第16周: 160 条

🏢 按部门分布:
  竞赛部: 640 条
  办公室: 640 条
  项目部: 640 条
  科普部: 640 条
```

### 4. 清空排班数据

```javascript
await clearSchedules()
```

## API 接口

### 重新生成排班数据

```http
POST /api/v1/demo/seed-schedules
Authorization: Bearer <token>

Response:
{
  "code": 200,
  "message": "排班数据重新生成成功",
  "data": {
    "schedules": 2560,
    "duties": 2560,
    "weekStats": { "1": 160, "2": 160, ... }
  }
}
```

**权限**: 仅系统管理员可用

### 清空排班数据

```http
POST /api/v1/demo/clear-schedules
Authorization: Bearer <token>
```

**权限**: 仅系统管理员可用

### 查看统计数据

```http
GET /api/v1/demo/stats
Authorization: Bearer <token>

Response:
{
  "code": 200,
  "data": {
    "users": 19,
    "availability": 7600,
    "schedules": 2560,
    "duties": 2560,
    "weekStats": { ... }
  }
}
```

## 排班数据结构

### Schedule（排班记录）

```javascript
{
  id: "sch_1_1_1_4",        // 唯一标识
  week: 1,                   // 周次
  weekday: 1,                // 星期（1-5）
  period: 1,                 // 节次（1-4）
  user_id: 4,                // 用户ID
  user_name: "张三",          // 用户姓名
  student_id: "2021001",     // 学号
  department: "竞赛部",       // 部门
  status: "confirmed",       // 状态：confirmed
  created_at: "2024-03-15T10:30:00Z"
}
```

### Duty（值班记录）

```javascript
{
  week: 1,
  weekday: 1,
  period: 1,
  user_id: 4,
  user_name: "张三",
  department: "竞赛部",
  status: "confirmed",           // 排班状态
  duty_status: "completed",      // 值班状态：pending/confirmed/completed
  check_in_time: "2024-03-15T14:30:00Z",  // 签到时间
  created_at: "2024-03-15T10:30:00Z"
}
```

## 数据状态说明

### 值班状态 (duty_status)

| 状态 | 说明 | 适用场景 |
|------|------|----------|
| `pending` | 待确认 | 未来周次或刚生成的排班 |
| `confirmed` | 已确认 | 已确认但未执行的值班 |
| `completed` | 已完成 | 已执行并签到的值班 |

当前模拟数据的时间设定：
- 第 1-8 周：已完成 (completed)
- 第 9-10 周：部分完成
- 第 11+ 周：已确认 (confirmed)

## 数据存储位置

所有数据存储在浏览器 localStorage 中：

```
mock_db_users          # 用户数据
mock_db_availability   # 无课表数据
mock_db_schedules_1    # 第1周排班
mock_db_schedules_2    # 第2周排班
...
mock_auth_token        # 登录token
mock_auth_user         # 当前用户
```

## 重置所有数据

如需完全重置数据，可以在浏览器控制台执行：

```javascript
// 清除所有 mock 数据
Object.keys(localStorage)
  .filter(k => k.startsWith('mock_'))
  .forEach(k => localStorage.removeItem(k))

// 刷新页面重新初始化
location.reload()
```

或者使用 MockAPI 提供的重置功能：

```javascript
await mockAPI.reset()
```
