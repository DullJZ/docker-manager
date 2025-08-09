#!/bin/bash

# 检查是否安装了 UPX
if ! command -v upx &> /dev/null; then
    echo "警告: UPX 未安装，将跳过压缩步骤"
    echo "请安装 UPX: brew install upx (macOS) 或 apt install upx (Ubuntu)"
    USE_UPX=false
else
    USE_UPX=true
    echo "使用 UPX 压缩二进制文件..."
fi

# 创建 build 目录
mkdir -p build

# Windows AMD64
echo "构建 Windows AMD64..."
go env -w CGO_ENABLED=0 GOOS=windows GOARCH=amd64
go build -o ./build/docker-manager_windows_amd64.exe api.go funcs.go
if [ "$USE_UPX" = true ]; then
    upx --best ./build/docker-manager_windows_amd64.exe
fi

# Windows 386
echo "构建 Windows 386..."
go env -w CGO_ENABLED=0 GOOS=windows GOARCH=386
go build -o ./build/docker-manager_windows_386.exe api.go funcs.go
if [ "$USE_UPX" = true ]; then
    upx --best ./build/docker-manager_windows_386.exe
fi

# Linux AMD64
echo "构建 Linux AMD64..."
go env -w CGO_ENABLED=0 GOOS=linux GOARCH=amd64
go build -o ./build/docker-manager_linux_amd64 api.go funcs.go
if [ "$USE_UPX" = true ]; then
    upx --best ./build/docker-manager_linux_amd64
fi

# Linux 386
echo "构建 Linux 386..."
go env -w CGO_ENABLED=0 GOOS=linux GOARCH=386
go build -o ./build/docker-manager_linux_386 api.go funcs.go
if [ "$USE_UPX" = true ]; then
    upx --best ./build/docker-manager_linux_386
fi

# Linux ARM64
echo "构建 Linux ARM64..."
go env -w CGO_ENABLED=0 GOOS=linux GOARCH=arm64
go build -o ./build/docker-manager_linux_arm64 api.go funcs.go
if [ "$USE_UPX" = true ]; then
    upx --best ./build/docker-manager_linux_arm64
fi

# Darwin AMD64
echo "构建 Darwin AMD64..."
go env -w CGO_ENABLED=0 GOOS=darwin GOARCH=amd64
go build -o ./build/docker-manager_darwin_amd64 api.go funcs.go
if [ "$USE_UPX" = true ]; then
    upx --best ./build/docker-manager_darwin_amd64
fi

# Darwin ARM64
echo "构建 Darwin ARM64..."
go env -w CGO_ENABLED=0 GOOS=darwin GOARCH=arm64
go build -o ./build/docker-manager_darwin_arm64 api.go funcs.go
if [ "$USE_UPX" = true ]; then
    upx --best ./build/docker-manager_darwin_arm64
fi

echo "构建完成！"
if [ "$USE_UPX" = true ]; then
    echo "所有二进制文件已使用 UPX 压缩"
else
    echo "二进制文件未压缩（UPX 未安装）"
fi
