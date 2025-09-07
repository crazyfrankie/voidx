package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/api_key/service"
	"github.com/crazyfrankie/voidx/internal/base/response"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/pkg/util"
	"github.com/crazyfrankie/voidx/types/errno"
)

type ApiKeyHandler struct {
	svc *service.ApiKeyService
}

func NewApiKeyHandler(svc *service.ApiKeyService) *ApiKeyHandler {
	return &ApiKeyHandler{svc: svc}
}

func (h *ApiKeyHandler) RegisterRoute(r *gin.RouterGroup) {
	apiKeyGroup := r.Group("api-keys")
	{
		apiKeyGroup.POST("", h.CreateApiKey())
		apiKeyGroup.GET("", h.GetApiKeysWithPage())
		apiKeyGroup.PUT("/:api_key_id", h.UpdateApiKey())
		apiKeyGroup.PUT("/:api_key_id/active", h.UpdateApiKeyIsActive())
		apiKeyGroup.DELETE("/:api_key_id", h.DeleteApiKey())
	}
}

// CreateApiKey 创建API秘钥
func (h *ApiKeyHandler) CreateApiKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		var createReq req.CreateApiKeyReq
		if err := c.ShouldBindJSON(&createReq); err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		err = h.svc.CreateApiKey(c.Request.Context(), userID, createReq)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Success(c)
	}
}

// GetApiKeysWithPage 获取当前登录账号的API秘钥分页列表信息
func (h *ApiKeyHandler) GetApiKeysWithPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var pageReq req.GetApiKeysWithPageReq
		if err := c.ShouldBindQuery(&pageReq); err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		// 设置默认值
		if pageReq.CurrentPage == 0 {
			pageReq.CurrentPage = 1
		}
		if pageReq.PageSize == 0 {
			pageReq.PageSize = 20
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		apiKeys, paginator, err := h.svc.GetApiKeysWithPage(c.Request.Context(), userID, pageReq)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		result := map[string]any{
			"list":      apiKeys,
			"paginator": paginator,
		}

		response.Data(c, result)
	}
}

// UpdateApiKey 根据传递的信息更新API秘钥
func (h *ApiKeyHandler) UpdateApiKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKeyIDStr := c.Param("api_key_id")
		apiKeyID, err := uuid.Parse(apiKeyIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		var updateReq req.UpdateApiKeyReq
		if err := c.ShouldBindJSON(&updateReq); err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		err = h.svc.UpdateApiKey(c.Request.Context(), apiKeyID, userID, updateReq)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Success(c)
	}
}

// UpdateApiKeyIsActive 根据传递的信息更新API秘钥激活状态
func (h *ApiKeyHandler) UpdateApiKeyIsActive() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKeyIDStr := c.Param("api_key_id")
		apiKeyID, err := uuid.Parse(apiKeyIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		var activeReq req.UpdateApiKeyIsActiveReq
		if err := c.ShouldBindJSON(&activeReq); err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		err = h.svc.UpdateApiKeyIsActive(c.Request.Context(), apiKeyID, userID, activeReq.IsActive)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Success(c)
	}
}

// DeleteApiKey 根据传递的id删除API秘钥
func (h *ApiKeyHandler) DeleteApiKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKeyIDStr := c.Param("api_key_id")
		apiKeyID, err := uuid.Parse(apiKeyIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		err = h.svc.DeleteApiKey(c.Request.Context(), apiKeyID, userID)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Success(c)
	}
}
