# 多阶段构建 - 前端
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install

COPY frontend/ ./
RUN npm run build

# 多阶段构建 - 后端
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN go build -o server ./cmd/main.go

# 最终镜像
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# 复制后端可执行文件
COPY --from=backend-builder /app/backend/server .

# 复制前端构建产物
COPY --from=frontend-builder /app/frontend/dist ./dist

# 复制数据库脚本
COPY backend/scripts/init_db.sql ./
COPY backend/migrations ./migrations

# 暴露端口
EXPOSE 8080

# 环境变量
ENV DB_HOST=mysql
ENV DB_PORT=3306
ENV DB_USER=root
ENV DB_PASSWORD=Schedule@2024
ENV DB_NAME=schedule_system_v2

# 启动命令
CMD ["./server"]
