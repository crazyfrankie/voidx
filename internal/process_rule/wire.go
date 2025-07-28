//go:build wireinject

package process_rule

import (
	"github.com/crazyfrankie/voidx/internal/process_rule/repository"
	"github.com/crazyfrankie/voidx/internal/process_rule/repository/dao"
	"github.com/crazyfrankie/voidx/internal/process_rule/service"
	"github.com/google/wire"
	"gorm.io/gorm"
)

type Service = service.ProcessRuleService

type ProcessRuleModule struct {
	Service *Service
}

func InitProcessRuleModule(db *gorm.DB) *ProcessRuleModule {
	wire.Build(
		dao.NewProcessRuleDao,
		repository.NewProcessRuleRepo,
		service.NewProcessRuleService,

		wire.Struct(new(ProcessRuleModule), "*"),
	)
	return new(ProcessRuleModule)
}
