#!/bin/bash

# 排班系统启动脚本

echo "正在启动排班系统..."

# 检查环境变量
if [ -z "$DB_PASSWORD" ]; then
    echo "警告: 未设置 DB_PASSWORD 环境变量，使用默认密码"
    export DB_PASSWORD="Schedule@2024"
fi

# 启动服务
./server
