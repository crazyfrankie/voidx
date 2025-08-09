package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/bytedance/sonic"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"

	"github.com/crazyfrankie/voidx/internal/ai/repository"
	"github.com/crazyfrankie/voidx/internal/conversation/service"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/errno"
)

const OptimizePromptTemplate = `# 角色
你是一位AI提示词工程师，你商场根据用户的需求，优化和组成AI提示词。

## 技能
- 确定用户给出的原始提示词的语言和意图
- 根据用户的提示（如果有）优化提示词
- 返回给用户优化后的提示词
- 根据样本提示词示例参考并返回优化后的提示词。以下是一个优化后样式提示词示例:

<example>
# 角色
你是一个幽默的电影评论员，擅长用轻松的语言解释电影情节和介绍最新电影，你擅长把复杂的电影概念解释得各类观众都能理解。

## 技能
### 技能1: 推荐新电影
- 发现用户最喜欢的电影类型。
- 如果提到的电影是未知的，搜索(site:douban.com)以确定其类型。
- 使用googleWebSearch()在https://movie.douban.com/cinema/nowplaying/beijing/上查找最新上映的电影。
- 根据用户的喜好，推荐几部正在上映或即将上映的电影。格式示例:
====
 - 电影名称: <电影名称>
 - 上映日期: <中国上映日期>
 - 故事简介: <100字以内的剧情简介>
====

### 技能2: 介绍电影
- 使用search(site:douban.com)找到用户查询电影的详细信息。
- 如果需要，可以使用googleWebSearch()获取更多信息。
- 根据搜索结果创建电影介绍。

### 技能3: 解释电影概念
- 使用recallDataset获取相关信息，并向用户解释概念。
- 使用熟悉的电影来说明此概念。

## 限制
- 只讨论与电影相关的话题。
- 固定提供的输出格式。
- 保持摘要在100字内。
- 使用知识库内容，对于未知电影，使用搜索和浏览。
- 采用^^ Markdown格式来引用数据源。
</example>

## 约束
- 只回答和提示词创建或优化相关的内容，如果用户提出其他问题，不要回答。
- 只使用原始提示所使用的语言。
- 只使用用户使用的语言。
- 请按照示例结果返回数据，不要携带<example>标签。`

type AIService struct {
	repo                *repository.AIRepo
	conversationService *service.ConversationService
}

func NewAIService(repo *repository.AIRepo, conversationService *service.ConversationService) *AIService {
	return &AIService{
		repo:                repo,
		conversationService: conversationService,
	}
}

// GenerateSuggestedQuestions 根据传递的消息id+账号生成建议问题列表
func (s *AIService) GenerateSuggestedQuestions(ctx context.Context, messageID, userID uuid.UUID) ([]string, error) {
	// 1. 查询消息并校验权限信息
	message, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("消息不存在"))
	}

	if message.CreatedBy != userID {
		return nil, errno.ErrForbidden.AppendBizMessage(errors.New("该条消息不存在或无权限"))
	}

	// 2. 构建对话历史列表
	histories := fmt.Sprintf("Human: %s\nAI: %s", message.Query, message.Answer)

	// 3. 调用会话服务生成建议问题
	return s.conversationService.GenerateSuggestedQuestions(ctx, histories)
}

// OptimizePrompt 根据传递的prompt进行优化生成
func (s *AIService) OptimizePrompt(ctx context.Context, prompt string) (<-chan string, error) {
	// 创建事件通道
	eventChan := make(chan string, 100)

	// 启动异步处理
	go s.processOptimizePrompt(ctx, prompt, eventChan)

	return eventChan, nil
}

// processOptimizePrompt 处理prompt优化
func (s *AIService) processOptimizePrompt(ctx context.Context, prompt string, eventChan chan<- string) {
	defer close(eventChan)

	// 1. 创建OpenAI客户端
	client := s.getOpenAIClient()

	// 2. 构建消息
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: OptimizePromptTemplate,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}

	// 3. 创建流式请求
	req := openai.ChatCompletionRequest{
		Model:       openai.GPT4oMini,
		Messages:    messages,
		Stream:      true,
		Temperature: 0.5,
	}

	// 4. 调用流式API
	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		// 发送错误事件
		eventChan <- ""
		return
	}
	defer stream.Close()

	// 5. 处理流式响应
	var optimizedPrompt strings.Builder
	for {
		response, err := stream.Recv()
		if err != nil {
			break
		}

		if len(response.Choices) > 0 {
			delta := response.Choices[0].Delta.Content
			if delta != "" {
				optimizedPrompt.WriteString(delta)

				// 发送增量事件
				data := resp.OptimizePromptEvent{OptimizePrompt: delta}
				eventData, _ := sonic.Marshal(data)
				select {
				case eventChan <- fmt.Sprintf("event: %s\ndata: %s\n\n", "optimize_prompt", string(eventData)):
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

// getOpenAIClient 获取OpenAI客户端
func (s *AIService) getOpenAIClient() *openai.Client {
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	if baseURL := os.Getenv("OPENAI_API_BASE"); baseURL != "" {
		config.BaseURL = baseURL
	}
	return openai.NewClientWithConfig(config)
}
