package dao

import (
	"context"

	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/google/uuid"
)

type WebAppDao struct {
	db *gorm.DB
}

func NewWebAppDao(db *gorm.DB) *WebAppDao {
	return &WebAppDao{db: db}
}

func (d *WebAppDao) GetAppByToken(ctx context.Context, token string) (*entity.App, error) {
	var app entity.App
	err := d.db.WithContext(ctx).
		Where("token = ? AND status = ?", token, "published").
		First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (d *WebAppDao) GetConversationByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
	var conversation entity.Conversation
	err := d.db.WithContext(ctx).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&conversation).Error
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}
