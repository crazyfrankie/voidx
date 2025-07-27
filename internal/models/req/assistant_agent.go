package req

import "github.com/google/uuid"

// AssistantAgentChatReq 辅助Agent会话请求
type AssistantAgentChatReq struct {
	Query     string   `json:"query" binding:"required" validate:"max=2000"`
	ImageUrls []string `json:"image_urls" validate:"max=5,dive,url"`
}

// GetAssistantAgentMessagesWithPageReq 获取辅助智能体消息列表分页请求
type GetAssistantAgentMessagesWithPageReq struct {
	Page     int   `form:"page" binding:"min=1"`
	PageSize int   `form:"page_size" binding:"min=1,max=100"`
	Ctime    int64 `form:"created_at" binding:"min=0"`
}

// StopAssistantAgentChatReq 停止辅助Agent会话请求
type StopAssistantAgentChatReq struct {
	TaskID uuid.UUID `json:"task_id" binding:"required"`
}
