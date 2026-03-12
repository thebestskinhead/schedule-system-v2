#!/bin/bash

# 排班系统构建脚本
# 用法: ./build.sh [输出目录]

set -e

OUTPUT_DIR=${1:-"./release"}
PROJECT_NAME="schedule-system-v2"

echo "================================"
echo "排班系统全量构建"
echo "================================"

# 清理旧构建
echo "[1/5] 清理旧构建..."
rm -rf $OUTPUT_DIR
mkdir -p $OUTPUT_DIR

# 构建前端
echo "[2/5] 构建前端..."
cd frontend
npm run build 2>&1 | tail -5
cd ..

# 构建后端
echo "[3/5] 构建后端..."
cd backend
go build -o ../$OUTPUT_DIR/server ./cmd/main.go 2>&1
cd ..

# 复制前端产物到后端dist目录
echo "[4/5] 整合前端资源..."
mkdir -p $OUTPUT_DIR/dist
cp -r frontend/dist/* $OUTPUT_DIR/dist/

# 复制必要文件
echo "[5/5] 复制配置文件..."
cp backend/scripts/init_db.sql $OUTPUT_DIR/
cp -r backend/migrations $OUTPUT_DIR/ 2>/dev/null || true

# 创建启动脚本
cat > $OUTPUT_DIR/start.sh << 'EOF'
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
EOF

chmod +x $OUTPUT_DIR/start.sh

# 创建Windows启动脚本
cat > $OUTPUT_DIR/start.bat << 'EOF'
@echo off
echo 正在启动排班系统...
if "%DB_PASSWORD%"=="" (
    echo 警告: 未设置 DB_PASSWORD 环境变量，使用默认密码
    set DB_PASSWORD=Schedule@2024
)
server.exe
EOF

echo ""
echo "================================"
echo "构建完成!"
echo "输出目录: $OUTPUT_DIR"
echo "================================"
echo ""
echo "文件清单:"
ls -lh $OUTPUT_DIR/
echo ""
echo "启动方式:"
echo "  1. 进入目录: cd $OUTPUT_DIR"
echo "  2. 运行: ./start.sh (Linux/Mac) 或 start.bat (Windows)"
echo ""
