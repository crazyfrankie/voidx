package entity

import (
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&App{}, &AppConfigVersion{}, AppConfig{}, &Conversation{}, &Message{},
		&AgentThought{}, &AppDatasetJoin{}, &Account{}, &UploadFile{})
}
