# 构建与运行

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
