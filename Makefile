# AlertAgent 开发环境管理 Makefile
# 作者: AlertAgent Team
# 版本: 1.0.0

.PHONY: help dev dev-stop dev-restart docker-dev docker-dev-stop docker-dev-restart clean build test lint deps check install

# 默认目标
help: ## 显示帮助信息
	@echo "AlertAgent 开发环境管理命令"
	@echo "============================"
	@echo ""
	@echo "本地开发环境:"
	@echo "  dev              启动本地开发环境 (MySQL + Redis + Go + React)"
	@echo "  dev-stop         停止本地开发环境"
	@echo "  dev-restart      重启本地开发环境"
	@echo ""
	@echo "Docker 开发环境:"
	@echo "  docker-dev       启动 Docker 开发环境"
	@echo "  docker-dev-stop  停止 Docker 开发环境"
	@echo "  docker-dev-restart 重启 Docker 开发环境"
	@echo "  docker-clean     停止并清理所有 Docker 资源"
	@echo ""
	@echo "项目管理:"
	@echo "  deps             安装项目依赖
	@echo "  build            构建项目"
	@echo "  test             运行测试"
	@echo "  lint             代码检查"
	@echo "  clean            清理构建文件"
	@echo "  check            检查开发环境"
	@echo "  install          安装开发工具"
	@echo ""
	@echo "数据库迁移:"
	@echo "  migrate          执行数据库迁移"
	@echo "  migrate-status   查看迁移状态"
	@echo "  migrate-rollback 回滚迁移 (需要 VERSION=版本号)"
	@echo "  migrate-validate 验证数据库状态"
	@echo "  migrate-info     显示详细迁移信息"
	@echo "  migrate-cleanup  清理迁移历史 (需要 DAYS=天数)"
	@echo ""
	@echo "n8n 集成:"
	@echo "  n8n-start        启动 n8n 服务"
	@echo "  n8n-stop         停止 n8n 服务"
	@echo "  n8n-logs         查看 n8n 日志"
	@echo "  n8n-demo         运行 n8n 演示应用"
	@echo "  n8n-demo-build   构建 n8n 演示应用"
	@echo "  n8n-demo-test    测试 n8n 演示功能"
	@echo "  n8n-setup        设置 n8n 演示环境"
	@echo ""
	@echo "使用示例:"
	@echo "  make dev         # 启动本地开发环境"
	@echo "  make docker-dev  # 启动 Docker 开发环境"
	@echo "  make test        # 运行测试"

# 本地开发环境
dev: ## 启动本地开发环境
	@echo "🚀 启动本地开发环境..."
	@chmod +x scripts/dev-setup.sh
	@./scripts/dev-setup.sh

dev-stop: ## 停止本地开发环境
	@echo "🛑 停止本地开发环境..."
	@chmod +x scripts/dev-stop.sh
	@./scripts/dev-stop.sh

dev-restart: ## 重启本地开发环境
	@echo "🔄 重启本地开发环境..."
	@chmod +x scripts/dev-restart.sh
	@./scripts/dev-restart.sh

# Docker 开发环境
docker-dev: ## 启动 Docker 开发环境
	@echo "🐳 启动 Docker 开发环境..."
	@chmod +x scripts/docker-dev-setup.sh
	@./scripts/docker-dev-setup.sh

docker-dev-stop: ## 停止 Docker 开发环境
	@echo "🐳 停止 Docker 开发环境..."
	@chmod +x scripts/docker-dev-stop.sh
	@./scripts/docker-dev-stop.sh

docker-dev-restart: ## 重启 Docker 开发环境
	@echo "🐳 重启 Docker 开发环境..."
	@make docker-dev-stop
	@sleep 2
	@make docker-dev

docker-clean: ## 停止并清理所有 Docker 资源
	@echo "🧹 清理 Docker 资源..."
	@chmod +x scripts/docker-dev-stop.sh
	@./scripts/docker-dev-stop.sh --cleanup

