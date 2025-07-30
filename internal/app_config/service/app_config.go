package service

import (
	"context"
	"errors"
	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	langchaintool "github.com/tmc/langchaingo/tools"
	"reflect"

	"github.com/crazyfrankie/voidx/internal/app_config/repository"
	"github.com/crazyfrankie/voidx/internal/core/llm"
	"github.com/crazyfrankie/voidx/internal/core/tools/api_tools/entities"
	apitools "github.com/crazyfrankie/voidx/internal/core/tools/api_tools/providers"
	builtin "github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers"
	"github.com/crazyfrankie/voidx/internal/core/workflow"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/consts"
	"github.com/crazyfrankie/voidx/pkg/errno"
)

type AppConfigService struct {
	repo            *repository.AppConfigRepo
	llmMgr          *llm.LanguageModelManager
	builtinProvider *builtin.BuiltinProviderManager
	apiProvider     *apitools.ApiProviderManager
}

func NewAppConfigService(repo *repository.AppConfigRepo, llmMgr *llm.LanguageModelManager,
	builtinProvider *builtin.BuiltinProviderManager, apiProvider *apitools.ApiProviderManager) *AppConfigService {
	return &AppConfigService{
		repo:            repo,
		llmMgr:          llmMgr,
		builtinProvider: builtinProvider,
		apiProvider:     apiProvider,
	}
}

// GetDraftAppConfig 根据传递的应用获取该应用的草稿配置
func (s *AppConfigService) GetDraftAppConfig(ctx context.Context, app *entity.App) (*resp.AppDraftConfigResp, error) {
	// 1. 提取应用的草稿配置
	if app.DraftAppConfigID == uuid.Nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("草稿配置不存在"))
	}

	draftAppConfig, err := s.repo.GetAppConfigVersion(ctx, app.DraftAppConfigID)
	if err != nil {
		return nil, err
	}

	// 2. 校验model_config信息，如果使用了不存在的提供者或者模型，则使用默认值(宽松校验)
	validateModelConfig := s.processAndValidateModelConfig(draftAppConfig.ModelConfig)
	if !s.compareJSON(draftAppConfig.ModelConfig, validateModelConfig) {
		err = s.repo.UpdateAppConfigVersion(ctx, app.DraftAppConfigID, map[string]any{
			"model_config": validateModelConfig,
		})
		if err != nil {
			return nil, err
		}
	}

	// 3. 循环遍历工具列表删除已经被删除的工具信息
	tools, validateTools := s.processAndValidateTools(ctx, draftAppConfig.Tools)

	// 4. 判断是否需要更新草稿配置中的工具列表信息
	if !s.compareJSON(draftAppConfig.Tools, validateTools) {
		err = s.repo.UpdateAppConfigVersion(ctx, app.DraftAppConfigID, map[string]any{
			"tools": tools,
		})
		if err != nil {
			return nil, err
		}
	}

	// 5. 校验知识库列表，如果引用了不存在/被删除的知识库，需要剔除数据并更新，同时获取知识库的额外信息
	datasets, validateDatasets := s.processAndValidateDatasets(ctx, draftAppConfig.Datasets)

	// 6. 判断是否存在已删除的知识库，如果存在则更新
	if !s.compareDatasetSlices(validateDatasets, draftAppConfig.Datasets) {
		err = s.repo.UpdateAppConfigVersion(ctx, app.DraftAppConfigID, map[string]any{
			"datasets": datasets,
		})
		if err != nil {
			return nil, err
		}
	}

	// 7. 校验工作流列表对应的数据
	workflows, validateWorkflows := s.processAndValidateWorkflows(ctx, draftAppConfig.Workflows)
	if !s.compareWorkflowSlices(validateWorkflows, draftAppConfig.Workflows) {
		err = s.repo.UpdateAppConfigVersion(ctx, app.DraftAppConfigID, map[string]any{
			"workflows": workflows,
		})
		if err != nil {
			return nil, err
		}
	}

	// 8. 将数据转换成字典后返回
	return s.processAndTransformAppConfig(validateModelConfig, tools, workflows, datasets, draftAppConfig), nil
}

