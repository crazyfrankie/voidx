package repository

import (
	"context"
	
	"github.com/crazyfrankie/voidx/internal/builtin_app/repository/dao"
	"github.com/crazyfrankie/voidx/internal/core/builtin_apps/entities"
	"github.com/crazyfrankie/voidx/internal/models/entity"
)

type BuiltinRepository struct {
	dao *dao.BuiltinDao
}

func NewBuiltinRepository(dao *dao.BuiltinDao) *BuiltinRepository {
	return &BuiltinRepository{dao: dao}
}

func (r *BuiltinRepository) AddBuiltinApp(ctx context.Context, builtinApp *entities.BuiltinAppEntity, app *entity.App) (*entity.App, error) {
	return r.dao.AddBuiltinApp(ctx, builtinApp, app)
}
