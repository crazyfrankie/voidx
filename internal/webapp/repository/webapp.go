package repository

import (
	"context"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/webapp/repository/dao"
	"github.com/google/uuid"
)

type WebAppRepo struct {
	dao *dao.WebAppDao
}

func NewWebAppRepo(d *dao.WebAppDao) *WebAppRepo {
	return &WebAppRepo{dao: d}
}

func (r *WebAppRepo) GetAppByToken(ctx context.Context, token string) (*entity.App, error) {
	return r.dao.GetAppByToken(ctx, token)
}

func (r *WebAppRepo) GetConversationByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
	return r.dao.GetConversationByID(ctx, id)
}
