package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/openapi/repository/dao"
)

type OpenAPIRepo struct {
	dao *dao.OpenAPIDao
}

func NewOpenAPIRepo(d *dao.OpenAPIDao) *OpenAPIRepo {
	return &OpenAPIRepo{dao: d}
}

// GetAppByID 根据ID获取应用
func (r *OpenAPIRepo) GetAppByID(ctx context.Context, id uuid.UUID) (*entity.App, error) {
	return r.dao.GetAppByID(ctx, id)
}

// GetEndUserByID 根据ID获取终端用户
func (r *OpenAPIRepo) GetEndUserByID(ctx context.Context, id uuid.UUID) (*entity.EndUser, error) {
	return r.dao.GetEndUserByID(ctx, id)
}

// CreateEndUser 创建终端用户
func (r *OpenAPIRepo) CreateEndUser(ctx context.Context, endUser *entity.EndUser) error {
	return r.dao.CreateEndUser(ctx, endUser)
}

// GetConversationByID 根据ID获取会话
func (r *OpenAPIRepo) GetConversationByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
	return r.dao.GetConversationByID(ctx, id)
}
