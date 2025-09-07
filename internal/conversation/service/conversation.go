package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/crazyfrankie/voidx/types/consts"
	"github.com/crazyfrankie/voidx/types/errno"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"

	"github.com/crazyfrankie/voidx/internal/conversation/repository"
	"github.com/crazyfrankie/voidx/internal/core/agent/entities"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/util"
)

type ConversationService struct {
	repo *repository.ConversationRepo
}

func NewConversationService(repo *repository.ConversationRepo) *ConversationService {
	return &ConversationService{repo: repo}
}

func (s *ConversationService) GetConversationMessagesWithPage(ctx context.Context,
	conversationID uuid.UUID, pageReq req.GetConversationMessagesWithPageReq) ([]entity.Message, resp.Paginator, error) {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return nil, resp.Paginator{}, err
	}

	// 验证会话权限
	conversation, err := s.repo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return nil, resp.Paginator{}, errno.ErrNotFound.AppendBizMessage(errors.New("会话不存在"))
	}

	if conversation.CreatedBy != userID {
		return nil, resp.Paginator{}, errno.ErrForbidden.AppendBizMessage(errors.New("无权限访问该会话"))
	}

	// 获取消息列表
	messages, total, err := s.repo.GetMessagesByConversationID(ctx, conversationID, pageReq)
	if err != nil {
		return nil, resp.Paginator{}, err
	}

	// 计算分页信息
	totalPages := (int(total) + pageReq.PageSize - 1) / pageReq.PageSize
	paginator := resp.Paginator{
		CurrentPage: pageReq.CurrentPage,
		PageSize:    pageReq.PageSize,
		TotalPage:   totalPages,
		TotalRecord: int(total),
	}

	return messages, paginator, nil
}

func (s *ConversationService) DeleteConversation(ctx context.Context, conversationID uuid.UUID) error {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	// 验证会话权限
	conversation, err := s.repo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("会话不存在"))
	}

	if conversation.CreatedBy != userID {
		return errno.ErrForbidden.AppendBizMessage(errors.New("无权限删除该会话"))
	}

	return s.repo.DeleteConversation(ctx, conversationID)
}

func (s *ConversationService) DeleteMessage(ctx context.Context, conversationID, messageID uuid.UUID) error {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	// 验证会话权限
	conversation, err := s.repo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("会话不存在"))
	}

	if conversation.CreatedBy != userID {
		return errno.ErrForbidden.AppendBizMessage(errors.New("无权限操作该会话"))
	}

	// 验证消息是否属于该会话
	message, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("消息不存在"))
	}

	if message.ConversationID != conversationID {
		return errno.ErrValidate.AppendBizMessage(errors.New("消息不属于该会话"))
	}

	return s.repo.DeleteMessage(ctx, messageID)
}

func (s *ConversationService) GetConversationName(ctx context.Context, conversationID uuid.UUID) (string, error) {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return "", err
	}

	// 验证会话权限
	conversation, err := s.repo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return "", errno.ErrNotFound.AppendBizMessage(errors.New("会话不存在"))
	}

	if conversation.CreatedBy != userID {
		return "", errno.ErrForbidden.AppendBizMessage(errors.New("无权限访问该会话"))
	}

	return conversation.Name, nil
}

func (s *ConversationService) UpdateConversationName(ctx context.Context, conversationID uuid.UUID, name string) error {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	// 验证会话权限
	conversation, err := s.repo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("会话不存在"))
	}

	if conversation.CreatedBy != userID {
		return errno.ErrForbidden.AppendBizMessage(errors.New("无权限修改该会话"))
	}

	return s.repo.UpdateConversationName(ctx, conversationID, name)
}

func (s *ConversationService) UpdateConversationIsPinned(ctx context.Context, conversationID uuid.UUID, isPinned bool) error {
	userID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	// 验证会话权限
	conversation, err := s.repo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("会话不存在"))
	}

	if conversation.CreatedBy != userID {
		return errno.ErrForbidden.AppendBizMessage(errors.New("无权限修改该会话"))
	}

	return s.repo.UpdateConversationIsPinned(ctx, conversationID, isPinned)
}

// These functions for other module call

func (s *ConversationService) CreateConversation(ctx context.Context, accountID uuid.UUID, createReq req.CreateConversationReq) (*entity.Conversation, error) {
	// 创建会话实体
	conversation := &entity.Conversation{
		ID:         uuid.New(),
		AppID:      createReq.AppID,
		Name:       createReq.Name,
		InvokeFrom: consts.InvokeFrom(createReq.InvokeFrom),
		IsPinned:   false,
		CreatedBy:  accountID,
	}

	err := s.repo.CreateConversation(ctx, conversation)
	if err != nil {
		return nil, err
	}

	return conversation, nil
}

func (s *ConversationService) CreateMessage(ctx context.Context, accountID uuid.UUID, createReq req.CreateMessageReq) (*entity.Message, error) {
	// 验证会话权限
	conversation, err := s.repo.GetConversationByID(ctx, createReq.ConversationID)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("会话不存在"))
	}

	if conversation.CreatedBy != accountID {
		return nil, errno.ErrForbidden.AppendBizMessage(errors.New("无权限操作该会话"))
	}

	// 创建消息实体
	message := &entity.Message{
		ID:             uuid.New(),
		ConversationID: createReq.ConversationID,
		AppID:          conversation.AppID,
		InvokeFrom:     consts.InvokeFrom(createReq.InvokeFrom),
		ImageUrls:      createReq.ImageUrls,
		CreatedBy:      accountID,
		Query:          createReq.Query,
		Status:         consts.MessageStatusNormal,
	}

	err = s.repo.CreateMessage(ctx, message)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (s *ConversationService) UpdateMessage(ctx context.Context, accountID uuid.UUID, messageID uuid.UUID, updateReq req.UpdateMessageReq) error {
	// 验证消息权限
	message, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("消息不存在"))
	}

	// 验证会话权限
	conversation, err := s.repo.GetConversationByID(ctx, message.ConversationID)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("会话不存在"))
	}

	if conversation.CreatedBy != accountID {
		return errno.ErrForbidden.AppendBizMessage(errors.New("无权限操作该会话"))
	}

	// 构建更新字段
	updates := make(map[string]any)
	if updateReq.Answer != "" {
		updates["answer"] = updateReq.Answer
	}
	if updateReq.Status != "" {
		updates["status"] = updateReq.Status
	}

	return s.repo.UpdateMessage(ctx, messageID, updates)
}