# 数据库迁移
migrate: build-migrate ## 执行数据库迁移
	@echo "🗄️  执行数据库迁移..."
	@$(MIGRATE_BINARY) -action=migrate

migrate-status: build-migrate ## 查看迁移状态
	@echo "📊 查看迁移状态..."
	@$(MIGRATE_BINARY) -action=status

migrate-rollback: build-migrate ## 回滚迁移 (需要指定版本)
	@echo "⏪ 回滚迁移到指定版本..."
	@if [ -z "$(MIGRATE_VERSION)" ]; then \
		echo "❌ 错误: 请指定版本号，例如: make migrate-rollback MIGRATE_VERSION=v2.0.0-001"; \
		exit 1; \
	fi
	@$(MIGRATE_BINARY) -action=rollback -version=$(MIGRATE_VERSION)

migrate-validate: build-migrate ## 验证数据库状态
	@echo "✅ 验证数据库状态..."
	@$(MIGRATE_BINARY) -action=validate

migrate-info: build-migrate ## 显示详细迁移信息
	@echo "ℹ️  显示详细迁移信息..."
	@$(MIGRATE_BINARY) -action=info

migrate-cleanup: build-migrate ## 清理迁移历史
	@echo "🧹 清理迁移历史..."
	@$(MIGRATE_BINARY) -action=cleanup -keep-days=$(DAYS)

# Docker 迁移相关命令
migrate-docker-build: ## 构建迁移 Docker 镜像
	@echo "🐳 构建迁移 Docker 镜像..."
	@docker build -f Dockerfile.migrate -t alertagent-migrate:latest .

migrate-docker: ## 使用 Docker Compose 运行迁移
	@echo "🐳 使用 Docker Compose 运行迁移..."
	@docker-compose -f docker-compose.dev.yml --profile migration up migrate

migrate-docker-status: ## 使用 Docker 检查迁移状态
	@echo "🐳 使用 Docker 检查迁移状态..."
	@docker run --rm --network alertagent_alertagent-network \
		-e DB_HOST=postgres \
		-e DB_PORT=5432 \
		-e DB_USER=postgres \
		-e DB_PASSWORD=password \
		-e DB_NAME=alert_agent \
		alertragent-migrate:latest ./migrate -action=status

migrate-docker-validate: ## 使用 Docker 验证数据库
	@echo "🐳 使用 Docker 验证数据库..."
	@docker run --rm --network alertagent_alertagent-network \
		-e DB_HOST=postgres \
		-e DB_PORT=5432 \
		-e DB_USER=postgres \
		-e DB_PASSWORD=password \
		-e DB_NAME=alert_agent \
		alertragent-migrate:latest ./migrate -action=validate

# 快速设置命令
migrate-setup: ## 快速设置数据库迁移环境
	@echo "🚀 快速设置数据库迁移环境..."
	@./scripts/migrate-setup.sh

migrate-setup-clean: ## 清理并重新设置迁移环境
	@echo "🧹 清理并重新设置迁移环境..."
	@./scripts/migrate-setup.sh --clean

migrate-check: ## 检查迁移状态
	@echo "📊 检查迁移状态..."
	@./scripts/migrate-setup.sh --status

migrate-verify: ## 验证数据库状态
	@echo "✅ 验证数据库状态..."
	@./scripts/migrate-setup.sh --validate

# 项目管理
deps: ## 安装项目依赖
	@echo "📦 安装项目依赖..."
	@echo "安装 Go 依赖..."
	@go mod download
	@go mod tidy
	@echo "安装前端依赖..."
	@cd web && npm install
	@echo "✅ 依赖安装完成"

# 项目信息
PROJECT_NAME := alertagent
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date +%Y-%m-%d_%H:%M:%S)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 目录配置
BIN_DIR := bin
CMD_DIR := cmd

