package dao

import (
	"context"
	"errors"
	"fmt"
	"github.com/crazyfrankie/voidx/pkg/consts"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type WechatDao struct {
	db *gorm.DB
}

func NewWechatDao(db *gorm.DB) *WechatDao {
	return &WechatDao{db: db}
}

func (d *WechatDao) GetApp(ctx context.Context, appID uuid.UUID) (*entity.App, error) {
	var app entity.App
	err := d.db.WithContext(ctx).Preload("WechatConfig").First(&app, "id = ?", appID).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (d *WechatDao) GetWechatConfig(ctx context.Context, appID uuid.UUID) (*entity.WechatConfig, error) {
	var config entity.WechatConfig
	err := d.db.WithContext(ctx).Model(&entity.WechatConfig{}).First(&config, "app_id = ?", appID).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (d *WechatDao) CreateEndUser(ctx context.Context, endUser *entity.EndUser) error {
	return d.db.WithContext(ctx).Create(endUser).Error
}

func (d *WechatDao) CreateWechatEndUser(ctx context.Context, endUser *entity.WechatEndUser) error {
	return d.db.WithContext(ctx).Create(endUser).Error
}

func (d *WechatDao) GetWechatEndUser(ctx context.Context, openid string, appID uuid.UUID) (*entity.WechatEndUser, error) {
	var wechatEndUser entity.WechatEndUser
	err := d.db.WithContext(ctx).Model(&entity.WechatEndUser{}).First(&wechatEndUser, "openid = ? AND app_id = ?", openid, appID).Error
	if err != nil {
		return nil, err
	}
	return &wechatEndUser, nil
}

func (d *WechatDao) GetLatestWechatMessage(ctx context.Context, wechatEndUserID uuid.UUID) (*entity.WechatMessage, error) {
	var wechatMessage entity.WechatMessage
	err := d.db.WithContext(ctx).Model(&entity.WechatMessage{}).Where("wechat_end_user_id = ?", wechatEndUserID).
		Order("ctime DESC").First(&wechatMessage).Error
	if err != nil {
		return nil, err
	}
	return &wechatMessage, nil
}

func (d *WechatDao) GetMessage(ctx context.Context, messageID uuid.UUID) (*entity.Message, error) {
	var message entity.Message
	err := d.db.WithContext(ctx).Model(&entity.Message{}).First(&message, "id = ?", messageID).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (d *WechatDao) CreateMessage(ctx context.Context, message *entity.Message) error {
	return d.db.WithContext(ctx).Create(message).Error
}

func (d *WechatDao) CreateWechatMessage(ctx context.Context, wechatMessage *entity.WechatMessage) error {
	return d.db.WithContext(ctx).Create(wechatMessage).Error
}

func (d *WechatDao) UpdateWechatMessage(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return d.db.WithContext(ctx).Model(&entity.WechatMessage{}).Where("id = ?", id).Updates(updates).Error
}

func (d *WechatDao) GetConversation(ctx context.Context, appID uuid.UUID, endUserID uuid.UUID) (*entity.Conversation, error) {
	// 1. 查询会话记录
	var conversation *entity.Conversation
	err := d.db.WithContext(ctx).
		Where("created_by = ? AND invoke_from = ? AND is_deleted = ?", endUserID, consts.InvokeFromServiceAPI, false).
		First(&conversation).Error

	// 2. 判断会话是否存在，不存在则创建
	if errors.Is(err, gorm.ErrRecordNotFound) {
		conversation = &entity.Conversation{
			AppID:      appID,
			Name:       "New Conversation",
			InvokeFrom: consts.InvokeFromServiceAPI,
			CreatedBy:  endUserID,
		}

		if err := d.db.WithContext(ctx).Model(&entity.Conversation{}).Create(&conversation).Error; err != nil {
			return nil, fmt.Errorf("创建会话失败: %w", err)
		}
		return conversation, nil
	} else if err != nil {
		return nil, fmt.Errorf("查询会话失败: %w", err)
	}

	return conversation, nil
}

func (d *WechatDao) GetConversationByID(ctx context.Context, conversationID uuid.UUID) (*entity.Conversation, error) {
	var conversation entity.Conversation
	err := d.db.WithContext(ctx).First(&conversation, "id = ?", conversationID).Error
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (d *WechatDao) CreateConversation(ctx context.Context, conversation *entity.Conversation) error {
	return d.db.WithContext(ctx).Create(conversation).Error
}
