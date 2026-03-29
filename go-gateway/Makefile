.PHONY: build run test clean docker-build docker-up docker-down

# 应用名称
APP_NAME=api-premium-gateway
MAIN_FILE=./cmd/server/main.go

# 构建
build:
	go build -o bin/$(APP_NAME) $(MAIN_FILE)

# 运行
run:
	go run $(MAIN_FILE)

# 测试
test:
	go test ./... -v

# 清理
clean:
	rm -rf bin/

# Docker 构建
docker-build:
	docker compose build

# Docker 启动
docker-up:
	docker compose up -d

# Docker 停止
docker-down:
	docker compose down

# Docker 重启
docker-restart:
	docker compose restart

# Docker 查看日志
docker-logs:
	docker compose logs -f

# 格式化代码
fmt:
	go fmt ./...

# 代码检查
vet:
	go vet ./...

# 下载依赖
deps:
	go mod tidy
	go mod download
