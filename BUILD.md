# 构建说明

## 方式一：快速开发构建（推荐开发使用）

```bash
./dev-build.sh
```

- 构建前端并复制到 `backend/dist/`
- 启动后端即可访问完整系统

```bash
cd backend
go run ./cmd/main.go
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

```bash
# 一键启动（包含MySQL）
docker-compose up -d

# 或者单独构建镜像
docker build -t schedule-system .
docker run -p 8080:8080 schedule-system
```

访问 http://localhost:8080

---

## 方式四：前后端分离部署

### 前端单独部署
```bash
cd frontend
npm run build
# 将 dist/ 目录部署到Nginx或CDN
```

### 后端单独部署
```bash
cd backend
go build -o server ./cmd/main.go
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
