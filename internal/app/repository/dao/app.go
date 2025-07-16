package dao

import (
	"context"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type AppDao struct {
	db *gorm.DB
}

func NewAppDao(db *gorm.DB) *AppDao {
	return &AppDao{db: db}
}

// CreateApp 创建应用
func (d *AppDao) CreateApp(ctx context.Context, app *entity.App) (*entity.App, error) {
	if err := d.db.WithContext(ctx).Create(app).Error; err != nil {
		return nil, err
	}

	return app, nil
}

// GetAppByID 根据ID获取应用
func (d *AppDao) GetAppByID(ctx context.Context, appID uuid.UUID) (*entity.App, error) {
	var app entity.App
	if err := d.db.WithContext(ctx).First(&app, "id = ?", appID).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

// UpdateApp 更新应用
func (d *AppDao) UpdateApp(ctx context.Context, app *entity.App) error {
	return d.db.WithContext(ctx).Save(app).Error
}

// DeleteApp 删除应用
func (d *AppDao) DeleteApp(ctx context.Context, appID uuid.UUID) error {
	return d.db.WithContext(ctx).Delete(&entity.App{}, "id = ?", appID).Error
}

// GetAppsWithPage 获取应用分页列表
func (d *AppDao) GetAppsWithPage(ctx context.Context, accountID uuid.UUID, page int, pageSize int, searchWord string) ([]*entity.App, int64, error) {
	var apps []*entity.App
	var total int64

	query := d.db.WithContext(ctx).Model(&entity.App{}).Where("account_id = ?", accountID)

	if searchWord != "" {
		query = query.Where("name LIKE ?", "%"+searchWord+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := query.Order("ctime DESC").Offset(offset).Limit(pageSize).Find(&apps).Error; err != nil {
		return nil, 0, err
	}

	return apps, total, nil
}

// CreateAppConfigVersion 创建应用配置版本
func (d *AppDao) CreateAppConfigVersion(ctx context.Context, appID uuid.UUID, version int, configType string, config map[string]interface{}) (*entity.AppConfigVersion, error) {
	appConfigVersion := &entity.AppConfigVersion{
		AppID:      appID,
		Version:    version,
		ConfigType: configType,
		Config:     config,
	}

	if err := d.db.WithContext(ctx).Create(appConfigVersion).Error; err != nil {
		return nil, err
	}

	return appConfigVersion, nil
}

// GetAppConfigVersion 获取应用配置版本
func (d *AppDao) GetAppConfigVersion(ctx context.Context, appConfigVersionID uuid.UUID) (*entity.AppConfigVersion, error) {
	var appConfigVersion entity.AppConfigVersion
	if err := d.db.WithContext(ctx).First(&appConfigVersion, "id = ?", appConfigVersionID).Error; err != nil {
		return nil, err
	}
	return &appConfigVersion, nil
}

// GetDraftAppConfigVersion 获取应用草稿配置版本
func (d *AppDao) GetDraftAppConfigVersion(ctx context.Context, appID uuid.UUID) (*entity.AppConfigVersion, error) {
	var appConfigVersion entity.AppConfigVersion
	if err := d.db.WithContext(ctx).First(&appConfigVersion, "app_id = ? AND config_type = ?", appID, "draft").Error; err != nil {
		return nil, err
	}
	return &appConfigVersion, nil
}

// UpdateAppConfigVersion 更新应用配置版本
func (d *AppDao) UpdateAppConfigVersion(ctx context.Context, appConfigVersion *entity.AppConfigVersion) error {
	return d.db.WithContext(ctx).Save(appConfigVersion).Error
}

// GetPublishHistoriesWithPage 获取发布历史分页列表
func (d *AppDao) GetPublishHistoriesWithPage(ctx context.Context, appID uuid.UUID, page int, pageSize int) ([]*entity.AppConfigVersion, int64, error) {
	var appConfigVersions []*entity.AppConfigVersion
	var total int64

	query := d.db.WithContext(ctx).Model(&entity.AppConfigVersion{}).
		Where("app_id = ? AND config_type = ?", appID, "published")

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := query.Order("version DESC").Offset(offset).Limit(pageSize).Find(&appConfigVersions).Error; err != nil {
		return nil, 0, err
	}

	return appConfigVersions, total, nil
}

// GetMaxPublishedVersion 获取最大发布版本号
func (d *AppDao) GetMaxPublishedVersion(ctx context.Context, appID uuid.UUID) (int, error) {
	var maxVersion int
	err := d.db.WithContext(ctx).Model(&entity.AppConfigVersion{}).
		Where("app_id = ? AND config_type = ?", appID, "published").
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVersion).Error
	return maxVersion, err
}

// CreateAppConfig 创建应用配置
func (d *AppDao) CreateAppConfig(ctx context.Context, appID uuid.UUID, config map[string]interface{}) (*entity.AppConfig, error) {
	data, err := sonic.Marshal(&config)
	if err != nil {
		return nil, err
	}
	appConfig := &entity.AppConfig{
		AppID:       appID,
		ModelConfig: data,
	}

	if err := d.db.WithContext(ctx).Create(appConfig).Error; err != nil {
		return nil, err
	}

	return appConfig, nil
}

// GetConversation 获取会话
func (d *AppDao) GetConversation(ctx context.Context, conversationID uuid.UUID) (*entity.Conversation, error) {
	var conversation entity.Conversation
	if err := d.db.WithContext(ctx).First(&conversation, "id = ?", conversationID).Error; err != nil {
		return nil, err
	}
	return &conversation, nil
}

// CreateConversation 创建会话
func (d *AppDao) CreateConversation(ctx context.Context, accountID, appID uuid.UUID) (*entity.Conversation, error) {
	conversation := &entity.Conversation{
		AccountID: accountID,
		AppID:     appID,
		Summary:   "",
	}

	if err := d.db.WithContext(ctx).Create(conversation).Error; err != nil {
		return nil, err
	}

	return conversation, nil
}

// UpdateConversation 更新会话
func (d *AppDao) UpdateConversation(ctx context.Context, conversation *entity.Conversation) error {
	return d.db.WithContext(ctx).Save(conversation).Error
}

// CreateMessage 创建消息
func (d *AppDao) CreateMessage(ctx context.Context, appID, conversationID, createdBy uuid.UUID, invokeFrom, query string, imageUrls []string) (*entity.Message, error) {
	message := &entity.Message{
		AppID:          appID,
		ConversationID: conversationID,
		InvokeFrom:     invokeFrom,
		CreatedBy:      createdBy,
		Query:          query,
		ImageUrls:      imageUrls,
		Status:         "normal",
	}

	if err := d.db.WithContext(ctx).Create(message).Error; err != nil {
		return nil, err
	}

	return message, nil
}

// UpdateMessage 更新消息
func (d *AppDao) UpdateMessage(ctx context.Context, message *entity.Message) error {
	return d.db.WithContext(ctx).Save(message).Error
}

// GetDebugConversationMessagesWithPage 获取调试会话消息分页列表
func (d *AppDao) GetDebugConversationMessagesWithPage(ctx context.Context, conversationID uuid.UUID, page int, pageSize int, ctime int64) ([]*entity.Message, int64, error) {
	var messages []*entity.Message
	var total int64

	query := d.db.WithContext(ctx).Model(&entity.Message{}).
		Where("conversation_id = ? AND status IN (?) AND answer != ''", conversationID, []string{"stop", "normal"})

	if ctime > 0 {
		createdAt := time.Unix(ctime, 0)
		query = query.Where("ctime <= ?", createdAt)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := query.Order("ctime DESC").Offset(offset).Limit(pageSize).
		Preload("AgentThoughts").Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// CreateAgentThought 创建智能体思考过程
func (d *AppDao) CreateAgentThought(ctx context.Context, messageID uuid.UUID, event, thought, observation, tool, toolInput, answer string, totalTokenCount int, totalPrice, latency float64) (*entity.AgentThought, error) {
	agentThought := &entity.AgentThought{
		MessageID:       messageID,
		Event:           event,
		Thought:         thought,
		Observation:     observation,
		Tool:            tool,
		ToolInput:       toolInput,
		Answer:          answer,
		TotalTokenCount: totalTokenCount,
		TotalPrice:      totalPrice,
		Latency:         latency,
	}

	if err := d.db.WithContext(ctx).Create(agentThought).Error; err != nil {
		return nil, err
	}

	return agentThought, nil
}

// GetAppDatasetJoins 获取应用关联的知识库
func (d *AppDao) GetAppDatasetJoins(ctx context.Context, appID uuid.UUID) ([]*entity.AppDatasetJoin, error) {
	var appDatasetJoins []*entity.AppDatasetJoin
	if err := d.db.WithContext(ctx).Where("app_id = ?", appID).Find(&appDatasetJoins).Error; err != nil {
		return nil, err
	}
	return appDatasetJoins, nil
}

// DeleteAppDatasetJoins 删除应用关联的知识库
func (d *AppDao) DeleteAppDatasetJoins(ctx context.Context, appID uuid.UUID) error {
	return d.db.WithContext(ctx).Where("app_id = ?", appID).Delete(&entity.AppDatasetJoin{}).Error
}

// CreateAppDatasetJoin 创建应用关联知识库
func (d *AppDao) CreateAppDatasetJoin(ctx context.Context, appID, datasetID uuid.UUID) error {
	return d.db.WithContext(ctx).Create(&entity.AppDatasetJoin{
		AppID:     appID,
		DatasetID: datasetID,
	}).Error
}
