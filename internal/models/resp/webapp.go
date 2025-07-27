package resp

import (
	"github.com/google/uuid"
)

// WebAppInfoResp WebApp信息响应
type WebAppInfoResp struct {
	ID                   uuid.UUID      `json:"id"`
	Name                 string         `json:"name"`
	Icon                 string         `json:"icon"`
	Description          string         `json:"description"`
	OpeningStatement     string         `json:"opening_statement"`
	OpeningQuestions     []string       `json:"opening_questions"`
	SpeechToText         map[string]any `json:"speech_to_text"`
	TextToSpeech         map[string]any `json:"text_to_speech"`
	SuggestedAfterAnswer map[string]any `json:"suggested_after_answer"`
	Features             []string       `json:"features"`
}

// WebAppConversationResp WebApp会话响应
type WebAppConversationResp struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	IsPinned   bool      `json:"is_pinned"`
	InvokeFrom string    `json:"invoke_from"`
	Ctime      int64     `json:"ctime"`
	Utime      int64     `json:"utime"`
}

// WebAppChatEventResp WebApp对话事件响应
type WebAppChatEventResp struct {
	ID              string  `json:"id"`
	ConversationID  string  `json:"conversation_id"`
	MessageID       string  `json:"message_id"`
	TaskID          string  `json:"task_id"`
	Event           string  `json:"event"`
	Thought         string  `json:"thought,omitempty"`
	Observation     string  `json:"observation,omitempty"`
	Tool            string  `json:"tool,omitempty"`
	ToolInput       any     `json:"tool_input,omitempty"`
	Answer          string  `json:"answer,omitempty"`
	TotalTokenCount int     `json:"total_token_count,omitempty"`
	TotalPrice      float64 `json:"total_price,omitempty"`
	Latency         float64 `json:"latency,omitempty"`
}
