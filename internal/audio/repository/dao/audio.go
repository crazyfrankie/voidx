package dao

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type AudioDao struct {
	db *gorm.DB
}

func NewAudioDao(db *gorm.DB) *AudioDao {
	return &AudioDao{db: db}
}

// GetMessageByID 根据ID获取消息
func (d *AudioDao) GetMessageByID(ctx context.Context, id uuid.UUID) (*entity.Message, error) {
	var message entity.Message
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// GetConversationByID 根据ID获取会话
func (d *AudioDao) GetConversationByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
	var conversation entity.Conversation
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&conversation).Error
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

// GetAppByID 根据ID获取应用
func (d *AudioDao) GetAppByID(ctx context.Context, id uuid.UUID) (*entity.App, error) {
	var app entity.App
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// GetAppConfig 获取应用配置
func (d *AudioDao) GetAppConfig(ctx context.Context, appID uuid.UUID, isDraft bool) (map[string]any, error) {
	var app entity.App
	err := d.db.WithContext(ctx).Where("id = ?", appID).First(&app).Error
	if err != nil {
		return nil, err
	}

	var configID uuid.UUID
	if isDraft {
		configID = app.DraftAppConfigID
	} else {
		configID = app.AppConfigID
	}

	if configID == uuid.Nil {
		return make(map[string]any), nil
	}

	var appConfig entity.AppConfig
	err = d.db.WithContext(ctx).Where("id = ?", configID).First(&appConfig).Error
	if err != nil {
		return nil, err
	}

	// 构建配置映射
	config := make(map[string]any)

	// 解析TextToSpeech配置
	if len(appConfig.TextToSpeech) > 0 {
		config["text_to_speech"] = appConfig.TextToSpeech
	}

	return config, nil
}
