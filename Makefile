# AlertAgent Makefile
# 提供统一的构建、测试和部署命令

# 变量定义
PROJECT_NAME := alertagent
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION := $(shell go version | cut -d' ' -f3)
GIT_COMMIT := $(shell git rev-parse HEAD)

# 构建标志
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -s -w"

# 目录定义
BIN_DIR := bin
DIST_DIR := dist
COVERAGE_DIR := coverage
TEST_RESULTS_DIR := test-results

# Go相关变量
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# Docker相关变量
DOCKER_REGISTRY := alertagent
DOCKER_TAG := $(VERSION)

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "AlertAgent 构建系统"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 清理
.PHONY: clean
clean: ## 清理构建产物
	@echo "清理构建产物..."
	@rm -rf $(BIN_DIR) $(DIST_DIR) $(COVERAGE_DIR) $(TEST_RESULTS_DIR)
	@go clean -cache -testcache -modcache
	@docker system prune -f

# 依赖管理
.PHONY: deps
deps: ## 下载依赖
	@echo "下载Go依赖..."
	@go mod download
	@go mod tidy
	@go mod verify

.PHONY: deps-update
deps-update: ## 更新依赖
	@echo "更新Go依赖..."
	@go get -u ./...
	@go mod tidy

# 代码质量检查
.PHONY: lint
lint: ## 运行代码检查
	@echo "运行代码检查..."
	@go vet ./...
	@golangci-lint run --timeout=5m

.PHONY: fmt
fmt: ## 格式化代码
	@echo "格式化代码..."
	@go fmt ./...
	@goimports -w .

.PHONY: security
security: ## 安全扫描
	@echo "运行安全扫描..."
	@gosec ./...

# 测试相关
.PHONY: test
test: ## 运行所有测试
	@./scripts/test_automation.sh all -r

.PHONY: test-unit
test-unit: ## 运行单元测试
	@./scripts/test_automation.sh unit -c

.PHONY: test-integration
test-integration: ## 运行集成测试
	@./scripts/test_automation.sh integration -v

.PHONY: test-performance
test-performance: ## 运行性能测试
	@./scripts/test_automation.sh performance -r

.PHONY: test-frontend
test-frontend: ## 运行前端测试
	@./scripts/test_automation.sh frontend

.PHONY: test-coverage
test-coverage: ## 生成测试覆盖率报告
	@echo "生成测试覆盖率报告..."
	@mkdir -p $(COVERAGE_DIR)
	@go test -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./internal/...
	@go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@go tool cover -func=$(COVERAGE_DIR)/coverage.out | tail -1

.PHONY: test-setup
test-setup: ## 设置测试环境
	@./scripts/test_automation.sh --setup

.PHONY: test-cleanup
test-cleanup: ## 清理测试环境
	@./scripts/test_automation.sh --clean

# 构建相关
.PHONY: build
build: deps ## 构建所有二进制文件
	@echo "构建二进制文件..."
	@mkdir -p $(BIN_DIR)
	@go build $(LDFLAGS) -o $(BIN_DIR)/$(PROJECT_NAME) cmd/main.go
	@go build $(LDFLAGS) -o $(BIN_DIR)/$(PROJECT_NAME)-worker cmd/worker/main.go
	@go build $(LDFLAGS) -o $(BIN_DIR)/$(PROJECT_NAME)-sidecar cmd/sidecar/main.go
	@go build $(LDFLAGS) -o $(BIN_DIR)/$(PROJECT_NAME)-migrate cmd/migrate/main.go

.PHONY: build-core
build-core: deps ## 构建核心服务
	@echo "构建核心服务..."
	@mkdir -p $(BIN_DIR)
	@go build $(LDFLAGS) -o $(BIN_DIR)/$(PROJECT_NAME) cmd/main.go

.PHONY: build-worker
build-worker: deps ## 构建Worker服务
	@echo "构建Worker服务..."
	@mkdir -p $(BIN_DIR)
	@go build $(LDFLAGS) -o $(BIN_DIR)/$(PROJECT_NAME)-worker cmd/worker/main.go

.PHONY: build-sidecar
build-sidecar: deps ## 构建Sidecar服务
	@echo "构建Sidecar服务..."
	@mkdir -p $(BIN_DIR)
	@go build $(LDFLAGS) -o $(BIN_DIR)/$(PROJECT_NAME)-sidecar cmd/sidecar/main.go

.PHONY: build-frontend
build-frontend: ## 构建前端
	@echo "构建前端..."
	@cd web && npm ci && npm run build

