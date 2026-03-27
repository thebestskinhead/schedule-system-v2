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
echo "[1/6] 清理旧构建..."
rm -rf $OUTPUT_DIR
mkdir -p $OUTPUT_DIR

# 构建电脑端前端
echo "[2/6] 构建电脑端前端..."
cd frontend-v2
npm run build 2>&1 | tail -5
cd ..

# 构建移动端前端
echo "[3/6] 构建移动端前端..."
cd frontend-mobile
npm run build 2>&1 | tail -5
cd ..

# 构建后端
echo "[4/6] 构建后端..."
cd backend
go build -o ../$OUTPUT_DIR/server ./cmd/server/main.go 2>&1
cd ..

# 复制前端产物到后端dist目录
echo "[5/6] 整合前端资源..."
# 电脑端
mkdir -p $OUTPUT_DIR/dist
cp -r frontend-v2/dist/* $OUTPUT_DIR/dist/
# 移动端
mkdir -p $OUTPUT_DIR/dist-mobile
cp -r frontend-mobile/dist/* $OUTPUT_DIR/dist-mobile/

# 复制必要文件
echo "[6/6] 复制配置文件..."
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
echo "前端资源:"
echo "  电脑端: $OUTPUT_DIR/dist/"
echo "  移动端: $OUTPUT_DIR/dist-mobile/"
echo ""
echo "启动方式:"
echo "  1. 进入目录: cd $OUTPUT_DIR"
echo "  2. 运行: ./start.sh (Linux/Mac) 或 start.bat (Windows)"
echo ""
echo "访问地址:"
echo "  电脑端: http://localhost:8080/"
echo "  移动端: http://localhost:8080/mobile/"
echo ""
