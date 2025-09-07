package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/base/response"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/platform/service"
	"github.com/crazyfrankie/voidx/pkg/util"
	"github.com/crazyfrankie/voidx/types/errno"
)

type PlatformHandler struct {
	svc *service.PlatformService
}

func NewPlatformHandler(svc *service.PlatformService) *PlatformHandler {
	return &PlatformHandler{svc: svc}
}

func (h *PlatformHandler) RegisterRoute(r *gin.RouterGroup) {
	platformGroup := r.Group("platform")
	{
		platformGroup.GET("apps/:app_id/wechat-config", h.GetWechatConfig())
		platformGroup.PUT("apps/:app_id/wechat-config", h.UpdateWechatConfig())
	}
}

// GetWechatConfig 根据传递的id获取指定应用的微信配置
func (h *PlatformHandler) GetWechatConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取路径参数中的应用ID
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		// 获取当前用户ID
		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		// 调用服务获取微信配置
		wechatConfig, err := h.svc.GetWechatConfig(c.Request.Context(), appID, userID)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Data(c, wechatConfig)
	}
}

// UpdateWechatConfig 根据传递的应用id更新该应用的微信发布配置
func (h *PlatformHandler) UpdateWechatConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取路径参数中的应用ID
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		// 绑定请求参数
		var updateReq req.UpdateWechatConfigReq
		if err := c.ShouldBindJSON(&updateReq); err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		// 获取当前用户ID
		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		// 调用服务更新微信配置
		err = h.svc.UpdateWechatConfig(c.Request.Context(), appID, userID, updateReq)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Success(c)
	}
}
