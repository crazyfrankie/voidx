package repository

import (
	"context"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/assistant_agent/repository/dao"
	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type AssistantAgentRepo struct {
	dao *dao.AssistantAgentDao
}

func NewAssistantAgentRepo(d *dao.AssistantAgentDao) *AssistantAgentRepo {
	return &AssistantAgentRepo{dao: d}
}

// GetAccountByID 根据ID获取账户信息
func (r *AssistantAgentRepo) GetAccountByID(ctx context.Context, id uuid.UUID) (*entity.Account, error) {
	return r.dao.GetAccountByID(ctx, id)
}

// UpdateAccountAssistantConversation 更新账户的辅助Agent会话ID
func (r *AssistantAgentRepo) UpdateAccountAssistantConversation(
	ctx context.Context,
	accountID uuid.UUID,
	conversationID *uuid.UUID,
) error {
	return r.dao.UpdateAccountAssistantConversation(ctx, accountID, conversationID)
}

// GetConversationByID 根据ID获取会话信息
func (r *AssistantAgentRepo) GetConversationByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
	return r.dao.GetConversationByID(ctx, id)
}

// GetMessagesByConversationID 根据会话ID获取消息列表
func (r *AssistantAgentRepo) GetMessagesByConversationID(
	ctx context.Context,
	conversationID uuid.UUID,
	page, pageSize int,
	createdAtBefore int64,
) ([]entity.Message, int64, error) {
	return r.dao.GetMessagesByConversationID(ctx, conversationID, page, pageSize, createdAtBefore)
}

// GetAppsByName 根据名称获取应用列表
func (r *AssistantAgentRepo) GetAppsByName(ctx context.Context, name string) ([]entity.App, error) {
	return r.dao.GetAppsByName(ctx, name)
}
