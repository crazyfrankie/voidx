package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/crazyfrankie/voidx/pkg/util"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"

	"github.com/crazyfrankie/voidx/internal/assistant_agent/repository"
	"github.com/crazyfrankie/voidx/internal/assistant_agent/task"
	"github.com/crazyfrankie/voidx/internal/conversation/service"
	"github.com/crazyfrankie/voidx/internal/core/agent"
	agenteneity "github.com/crazyfrankie/voidx/internal/core/agent/entities"
	llmentity "github.com/crazyfrankie/voidx/internal/core/llm/entity"
	"github.com/crazyfrankie/voidx/internal/core/memory"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/consts"
	"github.com/crazyfrankie/voidx/pkg/errno"
)

type AssistantAgentService struct {
	repo                *repository.AssistantAgentRepo
	conversationService *service.ConversationService
	appProducer         *task.AppProducer
	tokenBufMem         *memory.TokenBufferMemory
	agentManager        *agent.AgentQueueManager
	llm                 llmentity.BaseLanguageModel
	// 用于跟踪活跃的聊天会话，按用户ID分组
	activeSessions sync.Map // map[string]map[string]context.CancelFunc - key: userID, value: map[taskID]cancelFunc
}

func NewAssistantAgentService(repo *repository.AssistantAgentRepo, conversationService *service.ConversationService,
	appProducer *task.AppProducer, llm llmentity.BaseLanguageModel, tokenBufMem *memory.TokenBufferMemory,
	agentManager *agent.AgentQueueManager) *AssistantAgentService {
	return &AssistantAgentService{
		llm:                 llm,
		repo:                repo,
		conversationService: conversationService,
		tokenBufMem:         tokenBufMem,
		appProducer:         appProducer,
		agentManager:        agentManager,
	}
}

// Chat 与辅助智能体进行对话聊天
func (s *AssistantAgentService) Chat(ctx context.Context, userID uuid.UUID, chatReq req.AssistantAgentChatReq) (<-chan string, error) {
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

	s.tokenBufMem = s.tokenBufMem.WithLLM(s.llm)
	history, err := s.tokenBufMem.GetHistoryPromptMessages(2000, 3)
	if err != nil {
		return nil, err
	}

	tool := s.convertCreateAppToTool(userID, s.appProducer)

	agentCfg := agenteneity.AgentConfig{
		UserID:               userID,
		InvokeFrom:           consts.InvokeFromDebugger,
		EnableLongTermMemory: true,
		Tools:                []tools.Tool{tool},
	}
	agentIns := agent.NewFunctionCallAgent(s.llm, agentCfg, s.agentManager)

	responseStream := make(chan string, 100)

	// 启动异步处理
	go s.processChat(ctx, assistantAgentID, userID, agentIns, history, message, conversation, chatReq, responseStream)

	return responseStream, nil
}

