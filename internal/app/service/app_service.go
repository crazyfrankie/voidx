package service

import (
	"context"
	"math/rand"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/app/repository"
	"github.com/crazyfrankie/voidx/internal/core/llm"
	llmService "github.com/crazyfrankie/voidx/internal/llm/service"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/consts"
	"github.com/crazyfrankie/voidx/pkg/errno"
)

type AppService struct {
	repo             *repository.AppRepo
	languageModelMgr *llm.LanguageModelManager
	llmService       *llmService.LLMService
}

func NewAppService(repo *repository.AppRepo, languageModelMgr *llm.LanguageModelManager, llmSvc *llmService.LLMService) *AppService {
	return &AppService{
		repo:             repo,
		languageModelMgr: languageModelMgr,
		llmService:       llmSvc,
	}
}

// CreateApp 创建应用
func (s *AppService) CreateApp(ctx context.Context, createReq req.CreateAppReq) (uuid.UUID, error) {
	// 获取当前用户ID
	accountID, err := getCurrentUserID(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	// 创建应用
	app, err := s.repo.CreateApp(ctx, &entity.App{
		AccountID:   accountID,
		Name:        createReq.Name,
		Icon:        createReq.Icon,
		Description: createReq.Description,
		Status:      "draft", // 默认为草稿状态
	})
	if err != nil {
		return uuid.Nil, err
	}

	// 创建草稿配置版本
	draftAppConfigVersion, err := s.repo.CreateAppConfigVersion(ctx, app.ID, 0, "draft", consts.DefaultAppConfig)
	if err != nil {
		return uuid.Nil, err
	}

	// 更新应用的草稿配置ID
	app.DraftAppConfigID = &draftAppConfigVersion.ID
	if err := s.repo.UpdateApp(ctx, app); err != nil {
		return uuid.Nil, err
	}

	return app.ID, nil
}

// GetApp 获取应用
func (s *AppService) GetApp(ctx context.Context, appID uuid.UUID) (*resp.AppResp, error) {
	// 获取当前用户ID
	accountID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	// 获取应用
	app, err := s.repo.GetAppByID(ctx, appID)
	if err != nil {
		return nil, err
	}

	// 检查权限
	if app.AccountID != accountID {
		return nil, errno.ErrForbidden.AppendBizMessage("无权访问该应用")
	}

	// 转换为响应
	return &resp.AppResp{
		ID:                  app.ID,
		Name:                app.Name,
		Icon:                app.Icon,
		Description:         app.Description,
		Status:              app.Status,
		AppConfigID:         app.AppConfigID,
		DraftAppConfigID:    app.DraftAppConfigID,
		DebugConversationID: app.DebugConversationID,
		Token:               app.Token,
		Ctime:               app.Ctime,
		Utime:               app.Utime,
	}, nil
}

// UpdateApp 更新应用
func (s *AppService) UpdateApp(ctx context.Context, appID uuid.UUID, updateReq req.UpdateAppReq) error {
	// 获取当前用户ID
	accountID, err := getCurrentUserID(ctx)
	if err != nil {
		return err
	}

	// 获取应用
	app, err := s.repo.GetAppByID(ctx, appID)
	if err != nil {
		return err
	}

	// 检查权限
	if app.AccountID != accountID {
		return errno.ErrForbidden.AppendBizMessage("无权访问该应用")
	}

	// 更新应用
	if updateReq.Name != "" {
		app.Name = updateReq.Name
	}
	if updateReq.Icon != "" {
		app.Icon = updateReq.Icon
	}
	if updateReq.Description != "" {
		app.Description = updateReq.Description
	}

	if err := s.repo.UpdateApp(ctx, app); err != nil {
		return err
	}

	return nil
}

// DeleteApp 删除应用
func (s *AppService) DeleteApp(ctx context.Context, appID uuid.UUID) error {
	// 获取当前用户ID
	accountID, err := getCurrentUserID(ctx)
	if err != nil {
		return err
	}

	// 获取应用
	app, err := s.repo.GetAppByID(ctx, appID)
	if err != nil {
		return err
	}

	// 检查权限
	if app.AccountID != accountID {
		return errno.ErrForbidden.AppendBizMessage("无权访问该应用")
	}

	// 删除应用
	if err := s.repo.DeleteApp(ctx, appID); err != nil {
		return err
	}

	return nil
}

// CopyApp 拷贝应用
func (s *AppService) CopyApp(ctx context.Context, appID uuid.UUID) (uuid.UUID, error) {
	// 获取当前用户ID
	accountID, err := getCurrentUserID(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	// 获取应用
	app, err := s.repo.GetAppByID(ctx, appID)
	if err != nil {
		return uuid.Nil, err
	}

	// 检查权限
	if app.AccountID != accountID {
		return uuid.Nil, errno.ErrForbidden.AppendBizMessage("无权访问该应用")
	}

	// 获取草稿配置
	draftAppConfigVersion, err := s.repo.GetAppConfigVersion(ctx, *app.DraftAppConfigID)
	if err != nil {
		return uuid.Nil, err
	}

	// 创建新应用
	newApp := &entity.App{
		AccountID:   accountID,
		Name:        app.Name + " (复制)",
		Icon:        app.Icon,
		Description: app.Description,
		Status:      "draft", // 默认为草稿状态
	}

	if err := s.repo.UpdateApp(ctx, newApp); err != nil {
		return uuid.Nil, err
	}

	// 创建新的草稿配置
	newDraftAppConfigVersion, err := s.repo.CreateAppConfigVersion(ctx, newApp.ID, 0, "draft", draftAppConfigVersion.Config)
	if err != nil {
		return uuid.Nil, err
	}

	// 更新新应用的草稿配置ID
	newApp.DraftAppConfigID = &newDraftAppConfigVersion.ID
	if err := s.repo.UpdateApp(ctx, newApp); err != nil {
		return uuid.Nil, err
	}

	return newApp.ID, nil
}

// GetAppsWithPage 获取应用分页列表
func (s *AppService) GetAppsWithPage(ctx context.Context, pageReq req.GetAppsWithPageReq) ([]*resp.AppResp, *resp.Paginator, error) {
	// 获取当前用户ID
	accountID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, nil, err
	}

	// 获取应用分页列表
	apps, total, err := s.repo.GetAppsWithPage(ctx, accountID, pageReq.Page, pageReq.PageSize, pageReq.SearchWord)
	if err != nil {
		return nil, nil, err
	}

	// 转换为响应
	appResps := make([]*resp.AppResp, 0, len(apps))
	for _, app := range apps {
		appResps = append(appResps, &resp.AppResp{
			ID:                  app.ID,
			Name:                app.Name,
			Icon:                app.Icon,
			Description:         app.Description,
			Status:              app.Status,
			AppConfigID:         app.AppConfigID,
			DraftAppConfigID:    app.DraftAppConfigID,
			DebugConversationID: app.DebugConversationID,
			Token:               app.Token,
			Ctime:               app.Ctime,
			Utime:               app.Utime,
		})
	}

	// 构建分页器
	paginator := &resp.Paginator{
		CurrentPage: pageReq.Page,
		PageSize:    pageReq.PageSize,
		TotalPage:   int(total) / pageReq.PageSize,
		TotalRecord: int(total),
	}
	if int(total)%pageReq.PageSize > 0 {
		paginator.TotalPage++
	}

	return appResps, paginator, nil
}

// updateAppDatasetJoins 更新应用知识库关联
func (s *AppService) updateAppDatasetJoins(ctx context.Context, appID uuid.UUID, draftConfig map[string]interface{}) error {
	// 先删除现有关联
	if err := s.repo.DeleteAppDatasetJoins(ctx, appID); err != nil {
		return err
	}

	// 添加新关联
	if datasets, exists := draftConfig["datasets"]; exists {
		if datasetList, ok := datasets.([]interface{}); ok {
			for _, dataset := range datasetList {
				if datasetMap, ok := dataset.(map[string]interface{}); ok {
					if idStr, exists := datasetMap["id"]; exists {
						if datasetID, err := uuid.Parse(idStr.(string)); err == nil {
							if err := s.repo.CreateAppDatasetJoin(ctx, appID, datasetID); err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// createPublishHistory 创建发布历史记录
func (s *AppService) createPublishHistory(ctx context.Context, appID uuid.UUID, draftConfig map[string]interface{}) error {
	// 获取当前最大发布版本
	maxVersion, err := s.repo.GetMaxPublishedVersion(ctx, appID)
	if err != nil {
		return err
	}

	// 创建发布历史记录
	_, err = s.repo.CreateAppConfigVersion(ctx, appID, maxVersion+1, "published", draftConfig)
	return err
}

func getCurrentUserID(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return uuid.Nil, errno.ErrUnauthorized.AppendBizMessage("未登录")
	}

	// 解析用户ID
	id, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, errno.ErrUnauthorized.AppendBizMessage("用户ID格式不正确")
	}

	return id, nil
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
