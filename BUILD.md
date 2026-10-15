# 构建与运行

## 测试

```powershell
# 后端单元测试 + 集成测试（与 CI 一致）
cd backend
go test ./... -count=1

# 需要真实数据库的集成测试（默认被 integration build tag 隔离）
go test -tags integration ./internal/handler/ -run TestP04 -v

# 前端测试（vitest）
cd frontend-v2
npm test
```

测试分层：

| 层次 | 位置 | 说明 |
|------|------|------|
| 后端单元测试 | `backend/internal/**/*_test.go` | 表驱动用例，中文断言信息，不依赖外网 |
| 后端集成测试 | `backend/internal/router/update_integration_test.go` | 真实 `SetupRouter` + httptest；含**前后端 API 契约测试**：解析 `frontend-v2/src/api/*.js` 的全部请求，逐一校验后端路由已注册，接口漂移在 CI 即失败 |
| 数据库集成测试 | `//go:build integration` 标签文件 | 需真实 MySQL，默认不随 `go test ./...` 运行 |
| 前端单元测试 | `frontend-v2/src/**/*.test.js` | vitest + @vue/test-utils；API 层校验请求方法与路径，组件层覆盖更新卡片全部交互（检查/立即更新/取消/轮询跳转/失败刷新） |

---

## 开发运行

```powershell
.\dev.ps1
```

在项目根目录执行。脚本会同时拉起电脑端与移动端的前端开发服务器，并由后端以代理模式转发页面请求，`/api` 由后端直接处理。

访问地址：

- 电脑端：http://localhost:8080/
- 移动端：http://localhost:8080/mobile/

修改前端代码后由 vite 热更新，无需重新构建。按 `Ctrl+C` 停止，前端开发服务器会一并关闭。

---

## 热更新（同大版本自动升级）

推送 `vX.Y.Z` 标签发布后，已部署的系统可在**不手动下载整包**的情况下完成同大版本升级。

### 使用方式

管理员登录后进入「系统管理 / SMTP 配置」页的「系统更新」卡片：

1. **检查更新**：查询 GitHub Release 中是否存在同大版本且版本号更高的发布；
2. **立即更新**：自动执行完整更新流程，期间 http://localhost:8080 显示更新进度页，可实时监控或随时「终止更新并恢复旧版本」。

也可在部署目录手动执行 `update.exe`（Windows）/ `./update`（Linux）：

```powershell
update.exe --check   # 仅检查是否有可用更新，不停止服务
update.exe           # 完整更新流程
```

### 更新流程

```
启动 update.exe
  → 终止 server 进程（按可执行路径精确匹配，等待进程退出）
  → 在 8080 端口启动进度监控页面（下载/安装进度、终止更新按钮）
  → 检查 GitHub Release：同大版本且版本号更高才继续
  → 下载发布压缩包（Release 附带 .sha256 时自动校验）
  → 备份当前部署物到 updates/backup/<版本>-<时间戳>/
  → 解压并替换 server 可执行文件与前后端构建产物（dist/、dist-mobile/）
  → 释放 8080，自动重启 server
  → 健康检查确认存活后 update.exe 自动退出
```

任一步骤失败或用户点击终止：自动从备份回滚并重启旧版本，结果写入 `updates/last-result.json`（管理端「上次更新」展示），详细日志见 `updates/update.log`。

### 安全边界

- **只替换构建产物**：`server.exe`/`server`、`dist/`、`dist-mobile/`；数据库配置（`configs/`）、迁移脚本（`migrations/`）、日志一律不触碰；
- **同大版本才允许热更新**：大版本可能包含数据库结构变更，需手动下载整包并按发布说明迁移；
- 失败自动回滚：安装前必先备份，回滚后重启旧版本并验证存活。

### 网络要求

热更新需服务端可访问 GitHub：

| 域名 | 用途 |
|------|------|
| `api.github.com` | 查询 Release 列表 |
| `github.com`、`objects.githubusercontent.com` | 下载发布包 |

处于受限网络（表现为 DNS 解析失败 / 连接超时，检查更新返回「无法解析 api.github.com」或「连接 GitHub 超时」）时，可设置代理后重启服务与更新程序：

```bash
# Windows（启动 server.exe 前设置）
set HTTPS_PROXY=http://127.0.0.1:7890

# Linux
export HTTPS_PROXY=http://127.0.0.1:7890
```

Go 的 HTTP 客户端自动识别 `HTTPS_PROXY` / `HTTP_PROXY` / `NO_PROXY` 环境变量，API 查询与安装包下载均走代理。此外可设置 `GITHUB_TOKEN` 提升 API 额度（未认证限流 60 次/小时）。

检查更新失败仅影响更新检查本身，不影响系统正常运行。

### 版本号约定

`version.txt` 由 CI 自动写入，规则见 `build.yml` 的 `Resolve version` 步骤：

| 触发方式 | 写入内容 | 说明 |
|----------|----------|------|
| tag 推送 `v3.1.0` | `v3.1.0` | Release 的唯一来源 |
| 分支推送 / PR / 手动触发 | `v3.0.1-dev+abc1234` | 最近 `v*.*.*` 标签 + dev + 短 sha |

比较遵循语义化版本规则：正式版高于预发布版（`v3.1.0-dev` < `v3.1.0`），因此 dev 构建部署后仍能收到后续正式版的更新。

---

## 生产部署

### 发布包构建

推送标签（形如 `v2.1.0`）即触发发布流程，产出 Linux 与 Windows 两种整包并挂到 Release 页面。产物包含后端可执行文件、电脑端资源 `dist/`、移动端资源 `dist-mobile/`、数据库迁移脚本与启动脚本。

### Docker 部署

```bash
# 一键启动（含 MySQL）
docker-compose up -d

# 查看日志
docker-compose logs -f app
```

在 `.env` 文件中可修改 `DB_PASSWORD`、`DB_NAME`、`DB_PORT`、`APP_PORT`。

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

## 数据库迁移

```bash
./migrate.sh
```

执行前自动将受影响表备份到 `backup/<日期>/`，备份失败即中止。
