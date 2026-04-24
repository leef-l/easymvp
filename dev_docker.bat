@echo off
setlocal enabledelayedexpansion

set "ROOT=%~dp0"
set "COMPOSE_FILE=%ROOT%docker-compose.dev.yml"
set "CORE_URL=http://127.0.0.1:8000/api/v3/system/healthz"

cd /d "%ROOT%" || goto :fail

call :resolve_compose || goto :fail

if /I "%~1"=="down" goto :down
if /I "%~1"=="stop" goto :down
if /I "%~1"=="clean" goto :clean
if /I "%~1"=="logs" goto :logs
if /I "%~1"=="ps" goto :ps
if /I "%~1"=="restart" goto :restart
if not "%~1"=="" goto :usage

goto :start

:start
echo == EasyMVP Docker dev start ==
echo Compose=%COMPOSE%
echo Config=docker\dev\config.yaml
call :build_core || goto :fail
%COMPOSE% -f "%COMPOSE_FILE%" up --build -d || goto :fail
%COMPOSE% -f "%COMPOSE_FILE%" ps
call :wait_health || goto :health_fail
echo.
echo EasyMVP core is ready: http://127.0.0.1:8000
echo Health: %CORE_URL%
echo.
echo Commands:
echo   dev_docker.bat logs
echo   dev_docker.bat down
echo   dev_docker.bat clean
goto :success

:restart
%COMPOSE% -f "%COMPOSE_FILE%" down --remove-orphans || goto :fail
goto :start

:down
%COMPOSE% -f "%COMPOSE_FILE%" down --remove-orphans || goto :fail
goto :success

:clean
%COMPOSE% -f "%COMPOSE_FILE%" down -v --remove-orphans || goto :fail
if exist "%ROOT%docker\dev\bin" rmdir /s /q "%ROOT%docker\dev\bin"
goto :success

:logs
%COMPOSE% -f "%COMPOSE_FILE%" logs -f easymvp-core
exit /b %errorlevel%

:ps
%COMPOSE% -f "%COMPOSE_FILE%" ps
exit /b %errorlevel%

:health_fail
echo.
echo Core health check failed. Recent logs:
%COMPOSE% -f "%COMPOSE_FILE%" logs --tail=120 easymvp-core
goto :fail

:wait_health
echo Waiting for core health...
powershell -NoProfile -ExecutionPolicy Bypass -Command "$ErrorActionPreference='SilentlyContinue'; for ($i = 1; $i -le 60; $i++) { try { $r = Invoke-WebRequest -UseBasicParsing '%CORE_URL%' -TimeoutSec 2; if ($r.StatusCode -ge 200 -and $r.StatusCode -lt 300) { Write-Host 'Core health OK'; exit 0 } } catch { } Start-Sleep -Seconds 1 }; exit 1"
exit /b %errorlevel%

:resolve_compose
docker compose version >nul 2>nul
if "%errorlevel%"=="0" (
  set "COMPOSE=docker compose"
  exit /b 0
)
docker-compose version >nul 2>nul
if "%errorlevel%"=="0" (
  set "COMPOSE=docker-compose"
  exit /b 0
)
echo Docker Compose was not found. Install Docker Desktop or docker-compose.
exit /b 1

:build_core
echo Building Linux core binary for Docker...
where go >nul 2>nul
if not "%errorlevel%"=="0" (
  echo Go was not found. Install Go 1.24+ and make sure go.exe is in PATH.
  exit /b 1
)
if not exist "%ROOT%docker\dev\bin" mkdir "%ROOT%docker\dev\bin" || exit /b 1
pushd "%ROOT%apps\core" || exit /b 1
set "CGO_ENABLED=0"
set "GOOS=linux"
set "GOARCH=amd64"
go build -trimpath -ldflags "-s -w" -o "%ROOT%docker\dev\bin\easymvp-core-linux-amd64" .\main.go
set "_BUILD_EXIT=%errorlevel%"
popd
if not "%_BUILD_EXIT%"=="0" (
  echo Core Linux binary build failed.
  exit /b %_BUILD_EXIT%
)
exit /b 0

:usage
echo Usage:
echo   dev_docker.bat
echo   dev_docker.bat restart
echo   dev_docker.bat logs
echo   dev_docker.bat ps
echo   dev_docker.bat down
echo   dev_docker.bat clean
goto :fail

:success
echo.
echo Done.
pause
exit /b 0

:fail
echo.
echo Failed.
pause
exit /b 1
