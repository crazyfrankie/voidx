package ioc

import "github.com/crazyfrankie/voidx/internal/core/llm"

func InitLLMCore() *llm.LanguageModelManager {
	llmCore, err := llm.NewLanguageModelManager()
	if err != nil {
		return nil
	}

	return llmCore
}
