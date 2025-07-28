package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/voidx/internal/ai/service"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/response"
	"github.com/crazyfrankie/voidx/pkg/util"
)

type AIHandler struct {
	svc *service.AIService
}

func NewAIHandler(svc *service.AIService) *AIHandler {
	return &AIHandler{svc: svc}
}

func (h *AIHandler) RegisterRoute(r *gin.RouterGroup) {
	aiGroup := r.Group("ai")
	{
		aiGroup.POST("optimize-prompt", h.OptimizePrompt())
		aiGroup.POST("suggested-questions", h.GenerateSuggestedQuestions())
	}
}

// OptimizePrompt 根据传递的预设prompt进行优化
func (h *AIHandler) OptimizePrompt() gin.HandlerFunc {
	return func(c *gin.Context) {
		var optimizeReq req.OptimizePromptReq
		if err := c.ShouldBindJSON(&optimizeReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		// 设置SSE响应头
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Access-Control-Allow-Origin", "*")

		// 获取流式响应
		eventChan, err := h.svc.OptimizePrompt(c.Request.Context(), optimizeReq.Prompt)
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
				fmt.Fprintf(w, "event: optimize_prompt\ndata: %s\n\n", string(eventData))

				// 刷新缓冲区
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				return true
			case <-c.Request.Context().Done():
				return false
			}
		})
	}
}

// GenerateSuggestedQuestions 根据传递的消息id生成建议问题列表
func (h *AIHandler) GenerateSuggestedQuestions() gin.HandlerFunc {
	return func(c *gin.Context) {
		var suggestedReq req.GenerateSuggestedQuestionsReq
		if err := c.ShouldBindJSON(&suggestedReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		// 调用服务生成建议问题列表
		suggestedQuestions, err := h.svc.GenerateSuggestedQuestions(c.Request.Context(), suggestedReq.MessageID, userID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, suggestedQuestions)
	}
}
