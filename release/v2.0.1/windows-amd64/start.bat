@echo off
chcp 65001 >nul
cd /d "%~dp0"

REM 检查配置文件
if not exist "configs\config.yaml" (
    if exist "configs\config.example.yaml" (
        echo 未找到 configs\config.yaml，正在从模板创建...
        copy configs\config.example.yaml configs\config.yaml
        echo 请编辑 configs\config.yaml 填入配置信息后重新启动
        pause
        exit /b 1
    ) else (
        echo 未找到配置文件
        pause
        exit /b 1
    )
)

REM 检查前端文件
if not exist "dist\index.html" (
    echo 未找到前端文件 (dist\index.html)
    pause
    exit /b 1
)

echo 启动排班管理系统 v2.0.1...
echo 访问地址: http://localhost:8080
echo.

schedule-server.exe

if %errorlevel% neq 0 (
    echo.
    echo 程序异常退出，错误码: %errorlevel%
)

pause
