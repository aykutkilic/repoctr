@echo off
setlocal enabledelayedexpansion

:: Build script for repo-ctr

set BINARY=repo-ctr.exe

if "%VERSION%"=="" set VERSION=dev

if "%1"=="" goto build
if "%1"=="build" goto build
if "%1"=="test" goto test
if "%1"=="install" goto install
if "%1"=="clean" goto clean
if "%1"=="fmt" goto fmt
if "%1"=="lint" goto lint
if "%1"=="coverage" goto coverage
if "%1"=="help" goto help
if "%1"=="--help" goto help
if "%1"=="-h" goto help

echo Error: Unknown command: %1
goto help

:build
echo ==^> Building %BINARY%...
go build -ldflags "-s -w -X main.Version=%VERSION%" -o "%BINARY%" ./cmd/repo-ctr
if errorlevel 1 goto error
echo Built: %BINARY%
goto end

:test
echo ==^> Running tests...
go test -v ./...
if errorlevel 1 goto error
goto end

:install
echo ==^> Installing %BINARY%...
go install -ldflags "-s -w -X main.Version=%VERSION%" ./cmd/repo-ctr
if errorlevel 1 goto error
for /f "tokens=*" %%i in ('go env GOPATH') do set GOPATH=%%i
echo Installed to: %GOPATH%\bin\%BINARY%
goto end

:clean
echo ==^> Cleaning build artifacts...
if exist "%BINARY%" del /q "%BINARY%"
if exist "coverage.out" del /q "coverage.out"
if exist "coverage.html" del /q "coverage.html"
echo Done.
goto end

:fmt
echo ==^> Formatting code...
go fmt ./...
if errorlevel 1 goto error
echo Done.
goto end

:lint
echo ==^> Running linter...
where golangci-lint >nul 2>nul
if errorlevel 1 (
    echo Warning: golangci-lint not installed. Install with:
    echo   go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest
    goto error
)
golangci-lint run
if errorlevel 1 goto error
goto end

:coverage
echo ==^> Generating coverage report...
go test -coverprofile=coverage.out ./...
if errorlevel 1 goto error
go tool cover -html=coverage.out -o coverage.html
if errorlevel 1 goto error
echo Coverage report: coverage.html
goto end

:help
echo Usage: %~nx0 [command]
echo.
echo Commands:
echo   build       Build the binary (default)
echo   test        Run tests
echo   install     Install to %%GOPATH%%\bin
echo   clean       Remove build artifacts
echo   fmt         Format code
echo   lint        Run linter
echo   coverage    Generate coverage report
echo   help        Show this help message
goto end

:error
echo.
echo Build failed!
exit /b 1

:end
endlocal
