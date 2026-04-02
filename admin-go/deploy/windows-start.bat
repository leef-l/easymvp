@echo off
chcp 65001 >nul
setlocal

:: ─────────────────────────────────────────────────────────────
:: EasyMVP Windows 一键启动脚本
:: 同时启动 system / ai / mvp 三个后端服务
:: ─────────────────────────────────────────────────────────────

set "BIN_DIR=%~dp0..\bin"

echo ============================================
echo   EasyMVP 服务启动
echo ============================================
echo.

:: 检查编译产物
if not exist "%BIN_DIR%\system.exe" (
    echo [错误] 未找到 system.exe，请先运行 windows-deploy.bat 编译
    pause
    exit /b 1
)
if not exist "%BIN_DIR%\ai.exe" (
    echo [错误] 未找到 ai.exe，请先运行 windows-deploy.bat 编译
    pause
    exit /b 1
)
if not exist "%BIN_DIR%\mvp.exe" (
    echo [错误] 未找到 mvp.exe，请先运行 windows-deploy.bat 编译
    pause
    exit /b 1
)

:: ─────────────────────────────────────────────
:: 启动服务（每个服务在独立窗口运行）
:: ─────────────────────────────────────────────
echo [启动] system 服务 (端口 9000)...
start "EasyMVP - System (9000)" cmd /k "cd /d %BIN_DIR%\system-config && %BIN_DIR%\system.exe"

echo [启动] ai 服务 (端口 9001)...
start "EasyMVP - AI (9001)" cmd /k "cd /d %BIN_DIR%\ai-config && %BIN_DIR%\ai.exe"

echo [启动] mvp 服务 (端口 9002)...
start "EasyMVP - MVP (9002)" cmd /k "cd /d %BIN_DIR%\mvp-config && %BIN_DIR%\mvp.exe"

echo.
echo [信息] 三个服务已在独立窗口启动
echo.
echo   system  http://localhost:9000  (Swagger: http://localhost:9000/swagger)
echo   ai      http://localhost:9001  (Swagger: http://localhost:9001/swagger)
echo   mvp     http://localhost:9002  (Swagger: http://localhost:9002/swagger)
echo.
echo [提示] 关闭各服务窗口即可停止对应服务
echo.
pause
