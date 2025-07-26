package handler

import (
	"net/http"
	
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/voidx/internal/builtin_tools/service"
	"github.com/crazyfrankie/voidx/pkg/response"
)

type BuiltinToolsHandler struct {
	svc *service.BuiltinToolsService
}

func NewBuiltinToolsHandler(svc *service.BuiltinToolsService) *BuiltinToolsHandler {
	return &BuiltinToolsHandler{svc: svc}
}

func (h *BuiltinToolsHandler) RegisterRoute(r *gin.RouterGroup) {
	builtinToolsGroup := r.Group("builtin-tools")
	{
		builtinToolsGroup.GET("", h.GetBuiltinTools())
		builtinToolsGroup.GET("/:provider_name/tools/:tool_name", h.GetProviderTool())
		builtinToolsGroup.GET("/:provider_name/icon", h.GetProviderIcon())
		builtinToolsGroup.GET("categories", h.GetCategories())
	}
}

func (h *BuiltinToolsHandler) GetBuiltinTools() gin.HandlerFunc {
	return func(c *gin.Context) {
		res := h.svc.GetBuiltinTools(c.Request.Context())

		response.SuccessWithData(c, res)
	}
}

func (h *BuiltinToolsHandler) GetProviderTool() gin.HandlerFunc {
	return func(c *gin.Context) {
		providerName := c.Param("provider_name")
		toolName := c.Param("tool_name")

		res, err := h.svc.GetProviderTool(c.Request.Context(), providerName, toolName)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, res)
	}
}

func (h *BuiltinToolsHandler) GetProviderIcon() gin.HandlerFunc {
	return func(c *gin.Context) {
		providerName := c.Param("provider_name")

		iconData, mimeType, err := h.svc.GetProviderIcon(c.Request.Context(), providerName)
		if err != nil {
			response.Error(c, err)
			return
		}

		c.Header("Content-Type", mimeType)
		c.Header("Cache-Control", "public, max-age=86400")
		c.Data(http.StatusOK, mimeType, iconData)
	}
}

func (h *BuiltinToolsHandler) GetCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		res := h.svc.GetCategories(c.Request.Context())

		response.SuccessWithData(c, res)
	}
}
