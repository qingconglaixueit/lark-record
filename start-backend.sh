#!/bin/bash

echo "========================================"
echo "飞书记录助手 - 后端服务启动脚本"
echo "========================================"
echo ""

# 检测操作系统
echo "[1/3] 检测操作系统..."
OS="$(uname -s)"
case "$OS" in
    Linux*)     EXECUTABLE="lark-record-server-linux";;
    Darwin*)    EXECUTABLE="lark-record-server-mac";;
    *)          echo "错误: 不支持的操作系统: $OS";
                echo "支持的操作系统: Linux, macOS";
                exit 1;;
esac
echo "✓ 检测到操作系统: $OS"
echo "✓ 将使用可执行文件: $EXECUTABLE"
echo ""

# 检查可执行文件是否存在
echo "[2/3] 检查可执行文件..."
if [ ! -f "dist/$EXECUTABLE" ]; then
    echo "错误: 可执行文件不存在: dist/$EXECUTABLE"
    echo "请先运行 build.sh 脚本编译程序"
    exit 1
fi

echo "✓ 可执行文件存在: dist/$EXECUTABLE"
echo ""

# 启动服务
echo "[3/3] 启动后端服务..."
cd dist

# 在后台运行服务
nohup ./$EXECUTABLE > ../server.log 2>&1 &

# 保存进程ID
PID=$!

# 检查服务是否成功启动
sleep 2
if ps -p $PID > /dev/null; then
    echo "✓ 后端服务已在后台启动"
    echo "✓ 进程ID: $PID"
    echo "✓ 服务地址: http://localhost:8080"
    echo "✓ 日志文件: server.log"
    echo ""
    echo "停止服务命令: kill $PID"
else
    echo "✗ 后端服务启动失败"
    echo "请查看 server.log 文件获取详细错误信息"
    exit 1
fi

echo ""
echo "========================================"
echo "服务启动完成！"
echo "========================================"