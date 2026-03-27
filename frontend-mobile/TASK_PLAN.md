# 排班系统移动端改造 - 任务计划

## 禁区文件（禁止修改）
- `src/api/request.js` - 请求拦截器，仅允许 ElMessage→Toast
- `src/api/*.js` - 所有 API 端点定义
- `src/stores/user.js` - 权限存储和计算
- `src/router/index.js` - 路由守卫逻辑（仅可改 component 引用路径）

## 修改清单

### T1: 基础配置文件 (并行)
| # | 文件 | 操作 | 说明 |
|---|------|------|------|
| T1.1 | `package.json` | 修改 | 移除 element-plus，添加 vant@4.8.0 |
| T1.2 | `vite.config.js` | 修改 | 添加 VantResolver，端口改5174 |
| T1.3 | `index.html` | 修改 | viewport 添加移动端参数 |
| T1.4 | `src/main.js` | 修改 | 移除 Element Plus，引入 Vant |
| T1.5 | `src/styles/variables.css` | 新建 | CSS 设计变量 |
| T1.6 | `src/utils/native.js` | 新建 | 原生桥接预留接口 |

### T2: 布局组件 (并行，依赖 T1)
| # | 文件 | 操作 | 说明 |
|---|------|------|------|
| T2.1 | `src/components/MobileLayout.vue` | 新建 | 底部 TabBar 布局 |
| T2.2 | `src/components/Navbar.vue` | 新建 | 顶部导航栏 |
| T2.3 | `src/components/PeriodCard.vue` | 新建 | 节次状态卡片 |
| T2.4 | `src/components/SafeArea.vue` | 新建 | 安全区域适配 |

### T3: 页面文件 (并行，依赖 T1)
| # | 文件 | 操作 | 说明 |
|---|------|------|------|
| T3.1 | `src/views/mobile/Home.vue` | 新建 | 首页占位 |
| T3.2 | `src/views/mobile/Availability.vue` | 新建 | 无课表占位 |
| T3.3 | `src/views/mobile/Schedule.vue` | 新建 | 排班占位 |
| T3.4 | `src/views/mobile/Profile.vue` | 新建 | 我的占位 |
| T3.5 | `src/views/mobile/MyDuty.vue` | 新建 | 我的值班占位 |
| T3.6 | `src/views/mobile/ScheduleResult.vue` | 新建 | 排班结果占位 |

### T4: 路由 + 根组件 (依赖 T1)
| # | 文件 | 操作 | 说明 |
|---|------|------|------|
| T4.1 | `src/router/index.js` | 修改 | Layout→MobileLayout，改 component 路径 |
| T4.2 | `src/App.vue` | 修改 | 重写全局样式为移动端 |

### T5: 依赖安装 (依赖 T1.1)
| # | 命令 | 说明 |
|---|------|------|
| T5.1 | `rm -rf node_modules package-lock.json` | 清理旧依赖 |
| T5.2 | `npm install` | 安装新依赖 |

### T6: 验证 (依赖全部)
| # | 说明 |
|---|------|
| T6.1 | `npm run dev` 启动无报错 |
| T6.2 | 页面正常渲染 |
