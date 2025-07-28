package dao

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type AIDao struct {
	db *gorm.DB
}

func NewAIDao(db *gorm.DB) *AIDao {
	return &AIDao{db: db}
}

// GetMessageByID 根据ID获取消息
func (d *AIDao) GetMessageByID(ctx context.Context, id uuid.UUID) (*entity.Message, error) {
	var message entity.Message
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}
