@echo off
setlocal

REM Try different Python commands
python3 %* 2>nul
if %errorlevel% == 0 goto :eof

python %* 2>nul
if %errorlevel% == 0 goto :eof

py -3 %* 2>nul
if %errorlevel% == 0 goto :eof

py %* 2>nul
if %errorlevel% == 0 goto :eof

echo Python not found. Please install Python or add it to PATH.
exit /b 1