# Go 配置
GO := go
GOFLAGS := -v
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# 构建目标
API_BINARY := $(BIN_DIR)/$(PROJECT_NAME)-api
CLI_BINARY := $(BIN_DIR)/$(PROJECT_NAME)-cli
WORKER_BINARY := $(BIN_DIR)/$(PROJECT_NAME)-worker
MIGRATE_BINARY := $(BIN_DIR)/$(PROJECT_NAME)-migrate
RULE_SERVER_BINARY := $(BIN_DIR)/rule-server
N8N_DEMO_BINARY := $(BIN_DIR)/n8n-demo
MAIN_BINARY := $(BIN_DIR)/alertagent

# 创建必要的目录
$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

build: $(BIN_DIR) build-all ## 构建项目
	@echo "✅ 所有构建完成"

build-all: build-main build-migrate build-rule-server build-n8n-demo build-frontend ## 构建所有组件
	@echo "🔨 构建所有组件完成"

build-main: $(BIN_DIR) ## 构建主程序
	@echo "🔨 构建主程序..."
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(MAIN_BINARY) cmd/main.go
	@echo "✅ 主程序构建完成: $(MAIN_BINARY)"

build-api: $(BIN_DIR) ## 构建 API 服务
	@echo "🔨 构建 API 服务..."
	@if [ -d "$(CMD_DIR)/api" ]; then \
		CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(API_BINARY) ./$(CMD_DIR)/api; \
		echo "✅ API 服务构建完成: $(API_BINARY)"; \
	else \
		echo "⚠️  API 服务目录不存在，跳过构建"; \
	fi

build-cli: $(BIN_DIR) ## 构建 CLI 工具
	@echo "🔨 构建 CLI 工具..."
	@if [ -d "$(CMD_DIR)/cli" ]; then \
		CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(CLI_BINARY) ./$(CMD_DIR)/cli; \
		echo "✅ CLI 工具构建完成: $(CLI_BINARY)"; \
	else \
		echo "⚠️  CLI 工具目录不存在，跳过构建"; \
	fi

build-worker: $(BIN_DIR) ## 构建 Worker 服务
	@echo "🔨 构建 Worker 服务..."
	@if [ -d "$(CMD_DIR)/worker" ]; then \
		CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(WORKER_BINARY) ./$(CMD_DIR)/worker; \
		echo "✅ Worker 服务构建完成: $(WORKER_BINARY)"; \
	else \
		echo "⚠️  Worker 服务目录不存在，跳过构建"; \
	fi

build-migrate: $(BIN_DIR) ## 构建数据库迁移工具
	@echo "🔨 构建数据库迁移工具..."
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(MIGRATE_BINARY) ./$(CMD_DIR)/migrate
	@echo "✅ 数据库迁移工具构建完成: $(MIGRATE_BINARY)"

build-rule-server: $(BIN_DIR) ## 构建规则服务器
	@echo "🔨 构建规则服务器..."
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(RULE_SERVER_BINARY) ./$(CMD_DIR)/rule-server
	@echo "✅ 规则服务器构建完成: $(RULE_SERVER_BINARY)"

build-n8n-demo: $(BIN_DIR) ## 构建 n8n 演示应用
	@echo "🔨 构建 n8n 演示应用..."
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(N8N_DEMO_BINARY) ./$(CMD_DIR)/n8n-demo
	@echo "✅ n8n 演示应用构建完成: $(N8N_DEMO_BINARY)"

build-frontend: ## 构建前端
	@echo "🔨 构建前端..."
	@if [ -d "web" ]; then \
		cd web && npm run build; \
		echo "✅ 前端构建完成"; \
	else \
		echo "⚠️  前端目录不存在，跳过构建"; \
	fi

# 测试配置
TEST_FLAGS := -v -race -coverprofile=$(BIN_DIR)/coverage.out
TEST_TIMEOUT := 30m

test: $(BIN_DIR) test-unit test-frontend ## 运行所有测试
	@echo "✅ 所有测试完成"

