@echo off
chcp 65001 > nul
echo ========================================
echo 飞书记录助手 - 后端服务启动脚本
echo ========================================
echo.

echo [1/3] 检测操作系统...
echo ✓ 检测到操作系统: Windows
echo ✓ 将使用可执行文件: lark-record-server.exe
echo.

echo [2/3] 检查可执行文件...
if not exist "dist\lark-record-server.exe" (
    echo 错误: 可执行文件不存在: dist\lark-record-server.exe
    echo 请先运行 build.sh 脚本编译程序（需要在WSL或Git Bash环境中运行）
    pause
    exit /b 1
)

echo ✓ 可执行文件存在: dist\lark-record-server.exe
echo.

echo [3/3] 启动后端服务...

REM 在后台启动服务，使用start命令
start "飞书记录助手后端服务" /min "dist\lark-record-server.exe"

REM 检查服务是否成功启动
TIMEOUT /T 3 /NOBREAK > nul

REM 检查是否有进程在运行
for /f "tokens=2 delims=," %%a in ('tasklist /fi "IMAGENAME eq lark-record-server.exe" /fo csv /nh') do (
    set PID=%%a
)

if defined PID (
    echo ✓ 后端服务已在后台启动
    echo ✓ 进程ID: %PID%
    echo ✓ 服务地址: http://localhost:8080
    echo ✓ 日志文件: server-%date:~0,4%-%date:~5,2%-%date:~8,2%.log
    echo.
    echo 停止服务方法: 
    echo 1. 在任务管理器中结束 lark-record-server.exe 进程
    echo 2. 或运行命令: taskkill /F /IM lark-record-server.exe
) else (
    echo ✗ 后端服务启动失败
    echo 请查看 server.log 文件获取详细错误信息
    pause
    exit /b 1
)

echo.
echo ========================================
echo 服务启动完成！
echo ========================================
echo 按任意键关闭窗口...
pause > nul