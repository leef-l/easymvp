@echo off
chcp 65001 >nul

:: ─────────────────────────────────────────────────────────────
:: EasyMVP Windows 一键停止脚本
:: 终止所有 EasyMVP 后端服务进程
:: ─────────────────────────────────────────────────────────────

echo ============================================
echo   EasyMVP 服务停止
echo ============================================
echo.

tasklist /fi "imagename eq system.exe" 2>nul | find /i "system.exe" >nul
if %errorlevel% equ 0 (
    echo [停止] system.exe ...
    taskkill /f /im system.exe >nul 2>nul
) else (
    echo [信息] system.exe 未运行
)

tasklist /fi "imagename eq ai.exe" 2>nul | find /i "ai.exe" >nul
if %errorlevel% equ 0 (
    echo [停止] ai.exe ...
    taskkill /f /im ai.exe >nul 2>nul
) else (
    echo [信息] ai.exe 未运行
)

tasklist /fi "imagename eq mvp.exe" 2>nul | find /i "mvp.exe" >nul
if %errorlevel% equ 0 (
    echo [停止] mvp.exe ...
    taskkill /f /im mvp.exe >nul 2>nul
) else (
    echo [信息] mvp.exe 未运行
)

echo.
echo [完成] 所有服务已停止
echo.
pause