# 跨平台构建
.PHONY: build-all
build-all: deps ## 构建所有平台的二进制文件
	@echo "构建所有平台的二进制文件..."
	@mkdir -p $(DIST_DIR)
	@for os in linux darwin windows; do \
		for arch in amd64 arm64; do \
			if [ "$$os" = "windows" ] && [ "$$arch" = "arm64" ]; then continue; fi; \
			echo "构建 $$os/$$arch..."; \
			GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $(DIST_DIR)/$(PROJECT_NAME)-$$os-$$arch cmd/main.go; \
			GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $(DIST_DIR)/$(PROJECT_NAME)-worker-$$os-$$arch cmd/worker/main.go; \
			GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $(DIST_DIR)/$(PROJECT_NAME)-sidecar-$$os-$$arch cmd/sidecar/main.go; \
			if [ "$$os" = "windows" ]; then \
				mv $(DIST_DIR)/$(PROJECT_NAME)-$$os-$$arch $(DIST_DIR)/$(PROJECT_NAME)-$$os-$$arch.exe; \
				mv $(DIST_DIR)/$(PROJECT_NAME)-worker-$$os-$$arch $(DIST_DIR)/$(PROJECT_NAME)-worker-$$os-$$arch.exe; \
				mv $(DIST_DIR)/$(PROJECT_NAME)-sidecar-$$os-$$arch $(DIST_DIR)/$(PROJECT_NAME)-sidecar-$$os-$$arch.exe; \
			fi; \
		done; \
	done

# Docker相关
.PHONY: docker-build
docker-build: ## 构建Docker镜像
	@echo "构建Docker镜像..."
	@docker build -t $(DOCKER_REGISTRY)/core:$(DOCKER_TAG) .
	@docker build -f Dockerfile.worker -t $(DOCKER_REGISTRY)/worker:$(DOCKER_TAG) .
	@docker build -f Dockerfile.sidecar -t $(DOCKER_REGISTRY)/sidecar:$(DOCKER_TAG) .

.PHONY: docker-push
docker-push: docker-build ## 推送Docker镜像
	@echo "推送Docker镜像..."
	@docker push $(DOCKER_REGISTRY)/core:$(DOCKER_TAG)
	@docker push $(DOCKER_REGISTRY)/worker:$(DOCKER_TAG)
	@docker push $(DOCKER_REGISTRY)/sidecar:$(DOCKER_TAG)

.PHONY: docker-run
docker-run: ## 运行Docker容器
	@echo "运行Docker容器..."
	@docker-compose up -d

.PHONY: docker-stop
docker-stop: ## 停止Docker容器
	@echo "停止Docker容器..."
	@docker-compose down

# 开发相关
.PHONY: dev
dev: ## 启动开发环境
	@echo "启动开发环境..."
	@./scripts/dev-setup.sh

.PHONY: dev-stop
dev-stop: ## 停止开发环境
	@echo "停止开发环境..."
	@./scripts/dev-stop.sh

.PHONY: dev-restart
dev-restart: dev-stop dev ## 重启开发环境

.PHONY: run
run: build ## 运行应用
	@echo "运行应用..."
	@./$(BIN_DIR)/$(PROJECT_NAME)

.PHONY: run-worker
run-worker: build-worker ## 运行Worker
	@echo "运行Worker..."
	@./$(BIN_DIR)/$(PROJECT_NAME)-worker

.PHONY: run-sidecar
run-sidecar: build-sidecar ## 运行Sidecar
	@echo "运行Sidecar..."
	@./$(BIN_DIR)/$(PROJECT_NAME)-sidecar

# 数据库相关
.PHONY: db-migrate
db-migrate: build ## 运行数据库迁移
	@echo "运行数据库迁移..."
	@./$(BIN_DIR)/$(PROJECT_NAME)-migrate

.PHONY: db-setup
db-setup: ## 设置数据库
	@echo "设置数据库..."
	@mysql -h localhost -u root -ppassword < scripts/init.sql

.PHONY: db-reset
db-reset: ## 重置数据库
	@echo "重置数据库..."
	@mysql -h localhost -u root -ppassword -e "DROP DATABASE IF EXISTS alertagent;"
	@mysql -h localhost -u root -ppassword < scripts/init.sql

# 部署相关
.PHONY: deploy-staging
deploy-staging: ## 部署到测试环境
	@echo "部署到测试环境..."
	@kubectl apply -f k8s/staging/

.PHONY: deploy-production
deploy-production: ## 部署到生产环境
	@echo "部署到生产环境..."
	@kubectl apply -f k8s/production/

# 监控和日志
.PHONY: logs
logs: ## 查看应用日志
	@echo "查看应用日志..."
	@docker-compose logs -f

.PHONY: logs-core
logs-core: ## 查看核心服务日志
	@echo "查看核心服务日志..."
	@docker-compose logs -f alertagent-core

