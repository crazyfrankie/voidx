package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/audio/repository/dao"
	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type AudioRepo struct {
	dao *dao.AudioDao
}

func NewAudioRepo(d *dao.AudioDao) *AudioRepo {
	return &AudioRepo{dao: d}
}

// GetMessageByID 根据ID获取消息
func (r *AudioRepo) GetMessageByID(ctx context.Context, id uuid.UUID) (*entity.Message, error) {
	return r.dao.GetMessageByID(ctx, id)
}

// GetConversationByID 根据ID获取会话
func (r *AudioRepo) GetConversationByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
	return r.dao.GetConversationByID(ctx, id)
}

// GetAppByID 根据ID获取应用
func (r *AudioRepo) GetAppByID(ctx context.Context, id uuid.UUID) (*entity.App, error) {
	return r.dao.GetAppByID(ctx, id)
}

// GetAppConfig 获取应用配置
func (r *AudioRepo) GetAppConfig(ctx context.Context, appID uuid.UUID, isDraft bool) (map[string]any, error) {
	return r.dao.GetAppConfig(ctx, appID, isDraft)
}
