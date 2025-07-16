package service

import (
	"context"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms/openai"

	"github.com/crazyfrankie/voidx/internal/app/repository"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/errno"
)

// AppService 应用服务
type AppService struct {
	repo *repository.AppRepo
	llm  *openai.LLM
}

// NewAppService 创建应用服务
func NewAppService(repo *repository.AppRepo, llm *openai.LLM) *AppService {
	return &AppService{
		repo: repo,
		llm:  llm,
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

	// 创建草稿配置
	draftConfig := map[string]interface{}{
		"model_config": map[string]interface{}{
			"provider": "openai",
			"model":    "gpt-3.5-turbo",
			"parameters": map[string]interface{}{
				"temperature":       0.7,
				"top_p":             1.0,
				"frequency_penalty": 0.0,
				"presence_penalty":  0.0,
				"max_tokens":        1000,
			},
		},
		"dialog_round":  10,
		"preset_prompt": "",
		"tools":         []interface{}{},
		"workflows":     []interface{}{},
		"datasets":      []interface{}{},
		"retrieval_config": map[string]interface{}{
			"retrieval_strategy": "semantic",
			"k":                  3,
			"score":              0.7,
		},
		"long_term_memory": map[string]interface{}{
			"enable": false,
		},
		"opening_statement":      "",
		"opening_questions":      []interface{}{},
		"speech_to_text":         map[string]interface{}{"enable": false},
		"text_to_speech":         map[string]interface{}{"enable": false, "voice": "alloy", "auto_play": false},
		"suggested_after_answer": map[string]interface{}{"enable": false},
		"review_config": map[string]interface{}{
			"enable":         false,
			"keywords":       []interface{}{},
			"inputs_config":  map[string]interface{}{"enable": false, "preset_response": ""},
			"outputs_config": map[string]interface{}{"enable": false},
		},
	}

	// 创建草稿配置版本
	draftAppConfigVersion, err := s.repo.CreateAppConfigVersion(ctx, app.ID, 0, "draft", draftConfig)
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

// 辅助函数：从上下文中获取当前用户ID
func getCurrentUserID(ctx context.Context) (uuid.UUID, error) {
	// 在实际项目中，这里应该从上下文中获取当前用户ID
	// 这里为了简化，我们假设上下文中有一个键为"user_id"的值
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

// 辅助函数：生成随机字符串
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
