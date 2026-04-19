@echo off
setlocal enabledelayedexpansion

set "ROOT=%~dp0.."
set "LOGFILE=%~dp0local-verify-completion-verdict-authority.log"

break > "%LOGFILE%"
call :main
set "EXIT_CODE=%errorlevel%"
echo.
echo LogFile=%LOGFILE%
echo ExitCode=%EXIT_CODE%
pause
exit /b %EXIT_CODE%

:main
call :log == 1. Validate apps\core service tests ==
cd /d "%ROOT%\apps\core" || goto :fail
call :run go version || goto :fail
call :run go test ./internal/service/... || goto :fail

call :log
call :log == 2. Validate apps\desktop typecheck ==
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

call :run pnpm run typecheck || goto :fail

call :log
call :log == 3. Validate packaged desktop smoke ==
call :run pnpm run package:dir || goto :fail
call :run pnpm run verify:package || goto :fail

call :log
call :log == 4. Validation Passed ==
call :log apps\core internal service tests passed
call :log apps\desktop typecheck passed
call :log apps\desktop packaged smoke passed
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
