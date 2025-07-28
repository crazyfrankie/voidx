package handler

import (
	"io"

	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/webapp/service"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/response"
	"github.com/crazyfrankie/voidx/pkg/util"
)

type WebAppHandler struct {
	svc *service.WebAppService
}

func NewWebAppHandler(svc *service.WebAppService) *WebAppHandler {
	return &WebAppHandler{svc: svc}
}

func (h *WebAppHandler) RegisterRoute(r *gin.RouterGroup) {
	webappGroup := r.Group("webapp")
	{
		webappGroup.GET("/:token/info", h.GetWebAppInfo())
		webappGroup.POST("/:token/chat", h.WebAppChat())
		webappGroup.POST("/:token/chat/:task_id/stop", h.StopWebAppChat())
		webappGroup.GET("/:token/conversations", h.GetConversations())
		webappGroup.GET("/:token/conversations/:conversation_id/messages", h.GetConversationMessages())
		webappGroup.DELETE("/:token/conversations/:conversation_id", h.DeleteConversation())
		webappGroup.PUT("/:token/conversations/:conversation_id/name", h.UpdateConversationName())
		webappGroup.PUT("/:token/conversations/:conversation_id/pin", h.UpdateConversationPin())
	}
}

func (h *WebAppHandler) GetWebAppInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		if token == "" {
			response.Error(c, errno.ErrValidate.AppendBizMessage("token不能为空"))
			return
		}

		info, err := h.svc.GetWebAppInfo(c.Request.Context(), token)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, info)
	}
}

func (h *WebAppHandler) WebAppChat() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		if token == "" {
			response.Error(c, errno.ErrValidate.AppendBizMessage("token不能为空"))
			return
		}

		var chatReq req.WebAppChatReq
		if err := c.ShouldBind(&chatReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		// 设置SSE响应头
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Access-Control-Allow-Origin", "*")

		// 获取响应流
		responseStream, err := h.svc.WebAppChat(c.Request.Context(), token, chatReq, userID)
		if err != nil {
			response.Error(c, err)
			return
		}

		// 流式输出
		c.Stream(func(w io.Writer) bool {
			select {
			case data, ok := <-responseStream:
				if !ok {
					return false
				}
				c.SSEvent("message", data)
				return true
			case <-c.Request.Context().Done():
				return false
			}
		})
	}
}

func (h *WebAppHandler) StopWebAppChat() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		if token == "" {
			response.Error(c, errno.ErrValidate.AppendBizMessage("token不能为空"))
			return
		}

		taskID := c.Param("task_id")
		if taskID == "" {
			response.Error(c, errno.ErrValidate.AppendBizMessage("task_id不能为空"))
			return
		}

		err := h.svc.StopWebAppChat(c.Request.Context(), token, taskID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}

func (h *WebAppHandler) GetConversations() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		if token == "" {
			response.Error(c, errno.ErrValidate.AppendBizMessage("token不能为空"))
			return
		}

		var getReq req.GetWebAppConversationsReq
		if err := c.ShouldBindQuery(&getReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		conversations, err := h.svc.GetConversations(c.Request.Context(), token, getReq)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, conversations)
	}
}

func (h *WebAppHandler) GetConversationMessages() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		if token == "" {
			response.Error(c, errno.ErrValidate.AppendBizMessage("token不能为空"))
			return
		}

		conversationID := c.Param("conversation_id")
		if conversationID == "" {
			response.Error(c, errno.ErrValidate.AppendBizMessage("conversation_id不能为空"))
			return
		}

		var pageReq req.GetWebAppConversationMessagesReq
		if err := c.ShouldBindQuery(&pageReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		messages, paginator, err := h.svc.GetConversationMessages(c.Request.Context(), token, conversationID, pageReq)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, gin.H{
			"list":      messages,
			"paginator": paginator,
		})
	}
}

func (h *WebAppHandler) DeleteConversation() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		if token == "" {
			response.Error(c, errno.ErrValidate.AppendBizMessage("token不能为空"))
			return
		}

		conversationID := c.Param("conversation_id")
		if conversationID == "" {
			response.Error(c, errno.ErrValidate.AppendBizMessage("conversation_id不能为空"))
			return
		}

		err := h.svc.DeleteConversation(c.Request.Context(), token, conversationID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}

func (h *WebAppHandler) UpdateConversationName() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		if token == "" {
			response.Error(c, errno.ErrValidate.AppendBizMessage("token不能为空"))
			return
		}

		conversationID := c.Param("conversation_id")
		if conversationID == "" {
			response.Error(c, errno.ErrValidate.AppendBizMessage("conversation_id不能为空"))
			return
		}

		var updateReq req.UpdateWebAppConversationNameReq
		if err := c.ShouldBind(&updateReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		err := h.svc.UpdateConversationName(c.Request.Context(), token, conversationID, updateReq.Name)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}

func (h *WebAppHandler) UpdateConversationPin() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		if token == "" {
			response.Error(c, errno.ErrValidate.AppendBizMessage("token不能为空"))
			return
		}

		conversationID := c.Param("conversation_id")
		if conversationID == "" {
			response.Error(c, errno.ErrValidate.AppendBizMessage("conversation_id不能为空"))
			return
		}

		var updateReq req.UpdateWebAppConversationPinReq
		if err := c.ShouldBind(&updateReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		err := h.svc.UpdateConversationPin(c.Request.Context(), token, conversationID, updateReq.IsPinned)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}
