#!/bin/bash

echo "正在关闭后端服务..."

# 查找占用8080端口的进程ID
PID=$(lsof -i :8080 -t)

if [ -z "$PID" ]; then
    echo "未检测到运行中的后端服务"
    exit 0
fi

echo "检测到运行中的后端服务，进程ID：$PID"

echo "正在终止进程..."
# 终止进程
kill "$PID"

if [ $? -eq 0 ]; then
    echo "后端服务已成功关闭"
else
    echo "关闭后端服务失败，请手动终止进程"
    exit 1
fi