#!/bin/bash

echo "========================================"
echo "飞书记录助手 - 跨平台编译脚本"
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

# 安装依赖
echo "[3/3] 安装依赖并开始编译..."
echo "正在安装依赖..."
go mod tidy || { echo "错误: 依赖安装失败"; exit 1; }
echo "✓ 依赖安装完成"
echo ""

# 创建输出目录
mkdir -p ../dist

# 编译Linux版本
echo "正在编译 Linux 版本..."
GOOS=linux GOARCH=amd64 go build -o ../dist/lark-record-server-linux main.go
if [ $? -eq 0 ]; then
    echo "✓ Linux 版本编译成功"
else
    echo "✗ Linux 版本编译失败"
fi

# 编译macOS版本
echo "正在编译 macOS 版本..."
GOOS=darwin GOARCH=amd64 go build -o ../dist/lark-record-server-mac main.go
if [ $? -eq 0 ]; then
    echo "✓ macOS 版本编译成功"
else
    echo "✗ macOS 版本编译失败"
fi

# 编译Windows版本
echo "正在编译 Windows 版本..."
GOOS=windows GOARCH=amd64 go build -o ../dist/lark-record-server.exe main.go
if [ $? -eq 0 ]; then
    echo "✓ Windows 版本编译成功"
else
    echo "✗ Windows 版本编译失败"
fi

# 复制配置文件到dist目录
echo "正在复制配置文件到dist目录..."
cp backend/config.json ../dist/ 2>/dev/null || true

# 返回根目录
cd ..

# 打包前端插件
echo ""
echo "========================================"
echo "正在打包前端插件..."

# 创建插件打包临时目录
mkdir -p dist/chrome-extension

# 复制插件必要文件
cp manifest.json dist/chrome-extension/
cp -r background dist/chrome-extension/
cp -r popup dist/chrome-extension/
cp -r options dist/chrome-extension/
cp -r styles dist/chrome-extension/
cp -r utils dist/chrome-extension/
cp -r icons dist/chrome-extension/
cp backend/config.json dist/chrome-extension/ 2>/dev/null || true

# 打包成zip文件
cd dist

# 检查系统类型，使用不同的打包命令
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
    # Windows系统，使用PowerShell的Compress-Archive命令
    powershell -Command "Compress-Archive -Path chrome-extension -DestinationPath lark-record-chrome-extension.zip -Force"
else
    # Linux/macOS系统，使用zip命令
    zip -r lark-record-chrome-extension.zip chrome-extension/
fi

# 清理临时目录
rm -rf chrome-extension

if [ $? -eq 0 ]; then
    echo "✓ 前端插件打包成功"
else
    echo "✗ 前端插件打包失败"
fi

# 返回根目录
cd ..

echo ""
echo "========================================"
echo "打包完成！"
echo "文件已生成在 dist 目录中："
echo ""
echo "后端服务可执行文件："
echo "- Linux: lark-record-server-linux"
echo "- macOS: lark-record-server-mac"
echo "- Windows: lark-record-server.exe"
echo ""
echo "前端Chrome插件："
echo "- lark-record-chrome-extension.zip"
echo ""
echo "使用方法："
echo "1. 后端服务：直接运行对应平台的可执行文件即可启动服务"
echo "2. Chrome插件：解压zip文件，在Chrome扩展管理页面加载解压后的目录"
echo "3. 详细使用说明请参考 USER_GUIDE.md 文件"
echo "========================================"