package controller

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"

	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/response"
)

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

type ChatHandler struct{}

func NewChatHandler() *ChatHandler {
	return &ChatHandler{}
}

func (h *ChatHandler) RegisterRoute(r *gin.RouterGroup) {
	chatGroup := r.Group("chat")
	{
		chatGroup.POST("", h.Completion())
	}
}

func (h *ChatHandler) Completion() gin.HandlerFunc {
	return func(c *gin.Context) {
		var chatReq req.ChatReq
		if err := c.ShouldBind(&chatReq); err != nil {
			response.Error(c, errno.Success)
			return
		}

		// Initialize the OpenAI client with Deepseek model
		llm, err := openai.New(MoonShot...)
		if err != nil {
			log.Fatal(err)
		}

		ctx := context.Background()

		// Create messages for the chat
		content := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeSystem, "You are a helpful assistant that explains complex topics step by step"),
			llms.TextParts(llms.ChatMessageTypeHuman, chatReq.Query),
		}

		// Generate content with streaming to see both reasoning and final answer in real-time
		completion, err := llm.GenerateContent(
			ctx,
			content,
			llms.WithMaxTokens(2000),
			llms.WithTemperature(0.7),
			llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
				fmt.Print(string(chunk))
				return nil
			}),
		)
		if err != nil {
			log.Fatal(err)
		}

		// Access the reasoning content and final answer separately
		if len(completion.Choices) > 0 {
			choice := completion.Choices[0]
			//fmt.Printf("\n\nReasoning Process:\n%s\n", choice.ReasoningContent)
			//fmt.Printf("\nFinal Answer:\n%s\n", choice.Content)
			response.SuccessWithData(c, map[string]any{
				"reasoningProcess": choice.ReasoningContent,
				"finalAnswer":      choice.Content,
			})
		}
	}
}
