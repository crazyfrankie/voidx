package req

import "github.com/google/uuid"

// GenerateSuggestedQuestionsReq 生成建议问题列表请求
type GenerateSuggestedQuestionsReq struct {
	MessageID uuid.UUID `json:"message_id" binding:"required"`
}

// OptimizePromptReq 优化预设prompt请求
type OptimizePromptReq struct {
	Prompt string `json:"prompt" binding:"required,max=2000"`
}