// GetAppConfig 根据传递的应用获取该应用的运行配置
func (s *AppConfigService) GetAppConfig(ctx context.Context, app *entity.App) (*resp.AppDraftConfigResp, error) {
	// 1. 提取应用的运行配置
	if app.AppConfigID == uuid.Nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("运行配置不存在"))
	}

	appConfig, err := s.repo.GetAppConfigByID(ctx, app.AppConfigID)
	if err != nil {
		return nil, err
	}

	// 2. 校验model_config信息，如果运行时配置里的model_config发生变化则进行更新
	validateModelConfig := s.processAndValidateModelConfig(appConfig.ModelConfig)
	if !s.compareJSON(appConfig.ModelConfig, validateModelConfig) {
		err = s.repo.UpdateAppConfig(ctx, app.AppConfigID, map[string]any{
			"model_config": validateModelConfig,
		})
		if err != nil {
			return nil, err
		}
	}

	// 3. 循环遍历工具列表删除已经被删除的工具信息
	tools, validateTools := s.processAndValidateTools(ctx, appConfig.Tools)

	// 4. 判断是否需要更新配置中的工具列表信息
	if !s.compareJSON(appConfig.Tools, validateTools) {
		err = s.repo.UpdateAppConfig(ctx, app.AppConfigID, map[string]any{
			"tools": validateTools,
		})
		if err != nil {
			return nil, err
		}
	}

	// 5. 校验知识库列表，如果引用了不存在/被删除的知识库，需要剔除数据并更新，同时获取知识库的额外信息
	appDatasetJoins, err := s.repo.GetAppDatasetJoins(ctx, app.ID)
	if err != nil {
		return nil, err
	}

	var originDatasets []string
	for _, join := range appDatasetJoins {
		originDatasets = append(originDatasets, join.ID.String())
	}

	datasets, validateDatasets := s.processAndValidateDatasets(ctx, originDatasets)

	// 6. 判断是否存在已删除的知识库，如果存在则更新
	for _, datasetID := range originDatasets {
		found := false
		for _, validateID := range validateDatasets {
			if datasetID == validateID {
				found = true
				break
			}
		}
		if !found {
			// 删除已不存在的知识库关联
			if datasetUUID, err := uuid.Parse(datasetID); err == nil {
				_ = s.repo.DeleteAppDatasetJoin(ctx, app.ID, datasetUUID)
			}
		}
	}

	// 7. 校验工作流列表对应的数据
	workflows, validateWorkflows := s.processAndValidateWorkflows(ctx, appConfig.Workflows)
	if !s.compareWorkflowSlices(validateWorkflows, appConfig.Workflows) {
		err = s.repo.UpdateAppConfig(ctx, app.AppConfigID, map[string]any{
			"workflows": validateWorkflows,
		})
		if err != nil {
			return nil, err
		}
	}

	// 8. 将数据转换成字典后返回
	return s.processAndTransformAppConfig(
		validateModelConfig,
		tools,
		workflows,
		datasets,
		&entity.AppConfigVersion{
			ModelConfig:          appConfig.ModelConfig,
			DialogRound:          appConfig.DialogRound,
			PresetPrompt:         appConfig.PresetPrompt,
			Tools:                appConfig.Tools,
			Workflows:            appConfig.Workflows,
			RetrievalConfig:      appConfig.RetrievalConfig,
			LongTermMemory:       appConfig.LongTermMemory,
			OpeningStatement:     appConfig.OpeningStatement,
			OpeningQuestions:     appConfig.OpeningQuestions,
			SpeechToText:         appConfig.SpeechToText,
			TextToSpeech:         appConfig.TextToSpeech,
			SuggestedAfterAnswer: appConfig.SuggestedAfterAnswer,
			ReviewConfig:         appConfig.ReviewConfig,
		},
	), nil
}

