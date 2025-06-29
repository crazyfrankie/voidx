package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/service"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/response"
)

type ChatHandler struct {
	svc *service.ChatService
}

func NewChatHandler(svc *service.ChatService) *ChatHandler {
	return &ChatHandler{svc: svc}
}

func (h *ChatHandler) RegisterRoute(r *gin.RouterGroup) {
	chatGroup := r.Group("app")
	{
		chatGroup.POST("/:appid", h.Completion())
	}
}

func (h *ChatHandler) Completion() gin.HandlerFunc {
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

		content, err := h.svc.Chat(c.Request.Context(), chatReq.Query)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, map[string]any{
			"content": content,
		})
	}
}
