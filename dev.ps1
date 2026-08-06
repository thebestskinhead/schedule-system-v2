# 开发环境脚本：启动前后端开发服务器（后端 go run + 前端 vite 代理模式）
# 用法：在项目根目录执行  .\dev.ps1
# 前置：已通过 use-xxx 安装 Go 1.24.0 / Node 20.15.1
#
# 逻辑说明：
#   后端以 go run 直接运行时进入开发模式，读取 backend/configs/dev.yaml，
#   将页面请求反向代理到前端 vite 开发服务器（/api 仍由后端直接处理）。
#   本脚本负责同时拉起前端 vite 与后端，退出时会一并关闭前端进程树。

$ErrorActionPreference = "Stop"

# 激活工具链（仅当前会话生效）
if (Get-Command use-go -ErrorAction SilentlyContinue) {
    use-go 1.24.0
    use-node 20.15.1
} else {
    Write-Warning "未找到 use-go / use-node，请确认 PowerShell Profile 已加载 use-xxx 模块，或手动将 go/node 加入 PATH。"
}

$root       = $PSScriptRoot
$backendDir = Join-Path $root "backend"
$configPath = Join-Path $backendDir "configs\config.yaml"
$devConfigPath = Join-Path $backendDir "configs\dev.yaml"

if (-not (Test-Path $configPath)) {
    Write-Error "未找到开发配置 $configPath，请先创建开发环境测试配置文件（backend/configs/config.yaml）。"
    exit 1
}
if (-not (Test-Path $devConfigPath)) {
    Write-Error "未找到开发模式配置 $devConfigPath，请从 backend/configs/dev.example.yaml 复制创建。"
    exit 1
}

# 解析 dev.yaml（简单键值读取，忽略注释）
$devCfg = @{}
Get-Content $devConfigPath -Encoding UTF8 | ForEach-Object {
    if ($_ -match '^\s*(enabled|frontend|mobile)\s*:\s*"?([^"#]+?)"?\s*(#.*)?$') {
        $devCfg[$Matches[1]] = $Matches[2].Trim()
    }
}
if ($devCfg["enabled"] -and $devCfg["enabled"] -ne "true") {
    Write-Warning "dev.yaml 中 enabled 不是 true，后端将回退为静态文件托管模式。"
}

$procs = @()

function Start-Vite($name, $dir) {
    if (-not (Test-Path (Join-Path $dir "node_modules"))) {
        Write-Host "[$name] 首次运行，安装依赖 npm install ..." -ForegroundColor Yellow
        Push-Location $dir
        npm install
        Pop-Location
    }
    Write-Host "[$name] 启动 vite 开发服务器: $dir" -ForegroundColor Cyan
    $script:procs += Start-Process -FilePath "npm.cmd" -ArgumentList "run", "dev" -WorkingDirectory $dir -PassThru
}

try {
    if ($devCfg["frontend"]) {
        Start-Vite "电脑端" (Join-Path $root "frontend-v2")
    }
    if ($devCfg["mobile"]) {
        Start-Vite "移动端" (Join-Path $root "frontend-mobile")
    }

    Set-Location $backendDir

    Write-Host ""
    Write-Host "=== 启动后端开发服务器 (go run, 代理模式) ===" -ForegroundColor Cyan
    Write-Host "电脑端: http://localhost:8080/"        -ForegroundColor Green
    Write-Host "移动端: http://localhost:8080/mobile/" -ForegroundColor Green
    Write-Host "按 Ctrl+C 停止（前端开发服务器会一并关闭）" -ForegroundColor Yellow
    Write-Host ""

    # 显式标记开发模式（go run 自身也会被自动识别，这里是双保险）
    $env:SCHEDULE_DEV = "1"
    go run ./cmd/server/main.go
}
finally {
    # 关闭前端开发服务器（连同子进程树）
    foreach ($p in $procs) {
        if ($p -and -not $p.HasExited) {
            taskkill /PID $p.Id /T /F 2>$null | Out-Null
        }
    }
}