// GetLangchainToolsByToolsConfig 根据传递的工具配置列表获取langchain工具列表
func (s *AppConfigService) GetLangchainToolsByToolsConfig(ctx context.Context, toolConfigs []map[string]any) ([]langchaintool.Tool, error) {
	// 1. 循环遍历所有工具配置列表信息
	var res []langchaintool.Tool
	for _, tool := range toolConfigs {
		toolType, ok := tool["type"].(string)
		if !ok {
			continue
		}

		// 3. 根据不同的工具类型执行不同的操作
		if toolType == "builtin_tool" {
			// 4. 内置工具，通过builtin_provider_manager获取工具实例
			provider, ok := tool["provider"].(map[string]any)
			if !ok {
				continue
			}
			providerID, ok := provider["id"].(string)
			if !ok {
				continue
			}
			toolInfo, ok := tool["tool"].(map[string]any)
			if !ok {
				continue
			}
			toolName, ok := toolInfo["name"].(string)
			if !ok {
				continue
			}

			// 创建内置工具
			builtinTool := s.builtinProvider.GetTool(providerID, toolName)
			if builtinTool != nil {
				res = append(res, builtinTool.(langchaintool.Tool))
			}
		} else {
			// 5. API工具，首先根据id找到ApiTool记录，然后创建实例
			toolInfo, ok := tool["tool"].(map[string]any)
			if !ok {
				continue
			}
			toolID, ok := toolInfo["id"].(string)
			if !ok {
				continue
			}

			toolUUID, err := uuid.Parse(toolID)
			if err != nil {
				continue
			}

			apiTool, err := s.repo.GetAPIByID(ctx, toolUUID)
			if err != nil || apiTool == nil {
				continue
			}
			apiToolProvider, err := s.repo.GetAPIProviderByID(ctx, apiTool.ProviderID)
			if err != nil || apiToolProvider == nil {
				continue
			}

			// 创建API工具
			apiToolInstance := s.apiProvider.GetTool(&entities.ToolEntity{
				ID:          apiTool.ID.String(),
				Name:        apiTool.Name,
				URL:         apiTool.URL,
				Method:      apiTool.Method,
				Description: apiTool.Description,
				Headers:     apiToolProvider.Headers,
				Parameters:  nil,
			})
			if apiToolInstance != nil {
				res = append(res, apiToolInstance)
			}
		}
	}

	return res, nil
}

// GetLangchainToolsByWorkflowIDs 根据传递的工作流配置列表获取langchain工具列表
func (s *AppConfigService) GetLangchainToolsByWorkflowIDs(ctx context.Context, workflowIDs []uuid.UUID) ([]langchaintool.Tool, error) {
	// 1. 根据传递的工作流id查询工作流记录信息
	var workflows []langchaintool.Tool

	for _, workflowID := range workflowIDs {
		workflowRecord, err := s.repo.GetWorkflowByID(ctx, workflowID)
		if err != nil || workflowRecord == nil {
			continue
		}

		// 检查工作流状态
		if workflowRecord.Status != consts.WorkflowStatusPublished {
			continue
		}

		// 2. 创建工作流工具
		workflowTool, err := workflow.NewWorkflow(map[string]any{
			"account_id":  workflowRecord.AccountID,
			"name":        workflowRecord.Name,
			"description": workflowRecord.Description,
			"nodes":       workflowRecord.Graph["node"],
			"edges":       workflowRecord.Graph["edge"],
		})
		if workflowTool != nil {
			workflows = append(workflows, workflowTool)
		}
	}

	return workflows, nil
}

// processAndTransformAppConfig 根据传递的插件列表、工作流列表、知识库列表以及应用配置创建字典信息
func (s *AppConfigService) processAndTransformAppConfig(modelConfig map[string]any, tools []map[string]any, workflows []map[string]any, datasets []map[string]any, appConfig *entity.AppConfigVersion) *resp.AppDraftConfigResp {
	return &resp.AppDraftConfigResp{
		Id:                   appConfig.ID,
		ModelConfig:          modelConfig,
		DialogRound:          appConfig.DialogRound,
		PresetPrompt:         appConfig.PresetPrompt,
		Tools:                tools,
		Workflows:            workflows,
		Datasets:             datasets,
		RetrievalConfig:      appConfig.RetrievalConfig,
		LongTermMemory:       appConfig.LongTermMemory,
		OpeningStatement:     appConfig.OpeningStatement,
		OpeningQuestions:     appConfig.OpeningQuestions,
		SpeechToText:         appConfig.SpeechToText,
		TextToSpeech:         appConfig.TextToSpeech,
		SuggestedAfterAnswer: appConfig.SuggestedAfterAnswer,
		ReviewConfig:         appConfig.ReviewConfig,
	}
}

