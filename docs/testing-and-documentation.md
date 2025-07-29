# AlertAgent API 文档和测试指南

本文档介绍了 AlertAgent 项目的 API 文档生成和测试套件的使用方法。

## 📚 API 文档

### OpenAPI 规范

项目提供了完整的 OpenAPI 3.0.3 规范文档：

- **文件位置**: `docs/openapi.yaml`
- **内容**: 包含所有 API 端点、请求/响应模型、认证方式和错误处理
- **模块**: 健康检查、告警分析、通道管理、集群管理、插件管理、认证等

### API 使用示例

详细的 API 使用示例文档：

- **文件位置**: `docs/api-examples.md`
- **内容**: 包含 curl 命令示例、响应示例、批量操作、性能测试等
- **覆盖**: 所有主要 API 端点的实际使用场景

### 文档生成命令

```bash
# 生成所有 API 文档
make docs

# 仅生成 Swagger 文档
make docs-swagger

# 验证 OpenAPI 规范
make docs-openapi

# 启动文档服务器 (http://localhost:8000)
make docs-serve

# 清理生成的文档
make docs-clean
```

## 🧪 测试套件

### 测试类型

项目提供了多种类型的测试：

#### 1. 单元测试
- **位置**: 各模块的 `*_test.go` 文件
- **用途**: 测试单个函数和方法的功能
- **运行**: `make test-unit`

#### 2. 集成测试
- **位置**: `test/integration/api_test.go`
- **用途**: 端到端 API 功能测试
- **覆盖**: 健康检查、认证、分析、通道、集群、插件管理
- **运行**: `make test-integration`

#### 3. 性能测试
- **位置**: `test/performance/load_test.go`
- **用途**: API 性能和负载测试
- **指标**: 响应时间、吞吐量、并发处理能力
- **运行**: `make test-performance`

#### 4. 负载测试
- **位置**: `test/performance/load_test.go`
- **用途**: 高并发场景下的系统稳定性测试
- **运行**: `make test-load`

#### 5. 兼容性测试
- **位置**: `test/compatibility/version_test.go`
- **用途**: API 版本间的兼容性验证
- **覆盖**: V1/V2 版本兼容性、向后兼容性
- **运行**: `make test-compatibility`

### 测试命令

```bash
# 运行所有类型的测试
make test-all

# 运行基础测试（单元 + 前端）
make test

# 运行单元测试
make test-unit

# 运行集成测试
make test-integration

# 运行性能测试
make test-performance

# 运行负载测试
make test-load

# 运行兼容性测试
make test-compatibility

# 生成测试覆盖率报告
make test-coverage

# 清理测试环境
make test-clean
```

### 测试脚本

项目提供了统一的测试脚本 `scripts/run-tests.sh`，支持：

- **多种测试类型**: unit, integration, performance, load, compatibility, coverage
- **命令行参数**: --verbose, --clean, --parallel, --timeout, --race, --bench
- **环境管理**: 自动设置测试环境变量
- **日志记录**: 详细的测试日志输出

#### 使用示例

```bash
# 运行所有测试（详细输出 + 竞态检测）
./scripts/run-tests.sh all --verbose --race

# 运行集成测试（并行 + 清理）
./scripts/run-tests.sh integration --parallel --clean

# 运行性能测试（基准测试）
./scripts/run-tests.sh performance --bench

# 生成覆盖率报告
./scripts/run-tests.sh coverage
```

## 🔧 环境配置

### 测试环境变量

测试需要以下环境变量：

```bash
# 集成测试
export INTEGRATION_TEST=true

# 数据库配置
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=test
export DB_PASSWORD=test
export DB_NAME=alertagent_test

# Redis 配置
export REDIS_HOST=localhost
export REDIS_PORT=6379

# API 配置
export API_PORT=8080
export JWT_SECRET=test-secret
```

### 依赖工具

确保安装以下工具：

```bash
# Go 测试工具
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/go-swagger/go-swagger/cmd/swagger@latest

# 测试依赖
go mod download
```

## 📊 测试报告

### 覆盖率报告

运行 `make test-coverage` 后，会生成：

- **文本报告**: 控制台输出
- **HTML 报告**: `coverage.html`
- **原始数据**: `coverage.out`

### 性能报告

性能测试会输出：

- **响应时间统计**: 平均值、中位数、P95、P99
- **吞吐量指标**: 每秒请求数 (RPS)
- **错误率**: 失败请求百分比
- **并发性能**: 不同并发级别下的表现

## 🚀 CI/CD 集成

### GitHub Actions

项目支持在 CI/CD 流水线中运行测试：

```yaml
- name: Run Tests
  run: |
    make test-all
    make test-coverage

- name: Generate Docs
  run: |
    make docs
```

### Docker 测试

支持在 Docker 环境中运行测试：

```bash
# 构建测试镜像
docker build -f Dockerfile.test -t alertagent-test .

# 运行测试
docker run --rm alertagent-test make test-all
```

## 📝 最佳实践

### 测试编写

1. **单元测试**:
   - 测试单一功能
   - 使用 mock 对象
   - 覆盖边界条件

2. **集成测试**:
   - 测试完整流程
   - 使用真实数据库
   - 验证 API 契约

3. **性能测试**:
   - 设置合理的负载
   - 监控关键指标
   - 建立性能基线

### 文档维护

1. **OpenAPI 规范**:
   - 保持与代码同步
   - 提供详细的描述
   - 包含完整的示例

2. **API 示例**:
   - 覆盖常见用例
   - 提供错误处理示例
   - 包含性能优化建议

## 🔍 故障排除

### 常见问题

1. **测试失败**:
   - 检查环境变量配置
   - 确认数据库连接
   - 查看详细错误日志

2. **文档生成失败**:
   - 安装 swag 工具
   - 检查 Go 代码注释
   - 验证 OpenAPI 语法

3. **性能测试异常**:
   - 调整并发参数
   - 检查系统资源
   - 分析瓶颈原因

### 调试技巧

```bash
# 详细测试输出
make test-unit ARGS="-v"

# 运行特定测试
go test -v ./test/integration -run TestHealthCheck

# 性能分析
go test -bench=. -cpuprofile=cpu.prof ./test/performance
```

## 📚 相关文档

- [OpenAPI 规范](./openapi.yaml)
- [API 使用示例](./api-examples.md)
- [构建系统文档](./build-system.md)
- [快速开始指南](./quick-start.md)