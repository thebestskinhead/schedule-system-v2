#!/bin/bash

# Windows一键端构建脚本
# 生成包含exe和前端资源的完整包

set -e

OUTPUT_DIR="./windows-release"
VERSION="v2.0"

echo "================================"
echo "Windows一键端构建"
echo "================================"

# 清理旧构建
echo "[1/6] 清理旧构建..."
rm -rf $OUTPUT_DIR
mkdir -p $OUTPUT_DIR

# 构建前端
echo "[2/6] 构建前端..."
cd frontend-v2
npm run build 2>&1 | tail -3
cd ..

# 交叉编译Windows后端
echo "[3/6] 交叉编译Windows后端..."
cd backend
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ../$OUTPUT_DIR/排班系统.exe ./cmd/server/main.go
cd ..

# 复制前端资源
echo "[4/6] 复制前端资源..."
mkdir -p $OUTPUT_DIR/dist
cp -r frontend-v2/dist/* $OUTPUT_DIR/dist/

# 复制数据库脚本
echo "[5/6] 复制配置文件..."
cp backend/scripts/init_db.sql $OUTPUT_DIR/
cp backend/migrations/*.sql $OUTPUT_DIR/ 2>/dev/null || true

# 创建Windows启动脚本
echo "[6/6] 创建启动脚本..."
cat > "$OUTPUT_DIR/启动.bat" << 'EOF'
@echo off
chcp 65001 >nul
title 排班系统
color 0a

echo ========================================
echo      排班管理系统 V2.0
echo ========================================
echo.
echo 正在启动服务...
echo.
echo 首次使用请在浏览器中完成安装向导
echo.
echo 访问地址: http://localhost:8080
echo.
echo 按 Ctrl+C 停止服务
echo.

排班系统.exe

pause
EOF

# 创建配置说明
cat > "$OUTPUT_DIR/使用说明.txt" << 'EOF'
============================================
     排班管理系统 V2.0 - Windows一键端
============================================

【系统要求】
- Windows 10/11 或 Windows Server 2016+
- MySQL 8.0+ (需要提前安装)
- 内存: 至少 2GB RAM

【安装步骤】

第1步：安装MySQL
  1. 下载并安装 MySQL 8.0+ 
     https://dev.mysql.com/downloads/installer/
  2. 记住设置的root密码

第2步：启动系统
  1. 双击 "启动.bat" 运行
  2. 等待服务启动

第3步：完成安装向导
  1. 浏览器访问 http://localhost:8080
  2. 按照向导配置数据库连接
  3. 创建管理员账号
  4. 开始使用系统

【安装向导步骤】

1. 配置数据库
   - 输入MySQL连接信息
   - 系统自动测试连接并创建数据库

2. 初始化数据
   - 系统自动创建数据表

3. 创建管理员
   - 设置管理员账号信息

4. 完成
   - 进入系统登录页面

【重新安装】
如需重新安装：
  1. 删除MySQL中的 schedule_system_v2 数据库
  2. 重新运行 "启动.bat"
  3. 再次完成安装向导

【技术支持】
如有问题，请查看控制台日志
============================================
EOF

# 创建快捷配置脚本
cat > "$OUTPUT_DIR/配置数据库.bat" << 'EOF'
@echo off
chcp 65001 >nul
echo ========================================
echo      数据库配置
echo ========================================
echo.
echo 当前配置:
echo   主机: localhost
echo   端口: 3306
echo   用户: root
echo   密码: %DB_PASSWORD% (若未设置则使用默认)
echo.
echo 要修改配置，请编辑 启动.bat 文件中的以下行:
echo   set DB_PASSWORD=你的密码
echo   set PORT=8080
echo.
pause
EOF

echo ""
echo "================================"
echo "Windows一键端构建完成!"
echo "================================"
echo ""
echo "输出目录: $OUTPUT_DIR"
echo ""
echo "文件清单:"
ls -lh $OUTPUT_DIR/
echo ""
echo "使用方法:"
echo "  1. 将 $OUTPUT_DIR 复制到Windows电脑"
echo "  2. 确保已安装MySQL 8.0+"
echo "  3. 双击 启动.bat 运行"
echo ""
echo "可选: 打包成zip"
echo "  cd $OUTPUT_DIR && zip -r ../排班系统-$VERSION-windows.zip ."
echo ""
