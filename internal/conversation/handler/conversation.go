package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/conversation/service"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/response"
)

type ConversationHandler struct {
	svc *service.ConversationService
}

func NewConversationHandler(svc *service.ConversationService) *ConversationHandler {
	return &ConversationHandler{svc: svc}
}

func (h *ConversationHandler) RegisterRoute(r *gin.RouterGroup) {
	conversationGroup := r.Group("conversations")
	{
		conversationGroup.GET("/:conversation_id/messages", h.GetConversationMessagesWithPage())
		conversationGroup.DELETE("/:conversation_id", h.DeleteConversation())
		conversationGroup.DELETE("/:conversation_id/messages/:message_id", h.DeleteMessage())
		conversationGroup.GET("/:conversation_id/name", h.GetConversationName())
		conversationGroup.PUT("/:conversation_id/name", h.UpdateConversationName())
		conversationGroup.PUT("/:conversation_id/is-pinned", h.UpdateConversationIsPinned())
	}
}

func (h *ConversationHandler) GetConversationMessagesWithPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationIDStr := c.Param("conversation_id")
		conversationID, err := uuid.Parse(conversationIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate)
			return
		}

		var pageReq req.GetConversationMessagesWithPageReq
		if err := c.ShouldBindQuery(&pageReq); err != nil {
			response.Error(c, errno.ErrValidate)
			return
		}

		messages, paginator, err := h.svc.GetConversationMessagesWithPage(c.Request.Context(), conversationID, pageReq)
		if err != nil {
			response.Error(c, err)
			return
		}

		res := make([]resp.GetConversationMessagePageResp, 0, len(messages))
		for _, message := range messages {
			dbAgentThoughts, err := h.svc.GetConversationAgentThoughts(c.Request.Context(), conversationID)
			if err != nil {
				continue
			}
			agentThoughts := make([]resp.GetConversationMessagePageAgentThought, 0, len(dbAgentThoughts))
			for _, at := range dbAgentThoughts {
				agentThoughts = append(agentThoughts, resp.GetConversationMessagePageAgentThought{
					ID:          at.ID,
					Position:    at.Position,
					Event:       at.Event,
					Thought:     at.Thought,
					Observation: at.Observation,
					Tool:        at.Tool,
					ToolInput:   at.ToolInput,
					Latency:     int(at.Latency),
					Ctime:       at.Ctime,
				})
			}
			res = append(res, resp.GetConversationMessagePageResp{
				ID:              message.ID,
				ConversationID:  message.ConversationID,
				Query:           message.Query,
				ImageUrls:       message.ImageUrls,
				Answer:          message.Answer,
				Latency:         message.Latency,
				TotalTokenCount: message.TotalTokenCount,
				AgentThoughts:   agentThoughts,
				Ctime:           message.Ctime,
			})

		}

		response.SuccessWithData(c, gin.H{
			"list":      res,
			"paginator": paginator,
		})
	}
}

func (h *ConversationHandler) DeleteConversation() gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationIDStr := c.Param("conversation_id")
		conversationID, err := uuid.Parse(conversationIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate)
			return
		}

		err = h.svc.DeleteConversation(c.Request.Context(), conversationID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}

func (h *ConversationHandler) DeleteMessage() gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationIDStr := c.Param("conversation_id")
		conversationID, err := uuid.Parse(conversationIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate)
			return
		}

		messageIDStr := c.Param("message_id")
		messageID, err := uuid.Parse(messageIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate)
			return
		}

		err = h.svc.DeleteMessage(c.Request.Context(), conversationID, messageID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}

func (h *ConversationHandler) GetConversationName() gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationIDStr := c.Param("conversation_id")
		conversationID, err := uuid.Parse(conversationIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate)
			return
		}

		name, err := h.svc.GetConversationName(c.Request.Context(), conversationID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, gin.H{"name": name})
	}
}

func (h *ConversationHandler) UpdateConversationName() gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationIDStr := c.Param("conversation_id")
		conversationID, err := uuid.Parse(conversationIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate)
			return
		}

		var updateReq req.UpdateConversationNameReq
		if err := c.ShouldBind(&updateReq); err != nil {
			response.Error(c, errno.ErrValidate)
			return
		}

		err = h.svc.UpdateConversationName(c.Request.Context(), conversationID, updateReq.Name)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}

func (h *ConversationHandler) UpdateConversationIsPinned() gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationIDStr := c.Param("conversation_id")
		conversationID, err := uuid.Parse(conversationIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate)
			return
		}

		var updateReq req.UpdateConversationIsPinnedReq
		if err := c.ShouldBind(&updateReq); err != nil {
			response.Error(c, errno.ErrValidate)
			return
		}

		err = h.svc.UpdateConversationIsPinned(c.Request.Context(), conversationID, updateReq.IsPinned)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}
