package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/service"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/response"
)

type AppHandler struct {
	svc *service.AppService
}

func NewAppHandler(svc *service.AppService) *AppHandler {
	return &AppHandler{svc: svc}
}

func (h *AppHandler) RegisterRoute(r *gin.RouterGroup) {
	appGroup := r.Group("apps")
	{
		appGroup.GET("/:appid", h.GetApplicationConf())
		appGroup.POST("/:appid/config", h.UpdateAppDraftConfig())
		appGroup.GET("/:appid/long-term-memory", h.GetAppDebugLongMemory())
		appGroup.POST("/:appid/long-term-memory", h.UpdateAppDebugLongMemory())
		appGroup.POST("/:appid", h.AppDebugChat())
		appGroup.GET("/:appid/messages", h.GetAppDebugHistoryList())
		appGroup.DELETE("/:appid/messages/:messageId/delete", h.DeleteDebugMessage())
	}
}

func (h *AppHandler) GetApplicationConf() gin.HandlerFunc {
	return func(c *gin.Context) {
		appid := c.Param("appid")

		resp, err := h.svc.GetAppConfig(c.Request.Context(), appid)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, resp)
	}
}

func (h *AppHandler) UpdateAppDraftConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO
	}
}

func (h *AppHandler) GetAppDebugLongMemory() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO
	}
}

func (h *AppHandler) UpdateAppDebugLongMemory() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO
	}
}

func (h *AppHandler) AppDebugChat() gin.HandlerFunc {
	return func(c *gin.Context) {
		var chatReq req.ChatReq
		if err := c.ShouldBind(&chatReq); err != nil {
			response.Error(c, errno.Success)
			return
		}

		appID := c.Param("appid")
		if _, err := uuid.Parse(appID); err != nil {
			response.Error(c, errno.ErrValidate)
			return
		}

		resp, err := h.svc.AppDebugChat(c.Request.Context(), chatReq.Query)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, resp)
	}
}

func (h *AppHandler) GetAppDebugHistoryList() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO
	}
}

func (h *AppHandler) DeleteDebugMessage() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO
	}
}
