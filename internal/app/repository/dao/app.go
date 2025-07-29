package dao

import (
	"context"
	"github.com/crazyfrankie/voidx/pkg/consts"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
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

func (d *AppDao) UpdatesApp(ctx context.Context, appID uuid.UUID, updates map[string]any) error {
	return d.db.WithContext(ctx).Where("id = ?", appID).Updates(updates).Error
}

// DeleteApp 删除应用
func (d *AppDao) DeleteApp(ctx context.Context, appID uuid.UUID) error {
	return d.db.WithContext(ctx).Delete(&entity.App{}, "id = ?", appID).Error
}

// GetAppsWithPage 获取应用分页列表
func (d *AppDao) GetAppsWithPage(ctx context.Context, accountID uuid.UUID, page int, pageSize int, searchWord string) ([]*entity.App, int64, int64, error) {
	var apps []*entity.App
	var totalRecords int64

	query := d.db.WithContext(ctx).Model(&entity.App{}).Where("account_id = ?", accountID)

	if searchWord != "" {
		query = query.Where("name LIKE ?", "%"+searchWord+"%")
	}

	// 获取总记录数
	if err := query.Count(&totalRecords).Error; err != nil {
		return nil, 0, 0, err
	}

	// 计算总页数
	totalPages := (totalRecords + int64(pageSize) - 1) / int64(pageSize)

	// 获取分页数据
	offset := (page - 1) * pageSize
	if err := query.Order("ctime DESC").Offset(offset).Limit(pageSize).Find(&apps).Error; err != nil {
		return nil, 0, 0, err
	}

	return apps, totalRecords, totalPages, nil
}

// CreateAppConfigVersion 创建应用配置版本
//
//	func (d *AppDao) CreateAppConfigVersion(ctx context.Context, appID uuid.UUID, version int, configType string, config map[string]any) (*entity.AppConfigVersion, error) {
//		var err error
//
//		appConfigVersion := &entity.AppConfigVersion{
//			AppID:      appID,
//			Version:    version,
//			ConfigType: configType,
//		}
//		appConfigVersion.ModelConfig, err = sonic.Marshal(config)
//		if err != nil {
//			return nil, err
//		}
//
//		if err := d.db.WithContext(ctx).Create(appConfigVersion).Error; err != nil {
//			return nil, err
//		}
//
//		return appConfigVersion, nil
//	}
func (d *AppDao) CreateAppConfigVersion(ctx context.Context, appCfgVer *entity.AppConfigVersion) (*entity.AppConfigVersion, error) {
	if err := d.db.WithContext(ctx).Create(appCfgVer).Error; err != nil {
		return nil, err
	}
	return appCfgVer, nil
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
func (d *AppDao) CreateAppConfig(ctx context.Context, appConfig *entity.AppConfig) (*entity.AppConfig, error) {
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
		CreatedBy: accountID,
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

// GetRecentMessages 获取最近的消息记录
func (d *AppDao) GetRecentMessages(ctx context.Context, conversationID uuid.UUID, limit int) ([]*entity.Message, error) {
	var messages []*entity.Message
	err := d.db.WithContext(ctx).
		Where("conversation_id = ? AND status IN (?) AND answer != ''", conversationID, []string{"stop", "normal"}).
		Order("ctime DESC").
		Limit(limit).
		Find(&messages).Error

	// 反转数组以获得正确的时间顺序
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, err
}

// GetPublishHistoriesWithPage 获取发布历史分页列表
func (d *AppDao) GetPublishHistoriesWithPage(ctx context.Context, appID uuid.UUID, pageReq req.GetPublishHistoriesWithPageReq) ([]*entity.AppConfigVersion, resp.Paginator, error) {
	var histories []*entity.AppConfigVersion
	var total int64

	// 计算总数
	err := d.db.WithContext(ctx).Model(&entity.AppConfigVersion{}).
		Where("app_id = ? AND config_type = ?", appID, "published").
		Count(&total).Error
	if err != nil {
		return nil, resp.Paginator{}, err
	}

	// 分页查询
	offset := (pageReq.CurrentPage - 1) * pageReq.PageSize
	err = d.db.WithContext(ctx).
		Where("app_id = ? AND config_type = ?", appID, "published").
		Order("version DESC").
		Offset(offset).
		Limit(pageReq.PageSize).
		Find(&histories).Error
	if err != nil {
		return nil, resp.Paginator{}, err
	}

	paginator := resp.Paginator{
		CurrentPage: pageReq.CurrentPage,
		PageSize:    pageReq.PageSize,
		TotalRecord: int(total),
		TotalPage:   (int(total) + pageReq.PageSize - 1) / pageReq.PageSize,
	}

	return histories, paginator, nil
}

// GetApiTool 获取 api tool
func (d *AppDao) GetApiTool(ctx context.Context, provider string, name string, accountID uuid.UUID) (*entity.ApiTool, error) {
	var apiTool *entity.ApiTool
	err := d.db.WithContext(ctx).Model(&entity.ApiTool{}).
		Where("provider_id = ? AND name = ? AND account_id = ?", provider, name, accountID).
		First(&apiTool).Error
	if err != nil {
		return nil, err
	}

	return apiTool, nil
}

// GetWorkflows 获取已发布的 workflow
func (d *AppDao) GetWorkflows(ctx context.Context, workflowIDs []uuid.UUID, accountID uuid.UUID, status consts.WorkflowStatus) ([]*entity.Workflow, error) {
	var workflows []*entity.Workflow
	if err := d.db.WithContext(ctx).Model(&entity.Workflow{}).Where("id IN (?) AND account_id = ? AND status = ?",
		workflowIDs, accountID, status).
		Find(&workflows).Error; err != nil {
		return nil, err
	}

	return workflows, nil
}

// GetDatasets 获取知识库
func (d *AppDao) GetDatasets(ctx context.Context, datasetIDs []uuid.UUID, accountID uuid.UUID) ([]*entity.Dataset, error) {
	var datasets []*entity.Dataset
	if err := d.db.WithContext(ctx).Model(&entity.Dataset{}).Where("id IN (?) AND account_id = ?",
		datasetIDs, accountID).
		Find(&datasets).Error; err != nil {
		return nil, err
	}

	return datasets, nil
}
