package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/app_config/repository/dao"
	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type AppConfigRepo struct {
	dao *dao.AppConfigDao
}

func NewAppConfigRepo(d *dao.AppConfigDao) *AppConfigRepo {
	return &AppConfigRepo{dao: d}
}

// GetAppConfigByID 根据ID获取应用配置
func (r *AppConfigRepo) GetAppConfigByID(ctx context.Context, id uuid.UUID) (*entity.AppConfig, error) {
	return r.dao.GetAppConfigByID(ctx, id)
}

// CreateAppConfig 创建应用配置
func (r *AppConfigRepo) CreateAppConfig(ctx context.Context, appConfig *entity.AppConfig) error {
	return r.dao.CreateAppConfig(ctx, appConfig)
}

// UpdateAppConfig 更新应用配置
func (r *AppConfigRepo) UpdateAppConfig(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return r.dao.UpdateAppConfig(ctx, id, updates)
}

// DeleteAppConfig 删除应用配置
func (r *AppConfigRepo) DeleteAppConfig(ctx context.Context, id uuid.UUID) error {
	return r.dao.DeleteAppConfig(ctx, id)
}

// CloneAppConfig 克隆应用配置
func (r *AppConfigRepo) CloneAppConfig(ctx context.Context, sourceID uuid.UUID) (*entity.AppConfig, error) {
	return r.dao.CloneAppConfig(ctx, sourceID)
}

// GetDefaultAppConfig 获取默认应用配置
func (r *AppConfigRepo) GetDefaultAppConfig(ctx context.Context) (*entity.AppConfig, error) {
	return r.dao.GetDefaultAppConfig(ctx)
}

// GetWorkflowByID 根据ID获取工作流
func (r *AppConfigRepo) GetWorkflowByID(ctx context.Context, id uuid.UUID) (*entity.Workflow, error) {
	return r.dao.GetWorkflowByID(ctx, id)
}

// GetAPIByID 根据ID获取API
func (r *AppConfigRepo) GetAPIByID(ctx context.Context, id uuid.UUID) (*entity.ApiTool, error) {
	return r.dao.GetAPIByID(ctx, id)
}

// GetAppConfigVersion 获取应用配置版本
func (r *AppConfigRepo) GetAppConfigVersion(ctx context.Context, appConfigVersionID uuid.UUID) (*entity.AppConfigVersion, error) {
	return r.dao.GetAppConfigVersion(ctx, appConfigVersionID)
}

func (r *AppConfigRepo) UpdateAppConfigVersion(ctx context.Context, appConfigVersionID uuid.UUID, updates map[string]any) error {
	return r.dao.UpdateAppConfigVersion(ctx, appConfigVersionID, updates)
}

func (r *AppConfigRepo) GetDatasetByID(ctx context.Context, id uuid.UUID) (*entity.Dataset, error) {
	return r.dao.GetDatasetByID(ctx, id)
}
func (r *AppConfigRepo) GetAPIToolByProviderAndName(ctx context.Context, providerID, toolName string) (*entity.ApiTool, error) {
	return r.dao.GetAPIToolByProviderAndName(ctx, providerID, toolName)
}

func (r *AppConfigRepo) GetAPIProviderByID(ctx context.Context, providerID uuid.UUID) (*entity.ApiToolProvider, error) {
	return r.dao.GetAPIProviderByID(ctx, providerID)
}

func (r *AppConfigRepo) GetAppDatasetJoins(ctx context.Context, appID uuid.UUID) ([]*entity.AppDatasetJoin, error) {
	return r.dao.GetAppDatasetJoins(ctx, appID)
}

func (r *AppConfigRepo) DeleteAppDatasetJoin(ctx context.Context, appID, datasetID uuid.UUID) error {
	return r.dao.DeleteAppDatasetJoin(ctx, appID, datasetID)
}
