@echo off
setlocal enabledelayedexpansion

rem Windows local validation entry for the current apps/core + apps/desktop codepath.
rem Run this on your local high-spec machine after git pull.

set "ROOT=%~dp0.."
set "LOGFILE=%~dp0local-verify-apps-core-desktop.log"

break > "%LOGFILE%"
call :main
set "EXIT_CODE=%errorlevel%"
echo.
echo LogFile=%LOGFILE%
echo ExitCode=%EXIT_CODE%
pause
exit /b %EXIT_CODE%

:main

call :log == 1. Update Repository ==
cd /d "%ROOT%" || goto :fail
call :run git pull || goto :fail
call :run git status --short || goto :fail

call :log
call :log == 2. Validate apps\core ==
cd /d "%ROOT%\apps\core" || goto :fail
call :run go version || goto :fail
call :run go test ./... || goto :fail

call :log
call :log == 3. Validate apps\desktop ==
cd /d "%ROOT%\apps\desktop" || goto :fail
call :run node -v || goto :fail
call :run pnpm -v || goto :fail

if exist pnpm-lock.yaml (
  call :log pnpm-lock.yaml detected, running pnpm install --frozen-lockfile
  call :run pnpm install --frozen-lockfile || goto :fail
) else (
  call :log pnpm-lock.yaml not found, running pnpm install
  call :run pnpm install || goto :fail
)

call :run pnpm run build || goto :fail

call :log
call :log == 4. Validation Passed ==
call :log apps\core go test passed
call :log apps\desktop build passed
call :log
call :log Optional smoke test:
call :log   cd /d "%ROOT%\apps\desktop" ^&^& pnpm run dev
exit /b 0

:fail
call :log
call :log == Validation Failed ==
exit /b 1

:log
if "%~1"=="" (
  echo(
  >> "%LOGFILE%" echo(
  exit /b 0
)
set "_VERIFY_LINE=%*"
echo %_VERIFY_LINE%
>> "%LOGFILE%" echo %_VERIFY_LINE%
set "_VERIFY_LINE="
exit /b 0

:run
set "_VERIFY_CMD=%*"
call :log ^> %_VERIFY_CMD%
>> "%LOGFILE%" echo [command] %_VERIFY_CMD%
cmd /d /s /c "%_VERIFY_CMD%"
set "_VERIFY_EXIT=%errorlevel%"
>> "%LOGFILE%" echo [exit %_VERIFY_EXIT%] %_VERIFY_CMD%
set "_VERIFY_CMD="
exit /b %_VERIFY_EXIT%
