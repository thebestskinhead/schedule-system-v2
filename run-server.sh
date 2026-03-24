#!/bin/bash
# 运行后端服务器脚本

set -e

cd /workspace/schedule-system-v2/backend

echo "=== 排班系统后端 ==="
echo ""
echo "检查配置..."

if [ ! -f "configs/config.yaml" ]; then
    echo "警告: 配置文件不存在，将使用安装模式"
fi

echo ""
echo "启动服务器..."
echo "访问地址: http://localhost:8080"
echo ""

# 运行服务器
./server
