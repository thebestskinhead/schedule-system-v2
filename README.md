<div align="center">

<img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go">
<img src="https://img.shields.io/badge/Vue-3-4FC08D?style=flat-square&logo=vue.js&logoColor=white" alt="Vue3">
<img src="https://img.shields.io/badge/Element_Plus-2.x-409EFF?style=flat-square&logo=element&logoColor=white" alt="Element Plus">
<img src="https://img.shields.io/badge/MySQL-8.0-4479A1?style=flat-square&logo=mysql&logoColor=white" alt="MySQL">
<img src="https://img.shields.io/badge/License-MIT-yellow?style=flat-square" alt="License">

# 📅 排班管理系统 V2

> 🎓 专为高校社团/部门设计的智能排班管理系统

一个现代化的值班安排解决方案，支持多部门管理、无课表导入、自动排班算法和临时权限授权。

[🚀 快速开始](#-快速开始) • [📖 使用指南](#-使用指南) • [📚 文档](#-文档目录) • [🔧 部署](#-部署指南)

</div>

---

## ✨ 功能特性

| 功能模块 | 特性描述 | 状态 |
|---------|---------|------|
| 🗓️ **智能排班** | 基于无课表数据自动生成最优排班方案 | ✅ |
| 📥 **无课表导入** | 支持教务系统抓取、Excel导入、手动录入 | ✅ |
| 👥 **多部门管理** | 支持办公室、竞赛部、项目部、科普部等多个部门 | ✅ |
| 🔐 **灵活权限** | RBAC + 临时权限的混合权限模型 | ✅ |
| 📊 **每周分工** | 设置各部门本周值班日期安排 | ✅ |
| ✉️ **邮件通知** | 支持密码重置邮件发送 | ✅ |

---

## 🏗️ 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                        前端 (Vue3)                          │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐          │
│  │ 首页    │ │ 排班管理│ │ 无课表  │ │ 用户管理│          │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      后端 (Go/Gin)                          │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐          │
│  │ 用户模块│ │ 排班模块│ │ 权限模块│ │ 通知模块│          │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      数据存储 (MySQL)                       │
└─────────────────────────────────────────────────────────────┘
```

---

## 🚀 快速开始

### 环境要求

- **后端**: Go 1.21+
- **前端**: Node.js 18+
- **数据库**: MySQL 8.0+
- **操作系统**: Linux/macOS/Windows

### 1️⃣ 克隆项目

```bash
git clone https://gitee.com/skinhand/schedule-system-v2.git
cd schedule-system-v2
```

### 2️⃣ 数据库配置

创建 MySQL 数据库：

```sql
CREATE DATABASE schedule_system_v2 
CHARACTER SET utf8mb4 
COLLATE utf8mb4_unicode_ci;
```

### 3️⃣ 后端启动

```bash
cd backend
cp config/config.example.yaml config/config.yaml
# 编辑 config.yaml 配置数据库连接信息
go mod download
go run ./cmd/server/main.go
```

### 4️⃣ 前端启动

```bash
cd frontend-v2
npm install
npm run dev
```

### 5️⃣ 访问系统

打开浏览器访问 `http://localhost:8080`，按照安装向导完成初始化。

---

## 📖 使用指南

### 🔐 权限角色

| 角色 | 权限范围 | 典型用户 |
|------|---------|---------|
| 👑 系统管理员 | 所有功能 | 系统维护人员 |
| 🏢 办公室管理员 | 所有部门管理 + 每周分工 | 办公室主任 |
| 👤 部门管理员 | 本部门排班和用户管理 | 部长/副部长 |
| 👥 部门成员 | 个人功能 | 普通成员 |

### 📝 快速上手

1. **注册账号** - 使用学号注册，完善个人信息
2. **录入无课表** - 手动录入或从教务系统导入
3. **生成排班** - 管理员根据无课表自动生成排班
4. **确认值班** - 成员查看值班安排并确认

更多详细使用说明请参考 [用户手册](docs/user-guide.md)

---

## 🛠️ 技术栈

### 后端

| 技术 | 版本 | 说明 |
|------|------|------|
| [Go](https://golang.org/) | 1.21+ | 高性能编程语言 |
| [Gin](https://gin-gonic.com/) | v1.9+ | Web框架 |
| [MySQL](https://mysql.com/) | 8.0+ | 关系型数据库 |
| [sqlx](https://github.com/jmoiron/sqlx) | - | 数据库操作 |
| [JWT](https://jwt.io/) | - | 身份认证 |

### 前端

| 技术 | 版本 | 说明 |
|------|------|------|
| [Vue 3](https://vuejs.org/) | 3.4+ | 前端框架 |
| [Vite](https://vitejs.dev/) | 5.x | 构建工具 |
| [Element Plus](https://element-plus.org/) | 2.x | UI组件库 |
| [Pinia](https://pinia.vuejs.org/) | 2.x | 状态管理 |

---

## 📁 项目结构

```
schedule-system-v2/
├── 📂 backend/              # Go后端
│   ├── 📂 cmd/server/       # 程序入口
│   ├── 📂 internal/         # 内部代码
│   │   ├── 📂 handler/      # HTTP处理器
│   │   ├── 📂 service/      # 业务逻辑
│   │   ├── 📂 dao/          # 数据访问
│   │   ├── 📂 model/        # 数据模型
│   │   ├── 📂 middleware/   # 中间件
│   │   └── 📂 router/       # 路由
│   ├── 📂 config/           # 配置文件
│   └── 📂 static/           # 前端静态资源
├── 📂 frontend-v2/          # Vue3前端
│   ├── 📂 src/
│   │   ├── 📂 views/        # 页面
│   │   ├── 📂 components/   # 组件
│   │   ├── 📂 api/          # API接口
│   │   ├── 📂 stores/       # 状态管理
│   │   └── 📂 router/       # 路由
│   └── 📄 package.json
├── 📂 docs/                 # 文档
└── 📄 README.md
```

---

## 📚 文档目录

| 文档 | 说明 | 适用人群 |
|------|------|---------|
| [📘 API接口文档](docs/api.md) | 完整的API接口说明 | 开发者 |
| [🔐 权限系统文档](docs/permission.md) | 权限体系详细说明 | 管理员 |
| [📖 用户手册](docs/user-guide.md) | 操作使用指南 | 所有用户 |
| [✨ 功能介绍](docs/features.md) | 功能模块详细说明 | 所有用户 |

---

## 🐳 部署指南

### Docker 部署

```bash
# 使用 Docker Compose
docker-compose up -d
```

### 生产环境部署

```bash
# 构建前端
cd frontend-v2
npm run build

# 构建后端
cd ../backend
go build -o server ./cmd/server/main.go

# 运行
./server
```

更多部署方式请参考 [BUILD.md](BUILD.md)

---

## 🖼️ 界面预览

### 首页
系统仪表盘，展示本周排班、快速操作入口。

### 无课表管理
可视化录入界面，支持30周的无课时间管理。

### 排班管理
智能排班算法，支持预览和手动调整。

### 用户管理
部门成员管理，角色分配，临时权限授权。

---

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

---

## 📄 许可证

本项目采用 [MIT](LICENSE) 许可证

---

## 👨‍💻 作者

**skinhand** - [Gitee](https://gitee.com/skinhand)

---

<div align="center">

**⭐ 如果这个项目对你有帮助，请给个 Star 支持一下！**

</div>