// processAndValidateTools 根据传递的原始工具信息进行处理和校验
func (s *AppConfigService) processAndValidateTools(ctx context.Context, toolConfigs []map[string]any) ([]map[string]any, []map[string]any) {
	// 2. 循环遍历工具列表删除已被删除的工具
	var validateTools []map[string]any
	var tools []map[string]any

	for _, tool := range toolConfigs {
		toolType, ok := tool["type"].(string)
		if !ok {
			continue
		}

		if toolType == "builtin_tool" {
			// 3. 查询内置工具提供者，并检测是否存在
			providerID, ok := tool["provider_id"].(string)
			if !ok {
				continue
			}

			// 获取内置工具提供者
			provider := s.builtinProvider.GetProvider(providerID)
			if provider == nil {
				continue
			}

			// 4. 获取提供者下的工具实体，并检测是否存在
			toolID, ok := tool["tool_id"].(string)
			if !ok {
				continue
			}

			toolEntity := provider.GetToolEntity(toolID)
			if toolEntity == nil {
				continue
			}

			// 5. 判断工具的params和草稿中的params是否一致，如果不一致则全部重置为默认值
			params, ok := tool["params"].(map[string]any)
			if !ok {
				params = make(map[string]any)
			}

			// 6. 数据都存在，并且参数已经校验完毕，可以将数据添加到validate_tools
			validateTool := make(map[string]any)
			for k, v := range tool {
				validateTool[k] = v
			}
			validateTool["params"] = params
			validateTools = append(validateTools, validateTool)

			// 7. 组装内置工具展示信息
			providerEntity := provider.ProviderEntity
			tools = append(tools, map[string]any{
				"type": "builtin_tool",
				"provider": map[string]any{
					"id":          providerEntity.Name,
					"name":        providerEntity.Name,
					"label":       providerEntity.Label,
					"icon":        "/api/builtin-tools/" + providerEntity.Name + "/icon",
					"description": providerEntity.Description,
				},
				"tool": map[string]any{
					"id":          toolEntity.Name,
					"name":        toolEntity.Name,
					"label":       toolEntity.Label,
					"description": toolEntity.Description,
					"params":      params,
				},
			})
		} else if toolType == "api_tool" {
			// 8. 查询数据库获取对应的工具记录，并检测是否存在
			providerID, ok := tool["provider_id"].(string)
			if !ok {
				continue
			}

			toolID, ok := tool["tool_id"].(string)
			if !ok {
				continue
			}

			// 通过repository查询API工具
			toolRecord, err := s.repo.GetAPIToolByProviderAndName(ctx, providerID, toolID)
			if err != nil || toolRecord == nil {
				continue
			}

			// 9. 数据校验通过，往validate_tools中添加数据
			validateTools = append(validateTools, tool)

			// 10. 组装api工具展示信息
			// 通过repository查询API提供者
			provider, err := s.repo.GetAPIProviderByID(ctx, toolRecord.ProviderID)
			if err != nil {
				continue
			}

			tools = append(tools, map[string]any{
				"type": "api_tool",
				"provider": map[string]any{
					"id":          provider.ID.String(),
					"name":        provider.Name,
					"label":       provider.Name,
					"icon":        provider.Icon,
					"description": provider.Description,
				},
				"tool": map[string]any{
					"id":          toolRecord.ID.String(),
					"name":        toolRecord.Name,
					"label":       toolRecord.Name,
					"description": toolRecord.Description,
					"params":      map[string]any{},
				},
			})
		}
	}

	return tools, validateTools
}

// processAndValidateDatasets 根据传递的知识库并返回知识库配置与校验后的数据
func (s *AppConfigService) processAndValidateDatasets(ctx context.Context, datasetIDs []string) ([]map[string]any, []string) {
	// 2. 校验知识库配置列表，如果引用了不存在的/被删除的知识库，则需要剔除数据并更新，同时获取知识库的额外信息
	var datasets []map[string]any
	var validateDatasets []string

	// 我们这边需要通过repository查询
	for _, datasetID := range datasetIDs {
		datasetUUID, err := uuid.Parse(datasetID)
		if err != nil {
			continue
		}

		dataset, err := s.repo.GetDatasetByID(ctx, datasetUUID)
		if err != nil || dataset == nil {
			continue
		}

		// 3. 数据存在，添加到结果中
		validateDatasets = append(validateDatasets, datasetID)
		datasets = append(datasets, map[string]any{
			"id":          dataset.ID.String(),
			"name":        dataset.Name,
			"icon":        dataset.Icon,
			"description": dataset.Description,
		})
	}

	return datasets, datasetIDs
}

