# 排班管理系统 v2.0.0 - Linux 发行版

## 目录结构

```
├── schedule-server          # 后端可执行文件
├── configs/
│   └── config.example.yaml  # 配置文件模板
├── dist/                    # 前端静态文件
├── docs/                    # 文档
├── start.sh                 # 启动脚本
└── README.md                # 本文件
```

## 快速开始

### 1. 配置

```bash
cd configs
cp config.example.yaml config.yaml
# 编辑 config.yaml，填入数据库连接信息等
vim config.yaml
```

### 2. 启动

```bash

start.bat
# 或直接运行
./schedule-server
```

### 3. 访问

打开浏览器访问: http://localhost:8080

默认管理员账号需要在数据库中手动设置 role='admin'。

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
| database.dbname | 数据库名 | schedule_system_v3 |
| jwt.secret | JWT 密钥 | |
| jwt.expire | Token 过期时间(小时) | 168 |

## 系统要求

- Windows (amd64)
- MySQL 5.7+
- 网络：需能访问数据库

## 文档

详细文档请查看 `docs/` 目录：

- `user-guide.md` - 用户操作手册
- `api.md` - API 接口文档
- `permission.md` - 权限系统说明
- `DEVELOPMENT_GUIDE.md` - 开发规范指南

## 更新日志

### v2.0.0

- 统一前后端接口格式
- 统一权限代码为冒号格式
- 新增临时权限申请系统
- 新增每周值班分工管理
- 支持多平台部署
