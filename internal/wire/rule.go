//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"gorm.io/gorm"
	apprule "alert_agent/internal/application/rule"
	domainrule "alert_agent/internal/domain/rule"
	"alert_agent/internal/handler"
	"alert_agent/internal/infrastructure/repository"
)

// RuleSet 规则模块的依赖注入集合
var RuleSet = wire.NewSet(
	repository.NewRuleRepository,
	wire.Bind(new(domainrule.Repository), new(domainrule.Repository)),
	apprule.NewService,
	wire.Bind(new(domainrule.Service), new(*apprule.Service)),
	handler.NewRuleHandler,
)

// InitializeRuleHandler 初始化规则处理器
func InitializeRuleHandler(db *gorm.DB) (*handler.RuleHandler, error) {
	wire.Build(RuleSet)
	return nil, nil
}