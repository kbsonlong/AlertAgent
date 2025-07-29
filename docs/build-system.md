# AlertAgent 构建系统文档

## 概述

AlertAgent 项目使用 Makefile 管理构建流程，所有二进制文件统一输出到 `bin/` 目录。构建系统支持模块化构建、交叉编译、测试、代码质量检查等功能。

## 目录结构

```
AlertAgent/
├── bin/                    # 构建输出目录
│   ├── alertagent         # 主程序
│   ├── alertagent-migrate # 数据库迁移工具
│   ├── rule-server        # 规则服务器
│   └── n8n-demo          # n8n 演示应用
├── cmd/                   # 命令行程序源码
├── internal/              # 内部包
└── Makefile              # 构建配置
```

## 快速开始

### 查看所有可用命令

```bash
make help
```

### 构建所有组件

```bash
make build
```

### 清理构建产物

```bash
make clean
```

## 构建命令

### 基础构建

| 命令 | 描述 | 输出文件 |
|------|------|----------|
| `make build` | 构建所有组件 | `bin/` 目录下所有二进制文件 |
| `make build-main` | 构建主程序 | `bin/alertagent` |
| `make build-migrate` | 构建迁移工具 | `bin/alertagent-migrate` |
| `make build-rule-server` | 构建规则服务器 | `bin/rule-server` |
| `make build-n8n-demo` | 构建 n8n 演示应用 | `bin/n8n-demo` |
| `make build-frontend` | 构建前端 | `web/dist/` |

### 高级构建

| 命令 | 描述 |
|------|------|
| `make build-cross` | 交叉编译多平台二进制文件 |
| `make quick` | 快速构建（跳过测试） |
| `make release` | 发布准备（清理+质量检查+交叉编译） |

## 运行命令

| 命令 | 描述 |
|------|------|
| `make run-main` | 运行主程序 |
| `make run-api` | 运行 API 服务 |
| `make run-worker` | 运行 Worker 服务 |
| `make run-rule-server` | 运行规则服务器 |
| `make n8n-demo` | 运行 n8n 演示应用 |

## 测试命令

| 命令 | 描述 |
|------|------|
| `make test` | 运行所有测试 |
| `make test-unit` | 运行单元测试 |
| `make test-integration` | 运行集成测试 |
| `make test-frontend` | 运行前端测试 |
| `make test-coverage` | 生成测试覆盖率报告 |
| `make bench` | 运行基准测试 |

## 代码质量

| 命令 | 描述 |
|------|------|
| `make fmt` | 格式化代码 |
| `make lint` | 代码检查 |
| `make security` | 安全扫描 |
| `make quality` | 完整代码质量检查 |

## 开发工具

| 命令 | 描述 |
|------|------|
| `make install-tools` | 安装开发工具 |
| `make generate` | 生成代码 |
| `make docs` | 生成 API 文档 |
| `make watch` | 监控文件变化并重新构建 |

## 数据库迁移

| 命令 | 描述 |
|------|------|
| `make migrate` | 执行数据库迁移 |
| `make migrate-status` | 查看迁移状态 |
| `make migrate-rollback MIGRATE_VERSION=v1.0.0` | 回滚到指定版本 |
| `make migrate-validate` | 验证数据库状态 |

## 环境管理

| 命令 | 描述 |
|------|------|
| `make dev` | 启动本地开发环境 |
| `make docker-dev` | 启动 Docker 开发环境 |
| `make dev-setup` | 设置开发环境 |
| `make check` | 检查开发环境 |

## 构建变量

构建系统支持以下环境变量：

- `PROJECT_NAME`: 项目名称（默认：alertagent）
- `VERSION`: 版本号（自动从 git 获取）
- `BUILD_TIME`: 构建时间（自动生成）
- `GIT_COMMIT`: Git 提交哈希（自动获取）
- `GOOS`: 目标操作系统
- `GOARCH`: 目标架构

## 交叉编译

支持的平台：
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64, arm64)

```bash
make build-cross
```

生成的文件命名格式：
- `alertagent-main-{os}-{arch}[.exe]`
- `alertagent-migrate-{os}-{arch}[.exe]`
- `rule-server-{os}-{arch}[.exe]`

## 热重载开发

项目支持使用 Air 进行热重载开发：

```bash
# 安装 Air
make install-tools

# 启动热重载
make watch
```

配置文件：`.air.toml`

## 构建信息

查看当前构建状态：

```bash
make check-build
```

显示项目统计信息：

```bash
make stats
```

## 最佳实践

1. **开发前准备**：
   ```bash
   make dev-setup
   ```

2. **日常开发**：
   ```bash
   make watch  # 启动热重载
   ```

3. **提交前检查**：
   ```bash
   make quality
   ```

4. **发布准备**：
   ```bash
   make release
   ```

5. **清理环境**：
   ```bash
   make clean
   ```

## 故障排除

### 构建失败

1. 检查 Go 版本：
   ```bash
   go version
   ```

2. 更新依赖：
   ```bash
   make deps
   ```

3. 清理缓存：
   ```bash
   make clean
   go clean -cache -testcache -modcache
   ```

### 工具缺失

安装所有开发工具：
```bash
make install-tools
```

### 权限问题

确保 bin 目录有写权限：
```bash
chmod 755 bin/
```

## 扩展构建系统

要添加新的构建目标，编辑 `Makefile`：

1. 添加新的二进制文件变量
2. 创建构建目标
3. 更新 `build-all` 依赖
4. 添加运行目标（可选）

示例：
```makefile
# 新的二进制文件
NEW_BINARY := $(BIN_DIR)/new-component

# 构建目标
build-new: $(BIN_DIR)
	@echo "🔨 构建新组件..."
	@CGO_ENABLED=0 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(NEW_BINARY) ./$(CMD_DIR)/new
	@echo "✅ 新组件构建完成: $(NEW_BINARY)"

# 运行目标
run-new: build-new
	@echo "🚀 启动新组件..."
	@$(NEW_BINARY)
```