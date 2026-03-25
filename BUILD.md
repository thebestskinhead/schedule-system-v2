# 构建说明

## 方式一：快速开发构建（推荐开发使用）

```bash
./dev-build.sh
```

- 构建前端并复制到 `backend/dist/`
- 启动后端即可访问完整系统

```bash
cd backend
go run ./cmd/server/main.go
```

访问 http://localhost:8080

---

## 方式二：全量构建（推荐生产部署）

```bash
./build.sh [输出目录]
```

示例：
```bash
./build.sh ./release
```

输出目录结构：
```
release/
├── server          # 后端可执行文件
├── dist/           # 前端静态资源
├── start.sh        # Linux/Mac 启动脚本
├── start.bat       # Windows 启动脚本
├── init_db.sql     # 数据库初始化脚本
└── migrations/     # 数据库迁移文件
```

启动：
```bash
cd release
./start.sh
```

---

## 方式三：Docker部署（推荐容器化部署）

### 一键启动（包含MySQL）

```bash
# 1. 复制环境变量配置（可选，修改密码等）
cp .env.example .env
# 编辑 .env 修改 DB_PASSWORD 等配置

# 2. 启动所有服务
docker-compose up -d

# 3. 查看日志
docker-compose logs -f app
```

访问 http://localhost:8080 ，首次使用需要完成安装向导。

### 自定义配置

在 `.env` 文件中修改：

```bash
DB_PASSWORD=your_password    # MySQL root 密码
DB_NAME=schedule_system_v2  # 数据库名
DB_PORT=3306                # MySQL 映射端口
APP_PORT=8080               # 应用映射端口
```

### 单独构建镜像

```bash
docker build -t schedule-system .
docker run -d \
  -p 8080:8080 \
  -e DB_HOST=your_mysql_host \
  -e DB_PASSWORD=your_password \
  --name schedule-app \
  schedule-system
```

### 常用命令

```bash
# 停止
docker-compose down

# 停止并清除数据
docker-compose down -v

# 重新构建（代码更新后）
docker-compose up -d --build

# 查看状态
docker-compose ps
```

---

## 方式四：前后端分离部署

### 前端单独部署
```bash
cd frontend-v2
npm run build
# 将 dist/ 目录部署到Nginx或CDN
```

### 后端单独部署
```bash
cd backend
go build -o server ./cmd/server/main.go
./server
```

后端API地址：http://localhost:8080/api/v1/

---

## 环境变量配置

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| DB_HOST | 数据库主机 | localhost |
| DB_PORT | 数据库端口 | 3306 |
| DB_USER | 数据库用户名 | root |
| DB_PASSWORD | 数据库密码 | Schedule@2024 |
| DB_NAME | 数据库名 | schedule_system_v2 |
| PORT | 服务端口 | 8080 |

---

## 推荐选择

| 场景 | 推荐方式 |
|------|----------|
| 本地开发 | 方式一：dev-build.sh |
| 单机部署 | 方式二：build.sh |
| 生产环境 | 方式三：Docker |
| 大型部署 | 方式四：前后端分离 + Nginx |
