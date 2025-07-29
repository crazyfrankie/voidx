package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/conversation/repository/dao"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
)

type ConversationRepo struct {
	dao *dao.ConversationDao
}

func NewConversationRepo(d *dao.ConversationDao) *ConversationRepo {
	return &ConversationRepo{dao: d}
}

func (r *ConversationRepo) GetConversationByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
	return r.dao.GetConversationByID(ctx, id)
}

func (r *ConversationRepo) GetMessagesByConversationID(
	ctx context.Context,
	conversationID uuid.UUID,
	pageReq req.GetConversationMessagesWithPageReq,
) ([]entity.Message, int64, error) {
	return r.dao.GetMessagesByConversationID(ctx, conversationID, pageReq)
}

func (r *ConversationRepo) GetMessageByID(ctx context.Context, id uuid.UUID) (*entity.Message, error) {
	return r.dao.GetMessageByID(ctx, id)
}

func (r *ConversationRepo) DeleteConversation(ctx context.Context, id uuid.UUID) error {
	return r.dao.DeleteConversation(ctx, id)
}

func (r *ConversationRepo) DeleteMessage(ctx context.Context, id uuid.UUID) error {
	return r.dao.DeleteMessage(ctx, id)
}

func (r *ConversationRepo) UpdateConversationName(ctx context.Context, id uuid.UUID, name string) error {
	return r.dao.UpdateConversationName(ctx, id, name)
}

func (r *ConversationRepo) UpdateConversationIsPinned(ctx context.Context, id uuid.UUID, isPinned bool) error {
	return r.dao.UpdateConversationIsPinned(ctx, id, isPinned)
}

func (r *ConversationRepo) CreateConversation(ctx context.Context, conversation *entity.Conversation) error {
	return r.dao.CreateConversation(ctx, conversation)
}

func (r *ConversationRepo) CreateMessage(ctx context.Context, message *entity.Message) error {
	return r.dao.CreateMessage(ctx, message)
}

func (r *ConversationRepo) UpdateMessage(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return r.dao.UpdateMessage(ctx, id, updates)
}

func (r *ConversationRepo) UpdateConversationSummary(ctx context.Context, conversationID uuid.UUID, summary string) error {
	return r.dao.UpdateConversationSummary(ctx, conversationID, summary)
}

func (r *ConversationRepo) GetConversationsByAccountID(
	ctx context.Context,
	accountID uuid.UUID,
	getReq req.GetConversationsReq,
) ([]entity.Conversation, error) {
	return r.dao.GetConversationsByAccountID(ctx, accountID, getReq)
}

func (r *ConversationRepo) CreateAgentThought(ctx context.Context, agentThought *entity.AgentThought) error {
	return r.dao.CreateAgentThought(ctx, agentThought)
}

func (r *ConversationRepo) GetConversationAgentThoughts(ctx context.Context, conversationID uuid.UUID) ([]entity.AgentThought, error) {
	return r.dao.GetConversationAgentThoughts(ctx, conversationID)
}
