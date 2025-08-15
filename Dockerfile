FROM alpine:latest AS builder

# 安装构建依赖
RUN apk add --no-cache \
    build-base \
    go \
    git

# 设置工作目录
WORKDIR /app

# 首先复制go.mod和go.sum来利用Docker缓存
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY *.go ./

# 构建应用
RUN go build -o /docker-manager


FROM alpine:latest

# 创建应用目录
WORKDIR /app

# 从编译阶段复制二进制文件
COPY --from=builder /docker-manager /app/docker-manager

# 设置执行权限
RUN chmod +x /app/docker-manager

# 暴露端口
EXPOSE 15000

# 运行应用
CMD ["/app/docker-manager", "-ip", "0.0.0.0", "-port", "15000"]