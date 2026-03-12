#!/bin/bash

# 快速开发构建脚本
# 将前端构建产物直接放入后端dist目录，方便测试

echo "================================"
echo "快速开发构建"
echo "================================"

# 构建前端
echo "[1/2] 构建前端..."
cd frontend
npm run build 2>&1 | grep -E "(error|built in)" || true
cd ..

# 复制到后端
echo "[2/2] 复制到后端dist目录..."
mkdir -p backend/dist
cp -r frontend-v2/dist/* backend/dist/

echo ""
echo "✅ 构建完成!"
echo "前端资源已复制到 backend/dist/"
echo ""
echo "启动后端即可访问:"
echo "  cd backend && go run ./cmd/main.go"
echo ""
