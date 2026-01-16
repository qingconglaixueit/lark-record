#!/bin/bash

echo "========================================"
echo "飞书记录助手 - 后端服务启动脚本"
echo "========================================"
echo ""

# 检查Go环境
echo "[1/3] 检查Go环境..."
if ! command -v go &> /dev/null; then
    echo "错误: 未检测到Go环境，请先安装Go 1.21或更高版本"
    echo "下载地址: https://golang.org/dl/"
    exit 1
fi
echo "✓ Go环境检查通过"
echo ""

# 进入后端目录
echo "[2/3] 进入后端目录..."
cd backend || { echo "错误: 无法进入backend目录"; exit 1; }
echo "✓ 进入后端目录成功"
echo ""

# 安装依赖并启动服务
echo "[3/3] 安装依赖并启动服务..."
echo "正在安装依赖..."
go mod tidy || { echo "错误: 依赖安装失败"; exit 1; }
echo "✓ 依赖安装完成"
echo ""

echo "========================================"
echo "启动后端服务..."
echo "服务地址: http://localhost:8080"
echo "按 Ctrl+C 停止服务"
echo "========================================"
echo ""

go run main.go