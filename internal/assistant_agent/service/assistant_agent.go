package service

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"

	"github.com/crazyfrankie/voidx/internal/assistant_agent/repository"
	"github.com/crazyfrankie/voidx/internal/assistant_agent/task"
	"github.com/crazyfrankie/voidx/internal/conversation/service"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/consts"
	"github.com/crazyfrankie/voidx/pkg/errno"
)

type AssistantAgentService struct {
	repo                *repository.AssistantAgentRepo
	conversationService *service.ConversationService
	openaiClient        *openai.Client
	appProducer         *task.AppProducer
	// 用于跟踪活跃的聊天会话，按用户ID分组
	activeSessions sync.Map // map[string]map[string]context.CancelFunc - key: userID, value: map[taskID]cancelFunc
}

func NewAssistantAgentService(repo *repository.AssistantAgentRepo,
	conversationService *service.ConversationService, appProducer *task.AppProducer) *AssistantAgentService {
	// 初始化OpenAI客户端
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	if baseURL := os.Getenv("OPENAI_API_BASE"); baseURL != "" {
		config.BaseURL = baseURL
	}
	openaiClient := openai.NewClientWithConfig(config)

	return &AssistantAgentService{
		repo:                repo,
		conversationService: conversationService,
		openaiClient:        openaiClient,
		appProducer:         appProducer,
	}
}

