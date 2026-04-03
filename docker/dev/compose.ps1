param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$ComposeArgs
)

$ErrorActionPreference = 'Stop'

$repoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..\..'))
$composeFile = Join-Path $PSScriptRoot 'docker-compose.cn.yml'
$sourceEnv = Join-Path $PSScriptRoot '.env'
$adminGoEnv = Join-Path $repoRoot 'admin-go\.env'

if (-not (Test-Path $sourceEnv)) {
    throw "Missing docker dev env file: $sourceEnv"
}

Copy-Item -Path $sourceEnv -Destination $adminGoEnv -Force

function Get-EnvMap {
    param([string]$Path)

    $map = @{}
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

function Test-ProfileEnabled {
    param(
        [string[]]$Args,
        [string]$ProfileName
    )

    for ($i = 0; $i -lt $Args.Count; $i++) {
        if ($Args[$i] -eq '--profile' -and $i + 1 -lt $Args.Count -and $Args[$i + 1] -eq $ProfileName) {
            return $true
        }
    }

    return $false
}

function Write-StartupGuide {
    Write-Host 'Docker dev services started.' -ForegroundColor Green
    Write-Host 'Default command starts core services: mysql, system, ai, mvp.' -ForegroundColor Yellow
    Write-Host 'Start frontend too: .\docker\dev\compose.ps1 --profile frontend up -d' -ForegroundColor Yellow
    Write-Host 'Start all services: .\docker\dev\compose.ps1 --profile frontend --profile ai-runtime up -d' -ForegroundColor Yellow
}

function Invoke-DevBuild {
    param(
        [hashtable]$EnvMap,
        [bool]$BuildFrontend = $false,
        [bool]$BuildAiRuntime = $false
    )

    $backendArgs = @(
        'build',
        '-f', (Join-Path $repoRoot 'docker\build\Dockerfile.admin-go.dev'),
        (Join-Path $repoRoot 'admin-go'),
        '-t', 'easymvp-admin-go-dev:latest',
        '--build-arg', "GO_BASE_IMAGE=$(Get-EnvValue -EnvMap $EnvMap -Name 'GO_BASE_IMAGE' -Default 'golang:1.25-bookworm')",
        '--build-arg', "GO_PROXY=$(Get-EnvValue -EnvMap $EnvMap -Name 'GO_PROXY' -Default 'https://goproxy.cn,direct')",
        '--build-arg', "PIP_INDEX_URL=$(Get-EnvValue -EnvMap $EnvMap -Name 'PIP_INDEX_URL' -Default 'https://pypi.org/simple')",
        '--build-arg', "APT_MIRROR=$(Get-EnvValue -EnvMap $EnvMap -Name 'APT_MIRROR' -Default '')",
        '--build-arg', "APT_SECURITY_MIRROR=$(Get-EnvValue -EnvMap $EnvMap -Name 'APT_SECURITY_MIRROR' -Default '')"
    )
    & docker @backendArgs
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }

    if ($BuildFrontend) {
        $frontendArgs = @(
            'build',
            '-f', (Join-Path $repoRoot 'docker\build\Dockerfile.web.dev'),
            (Join-Path $repoRoot 'vue-vben-admin'),
            '-t', 'easymvp-web-dev:latest',
            '--build-arg', "WEB_BASE_IMAGE=$(Get-EnvValue -EnvMap $EnvMap -Name 'WEB_BASE_IMAGE' -Default 'node:22-bookworm')",
            '--build-arg', "NPM_REGISTRY=$(Get-EnvValue -EnvMap $EnvMap -Name 'NPM_REGISTRY' -Default 'https://registry.npmmirror.com')"
        )
        & docker @frontendArgs
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }
    }

    if ($BuildAiRuntime) {
        $aiRuntimeArgs = @(
            'build',
            '-f', (Join-Path $repoRoot 'docker\build\Dockerfile.openhands.runtime'),
            (Join-Path $repoRoot 'admin-go'),
            '-t', 'easymvp-openhands-local:latest',
            '--build-arg', "OPENHANDS_RUNTIME_BASE_IMAGE=$(Get-EnvValue -EnvMap $EnvMap -Name 'OPENHANDS_RUNTIME_BASE_IMAGE' -Default 'python:3.12-slim')"
        )
        & docker @aiRuntimeArgs
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }
    }
}

$dockerArgs = @(
    'compose',
    '--project-name', 'easymvp',
    '--env-file', $sourceEnv,
    '-f', $composeFile
)

Push-Location $repoRoot
try {
    $envMap = Get-EnvMap -Path $sourceEnv

    if ($ComposeArgs.Count -eq 0) {
        Invoke-DevBuild -EnvMap $envMap
        & docker @dockerArgs 'up' '-d'
        if ($LASTEXITCODE -eq 0) {
            Write-StartupGuide
        }
    } elseif ($ComposeArgs -contains 'up') {
        Invoke-DevBuild -EnvMap $envMap -BuildFrontend:(Test-ProfileEnabled -Args $ComposeArgs -ProfileName 'frontend') -BuildAiRuntime:(Test-ProfileEnabled -Args $ComposeArgs -ProfileName 'ai-runtime')
        & docker @dockerArgs @ComposeArgs
        if ($LASTEXITCODE -eq 0) {
            Write-StartupGuide
        }
    } else {
        & docker @dockerArgs @ComposeArgs
    }
} finally {
    Pop-Location
}
