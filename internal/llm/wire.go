//go:build wireinject

package llm

import (
	"github.com/google/wire"

	"github.com/crazyfrankie/voidx/internal/core/llm"
	"github.com/crazyfrankie/voidx/internal/llm/handler"
	"github.com/crazyfrankie/voidx/internal/llm/service"
)

type Handler = handler.LLMHandler
type Service = service.LLMService
type LLMModule struct {
	Handler *Handler
	Service *Service
}

func InitLLMModule(llmCore *llm.LanguageModelManager) *LLMModule {
	wire.Build(
		service.NewLLMService,
		handler.NewLLMHandler,

		wire.Struct(new(LLMModule), "*"),
	)
	return new(LLMModule)
}
