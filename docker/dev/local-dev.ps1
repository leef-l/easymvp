<#
.SYNOPSIS
    EasyMVP 本地开发启动脚本：Docker 基础设施 + 本地 gf run 热重启 + 前端 pnpm dev
.DESCRIPTION
    MySQL/Redis 跑 Docker，Go 三端用 gf run 热重启，前端用 pnpm dev。
.PARAMETER NoInfra
    跳过 Docker 基础设施启动（已经在跑了）
.PARAMETER NoWeb
    不启动前端
.PARAMETER StopInfra
    停止 Docker 基础设施
.EXAMPLE
    .\docker\dev\local-dev.ps1              # 启动全部（基础设施 + 三端 + 前端）
    .\docker\dev\local-dev.ps1 -NoWeb       # 仅后端
    .\docker\dev\local-dev.ps1 -NoInfra     # 跳过 Docker（已在跑）
    .\docker\dev\local-dev.ps1 -StopInfra   # 停止 Docker 基础设施
#>
param(
    [switch]$NoInfra,
    [switch]$NoWeb,
    [switch]$StopInfra
)

$ErrorActionPreference = 'Stop'

$repoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..\..'))
$infraCompose = Join-Path $PSScriptRoot 'docker-compose.infra.yml'
$envFile = Join-Path $PSScriptRoot '.env'
$adminGo = Join-Path $repoRoot 'admin-go'
$webDir = Join-Path $repoRoot 'vue-vben-admin'

# --- 停止模式 ---
if ($StopInfra) {
    Write-Host '[local-dev] Stopping infra...' -ForegroundColor Yellow
    docker compose --project-name easymvp-infra --env-file $envFile -f $infraCompose down
    exit 0
}

# --- 启动基础设施 ---
if (-not $NoInfra) {
    # 清理可能冲突的旧容器（不管是哪个 compose 项目创建的）
    $conflictContainers = @('easymvp-mysql', 'easymvp-redis')
    foreach ($c in $conflictContainers) {
        $exists = docker ps -a --filter "name=^/${c}$" --format "{{.Names}}" 2>$null
        if ($exists) {
            Write-Host "[local-dev] Removing conflicting container: $c" -ForegroundColor Yellow
            docker rm -f $c 2>$null | Out-Null
        }
    }

    Write-Host '[local-dev] Starting Docker infra (MySQL + Redis)...' -ForegroundColor Cyan
    docker compose --project-name easymvp-infra --env-file $envFile -f $infraCompose up -d
    if ($LASTEXITCODE -ne 0) { throw 'Failed to start Docker infra' }

    # 等待 MySQL 就绪
    Write-Host '[local-dev] Waiting for MySQL...' -ForegroundColor Yellow
    $maxWait = 60
    $waited = 0
    while ($waited -lt $maxWait) {
        try {
            $result = docker exec easymvp-mysql mysqladmin ping -h 127.0.0.1 -uroot -proot --silent 2>&1 | Out-String
        } catch {
            $result = ''
        }
        if ($result -match 'alive') {
            Write-Host '[local-dev] MySQL is ready.' -ForegroundColor Green
            break
        }
        Start-Sleep -Seconds 2
        $waited += 2
    }
    if ($waited -ge $maxWait) {
        Write-Host '[local-dev] WARNING: MySQL may not be ready yet, proceeding anyway...' -ForegroundColor Red
    }
}

# --- 启动 Go 服务（gf run 热重启）---
$jobs = @()

$goApps = @(
    @{ Name = 'system'; Dir = Join-Path $adminGo 'app\system' },
    @{ Name = 'ai';     Dir = Join-Path $adminGo 'app\ai' },
    @{ Name = 'mvp';    Dir = Join-Path $adminGo 'app\mvp' }
)

foreach ($app in $goApps) {
    $appName = $app.Name
    $appDir = $app.Dir
    Write-Host "[local-dev] Starting $appName (gf run)..." -ForegroundColor Cyan
    $jobs += Start-Job -Name "go-$appName" -ScriptBlock {
        param($dir, $name)
        Set-Location $dir
        & gf run main.go 2>&1 | ForEach-Object { "[$name] $_" }
    } -ArgumentList $appDir, $appName
}

# --- 启动前端 ---
if (-not $NoWeb) {
    Write-Host '[local-dev] Starting web (pnpm dev)...' -ForegroundColor Cyan
    $jobs += Start-Job -Name 'web' -ScriptBlock {
        param($dir)
        Set-Location $dir
        $env:VITE_PROXY_SYSTEM_TARGET = 'http://localhost:9000'
        $env:VITE_PROXY_AI_TARGET = 'http://localhost:9001'
        $env:VITE_PROXY_MVP_TARGET = 'http://localhost:9002'
        & pnpm -F @vben/web-antd run dev 2>&1 | ForEach-Object { "[web] $_" }
    } -ArgumentList $webDir
}

# --- 输出状态 ---
Write-Host ''
Write-Host '========================================' -ForegroundColor Green
Write-Host '  EasyMVP Local Dev Environment' -ForegroundColor Green
Write-Host '========================================' -ForegroundColor Green
Write-Host "  MySQL:  127.0.0.1:3306  (Docker)" -ForegroundColor White
Write-Host "  Redis:  127.0.0.1:6379  (Docker)" -ForegroundColor White
Write-Host "  system: http://localhost:9000  (gf run)" -ForegroundColor White
Write-Host "  ai:     http://localhost:9001  (gf run)" -ForegroundColor White
Write-Host "  mvp:    http://localhost:9002  (gf run)" -ForegroundColor White
if (-not $NoWeb) {
    Write-Host "  web:    http://localhost:5173  (pnpm dev)" -ForegroundColor White
}
Write-Host '========================================' -ForegroundColor Green
Write-Host 'Press Ctrl+C to stop all services.' -ForegroundColor Yellow
Write-Host ''

# --- 流式输出日志，Ctrl+C 退出 ---
try {
    while ($true) {
        foreach ($job in $jobs) {
            Receive-Job -Job $job -ErrorAction SilentlyContinue | ForEach-Object {
                Write-Host $_
            }
        }
        Start-Sleep -Milliseconds 500
    }
} finally {
    Write-Host ''
    Write-Host '[local-dev] Stopping all services...' -ForegroundColor Yellow
    $jobs | ForEach-Object {
        Stop-Job -Job $_ -ErrorAction SilentlyContinue
        Remove-Job -Job $_ -Force -ErrorAction SilentlyContinue
    }
    Write-Host '[local-dev] All services stopped. Docker infra is still running.' -ForegroundColor Green
    Write-Host '[local-dev] To stop Docker: .\docker\dev\local-dev.ps1 -StopInfra' -ForegroundColor Yellow
}
