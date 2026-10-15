#!/usr/bin/env bash
# 组装发布包目录结构。由 CI 调用，TARGET 取 linux 或 windows。
#   linux   -> release-linux/    (server + update + dist + dist-mobile + migrations + 启动脚本)
#   windows -> release-windows/  (server.exe + update.exe + dist + dist-mobile + migrations + 启动脚本)
# 说明：VERSION 传入本次发布的版本号（如 v3.1.0），写入 version.txt 供热更新比对。
set -euo pipefail

TARGET="${TARGET:?TARGET 必须为 linux 或 windows}"
VERSION="${VERSION:-dev}"

case "$TARGET" in
  linux)
    PKG="release-linux"
    mkdir -p "$PKG/dist" "$PKG/dist-mobile" "$PKG/migrations"
    cp -r frontend-v2/dist/.   "$PKG/dist/"
    cp -r frontend-mobile/dist/. "$PKG/dist-mobile/"
    cp backend/scripts/init_db.sql "$PKG/"
    cp backend/migrations/*.sql "$PKG/migrations/" 2>/dev/null || true
    chmod +x "$PKG/server"

    # 热更新程序（由 CI 直接构建到本目录，与 server 同级，供管理端「系统更新」或手动 ./update 调用）
    [ -f "$PKG/update" ] && chmod +x "$PKG/update"
    echo "$VERSION" > "$PKG/version.txt"

    cat > "$PKG/start.sh" <<'EOF'
#!/bin/bash
cd "$(dirname "$0")"
if [ -z "$DB_PASSWORD" ]; then
  echo "警告: 未设置 DB_PASSWORD 环境变量，使用默认密码"
  export DB_PASSWORD="Schedule@2024"
fi
exec ./server
EOF
    chmod +x "$PKG/start.sh"

    cat > "$PKG/start.bat" <<'EOF'
@echo off
echo 正在启动排班系统...
if "%DB_PASSWORD%"=="" set DB_PASSWORD=Schedule@2024
server.exe
EOF

    cat > "$PKG/使用说明.txt" <<'EOF'
排班管理系统 - Linux 部署包

【系统要求】
- Linux x86_64
- 已安装 MySQL 8.0+

【启动步骤】
1. 执行: chmod +x start.sh server && ./start.sh
2. 浏览器访问 http://localhost:8080
3. 按向导配置数据库连接、初始化数据、创建管理员

【目录说明】
- server          后端可执行文件
- update          热更新程序（同大版本升级用，通常由管理端页面自动调用）
- dist/           电脑端前端资源
- dist-mobile/    移动端前端资源
- migrations/     数据库迁移脚本
- init_db.sql     数据库初始化脚本
- version.txt     当前版本号（热更新据此判断是否有新版本）

【热更新】
登录管理员账号后进入「系统管理 / SMTP 配置」页的「系统更新」卡片：
- 检查更新：查询 GitHub Release 中同大版本的新版本
- 立即更新：自动停止服务 -> 下载 -> 备份 -> 替换产物 -> 重启服务，
  期间 http://localhost:8080 显示更新进度页，可随时终止并恢复旧版本
- 跨大版本更新可能包含数据库结构变更，需手动下载整包升级
EOF
    ;;

  windows)
    PKG="release-windows"
    mkdir -p "$PKG/dist" "$PKG/dist-mobile" "$PKG/migrations"
    cp -r frontend-v2/dist/.   "$PKG/dist/"
    cp -r frontend-mobile/dist/. "$PKG/dist-mobile/"
    cp backend/scripts/init_db.sql "$PKG/"
    cp backend/migrations/*.sql "$PKG/migrations/" 2>/dev/null || true

    # 热更新程序（由 CI 直接构建到本目录，与 server.exe 同级）
    echo "$VERSION" > "$PKG/version.txt"

    cat > "$PKG/启动.bat" <<'EOF'
@echo off
chcp 65001 >nul
title 排班系统
echo ========================================
echo       排班管理系统
echo ========================================
echo.
echo 正在启动服务，访问地址: http://localhost:8080
echo 首次使用请在浏览器中完成安装向导
echo.
server.exe
pause
EOF

    cat > "$PKG/使用说明.txt" <<'EOF'
排班管理系统 - Windows 部署包

【系统要求】
- Windows 10/11 或 Windows Server 2016+
- 已安装 MySQL 8.0+

【启动步骤】
1. 双击 "启动.bat"
2. 浏览器访问 http://localhost:8080
3. 按向导配置数据库连接、初始化数据、创建管理员

【目录说明】
- server.exe      后端可执行文件
- update.exe      热更新程序（同大版本升级用，通常由管理端页面自动调用）
- dist/           电脑端前端资源
- dist-mobile/    移动端前端资源
- migrations/     数据库迁移脚本
- init_db.sql     数据库初始化脚本
- version.txt     当前版本号（热更新据此判断是否有新版本）

【热更新】
登录管理员账号后进入「系统管理 / SMTP 配置」页的「系统更新」卡片：
- 检查更新：查询 GitHub Release 中同大版本的新版本
- 立即更新：自动停止服务 -> 下载 -> 备份 -> 替换产物 -> 重启服务，
  期间 http://localhost:8080 显示更新进度页，可随时终止并恢复旧版本
- 跨大版本更新可能包含数据库结构变更，需手动下载整包升级
EOF
    ;;

  *)
    echo "未知 TARGET: $TARGET" >&2
    exit 1
    ;;
esac

echo "已生成 $PKG"
find "$PKG" -maxdepth 2 | head -40
