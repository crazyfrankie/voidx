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
	// API工具提供商路由
	providerGroup := r.Group("api-tool-providers")
	{
		providerGroup.GET("", h.GetApiToolProvidersWithPage())
		providerGroup.GET("/:provider_id", h.GetApiToolProvider())
		providerGroup.PUT("/:provider_id", h.UpdateApiToolProvider())
		providerGroup.DELETE("/:provider_id", h.DeleteApiToolProvider())
	}

	// API工具路由
	toolGroup := r.Group("api-tools")
	{
		toolGroup.POST("", h.CreateApiTool())
		toolGroup.GET("/:tool_id", h.GetApiTool())
	}
}

func (h *ApiToolHandler) CreateApiTool() gin.HandlerFunc {
	return func(c *gin.Context) {
		var createReq req.CreateApiToolReq
		if err := c.ShouldBind(&createReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
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
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
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
			response.Error(c, errno.ErrValidate.AppendBizMessage("提供商ID格式错误"))
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
			response.Error(c, errno.ErrValidate.AppendBizMessage("提供商ID格式错误"))
			return
		}

		var updateReq req.UpdateApiToolProviderReq
		if err := c.ShouldBind(&updateReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
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
			response.Error(c, errno.ErrValidate.AppendBizMessage("提供商ID格式错误"))
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
		toolIDStr := c.Param("tool_id")
		toolID, err := uuid.Parse(toolIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("工具ID格式错误"))
			return
		}

		tool, err := h.svc.GetApiTool(c.Request.Context(), toolID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, tool)
	}
}
