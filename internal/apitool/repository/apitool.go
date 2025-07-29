package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/apitool/repository/dao"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
)

type ApiToolRepo struct {
	dao *dao.ApiToolDao
}

func NewApiToolRepo(d *dao.ApiToolDao) *ApiToolRepo {
	return &ApiToolRepo{dao: d}
}

func (r *ApiToolRepo) CreateApiToolProvider(ctx context.Context, provider *entity.ApiToolProvider) error {
	return r.dao.CreateApiToolProvider(ctx, provider)
}

func (r *ApiToolRepo) GetApiToolProviderByID(ctx context.Context, id uuid.UUID) (*entity.ApiToolProvider, error) {
	return r.dao.GetApiToolProviderByID(ctx, id)
}

func (r *ApiToolRepo) GetApiToolProviderByName(ctx context.Context, userID uuid.UUID, name string) (*entity.ApiToolProvider, error) {
	return r.dao.GetApiToolProviderByName(ctx, userID, name)
}

func (r *ApiToolRepo) GetApiToolProvidersByAccountID(ctx context.Context, accountID uuid.UUID, pageReq req.GetApiToolProvidersWithPageReq) ([]entity.ApiToolProvider, int64, error) {
	return r.dao.GetApiToolProvidersByAccountID(ctx, accountID, pageReq)
}

func (r *ApiToolRepo) UpdateApiToolProvider(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return r.dao.UpdateApiToolProvider(ctx, id, updates)
}

func (r *ApiToolRepo) DeleteApiToolProvider(ctx context.Context, id uuid.UUID) error {
	return r.dao.DeleteApiToolProvider(ctx, id)
}

func (r *ApiToolRepo) CheckProviderNameExistsExclude(ctx context.Context, accountID uuid.UUID, name string, excludeID uuid.UUID) (bool, error) {
	return r.dao.CheckProviderNameExistsExclude(ctx, accountID, name, excludeID)
}

// API工具相关方法
func (r *ApiToolRepo) CreateApiTool(ctx context.Context, tool *entity.ApiTool) error {
	return r.dao.CreateApiTool(ctx, tool)
}

func (r *ApiToolRepo) GetApiToolByProviderID(ctx context.Context, providerID uuid.UUID, toolName string) (*entity.ApiTool, error) {
	return r.dao.GetApiToolByProviderID(ctx, providerID, toolName)
}

func (r *ApiToolRepo) GetApiTools(ctx context.Context, providerID uuid.UUID) ([]entity.ApiTool, error) {
	return r.dao.GetApiTools(ctx, providerID)
}
