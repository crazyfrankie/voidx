package dao

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type AppConfigDao struct {
	db *gorm.DB
}

func NewAppConfigDao(db *gorm.DB) *AppConfigDao {
	return &AppConfigDao{db: db}
}

// GetAppConfigByID 根据ID获取应用配置
func (d *AppConfigDao) GetAppConfigByID(ctx context.Context, id uuid.UUID) (*entity.AppConfig, error) {
	var appConfig entity.AppConfig
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&appConfig).Error
	if err != nil {
		return nil, err
	}
	return &appConfig, nil
}

// CreateAppConfig 创建应用配置
func (d *AppConfigDao) CreateAppConfig(ctx context.Context, appConfig *entity.AppConfig) error {
	return d.db.WithContext(ctx).Create(appConfig).Error
}

// UpdateAppConfig 更新应用配置
func (d *AppConfigDao) UpdateAppConfig(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return d.db.WithContext(ctx).Model(&entity.AppConfig{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteAppConfig 删除应用配置
func (d *AppConfigDao) DeleteAppConfig(ctx context.Context, id uuid.UUID) error {
	return d.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.AppConfig{}).Error
}

// CloneAppConfig 克隆应用配置
func (d *AppConfigDao) CloneAppConfig(ctx context.Context, sourceID uuid.UUID) (*entity.AppConfig, error) {
	// 1. 获取源配置
	var sourceConfig entity.AppConfig
	err := d.db.WithContext(ctx).Where("id = ?", sourceID).First(&sourceConfig).Error
	if err != nil {
		return nil, err
	}

	// 2. 创建新配置
	newConfig := entity.AppConfig{
		ModelConfig:          sourceConfig.ModelConfig,
		OpeningStatement:     sourceConfig.OpeningStatement,
		SuggestedAfterAnswer: sourceConfig.SuggestedAfterAnswer,
		RetrievalConfig:      sourceConfig.RetrievalConfig,
		LongTermMemory:       sourceConfig.LongTermMemory,
		TextToSpeech:         sourceConfig.TextToSpeech,
		SpeechToText:         sourceConfig.SpeechToText,
	}

	err = d.db.WithContext(ctx).Create(&newConfig).Error
	if err != nil {
		return nil, err
	}

	return &newConfig, nil
}

// GetDefaultAppConfig 获取默认应用配置
func (d *AppConfigDao) GetDefaultAppConfig(ctx context.Context) (*entity.AppConfig, error) {
	// 创建默认配置
	modelConfig := map[string]any{
		"provider":    "openai",
		"model":       "gpt-3.5-turbo",
		"temperature": 0.7,
		"max_tokens":  2000,
	}

	retrievalConfig := map[string]any{
		"retrieval_strategy": "semantic",
		"k":                  3,
		"score":              0.7,
	}

	longTermMemory := map[string]any{
		"enable": false,
	}

	textToSpeech := map[string]any{
		"enable": false,
		"voice":  "alloy",
	}

	speechToText := map[string]any{
		"enable": false,
	}

	appConfig := &entity.AppConfig{
		ModelConfig:          modelConfig,
		OpeningStatement:     "Hello! How can I help you today?",
		SuggestedAfterAnswer: nil,
		RetrievalConfig:      retrievalConfig,
		LongTermMemory:       longTermMemory,
		TextToSpeech:         textToSpeech,
		SpeechToText:         speechToText,
	}

	err := d.db.WithContext(ctx).Create(appConfig).Error
	if err != nil {
		return nil, err
	}

	return appConfig, nil
}

// GetWorkflowByID 根据ID获取工作流
func (d *AppConfigDao) GetWorkflowByID(ctx context.Context, id uuid.UUID) (*entity.Workflow, error) {
	var workflow entity.Workflow
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&workflow).Error
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

// GetAPIByID 根据ID获取API
func (d *AppConfigDao) GetAPIByID(ctx context.Context, id uuid.UUID) (*entity.ApiTool, error) {
	var api entity.ApiTool
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&api).Error
	if err != nil {
		return nil, err
	}
	return &api, nil
}

// GetAppConfigVersion 获取应用配置版本
func (d *AppConfigDao) GetAppConfigVersion(ctx context.Context, appConfigVersionID uuid.UUID) (*entity.AppConfigVersion, error) {
	var appConfigVersion entity.AppConfigVersion
	if err := d.db.WithContext(ctx).First(&appConfigVersion, "id = ?", appConfigVersionID).Error; err != nil {
		return nil, err
	}
	return &appConfigVersion, nil
}

// UpdateAppConfigVersion 更新应用配置版本
func (d *AppConfigDao) UpdateAppConfigVersion(ctx context.Context, appConfigVersionID uuid.UUID, updates map[string]any) error {
	return d.db.WithContext(ctx).Model(&entity.AppConfigVersion{}).Where("id = ?", appConfigVersionID).Updates(updates).Error
}

func (d *AppConfigDao) GetDatasetByID(ctx context.Context, id uuid.UUID) (*entity.Dataset, error) {
	var dataset entity.Dataset
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&dataset).Error
	if err != nil {
		return nil, err
	}
	return &dataset, nil
}

func (d *AppConfigDao) GetAPIToolByProviderAndName(ctx context.Context, providerID, toolName string) (*entity.ApiTool, error) {
	var apiTool entity.ApiTool
	err := d.db.WithContext(ctx).
		Where("provider_id = ? AND name = ?", providerID, toolName).
		First(&apiTool).Error
	if err != nil {
		return nil, err
	}
	return &apiTool, nil
}

func (d *AppConfigDao) GetAPIProviderByID(ctx context.Context, providerID uuid.UUID) (*entity.ApiToolProvider, error) {
	var apiProvider entity.ApiToolProvider
	err := d.db.WithContext(ctx).Model(&entity.ApiToolProvider{}).Where("id = ?", providerID).First(&apiProvider).Error
	if err != nil {
		return nil, err
	}
	return &apiProvider, nil
}

func (d *AppConfigDao) GetAppDatasetJoins(ctx context.Context, appID uuid.UUID) ([]*entity.AppDatasetJoin, error) {
	var joins []*entity.AppDatasetJoin
	err := d.db.WithContext(ctx).Where("app_id = ?", appID).Find(&joins).Error
	if err != nil {
		return nil, err
	}
	return joins, nil
}

func (d *AppConfigDao) DeleteAppDatasetJoin(ctx context.Context, appID, datasetID uuid.UUID) error {
	return d.db.WithContext(ctx).
		Where("app_id = ? AND dataset_id = ?", appID, datasetID).
		Delete(&entity.AppDatasetJoin{}).Error
}
