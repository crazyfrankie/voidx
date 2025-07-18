package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/util"
)

// GetPublishedConfig 获取已发布的配置
func (s *AppService) GetPublishedConfig(ctx context.Context, appID uuid.UUID) (map[string]interface{}, error) {
	// 获取当前用户ID
	accountID, err := util.GetCurrentUserID(ctx)
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

	// 构建发布配置响应
	publishedConfig := map[string]interface{}{
		"web_app": map[string]interface{}{
			"token":  app.Token,
			"status": app.Status,
		},
	}

	return publishedConfig, nil
}

// RegenerateWebAppToken 重新生成WebApp令牌
func (s *AppService) RegenerateWebAppToken(ctx context.Context, appID uuid.UUID) (string, error) {
	// 获取当前用户ID
	accountID, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return "", err
	}

	// 获取应用
	app, err := s.repo.GetAppByID(ctx, appID)
	if err != nil {
		return "", err
	}

	// 检查权限
	if app.AccountID != accountID {
		return "", errno.ErrForbidden.AppendBizMessage("无权访问该应用")
	}

	// 检查应用是否已发布
	if app.Status != "published" {
		return "", errno.ErrValidate.AppendBizMessage("应用未发布，无法生成WebApp令牌")
	}

	// 生成新的令牌
	token := generateRandomString(16)
	app.Token = token

	// 更新应用
	if err := s.repo.UpdateApp(ctx, app); err != nil {
		return "", err
	}

	return token, nil
}
