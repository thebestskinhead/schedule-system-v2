# 构建脚本：编译前端（电脑端 + 移动端）与后端，
# 并将前端编译产物分别放到 backend/dist 和 backend/dist-mobile。
# 用法：在项目根目录执行  .\build.ps1
# 前置：已通过 use-xxx 安装 Go 1.24.0 / Node 20.15.1

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
$feDir      = Join-Path $root "frontend-v2"
$mobileDir  = Join-Path $root "frontend-mobile"

function Invoke-Step {
    param([scriptblock]$Cmd, [string]$Name)
    Write-Host ">>> $Name" -ForegroundColor Cyan
    & $Cmd
    if ($LASTEXITCODE -ne 0) {
        Write-Error "$Name 失败 (exit $LASTEXITCODE)"
        exit $LASTEXITCODE
    }
}

# 1/4 构建电脑端前端
Write-Host "=== [1/4] 构建电脑端前端 ===" -ForegroundColor Cyan
Set-Location $feDir
Invoke-Step { npm install } "npm install (frontend-v2)"
Invoke-Step { npm run build } "npm run build (frontend-v2)"

$destDist = Join-Path $backendDir "dist"
if (Test-Path $destDist) { Remove-Item $destDist -Recurse -Force }
Copy-Item (Join-Path $feDir "dist") $destDist -Recurse
Write-Host "已复制前端产物 -> $destDist" -ForegroundColor Green

# 2/4 构建移动端前端
Write-Host "=== [2/4] 构建移动端前端 ===" -ForegroundColor Cyan
Set-Location $mobileDir
Invoke-Step { npm install } "npm install (frontend-mobile)"
Invoke-Step { npm run build } "npm run build (frontend-mobile)"

$destMobile = Join-Path $backendDir "dist-mobile"
if (Test-Path $destMobile) { Remove-Item $destMobile -Recurse -Force }
Copy-Item (Join-Path $mobileDir "dist") $destMobile -Recurse
Write-Host "已复制移动端产物 -> $destMobile" -ForegroundColor Green

# 3/4 构建后端
Write-Host "=== [3/4] 构建后端 ===" -ForegroundColor Cyan
Set-Location $backendDir
Invoke-Step { go mod download } "go mod download"
Invoke-Step { go build -ldflags="-s -w" -o server.exe ./cmd/server/main.go } "go build"

# 4/4 完成
Write-Host "=== [4/4] 构建完成 ===" -ForegroundColor Green
Write-Host "可执行文件 : $backendDir\server.exe" -ForegroundColor White
Write-Host "前端电脑端 : $destDist" -ForegroundColor White
Write-Host "前端移动端 : $destMobile" -ForegroundColor White
Write-Host "运行后端   : cd backend && .\server.exe" -ForegroundColor White