test-unit: $(BIN_DIR) ## 运行单元测试
	@echo "🧪 运行单元测试..."
	@$(GO) test $(TEST_FLAGS) -timeout $(TEST_TIMEOUT) -short ./...
	@echo "✅ 单元测试完成"

test-integration: $(BIN_DIR) ## 运行集成测试
	@echo "🧪 运行集成测试..."
	@$(GO) test $(TEST_FLAGS) -timeout $(TEST_TIMEOUT) -tags=integration ./test/...
	@echo "✅ 集成测试完成"

test-frontend: ## 运行前端测试
	@echo "🧪 运行前端测试..."
	@if [ -d "web" ]; then \
		cd web && npm test; \
		echo "✅ 前端测试完成"; \
	else \
		echo "⚠️  前端目录不存在，跳过测试"; \
	fi

test-coverage: test-unit ## 生成测试覆盖率报告
	@echo "📊 生成覆盖率报告..."
	@$(GO) tool cover -html=$(BIN_DIR)/coverage.out -o $(BIN_DIR)/coverage.html
	@echo "✅ 覆盖率报告生成完成: $(BIN_DIR)/coverage.html"

bench: ## 运行基准测试
	@echo "🏃 运行基准测试..."
	@$(GO) test -bench=. -benchmem -run=^$$ ./...
	@echo "✅ 基准测试完成"

lint: ## 代码检查
	@echo "🔍 代码检查..."
	@echo "检查 Go 代码..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint 未安装，跳过 Go 代码检查"; \
	fi
	@echo "检查前端代码..."
	@cd web && npm run lint
	@echo "✅ 代码检查完成"

clean: ## 清理构建文件
	@echo "🧹 清理构建文件..."
	@rm -rf $(BIN_DIR)/
	@rm -rf web/dist/
	@rm -rf logs/
	@rm -f .backend.pid .frontend.pid
	@rm -f coverage.out coverage.html
	@$(GO) clean -cache -testcache -modcache
	@echo "✅ 清理完成"

check: ## 检查开发环境
	@echo "🔍 运行详细环境检查..."
	@chmod +x scripts/check-env.sh
	@./scripts/check-env.sh

check-simple: ## 简单环境检查
	@echo "🔍 检查开发环境..."
	@echo "检查 Go 版本:"
	@go version || echo "❌ Go 未安装"
	@echo "检查 Node.js 版本:"
	@node --version || echo "❌ Node.js 未安装"
	@echo "检查 npm 版本:"
	@npm --version || echo "❌ npm 未安装"
	@echo "检查 MySQL:"
	@mysql --version || echo "❌ MySQL 未安装"
	@echo "检查 Redis:"
	@redis-server --version || echo "❌ Redis 未安装"
	@echo "检查 Docker:"
	@docker --version || echo "❌ Docker 未安装"
	@echo "检查 Docker Compose:"
	@docker-compose --version || docker compose version || echo "❌ Docker Compose 未安装"
	@echo "✅ 环境检查完成"

install: ## 安装开发工具
	@echo "🛠️  安装开发工具..."
	@echo "安装 golangci-lint..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2; \
	else \
		echo "golangci-lint 已安装"; \
	fi
	@echo "安装 air (热重载工具)..."
	@if ! command -v air >/dev/null 2>&1; then \
		go install github.com/cosmtrek/air@latest; \
	else \
		echo "air 已安装"; \
	fi
	@echo "✅ 开发工具安装完成"

# 快捷命令别名
start: dev ## 启动开发环境 (dev 的别名)
stop: dev-stop ## 停止开发环境 (dev-stop 的别名)
restart: dev-restart ## 重启开发环境 (dev-restart 的别名)

