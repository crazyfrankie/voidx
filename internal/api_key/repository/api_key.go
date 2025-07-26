package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/api_key/repository/dao"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
)

type ApiKeyRepo struct {
	dao *dao.ApiKeyDao
}

func NewApiKeyRepo(d *dao.ApiKeyDao) *ApiKeyRepo {
	return &ApiKeyRepo{dao: d}
}

// CreateApiKey 创建API秘钥
func (r *ApiKeyRepo) CreateApiKey(ctx context.Context, apiKey *entity.ApiKey) error {
	return r.dao.CreateApiKey(ctx, apiKey)
}

// GetApiKeyByID 根据ID获取API秘钥
func (r *ApiKeyRepo) GetApiKeyByID(ctx context.Context, id uuid.UUID) (*entity.ApiKey, error) {
	return r.dao.GetApiKeyByID(ctx, id)
}

// GetApiKeyByCredential 根据凭证获取API秘钥
func (r *ApiKeyRepo) GetApiKeyByCredential(ctx context.Context, credential string) (*entity.ApiKey, error) {
	return r.dao.GetApiKeyByCredential(ctx, credential)
}

// GetApiKeysByAccountID 根据账户ID获取API秘钥列表
func (r *ApiKeyRepo) GetApiKeysByAccountID(
	ctx context.Context,
	accountID uuid.UUID,
	pageReq req.GetApiKeysWithPageReq,
) ([]entity.ApiKey, int64, error) {
	return r.dao.GetApiKeysByAccountID(ctx, accountID, pageReq)
}

// UpdateApiKey 更新API秘钥
func (r *ApiKeyRepo) UpdateApiKey(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return r.dao.UpdateApiKey(ctx, id, updates)
}

// DeleteApiKey 删除API秘钥
func (r *ApiKeyRepo) DeleteApiKey(ctx context.Context, id uuid.UUID) error {
	return r.dao.DeleteApiKey(ctx, id)
}
