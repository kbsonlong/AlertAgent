//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"gorm.io/gorm"
	"alert_agent/internal/application/rule"
	"alert_agent/internal/domain/rule"
	"alert_agent/internal/handler"
	"alert_agent/internal/infrastructure/repository"
)

// RuleSet 规则模块的依赖注入集合
var RuleSet = wire.NewSet(
	repository.NewRuleRepository,
	wire.Bind(new(rule.Repository), new(*repository.RuleRepository)),
	rule.NewService,
	wire.Bind(new(rule.Service), new(*rule.Service)),
	handler.NewRuleHandler,
)

// InitializeRuleHandler 初始化规则处理器
func InitializeRuleHandler(db *gorm.DB) (*handler.RuleHandler, error) {
	wire.Build(RuleSet)
	return nil, nil
}