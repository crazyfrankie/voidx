package llm

import (
	"github.com/tmc/langchaingo/llms/openai"
)

type LanguageModel struct {
	openai.LLM
}
