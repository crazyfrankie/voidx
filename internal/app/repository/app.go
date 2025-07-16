package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/app/repository/dao"
	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type AppRepo struct {
	dao *dao.AppDao
}

func NewAppRepo(d *dao.AppDao) *AppRepo {
	return &AppRepo{dao: d}
}

// CreateApp 创建应用
func (r *AppRepo) CreateApp(ctx context.Context, app *entity.App) (*entity.App, error) {
	return r.dao.CreateApp(ctx, app)
}

// GetAppByID 根据ID获取应用
func (r *AppRepo) GetAppByID(ctx context.Context, appID uuid.UUID) (*entity.App, error) {
	return r.dao.GetAppByID(ctx, appID)
}

// UpdateApp 更新应用
func (r *AppRepo) UpdateApp(ctx context.Context, app *entity.App) error {
	return r.dao.UpdateApp(ctx, app)
}

// DeleteApp 删除应用
func (r *AppRepo) DeleteApp(ctx context.Context, appID uuid.UUID) error {
	return r.dao.DeleteApp(ctx, appID)
}

// GetAppsWithPage 获取应用分页列表
func (r *AppRepo) GetAppsWithPage(ctx context.Context, accountID uuid.UUID, page int, pageSize int, searchWord string) ([]*entity.App, int64, error) {
	return r.dao.GetAppsWithPage(ctx, accountID, page, pageSize, searchWord)
}

// CreateAppConfigVersion 创建应用配置版本
func (r *AppRepo) CreateAppConfigVersion(ctx context.Context, appID uuid.UUID, version int, configType string, config map[string]interface{}) (*entity.AppConfigVersion, error) {
	return r.dao.CreateAppConfigVersion(ctx, appID, version, configType, config)
}

// GetAppConfigVersion 获取应用配置版本
func (r *AppRepo) GetAppConfigVersion(ctx context.Context, appConfigVersionID uuid.UUID) (*entity.AppConfigVersion, error) {
	return r.dao.GetAppConfigVersion(ctx, appConfigVersionID)
}

// GetDraftAppConfigVersion 获取应用草稿配置版本
func (r *AppRepo) GetDraftAppConfigVersion(ctx context.Context, appID uuid.UUID) (*entity.AppConfigVersion, error) {
	return r.dao.GetDraftAppConfigVersion(ctx, appID)
}

// UpdateAppConfigVersion 更新应用配置版本
func (r *AppRepo) UpdateAppConfigVersion(ctx context.Context, appConfigVersion *entity.AppConfigVersion) error {
	return r.dao.UpdateAppConfigVersion(ctx, appConfigVersion)
}

// GetPublishHistoriesWithPage 获取发布历史分页列表
func (r *AppRepo) GetPublishHistoriesWithPage(ctx context.Context, appID uuid.UUID, page int, pageSize int) ([]*entity.AppConfigVersion, int64, error) {
	return r.dao.GetPublishHistoriesWithPage(ctx, appID, page, pageSize)
}

// GetMaxPublishedVersion 获取最大发布版本号
func (r *AppRepo) GetMaxPublishedVersion(ctx context.Context, appID uuid.UUID) (int, error) {
	return r.dao.GetMaxPublishedVersion(ctx, appID)
}

// CreateAppConfig 创建应用配置
func (r *AppRepo) CreateAppConfig(ctx context.Context, appID uuid.UUID, config map[string]interface{}) (*entity.AppConfig, error) {
	return r.dao.CreateAppConfig(ctx, appID, config)
}

// GetConversation 获取会话
func (r *AppRepo) GetConversation(ctx context.Context, conversationID uuid.UUID) (*entity.Conversation, error) {
	return r.dao.GetConversation(ctx, conversationID)
}

// CreateConversation 创建会话
func (r *AppRepo) CreateConversation(ctx context.Context, accountID, appID uuid.UUID) (*entity.Conversation, error) {
	return r.dao.CreateConversation(ctx, accountID, appID)
}

// UpdateConversation 更新会话
func (r *AppRepo) UpdateConversation(ctx context.Context, conversation *entity.Conversation) error {
	return r.dao.UpdateConversation(ctx, conversation)
}

// CreateMessage 创建消息
func (r *AppRepo) CreateMessage(ctx context.Context, appID, conversationID, createdBy uuid.UUID, invokeFrom, query string, imageUrls []string) (*entity.Message, error) {
	return r.dao.CreateMessage(ctx, appID, conversationID, createdBy, invokeFrom, query, imageUrls)
}

// UpdateMessage 更新消息
func (r *AppRepo) UpdateMessage(ctx context.Context, message *entity.Message) error {
	return r.dao.UpdateMessage(ctx, message)
}

// GetDebugConversationMessagesWithPage 获取调试会话消息分页列表
func (r *AppRepo) GetDebugConversationMessagesWithPage(ctx context.Context, conversationID uuid.UUID, page int, pageSize int, ctime int64) ([]*entity.Message, int64, error) {
	return r.dao.GetDebugConversationMessagesWithPage(ctx, conversationID, page, pageSize, ctime)
}

// CreateAgentThought 创建智能体思考过程
func (r *AppRepo) CreateAgentThought(ctx context.Context, messageID uuid.UUID, event, thought, observation, tool, toolInput, answer string, totalTokenCount int, totalPrice, latency float64) (*entity.AgentThought, error) {
	return r.dao.CreateAgentThought(ctx, messageID, event, thought, observation, tool, toolInput, answer, totalTokenCount, totalPrice, latency)
}

// GetAppDatasetJoins 获取应用关联的知识库
func (r *AppRepo) GetAppDatasetJoins(ctx context.Context, appID uuid.UUID) ([]*entity.AppDatasetJoin, error) {
	return r.dao.GetAppDatasetJoins(ctx, appID)
}

// DeleteAppDatasetJoins 删除应用关联的知识库
func (r *AppRepo) DeleteAppDatasetJoins(ctx context.Context, appID uuid.UUID) error {
	return r.dao.DeleteAppDatasetJoins(ctx, appID)
}

// CreateAppDatasetJoin 创建应用关联知识库
func (r *AppRepo) CreateAppDatasetJoin(ctx context.Context, appID, datasetID uuid.UUID) error {
	return r.dao.CreateAppDatasetJoin(ctx, appID, datasetID)
}
