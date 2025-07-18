package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/util"
)

// GetDraftAppConfig 获取应用的最新草稿配置
func (s *AppService) GetDraftAppConfig(ctx context.Context, appID uuid.UUID) (map[string]interface{}, error) {
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

	// 检查草稿配置ID是否存在
	if app.DraftAppConfigID == nil {
		return nil, errno.ErrNotFound.AppendBizMessage("应用草稿配置不存在")
	}

	// 获取草稿配置
	draftAppConfigVersion, err := s.repo.GetAppConfigVersion(ctx, *app.DraftAppConfigID)
	if err != nil {
		return nil, err
	}

	config := draftAppConfigVersion.Config
	// 验证和处理模型配置
	if modelConfig, exists := config["model_config"]; exists {
		if modelConfigMap, ok := modelConfig.(map[string]interface{}); ok {
			validatedModelConfig, err := s.llmService.ProcessAndValidateModelConfig(modelConfigMap)
			if err != nil {
				return nil, fmt.Errorf("failed to validate model config: %w", err)
			}

			// 如果模型配置发生变化，更新数据库
			if !s.isModelConfigEqual(modelConfigMap, validatedModelConfig) {
				config["model_config"] = validatedModelConfig
				draftAppConfigVersion.Config = config
				if err := s.repo.UpdateAppConfigVersion(ctx, draftAppConfigVersion); err != nil {
					return nil, fmt.Errorf("failed to update draft app config: %w", err)
				}
			}
			config["model_config"] = validatedModelConfig
		}
	}

	return config, nil
}

// UpdateDraftAppConfig 更新应用的最新草稿配置
func (s *AppService) UpdateDraftAppConfig(ctx context.Context, appID uuid.UUID, draftConfig map[string]interface{}) error {
	// 获取当前用户ID
	accountID, err := util.GetCurrentUserID(ctx)
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
	validatedConfig, err := s.validateDraftAppConfig(draftConfig, accountID)
	if err != nil {
		return err
	}

	// 更新草稿配置
	draftAppConfigVersion.Config = validatedConfig
	if err := s.repo.UpdateAppConfigVersion(ctx, draftAppConfigVersion); err != nil {
		return err
	}

	return nil
}

// validateDraftAppConfig 校验草稿配置
func (s *AppService) validateDraftAppConfig(draftConfig map[string]interface{}, accountID uuid.UUID) (map[string]interface{}, error) {
	// 定义可接受的字段
	acceptableFields := []string{
		"model_config", "dialog_round", "preset_prompt",
		"tools", "workflows", "datasets", "retrieval_config",
		"long_term_memory", "opening_statement", "opening_questions",
		"speech_to_text", "text_to_speech", "suggested_after_answer", "review_config",
	}

	// 验证字段
	for key := range draftConfig {
		found := false
		for _, field := range acceptableFields {
			if key == field {
				found = true
				break
			}
		}
		if !found {
			return nil, errno.ErrValidate.AppendBizMessage("草稿配置包含无效字段: " + key)
		}
	}

	// 验证模型配置
	if modelConfig, exists := draftConfig["model_config"]; exists {
		if modelConfigMap, ok := modelConfig.(map[string]interface{}); ok {
			validatedModelConfig, err := s.llmService.ProcessAndValidateModelConfig(modelConfigMap)
			if err != nil {
				return nil, fmt.Errorf("failed to validate model config: %w", err)
			}
			draftConfig["model_config"] = validatedModelConfig
		} else {
			return nil, errno.ErrValidate.AppendBizMessage("模型配置格式错误")
		}
	}

	// 验证对话轮数
	if dialogRound, exists := draftConfig["dialog_round"]; exists {
		if rounds, ok := dialogRound.(float64); ok {
			if rounds < 0 || rounds > 100 {
				return nil, errno.ErrValidate.AppendBizMessage("对话轮数范围为0-100")
			}
			draftConfig["dialog_round"] = int(rounds)
		} else {
			return nil, errno.ErrValidate.AppendBizMessage("对话轮数必须是数字")
		}
	}

	// 验证预设提示
	if presetPrompt, exists := draftConfig["preset_prompt"]; exists {
		if prompt, ok := presetPrompt.(string); ok {
			if len(prompt) > 2000 {
				return nil, errno.ErrValidate.AppendBizMessage("预设提示长度不能超过2000字符")
			}
		} else {
			return nil, errno.ErrValidate.AppendBizMessage("预设提示必须是字符串")
		}
	}

	// 验证工具配置
	if toolsConfig, exists := draftConfig["tools"]; exists {
		if toolsList, ok := toolsConfig.([]interface{}); ok {
			if len(toolsList) > 5 {
				return nil, errno.ErrValidate.AppendBizMessage("工具数量不能超过5个")
			}
			// 这里可以添加更详细的工具验证逻辑
		} else {
			return nil, errno.ErrValidate.AppendBizMessage("工具配置必须是数组")
		}
	}

	// 验证其他配置字段...
	// 为了简化，这里只验证关键字段，可以根据需要添加更多验证

	return draftConfig, nil
}

// isModelConfigEqual 比较两个模型配置是否相等
func (s *AppService) isModelConfigEqual(config1, config2 map[string]interface{}) bool {
	// 简单的深度比较
	if len(config1) != len(config2) {
		return false
	}

	for key, value1 := range config1 {
		value2, exists := config2[key]
		if !exists {
			return false
		}

		// 对于嵌套的 map，需要递归比较
		if map1, ok1 := value1.(map[string]interface{}); ok1 {
			if map2, ok2 := value2.(map[string]interface{}); ok2 {
				if !s.isModelConfigEqual(map1, map2) {
					return false
				}
			} else {
				return false
			}
		} else if value1 != value2 {
			return false
		}
	}

	return true
}