// TODO
// Chat 与辅助智能体进行对话聊天
func (s *AssistantAgentService) Chat(ctx context.Context, userID uuid.UUID, chatReq req.AssistantAgentChatReq) (<-chan resp.AssistantAgentChatEvent, error) {
	assistantAgentID, _ := uuid.Parse(consts.AssistantAgentID)

	// 1. 获取或创建辅助Agent会话
	conversation, err := s.getOrCreateAssistantConversation(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 2. 创建消息记录
	message, err := s.conversationService.RawCreateMessage(ctx, &entity.Message{
		AppID:          assistantAgentID,
		ConversationID: conversation.ID,
		CreatedBy:      userID,
		Query:          chatReq.Query,
		ImageUrls:      chatReq.ImageUrls,
		InvokeFrom:     consts.InvokeFromAssistantAgent,
		Status:         consts.MessageStatusNormal,
	})
	if err != nil {
		return nil, err
	}

	// 3. 创建可取消的context和taskID
	taskID := uuid.New()
	chatCtx, cancel := context.WithCancel(ctx)

	// 4. 将取消函数存储到用户的活跃会话中
	s.addUserSession(userID.String(), taskID.String(), cancel)

	// 5. 创建事件通道
	eventChan := make(chan resp.AssistantAgentChatEvent, 100)

	// 6. 启动异步处理
	go s.processAssistantChat(chatCtx, userID, conversation, message, chatReq, eventChan, taskID)

	return eventChan, nil
}

// StopChat 停止与辅助智能体的对话聊天
func (s *AssistantAgentService) StopChat(ctx context.Context, taskID, userID uuid.UUID) error {
	// 1. 验证用户权限
	account, err := s.repo.GetAccountByID(ctx, userID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage("用户不存在")
	}

	// 2. 检查是否有正在进行的会话
	if account.AssistantAgentConversationID == uuid.Nil {
		return errno.ErrValidate.AppendBizMessage("没有正在进行的对话")
	}

	// 3. 查找并取消对应的任务
	if !s.cancelUserSession(userID.String(), taskID.String()) {
		return errno.ErrValidate.AppendBizMessage("未找到对应的活跃会话")
	}

	return nil
}

// GetMessagesWithPage 获取辅助智能体消息分页列表
func (s *AssistantAgentService) GetMessagesWithPage(ctx context.Context, userID uuid.UUID, pageReq req.GetAssistantAgentMessagesWithPageReq) ([]entity.Message, []entity.AgentThought, resp.Paginator, error) {
	// 获取辅助Agent会话
	conversation, err := s.getOrCreateAssistantConversation(ctx, userID)
	if err != nil {
		return nil, nil, resp.Paginator{}, err
	}

	// 获取消息列表
	messages, total, err := s.repo.GetMessagesByConversationID(
		ctx,
		conversation.ID,
		pageReq.Page,
		pageReq.PageSize,
		pageReq.Ctime,
	)
	if err != nil {
		return nil, nil, resp.Paginator{}, err
	}

	// 计算分页信息
	totalPages := (int(total) + pageReq.PageSize - 1) / pageReq.PageSize
	paginator := resp.Paginator{
		CurrentPage: pageReq.Page,
		PageSize:    pageReq.PageSize,
		TotalPage:   totalPages,
		TotalRecord: int(total),
	}

	return messages, nil, paginator, nil
}

// DeleteConversation 清空辅助Agent智能体会话消息列表
func (s *AssistantAgentService) DeleteConversation(ctx context.Context, userID uuid.UUID) error {
	// 1. 获取用户账户信息
	account, err := s.repo.GetAccountByID(ctx, userID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage("用户不存在")
	}

	// 2. 如果存在辅助Agent会话，则删除
	if account.AssistantAgentConversationID != uuid.Nil {
		err = s.conversationService.DeleteConversation(ctx, account.AssistantAgentConversationID)
		if err != nil {
			return err
		}

		// 3. 更新用户账户，清空辅助Agent会话ID
		err = s.repo.UpdateAccountAssistantConversation(ctx, userID, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

// getOrCreateAssistantConversation 获取或创建辅助Agent会话
func (s *AssistantAgentService) getOrCreateAssistantConversation(ctx context.Context, userID uuid.UUID) (*entity.Conversation, error) {
	// 1. 获取用户账户信息
	account, err := s.repo.GetAccountByID(ctx, userID)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage("用户不存在")
	}

	// 2. 如果已存在辅助Agent会话，直接返回
	if account.AssistantAgentConversationID != uuid.Nil {
		conversation, err := s.repo.GetConversationByID(ctx, account.AssistantAgentConversationID)
		if err == nil && conversation != nil {
			return conversation, nil
		}
	}

	// 3. 创建新的辅助Agent会话
	assistantAppID, _ := uuid.Parse(consts.AssistantAgentID)
	conversation, err := s.conversationService.CreateConversation(ctx, userID, req.CreateConversationReq{
		AppID:      assistantAppID,
		Name:       "Assistant Agent Conversation",
		InvokeFrom: string(consts.InvokeFromAssistantAgent),
	})
	if err != nil {
		return nil, err
	}

	// 4. 更新用户账户，关联辅助Agent会话
	err = s.repo.UpdateAccountAssistantConversation(ctx, userID, &conversation.ID)
	if err != nil {
		return nil, err
	}

	return conversation, nil
}

// processAssistantChat 处理辅助Agent对话
func (s *AssistantAgentService) processAssistantChat(ctx context.Context, userID uuid.UUID, conversation *entity.Conversation, message *entity.Message, chatReq req.AssistantAgentChatReq, eventChan chan<- resp.AssistantAgentChatEvent, taskID uuid.UUID) {
	defer close(eventChan)
	defer s.removeUserSession(userID.String(), taskID.String()) // 清理活跃会话

	startTime := time.Now()

	// 检查context是否已被取消
	select {
	case <-ctx.Done():
		s.sendEvent(eventChan, resp.AssistantAgentChatEvent{
			ID:             uuid.New().String(),
			ConversationID: conversation.ID.String(),
			MessageID:      message.ID.String(),
			TaskID:         taskID.String(),
			Event:          "error",
			Answer:         "对话已被取消",
			Latency:        time.Since(startTime).Seconds(),
		})
		return
	default:
	}

	// 发送思考事件
	s.sendEvent(eventChan, resp.AssistantAgentChatEvent{
		ID:             uuid.New().String(),
		ConversationID: conversation.ID.String(),
		MessageID:      message.ID.String(),
		TaskID:         taskID.String(),
		Event:          "agent_thought",
		Thought:        "正在分析用户问题...",
		Latency:        time.Since(startTime).Seconds(),
	})

	// 检查是否被取消
	select {
	case <-ctx.Done():
		return
	default:
	}

	// 获取对话历史
	historyMessages, _, err := s.repo.GetMessagesByConversationID(ctx, conversation.ID, 1, 10, 0)
	if err != nil {
		s.sendErrorEvent(eventChan, conversation, message, taskID, "获取对话历史失败")
		return
	}

	// 构建对话上下文
	s.sendEvent(eventChan, resp.AssistantAgentChatEvent{
		ID:             uuid.New().String(),
		ConversationID: conversation.ID.String(),
		MessageID:      message.ID.String(),
		TaskID:         taskID.String(),
		Event:          "agent_thought",
		Thought:        "正在构建对话上下文...",
		Latency:        time.Since(startTime).Seconds(),
	})

	messages := s.buildChatMessages(historyMessages, chatReq.Query)

	// 检查是否被取消
	select {
	case <-ctx.Done():
		return
	default:
	}

	// 调用OpenAI API
	s.sendEvent(eventChan, resp.AssistantAgentChatEvent{
		ID:             uuid.New().String(),
		ConversationID: conversation.ID.String(),
		MessageID:      message.ID.String(),
		TaskID:         taskID.String(),
		Event:          "agent_thought",
		Thought:        "正在生成AI回复...",
		Latency:        time.Since(startTime).Seconds(),
	})

	// 定义工具
	tools := []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "create_app",
				Description: "如果用户提出了需要创建一个Agent/应用，你可以调用此工具，参数的输入是应用的名称+描述，返回的数据是创建后的成功提示",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "需要创建的Agent/应用名称，长度不超过50个字符",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "需要创建的Agent/应用描述，请详细概括该应用的功能",
						},
					},
					"required": []string{"name", "description"},
				},
			},
		},
	}

	// 创建流式请求
	stream, err := s.openaiClient.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       openai.GPT4oMini,
		Messages:    messages,
		Tools:       tools,
		ToolChoice:  "auto",
		Temperature: 0.7,
		MaxTokens:   2000,
		Stream:      true,
	})
	if err != nil {
		s.sendErrorEvent(eventChan, conversation, message, taskID, fmt.Sprintf("AI调用失败: %v", err))
		return
	}
	defer stream.Close()

	// 处理流式响应
	var fullAnswer strings.Builder
	var totalTokens int
	var toolCalls []openai.ToolCall

	for {
		// 检查是否被取消
		select {
		case <-ctx.Done():
			return
		default:
		}

		response, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			s.sendErrorEvent(eventChan, conversation, message, taskID, fmt.Sprintf("接收AI响应失败: %v", err))
			return
		}

		if len(response.Choices) > 0 {
			choice := response.Choices[0]

			// 处理文本内容
			if choice.Delta.Content != "" {
				fullAnswer.WriteString(choice.Delta.Content)

				// 发送增量消息事件
				s.sendEvent(eventChan, resp.AssistantAgentChatEvent{
					ID:             uuid.New().String(),
					ConversationID: conversation.ID.String(),
					MessageID:      message.ID.String(),
					TaskID:         taskID.String(),
					Event:          "agent_message",
					Answer:         fullAnswer.String(),
					Latency:        time.Since(startTime).Seconds(),
				})
			}

			// 处理工具调用
			if len(choice.Delta.ToolCalls) > 0 {
				toolCalls = append(toolCalls, choice.Delta.ToolCalls...)
			}
		}

		if response.Usage != nil {
			totalTokens = response.Usage.TotalTokens
		}
	}

	// 处理工具调用
	if len(toolCalls) > 0 {
		for _, toolCall := range toolCalls {
			if toolCall.Function.Name == "create_app" {
				s.sendEvent(eventChan, resp.AssistantAgentChatEvent{
					ID:             uuid.New().String(),
					ConversationID: conversation.ID.String(),
					MessageID:      message.ID.String(),
					TaskID:         taskID.String(),
					Event:          "agent_thought",
					Thought:        "正在创建应用...",
					Tool:           toolCall.Function.Name,
					ToolInput:      toolCall.Function.Arguments,
					Latency:        time.Since(startTime).Seconds(),
				})

				// 解析工具参数
				var args struct {
					Name        string `json:"name"`
					Description string `json:"description"`
				}
				if err := sonic.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
					s.sendErrorEvent(eventChan, conversation, message, taskID, "解析工具参数失败")
					continue
				}

				// 调用异步任务
				if err := s.appProducer.PublishAutoCreateAppTask(ctx, args.Name, args.Description, userID); err != nil {
					s.sendErrorEvent(eventChan, conversation, message, taskID, "发布创建应用任务失败")
					continue
				}

				// 发送工具执行结果
				toolResult := fmt.Sprintf("已调用后端异步任务创建Agent应用。\n应用名称: %s\n应用描述: %s", args.Name, args.Description)
				fullAnswer.WriteString("\n\n" + toolResult)

				s.sendEvent(eventChan, resp.AssistantAgentChatEvent{
					ID:             uuid.New().String(),
					ConversationID: conversation.ID.String(),
					MessageID:      message.ID.String(),
					TaskID:         taskID.String(),
					Event:          "agent_thought",
					Observation:    toolResult,
					Latency:        time.Since(startTime).Seconds(),
				})
			}
		}
	}

	// 发送最终消息事件
	finalAnswer := fullAnswer.String()
	s.sendEvent(eventChan, resp.AssistantAgentChatEvent{
		ID:              uuid.New().String(),
		ConversationID:  conversation.ID.String(),
		MessageID:       message.ID.String(),
		TaskID:          taskID.String(),
		Event:           "message_end",
		Answer:          finalAnswer,
		Latency:         time.Since(startTime).Seconds(),
		TotalTokenCount: totalTokens,
	})

	// 更新消息状态
	s.conversationService.UpdateMessage(ctx, userID, message.ID, req.UpdateMessageReq{
		Answer: finalAnswer,
		Status: string(consts.MessageStatusNormal),
	})
}