// processAndValidateModelConfig 根据传递的模型配置处理并校验，随后返回校验后的信息
func (s *AppConfigService) processAndValidateModelConfig(modelConfigMap map[string]any) map[string]any {
	// 1. 提取origin_model_config中provider、model、parameters对应的信息
	modelConfig := map[string]any{
		"provider":   "",
		"model":      "",
		"parameters": map[string]any{},
	}

	if provider, ok := modelConfigMap["provider"]; ok {
		modelConfig["provider"] = provider
	}
	if model, ok := modelConfigMap["model"]; ok {
		modelConfig["model"] = model
	}
	if parameters, ok := modelConfigMap["parameters"]; ok {
		modelConfig["parameters"] = parameters
	}

	// 2. 判断provider是否存在、类型是否正确，如果不符合规则则返回默认值
	provider, ok := modelConfig["provider"].(string)
	if !ok || provider == "" {
		return consts.DefaultAppConfig["model_config"].(map[string]any)
	}

	if _, err := s.llmMgr.GetProvider(provider); err != nil {
		return consts.DefaultAppConfig["model_config"].(map[string]any)
	}

	// 3. 判断model是否存在、类型是否正确，如果不符合则返回默认值
	model, ok := modelConfig["model"].(string)
	if !ok || model == "" {
		return consts.DefaultAppConfig["model_config"].(map[string]any)
	}

	if _, err := s.llmMgr.GetProvider(provider); err != nil {
		return consts.DefaultAppConfig["model_config"].(map[string]any)
	}

	// 4. 判断parameters信息类型是否错误，如果错误则设置为默认值
	if _, ok := modelConfig["parameters"].(map[string]any); !ok {
		modelConfig["parameters"] = s.getModelDefaultParameters(provider, model)
	}

	// 5. 验证参数
	validatedParams := s.validateModelParameters(provider, model, modelConfig["parameters"].(map[string]any))
	modelConfig["parameters"] = validatedParams

	return modelConfig
}

// processAndValidateWorkflows 根据传递的工作流列表并返回工作流配置和校验后的数据
func (s *AppConfigService) processAndValidateWorkflows(ctx context.Context, workflowIDs []string) ([]map[string]any, []string) {
	// 1. 校验工作流配置列表，如果引用了不存在/被删除的工作流，则需要剔除数据并更新，同时获取工作流的额外信息
	var workflows []map[string]any
	var validateWorkflows []string

	for _, workflowID := range workflowIDs {
		workflowUUID, err := uuid.Parse(workflowID)
		if err != nil {
			continue
		}

		wf, err := s.repo.GetWorkflowByID(ctx, workflowUUID)
		if err != nil || wf == nil {
			continue
		}

		// 检查工作流状态
		if wf.Status != consts.WorkflowStatusPublished {
			continue
		}

		// 3. 数据存在且已发布，添加到结果中
		validateWorkflows = append(validateWorkflows, workflowID)
		workflows = append(workflows, map[string]any{
			"id":          wf.ID.String(),
			"name":        wf.Name,
			"icon":        wf.Icon,
			"description": wf.Description,
		})
	}

	return workflows, validateWorkflows
}

// 辅助方法

// compareJSON 比较两个JSON是否相等
func (s *AppConfigService) compareJSON(a, b any) bool {
	if a == nil && b == nil {
		return true
	}

	aValue := reflect.ValueOf(a)
	bValue := reflect.ValueOf(b)

	if aValue.Kind() == reflect.Slice && bValue.Kind() == reflect.Slice {
		if (aValue.IsNil() || aValue.Len() == 0) &&
			(bValue.IsNil() || bValue.Len() == 0) {
			return true
		}
	}

	aJson, _ := sonic.Marshal(a)
	bJson, _ := sonic.Marshal(b)

	return string(aJson) == string(bJson)
}

// compareDatasetSlices 比较两个数据集切片是否相等
func (s *AppConfigService) compareDatasetSlices(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	aSet := make(map[string]bool)
	for _, item := range a {
		aSet[item] = true
	}

	for _, item := range b {
		if !aSet[item] {
			return false
		}
	}

	return true
}

// compareWorkflowSlices 比较两个工作流切片是否相等
func (s *AppConfigService) compareWorkflowSlices(a []string, b []string) bool {
	return s.compareDatasetSlices(a, b)
}

// validateModelParameters 验证模型参数
func (s *AppConfigService) validateModelParameters(provider, model string, parameters map[string]any) map[string]any {
	// 使用LLM管理器验证模型参数
	if s.llmMgr != nil {
		if err := s.llmMgr.ValidateModelConfig(provider, model, parameters); err == nil {
			return parameters
		}
	}
	// 验证失败时返回默认参数
	return s.getModelDefaultParameters(provider, model)
}

// getModelDefaultParameters 获取模型默认参数
func (s *AppConfigService) getModelDefaultParameters(provider, model string) map[string]any {
	// 实现获取模型默认参数逻辑
	if s.llmMgr != nil {
		if modelEntity, err := s.llmMgr.GetModelEntity(provider, model); err == nil {
			defaultParams := make(map[string]any)
			for _, param := range modelEntity.Parameters {
				defaultParams[param.Name] = param.Default
			}
			return defaultParams
		}
	}
	// 返回通用默认参数
	return map[string]any{
		"temperature": 0.7,
		"max_tokens":  1000,
	}
}
