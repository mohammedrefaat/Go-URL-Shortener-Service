@echo off
setlocal enabledelayedexpansion

rem Define color codes for Windows console
set "GREEN=[92m"
set "YELLOW=[93m"
set "RED=[91m"
set "BLUE=[94m"
set "RESET=[0m"

rem Get ANSI escape character
for /f %%a in ('echo prompt $E^| cmd') do set "ESC=%%a"

rem Default settings
set "GOVULNCHECK_AVAILABLE=1"
set "USE_ZIP=0"
set "DO_COMPRESS=1"
set "EXE_NAME="

rem Check for help parameter first
for %%a in (%*) do (
    if /i "%%a"=="--help" (
        call :ShowHelp
        exit /b 0
    )
    if /i "%%a"=="-h" (
        call :ShowHelp
        exit /b 0
    )
)

rem Process parameters
:param_loop
if "%1"=="" goto end_param_loop
if /i "%1"=="--skip" (
    set "GOVULNCHECK_AVAILABLE=0"
    echo !ESC!%YELLOW%Detected --skip flag. Vulnerability check will be skipped.!ESC!%RESET%
)
if /i "%1"=="--zip" (
    set "USE_ZIP=1"
    echo !ESC!%YELLOW%Using ZIP format instead of RAR.!ESC!%RESET%
)
if /i "%1"=="--no-compress" (
    set "DO_COMPRESS=0"
    echo !ESC!%YELLOW%Compression disabled. Only copying executable.!ESC!%RESET%
)
if /i "%1"=="--name" (
    shift
    call set "TEMP_NAME=%%~1"
    if defined TEMP_NAME (
        set "EXE_NAME=!TEMP_NAME!.exe"
        echo !ESC!%BLUE%Using custom executable name: !TEMP_NAME!.exe!ESC!%RESET%
        shift
    ) else (
        echo !ESC!%RED%Error: --name requires a value!ESC!%RESET%
        exit /b 1
    )
)
shift
goto param_loop
:end_param_loop

echo !ESC!%BLUE%Build process started!ESC!%RESET%

rem Check for WinRAR only if we're using RAR compression
if "!DO_COMPRESS!"=="1" if "!USE_ZIP!"=="0" (
    echo !ESC!%BLUE%Checking if WinRAR is installed...!ESC!%RESET%
    set "WINRAR_PATH=C:\Program Files\WinRAR\rar.exe"
    if not exist "!WINRAR_PATH!" (
        echo !ESC!%RED%WinRAR is not installed at the default location: '!WINRAR_PATH!'. Please install WinRAR or use --zip option.!ESC!%RESET%
        exit /b 1
    )
)

if "!GOVULNCHECK_AVAILABLE!"=="1" (
    echo !ESC!%BLUE%Starting vulnerability check process...!ESC!%RESET%
    echo !ESC!%GREEN%Installing or updating govulncheck...!ESC!%RESET%
    go install golang.org/x/vuln/cmd/govulncheck@latest

    echo !ESC!%GREEN%Running govulncheck to scan for vulnerabilities...!ESC!%RESET%
    for /f "tokens=*" %%i in ('govulncheck ./...') do set "GOVULNCHECK_OUTPUT=%%i"

    echo !GOVULNCHECK_OUTPUT! | findstr /i /c:"No vulnerabilities found." >nul
    if errorlevel 1 (
        echo !ESC!%RED%Vulnerabilities found. Stopping the build process.!ESC!%RESET%
        exit /b 1
    )
    echo !ESC!%GREEN%No vulnerabilities found. Proceeding with the build...!ESC!%RESET%
) else (
    echo !ESC!%YELLOW%Skipping vulnerability check...!ESC!%RESET%
)

rem Get the current directory name if no custom name provided
if "!EXE_NAME!"=="" (
    for %%F in ("%cd%") do set "EXE_NAME=%%~nxF.exe"
    echo !ESC!%BLUE%Using default executable name: !EXE_NAME!!ESC!%RESET%
)

echo !ESC!%BLUE%Running go mod tidy...!ESC!%RESET%
go mod tidy

echo !ESC!%BLUE%Building application...!ESC!%RESET%
go build -ldflags "-s -w" -o "!EXE_NAME!"

if exist "!EXE_NAME!" (
    echo !ESC!%GREEN%Build successful!!ESC!%RESET%
    
    if "!DO_COMPRESS!"=="1" (
        set "ZIP_FILE=!EXE_NAME:.exe=.zip!"
        set "RAR_FILE=!EXE_NAME:.exe=.rar!"

        if exist "!ZIP_FILE!" (
            del "!ZIP_FILE!"
            echo !ESC!%YELLOW%Deleted old ZIP archive: '!ZIP_FILE!'!ESC!%RESET%
        )
        if exist "!RAR_FILE!" (
            del "!RAR_FILE!"
            echo !ESC!%YELLOW%Deleted old RAR archive: '!RAR_FILE!'!ESC!%RESET%
        )

        if "!USE_ZIP!"=="1" (
            echo !ESC!%BLUE%Creating ZIP archive...!ESC!%RESET%
            powershell -command "Compress-Archive -Path '!EXE_NAME!' -DestinationPath '!ZIP_FILE!' -Force"
            echo !ESC!%GREEN%Created ZIP archive: '!ZIP_FILE!'!ESC!%RESET%
        ) else (
            echo !ESC!%BLUE%Creating RAR archive...!ESC!%RESET%
            "!WINRAR_PATH!" a "!RAR_FILE!" "!EXE_NAME!"
            echo !ESC!%GREEN%Created RAR archive: '!RAR_FILE!'!ESC!%RESET%
        )
    ) else (
        echo !ESC!%BLUE%Skipping compression as requested!ESC!%RESET%
    )
) else (
    echo !ESC!%RED%Build failed. Executable not found.!ESC!%RESET%
)

exit /b 0

:ShowHelp
echo !ESC!%BLUE%Build Script Help!ESC!%RESET%
echo.
echo Usage: build.bat [options]
echo.
echo Options:
echo   --help, -h        Show this help message
echo   --skip            Skip vulnerability check
echo   --zip             Create ZIP archive instead of RAR
echo   --no-compress     Skip compression (no ZIP/RAR)
echo   --name [name]     Set custom executable name (without .exe)
echo.
echo Examples:
echo   build.bat                    Regular build with vulnerability check (RAR)
echo   build.bat --skip             Build without vulnerability check (RAR)
echo   build.bat --zip              Build with ZIP archive instead of RAR
echo   build.bat --name myapp       Build with custom name myapp.exe
echo   build.bat --no-compress      Build without creating archive
echo   build.bat --skip --zip       Build without check and use ZIP
echo.
echo Color Legend:
echo !ESC!%GREEN%Green!ESC!%RESET%  - Success messages
echo !ESC!%YELLOW%Yellow!ESC!%RESET% - Warnings and notifications
echo !ESC!%RED%Red!ESC!%RESET%    - Errors and failures
echo !ESC!%BLUE%Blue!ESC!%RESET%   - Log messages
echo.
exit /b 0
endlocal