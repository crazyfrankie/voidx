package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/builtin_app/repository"
	"github.com/crazyfrankie/voidx/internal/core/builtin_apps"
	"github.com/crazyfrankie/voidx/internal/core/builtin_apps/entities"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/pkg/consts"
	"github.com/crazyfrankie/voidx/pkg/errno"
)

type BuiltinService struct {
	repo              *repository.BuiltinRepository
	builtinAppManager *builtin_apps.BuiltinAppManager
}

func NewBuiltinService(repo *repository.BuiltinRepository, builtinAppManager *builtin_apps.BuiltinAppManager) *BuiltinService {
	return &BuiltinService{repo: repo, builtinAppManager: builtinAppManager}
}

func (s *BuiltinService) GetBuiltinAppCategories(ctx context.Context) []*entities.CategoryEntity {
	return s.builtinAppManager.GetCategories()
}

func (s *BuiltinService) GetBuiltinApps(ctx context.Context) []*entities.BuiltinAppEntity {
	return s.builtinAppManager.GetBuiltinApps()
}

func (s *BuiltinService) AddBuiltinAppToSpace(ctx context.Context, userID uuid.UUID, addReq req.AddBuiltinAppReq) (*entity.App, error) {
	builtinApp := s.builtinAppManager.GetBuiltinApp(addReq.BuiltinAppID.String())
	if builtinApp == nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("该内置应用不存在，请核实后重试"))
	}

	app, err := s.repo.AddBuiltinApp(ctx, builtinApp, &entity.App{
		AccountID:   userID,
		Status:      consts.AppStatusDraft,
		Name:        builtinApp.Name,
		Icon:        builtinApp.Icon,
		Description: builtinApp.Description,
	})
	if err != nil {
		return nil, err
	}

	return app, nil
}