func (s *ConversationService) GetConversations(ctx context.Context, accountID uuid.UUID, getReq req.GetConversationsReq) ([]entity.Conversation, error) {
	return s.repo.GetConversationsByAccountID(ctx, accountID, getReq)
}

func (s *ConversationService) GenerateSuggestedQuestions(ctx context.Context, histories string) ([]string, error) {
	// 使用OpenAI生成建议问题
	client := s.getOpenAIClient()

	prompt := fmt.Sprintf(`基于以下对话历史，生成3个相关的后续问题建议。请直接返回问题列表，每行一个问题，不要添加编号或其他格式：

对话历史：
%s

请生成3个自然、相关的后续问题：`, histories)

	req := openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   200,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		// 如果API调用失败，返回默认建议问题
		return []string{
			"能否详细解释一下？",
			"还有其他相关的信息吗？",
			"这个问题有什么实际应用？",
		}, nil
	}

	if len(resp.Choices) == 0 {
		return []string{}, nil
	}

	// 解析生成的问题
	content := resp.Choices[0].Message.Content
	questions := strings.Split(strings.TrimSpace(content), "\n")

	// 清理和过滤问题
	var result []string
	for _, q := range questions {
		q = strings.TrimSpace(q)
		// 移除可能的编号前缀
		q = regexp.MustCompile(`^\d+\.\s*`).ReplaceAllString(q, "")
		q = regexp.MustCompile(`^-\s*`).ReplaceAllString(q, "")
		if q != "" && len(q) > 5 {
			result = append(result, q)
		}
	}

	// 限制返回最多3个问题
	if len(result) > 3 {
		result = result[:3]
	}

	return result, nil
}

// getOpenAIClient 获取OpenAI客户端
func (s *ConversationService) getOpenAIClient() *openai.Client {
	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	if baseURL := os.Getenv("OPENAI_API_BASE"); baseURL != "" {
		config.BaseURL = baseURL
	}
	return openai.NewClientWithConfig(config)
}

// SaveAgentThoughts 保存Agent思考过程到数据库
func (s *ConversationService) SaveAgentThoughts(ctx context.Context, accountID, appID, conversationID, messageID uuid.UUID, agentThoughts []entities.AgentThought) error {
	// 构建最终答案
	var finalAnswer string
	var totalTokens int
	var totalPrice float64

	// 从思考过程中提取最终答案和统计信息
	for _, thought := range agentThoughts {
		if thought.Event == entities.EventAgentMessage && thought.Answer != "" {
			finalAnswer = thought.Answer
		}
		totalTokens += thought.TotalTokenCount
		totalPrice += thought.TotalPrice
	}

	// 更新消息的答案
	if finalAnswer != "" {
		err := s.UpdateMessage(ctx, accountID, messageID, req.UpdateMessageReq{
			Answer: finalAnswer,
		})
		if err != nil {
			return fmt.Errorf("failed to update message answer: %w", err)
		}
	}

	// 保存Agent思考过程
	for _, thought := range agentThoughts {
		agentThoughtEntity := &entity.AgentThought{
			ID:                thought.ID,
			AppID:             appID,
			MessageID:         messageID,
			ConversationID:    conversationID,
			CreatedBy:         accountID,
			Event:             string(thought.Event),
			Thought:           thought.Thought,
			Observation:       thought.Observation,
			Tool:              thought.Tool,
			ToolInput:         thought.ToolInput,
			Answer:            thought.Answer,
			MessageTokenCount: thought.MessageTokenCount,
			AnswerTokenCount:  thought.AnswerTokenCount,
			TotalTokenCount:   thought.TotalTokenCount,
			TotalPrice:        thought.TotalPrice,
			Latency:           thought.Latency,
		}

		err := s.repo.CreateAgentThought(ctx, agentThoughtEntity)
		if err != nil {
			// 记录错误但继续处理其他思考过程
			continue
		}
	}

	return nil
}

func (s *ConversationService) GetConversationByID(ctx context.Context, conversationID uuid.UUID) (*entity.Conversation, error) {
	return s.repo.GetConversationByID(ctx, conversationID)
}

func (s *ConversationService) RawCreateMessage(ctx context.Context, msg *entity.Message) (*entity.Message, error) {
	err := s.repo.CreateMessage(ctx, msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *ConversationService) RawCreateConversation(ctx context.Context, userID uuid.UUID, appID uuid.UUID) (*entity.Conversation, error) {
	convers := &entity.Conversation{
		AppID:     appID,
		CreatedBy: userID,
	}
	if err := s.repo.CreateConversation(ctx, convers); err != nil {
		return nil, err
	}

	return convers, nil
}

func (s *ConversationService) UpdateConversationSummary(ctx context.Context, conversationID uuid.UUID, summary string) error {
	return s.repo.UpdateConversationSummary(ctx, conversationID, summary)
}

func (s *ConversationService) GetConversationAgentThoughts(ctx context.Context, conversationID uuid.UUID) ([]entity.AgentThought, error) {
	return s.repo.GetConversationAgentThoughts(ctx, conversationID)
}
