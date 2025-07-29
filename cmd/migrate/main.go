package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"alert_agent/internal/infrastructure/config"
	"alert_agent/internal/infrastructure/database"
	"alert_agent/internal/infrastructure/migration"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// 定义命令行参数
	action := flag.String("action", "migrate", "操作类型: migrate, rollback, status, validate, repair, info, cleanup, check")
	version := flag.String("version", "", "目标版本 (用于 rollback 和 repair)")
	dbHost := flag.String("db-host", getEnv("DB_HOST", "localhost"), "数据库主机")
	dbPort := flag.Int("db-port", getEnvInt("DB_PORT", 5432), "数据库端口")
	dbUser := flag.String("db-user", getEnv("DB_USER", "postgres"), "数据库用户名")
	dbPassword := flag.String("db-password", getEnv("DB_PASSWORD", ""), "数据库密码")
	dbName := flag.String("db-name", getEnv("DB_NAME", "alert_agent"), "数据库名称")
	logLevel := flag.String("log-level", "info", "日志级别: debug, info, warn, error")
	timeout := flag.Duration("timeout", 30*time.Minute, "操作超时时间")
	keepDays := flag.Int("keep-days", 30, "清理历史时保留天数")
	help := flag.Bool("help", false, "显示帮助信息")

	flag.Parse()

	if *help {
		printUsage()
		return
	}

	// 初始化日志
	logger := initLogger(*logLevel)
	defer logger.Sync()

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// 创建数据库配置
	dbConfig := &config.DatabaseConfig{
		Host:            *dbHost,
		Port:            *dbPort,
		User:            *dbUser,
		Password:        *dbPassword,
		Name:            *dbName,
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 300,
	}

	// 连接数据库
	db, err := database.NewConnection(*dbConfig)
	if err != nil {
		logger.Fatal("数据库连接失败", zap.Error(err))
	}

	// 创建迁移管理器
	manager := migration.NewManager(db, logger)

	// 执行操作
	switch *action {
	case "migrate":
		if err := manager.MigrateToLatest(ctx); err != nil {
			logger.Fatal("迁移失败", zap.Error(err))
		}
		logger.Info("迁移完成")

	case "rollback":
		if *version == "" {
			logger.Fatal("回滚操作需要指定版本")
		}
		if err := manager.RollbackToVersion(ctx, *version); err != nil {
			logger.Fatal("回滚失败", zap.Error(err))
		}
		logger.Info("回滚完成", zap.String("version", *version))

	case "status":
		printMigrationStatus(manager, logger)

	case "validate":
		if err := manager.ValidateDatabase(); err != nil {
			logger.Fatal("数据库验证失败", zap.Error(err))
		}
		logger.Info("数据库验证通过")

	case "repair":
		if *version == "" {
			logger.Fatal("修复操作需要指定版本")
		}
		if err := manager.RepairFailedMigration(*version); err != nil {
			logger.Fatal("修复失败", zap.Error(err))
		}
		logger.Info("修复完成", zap.String("version", *version))

	case "info":
		printMigrationInfo(manager, logger)

	case "cleanup":
		if err := manager.CleanupMigrationHistory(*keepDays); err != nil {
			logger.Fatal("清理失败", zap.Error(err))
		}
		logger.Info("清理完成", zap.Int("keep_days", *keepDays))

	case "check":
		isUpToDate, err := manager.IsUpToDate()
		if err != nil {
			logger.Fatal("检查失败", zap.Error(err))
		}
		if isUpToDate {
			logger.Info("数据库已是最新版本")
			os.Exit(0)
		} else {
			logger.Info("数据库需要更新")
			os.Exit(1)
		}

	default:
		logger.Fatal("未知操作", zap.String("action", *action))
	}
}

// initLogger 初始化日志
func initLogger(level string) *zap.Logger {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("初始化日志失败: %v", err))
	}

	return logger
}

