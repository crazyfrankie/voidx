package dao

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type PlatformDao struct {
	db *gorm.DB
}

func NewPlatformDao(db *gorm.DB) *PlatformDao {
	return &PlatformDao{db: db}
}

// GetAppByID 根据ID获取应用信息
func (d *PlatformDao) GetAppByID(ctx context.Context, id uuid.UUID) (*entity.App, error) {
	var app entity.App
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// GetWechatConfigByAppID 根据应用ID获取微信配置
func (d *PlatformDao) GetWechatConfigByAppID(ctx context.Context, appID uuid.UUID) (*entity.WechatConfig, error) {
	var wechatConfig entity.WechatConfig
	err := d.db.WithContext(ctx).Where("app_id = ?", appID).First(&wechatConfig).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &wechatConfig, nil
}

// CreateWechatConfig 创建微信配置
func (d *PlatformDao) CreateWechatConfig(ctx context.Context, wechatConfig *entity.WechatConfig) error {
	return d.db.WithContext(ctx).Create(wechatConfig).Error
}

// UpdateWechatConfig 更新微信配置
func (d *PlatformDao) UpdateWechatConfig(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return d.db.WithContext(ctx).Model(&entity.WechatConfig{}).Where("id = ?", id).Updates(updates).Error
}

// GetWechatEndUserByOpenIDAndAppID 根据OpenID和应用ID获取微信终端用户
func (d *PlatformDao) GetWechatEndUserByOpenIDAndAppID(ctx context.Context, openID string, appID uuid.UUID) (*entity.WechatEndUser, error) {
	var wechatEndUser entity.WechatEndUser
	err := d.db.WithContext(ctx).Where("openid = ? AND app_id = ?", openID, appID).First(&wechatEndUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &wechatEndUser, nil
}

// CreateWechatEndUser 创建微信终端用户
func (d *PlatformDao) CreateWechatEndUser(ctx context.Context, wechatEndUser *entity.WechatEndUser) error {
	return d.db.WithContext(ctx).Create(wechatEndUser).Error
}

// CreateWechatMessage 创建微信消息记录
func (d *PlatformDao) CreateWechatMessage(ctx context.Context, wechatMessage *entity.WechatMessage) error {
	return d.db.WithContext(ctx).Create(wechatMessage).Error
}

// GetUnpushedWechatMessages 获取未推送的微信消息
func (d *PlatformDao) GetUnpushedWechatMessages(ctx context.Context, wechatEndUserID uuid.UUID) ([]entity.WechatMessage, error) {
	var messages []entity.WechatMessage
	err := d.db.WithContext(ctx).
		Where("wechat_end_user_id = ? AND is_pushed = ?", wechatEndUserID, false).
		Order("ctime ASC").
		Find(&messages).Error
	return messages, err
}

// UpdateWechatMessagePushed 更新微信消息推送状态
func (d *PlatformDao) UpdateWechatMessagePushed(ctx context.Context, id uuid.UUID, isPushed bool) error {
	return d.db.WithContext(ctx).Model(&entity.WechatMessage{}).
		Where("id = ?", id).
		Update("is_pushed", isPushed).Error
}