# 交叉编译
build-cross: $(BIN_DIR) ## 交叉编译多平台二进制文件
	@echo "🔨 交叉编译多平台二进制文件..."
	@for os in linux darwin windows; do \
		for arch in amd64 arm64; do \
			if [ "$$os" = "windows" ]; then \
				ext=".exe"; \
			else \
				ext=""; \
			fi; \
			echo "构建 $$os/$$arch..."; \
			CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch $(GO) build $(GOFLAGS) $(LDFLAGS) \
				-o $(BIN_DIR)/$(PROJECT_NAME)-main-$$os-$$arch$$ext cmd/main.go; \
			CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch $(GO) build $(GOFLAGS) $(LDFLAGS) \
				-o $(BIN_DIR)/$(PROJECT_NAME)-migrate-$$os-$$arch$$ext ./$(CMD_DIR)/migrate; \
			CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch $(GO) build $(GOFLAGS) $(LDFLAGS) \
				-o $(BIN_DIR)/rule-server-$$os-$$arch$$ext ./$(CMD_DIR)/rule-server; \
		done; \
	done
	@echo "✅ 交叉编译完成"

# 安装开发工具
install-tools: ## 安装开发工具
	@echo "🛠️  安装开发工具..."
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GO) install golang.org/x/tools/cmd/goimports@latest
	@$(GO) install github.com/google/wire/cmd/wire@latest
	@$(GO) install github.com/swaggo/swag/cmd/swag@latest
	@$(GO) install github.com/cosmtrek/air@latest
	@echo "✅ 开发工具安装完成"

# 代码格式化
fmt: ## 格式化代码
	@echo "🎨 格式化代码..."
	@$(GO) fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w -local alert_agent .; \
	else \
		echo "⚠️  goimports 未安装，跳过导入整理"; \
	fi
	@echo "✅ 代码格式化完成"

# 生成代码
generate: ## 生成代码
	@echo "⚙️  生成代码..."
	@$(GO) generate ./...
	@if command -v wire >/dev/null 2>&1; then \
		wire ./internal/wire/...; \
	else \
		echo "⚠️  wire 未安装，跳过依赖注入代码生成"; \
	fi
	@echo "✅ 代码生成完成"

# 生成 API 文档
docs: ## 生成 API 文档
	@echo "📚 生成 API 文档..."
	@mkdir -p docs/swagger
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g ./$(CMD_DIR)/api/main.go -o ./docs/swagger; \
		echo "✅ API 文档生成完成"; \
	else \
		echo "⚠️  swag 未安装，跳过 API 文档生成"; \
	fi

# 开发环境设置
dev-setup: install-tools deps generate ## 设置开发环境
	@echo "✅ 开发环境设置完成"

# 代码质量检查
quality: fmt lint test ## 运行代码质量检查
	@echo "✅ 代码质量检查完成"

# 发布准备
release: clean quality build-cross ## 准备发布
	@echo "✅ 发布准备完成"

# 快速构建（跳过测试）
quick: deps generate build ## 快速构建（跳过测试）
	@echo "✅ 快速构建完成"

# 检查构建状态
check-build: ## 检查构建状态
	@echo "📊 项目信息:"
	@echo "  项目名称: $(PROJECT_NAME)"
	@echo "  版本: $(VERSION)"
	@echo "  构建时间: $(BUILD_TIME)"
	@echo "  Git 提交: $(GIT_COMMIT)"
	@echo "  Go 版本: $(shell $(GO) version)"
	@echo "  操作系统: $(GOOS)"
	@echo "  架构: $(GOARCH)"
	@echo ""
	@echo "📁 构建目标:"
	@ls -la $(BIN_DIR)/ 2>/dev/null || echo "  无构建产物"

# 监控文件变化并重新构建
watch: ## 监控文件变化并重新构建
	@echo "👀 监控文件变化..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	elif command -v fswatch >/dev/null 2>&1; then \
		fswatch -o . -e ".*" -i "\.go$$" | xargs -n1 -I{} make build-main; \
	else \
		echo "❌ 请安装 air 或 fswatch: go install github.com/cosmtrek/air@latest 或 brew install fswatch"; \
	fi

# 安全扫描
security: ## 运行安全扫描
	@echo "🔒 运行安全扫描..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "⚠️  gosec 未安装，正在安装..."; \
		$(GO) install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
		gosec ./...; \
	fi
	@echo "✅ 安全扫描完成"

