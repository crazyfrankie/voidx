package entity

import "github.com/google/uuid"

// Account 账号模型
type Account struct {
	ID                           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name                         string    `gorm:"size:255;not null;default:''" json:"name"`
	Email                        string    `gorm:"size:255;not null;default:'';index:account_email_idx" json:"email"`
	Avatar                       string    `gorm:"size:255;not null;default:''" json:"avatar"`
	Password                     string    `gorm:"size:255;default:''" json:"-"`
	AssistantAgentConversationID uuid.UUID `gorm:"type:uuid" json:"assistant_agent_conversation_id"`
	LastLoginAt                  int64     `gorm:"autoCreateTime" json:"last_login_at"`
	LastLoginIP                  string    `gorm:"size:255;not null;default:''" json:"last_login_ip"`
	Utime                        int64     `gorm:"autoUpdateTime" json:"utime"`
	Ctime                        int64     `gorm:"autoCreateTime" json:"ctime"`
}

// AccountOAuth 账号与第三方授权认证记录表
type AccountOAuth struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AccountID      uuid.UUID `gorm:"type:uuid;not null;index:account_oauth_account_id_idx" json:"account_id"`
	Provider       string    `gorm:"size:255;not null;default:'';index:account_oauth_openid_provider_idx,composite:openid_provider" json:"provider"`
	OpenID         string    `gorm:"size:255;not null;default:'';index:account_oauth_openid_provider_idx,composite:openid_provider" json:"openid"`
	EncryptedToken string    `gorm:"size:255;not null;default:''" json:"encrypted_token"`
	Utime          int64     `gorm:"autoUpdateTime" json:"utime"`
	Ctime          int64     `gorm:"autoCreateTime" json:"ctime"`
}
