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
	@echo "  deps             安装项目依赖"
	@echo "  build            构建项目"
	@echo "  test             运行测试"
	@echo "  lint             代码检查"
	@echo "  clean            清理构建文件"
	@echo "  check            检查开发环境"
	@echo "  install          安装开发工具"
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

# 项目管理
deps: ## 安装项目依赖
	@echo "📦 安装项目依赖..."
	@echo "安装 Go 依赖..."
	@go mod download
	@go mod tidy
	@echo "安装前端依赖..."
	@cd web && npm install
	@echo "✅ 依赖安装完成"

build: ## 构建项目
	@echo "🔨 构建项目..."
	@echo "构建后端..."
	@go build -o bin/alertagent cmd/main.go
	@echo "构建前端..."
	@cd web && npm run build
	@echo "✅ 构建完成"

test: ## 运行测试
	@echo "🧪 运行测试..."
	@echo "运行后端测试..."
	@go test -v ./...
	@echo "运行前端测试..."
	@cd web && npm test
	@echo "✅ 测试完成"

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
	@rm -rf bin/
	@rm -rf web/dist/
	@rm -rf logs/
	@rm -f .backend.pid .frontend.pid
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