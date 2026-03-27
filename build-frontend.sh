#!/bin/bash
# 前端构建脚本 - 构建电脑端和移动端

set -e

echo "=== 开始构建前端 ==="

BACKEND_DIR="/workspace/schedule-system-v2/backend"

# 构建电脑端前端
echo ""
echo ">>> 构建电脑端前端..."
cd /workspace/schedule-system-v2/frontend-v2

# 安装依赖
echo "安装电脑端依赖..."
npm install

# 构建
echo "构建电脑端..."
npm run build

# 复制到后端目录
echo "复制到后端 dist 目录..."
rm -rf $BACKEND_DIR/dist
cp -r dist $BACKEND_DIR/dist

echo "电脑端构建成功!"

# 构建移动端前端
echo ""
echo ">>> 构建移动端前端..."
cd /workspace/schedule-system-v2/frontend-mobile

# 安装依赖
echo "安装移动端依赖..."
npm install

# 构建
echo "构建移动端..."
npm run build

# 复制到后端目录
echo "复制到后端 dist-mobile 目录..."
rm -rf $BACKEND_DIR/dist-mobile
cp -r dist $BACKEND_DIR/dist-mobile

echo "移动端构建成功!"

echo ""
echo "=== 前端构建完成 ==="
echo "后端目录内容:"
ls -la $BACKEND_DIR/ | grep dist
