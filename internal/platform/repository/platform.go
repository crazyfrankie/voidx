package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/platform/repository/dao"
)

type PlatformRepo struct {
	dao *dao.PlatformDao
}

func NewPlatformRepo(d *dao.PlatformDao) *PlatformRepo {
	return &PlatformRepo{dao: d}
}

// GetAppByID 根据ID获取应用信息
func (r *PlatformRepo) GetAppByID(ctx context.Context, id uuid.UUID) (*entity.App, error) {
	return r.dao.GetAppByID(ctx, id)
}

// GetWechatConfigByAppID 根据应用ID获取微信配置
func (r *PlatformRepo) GetWechatConfigByAppID(ctx context.Context, appID uuid.UUID) (*entity.WechatConfig, error) {
	return r.dao.GetWechatConfigByAppID(ctx, appID)
}

// CreateWechatConfig 创建微信配置
func (r *PlatformRepo) CreateWechatConfig(ctx context.Context, wechatConfig *entity.WechatConfig) error {
	return r.dao.CreateWechatConfig(ctx, wechatConfig)
}

// UpdateWechatConfig 更新微信配置
func (r *PlatformRepo) UpdateWechatConfig(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return r.dao.UpdateWechatConfig(ctx, id, updates)
}

// GetWechatEndUserByOpenIDAndAppID 根据OpenID和应用ID获取微信终端用户
func (r *PlatformRepo) GetWechatEndUserByOpenIDAndAppID(ctx context.Context, openID string, appID uuid.UUID) (*entity.WechatEndUser, error) {
	return r.dao.GetWechatEndUserByOpenIDAndAppID(ctx, openID, appID)
}

// CreateWechatEndUser 创建微信终端用户
func (r *PlatformRepo) CreateWechatEndUser(ctx context.Context, wechatEndUser *entity.WechatEndUser) error {
	return r.dao.CreateWechatEndUser(ctx, wechatEndUser)
}

// CreateWechatMessage 创建微信消息记录
func (r *PlatformRepo) CreateWechatMessage(ctx context.Context, wechatMessage *entity.WechatMessage) error {
	return r.dao.CreateWechatMessage(ctx, wechatMessage)
}

// GetUnpushedWechatMessages 获取未推送的微信消息
func (r *PlatformRepo) GetUnpushedWechatMessages(ctx context.Context, wechatEndUserID uuid.UUID) ([]entity.WechatMessage, error) {
	return r.dao.GetUnpushedWechatMessages(ctx, wechatEndUserID)
}

// UpdateWechatMessagePushed 更新微信消息推送状态
func (r *PlatformRepo) UpdateWechatMessagePushed(ctx context.Context, id uuid.UUID, isPushed bool) error {
	return r.dao.UpdateWechatMessagePushed(ctx, id, isPushed)
}
