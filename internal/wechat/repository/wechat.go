package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/wechat/repository/dao"
)

var (
	ErrMessageNotFound = dao.ErrMessageNotFound
	ErrEndUserNotFound = dao.ErrEndUserNotFound
)

type WechatRepository struct {
	dao *dao.WechatDao
}

func NewWechatRepository(dao *dao.WechatDao) *WechatRepository {
	return &WechatRepository{dao: dao}
}

func (r *WechatRepository) GetApp(ctx context.Context, appID uuid.UUID) (*entity.App, error) {
	return r.dao.GetApp(ctx, appID)
}

func (r *WechatRepository) GetWechatConfig(ctx context.Context, appID uuid.UUID) (*entity.WechatConfig, error) {
	return r.dao.GetWechatConfig(ctx, appID)
}

func (r *WechatRepository) CreateEndUser(ctx context.Context, endUser *entity.EndUser) error {
	return r.dao.CreateEndUser(ctx, endUser)
}

func (r *WechatRepository) CreateMessage(ctx context.Context, message *entity.Message) error {
	return r.dao.CreateMessage(ctx, message)
}

func (r *WechatRepository) CreateWechatMessage(ctx context.Context, message *entity.WechatMessage) error {
	return r.dao.CreateWechatMessage(ctx, message)
}

func (r *WechatRepository) CreateWechatEndUser(ctx context.Context, endUser *entity.WechatEndUser) error {
	return r.dao.CreateWechatEndUser(ctx, endUser)
}

func (r *WechatRepository) GetWechatEndUser(ctx context.Context, openid string, appID uuid.UUID) (*entity.WechatEndUser, error) {
	return r.dao.GetWechatEndUser(ctx, openid, appID)
}

func (r *WechatRepository) GetLatestWechatMessage(ctx context.Context, userID uuid.UUID) (*entity.WechatMessage, error) {
	return r.dao.GetLatestWechatMessage(ctx, userID)
}

func (r *WechatRepository) GetConversation(ctx context.Context, appID uuid.UUID, endUserID uuid.UUID) (*entity.Conversation, error) {
	return r.dao.GetConversation(ctx, appID, endUserID)
}

func (r *WechatRepository) GetConversationByID(ctx context.Context, conversationID uuid.UUID) (*entity.Conversation, error) {
	return r.dao.GetConversationByID(ctx, conversationID)
}

func (r *WechatRepository) GetMessage(ctx context.Context, messageID uuid.UUID) (*entity.Message, error) {
	return r.dao.GetMessage(ctx, messageID)
}

func (r *WechatRepository) UpdateWechatMessage(ctx context.Context, messageID uuid.UUID, updates map[string]any) error {
	return r.dao.UpdateWechatMessage(ctx, messageID, updates)
}
