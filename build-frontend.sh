#!/bin/bash
# 前端构建脚本

set -e

echo "=== 开始构建前端 ==="

# 进入前端目录
cd /workspace/schedule-system-v2/frontend-v2

# 安装依赖
echo "安装依赖..."
npm install

# 构建
echo "构建前端..."
npm run build

# 检查构建产物
if [ -d "dist" ]; then
    echo "构建成功!"
    echo "构建产物:"
    ls -la dist/
else
    echo "构建失败，dist 目录不存在"
    exit 1
fi

# 复制到后端目录
echo "复制到后端 dist 目录..."
rm -rf /workspace/schedule-system-v2/backend/dist
cp -r dist /workspace/schedule-system-v2/backend/dist

echo "=== 前端构建完成 ==="
echo "后端 dist 目录内容:"
ls -la /workspace/schedule-system-v2/backend/dist/
