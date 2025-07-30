package entity

import (
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		// Account 相关表
		&Account{},
		&AccountOAuth{},

		// API 相关表
		&ApiKey{},
		&ApiToolProvider{},
		&ApiTool{},

		// App 相关表
		&App{},
		&AppConfig{},
		&AppConfigVersion{},

		// Conversation 相关表
		&Conversation{},
		&Message{},
		&AgentThought{},

		// Dataset 相关表
		&Dataset{},
		&Document{},
		&Segment{},
		&Keyword{},
		&DatasetQuery{},
		&ProcessRule{},

		// EndUser 相关表
		&EndUser{},

		// Platform 相关表
		&WechatConfig{},
		&WechatEndUser{},
		&WechatMessage{},

		// Upload 相关表
		&UploadFile{},

		// Workflow 相关表
		&Workflow{},
		&WorkflowResult{},
	)
}
