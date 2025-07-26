package req

import (
	"github.com/google/uuid"
)

// CreateConversationReq 创建会话请求
type CreateConversationReq struct {
	AppID      uuid.UUID `json:"app_id" binding:"required"`
	Name       string    `json:"name" binding:"required"`
	InvokeFrom string    `json:"invoke_from" binding:"required"`
}

// CreateMessageReq 创建消息请求
type CreateMessageReq struct {
	ConversationID uuid.UUID `json:"conversation_id" binding:"required"`
	Query          string    `json:"query" binding:"required"`
	ImageUrls      []string  `json:"image_urls"`
	InvokeFrom     string    `json:"invoke_from" binding:"required"`
}

// UpdateMessageReq 更新消息请求
type UpdateMessageReq struct {
	Answer string `json:"answer"`
	Status string `json:"status"`
}

// GetConversationsReq 获取会话列表请求
type GetConversationsReq struct {
	AppID      uuid.UUID `json:"app_id"`
	IsPinned   *bool     `json:"is_pinned"`
	InvokeFrom string    `json:"invoke_from"`
}
