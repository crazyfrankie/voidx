package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/apitool/service"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/response"
)

type ApiToolHandler struct {
	svc *service.ApiToolService
}

func NewApiToolHandler(svc *service.ApiToolService) *ApiToolHandler {
	return &ApiToolHandler{svc: svc}
}

func (h *ApiToolHandler) RegisterRoute(r *gin.RouterGroup) {
	toolGroup := r.Group("api-tools")
	{
		toolGroup.GET("", h.GetApiToolProvidersWithPage())
		toolGroup.POST("validate-openapi-schema", h.ValidateOpenApiSchema())
		toolGroup.GET("/:provider_id", h.GetApiToolProvider())
		toolGroup.PUT("/:provider_id", h.UpdateApiToolProvider())
		toolGroup.DELETE("/:provider_id", h.DeleteApiToolProvider())
		toolGroup.POST("", h.CreateApiTool())
		toolGroup.GET("/:provider_id/tools/:tool_name", h.GetApiTool())
	}
}

func (h *ApiToolHandler) CreateApiTool() gin.HandlerFunc {
	return func(c *gin.Context) {
		var createReq req.CreateApiToolReq
		if err := c.ShouldBind(&createReq); err != nil {
			response.Error(c, errno.ErrValidate)
			return
		}

		err := h.svc.CreateApiTool(c.Request.Context(), createReq)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}

func (h *ApiToolHandler) GetApiToolProvidersWithPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var pageReq req.GetApiToolProvidersWithPageReq
		if err := c.ShouldBindQuery(&pageReq); err != nil {
			response.Error(c, errno.ErrValidate)
			return
		}

		res, err := h.svc.GetApiToolProvidersWithPage(c.Request.Context(), pageReq)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, res)
	}
}

func (h *ApiToolHandler) GetApiToolProvider() gin.HandlerFunc {
	return func(c *gin.Context) {
		providerIDStr := c.Param("provider_id")
		providerID, err := uuid.Parse(providerIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate)
			return
		}

		provider, err := h.svc.GetApiToolProvider(c.Request.Context(), providerID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, provider)
	}
}

func (h *ApiToolHandler) UpdateApiToolProvider() gin.HandlerFunc {
	return func(c *gin.Context) {
		providerIDStr := c.Param("provider_id")
		providerID, err := uuid.Parse(providerIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate)
			return
		}

		var updateReq req.UpdateApiToolProviderReq
		if err := c.ShouldBind(&updateReq); err != nil {
			response.Error(c, errno.ErrValidate)
			return
		}

		err = h.svc.UpdateApiToolProvider(c.Request.Context(), providerID, updateReq)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}

func (h *ApiToolHandler) DeleteApiToolProvider() gin.HandlerFunc {
	return func(c *gin.Context) {
		providerIDStr := c.Param("provider_id")
		providerID, err := uuid.Parse(providerIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate)
			return
		}

		err = h.svc.DeleteApiToolProvider(c.Request.Context(), providerID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}

func (h *ApiToolHandler) GetApiTool() gin.HandlerFunc {
	return func(c *gin.Context) {
		providerIdStr := c.Param("provider_id")
		toolName := c.Param("tool_name")
		providerId, err := uuid.Parse(providerIdStr)
		if err != nil {
			response.Error(c, errno.ErrValidate)
			return
		}

		tool, err := h.svc.GetApiTool(c.Request.Context(), providerId, toolName)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, tool)
	}
}

func (h *ApiToolHandler) ValidateOpenApiSchema() gin.HandlerFunc {
	return func(c *gin.Context) {
		var validateReq req.ValidateOpenApiSchemaReq
		if err := c.ShouldBind(&validateReq); err != nil {
			response.Error(c, errno.ErrValidate)
			return
		}

		err := h.svc.ValidateOpenapiSchema(c.Request.Context(), validateReq.OpenApiSchema)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}
