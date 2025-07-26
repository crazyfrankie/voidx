package resp

// AudioToTextResp 语音转文本响应
type AudioToTextResp struct {
	Text string `json:"text"`
}

// TTSEvent 文本转语音事件
type TTSEvent struct {
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id"`
	Audio          string `json:"audio"` // base64编码的音频数据
}
