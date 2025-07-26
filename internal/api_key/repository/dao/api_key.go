package dao

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
)

type ApiKeyDao struct {
	db *gorm.DB
}

func NewApiKeyDao(db *gorm.DB) *ApiKeyDao {
	return &ApiKeyDao{db: db}
}

// CreateApiKey 创建API秘钥
func (d *ApiKeyDao) CreateApiKey(ctx context.Context, apiKey *entity.ApiKey) error {
	return d.db.WithContext(ctx).Create(apiKey).Error
}

// GetApiKeyByID 根据ID获取API秘钥
func (d *ApiKeyDao) GetApiKeyByID(ctx context.Context, id uuid.UUID) (*entity.ApiKey, error) {
	var apiKey entity.ApiKey
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// GetApiKeyByCredential 根据凭证获取API秘钥
func (d *ApiKeyDao) GetApiKeyByCredential(ctx context.Context, credential string) (*entity.ApiKey, error) {
	var apiKey entity.ApiKey
	err := d.db.WithContext(ctx).Where("api_key = ?", credential).First(&apiKey).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &apiKey, nil
}

// GetApiKeysByAccountID 根据账户ID获取API秘钥列表
func (d *ApiKeyDao) GetApiKeysByAccountID(ctx context.Context, accountID uuid.UUID, pageReq req.GetApiKeysWithPageReq) ([]entity.ApiKey, int64, error) {
	query := d.db.WithContext(ctx).Where("account_id = ?", accountID)

	// 获取总数
	var total int64
	if err := query.Model(&entity.ApiKey{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	var apiKeys []entity.ApiKey
	offset := (pageReq.Page - 1) * pageReq.PageSize
	err := query.Order("ctime DESC").Offset(offset).Limit(pageReq.PageSize).Find(&apiKeys).Error
	if err != nil {
		return nil, 0, err
	}

	return apiKeys, total, nil
}

// UpdateApiKey 更新API秘钥
func (d *ApiKeyDao) UpdateApiKey(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return d.db.WithContext(ctx).Model(&entity.ApiKey{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteApiKey 删除API秘钥
func (d *ApiKeyDao) DeleteApiKey(ctx context.Context, id uuid.UUID) error {
	return d.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.ApiKey{}).Error
}
