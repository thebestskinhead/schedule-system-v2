#!/bin/bash

# 排班系统一键部署脚本

echo "========================================="
echo "  排班系统 V2 - 部署脚本"
echo "========================================="
echo ""

# 构建前端
echo "[1/3] 构建前端..."
cd frontend-v2
npm install
npm run build
if [ $? -ne 0 ]; then
    echo "前端构建失败"
    exit 1
fi
echo "前端构建成功"
echo ""

# 复制到后端
echo "[2/3] 复制前端文件到后端..."
cd ..
rm -rf backend/dist
cp -r frontend-v2/dist backend/
echo "复制完成"
echo ""

# 编译后端
echo "[3/3] 编译后端..."
cd backend
go build -o server cmd/server/main.go
if [ $? -ne 0 ]; then
    echo "后端编译失败"
    exit 1
fi
echo "后端编译成功"
echo ""

echo "========================================="
echo "  部署完成！"
echo "========================================="
echo ""
echo "启动命令:"
echo "  cd backend && ./server"
echo ""
echo "访问地址:"
echo "  http://localhost:8080"
echo ""
echo "前后端已合并到同一域名下:"
echo "  - 前端页面: http://localhost:8080/"
echo "  - API接口: http://localhost:8080/api/v1/..."
echo ""
