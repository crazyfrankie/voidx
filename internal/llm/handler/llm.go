package handler

import "github.com/crazyfrankie/voidx/internal/llm/service"

type LLMHandler struct {
	svc *service.LLMService
}

func NewLLMHandler(svc *service.LLMService) *LLMHandler {
	return &LLMHandler{svc: svc}
}
