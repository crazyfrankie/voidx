package dao

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
)

type ApiToolDao struct {
	db *gorm.DB
}

func NewApiToolDao(db *gorm.DB) *ApiToolDao {
	return &ApiToolDao{db: db}
}

func (d *ApiToolDao) CreateApiToolProvider(ctx context.Context, provider *entity.ApiToolProvider) error {
	return d.db.WithContext(ctx).Create(provider).Error
}

func (d *ApiToolDao) GetApiToolProviderByID(ctx context.Context, id uuid.UUID) (*entity.ApiToolProvider, error) {
	var provider entity.ApiToolProvider
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&provider).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (d *ApiToolDao) GetApiToolProviderByName(ctx context.Context, userID uuid.UUID, name string) (*entity.ApiToolProvider, error) {
	var provider entity.ApiToolProvider
	err := d.db.WithContext(ctx).Where("account_id = ? AND name = ?", userID, name).First(&provider).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (d *ApiToolDao) GetApiToolProvidersByAccountID(ctx context.Context, accountID uuid.UUID, pageReq req.GetApiToolProvidersWithPageReq) ([]entity.ApiToolProvider, error) {
	var providers []entity.ApiToolProvider

	query := d.db.WithContext(ctx).Where("account_id = ?", accountID)

	// 添加搜索条件
	if pageReq.SearchWord != "" {
		query = query.Where("name ILIKE ?", "%"+pageReq.SearchWord+"%")
	}

	// 分页查询
	offset := (pageReq.Page - 1) * pageReq.PageSize
	err := query.Order("ctime DESC").
		Offset(offset).
		Limit(pageReq.PageSize).
		Find(&providers).Error

	if err != nil {
		return nil, err
	}

	return providers, nil
}

func (d *ApiToolDao) UpdateApiToolProvider(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return d.db.WithContext(ctx).Model(&entity.ApiToolProvider{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (d *ApiToolDao) DeleteApiToolProvider(ctx context.Context, id uuid.UUID) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除相关的API工具
		if err := tx.Where("provider_id = ?", id).Delete(&entity.ApiTool{}).Error; err != nil {
			return err
		}

		// 删除提供商
		return tx.Where("id = ?", id).Delete(&entity.ApiToolProvider{}).Error
	})
}

func (d *ApiToolDao) CheckProviderNameExistsExclude(ctx context.Context, accountID uuid.UUID, name string, excludeID uuid.UUID) (bool, error) {
	var count int64
	err := d.db.WithContext(ctx).Model(&entity.ApiToolProvider{}).
		Where("account_id = ? AND name = ? AND id != ?", accountID, name, excludeID).
		Count(&count).Error
	return count > 0, err
}

func (d *ApiToolDao) CreateApiTool(ctx context.Context, tool *entity.ApiTool) error {
	return d.db.WithContext(ctx).Create(tool).Error
}

func (d *ApiToolDao) GetApiToolByID(ctx context.Context, id uuid.UUID) (*entity.ApiTool, error) {
	var tool entity.ApiTool
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&tool).Error
	if err != nil {
		return nil, err
	}
	return &tool, nil
}
