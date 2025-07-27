package req

import "github.com/google/uuid"

// AudioToTextReq 语音转文本请求
type AudioToTextReq struct {
	File []byte `json:"file" binding:"required"`
}

// MessageToAudioReq 消息转流式事件语音请求
type MessageToAudioReq struct {
	MessageID uuid.UUID `json:"message_id" binding:"required"`
}