// sendEvent 发送事件到通道
func (s *AssistantAgentService) sendEvent(eventChan chan<- resp.AssistantAgentChatEvent, event resp.AssistantAgentChatEvent) {
	select {
	case eventChan <- event:
	default:
		// 通道已满或已关闭，忽略
	}
}

// sendErrorEvent 发送错误事件
func (s *AssistantAgentService) sendErrorEvent(eventChan chan<- resp.AssistantAgentChatEvent,
	conversation *entity.Conversation, message *entity.Message, taskID uuid.UUID, errorMsg string) {
	s.sendEvent(eventChan, resp.AssistantAgentChatEvent{
		ID:             uuid.New().String(),
		ConversationID: conversation.ID.String(),
		MessageID:      message.ID.String(),
		TaskID:         taskID.String(),
		Event:          "error",
		Answer:         fmt.Sprintf("抱歉，处理您的请求时出现错误: %s", errorMsg),
		Latency:        0,
	})
}

// buildChatMessages 构建对话消息
func (s *AssistantAgentService) buildChatMessages(historyMessages []entity.Message, currentQuery string) []openai.ChatCompletionMessage {
	messages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleSystem,
			Content: `你是一个智能助手，专门帮助用户解决各种问题。请遵循以下原则：
1. 提供准确、有用的信息
2. 保持友好和专业的语调
3. 如果不确定答案，请诚实说明
4. 尽量提供具体的建议和解决方案
5. 根据上下文理解用户的真实需求`,
		},
	}

	// 添加历史对话（最近的5轮对话）
	maxHistory := 5
	if len(historyMessages) > maxHistory {
		historyMessages = historyMessages[:maxHistory]
	}

	// 倒序遍历，最新的在前
	for i := len(historyMessages) - 1; i >= 0; i-- {
		msg := historyMessages[i]
		if msg.Query != "" {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: msg.Query,
			})
		}
		if msg.Answer != "" {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: msg.Answer,
			})
		}
	}

	// 添加当前问题
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: currentQuery,
	})

	return messages
}

