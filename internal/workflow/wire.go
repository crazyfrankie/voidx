//go:build wireinject
// +build wireinject

package workflow

import (
	"github.com/google/wire"
	"gorm.io/gorm"

	builtin "github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers"
	"github.com/crazyfrankie/voidx/internal/workflow/handler"
	"github.com/crazyfrankie/voidx/internal/workflow/repository"
	"github.com/crazyfrankie/voidx/internal/workflow/repository/dao"
	"github.com/crazyfrankie/voidx/internal/workflow/service"
)

type Handler = handler.WorkflowHandler
type Service = service.WorkflowService

type WorkflowModule struct {
	Handler *Handler
	Service *Service
}

var ProviderSet = wire.NewSet(
	dao.NewWorkflowDao,
	repository.NewWorkflowRepo,
	service.NewWorkflowService,
	handler.NewWorkflowHandler,
)

func InitWorkflowModule(db *gorm.DB, builtinProvider *builtin.BuiltinProviderManager) *WorkflowModule {
	wire.Build(
		ProviderSet,

		wire.Struct(new(WorkflowModule), "*"),
	)
	return new(WorkflowModule)
}
