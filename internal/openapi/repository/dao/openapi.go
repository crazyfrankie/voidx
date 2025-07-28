package dao

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type OpenAPIDao struct {
	db *gorm.DB
}

func NewOpenAPIDao(db *gorm.DB) *OpenAPIDao {
	return &OpenAPIDao{db: db}
}

// GetAppByID 根据ID获取应用
func (d *OpenAPIDao) GetAppByID(ctx context.Context, id uuid.UUID) (*entity.App, error) {
	var app entity.App
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// GetEndUserByID 根据ID获取终端用户
func (d *OpenAPIDao) GetEndUserByID(ctx context.Context, id uuid.UUID) (*entity.EndUser, error) {
	var endUser entity.EndUser
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&endUser).Error
	if err != nil {
		return nil, err
	}
	return &endUser, nil
}

// CreateEndUser 创建终端用户
func (d *OpenAPIDao) CreateEndUser(ctx context.Context, endUser *entity.EndUser) error {
	return d.db.WithContext(ctx).Create(endUser).Error
}

// GetConversationByID 根据ID获取会话
func (d *OpenAPIDao) GetConversationByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
	var conversation entity.Conversation
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&conversation).Error
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}
