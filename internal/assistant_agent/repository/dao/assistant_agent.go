package dao

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type AssistantAgentDao struct {
	db *gorm.DB
}

func NewAssistantAgentDao(db *gorm.DB) *AssistantAgentDao {
	return &AssistantAgentDao{db: db}
}

// GetAccountByID 根据ID获取账户信息
func (d *AssistantAgentDao) GetAccountByID(ctx context.Context, id uuid.UUID) (*entity.Account, error) {
	var account entity.Account
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// UpdateAccountAssistantConversation 更新账户的辅助Agent会话ID
func (d *AssistantAgentDao) UpdateAccountAssistantConversation(
	ctx context.Context,
	accountID uuid.UUID,
	conversationID *uuid.UUID,
) error {
	return d.db.WithContext(ctx).
		Model(&entity.Account{}).
		Where("id = ?", accountID).
		Update("assistant_agent_conversation_id", conversationID).Error
}

// GetConversationByID 根据ID获取会话信息
func (d *AssistantAgentDao) GetConversationByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
	var conversation entity.Conversation
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&conversation).Error
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

// GetMessagesByConversationID 根据会话ID获取消息列表
func (d *AssistantAgentDao) GetMessagesByConversationID(
	ctx context.Context,
	conversationID uuid.UUID,
	page, pageSize int,
	createdAtBefore int64,
) ([]entity.Message, int64, error) {
	query := d.db.WithContext(ctx).
		Where("conversation_id = ? AND status IN ? AND answer != ?",
			conversationID, []string{"normal", "stop"}, "")

	if createdAtBefore != 0 {
		query = query.Where("ctime <= ?", createdAtBefore)
	}

	// 获取总数
	var total int64
	if err := query.Model(&entity.Message{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	var messages []entity.Message
	offset := (page - 1) * pageSize
	err := query.Order("ctime DESC").Offset(offset).Limit(pageSize).Find(&messages).Error
	if err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// GetAppsByName 根据名称获取应用列表
func (d *AssistantAgentDao) GetAppsByName(ctx context.Context, name string) ([]entity.App, error) {
	var apps []entity.App
	err := d.db.WithContext(ctx).Where("name = ?", name).Find(&apps).Error
	if err != nil {
		return nil, err
	}
	return apps, nil
}
