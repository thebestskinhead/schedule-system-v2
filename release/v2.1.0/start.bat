@echo off
echo 正在启动排班系统...
if "%DB_PASSWORD%"=="" (
    echo 警告: 未设置 DB_PASSWORD 环境变量，使用默认密码
    set DB_PASSWORD=Schedule@2024
)
server.exe
