package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"

	"github.com/crazyfrankie/voidx/internal/core/agent/entities"
)

// BaseAgent represents the interface for all agent implementations
// 基于Runnable的基础智能体基类
type BaseAgent interface {
	// Invoke executes the agent with the given input and returns a result
	// 块内容响应，一次性生成完整内容后返回
	Invoke(ctx context.Context, input entities.AgentState) (entities.AgentResult, error)

	// Stream executes the agent and returns a channel of thoughts
	// 流式输出，每个Node节点或者LLM每生成一个token时则会返回相应内容
	Stream(ctx context.Context, input entities.AgentState) (<-chan entities.AgentThought, error)

	// GetQueueManager returns the agent's queue manager
	GetQueueManager() *AgentQueueManager
}

// baseAgentImpl provides a base implementation of the BaseAgent interface
type baseAgentImpl struct {
	llm          llms.Model
	agentConfig  entities.AgentConfig
	queueManager *AgentQueueManager
}

// NewBaseAgent creates a new base agent implementation
func NewBaseAgent(llm llms.Model, config entities.AgentConfig, agentQueueManager *AgentQueueManager) BaseAgent {
	return &baseAgentImpl{
		llm:          llm,
		agentConfig:  config,
		queueManager: agentQueueManager,
	}
}

// GetQueueManager returns the agent's queue manager
func (b *baseAgentImpl) GetQueueManager() *AgentQueueManager {
	return b.queueManager
}

// Invoke implements the BaseAgent interface
func (b *baseAgentImpl) Invoke(ctx context.Context, input entities.AgentState) (entities.AgentResult, error) {
	// 1.调用stream方法获取流式事件输出数据
	content := ""
	query := ""
	var imageURLs []string

	if len(input.Messages) > 0 {
		lastMsg := input.Messages[len(input.Messages)-1]
		content = lastMsg.GetContent()
		query = content
		imageURLs = extractImageURLsFromMessage(lastMsg)
	}

	agentResult := entities.AgentResult{
		Query:     query,
		ImageURLs: imageURLs,
		CreatedAt: time.Now(),
		Status:    entities.EventAgentMessage,
	}

	agentThoughts := make(map[string]entities.AgentThought)

	// Get streaming channel
	thoughtChan, err := b.Stream(ctx, input)
	if err != nil {
		return agentResult, fmt.Errorf("failed to start streaming: %w", err)
	}

	// 2.提取事件id并转换成字符串
	for agentThought := range thoughtChan {
		eventID := agentThought.ID.String()

		// 3.除了ping事件，其他事件全部记录
		if agentThought.Event != entities.EventPing {
			// 4.单独处理agent_message事件，因为该事件为数据叠加
			if agentThought.Event == entities.EventAgentMessage {
				// 5.检测是否已存储了事件
				if existing, exists := agentThoughts[eventID]; exists {
					// 6.叠加智能体消息事件
					existing.Thought = existing.Thought + agentThought.Thought
					existing.Answer = existing.Answer + agentThought.Answer
					existing.Latency = agentThought.Latency
					agentThoughts[eventID] = existing
				} else {
					// 7.初始化智能体消息事件
					agentThoughts[eventID] = agentThought
				}
				// 8.更新智能体消息答案
				agentResult.Answer += agentThought.Answer
			} else {
				// 9.处理其他类型的智能体事件，类型均为覆盖
				agentThoughts[eventID] = agentThought

				// 10.单独判断是否为异常消息类型，如果是则修改状态并记录错误
				if agentThought.Event == entities.EventStop ||
					agentThought.Event == entities.EventTimeout ||
					agentThought.Event == entities.EventError {
					agentResult.Status = agentThought.Event
					if agentThought.Event == entities.EventError {
						agentResult.Error = agentThought.Observation
					} else {
						agentResult.Error = ""
					}
				}
			}
		}
	}

	// 11.将推理字典转换成列表并存储
	agentResult.AgentThoughts = make([]entities.AgentThought, 0, len(agentThoughts))
	for _, agentThought := range agentThoughts {
		agentResult.AgentThoughts = append(agentResult.AgentThoughts, agentThought)
	}

	// 12.完善message
	for _, agentThought := range agentThoughts {
		if agentThought.Event == entities.EventAgentMessage && len(agentThought.Message) > 0 {
			agentResult.Message = agentThought.Message
			break
		}
	}

	// 13.更新总耗时
	for _, agentThought := range agentResult.AgentThoughts {
		agentResult.Latency += agentThought.Latency
	}

	return agentResult, nil
}

// Stream implements the BaseAgent interface
func (b *baseAgentImpl) Stream(ctx context.Context, input entities.AgentState) (<-chan entities.AgentThought, error) {
	// 1.检测子类是否已构建Agent智能体，如果未构建则抛出错误
	// 这里是基础实现，具体的智能体会重写这个方法

	// 2.构建对应的任务id及数据初始化
	if input.TaskID == uuid.Nil {
		input.TaskID = uuid.New()
	}
	if input.History == nil {
		input.History = make([]llms.ChatMessage, 0)
	}
	if input.IterationCount == 0 {
		input.IterationCount = 0
	}

	// 3.创建队列
	thoughtChan, err := b.queueManager.Listen(input.TaskID, b.agentConfig.UserID, b.agentConfig.InvokeFrom)
	if err != nil {
		return nil, fmt.Errorf("failed to create queue: %w", err)
	}

	// 4.创建子线程并执行
	go func() {
		defer func() {
			// 发送结束事件
			b.queueManager.Publish(input.TaskID, entities.AgentThought{
				ID:     uuid.New(),
				TaskID: input.TaskID,
				Event:  entities.EventAgentEnd,
			})
		}()

		// 基础实现 - 简单回显输入
		if len(input.Messages) > 0 {
			lastMsg := input.Messages[len(input.Messages)-1]
			content := extractQueryFromMessage(lastMsg)

			b.queueManager.Publish(input.TaskID, entities.AgentThought{
				ID:      uuid.New(),
				TaskID:  input.TaskID,
				Event:   entities.EventAgentMessage,
				Thought: content,
				Answer:  content,
				Latency: 0.1,
			})
		}
	}()

	// 5.调用队列管理器监听数据并返回迭代器
	return thoughtChan, nil
}

// extractQueryFromMessage extracts the text content from a chat message
func extractQueryFromMessage(msg llms.ChatMessage) string {
	return msg.GetContent()
}

// extractImageURLsFromMessage extracts image URLs from a chat message
func extractImageURLsFromMessage(msg llms.ChatMessage) []string {
	content := msg.GetContent()
	var imageURLs []string

	// Simple check for image URLs
	if strings.Contains(content, "http") && (strings.Contains(content, ".jpg") ||
		strings.Contains(content, ".png") || strings.Contains(content, ".gif") ||
		strings.Contains(content, ".jpeg")) {
		words := strings.Fields(content)
		for _, word := range words {
			if strings.HasPrefix(word, "http") && (strings.Contains(word, ".jpg") ||
				strings.Contains(word, ".png") || strings.Contains(word, ".gif") ||
				strings.Contains(word, ".jpeg")) {
				imageURLs = append(imageURLs, word)
			}
		}
	}

	return imageURLs
}