# 依赖检查
deps-check: ## 检查依赖更新
	@echo "📦 检查依赖更新..."
	@$(GO) list -u -m all
	@echo "✅ 依赖检查完成"

# 显示项目统计信息
stats: ## 显示项目统计信息
	@echo "📈 项目统计信息:"
	@echo "  Go 文件数量: $(shell find . -name '*.go' | wc -l)"
	@echo "  代码行数: $(shell find . -name '*.go' -exec wc -l {} + | tail -1 | awk '{print $$1}')"
	@echo "  包数量: $(shell $(GO) list ./... | wc -l)"
	@echo "  依赖数量: $(shell $(GO) list -m all | wc -l)"

# 验证构建
verify: clean deps generate build test ## 完整验证构建
	@echo "✅ 构建验证完成"

# 显示当前状态
status: ## 显示服务状态
	@echo "📊 服务状态检查"
	@echo "================"
	@echo "检查端口占用情况:"
	@echo "后端 (8080):"
	@lsof -i :8080 || echo "  端口 8080 未被占用"
	@echo "前端 (5173):"
	@lsof -i :5173 || echo "  端口 5173 未被占用"
	@echo "MySQL (3306):"
	@lsof -i :3306 || echo "  端口 3306 未被占用"
	@echo "Redis (6379):"
	@lsof -i :6379 || echo "  端口 6379 未被占用"
	@echo "Ollama (11434):"
	@lsof -i :11434 || echo "  端口 11434 未被占用"
	@echo ""
	@echo "检查 Docker 容器:"
	@if command -v docker >/dev/null 2>&1; then \
		docker ps --filter "name=alertagent" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" || echo "  没有运行的 AlertAgent 容器"; \
	else \
		echo "  Docker 未安装"; \
	fi

# 演示和测试
demo: ## 运行功能演示
	@echo "🎯 运行功能演示..."
	@chmod +x scripts/demo.sh
	@./scripts/demo.sh

demo-api: ## 演示 API 功能
	@echo "🎯 演示 API 功能..."
	@chmod +x scripts/demo.sh
	@./scripts/demo.sh --api

demo-frontend: ## 演示前端功能
	@echo "🎯 演示前端功能..."
	@chmod +x scripts/demo.sh
	@./scripts/demo.sh --frontend

# 日志查看
logs: ## 查看应用日志
	@echo "📋 查看应用日志"
	@if [ -f "logs/alert_agent.log" ]; then \
		tail -f logs/alert_agent.log; \
	else \
		echo "日志文件不存在: logs/alert_agent.log"; \
	fi

docker-logs: ## 查看 Docker 服务日志
	@echo "📋 查看 Docker 服务日志"
	@if [ -f "docker-compose.dev.yml" ]; then \
		if docker compose version >/dev/null 2>&1; then \
			docker compose -f docker-compose.dev.yml logs -f; \
		else \
			docker-compose -f docker-compose.dev.yml logs -f; \
		fi; \
	else \
		echo "docker-compose.dev.yml 文件不存在"; \
	fi

# 运行服务
run-main: build-main ## 运行主程序
	@echo "🚀 启动主程序..."
	@$(MAIN_BINARY)

run-api: build-api ## 运行 API 服务
	@echo "🚀 启动 API 服务..."
	@$(API_BINARY)

run-worker: build-worker ## 运行 Worker 服务
	@echo "🚀 启动 Worker 服务..."
	@$(WORKER_BINARY)

# Run rule server locally
run-rule-server: build-rule-server ## 运行规则服务器
	@echo "🚀 启动规则服务器..."
	DB_HOST=localhost DB_PORT=3306 DB_USER=root DB_PASSWORD=password DB_NAME=alert_agent PORT=8080 $(RULE_SERVER_BINARY)

# Test rule server APIs
test-rule-server:
	@echo "Testing rule server APIs..."
	./scripts/test-api.sh

