# 多阶段构建 - 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建二进制文件
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o config-syncer \
    ./cmd/config-syncer

# 运行阶段
FROM alpine:3.18

# 安装必要的包
RUN apk --no-cache add ca-certificates tzdata curl

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/config-syncer .

# 创建配置目录
RUN mkdir -p /etc/config && \
    chown -R appuser:appgroup /etc/config /app

# 切换到非root用户
USER appuser

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 暴露端口（用于健康检查和指标）
EXPOSE 8080

# 设置默认环境变量
ENV APP_ENV=production
ENV SYNC_INTERVAL=30s
ENV HTTP_TIMEOUT=30s
ENV MAX_RETRIES=3
ENV RETRY_BACKOFF=5s

# 启动命令
CMD ["./config-syncer"]