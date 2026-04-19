@echo off
setlocal enabledelayedexpansion

rem Windows local validation entry for the current apps/core + apps/desktop codepath.
rem Run this on your local high-spec machine after git pull.

set "ROOT=%~dp0.."

echo == 1. Update Repository ==
cd /d "%ROOT%" || goto :fail
git pull || goto :fail
git status --short || goto :fail

echo.
echo == 2. Validate apps\core ==
cd /d "%ROOT%\apps\core" || goto :fail
go version || goto :fail
go test ./... || goto :fail

echo.
echo == 3. Validate apps\desktop ==
cd /d "%ROOT%\apps\desktop" || goto :fail
node -v || goto :fail
npm -v || goto :fail

if exist package-lock.json (
  echo package-lock.json detected, running npm ci
  npm ci || goto :fail
) else (
  echo package-lock.json not found, running npm install
  npm install || goto :fail
)

npm run build || goto :fail

echo.
echo == 4. Validation Passed ==
echo apps\core go test passed
echo apps\desktop build passed
echo.
echo Optional smoke test:
echo   cd /d "%ROOT%\apps\desktop" ^&^& npm run dev
echo.
echo ExitCode=0
pause
exit /b 0

:fail
echo.
echo == Validation Failed ==
echo ExitCode=1
pause
exit /b 1
