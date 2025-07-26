package entity

import (
	"github.com/crazyfrankie/voidx/pkg/consts"
	"github.com/google/uuid"
)

// WechatConfig Agent微信配置信息
type WechatConfig struct {
	ID              uuid.UUID                 `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AppID           uuid.UUID                 `gorm:"type:uuid;not null;index:wechat_config_app_id_idx" json:"app_id"`
	WechatAppID     string                    `gorm:"size:255;default:''" json:"wechat_app_id"`
	WechatAppSecret string                    `gorm:"size:255;default:''" json:"wechat_app_secret"`
	WechatToken     string                    `gorm:"size:255;default:''" json:"wechat_token"`
	Status          consts.WechatConfigStatus `gorm:"size:255;not null;default:''" json:"status"`
	Utime           int64                     `gorm:"autoUpdateTime" json:"utime"`
	Ctime           int64                     `gorm:"autoCreateTime" json:"ctime"`
}

// WechatEndUser 微信公众号与终端用户标识关联表
type WechatEndUser struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OpenID    string    `gorm:"type:text;not null;index:wechat_end_user_openid_app_id_idx,composite:openid_app" json:"openid"`
	AppID     uuid.UUID `gorm:"type:uuid;not null;index:wechat_end_user_openid_app_id_idx,composite:openid_app" json:"app_id"`
	EndUserID uuid.UUID `gorm:"type:uuid;not null" json:"end_user_id"`
	Utime     int64     `gorm:"autoUpdateTime" json:"utime"`
	Ctime     int64     `gorm:"autoCreateTime" json:"ctime"`
}

// WechatMessage 微信公众号消息模型，用于记录未推送的消息记录
type WechatMessage struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WechatEndUserID uuid.UUID `gorm:"type:uuid;not null;index:wechat_message_wechat_end_user_id_idx" json:"wechat_end_user_id"`
	MessageID       uuid.UUID `gorm:"type:uuid;not null" json:"message_id"`
	IsPushed        bool      `gorm:"not null;default:false" json:"is_pushed"`
	Utime           int64     `gorm:"autoUpdateTime" json:"utime"`
	Ctime           int64     `gorm:"autoCreateTime" json:"ctime"`
}