// printUsage 打印使用说明
func printUsage() {
	fmt.Println(`AlertAgent 数据库迁移工具

用法:
  migrate [选项]

操作类型:
  migrate   - 迁移到最新版本 (默认)
  rollback  - 回滚到指定版本 (需要 -version 参数)
  status    - 显示迁移状态
  validate  - 验证数据库状态
  repair    - 修复失败的迁移 (需要 -version 参数)
  info      - 显示详细迁移信息
  cleanup   - 清理迁移历史 (使用 -keep-days 参数)
  check     - 检查数据库是否为最新版本

选项:
  -action string      操作类型 (默认: migrate)
  -version string     目标版本 (用于 rollback 和 repair)
  -db-host string     数据库主机 (默认: localhost)
  -db-port int        数据库端口 (默认: 5432)
  -db-user string     数据库用户名 (默认: postgres)
  -db-password string 数据库密码
  -db-name string     数据库名称 (默认: alert_agent)
  -log-level string   日志级别 (默认: info)
  -timeout duration   操作超时时间 (默认: 30m)
  -keep-days int      清理历史时保留天数 (默认: 30)
  -help              显示此帮助信息

示例:
  # 迁移到最新版本
  migrate -action=migrate -db-password=secret

  # 回滚到指定版本
  migrate -action=rollback -version=v2.0.0-010 -db-password=secret

  # 查看迁移状态
  migrate -action=status -db-password=secret

  # 验证数据库
  migrate -action=validate -db-password=secret

  # 修复失败的迁移
  migrate -action=repair -version=v2.0.0-005 -db-password=secret

  # 清理30天前的迁移历史
  migrate -action=cleanup -keep-days=30 -db-password=secret

环境变量:
  DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME 可用于设置数据库连接参数`)
}

// printMigrationStatus 打印迁移状态
func printMigrationStatus(manager *migration.Manager, logger *zap.Logger) {
	status, err := manager.GetMigrationStatus()
	if err != nil {
		logger.Fatal("获取迁移状态失败", zap.Error(err))
	}

	if len(status) == 0 {
		fmt.Println("没有找到迁移记录")
		return
	}

	fmt.Printf("%-20s %-40s %-10s %-20s %-10s\n", "版本", "名称", "状态", "执行时间", "耗时(ms)")
	fmt.Println("--------------------------------------------------------------------------------------------")

	for _, migration := range status {
		status := "✗"
		if migration.Success {
			status = "✓"
		}

		executedAt := migration.ExecutedAt.Format("2006-01-02 15:04:05")

		errorMsg := ""
		if migration.ErrorMsg != "" {
			errorMsg = fmt.Sprintf(" (Error: %s)", migration.ErrorMsg)
		}

		fmt.Printf("%-20s %-40s %-10s %-20s %-10d%s\n",
			migration.Version,
			migration.Name,
			status,
			executedAt,
			migration.Duration,
			errorMsg)
	}
}

// printMigrationInfo 打印详细迁移信息
func printMigrationInfo(manager *migration.Manager, logger *zap.Logger) {
	status, err := manager.GetMigrationStatus()
	if err != nil {
		logger.Fatal("获取迁移信息失败", zap.Error(err))
	}

	if len(status) == 0 {
		fmt.Println("没有找到迁移记录")
		return
	}

	fmt.Println("=== 迁移详细信息 ===")
	for _, migration := range status {
		fmt.Printf("\n版本: %s\n", migration.Version)
		fmt.Printf("名称: %s\n", migration.Name)
		fmt.Printf("描述: %s\n", migration.Description)
		fmt.Printf("状态: %s\n", func() string {
			if migration.Success {
				return "成功"
			}
			return "失败"
		}())
		fmt.Printf("执行时间: %s\n", migration.ExecutedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("耗时: %d ms\n", migration.Duration)
		fmt.Printf("校验和: %s\n", migration.Checksum)
		if migration.ErrorMsg != "" {
			fmt.Printf("错误信息: %s\n", migration.ErrorMsg)
		}
		fmt.Println("---")
	}
}

// getEnv 获取环境变量字符串值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取环境变量整数值
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}