#!/bin/bash
# 构建并运行后端脚本

set -e

echo "=== 排班系统后端构建与运行 ==="

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo "错误: 未找到 Go 环境，请先安装 Go 1.21+"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo "Go 版本: $GO_VERSION"

# 进入后端目录
cd /workspace/schedule-system-v2/backend

# 下载依赖
echo "下载 Go 依赖..."
go mod download

# 构建后端
echo "构建后端..."
go build -o server ./cmd/server/main.go

# 检查构建结果
if [ ! -f "server" ]; then
    echo "构建失败，未生成可执行文件"
    exit 1
fi

echo "构建成功!"
echo "可执行文件: $(pwd)/server"

# 检查配置文件
if [ ! -f "configs/config.yaml" ]; then
    echo "警告: 配置文件不存在，使用默认配置"
fi

# 运行后端
echo ""
echo "启动后端服务器..."
echo "访问地址: http://localhost:8080"
echo "按 Ctrl+C 停止服务器"
echo ""

./server