# Build rule server Docker image
docker-rule-server:
	@echo "Building rule server Docker image..."
	docker build -f Dockerfile.rule-server -t alert-agent/rule-server:latest .

# Start rule server with Docker Compose
docker-up-rule-server:
	@echo "Starting rule server with Docker Compose..."
	docker-compose -f docker-compose.rule-server.yml up -d

# Stop rule server Docker Compose
docker-down-rule-server:
	@echo "Stopping rule server Docker Compose..."
	docker-compose -f docker-compose.rule-server.yml down

# View rule server logs
logs-rule-server:
	@echo "Viewing rule server logs..."
	docker-compose -f docker-compose.rule-server.yml logs -f rule-server

# Clean rule server artifacts
clean-rule-server:
	@echo "Cleaning rule server artifacts..."
	rm -f bin/rule-server
	rm -f rule-server

# n8n 集成相关命令
.PHONY: n8n-start n8n-stop n8n-logs n8n-demo n8n-demo-build n8n-demo-test n8n-setup

# n8n 服务管理
n8n-start: ## 启动 n8n 服务
	@echo "🚀 启动 n8n 服务..."
	docker run -d \
		--name n8n \
		-p 5678:5678 \
		-e GENERIC_TIMEZONE="Asia/Shanghai" \
		-e TZ="Asia/Shanghai" \
		-v n8n_data:/home/node/.n8n \
		n8nio/n8n
	@echo "✅ n8n 服务已启动"
	@echo "📱 管理界面: http://localhost:5678"

n8n-stop: ## 停止 n8n 服务
	@echo "🛑 停止 n8n 服务..."
	docker stop n8n || true
	docker rm n8n || true
	@echo "✅ n8n 服务已停止"

n8n-logs: ## 查看 n8n 日志
	@echo "📋 查看 n8n 日志..."
	docker logs -f n8n

# n8n 演示应用
n8n-demo: build-n8n-demo ## 运行 n8n 演示应用
	@echo "🚀 启动 n8n 演示应用..."
	DB_HOST=localhost \
	DB_PORT=3306 \
	DB_USER=alertagent \
	DB_PASSWORD=alertagent123 \
	DB_NAME=alertagent \
	N8N_BASE_URL=http://localhost:5678 \
	N8N_API_KEY=your-n8n-api-key \
	PORT=8080 \
	GIN_MODE=debug \
	$(N8N_DEMO_BINARY)

n8n-demo-test: ## 测试 n8n 演示功能
	@echo "🧪 测试 n8n 演示功能..."
	@echo "创建测试告警..."
	@curl -X POST http://localhost:8080/api/v1/demo/alerts \
		-H "Content-Type: application/json" \
		-d '{
			"title": "测试告警",
			"description": "这是一个 n8n 集成测试告警",
			"severity": "high",
			"source": "demo",
			"metadata": {"test": true, "integration": "n8n"}
		}' || echo "❌ 创建告警失败，请确保演示应用正在运行"
	@echo ""
	@echo "获取演示统计..."
	@curl -X GET http://localhost:8080/api/v1/demo/stats || echo "❌ 获取统计失败"
	@echo ""
	@echo "检查健康状态..."
	@curl -X GET http://localhost:8080/health || echo "❌ 健康检查失败"

n8n-setup: n8n-start ## 设置完整的 n8n 演示环境
	@echo "⚙️  设置 n8n 演示环境..."
	@echo "等待 n8n 服务启动..."
	sleep 10
	@echo "✅ n8n 演示环境设置完成!"
	@echo ""
	@echo "🎯 下一步操作:"
	@echo "1. 访问 n8n 管理界面: http://localhost:5678"
	@echo "2. 创建工作流模板"
	@echo "3. 运行演示应用: make n8n-demo"
	@echo "4. 测试功能: make n8n-demo-test"
	@echo ""
	@echo "📚 更多信息请查看: docs/n8n-integration-guide.md"