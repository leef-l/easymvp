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

Push-Location $repoRoot
try {
    $envMap = @{}
    foreach ($line in Get-Content -Path $sourceEnv) {
        $trimmed = $line.Trim()
        if ([string]::IsNullOrWhiteSpace($trimmed) -or $trimmed.StartsWith('#')) {
            continue
        }
        $parts = $trimmed -split '=', 2
        if ($parts.Count -eq 2) {
            $envMap[$parts[0]] = $parts[1]
        }
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

    $buildFrontend = $true
    $buildAiRuntime = $false

    for ($i = 0; $i -lt $ComposeArgs.Count; $i++) {
        if ($ComposeArgs[$i] -eq '--profile' -and $i + 1 -lt $ComposeArgs.Count) {
            if ($ComposeArgs[$i + 1] -eq 'ai-runtime') {
                $buildAiRuntime = $true
            }
            if ($ComposeArgs[$i + 1] -eq 'frontend') {
                $buildFrontend = $true
            }
        }
    }

    $backendArgs = @(
        'build',
        '-f', (Join-Path $repoRoot 'docker\build\Dockerfile.admin-go.dev'),
        (Join-Path $repoRoot 'admin-go'),
        '-t', 'easymvp-admin-go-dev:latest',
        '--build-arg', "GO_PROXY=$(Get-EnvValue -EnvMap $envMap -Name 'GO_PROXY' -Default 'https://goproxy.cn,direct')",
        '--build-arg', "PIP_INDEX_URL=$(Get-EnvValue -EnvMap $envMap -Name 'PIP_INDEX_URL' -Default 'https://pypi.org/simple')",
        '--build-arg', "APT_MIRROR=$(Get-EnvValue -EnvMap $envMap -Name 'APT_MIRROR' -Default '')",
        '--build-arg', "APT_SECURITY_MIRROR=$(Get-EnvValue -EnvMap $envMap -Name 'APT_SECURITY_MIRROR' -Default '')"
    )
    & docker @backendArgs
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }

    if ($buildFrontend) {
        $frontendArgs = @(
            'build',
            '-f', (Join-Path $repoRoot 'docker\build\Dockerfile.web.dev'),
            (Join-Path $repoRoot 'vue-vben-admin'),
            '-t', 'easymvp-web-dev:latest',
            '--build-arg', "NPM_REGISTRY=$(Get-EnvValue -EnvMap $envMap -Name 'NPM_REGISTRY' -Default 'https://registry.npmmirror.com')"
        )
        & docker @frontendArgs
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }
    }

    if ($buildAiRuntime) {
        $aiRuntimeArgs = @(
            'build',
            '-f', (Join-Path $repoRoot 'docker\build\Dockerfile.openhands.runtime'),
            (Join-Path $repoRoot 'admin-go'),
            '-t', 'easymvp-openhands-local:latest'
        )
        & docker @aiRuntimeArgs
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }
    }
} finally {
    Pop-Location
}
