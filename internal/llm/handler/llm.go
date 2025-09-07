package handler

import (
	"net/http"

	"github.com/crazyfrankie/voidx/internal/base/response"
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/voidx/internal/llm/service"
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
		llmGroup.GET("/:provider/icon", h.GetProviderIcon())
		llmGroup.GET("/:provider/:model", h.GetModelEntity())
	}
}

// GetProviders 获取所有模型提供商
func (h *LLMHandler) GetProviders() gin.HandlerFunc {
	return func(c *gin.Context) {
		providers, err := h.llmService.GetProviders(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Data(c, providers)
	}
}

// GetModelEntity 获取模型实体信息
func (h *LLMHandler) GetModelEntity() gin.HandlerFunc {
	return func(c *gin.Context) {
		provider := c.Param("provider")
		modelName := c.Param("model")

		entity, err := h.llmService.GetModelEntity(c.Request.Context(), provider, modelName)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Data(c, entity)
	}
}

// GetProviderIcon 获取模型提供商图标
func (h *LLMHandler) GetProviderIcon() gin.HandlerFunc {
	return func(c *gin.Context) {
		providerName := c.Param("provider")

		iconData, mimeType, err := h.llmService.GetProviderIcon(c.Request.Context(), providerName)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		c.Header("Content-Type", mimeType)
		c.Header("Cache-Control", "public, max-age=86400")
		c.Data(http.StatusOK, mimeType, iconData)
	}
}
