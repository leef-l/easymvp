@echo off
setlocal

cd /d "%~dp0"

echo.
echo [EasyMVP] Pushing branch v3 to origin...
echo.

git push origin v3
set "EXIT_CODE=%errorlevel%"

echo.
if "%EXIT_CODE%"=="0" (
  echo [EasyMVP] Push succeeded.
) else (
  echo [EasyMVP] Push failed. ExitCode=%EXIT_CODE%
)

pause
exit /b %EXIT_CODE%