// getUserSessions 获取用户的会话map，如果不存在则创建
func (s *AssistantAgentService) getUserSessions(userID string) map[string]context.CancelFunc {
	if sessions, exists := s.activeSessions.Load(userID); exists {
		return sessions.(map[string]context.CancelFunc)
	}

	// 创建新的用户会话map
	userSessions := make(map[string]context.CancelFunc)
	s.activeSessions.Store(userID, userSessions)
	return userSessions
}

// addUserSession 添加用户会话
func (s *AssistantAgentService) addUserSession(userID, taskID string, cancel context.CancelFunc) {
	if sessions, exists := s.activeSessions.Load(userID); exists {
		userSessions := sessions.(map[string]context.CancelFunc)
		userSessions[taskID] = cancel
	} else {
		userSessions := make(map[string]context.CancelFunc)
		userSessions[taskID] = cancel
		s.activeSessions.Store(userID, userSessions)
	}
}

// removeUserSession 移除用户会话
func (s *AssistantAgentService) removeUserSession(userID, taskID string) {
	if sessions, exists := s.activeSessions.Load(userID); exists {
		userSessions := sessions.(map[string]context.CancelFunc)
		delete(userSessions, taskID)

		// 如果用户没有活跃会话了，清理用户记录
		if len(userSessions) == 0 {
			s.activeSessions.Delete(userID)
		}
	}
}

// cancelUserSession 取消用户的特定会话
func (s *AssistantAgentService) cancelUserSession(userID, taskID string) bool {
	if sessions, exists := s.activeSessions.Load(userID); exists {
		userSessions := sessions.(map[string]context.CancelFunc)
		if cancel, exists := userSessions[taskID]; exists {
			cancel()
			delete(userSessions, taskID)

			// 如果用户没有活跃会话了，清理用户记录
			if len(userSessions) == 0 {
				s.activeSessions.Delete(userID)
			}
			return true
		}
	}
	return false
}
