# 多阶段构建
FROM golang:1.18-alpine AS builder

WORKDIR /app

# 安装依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/gateway ./cmd/server/main.go

# 运行阶段
FROM alpine:3.18

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制文件和配置
COPY --from=builder /app/gateway .
COPY --from=builder /app/configs ./configs

EXPOSE 8081

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8081/api/health || exit 1

CMD ["./gateway"]
