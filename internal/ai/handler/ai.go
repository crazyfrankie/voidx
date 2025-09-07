package handler

import (
	"io"

	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/voidx/internal/ai/service"
	"github.com/crazyfrankie/voidx/internal/base/response"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/pkg/util"
	"github.com/crazyfrankie/voidx/types/errno"
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
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		// 获取流式响应
		eventChan, err := h.svc.OptimizePrompt(c.Request.Context(), optimizeReq.Prompt)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		// 流式输出
		c.Stream(func(w io.Writer) bool {
			select {
			case event, ok := <-eventChan:
				if !ok {
					return false
				}
				c.SSEvent("message", event)
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
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		// 调用服务生成建议问题列表
		suggestedQuestions, err := h.svc.GenerateSuggestedQuestions(c.Request.Context(), suggestedReq.MessageID, userID)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Data(c, suggestedQuestions)
	}
}
