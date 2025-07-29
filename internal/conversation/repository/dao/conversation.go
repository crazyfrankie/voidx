package dao

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
)

type ConversationDao struct {
	db *gorm.DB
}

func NewConversationDao(db *gorm.DB) *ConversationDao {
	return &ConversationDao{db: db}
}

func (d *ConversationDao) GetConversationByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
	var conversation entity.Conversation
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&conversation).Error
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (d *ConversationDao) GetMessagesByConversationID(ctx context.Context, conversationID uuid.UUID, pageReq req.GetConversationMessagesWithPageReq) ([]entity.Message, int64, error) {
	var messages []entity.Message
	var total int64

	query := d.db.WithContext(ctx).Where("conversation_id = ?", conversationID)

	// 添加时间过滤条件
	if pageReq.Ctime > 0 {
		query = query.Where("ctime <= ?", pageReq.Ctime)
	}

	// 计算总数
	if err := query.Model(&entity.Message{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageReq.CurrentPage - 1) * pageReq.PageSize
	err := query.Order("ctime DESC").
		Offset(offset).
		Limit(pageReq.PageSize).
		Find(&messages).Error

	if err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

func (d *ConversationDao) GetMessageByID(ctx context.Context, id uuid.UUID) (*entity.Message, error) {
	var message entity.Message
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (d *ConversationDao) DeleteConversation(ctx context.Context, id uuid.UUID) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除相关的消息和AgentThoughts
		if err := tx.Where("conversation_id = ?", id).Delete(&entity.Message{}).Error; err != nil {
			return err
		}

		if err := tx.Where("conversation_id = ?", id).Delete(&entity.AgentThought{}).Error; err != nil {
			return err
		}

		// 删除会话
		return tx.Where("id = ?", id).Delete(&entity.Conversation{}).Error
	})
}

func (d *ConversationDao) DeleteMessage(ctx context.Context, id uuid.UUID) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除相关的AgentThoughts
		if err := tx.Where("message_id = ?", id).Delete(&entity.AgentThought{}).Error; err != nil {
			return err
		}

		// 删除消息
		return tx.Where("id = ?", id).Delete(&entity.Message{}).Error
	})
}

func (d *ConversationDao) UpdateConversationName(ctx context.Context, id uuid.UUID, name string) error {
	return d.db.WithContext(ctx).Model(&entity.Conversation{}).
		Where("id = ?", id).
		Update("name", name).Error
}

func (d *ConversationDao) UpdateConversationIsPinned(ctx context.Context, id uuid.UUID, isPinned bool) error {
	return d.db.WithContext(ctx).Model(&entity.Conversation{}).
		Where("id = ?", id).
		Update("is_pinned", isPinned).Error
}

func (d *ConversationDao) CreateConversation(ctx context.Context, conversation *entity.Conversation) error {
	return d.db.WithContext(ctx).Create(conversation).Error
}

func (d *ConversationDao) CreateMessage(ctx context.Context, message *entity.Message) error {
	return d.db.WithContext(ctx).Create(message).Error
}

func (d *ConversationDao) UpdateMessage(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return d.db.WithContext(ctx).Model(&entity.Message{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (d *ConversationDao) UpdateConversationSummary(ctx context.Context, conversationID uuid.UUID, summary string) error {
	return d.db.WithContext(ctx).Model(&entity.Conversation{}).
		Where("id = ?", conversationID).Update("summary", summary).Error
}

func (d *ConversationDao) GetConversationsByAccountID(
	ctx context.Context,
	accountID uuid.UUID,
	getReq req.GetConversationsReq,
) ([]entity.Conversation, error) {
	var conversations []entity.Conversation

	query := d.db.WithContext(ctx).Where("account_id = ?", accountID)

	// 添加应用ID过滤
	if getReq.AppID != uuid.Nil {
		query = query.Where("app_id = ?", getReq.AppID)
	}

	// 添加置顶过滤
	if getReq.IsPinned != nil {
		query = query.Where("is_pinned = ?", *getReq.IsPinned)
	}

	// 添加调用来源过滤
	if getReq.InvokeFrom != "" {
		query = query.Where("invoke_from = ?", getReq.InvokeFrom)
	}

	err := query.Order("is_pinned DESC, utime DESC").
		Limit(50). // 限制返回数量
		Find(&conversations).Error

	return conversations, err
}

func (d *ConversationDao) CreateAgentThought(ctx context.Context, agentThought *entity.AgentThought) error {
	return d.db.WithContext(ctx).Create(agentThought).Error
}

func (d *ConversationDao) GetConversationAgentThoughts(ctx context.Context, conversationID uuid.UUID) ([]entity.AgentThought, error) {
	var agentThoughts []entity.AgentThought
	if err := d.db.WithContext(ctx).Model(&entity.AgentThought{}).
		Where("conversation_id = ?", conversationID).
		Find(&agentThoughts).Error; err != nil {
		return nil, err
	}

	return agentThoughts, nil
}
