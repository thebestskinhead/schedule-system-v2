#!/bin/bash
# 排班管理系统启动脚本 (Linux)

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

# 检查配置文件
if [ ! -f "configs/config.yaml" ]; then
    if [ -f "configs/config.example.yaml" ]; then
        echo "⚠️  未找到 configs/config.yaml，正在从模板创建..."
        cp configs/config.example.yaml configs/config.yaml
        echo "请编辑 configs/config.yaml 填入配置信息后重新启动"
        exit 1
    else
        echo "❌ 未找到配置文件"
        exit 1
    fi
fi

# 检查前端文件
if [ ! -f "dist/index.html" ]; then
    echo "❌ 未找到前端文件 (dist/index.html)"
    exit 1
fi

echo "🚀 启动排班管理系统..."
exec ./schedule-server
