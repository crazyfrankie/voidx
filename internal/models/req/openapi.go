package req

import "github.com/google/uuid"

// OpenAPIChatReq 开放API聊天接口请求
type OpenAPIChatReq struct {
	AppID          uuid.UUID `json:"app_id" binding:"required"`
	EndUserID      uuid.UUID `json:"end_user_id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	Query          string    `json:"query" binding:"required"`
	ImageUrls      []string  `json:"image_urls" validate:"max=5,dive,url"`
	Stream         bool      `json:"stream"`
}
