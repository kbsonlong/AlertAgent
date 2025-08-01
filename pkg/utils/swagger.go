package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"alert_agent/internal/pkg/logger"
	"go.uber.org/zap"
)

// RegenerateSwaggerDocs 重新生成Swagger文档
// 该函数会自动执行 swag init 命令重新生成API文档
// 如果swag命令不存在，只记录警告不中断程序执行
func RegenerateSwaggerDocs() error {
	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// 确定项目根目录
	// 如果当前在cmd目录下，则切换到上级目录
	projectRoot := wd
	if filepath.Base(wd) == "cmd" {
		projectRoot = filepath.Dir(wd)
	}

	// 执行 swag init 命令
	cmd := exec.Command("swag", "init", "-g", "cmd/main.go")
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.L.Info("Regenerating Swagger documentation...")
	if err := cmd.Run(); err != nil {
		// 如果swag命令不存在，记录警告但不中断启动
		logger.L.Warn("Failed to regenerate Swagger docs (swag command may not be installed)", zap.Error(err))
		return nil
	}

	logger.L.Info("Swagger documentation regenerated successfully")
	return nil
}

// CheckSwagInstallation 检查swag命令是否已安装
func CheckSwagInstallation() bool {
	_, err := exec.LookPath("swag")
	return err == nil
}

// InstallSwagIfNeeded 如果需要的话安装swag工具
func InstallSwagIfNeeded() error {
	if CheckSwagInstallation() {
		return nil
	}

	logger.L.Info("Installing swag tool...")
	cmd := exec.Command("go", "install", "github.com/swaggo/swag/cmd/swag@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install swag: %w", err)
	}

	logger.L.Info("Swag tool installed successfully")
	return nil
}