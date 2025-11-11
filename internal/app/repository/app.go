package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/app/repository/dao"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/types/consts"
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
func (r *AppRepo) UpdateApp(ctx context.Context, appID uuid.UUID, updates map[string]any) error {
	return r.dao.UpdateApp(ctx, appID, updates)
}

func (r *AppRepo) UpdatesApp(ctx context.Context, appID uuid.UUID, updates map[string]any) error {
	return r.dao.UpdatesApp(ctx, appID, updates)
}

// DeleteApp 删除应用
func (r *AppRepo) DeleteApp(ctx context.Context, appID uuid.UUID) error {
	return r.dao.DeleteApp(ctx, appID)
}

// GetAppsWithPage 获取应用分页列表
func (r *AppRepo) GetAppsWithPage(ctx context.Context, accountID uuid.UUID, page int, pageSize int, searchWord string) ([]*entity.App, int64, int64, error) {
	return r.dao.GetAppsWithPage(ctx, accountID, page, pageSize, searchWord)
}

// CreateAppConfigVersion 创建应用配置版本
func (r *AppRepo) CreateAppConfigVersion(ctx context.Context, appCfgVer *entity.AppConfigVersion) (*entity.AppConfigVersion, error) {
	return r.dao.CreateAppConfigVersion(ctx, appCfgVer)
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
func (r *AppRepo) UpdateAppConfigVersion(ctx context.Context, appConfigVersionID uuid.UUID, updates map[string]any) error {
	return r.dao.UpdateAppConfigVersion(ctx, appConfigVersionID, updates)
}

// GetMaxPublishedVersion 获取最大发布版本号
func (r *AppRepo) GetMaxPublishedVersion(ctx context.Context, appID uuid.UUID) (int, error) {
	return r.dao.GetMaxPublishedVersion(ctx, appID)
}

// CreateAppConfig 创建应用配置
func (r *AppRepo) CreateAppConfig(ctx context.Context, appCfg *entity.AppConfig) (*entity.AppConfig, error) {
	return r.dao.CreateAppConfig(ctx, appCfg)
}

// GetConversation 获取会话
func (r *AppRepo) GetConversation(ctx context.Context, conversationID uuid.UUID) (*entity.Conversation, error) {
	return r.dao.GetConversation(ctx, conversationID)
}

// CreateConversation 创建会话
func (r *AppRepo) CreateConversation(ctx context.Context, accountID, appID uuid.UUID) (*entity.Conversation, error) {
	return r.dao.CreateConversation(ctx, accountID, appID)
}

// GetDebugConversationMessagesWithPage 获取调试会话消息分页列表
func (r *AppRepo) GetDebugConversationMessagesWithPage(ctx context.Context, conversationID uuid.UUID, page int, pageSize int, ctime int64) ([]*entity.Message, int64, error) {
	return r.dao.GetDebugConversationMessagesWithPage(ctx, conversationID, page, pageSize, ctime)
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

// GetRecentMessages 获取最近的消息记录
func (r *AppRepo) GetRecentMessages(ctx context.Context, conversationID uuid.UUID, limit int) ([]*entity.Message, error) {
	return r.dao.GetRecentMessages(ctx, conversationID, limit)
}

func (r *AppRepo) GetPublishHistoriesWithPage(ctx context.Context, appID uuid.UUID, pageReq req.GetPublishHistoriesWithPageReq) ([]*entity.AppConfigVersion, resp.Paginator, error) {
	return r.dao.GetPublishHistoriesWithPage(ctx, appID, pageReq)
}

func (r *AppRepo) GetApiTool(ctx context.Context, provider string, name string, accountID uuid.UUID) (*entity.ApiTool, error) {
	return r.dao.GetApiTool(ctx, provider, name, accountID)
}

func (r *AppRepo) GetWorkflows(ctx context.Context, workflowIDs []uuid.UUID, accountID uuid.UUID, status consts.WorkflowStatus) ([]*entity.Workflow, error) {
	return r.dao.GetWorkflows(ctx, workflowIDs, accountID, status)
}

func (r *AppRepo) GetDatasets(ctx context.Context, workflowIDs []uuid.UUID, accountID uuid.UUID) ([]*entity.Dataset, error) {
	return r.dao.GetDatasets(ctx, workflowIDs, accountID)
}
