# AlertAgent 数据模型测试脚本

本目录包含了为 AlertAgent 项目数据模型生成的完整测试脚本集合。

## 文件说明

### 1. `integration_test.go`
**集成测试脚本**
- 测试所有数据模型的基本功能
- 验证模型字段类型和约束
- 测试模型间的关系和依赖
- 包含常量值验证

**运行方式:**
```bash
cd /Users/zengshenglong/Code/GoWorkSpace/AlertAgent/test
go test -v integration_test.go
```

### 2. `benchmark_test.go`
**性能基准测试脚本**
- 测试模型创建性能
- JSON 序列化/反序列化性能
- 验证方法性能
- 大批量数据处理性能

**运行方式:**
```bash
cd /Users/zengshenglong/Code/GoWorkSpace/AlertAgent/test
go test -bench=. benchmark_test.go
```

### 3. `mock_data_generator.go`
**模拟数据生成器**
- 为所有模型生成真实的测试数据
- 支持批量数据生成
- 包含数据关联和依赖关系
- 可用于开发和测试环境

**使用示例:**
```go
package main

import (
    "fmt"
    "alert_agent/test"
)

func main() {
    generator := test.NewMockDataGenerator()
    
    // 生成单个模型数据
    rule := generator.GenerateRule(1)
    alert := generator.GenerateAlert(1)
    
    // 生成批量数据
    batchData := generator.GenerateBatchData(100)
    
    fmt.Printf("生成了 %d 条规则数据\n", len(batchData.Rules))
}
```

## 测试覆盖的模型

### 核心业务模型
- **Rule**: 告警规则模型
- **Alert**: 告警实例模型
- **Provider**: 数据源提供者模型
- **User**: 用户模型

### 通知系统模型
- **NotifyTemplate**: 通知模板
- **NotifyGroup**: 通知组
- **NotifyRecord**: 通知记录
- **NotificationPlugin**: 通知插件

### 任务队列模型
- **TaskQueue**: 任务队列
- **WorkerInstance**: 工作实例
- **TaskExecutionHistory**: 任务执行历史

### 配置同步模型
- **ConfigSyncStatus**: 配置同步状态
- **ConfigSyncTrigger**: 配置同步触发器
- **ConfigSyncHistory**: 配置同步历史

### 知识库模型
- **Knowledge**: 知识库条目
- **Settings**: 系统设置

## 测试特性

### 1. 数据完整性测试
- 字段类型验证
- 必填字段检查
- 数据格式验证
- 约束条件测试

### 2. 业务逻辑测试
- 模型验证方法
- JSON 序列化/反序列化
- 字段映射操作
- 状态转换逻辑

### 3. 性能测试
- 模型创建性能
- 批量操作性能
- 内存使用情况
- 并发安全性

### 4. 关系测试
- 外键关联
- 级联操作
- 数据一致性

## 运行所有测试

```bash
# 进入测试目录
cd /Users/zengshenglong/Code/GoWorkSpace/AlertAgent/test

# 运行集成测试
go test -v integration_test.go

# 运行性能测试
go test -bench=. benchmark_test.go

# 运行所有测试（如果在项目根目录）
go test ./test/...
```

## 测试数据说明

### 生成的测试数据特点
- **真实性**: 模拟真实业务场景的数据
- **多样性**: 覆盖各种边界情况和状态
- **关联性**: 保持模型间的数据关联
- **可控性**: 支持指定数据量和特定场景

### 数据生成策略
- 使用随机种子确保可重现性
- 模拟不同时间段的数据
- 包含正常和异常状态数据
- 支持自定义数据模式

## 扩展和自定义

### 添加新的测试用例
1. 在 `integration_test.go` 中添加新的测试函数
2. 遵循 `Test[ModelName][Feature]` 命名规范
3. 使用 `assert` 包进行断言

### 添加新的性能测试
1. 在 `benchmark_test.go` 中添加新的基准测试
2. 遵循 `Benchmark[Operation]` 命名规范
3. 使用 `b.ResetTimer()` 和 `b.StopTimer()` 控制计时

### 扩展数据生成器
1. 在 `mock_data_generator.go` 中添加新的生成方法
2. 保持数据的真实性和一致性
3. 支持批量生成和自定义参数

## 注意事项

1. **依赖管理**: 确保项目依赖已正确安装
2. **数据库连接**: 某些测试可能需要数据库连接
3. **并发安全**: 注意测试的并发安全性
4. **资源清理**: 测试后及时清理临时数据

## 故障排除

### 常见问题
1. **导入路径错误**: 确保在正确的目录下运行测试
2. **依赖缺失**: 运行 `go mod tidy` 安装依赖
3. **权限问题**: 确保有足够的文件系统权限

### 调试技巧
1. 使用 `-v` 参数查看详细输出
2. 使用 `-run` 参数运行特定测试
3. 使用 `-count` 参数重复运行测试

---

**生成时间**: $(date)
**项目**: AlertAgent
**版本**: 1.0.0