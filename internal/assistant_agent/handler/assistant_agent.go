package handler

import (
	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"io"

	"github.com/crazyfrankie/voidx/internal/assistant_agent/service"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/response"
	"github.com/crazyfrankie/voidx/pkg/util"
)

type AssistantAgentHandler struct {
	svc *service.AssistantAgentService
}

func NewAssistantAgentHandler(svc *service.AssistantAgentService) *AssistantAgentHandler {
	return &AssistantAgentHandler{svc: svc}
}

func (h *AssistantAgentHandler) RegisterRoute(r *gin.RouterGroup) {
	assistantGroup := r.Group("assistant-agent")
	{
		assistantGroup.POST("chat", h.Chat())
		assistantGroup.POST("chat/:task_id/stop", h.StopChat())
		assistantGroup.GET("messages", h.GetMessagesWithPage())
		assistantGroup.DELETE("conversation", h.DeleteConversation())
	}
}

// Chat 与辅助智能体进行对话聊天
func (h *AssistantAgentHandler) Chat() gin.HandlerFunc {
	return func(c *gin.Context) {
		var chatReq req.AssistantAgentChatReq
		if err := c.ShouldBindJSON(&chatReq); err != nil {
			response.Error(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		// 获取流式响应
		eventChan, err := h.svc.Chat(c.Request.Context(), userID, chatReq)
		if err != nil {
			response.Error(c, err)
			return
		}

		// 流式输出
		c.Stream(func(w io.Writer) bool {
			select {
			case event, ok := <-eventChan:
				if !ok {
					return false
				}

				eventData, _ := sonic.Marshal(event)
				c.SSEvent("message", eventData)
				return true
			case <-c.Request.Context().Done():
				return false
			}
		})
	}
}

// StopChat 停止与辅助智能体的对话聊天
func (h *AssistantAgentHandler) StopChat() gin.HandlerFunc {
	return func(c *gin.Context) {
		var stopReq req.StopAssistantAgentChatReq
		if err := c.ShouldBindUri(&stopReq); err != nil {
			response.Error(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		err = h.svc.StopChat(c.Request.Context(), stopReq.TaskID, userID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}

// GetMessagesWithPage 获取辅助智能体消息分页列表
func (h *AssistantAgentHandler) GetMessagesWithPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var pageReq req.GetAssistantAgentMessagesWithPageReq
		if err := c.ShouldBindQuery(&pageReq); err != nil {
			response.Error(c, errno.ErrValidate)
			return
		}

		// 设置默认值
		if pageReq.CurrentPage == 0 {
			pageReq.Ctime = 1
		}
		if pageReq.PageSize == 0 {
			pageReq.PageSize = 20
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		messages, agentsThoughts, paginator, err := h.svc.GetMessagesWithPage(c.Request.Context(), userID, pageReq)
		if err != nil {
			response.Error(c, err)
			return
		}

		// 转换为响应格式
		messageResps := make([]resp.AssistantAgentMessageResp, len(messages))
		for i, msg := range messages {
			thoughtResps := make([]resp.AssistantAgentThoughtResp, len(agentsThoughts))
			for j, thought := range agentsThoughts {
				thoughtResps[j] = resp.AssistantAgentThoughtResp{
					ID:          thought.ID,
					Event:       thought.Event,
					Thought:     thought.Thought,
					Observation: thought.Observation,
					Tool:        thought.Tool,
					ToolInput:   thought.ToolInput,
					Latency:     thought.Latency,
					Ctime:       thought.Ctime,
				}
			}

			messageResps[i] = resp.AssistantAgentMessageResp{
				ID:              msg.ID,
				ConversationID:  msg.ConversationID,
				Query:           msg.Query,
				Answer:          msg.Answer,
				TotalTokenCount: msg.TotalTokenCount,
				Latency:         msg.Latency,
				ImageUrls:       msg.ImageUrls,
				AgentThoughts:   thoughtResps,
				Ctime:           msg.Ctime,
			}
		}

		result := map[string]any{
			"list":      messageResps,
			"paginator": paginator,
		}

		response.SuccessWithData(c, result)
	}
}

// DeleteConversation 清空辅助Agent智能体会话消息列表
func (h *AssistantAgentHandler) DeleteConversation() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		err = h.svc.DeleteConversation(c.Request.Context(), userID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}