.PHONY: logs-worker
logs-worker: ## 查看Worker日志
	@echo "查看Worker日志..."
	@docker-compose logs -f alertagent-worker

# 性能分析
.PHONY: profile
profile: ## 运行性能分析
	@echo "运行性能分析..."
	@go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./tests/performance/...
	@go tool pprof cpu.prof
	@go tool pprof mem.prof

.PHONY: benchmark
benchmark: ## 运行基准测试
	@echo "运行基准测试..."
	@go test -bench=. -benchmem ./...

# 文档生成
.PHONY: docs
docs: ## 生成文档
	@echo "生成文档..."
	@godoc -http=:6060

.PHONY: swagger
swagger: ## 生成API文档
	@echo "生成API文档..."
	@swag init -g cmd/main.go

# 发布相关
.PHONY: release
release: clean test build-all docker-build ## 创建发布版本
	@echo "创建发布版本 $(VERSION)..."
	@mkdir -p $(DIST_DIR)/release
	@tar -czf $(DIST_DIR)/release/$(PROJECT_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(DIST_DIR) $(PROJECT_NAME)-linux-amd64 $(PROJECT_NAME)-worker-linux-amd64 $(PROJECT_NAME)-sidecar-linux-amd64
	@tar -czf $(DIST_DIR)/release/$(PROJECT_NAME)-$(VERSION)-darwin-amd64.tar.gz -C $(DIST_DIR) $(PROJECT_NAME)-darwin-amd64 $(PROJECT_NAME)-worker-darwin-amd64 $(PROJECT_NAME)-sidecar-darwin-amd64
	@zip -j $(DIST_DIR)/release/$(PROJECT_NAME)-$(VERSION)-windows-amd64.zip $(DIST_DIR)/$(PROJECT_NAME)-windows-amd64.exe $(DIST_DIR)/$(PROJECT_NAME)-worker-windows-amd64.exe $(DIST_DIR)/$(PROJECT_NAME)-sidecar-windows-amd64.exe

.PHONY: tag
tag: ## 创建Git标签
	@echo "创建Git标签 $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)

# CI/CD相关
.PHONY: ci
ci: deps lint test build ## CI流水线
	@echo "CI流水线完成"

.PHONY: cd
cd: ci docker-build docker-push ## CD流水线
	@echo "CD流水线完成"

# 质量门禁
.PHONY: quality-gate
quality-gate: ## 质量门禁检查
	@echo "运行质量门禁检查..."
	@./scripts/quality_gate.sh

# 安装工具
.PHONY: install-tools
install-tools: ## 安装开发工具
	@echo "安装开发工具..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/swaggo/swag/cmd/swag@latest

# 版本信息
.PHONY: version
version: ## 显示版本信息
	@echo "项目: $(PROJECT_NAME)"
	@echo "版本: $(VERSION)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "Go版本: $(GO_VERSION)"
	@echo "Git提交: $(GIT_COMMIT)"

# 健康检查
.PHONY: health
health: ## 健康检查
	@echo "运行健康检查..."
	@curl -f http://localhost:8080/health || echo "服务未运行"

# 备份和恢复
.PHONY: backup
backup: ## 备份数据
	@echo "备份数据..."
	@./scripts/backup.sh

.PHONY: restore
restore: ## 恢复数据
	@echo "恢复数据..."
	@./scripts/restore.sh

# 监控指标
.PHONY: metrics
metrics: ## 查看监控指标
	@echo "查看监控指标..."
	@curl -s http://localhost:8080/metrics

# 调试相关
.PHONY: debug
debug: ## 调试模式运行
	@echo "调试模式运行..."
	@dlv debug cmd/main.go

# 检查依赖更新
.PHONY: check-updates
check-updates: ## 检查依赖更新
	@echo "检查依赖更新..."
	@go list -u -m all

# 生成模拟数据
.PHONY: mock-data
mock-data: ## 生成模拟数据
	@echo "生成模拟数据..."
	@./scripts/generate_mock_data.sh

# 压力测试
.PHONY: stress-test
stress-test: ## 运行压力测试
	@echo "运行压力测试..."
	@./scripts/stress_test.sh

# 确保脚本可执行
$(shell chmod +x scripts/*.sh)

# 检查必要的工具
.PHONY: check-tools
check-tools: ## 检查必要工具
	@echo "检查必要工具..."
	@command -v go >/dev/null 2>&1 || { echo "Go未安装"; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "Docker未安装"; exit 1; }
	@command -v kubectl >/dev/null 2>&1 || { echo "kubectl未安装"; exit 1; }
	@echo "所有必要工具已安装"