@echo off
setlocal enabledelayedexpansion

set "MOD_NAME=intermark"
set "VERSION_VAR_PATH=%MOD_NAME%/internal/app.Version"

set "PROJECT_ROOT=%~dp0"
set "ENTRY_POINT=%PROJECT_ROOT%cmd\%MOD_NAME%"
set "BIN_DIR=%PROJECT_ROOT%bin"

REM Clean binary directory
if exist "%BIN_DIR%" (
    rmdir /s /q "%BIN_DIR%"
    if errorlevel 1 (
        echo Failed to remove existing bin directory.
        exit /b 1
    )
)
mkdir "%BIN_DIR%"
if errorlevel 1 (
    echo Failed to create bin directory.
    exit /b 1
)
echo Cleaned binary directory.

REM Set the version if it's not set
if "%VERSION%"=="" (
    set "VERSION=vX.X.X"
)
echo Output version set to %VERSION%

call :build windows amd64
if errorlevel 1 (
    echo Build process failed.
    exit /b 1
)

goto :EOF

REM Function to build a binary
REM Usage: call :build <GOOS> <GOARCH>
:build
if "%~1"=="" (
    echo GOOS not specified.
    exit /b 1
)
if "%~2"=="" (
    echo GOARCH not specified.
    exit /b 1
)

set "GOOS=%~1"
set "GOARCH=%~2"
set "OUTPUT_PATH=%BIN_DIR%\%MOD_NAME%-%GOOS%-%GOARCH%"
if /I "%GOOS%"=="windows" (
    set "OUTPUT_PATH=%OUTPUT_PATH%.exe"
)

go build -ldflags="-X \"%VERSION_VAR_PATH%=%VERSION%\"" -o "%OUTPUT_PATH%" "%ENTRY_POINT%"
if errorlevel 1 (
    echo Go build failed for %GOOS%/%GOARCH%.
    exit /b 1
)
echo Built %MOD_NAME% for %GOOS%/%GOARCH%.

exit /b 0
