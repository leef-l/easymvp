@echo off
setlocal enabledelayedexpansion

set "ROOT=%~dp0"
set "CORE_BIN=%ROOT%bin\easymvp-core.exe"
set "CORE_PID_FILE=%ROOT%bin\.core_pid"

cd /d "%ROOT%" || goto :fail

if /I "%~1"=="stop" goto :stop
if /I "%~1"=="build" goto :build_only
if /I "%~1"=="core" goto :core_only
if not "%~1"=="" goto :usage

goto :start

:start
echo == EasyMVP Local dev start ==

call :build_core || goto :fail
call :start_core || goto :fail
call :start_desktop || goto :fail
goto :success

:build_only
echo == EasyMVP Local build ==
call :build_core || goto :fail
echo.
echo Build complete.
echo Core binary: %CORE_BIN%
goto :success

:core_only
echo == EasyMVP Core only ==
call :build_core || goto :fail
call :start_core || goto :fail
echo.
echo Core is running at http://127.0.0.1:8000
echo Use 'dev_local.bat stop' to stop.
goto :success

:stop
echo Stopping EasyMVP core...
if exist "%CORE_PID_FILE%" (
    set /p CORE_PID=<"%CORE_PID_FILE%"
    if not "!CORE_PID!"=="" (
        taskkill /PID !CORE_PID! /F >nul 2>nul
    )
    del "%CORE_PID_FILE%" >nul 2>nul
)
taskkill /FI "WINDOWTITLE eq EasyMVP Core*" /F >nul 2>nul
echo Stopped.
goto :success

:build_core
echo Building core binary...
where go >nul 2>nul
if not "%errorlevel%"=="0" (
    echo Go was not found. Install Go 1.24+ and make sure go.exe is in PATH.
    exit /b 1
)
if not exist "%ROOT%bin" mkdir "%ROOT%bin" || exit /b 1
pushd "%ROOT%apps\core" || exit /b 1
go build -trimpath -ldflags "-s -w" -o "%CORE_BIN%" .\main.go
set "_BUILD_EXIT=%errorlevel%"
popd
if not "%_BUILD_EXIT%"=="0" (
    echo Core build failed.
    exit /b %_BUILD_EXIT%
)
echo Core build OK.
exit /b 0

:start_core
echo Starting core...
if exist "%CORE_PID_FILE%" (
    set /p OLD_PID=<"%CORE_PID_FILE%"
    if not "!OLD_PID!"=="" (
        taskkill /PID !OLD_PID! /F >nul 2>nul
    )
)
start "EasyMVP Core" /min cmd /c "cd /d "%ROOT%apps\core" && "%CORE_BIN%" && pause"
:: Give it a moment to start so we can capture the window PID indirectly via title
powershell -NoProfile -ExecutionPolicy Bypass -Command "Start-Sleep -Milliseconds 500; $p = Get-Process -Name 'easymvp-core' -ErrorAction SilentlyContinue | Select-Object -First 1; if ($p) { $p.Id | Out-File -Encoding ASCII '%CORE_PID_FILE%' }"
echo Core started at http://127.0.0.1:8000
exit /b 0

:start_desktop
echo Starting desktop dev...
cd /d "%ROOT%apps\desktop" || exit /b 1
echo.
echo Desktop dev starting. Press Ctrl+C to stop, then run 'dev_local.bat stop' to stop core.
echo.
npm run dev
exit /b %errorlevel%

:usage
echo Usage:
echo   dev_local.bat         Build and start core + desktop
echo   dev_local.bat build   Build core only
echo   dev_local.bat core    Build and start core only
echo   dev_local.bat stop    Stop background core process
goto :fail

:success
echo.
echo Done.
exit /b 0

:fail
echo.
echo Failed.
exit /b 1
