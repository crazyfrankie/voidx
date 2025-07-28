package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/ai/repository/dao"
	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type AIRepo struct {
	dao *dao.AIDao
}

func NewAIRepo(d *dao.AIDao) *AIRepo {
	return &AIRepo{dao: d}
}

// GetMessageByID 根据ID获取消息
func (r *AIRepo) GetMessageByID(ctx context.Context, id uuid.UUID) (*entity.Message, error) {
	return r.dao.GetMessageByID(ctx, id)
}
