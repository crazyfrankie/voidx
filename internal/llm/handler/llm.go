package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/voidx/internal/llm/service"
	"github.com/crazyfrankie/voidx/pkg/response"
)

type LLMHandler struct {
	llmService *service.LLMService
}

func NewLLMHandler(llmService *service.LLMService) *LLMHandler {
	return &LLMHandler{
		llmService: llmService,
	}
}

func (h *LLMHandler) RegisterRoute(r *gin.RouterGroup) {
	llmGroup := r.Group("llms")
	{
		llmGroup.GET("", h.GetProviders())
		llmGroup.GET("/:provider/:model", h.GetModelEntity())
	}
}

// GetProviders 获取所有模型提供商
func (h *LLMHandler) GetProviders() gin.HandlerFunc {
	return func(c *gin.Context) {
		providers, err := h.llmService.GetProviders(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, providers)
	}
}

// GetModelEntity 获取模型实体信息
func (h *LLMHandler) GetModelEntity() gin.HandlerFunc {
	return func(c *gin.Context) {
		provider := c.Param("provider")
		modelName := c.Param("model")

		entity, err := h.llmService.GetModelEntity(c.Request.Context(), provider, modelName)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, entity)
	}
}
