package ioc

import "github.com/tmc/langchaingo/llms/openai"

var (
	MoonShot = []openai.Option{
		openai.WithModel("moonshot-v1-8k"),
		openai.WithBaseURL("https://api.moonshot.cn/v1"),
	}
	DeepSeek = []openai.Option{
		openai.WithModel("deepseek-reasoner"),
		openai.WithBaseURL("https://api.deepseek.com/v1"),
	}
)

func InitLLM() *openai.LLM {
	llm, err := openai.New(MoonShot...)
	if err != nil {
		panic(err)
	}

	return llm
}
