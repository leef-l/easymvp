@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

:: ─────────────────────────────────────────────────────────────
:: EasyMVP Windows 本地部署脚本
:: 编译并启动所有后端服务（system / ai / mvp）
:: ─────────────────────────────────────────────────────────────

set "PROJECT_DIR=%~dp0.."
set "OUTPUT_DIR=%PROJECT_DIR%\bin"
set "GOPROXY=https://goproxy.cn,direct"

echo ============================================
echo   EasyMVP Windows 本地部署
echo ============================================
echo.

:: ─────────────────────────────────────────────
:: 检查 Go 环境
:: ─────────────────────────────────────────────
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [错误] 未检测到 Go 环境，请先安装 Go 1.25+
    echo 下载地址: https://go.dev/dl/
    pause
    exit /b 1
)

for /f "tokens=3" %%v in ('go version') do set "GO_VERSION=%%v"
echo [信息] Go 版本: %GO_VERSION%
echo [信息] 项目目录: %PROJECT_DIR%
echo [信息] 输出目录: %OUTPUT_DIR%
echo.

:: 创建输出目录
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

:: ─────────────────────────────────────────────
:: 检查 MySQL 连接（可选）
:: ─────────────────────────────────────────────
where mysql >nul 2>nul
if %errorlevel% equ 0 (
    echo [信息] 检查 MySQL 连接...
    mysql -u easymvp -pJKcHFJYXnkrB6BXE -h 127.0.0.1 -P 3306 -e "SELECT 1" easymvp >nul 2>nul
    if !errorlevel! neq 0 (
        echo [警告] MySQL 连接失败，请确认：
        echo         - MySQL 已启动
        echo         - 数据库 easymvp 已创建
        echo         - 用户名/密码正确
        echo.
        set /p "CONTINUE=是否继续编译？(y/n): "
        if /i "!CONTINUE!" neq "y" exit /b 1
    ) else (
        echo [信息] MySQL 连接正常
    )
) else (
    echo [信息] 未检测到 mysql 客户端，跳过连接检查
)
echo.

:: ─────────────────────────────────────────────
:: 下载依赖
:: ─────────────────────────────────────────────
echo [步骤 1/4] 下载 Go 依赖...
cd /d "%PROJECT_DIR%"
go mod download
if %errorlevel% neq 0 (
    echo [错误] 依赖下载失败
    pause
    exit /b 1
)
echo [完成] 依赖下载成功
echo.

:: ─────────────────────────────────────────────
:: 编译三个服务
:: ─────────────────────────────────────────────
echo [步骤 2/4] 编译 system 服务 (端口 9000)...
go build -ldflags "-s -w" -o "%OUTPUT_DIR%\system.exe" ./app/system/main.go
if %errorlevel% neq 0 (
    echo [错误] system 编译失败
    pause
    exit /b 1
)
echo [完成] system.exe

echo [步骤 3/4] 编译 ai 服务 (端口 9001)...
go build -ldflags "-s -w" -o "%OUTPUT_DIR%\ai.exe" ./app/ai/main.go
if %errorlevel% neq 0 (
    echo [错误] ai 编译失败
    pause
    exit /b 1
)
echo [完成] ai.exe

echo [步骤 4/4] 编译 mvp 服务 (端口 9002)...
go build -ldflags "-s -w" -o "%OUTPUT_DIR%\mvp.exe" ./app/mvp/main.go
if %errorlevel% neq 0 (
    echo [错误] mvp 编译失败
    pause
    exit /b 1
)
echo [完成] mvp.exe
echo.

:: ─────────────────────────────────────────────
:: 复制配置文件
:: ─────────────────────────────────────────────
echo [信息] 复制配置文件...

if not exist "%OUTPUT_DIR%\system-config" mkdir "%OUTPUT_DIR%\system-config"
if not exist "%OUTPUT_DIR%\ai-config" mkdir "%OUTPUT_DIR%\ai-config"
if not exist "%OUTPUT_DIR%\mvp-config" mkdir "%OUTPUT_DIR%\mvp-config"

xcopy /Y /E /I "%PROJECT_DIR%\app\system\manifest" "%OUTPUT_DIR%\system-config\manifest" >nul 2>nul
xcopy /Y /E /I "%PROJECT_DIR%\app\ai\manifest" "%OUTPUT_DIR%\ai-config\manifest" >nul 2>nul
xcopy /Y /E /I "%PROJECT_DIR%\app\mvp\manifest" "%OUTPUT_DIR%\mvp-config\manifest" >nul 2>nul

echo [完成] 配置文件已复制
echo.

:: ─────────────────────────────────────────────
:: 完成
:: ─────────────────────────────────────────────
echo ============================================
echo   编译完成！
echo ============================================
echo.
echo   输出目录: %OUTPUT_DIR%
echo.
echo   服务列表:
echo     system.exe  - 系统管理服务  端口 9000
echo     ai.exe      - AI 服务       端口 9001
echo     mvp.exe     - MVP 服务      端口 9002
echo.
echo   启动方式:
echo     方式1: 运行 windows-start.bat 一键启动全部服务
echo     方式2: 分别在各目录启动:
echo       cd %OUTPUT_DIR%\system-config ^&^& ..\system.exe
echo       cd %OUTPUT_DIR%\ai-config ^&^& ..\ai.exe
echo       cd %OUTPUT_DIR%\mvp-config ^&^& ..\mvp.exe
echo.
echo   前端开发服务:
echo     cd vue-vben-admin ^&^& pnpm dev
echo     访问 http://localhost:5666
echo.
pause
