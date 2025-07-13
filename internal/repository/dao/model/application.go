package model

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(&App{}, AppConfig{}, &Conversation{}, &Message{}, &Account{})
}

type App struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AccountID            uuid.UUID  `gorm:"type:uuid;index:idx_app_account_id;not null" json:"accountId"`
	PublishedAppConfigID *uuid.UUID `gorm:"type:uuid" json:"publishedAppConfigId"` // 可为空
	DraftedAppConfigID   *uuid.UUID `gorm:"type:uuid" json:"draftedAppConfigId"`   // 可为空
	DebugConversationID  uuid.UUID  `gorm:"type:uuid;not null" json:"debugConversationId"`
	Name                 string     `gorm:"size:255;not null" json:"name"`
	Icon                 string     `gorm:"size:255;not null" json:"icon"`
	Description          string     `gorm:"type:text;not null" json:"description"`
	Status               string     `gorm:"type:varchar(100);not null" json:"status"`
	Utime                int64      `gorm:"autoUpdateTime:milli" json:"utime"`
	Ctime                int64      `gorm:"autoCreateTime:milli" json:"ctime"`
}

type AppConfig struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AppID       uuid.UUID      `gorm:"type:uuid;index;not null" json:"appId"`
	ModelConfig datatypes.JSON `gorm:"type:jsonb" json:"modelConfig"` // 使用 gorm 的 datatypes.JSON
	MemoryMode  string         `gorm:"type:varchar(255);not null" json:"memoryMode"`
	Status      string         `gorm:"type:varchar(100);not null" json:"status"`
	Utime       int64          `gorm:"autoUpdateTime:milli" json:"utime"`
	Ctime       int64          `gorm:"autoCreateTime:milli" json:"ctime"`
}

type Conversation struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	AccountID uuid.UUID `gorm:"type:uuid;index;not null" json:"accountId"`
	AppID     uuid.UUID `gorm:"type:uuid;index;not null" json:"appId"`
	Utime     int64     `gorm:"autoUpdateTime:milli" json:"utime"`
	Ctime     int64     `gorm:"autoCreateTime:milli" json:"ctime"`
}

type Message struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ConversationID  uuid.UUID `gorm:"type:uuid;index;not null" json:"conversationId"`
	AccountID       uuid.UUID `gorm:"type:uuid;index;not null" json:"accountId"`
	Query           string    `gorm:"type:text;not null" json:"query"`
	Answer          string    `gorm:"type:text;not null" json:"answer"`
	AnswerTokens    int       `gorm:"not null;default:0" json:"answerTokens"`
	ResponseLatency float64   `gorm:"not null;default:0" json:"responseLatency"`
	Utime           int64     `gorm:"autoUpdateTime:milli" json:"utime"`
	Ctime           int64     `gorm:"autoCreateTime:milli" json:"ctime"`
}

type Account struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name     string    `gorm:"size:255;not null" json:"name"`
	Email    string    `gorm:"size:255;not null;unique" json:"email"`
	Password string    `gorm:"size:255;not null" json:"-"`
	Utime    int64     `gorm:"autoUpdateTime:milli" json:"utime"`
	Ctime    int64     `gorm:"autoCreateTime:milli" json:"ctime"`
}
