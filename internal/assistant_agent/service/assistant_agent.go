package service

import (
	"context"
	"errors"
	"github.com/crazyfrankie/voidx/internal/core/memory"
	"sync"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"

	"github.com/crazyfrankie/voidx/internal/assistant_agent/repository"
	"github.com/crazyfrankie/voidx/internal/assistant_agent/task"
	"github.com/crazyfrankie/voidx/internal/conversation/service"
	llmentity "github.com/crazyfrankie/voidx/internal/core/llm/entity"
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
	llm                 llmentity.BaseLanguageModel
	// 用于跟踪活跃的聊天会话，按用户ID分组
	activeSessions sync.Map // map[string]map[string]context.CancelFunc - key: userID, value: map[taskID]cancelFunc
}

func NewAssistantAgentService(repo *repository.AssistantAgentRepo, conversationService *service.ConversationService,
	appProducer *task.AppProducer, llm llmentity.BaseLanguageModel, tokenBufMem *memory.TokenBufferMemory) *AssistantAgentService {
	return &AssistantAgentService{
		llm:                 llm,
		repo:                repo,
		conversationService: conversationService,
		tokenBufMem:         tokenBufMem,
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

	s.tokenBufMem = s.tokenBufMem.WithLLM(s.llm)
	history, err := s.tokenBufMem.GetHistoryPromptMessages(2000, 3)
	if err != nil {
		return nil, err
	}

	tools, err :=

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

	// 3. 查找并取消对应的任务
	if !s.cancelUserSession(userID.String(), taskID.String()) {
		return errno.ErrValidate.AppendBizMessage(errors.New("未找到对应的活跃会话"))
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
