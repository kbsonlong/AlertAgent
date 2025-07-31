# AlertAgent Core Dockerfile
# 多阶段构建：构建阶段
FROM golang:1.23-alpine AS go-builder

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

# 构建AlertAgent Core二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o alertagent-core cmd/main.go

# 前端构建阶段
FROM node:18-alpine AS web-builder

WORKDIR /app/web

# 复制前端依赖文件
COPY web/package*.json ./

# 安装前端依赖
RUN npm ci --only=production

# 复制前端源代码
COPY web/ ./

# 构建前端
RUN npm run build

# 运行阶段
FROM alpine:latest

# 安装ca证书和时区数据
RUN apk --no-cache add ca-certificates tzdata curl

# 创建非root用户
RUN addgroup -g 1001 -S alertagent && \
    adduser -u 1001 -S alertagent -G alertagent

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=go-builder /app/alertagent-core .

# 从前端构建阶段复制静态文件
COPY --from=web-builder /app/web/dist ./web/dist

# 复制配置文件
COPY config/config.yaml.example ./config/config.yaml

# 创建必要的目录
RUN mkdir -p logs && \
    chown -R alertagent:alertagent /app

# 切换到非root用户
USER alertagent

# 设置时区
ENV TZ=Asia/Shanghai

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/api/v1/health || exit 1

# 运行AlertAgent Core
ENTRYPOINT ["./alertagent-core"]