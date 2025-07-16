package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/pkg/errno"
)

// GetDraftAppConfig 获取应用的最新草稿配置
func (s *AppService) GetDraftAppConfig(ctx context.Context, appID uuid.UUID) (map[string]interface{}, error) {
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

	// 检查草稿配置ID是否存在
	if app.DraftAppConfigID == nil {
		return nil, errno.ErrNotFound.AppendBizMessage("应用草稿配置不存在")
	}

	// 获取草稿配置
	draftAppConfigVersion, err := s.repo.GetAppConfigVersion(ctx, *app.DraftAppConfigID)
	if err != nil {
		return nil, err
	}

	return draftAppConfigVersion.Config, nil
}

// UpdateDraftAppConfig 更新应用的最新草稿配置
func (s *AppService) UpdateDraftAppConfig(ctx context.Context, appID uuid.UUID, draftConfig map[string]interface{}) error {
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

	// 检查草稿配置ID是否存在
	if app.DraftAppConfigID == nil {
		return errno.ErrNotFound.AppendBizMessage("应用草稿配置不存在")
	}

	// 获取草稿配置
	draftAppConfigVersion, err := s.repo.GetAppConfigVersion(ctx, *app.DraftAppConfigID)
	if err != nil {
		return err
	}

	// 校验草稿配置
	if err := validateDraftAppConfig(draftConfig, accountID); err != nil {
		return err
	}

	// 更新草稿配置
	draftAppConfigVersion.Config = draftConfig
	if err := s.repo.UpdateAppConfigVersion(ctx, draftAppConfigVersion); err != nil {
		return err
	}

	return nil
}

// validateDraftAppConfig 校验草稿配置
func validateDraftAppConfig(draftConfig map[string]interface{}, accountID uuid.UUID) error {
	// 这里应该实现完整的配置校验逻辑
	// 为了简化，我们只做基本的校验

	// 检查必要字段是否存在
	requiredFields := []string{
		"model_config", "dialog_round", "preset_prompt",
		"tools", "workflows", "datasets", "retrieval_config",
		"long_term_memory", "opening_statement", "opening_questions",
		"speech_to_text", "text_to_speech", "suggested_after_answer", "review_config",
	}

	for _, field := range requiredFields {
		if _, ok := draftConfig[field]; !ok {
			return errno.ErrValidate.AppendBizMessage("草稿配置缺少必要字段: " + field)
		}
	}

	return nil
}
