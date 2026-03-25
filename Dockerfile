# ============================================
# 排班管理系统 V2 - 多阶段 Docker 构建
# ============================================

# ---- 阶段1: 构建前端 ----
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

# 先复制依赖文件，利用 Docker 缓存层
COPY frontend-v2/package*.json ./
RUN npm ci

# 再复制源码并构建
COPY frontend-v2/ ./
RUN npm run build

# ---- 阶段2: 构建后端 ----
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app/backend

# 先复制依赖文件，利用 Docker 缓存层
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 再复制源码并构建
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/server/main.go

# ---- 阶段3: 最终运行镜像 ----
FROM alpine:3.19

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制产物
COPY --from=backend-builder /app/backend/server .

# 复制前端构建产物（用于 SPA 静态服务）
COPY --from=frontend-builder /app/frontend/dist ./dist

# 复制数据库迁移脚本（安装向导使用）
COPY backend/migrations ./migrations

# 创建非 root 用户运行
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/system/installed || exit 1

# 环境变量默认值
ENV DB_HOST= \
    DB_PORT=3306 \
    DB_USER=root \
    DB_PASSWORD= \
    DB_NAME=schedule_system_v2 \
    PORT=8080

# 启动
CMD ["./server"]
