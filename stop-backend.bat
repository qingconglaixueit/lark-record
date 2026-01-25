@echo off
echo 正在关闭后端服务...

rem 查找占用8080端口的进程ID
for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8080') do set PID=%%a

if "%PID%"=="" (
    echo 未检测到运行中的后端服务
    pause
    exit /b 0
)

echo 检测到运行中的后端服务，进程ID：%PID%

echo 正在终止进程...
rem 终止进程
taskkill /PID %PID% /F

if %errorlevel% equ 0 (
    echo 后端服务已成功关闭
) else (
    echo 关闭后端服务失败，请手动在任务管理器中终止进程
    pause
    exit /b 1
)

echo 按任意键退出...
pause >nul