# 排班管理系统 v2.0.1 - Windows 发行版

## 目录结构

```
├── schedule-server.exe       # 后端可执行文件
├── configs/
│   └── config.example.yaml   # 配置文件模板
├── dist/                     # 前端静态文件
├── docs/                     # 文档
├── start.bat                 # 启动脚本
└── README.md                 # 本文件
```

## 快速开始

### 1. 配置

进入 `configs` 目录，复制 `config.example.yaml` 为 `config.yaml`，填入数据库连接信息。

### 2. 启动

双击 `start.bat` 或在命令行中运行：

```cmd
start.bat
```

### 3. 访问

打开浏览器访问: http://localhost:8080

首次使用会自动进入安装向导，按提示完成数据库初始化和管理员账号创建。

## 配置说明

主要配置项（`configs/config.yaml`）：

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| server.port | 服务端口 | 8080 |
| server.mode | 运行模式 (debug/release) | release |
| database.host | 数据库地址 | localhost |
| database.port | 数据库端口 | 3306 |
| database.user | 数据库用户 | root |
| database.password | 数据库密码 | |
| database.dbname | 数据库名 | schedule_system_v2 |
| jwt.secret | JWT 密钥 | |
| jwt.expire | Token 过期时间(小时) | 168 |

## 系统要求

- Windows 10/11 或 Windows Server 2016+
- MySQL 8.0+
- 网络：需能访问数据库

## 文档

详细文档请查看 `docs/` 目录：

- `api.md` - API 接口文档
- `permission.md` - 权限系统说明
- `user-guide.md` - 用户操作手册
- `features.md` - 功能详细介绍
- `DEVELOPMENT_GUIDE.md` - 开发规范指南

## 更新日志

### v2.0.1

- 修复文档与代码不一致问题，全面更新 API 接口文档
- 修复 Docker 部署文件中的路径错误
- 修复构建脚本中的前端目录引用（frontend → frontend-v2）
- 优化 Dockerfile：多阶段缓存层、CGO 静态编译、非 root 运行、健康检查
- 优化 docker-compose：支持 .env 配置、MySQL 字符集和时区
- 更新 .gitignore：排除 release 目录中的二进制文件和配置文件

### v2.0.0

- 统一前后端接口格式
- 统一权限代码为冒号格式
- 新增临时权限申请系统
- 新增每周值班分工管理
- 支持多平台部署