func (s *AssistantAgentService) processChat(ctx context.Context,
	appID uuid.UUID, userID uuid.UUID,
	agent agent.BaseAgent, history []llms.MessageContent,
	message *entity.Message, conversation *entity.Conversation,
	chatReq req.AssistantAgentChatReq,
	responseStream chan string) {

	defer close(responseStream)

	// 实现完整的Agent流式处理逻辑
	agentState := agenteneity.AgentState{
		TaskID:         uuid.New(),
		Messages:       util.MessageContentToChatMessages(history),
		History:        util.MessageContentToChatMessages(history),
		LongTermMemory: conversation.Summary,
		IterationCount: 0,
	}

	// 添加当前用户消息
	if len(chatReq.Query) > 0 {
		userMsg := llms.HumanChatMessage{Content: chatReq.Query}
		agentState.Messages = append(agentState.Messages, userMsg)
	}

	// 获取Agent流式输出
	thoughtChan, err := agent.Stream(ctx, agentState)
	if err != nil {
		select {
		case responseStream <- fmt.Sprintf("event: error\ndata: %s\n\n", err.Error()):
		case <-ctx.Done():
		}
		return
	}

	// 存储agent思考过程
	agentThoughts := make(map[string]agenteneity.AgentThought)

	// 处理流式输出
	for agentThought := range thoughtChan {
		eventID := agentThought.ID.String()

		// 除了ping事件，其他事件全部记录
		if agentThought.Event != agenteneity.EventPing {
			// 单独处理agent_message事件，因为该事件为数据叠加
			if agentThought.Event == agenteneity.EventAgentMessage {
				if existing, exists := agentThoughts[eventID]; exists {
					// 叠加智能体消息事件
					existing.Thought = existing.Thought + agentThought.Thought
					existing.Answer = existing.Answer + agentThought.Answer
					existing.Latency = agentThought.Latency
					agentThoughts[eventID] = existing
				} else {
					// 初始化智能体消息事件
					agentThoughts[eventID] = agentThought
				}
			} else {
				// 处理其他类型事件的消息
				agentThoughts[eventID] = agentThought
			}
		}

		// 构建响应数据
		data := map[string]any{
			"id":              eventID,
			"conversation_id": conversation.ID.String(),
			"message_id":      message.ID.String(),
			"task_id":         agentState.TaskID.String(),
			"event":           string(agentThought.Event),
			"thought":         agentThought.Thought,
			"observation":     agentThought.Observation,
			"tool":            agentThought.Tool,
			"tool_input":      agentThought.ToolInput,
			"answer":          agentThought.Answer,
			"latency":         agentThought.Latency,
		}

		jsonData, _ := sonic.Marshal(data)
		eventStr := fmt.Sprintf("event: %s\ndata: %s\n\n", agentThought.Event, string(jsonData))

		select {
		case responseStream <- eventStr:
		case <-ctx.Done():
			return
		}
	}

	// 将agent思考过程转换为切片
	agentThoughtsList := make([]agenteneity.AgentThought, 0, len(agentThoughts))
	for _, thought := range agentThoughts {
		agentThoughtsList = append(agentThoughtsList, thought)
	}

	// 将消息以及推理过程添加到数据库
	err = s.conversationService.SaveAgentThoughts(ctx, userID, appID, conversation.ID, message.ID, agentThoughtsList)
	if err != nil {
		// 记录错误但不中断流程
		fmt.Printf("Failed to save agent thoughts: %v\n", err)
	}
}

// StopChat 停止与辅助智能体的对话聊天
func (s *AssistantAgentService) StopChat(ctx context.Context, taskID, userID uuid.UUID) error {
	// 1. 验证用户权限
	account, err := s.repo.GetAccountByID(ctx, userID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("用户不存在"))
	}

	// 2. 检查是否有正在进行的会话
	if account.AssistantAgentConversationID == uuid.Nil {
		return errno.ErrValidate.AppendBizMessage(errors.New("没有正在进行的对话"))
	}

	return s.agentManager.SetStopFlag(taskID, consts.InvokeFromWebApp, userID)
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
		pageReq.CurrentPage,
		pageReq.PageSize,
		pageReq.Ctime,
	)
	if err != nil {
		return nil, nil, resp.Paginator{}, err
	}

	// 计算分页信息
	totalPages := (int(total) + pageReq.PageSize - 1) / pageReq.PageSize
	paginator := resp.Paginator{
		CurrentPage: pageReq.CurrentPage,
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
		return errno.ErrNotFound.AppendBizMessage(errors.New("用户不存在"))
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
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("用户不存在"))
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

type CreateAppTool struct {
	accountID uuid.UUID

	appProducer *task.AppProducer
}

type CreateAppInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (s *AssistantAgentService) convertCreateAppToTool(accountID uuid.UUID, producer *task.AppProducer) tools.Tool {
	return &CreateAppTool{
		accountID:   accountID,
		appProducer: producer,
	}
}

func (t *CreateAppTool) Name() string {
	return "需要创建的Agent/应用名称，长度不超过50个字符"
}

func (t *CreateAppTool) Description() string {
	return "需要创建的Agent/应用描述，请详细概括该应用的功能"
}

func (t *CreateAppTool) Call(ctx context.Context, input string) (string, error) {
	// 解析输入参数
	var params CreateAppInput
	err := sonic.Unmarshal([]byte(input), &params)
	if err != nil {
		return "", fmt.Errorf("invalid input format: %v", err)
	}

	// 验证参数
	if params.Name == "" || len(params.Name) > 50 {
		return "", fmt.Errorf("name must be between 1 and 50 characters")
	}
	if params.Description == "" {
		return "", fmt.Errorf("description cannot be empty")
	}

	// 1. 调用异步任务在后端创建应用
	err = t.appProducer.PublishAutoCreateAppTask(ctx, params.Name, params.Description, t.accountID)
	if err != nil {
		return "", err
	}

	// 2. 返回成功提示
	result := fmt.Sprintf("已调用后端异步任务创建Agent应用。\n应用名称: %s\n应用描述: %s",
		params.Name, params.Description)
	return result, nil
}
