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
$adminGoEnv = Join-Path $adminGo '.env'
$webDir = Join-Path $repoRoot 'vue-vben-admin'

function Get-EnvMap {
    param([string]$Path)

    $map = @{}
    if (-not (Test-Path $Path)) {
        return $map
    }
    foreach ($line in Get-Content -Path $Path) {
        $trimmed = $line.Trim()
        if ([string]::IsNullOrWhiteSpace($trimmed) -or $trimmed.StartsWith('#')) {
            continue
        }
        $parts = $trimmed -split '=', 2
        if ($parts.Count -eq 2) {
            $map[$parts[0]] = $parts[1]
        }
    }
    return $map
}

function Get-EnvValue {
    param(
        [hashtable]$EnvMap,
        [string]$Name,
        [string]$Default
    )

    if ($EnvMap.ContainsKey($Name) -and -not [string]::IsNullOrWhiteSpace($EnvMap[$Name])) {
        return $EnvMap[$Name]
    }
    return $Default
}

function To-PosixPath {
    param([string]$Path)

    return ($Path -replace '\\', '/')
}

function Resolve-LocalHostName {
    param(
        [string]$Host,
        [string]$Fallback = '127.0.0.1'
    )

    if ([string]::IsNullOrWhiteSpace($Host)) {
        return $Fallback
    }
    if ($Host -in @('mysql', 'redis', 'localhost', '127.0.0.1')) {
        return '127.0.0.1'
    }
    return $Host
}

function Resolve-LocalAddress {
    param(
        [string]$Address,
        [string]$DefaultPort
    )

    if ([string]::IsNullOrWhiteSpace($Address)) {
        return "127.0.0.1:$DefaultPort"
    }

    $parts = $Address.Split(':')
    if ($parts.Count -ge 2) {
        $host = Resolve-LocalHostName -Host ($parts[0].Trim())
        $port = $parts[$parts.Count - 1].Trim()
        if ([string]::IsNullOrWhiteSpace($port)) {
            $port = $DefaultPort
        }
        return "${host}:$port"
    }

    return "$(Resolve-LocalHostName -Host $Address):$DefaultPort"
}

function Resolve-LocalLogBasePath {
    param(
        [hashtable]$EnvMap
    )

    $raw = Get-EnvValue -EnvMap $EnvMap -Name 'GF_LOG_PATH' -Default (Join-Path $adminGo 'logs')
    if ($raw.StartsWith('/workspace/admin-go')) {
        return Join-Path $adminGo 'logs'
    }
    return $raw
}

function Write-RuntimeConfig {
    param(
        [string]$AppName,
        [int]$Port,
        [hashtable]$EnvMap
    )

    $runtimeConfigDir = Join-Path $adminGo '.runtime-config'
    $logBasePath = Resolve-LocalLogBasePath -EnvMap $EnvMap
    $logDir = Join-Path $logBasePath $AppName
    $configFile = Join-Path $runtimeConfigDir "$AppName.yaml"

    New-Item -ItemType Directory -Force -Path $runtimeConfigDir | Out-Null
    New-Item -ItemType Directory -Force -Path $logDir | Out-Null

    $dbHost = Resolve-LocalHostName -Host (Get-EnvValue -EnvMap $EnvMap -Name 'DB_HOST' -Default '127.0.0.1')
    $dbPort = Get-EnvValue -EnvMap $EnvMap -Name 'DB_PORT' -Default '3306'
    $dbUser = Get-EnvValue -EnvMap $EnvMap -Name 'DB_USER' -Default 'easymvp'
    $dbPassword = Get-EnvValue -EnvMap $EnvMap -Name 'DB_PASSWORD' -Default ''
    $dbName = Get-EnvValue -EnvMap $EnvMap -Name 'DB_NAME' -Default 'easymvp'

    $redisConfig = ''
    $redisAddr = Get-EnvValue -EnvMap $EnvMap -Name 'REDIS_ADDR' -Default ''
    if (-not [string]::IsNullOrWhiteSpace($redisAddr)) {
        $localRedisAddr = Resolve-LocalAddress -Address $redisAddr -DefaultPort '6379'
        $redisPass = Get-EnvValue -EnvMap $EnvMap -Name 'REDIS_PASS' -Default ''
        $redisConfig = @"
redis:
  default:
    address: "$localRedisAddr"
    pass: "$redisPass"
    db: 0

"@
    }

$config = @"
server:
  address: ":$Port"
  openapiPath: "/api.json"
  swaggerPath: "/swagger"

database:
  default:
    link: "mysql:$dbUser:$dbPassword@tcp($dbHost:$dbPort)/$dbName?charset=utf8mb4&loc=Local&parseTime=true"
    debug: false

${redisConfig}logger:
  path: "$(To-PosixPath -Path $logDir)"
  level: "$(Get-EnvValue -EnvMap $EnvMap -Name 'GF_LOG_LEVEL' -Default 'all')"
  stdout: $(Get-EnvValue -EnvMap $EnvMap -Name 'GF_LOG_STDOUT' -Default 'false')
  rotateSize: "$(Get-EnvValue -EnvMap $EnvMap -Name 'GF_LOG_ROTATE_SIZE' -Default '100M')"
  rotateExpire: "$(Get-EnvValue -EnvMap $EnvMap -Name 'GF_LOG_ROTATE_EXPIRE' -Default '7d')"
  rotateBackupLimit: $(Get-EnvValue -EnvMap $EnvMap -Name 'GF_LOG_ROTATE_BACKUP_LIMIT' -Default '10')
  stStatus: 0

jwt:
  secret: "$(Get-EnvValue -EnvMap $EnvMap -Name 'JWT_SECRET' -Default 'easymvp-secret-key')"
  expire: 24
"@

    Set-Content -Path $configFile -Value $config -Encoding UTF8
    return $configFile
}

if (Test-Path $envFile) {
    Copy-Item -Path $envFile -Destination $adminGoEnv -Force
}
$envMap = Get-EnvMap -Path $envFile

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
    $mysqlRootPassword = Get-EnvValue -EnvMap $envMap -Name 'MYSQL_ROOT_PASSWORD' -Default 'root'
    while ($waited -lt $maxWait) {
        try {
            $result = docker exec easymvp-mysql mysqladmin ping -h 127.0.0.1 -uroot "-p$mysqlRootPassword" --silent 2>&1 | Out-String
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
    @{ Name = 'system'; Dir = Join-Path $adminGo 'app\system'; Port = 9000 },
    @{ Name = 'ai';     Dir = Join-Path $adminGo 'app\ai'; Port = 9001 },
    @{ Name = 'mvp';    Dir = Join-Path $adminGo 'app\mvp'; Port = 9002 }
)

foreach ($app in $goApps) {
    $appName = $app.Name
    $appDir = $app.Dir
    $configPath = Write-RuntimeConfig -AppName $appName -Port $app.Port -EnvMap $envMap
    Write-Host "[local-dev] Starting $appName (gf run)..." -ForegroundColor Cyan
    $jobs += Start-Job -Name "go-$appName" -ScriptBlock {
        param($dir, $name, $cfg)
        Set-Location $dir
        $env:GF_GCFG_FILE = $cfg
        & gf run main.go 2>&1 | ForEach-Object { "[$name] $_" }
    } -ArgumentList $appDir, $appName, $configPath
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
