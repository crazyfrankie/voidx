package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/app/repository"
	"github.com/crazyfrankie/voidx/internal/app_config"
	"github.com/crazyfrankie/voidx/internal/conversation"
	"github.com/crazyfrankie/voidx/internal/core/agent"
	agenteneity "github.com/crazyfrankie/voidx/internal/core/agent/entities"
	"github.com/crazyfrankie/voidx/internal/core/llm"
	llmentity "github.com/crazyfrankie/voidx/internal/core/llm/entities"
	"github.com/crazyfrankie/voidx/internal/core/memory"
	"github.com/crazyfrankie/voidx/internal/core/tools/api_tools/providers"
	builtin "github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/internal/retriever"
	"github.com/crazyfrankie/voidx/internal/upload"
	"github.com/crazyfrankie/voidx/pkg/dalle"
	"github.com/crazyfrankie/voidx/pkg/logs"
	"github.com/crazyfrankie/voidx/pkg/sonic"
	"github.com/crazyfrankie/voidx/pkg/util"
	"github.com/crazyfrankie/voidx/types/consts"
	"github.com/crazyfrankie/voidx/types/errno"
)

type AppService struct {
	repo                *repository.AppRepo
	appConfigService    *app_config.Service
	conversationService *conversation.Service
	apiProvider         *providers.APIProviderManager
	builtinProvider     *builtin.BuiltinProviderManager
	retrieverSvc        *retriever.Service
	ossSvc              *upload.Service
	agentManager        *agent.AgentQueueManagerFactory
	llmService          *llm.LanguageModelManager
	llm                 llmentity.BaseLanguageModel
	tokenBufMem         *memory.TokenBufferMemory
	activeSessions      sync.Map
}

func NewAppService(repo *repository.AppRepo,
	appConfigSvc *app_config.Service, conversationSvc *conversation.Service,
	retrieverSvc *retriever.Service, ossSvc *upload.Service, apiProvider *providers.APIProviderManager,
	builtinProvider *builtin.BuiltinProviderManager, agentManager *agent.AgentQueueManagerFactory,
	llmService *llm.LanguageModelManager, tokenBufMem *memory.TokenBufferMemory, llm llmentity.BaseLanguageModel) *AppService {
	return &AppService{
		repo:                repo,
		appConfigService:    appConfigSvc,
		conversationService: conversationSvc,
		retrieverSvc:        retrieverSvc,
		ossSvc:              ossSvc,
		apiProvider:         apiProvider,
		builtinProvider:     builtinProvider,
		agentManager:        agentManager,
		llmService:          llmService,
		tokenBufMem:         tokenBufMem,
		llm:                 llm,
	}
}

// AutoCreateApp 根据传递的应用名称、描述、账号id利用AI创建一个Agent智能体
func (s *AppService) AutoCreateApp(ctx context.Context, name, description string, accountID uuid.UUID) error {
	// 创建DallEApiWrapper包装器
	dalleCli := dalle.NewClient(os.Getenv("OPENAI_API_KEY"))

	iconPrompt := fmt.Sprintf(consts.GenerateIconPromptTemplate, name, description)

	var iconURL string
	var generatedPresetPrompt string

	// 调用 AI 生成图片和 Prompt
	size := dalle.Large
	userID := accountID.String()
	res, err := dalleCli.Generate(iconPrompt, &size, nil, &userID, nil)
	if err != nil {
		return err
	}
	iconURL = res[0].URL

	// 使用新的LLM服务生成预设提示词
	model, err := s.llmService.CreateModel(ctx, "openai", "gpt-3.5-turbo", map[string]any{})
	if err != nil {
		return err
	}

	messages := []*schema.Message{
		schema.SystemMessage(consts.OptimizePromptTemplate),
		schema.UserMessage(fmt.Sprintf("应用名称: %s\n\n应用描述: %s", name, description)),
	}

	response, err := model.Generate(ctx, messages)
	if err != nil {
		return err
	}
	generatedPresetPrompt = response.Content

	// 5. 将图片下载到本地后上传到OSS中
	if iconURL != "" {
		httpResp, err := http.Get(iconURL)
		if err != nil {
			return errno.ErrInternalServer.AppendBizMessage(errors.New("下载生成的图标失败"))
		}
		defer httpResp.Body.Close()

		if httpResp.StatusCode != 200 {
			return errno.ErrInternalServer.AppendBizMessage(errors.New("生成应用icon图标出错"))
		}

		data, err := io.ReadAll(httpResp.Body)
		if err != nil {
			return err
		}
		// 上传到OSS
		uploadedIconURL, err := s.ossSvc.UploadFile(ctx, data, true, "icon.png", accountID)
		if err != nil {
			return errno.ErrInternalServer.AppendBizMessage(errors.New("上传图标失败"))
		}
		iconURL = uploadedIconURL.URL
	}

	// 创建应用记录并刷新数据，从而可以拿到应用id
	app := &entity.App{
		ID:          uuid.New(),
		AccountID:   accountID,
		Name:        name,
		Icon:        iconURL,
		Description: description,
		Status:      consts.AppStatusDraft,
	}

	_, err = s.repo.CreateApp(ctx, app)
	if err != nil {
		return err
	}

	// 添加草稿记录
	defaultConfig := consts.DefaultAppConfig
	if generatedPresetPrompt != "" {
		defaultConfig["preset_prompt"] = generatedPresetPrompt
	}

	appConfigVersion := &entity.AppConfigVersion{
		ID:           uuid.New(),
		AppID:        app.ID,
		Version:      0,
		ConfigType:   consts.AppConfigTypeDraft,
		ModelConfig:  defaultConfig,
		PresetPrompt: generatedPresetPrompt,
	}

	_, err = s.repo.CreateAppConfigVersion(ctx, appConfigVersion)
	if err != nil {
		return err
	}

	// 更新应用配置id
	return s.repo.UpdateApp(ctx, app.ID, map[string]any{
		"id":                  app.ID,
		"draft_app_config_id": appConfigVersion.ID,
	})
}

// CreateApp 创建Agent应用服务
func (s *AppService) CreateApp(ctx context.Context, accountID uuid.UUID, req req.CreateAppReq) (*entity.App, error) {
	// 创建应用记录，并刷新数据，从而可以拿到应用id
	app := &entity.App{
		ID:          uuid.New(),
		AccountID:   accountID,
		Name:        req.Name,
		Icon:        req.Icon,
		Description: req.Description,
		Status:      consts.AppStatusDraft,
	}

	_, err := s.repo.CreateApp(ctx, app)
	if err != nil {
		return nil, err
	}

	// 添加草稿记录
	appConfigVersion := &entity.AppConfigVersion{
		ID:          uuid.New(),
		AppID:       app.ID,
		Version:     0,
		ConfigType:  consts.AppConfigTypeDraft,
		ModelConfig: consts.DefaultAppConfig,
	}

	_, err = s.repo.CreateAppConfigVersion(ctx, appConfigVersion)
	if err != nil {
		return nil, err
	}

	// 为应用添加草稿配置id
	err = s.repo.UpdateApp(ctx, app.ID, map[string]any{
		"id":                  app.ID,
		"draft_app_config_id": appConfigVersion.ID,
	})
	if err != nil {
		return nil, err
	}

	// 返回创建的应用记录
	return app, nil
}

// GetApp 根据传递的id获取应用的基础信息
func (s *AppService) GetApp(ctx context.Context, appID uuid.UUID, accountID uuid.UUID) (*resp.AppResp, error) {
	// 1. 查询数据库获取应用基础信息
	app, err := s.repo.GetAppByID(ctx, appID)
	if err != nil {
		return nil, err
	}

	// 2. 判断应用是否存在
	if app == nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("该应用不存在，请核实后重试"))
	}

	// 3. 判断当前账号是否有权限访问该应用
	if app.AccountID != accountID {
		return nil, errno.ErrForbidden.AppendBizMessage(errors.New("当前账号无权限访问该应用，请核实后尝试"))
	}

	draft, err := s.repo.GetDraftAppConfigVersion(ctx, app.ID)
	if err != nil {
		return nil, err
	}

	return &resp.AppResp{
		ID:                  app.ID,
		DebugConversationID: app.DebugConversationID,
		Name:                app.Name,
		Icon:                app.Icon,
		Description:         app.Description,
		Status:              app.Status.String(),
		DraftUtime:          draft.Utime,
		Ctime:               app.Ctime,
		Utime:               app.Utime,
	}, nil
}

// RawGetApp 根据传递的id获取应用的基础信息
func (s *AppService) RawGetApp(ctx context.Context, appID uuid.UUID, accountID uuid.UUID) (*entity.App, error) {
	// 1. 查询数据库获取应用基础信息
	app, err := s.repo.GetAppByID(ctx, appID)
	if err != nil {
		return nil, err
	}

	// 2. 判断应用是否存在
	if app == nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("该应用不存在，请核实后重试"))
	}

	// 3. 判断当前账号是否有权限访问该应用
	if app.AccountID != accountID {
		return nil, errno.ErrForbidden.AppendBizMessage(errors.New("当前账号无权限访问该应用，请核实后尝试"))
	}

	return app, nil
}

// DeleteApp 根据传递的应用id+账号，删除指定的应用信息，目前仅删除应用基础信息即可
func (s *AppService) DeleteApp(ctx context.Context, appID uuid.UUID, accountID uuid.UUID) error {
	app, err := s.GetApp(ctx, appID, accountID)
	if err != nil {
		return err
	}

	err = s.repo.DeleteApp(ctx, app.ID)
	if err != nil {
		return err
	}

	return nil
}

// UpdateApp 根据传递的应用id+账号+信息，更新指定的应用
func (s *AppService) UpdateApp(ctx context.Context, appID uuid.UUID, accountID uuid.UUID, updateReq req.UpdateAppReq) error {
	_, err := s.GetApp(ctx, appID, accountID)
	if err != nil {
		return err
	}

	updates := make(map[string]any)
	if updateReq.Name != "" {
		updates["name"] = updateReq.Name
	}
	if updateReq.Icon != "" {
		updates["icon"] = updateReq.Icon
	}
	if updateReq.Description != "" {
		updates["description"] = updateReq.Description
	}

	return s.repo.UpdatesApp(ctx, appID, updates)
}

// CopyApp 根据传递的应用id，拷贝Agent相关信息并创建一个新Agent
func (s *AppService) CopyApp(ctx context.Context, appID uuid.UUID, accountID uuid.UUID) error {
	// 1. 获取App+草稿配置，并校验权限
	app, err := s.RawGetApp(ctx, appID, accountID)
	if err != nil {
		return err
	}

	var draftAppConfig *entity.AppConfigVersion
	if app.DraftAppConfigID != uuid.Nil {
		draftAppConfig, err = s.repo.GetAppConfigVersion(ctx, app.DraftAppConfigID)
		if err != nil {
			return err
		}
	}

	// 2. 创建一个新的应用记录
	newApp := &entity.App{
		ID:          uuid.New(),
		AccountID:   app.AccountID,
		Name:        app.Name,
		Icon:        app.Icon,
		Description: app.Description,
		Status:      consts.AppStatusDraft,
	}

	_, err = s.repo.CreateApp(ctx, newApp)
	if err != nil {
		return err
	}

	// 3. 添加草稿配置
	if draftAppConfig != nil {
		newDraftAppConfig := &entity.AppConfigVersion{
			ID:                   uuid.New(),
			AppID:                newApp.ID,
			Version:              0,
			ConfigType:           consts.AppConfigTypeDraft,
			ModelConfig:          draftAppConfig.ModelConfig,
			DialogRound:          draftAppConfig.DialogRound,
			PresetPrompt:         draftAppConfig.PresetPrompt,
			Tools:                draftAppConfig.Tools,
			Workflows:            draftAppConfig.Workflows,
			Datasets:             draftAppConfig.Datasets,
			RetrievalConfig:      draftAppConfig.RetrievalConfig,
			LongTermMemory:       draftAppConfig.LongTermMemory,
			OpeningStatement:     draftAppConfig.OpeningStatement,
			OpeningQuestions:     draftAppConfig.OpeningQuestions,
			SpeechToText:         draftAppConfig.SpeechToText,
			TextToSpeech:         draftAppConfig.TextToSpeech,
			SuggestedAfterAnswer: draftAppConfig.SuggestedAfterAnswer,
			ReviewConfig:         draftAppConfig.ReviewConfig,
		}

		_, err = s.repo.CreateAppConfigVersion(ctx, newDraftAppConfig)
		if err != nil {
			return err
		}

		// 4. 更新应用的草稿配置id
		err = s.repo.UpdateApp(ctx, app.ID, map[string]any{
			"id":                  app.ID,
			"draft_app_config_id": newDraftAppConfig.ID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// GetAppsWithPage 根据传递的分页参数获取当前登录账号下的应用分页列表数据
func (s *AppService) GetAppsWithPage(ctx context.Context, accountID uuid.UUID, req req.GetAppsWithPageReq) (*resp.GetAppsWithPageResp, error) {
	// 执行分页操作
	apps, totalRecords, totalPages, err := s.repo.GetAppsWithPage(ctx, accountID, req.CurrentPage, req.PageSize, req.SearchWord)
	if err != nil {
		return nil, err
	}

	appCfgs, err := s.getAppConfigs(ctx, apps)
	if err != nil {
		return nil, err
	}
	if len(apps) != len(appCfgs) {
		return nil, errno.ErrInternalServer
	}

	res := &resp.GetAppsWithPageResp{
		Paginator: resp.Paginator{
			CurrentPage: req.CurrentPage,
			PageSize:    req.PageSize,
			TotalPage:   int(totalPages),
			TotalRecord: int(totalRecords),
		},
	}
	res.List = make([]resp.AppWithPage, 0, len(apps))
	for i, app := range apps {
		res.List = append(res.List, resp.AppWithPage{
			ID:           app.ID,
			Name:         app.Name,
			Icon:         app.Icon,
			Description:  app.Description,
			PresetPrompt: appCfgs[i].PresetPrompt,
			ModelConfig:  appCfgs[i].ModelConfig,
			Status:       string(app.Status),
			Ctime:        app.Ctime,
			Utime:        app.Utime,
		})
	}

	return res, nil
}

func (s *AppService) getAppConfigs(ctx context.Context, apps []*entity.App) ([]*resp.AppDraftConfigResp, error) {
	appConfigs := make([]*resp.AppDraftConfigResp, 0, len(apps))
	for _, app := range apps {
		var appCfg *resp.AppDraftConfigResp
		var err error
		if app.Status == consts.AppStatusDraft {
			appCfg, err = s.appConfigService.GetDraftAppConfig(ctx, app)
			if err != nil {
				continue
			}
		} else {
			appCfg, err = s.appConfigService.GetAppConfig(ctx, app)
			if err != nil {
				continue
			}
		}
		appConfigs = append(appConfigs, appCfg)
	}

	return appConfigs, nil
}

// GetDraftAppConfig 根据传递的应用id，获取指定的应用草稿配置信息
func (s *AppService) GetDraftAppConfig(ctx context.Context, appID uuid.UUID, accountID uuid.UUID) (*resp.AppDraftConfigResp, error) {
	app, err := s.RawGetApp(ctx, appID, accountID)
	if err != nil {
		return nil, err
	}

	return s.appConfigService.GetDraftAppConfig(ctx, app)
}

// UpdateDraftAppConfig 根据传递的应用id+草稿配置修改指定应用的最新草稿
func (s *AppService) UpdateDraftAppConfig(ctx context.Context, appID uuid.UUID, accountID uuid.UUID, draftAppConfig map[string]any) error {
	// 1. 获取应用信息并校验
	app, err := s.RawGetApp(ctx, appID, accountID)
	if err != nil {
		return err
	}

	// 2. 校验传递的草稿配置信息
	validatedConfig, err := s.validateDraftAppConfig(draftAppConfig, accountID)
	if err != nil {
		return err
	}

	// 3. 获取当前应用的最新草稿信息
	if app.DraftAppConfigID == uuid.Nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("草稿配置不存在"))
	}

	_, err = s.repo.GetAppConfigVersion(ctx, app.DraftAppConfigID)
	if err != nil {
		return err
	}

	// app.DraftAppConfigID, validatedConfig
	err = s.repo.UpdateAppConfigVersion(ctx, app.DraftAppConfigID, map[string]any{
		"model_config": validatedConfig,
	})
	if err != nil {
		return err
	}

	return nil
}

// PublishDraftAppConfig 根据传递的应用id+账号，发布/更新指定的应用草稿配置为运行时配置
func (s *AppService) PublishDraftAppConfig(ctx context.Context, appID uuid.UUID, accountID uuid.UUID) error {
	// 1. 获取应用的信息以及草稿信息
	app, err := s.RawGetApp(ctx, appID, accountID)
	if err != nil {
		return err
	}

	draftAppConfig, err := s.GetDraftAppConfig(ctx, appID, accountID)
	if err != nil {
		return err
	}

	// 2. 创建应用运行配置（在这里暂时不删除历史的运行配置）
	// 处理工具配置
	var processedTools []map[string]any
	if tools := draftAppConfig.Tools; len(tools) > 0 {
		for _, tool := range tools {
			processedTool := map[string]any{
				"type": tool["type"],
			}
			if provider, exists := tool["provider"]; exists {
				if providerMap, ok := provider.(map[string]any); ok {
					processedTool["provider_id"] = providerMap["id"]
				}
			}
			if toolInfo, exists := tool["tool"]; exists {
				if toolInfoMap, ok := toolInfo.(map[string]any); ok {
					processedTool["tool_id"] = toolInfoMap["name"]
					processedTool["params"] = toolInfoMap["params"]
				}
			}
			processedTools = append(processedTools, processedTool)
		}
	}

	// 处理工作流配置
	var workflowIDs []string
	if workflows := draftAppConfig.Workflows; len(workflows) > 0 {
		for _, workflow := range workflows {
			if idStr, ok := workflow["id"].(string); ok {
				workflowIDs = append(workflowIDs, idStr)
			}
		}
	}

	appConfig := &entity.AppConfig{
		ID:                   uuid.New(),
		AppID:                appID,
		ModelConfig:          draftAppConfig.ModelConfig,
		DialogRound:          draftAppConfig.DialogRound,
		PresetPrompt:         draftAppConfig.PresetPrompt,
		Tools:                processedTools,
		Workflows:            workflowIDs,
		RetrievalConfig:      draftAppConfig.RetrievalConfig,
		LongTermMemory:       draftAppConfig.LongTermMemory,
		OpeningStatement:     draftAppConfig.OpeningStatement,
		OpeningQuestions:     draftAppConfig.OpeningQuestions,
		SpeechToText:         draftAppConfig.SpeechToText,
		TextToSpeech:         draftAppConfig.TextToSpeech,
		SuggestedAfterAnswer: draftAppConfig.SuggestedAfterAnswer,
		ReviewConfig:         draftAppConfig.ReviewConfig,
	}

	_, err = s.repo.CreateAppConfig(ctx, appConfig)
	if err != nil {
		return err
	}

	// 3. 更新应用关联的运行时配置以及状态
	err = s.repo.UpdateApp(ctx, appID, map[string]any{
		"app_config_id": appConfig.ID,
		"status":        consts.AppStatusPublished,
	})
	if err != nil {
		return err
	}

	// 4. 先删除原有的知识库关联记录
	err = s.repo.DeleteAppDatasetJoins(ctx, appID)
	if err != nil {
		return err
	}

	// 5. 新增新的知识库关联记录
	if datasets := draftAppConfig.Datasets; len(datasets) > 0 {
		for _, dataset := range datasets {
			if idStr, ok := dataset["id"].(string); ok {
				if datasetID, err := uuid.Parse(idStr); err == nil {
					err := s.repo.CreateAppDatasetJoin(ctx, appID, datasetID)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	// 6. 获取应用草稿记录
	if app.DraftAppConfigID == uuid.Nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("草稿配置不存在"))
	}

	draftAppConfigCopy, err := s.repo.GetAppConfigVersion(ctx, app.DraftAppConfigID)
	if err != nil {
		return err
	}

	// 7. 获取当前最大的发布版本
	maxVersion, err := s.repo.GetMaxPublishedVersion(ctx, appID)
	if err != nil {
		maxVersion = 0
	}

	// 8. 新增发布历史配置
	publishedVersion := &entity.AppConfigVersion{
		ID:                   uuid.New(),
		AppID:                appID,
		Version:              maxVersion + 1,
		ConfigType:           consts.AppConfigTypePublished,
		ModelConfig:          draftAppConfigCopy.ModelConfig,
		DialogRound:          draftAppConfigCopy.DialogRound,
		PresetPrompt:         draftAppConfigCopy.PresetPrompt,
		Tools:                draftAppConfigCopy.Tools,
		Workflows:            draftAppConfigCopy.Workflows,
		Datasets:             draftAppConfigCopy.Datasets,
		RetrievalConfig:      draftAppConfigCopy.RetrievalConfig,
		LongTermMemory:       draftAppConfigCopy.LongTermMemory,
		OpeningStatement:     draftAppConfigCopy.OpeningStatement,
		OpeningQuestions:     draftAppConfigCopy.OpeningQuestions,
		SpeechToText:         draftAppConfigCopy.SpeechToText,
		TextToSpeech:         draftAppConfigCopy.TextToSpeech,
		SuggestedAfterAnswer: draftAppConfigCopy.SuggestedAfterAnswer,
		ReviewConfig:         draftAppConfigCopy.ReviewConfig,
	}

	_, err = s.repo.CreateAppConfigVersion(ctx, publishedVersion)
	if err != nil {
		return err
	}

	return nil
}

// CancelPublishAppConfig 根据传递的应用id+账号，取消发布指定的应用配置
func (s *AppService) CancelPublishAppConfig(ctx context.Context, appID uuid.UUID, accountID uuid.UUID) error {
	// 1. 获取应用信息并校验权限
	app, err := s.RawGetApp(ctx, appID, accountID)
	if err != nil {
		return err
	}

	// 2. 检测下当前应用的状态是否为已发布
	if app.Status != consts.AppStatusPublished {
		return errno.ErrValidate.AppendBizMessage(errors.New("当前应用未发布，请核实后重试"))
	}

	// 3. 修改账号的发布状态，并清空关联配置id
	err = s.repo.UpdateApp(ctx, appID, map[string]any{
		"status":        consts.AppStatusDraft,
		"app_config_Id": uuid.Nil,
	})
	if err != nil {
		return err
	}

	// 4. 删除应用关联的知识库信息
	err = s.repo.DeleteAppDatasetJoins(ctx, appID)
	if err != nil {
		return err
	}

	return nil
}

// GetPublishHistoriesWithPage 根据传递的应用id+请求数据，获取指定应用的发布历史配置列表信息
func (s *AppService) GetPublishHistoriesWithPage(ctx context.Context, appID uuid.UUID, req req.GetPublishHistoriesWithPageReq, accountID uuid.UUID) ([]resp.GetPublishHistoriesWithPageResp, resp.Paginator, error) {
	// 1. 获取应用信息并校验权限
	_, err := s.GetApp(ctx, appID, accountID)
	if err != nil {
		return nil, resp.Paginator{}, err
	}

	// 2. 执行分页并获取数据
	appConfigVersions, paginator, err := s.repo.GetPublishHistoriesWithPage(ctx, appID, req)
	if err != nil {
		return nil, resp.Paginator{}, err
	}

	res := make([]resp.GetPublishHistoriesWithPageResp, 0, len(appConfigVersions))
	for _, a := range appConfigVersions {
		res = append(res, resp.GetPublishHistoriesWithPageResp{
			ID:      a.ID,
			Version: a.Version,
			Ctime:   a.Ctime,
		})
	}

	return res, paginator, nil
}

// FallbackHistoryToDraft 根据传递的应用id、历史配置版本id、账号信息，回退特定配置到草稿
func (s *AppService) FallbackHistoryToDraft(ctx context.Context, appID, appConfigVersionID uuid.UUID, accountID uuid.UUID) (*entity.AppConfigVersion, error) {
	// 1. 校验应用权限并获取信息
	app, err := s.RawGetApp(ctx, appID, accountID)
	if err != nil {
		return nil, err
	}

	// 2. 查询指定的历史版本配置id
	appConfigVersion, err := s.repo.GetAppConfigVersion(ctx, appConfigVersionID)
	if err != nil {
		return nil, err
	}

	if appConfigVersion == nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("该历史版本配置不存在，请核实后重试"))
	}

	// 3. 校验历史版本配置信息（剔除已删除的工具、知识库、工作流）
	draftAppConfigDict := map[string]any{
		"model_config":      appConfigVersion.ModelConfig,
		"dialog_round":      appConfigVersion.DialogRound,
		"preset_prompt":     appConfigVersion.PresetPrompt,
		"tools":             appConfigVersion.Tools,
		"workflows":         appConfigVersion.Workflows,
		"datasets":          appConfigVersion.Datasets,
		"long_term_memory":  appConfigVersion.LongTermMemory,
		"opening_statement": appConfigVersion.OpeningStatement,
		"opening_questions": appConfigVersion.OpeningQuestions,
		"speech_to_text":    appConfigVersion.SpeechToText,
		"text_to_speech":    appConfigVersion.TextToSpeech,
		"review_config":     appConfigVersion.ReviewConfig,
	}

	// 4. 校验历史版本配置信息
	validatedConfig, err := s.validateDraftAppConfig(draftAppConfigDict, accountID)
	if err != nil {
		return nil, err
	}

	// 5. 更新草稿配置信息
	if app.DraftAppConfigID == uuid.Nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("草稿配置不存在"))
	}

	err = s.repo.UpdateAppConfigVersion(ctx, app.DraftAppConfigID, map[string]any{
		"model_config": validatedConfig,
	})
	if err != nil {
		return nil, err
	}

	draftAppConfigRecord, err := s.repo.GetAppConfigVersion(ctx, app.DraftAppConfigID)
	if err != nil {
		return nil, err
	}

	return draftAppConfigRecord, nil
}

// GetDebugConversationSummary 根据传递的应用id+账号获取指定应用的调试会话长期记忆
func (s *AppService) GetDebugConversationSummary(ctx context.Context, appID uuid.UUID, accountID uuid.UUID) (string, error) {
	// 1. 获取应用信息并校验权限
	app, err := s.GetApp(ctx, appID, accountID)
	if err != nil {
		return "", err
	}

	// 2. 获取应用的草稿配置，并校验长期记忆是否启用
	draftAppConfig, err := s.GetDraftAppConfig(ctx, appID, accountID)
	if err != nil {
		return "", err
	}

	if longTermMemory := draftAppConfig.LongTermMemory; longTermMemory != nil {
		if enable, ok := longTermMemory["enable"].(bool); !ok || !enable {
			return "", errno.ErrValidate.AppendBizMessage(errors.New("该应用并未开启长期记忆，无法获取"))
		}
	}

	// 3. 获取调试会话
	if app.DebugConversationID == uuid.Nil {
		return "", errno.ErrNotFound.AppendBizMessage(errors.New("调试会话不存在"))
	}

	convers, err := s.conversationService.GetConversationByID(ctx, app.DebugConversationID)
	if err != nil {
		return "", err
	}

	return convers.Summary, nil
}

// UpdateDebugConversationSummary 根据传递的应用id+总结更新指定应用的调试长期记忆
func (s *AppService) UpdateDebugConversationSummary(ctx context.Context, appID uuid.UUID, accountID uuid.UUID, summary string) error {
	// 1. 获取应用信息并校验权限
	app, err := s.GetApp(ctx, appID, accountID)
	if err != nil {
		return err
	}

	// 2. 获取应用的草稿配置，并校验长期记忆是否启用
	draftAppConfig, err := s.GetDraftAppConfig(ctx, appID, accountID)
	if err != nil {
		return err
	}

	if longTermMemory := draftAppConfig.LongTermMemory; longTermMemory != nil {
		if enable, ok := longTermMemory["enable"].(bool); !ok || !enable {
			return errno.ErrValidate.AppendBizMessage(errors.New("该应用并未开启长期记忆，无法获取"))
		}
	}

	// 3. 更新应用长期记忆
	if app.DebugConversationID == uuid.Nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("调试会话不存在"))
	}

	err = s.conversationService.UpdateConversationSummary(ctx, app.DebugConversationID, summary)
	if err != nil {
		return err
	}

	return nil
}

// DeleteDebugConversation 根据传递的应用id，删除指定的应用调试会话
func (s *AppService) DeleteDebugConversation(ctx context.Context, appID uuid.UUID, accountID uuid.UUID) error {
	// 1. 获取应用信息并校验权限
	app, err := s.GetApp(ctx, appID, accountID)
	if err != nil {
		return err
	}

	// 2. 判断是否存在debug_conversation_id这个数据，如果不存在表示没有会话，无需执行任何操作
	if app.DebugConversationID == uuid.Nil {
		return nil
	}

	// 3. 否则将debug_conversation_id的值重置为None
	err = s.repo.UpdateApp(ctx, app.ID, map[string]any{
		"id":                  app.ID,
		"draft_app_config_id": uuid.Nil,
	})
	if err != nil {
		return err
	}

	return nil
}

// DebugChat 根据传递的应用id+提问query向特定的应用发起会话调试
func (s *AppService) DebugChat(ctx context.Context, appID uuid.UUID, accountID uuid.UUID, chatReq req.DebugChatReq) (<-chan string, error) {
	// 1. 获取应用信息并校验权限
	app, err := s.GetApp(ctx, appID, accountID)
	if err != nil {
		return nil, err
	}

	// 2. 获取应用的最新草稿配置信息
	draftAppConfig, err := s.GetDraftAppConfig(ctx, appID, accountID)
	if err != nil {
		return nil, err
	}

	// 3. 获取当前应用的调试会话信息
	var debugConversation *entity.Conversation
	if app.DebugConversationID != uuid.Nil {
		debugConversation, err = s.conversationService.GetConversationByID(ctx, app.DebugConversationID)
		if err != nil {
			return nil, err
		}
	} else {
		// 创建调试会话
		debugConversation, err = s.conversationService.RawCreateConversation(ctx, appID, accountID)
		if err != nil {
			return nil, err
		}

		// 更新应用的调试会话ID
		err = s.repo.UpdateApp(ctx, appID, map[string]any{
			"debug_conversation_id": debugConversation.ID,
		})
		if err != nil {
			return nil, err
		}
	}

	// 4. 新建一条消息记录
	message, err := s.conversationService.RawCreateMessage(ctx, &entity.Message{
		ID:             uuid.New(),
		AppID:          appID,
		ConversationID: debugConversation.ID,
		InvokeFrom:     consts.InvokeFromDebugger,
		CreatedBy:      accountID,
		Query:          chatReq.Query,
		ImageUrls:      chatReq.ImageUrls,
		Status:         consts.MessageStatusNormal,
	})
	if err != nil {
		return nil, err
	}

	// 6. 使用TokenBufferMemory用于提取短期记忆
	s.tokenBufMem.WithConversationID(debugConversation.ID)
	history, err := s.tokenBufMem.GetHistoryPromptMessages(2000, 10)
	if err != nil {
		return nil, err
	}

	// 7. 将草稿配置中的tools转换成eino工具
	tools, err := s.appConfigService.GetToolsByToolsConfig(ctx, draftAppConfig.Tools)
	if err != nil {
		return nil, err
	}

	// 8. 检测是否关联了知识库
	if draftAppConfig.Datasets != nil {
		datasets := make([]uuid.UUID, 0, len(draftAppConfig.Datasets))
		for _, dataset := range draftAppConfig.Datasets {
			datasets = append(datasets, dataset["id"].(uuid.UUID))
		}
		datasetRetrieval, err := s.retrieverSvc.CreateToolFromSearch(ctx, accountID, datasets, consts.RetrievalSourceApp, draftAppConfig.ReviewConfig)
		if err != nil {
			return nil, err
		}
		tools = append(tools, datasetRetrieval)
	}

	// 9. 检测是否关联工作流
	if draftAppConfig.Workflows != nil {
		workflows := make([]uuid.UUID, 0, len(draftAppConfig.Workflows))
		for _, workflow := range draftAppConfig.Workflows {
			workflows = append(workflows, workflow["id"].(uuid.UUID))
		}
		workflowTools, err := s.appConfigService.GetToolsByWorkflowIDs(ctx, workflows)
		if err != nil {
			return nil, err
		}
		tools = append(tools, workflowTools...)
	}

	// 10. 根据LLM是否支持tool_call决定使用不同的Agent
	agentCfg := &agenteneity.AgentConfig{
		UserID:               accountID,
		InvokeFrom:           consts.InvokeFromDebugger,
		PresetPrompt:         draftAppConfig.PresetPrompt,
		EnableLongTermMemory: draftAppConfig.LongTermMemory["enabled"].(bool),
		Tools:                tools,
	}
	if err := util.ConvertViaJSON(&agentCfg.ReviewConfig, draftAppConfig.ReviewConfig); err != nil {
		return nil, err
	}
	agentIns := agent.NewFunctionCallAgent(s.llm, agentCfg, s.agentManager)
	for _, f := range s.llm.GetFeatures() {
		if f == llmentity.FeatureToolCall {
			agentIns = agent.NewFunctionCallAgent(s.llm, agentCfg, s.agentManager)
			break
		}
	}
	// 创建响应流通道
	responseStream := make(chan string, 100)

	// 启动异步处理
	go s.processDebugChat(ctx, appID, agentIns, history, debugConversation, message, chatReq, accountID, responseStream)

	return responseStream, nil
}

// processDebugChat 处理调试对话的异步逻辑
func (s *AppService) processDebugChat(
	ctx context.Context,
	appID uuid.UUID,
	agentIns agent.BaseAgent,
	history []*schema.Message,
	debugConversation *entity.Conversation,
	message *entity.Message,
	chatReq req.DebugChatReq,
	accountID uuid.UUID,
	responseStream chan<- string,
) {
	defer close(responseStream)

	// 实现完整的Agent流式处理逻辑
	agentState := agenteneity.AgentState{
		TaskID:         uuid.New(),
		Messages:       history,
		History:        history,
		LongTermMemory: debugConversation.Summary,
		IterationCount: 0,
	}

	// 添加当前用户消息
	if len(chatReq.Query) > 0 {
		userMsg := schema.UserMessage(chatReq.Query)
		agentState.Messages = append(agentState.Messages, userMsg)
	}

	// 获取Agent流式输出
	thoughtChan, err := agentIns.Stream(ctx, agentState)
	if err != nil {
		select {
		case responseStream <- fmt.Sprintf("event: error\ndata: %s\n\n", err.Error()):
		case <-ctx.Done():
		}
		return
	}

	// 存储agent思考过程
	agentThoughts := make(map[string]*agenteneity.AgentThought)

	// 处理流式输出
	for agentThought := range thoughtChan {
		eventID := agentThought.ID.String()

		// 除了ping事件，其他事件全部记录
		if agentThought.Event != agenteneity.EventPing {
			// 单独处理agent_message事件，因为该事件为数据叠加
			if agentThought.Event == agenteneity.EventAgentMessage {
				if existing, exists := agentThoughts[eventID]; exists {
					// 叠加智能体消息事件
					existing.Thought = existing.Thought + agentThought.Thought
					existing.Answer = existing.Answer + agentThought.Answer
					existing.Latency = agentThought.Latency
					agentThoughts[eventID] = existing
				} else {
					// 初始化智能体消息事件
					agentThoughts[eventID] = agentThought
				}
			} else {
				// 处理其他类型事件的消息
				agentThoughts[eventID] = agentThought
			}
		}

		// 构建响应数据
		data := map[string]any{
			"id":              eventID,
			"conversation_id": debugConversation.ID.String(),
			"message_id":      message.ID.String(),
			"task_id":         agentState.TaskID.String(),
			"event":           string(agentThought.Event),
			"thought":         agentThought.Thought,
			"observation":     agentThought.Observation,
			"tool":            agentThought.Tool,
			"tool_input":      agentThought.ToolInput,
			"answer":          agentThought.Answer,
			"latency":         agentThought.Latency,
		}

		jsonData, _ := sonic.Marshal(data)
		eventStr := fmt.Sprintf("event: %s\ndata: %s\n\n", agentThought.Event, string(jsonData))

		select {
		case responseStream <- eventStr:
		case <-ctx.Done():
			return
		}
	}

	// 将agent思考过程转换为切片
	agentThoughtsList := make([]agenteneity.AgentThought, 0, len(agentThoughts))
	for _, thought := range agentThoughts {
		agentThoughtsList = append(agentThoughtsList, *thought)
	}

	// 更新消息状态和结果
	finalAnswer := ""
	if len(agentThoughtsList) > 0 {
		finalAnswer = agentThoughtsList[len(agentThoughtsList)-1].Answer
	}

	// 更新消息记录
	err = s.conversationService.UpdateMessage(ctx, accountID, message.ID, &req.UpdateMessageReq{
		Answer: finalAnswer,
		Status: consts.MessageStatusNormal.String(),
	})
	if err != nil {
		logs.Error("Failed to update message: %v", err)
	}

	// 将消息以及推理过程添加到数据库
	err = s.conversationService.SaveAgentThoughts(ctx, accountID, appID, debugConversation.ID, message.ID, agentThoughtsList)
	if err != nil {
		// 记录错误但不中断流程
		logs.Errorf("Failed to save agent thoughts: %v", err)
	}
}

// StopDebugChat 根据传递的应用id+任务id+账号，停止某个应用的调试会话，中断流式事件
func (s *AppService) StopDebugChat(ctx context.Context, appID, taskID uuid.UUID, accountID uuid.UUID) error {
	// 1. 获取应用信息并校验权限
	_, err := s.GetApp(ctx, appID, accountID)
	if err != nil {
		return err
	}

	// 2. 调用智能体队列管理器停止特定任务

	return s.agentManager.StopTask(ctx, taskID, accountID, consts.InvokeFromDebugger)
}

// GetDebugConversationMessagesWithPage 根据传递的应用id+请求数据，获取调试会话消息列表分页数据
func (s *AppService) GetDebugConversationMessagesWithPage(ctx context.Context, appID uuid.UUID, getReq req.GetDebugConversationMessagesWithPageReq, accountID uuid.UUID) ([]resp.DebugConversationMessageResp, resp.Paginator, error) {
	// 1. 获取应用信息并校验权限
	app, err := s.GetApp(ctx, appID, accountID)
	if err != nil {
		return nil, resp.Paginator{}, err
	}

	// 2. 获取应用的调试会话
	if app.DebugConversationID == uuid.Nil {
		return nil, resp.Paginator{
			CurrentPage: getReq.CurrentPage,
			PageSize:    getReq.PageSize,
		}, nil
	}

	// 3. 执行分页并查询数据
	messages, paginator, err := s.conversationService.GetConversationMessagesWithPage(ctx, app.DebugConversationID, req.GetConversationMessagesWithPageReq{
		CurrentPage: getReq.CurrentPage,
		PageSize:    getReq.PageSize,
		Ctime:       getReq.Ctime,
	})
	if err != nil {
		return nil, resp.Paginator{}, err
	}

	res := make([]resp.DebugConversationMessageResp, 0, len(messages))
	for _, message := range messages {
		dbAgentThoughts, err := s.conversationService.GetConversationAgentThoughts(ctx, message.ConversationID)
		if err != nil {
			continue
		}
		agentThoughts := make([]resp.AgentThought, 0, len(dbAgentThoughts))
		for _, at := range dbAgentThoughts {
			agentThoughts = append(agentThoughts, resp.AgentThought{
				ID:              at.ID,
				MessageID:       message.ID,
				Event:           at.Event,
				Thought:         at.Thought,
				Observation:     at.Observation,
				Tool:            at.Tool,
				ToolInput:       at.ToolInput,
				Answer:          at.Answer,
				TotalTokenCount: at.TotalTokenCount,
				TotalPrice:      at.TotalPrice,
				Latency:         at.Latency,
				Ctime:           at.Ctime,
			})
		}
		res = append(res, resp.DebugConversationMessageResp{
			ID:              message.ID,
			ConversationID:  message.ConversationID,
			Query:           message.Query,
			ImageUrls:       message.ImageUrls,
			Answer:          message.Answer,
			Latency:         message.Latency,
			TotalTokenCount: message.TotalTokenCount,
			AgentThoughts:   agentThoughts,
			Ctime:           message.Ctime,
		})
	}

	return res, paginator, nil
}

// GetPublishedConfig 根据传递的应用id+账号，获取应用的发布配置
func (s *AppService) GetPublishedConfig(ctx context.Context, appID uuid.UUID, accountID uuid.UUID) (map[string]any, error) {
	// 1. 获取应用信息并校验权限
	app, err := s.GetApp(ctx, appID, accountID)
	if err != nil {
		return nil, err
	}

	// 2. 构建发布配置并返回
	token, err := s.generateDefaultToken(ctx, appID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"web_app": map[string]any{
			"token":  token,
			"status": app.Status,
		},
	}, nil
}

// RegenerateWebAppToken 根据传递的应用id+账号，重新生成WebApp凭证标识
func (s *AppService) RegenerateWebAppToken(ctx context.Context, appID uuid.UUID, accountID uuid.UUID) (string, error) {
	// 1. 获取应用信息并校验权限
	app, err := s.RawGetApp(ctx, appID, accountID)
	if err != nil {
		return "", err
	}

	// 2. 判断应用是否已发布
	if app.Status != consts.AppStatusPublished {
		return "", errno.ErrValidate.AppendBizMessage(errors.New("应用未发布，无法生成WebApp凭证标识"))
	}

	// 3. 重新生成token并更新数据
	token := s.generateRandomString(16)
	err = s.repo.UpdateApp(ctx, appID, map[string]any{
		"token": token,
	})
	if err != nil {
		return "", err
	}

	return token, nil
}

// validateDraftAppConfig 校验传递的应用草稿配置信息，返回校验后的数据
func (s *AppService) validateDraftAppConfig(draftAppConfig map[string]any, accountID uuid.UUID) (map[string]any, error) {
	// 1. 校验上传的草稿配置中对应的字段，至少拥有一个可以更新的配置
	acceptableFields := []string{
		"model_config", "dialog_round", "preset_prompt",
		"tools", "workflows", "datasets", "retrieval_config",
		"long_term_memory", "opening_statement", "opening_questions",
		"speech_to_text", "text_to_speech", "suggested_after_answer", "review_config",
	}

	// 2. 判断传递的草稿配置是否在可接受字段内
	if draftAppConfig == nil || len(draftAppConfig) == 0 {
		return nil, errno.ErrValidate.AppendBizMessage(errors.New("草稿配置字段出错，请核实后重试"))
	}

	for key := range draftAppConfig {
		found := false
		for _, field := range acceptableFields {
			if key == field {
				found = true
				break
			}
		}
		if !found {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("草稿配置字段出错，请核实后重试"))
		}
	}

	// 3. 校验model_config字段
	if modelConfig, exists := draftAppConfig["model_config"]; exists {
		if modelConfigMap, ok := modelConfig.(map[string]any); ok {
			// 3.1 判断model_config键信息是否正确
			requiredKeys := []string{"provider", "model", "parameters"}
			if len(modelConfigMap) != 3 {
				return nil, errno.ErrValidate.AppendBizMessage(errors.New("模型键配置格式错误，请核实后重试"))
			}

			for _, key := range requiredKeys {
				if _, exists := modelConfigMap[key]; !exists {
					return nil, errno.ErrValidate.AppendBizMessage(errors.New("模型键配置格式错误，请核实后重试"))
				}
			}

			// 3.2 判断模型提供者信息是否正确
			if provider, ok := modelConfigMap["provider"].(string); !ok || provider == "" {
				return nil, errno.ErrValidate.AppendBizMessage(errors.New("模型服务提供商类型必须为字符串"))
			}
			provider, err := s.llmService.GetProvider(modelConfigMap["provider"].(string))
			if err != nil {
				return nil, err
			}
			if provider == nil {
				return nil, errno.ErrNotFound.AppendBizMessage(errors.New("该模型服务提供商不存在，请核实后重试"))
			}

			// 3.3 判断模型信息是否正确
			if model, ok := modelConfigMap["model"].(string); !ok || model == "" {
				return nil, errno.ErrValidate.AppendBizMessage(errors.New("模型名字必须是字符串"))
			}
			model, err := provider.GetModelEntity(modelConfigMap["model"].(string))
			if err != nil {
				return nil, err
			}
			if model == nil {
				return nil, errno.ErrNotFound.AppendBizMessage(errors.New("该服务提供商下不存在该模型，请核实后重试"))
			}

			var parameters map[string]any
			// 3.4 判断传递的parameters是否正确，如果不正确则设置默认值，并剔除多余字段，补全未传递的字段
			for _, parameter := range model.Parameters {
				// 3.5 从model_config中获取参数值，如果不存在则设置为默认值
				parameterVal := modelConfigMap["parameters"].(map[string]any)
				parameterValue := s.getParamWithDefault(parameterVal, parameter.Name, parameter.Default)

				// 3.6 判断参数是否必填
				if parameter.Required {
					// 3.7 参数必填，则值不允许为 nil，如果为 nil 则设置默认值
					if parameterValue == nil {
						parameterValue = parameter.Default
					} else {
						// 3.8 值非空则校验数据类型是否正确，不正确则设置默认值
						if util.GetValueType(parameterValue) != string(parameter.Type) {
							parameterValue = parameter.Default
						}
					}
				} else {
					// 3.9 参数非必填，数据非空的情况下需要校验
					if parameterValue != nil {
						if util.GetValueType(parameterValue) != string(parameter.Type) {
							parameterValue = parameter.Default
						}
					}
				}

				// 3.10 判断参数是否存在options，如果存在则数值必须在options中选择
				if parameter.Options != nil && !util.Contains(parameter.Options, parameterValue) {
					parameterValue = parameter.Default
				}

				// 3.11 参数类型为int/float，如果存在min/max时候需要校验
				if !util.Contains([]llmentity.ModelParameterType{llmentity.ParameterTypeFloat,
					llmentity.ParameterTypeInt}, parameter.Type) && parameterValue != nil {
					// 3.12 校验数值的min/max
					if (parameter.Min != nil && util.LessThan(parameterValue, parameter.Min)) ||
						(parameter.Max != nil && util.GreaterThan(parameterValue, parameter.Max)) {
						parameterValue = parameter.Default
					}
				}

				parameters[parameter.Name] = parameterValue
			}

			// 3.13 覆盖Agent配置中的模型配置
			modelConfigMap["parameters"] = parameters
			draftAppConfig["model_config"] = modelConfigMap

		} else {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("模型配置格式错误，请核实后重试"))
		}
	}

	// 4. 校验dialog_round上下文轮数
	if dialogRound, exists := draftAppConfig["dialog_round"]; exists {
		if round, ok := dialogRound.(int); ok {
			if round < 0 || round > 100 {
				return nil, errno.ErrValidate.AppendBizMessage(errors.New("携带上下文轮数范围为0-100"))
			}
		} else {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("携带上下文轮数必须为整数"))
		}
	}

	// 5. 校验preset_prompt
	if presetPrompt, exists := draftAppConfig["preset_prompt"]; exists {
		if prompt, ok := presetPrompt.(string); ok {
			if len(prompt) > 2000 {
				return nil, errno.ErrValidate.AppendBizMessage(errors.New("人设与回复逻辑必须是字符串，长度在0-2000个字符"))
			}
		} else {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("人设与回复逻辑必须是字符串"))
		}
	}

	// 6.校验tools工具
	if tools, exists := draftAppConfig["tools"]; exists {
		toolsSlice, ok := tools.([]map[string]any)
		if !ok {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("工具列表必须是列表型数据"))
		}

		// 6.1 tools的长度不能超过5
		if len(toolsSlice) > 5 {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("Agent绑定的工具数不能超过5"))
		}

		validateTools := make([]map[string]any, 0)
		toolKeys := make(map[string]bool)

		// 6.2 循环校验工具里的每一个参数
		for _, tool := range toolsSlice {
			if tool == nil {
				return nil, errno.ErrValidate.AppendBizMessage(errors.New("绑定插件工具参数出错"))
			}

			// 6.3 校验工具的参数是不是type、provider_id、tool_id、params
			requiredKeys := map[string]bool{
				"type":        true,
				"provider_id": true,
				"tool_id":     true,
				"params":      true,
			}

			for k := range tool {
				delete(requiredKeys, k)
			}
			if len(requiredKeys) > 0 {
				return nil, errno.ErrValidate.AppendBizMessage(errors.New("绑定插件工具参数出错"))
			}

			// 6.4 校验type类型是否为builtin_tool以及api_tool
			toolType, ok := tool["type"].(string)
			if !ok || (toolType != "builtin_tool" && toolType != "api_tool") {
				return nil, errno.ErrValidate.AppendBizMessage(errors.New("绑定插件工具参数出错"))
			}

			// 6.5 校验provider_id和tool_id
			providerID, ok1 := tool["provider_id"].(string)
			toolID, ok2 := tool["tool_id"].(string)
			if !ok1 || !ok2 || providerID == "" || toolID == "" {
				return nil, errno.ErrValidate.AppendBizMessage(errors.New("插件提供者或者插件标识参数出错"))
			}

			// 6.6 校验params参数，类型为字典
			if _, ok := tool["params"].(map[string]any); !ok {
				return nil, errno.ErrValidate.AppendBizMessage(errors.New("插件自定义参数格式错误"))
			}

			// 6.7 校验对应的工具是否存在
			if toolType == "builtin_tool" {
				builtinTool, err := s.builtinProvider.GetTool(toolID)
				if err != nil || builtinTool == nil {
					continue
				}
			} else {
				apiTool, err := s.repo.GetApiTool(context.Background(), providerID, toolID, accountID)
				if err != nil || apiTool == nil {
					continue
				}
			}

			// 检查工具是否重复
			key := providerID + "_" + toolID
			if toolKeys[key] {
				return nil, errno.ErrValidate.AppendBizMessage(errors.New("绑定插件存在重复"))
			}
			toolKeys[key] = true

			validateTools = append(validateTools, tool)
		}

		// 6.8 重新赋值工具
		draftAppConfig["tools"] = validateTools
	}

	// 7. 校验workflow
	if workflows, exists := draftAppConfig["workflows"]; exists {
		workflowsSlice, ok := workflows.([]any)
		if !ok {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("绑定工作流列表参数格式错误"))
		}

		// 7.2 判断关联的工作流列表是否超过5个
		if len(workflowsSlice) > 5 {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("Agent绑定的工作流数量不能超过5个"))
		}

		// 7.3 校验每个工作流ID是否为UUID
		workflowIDs := make([]uuid.UUID, 0, len(workflowsSlice))
		for _, item := range workflowsSlice {
			workflowID, ok := item.(uuid.UUID)
			if !ok {
				return nil, errno.ErrValidate.AppendBizMessage(errors.New("工作流参数必须是UUID"))
			}
			workflowIDs = append(workflowIDs, workflowID)
		}

		// 7.4 判断是否重复关联了工作流
		if len(workflowIDs) != len(util.UniqueUUID(workflowIDs)) {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("绑定工作流存在重复"))
		}

		// 7.5 校验关联工作流的权限
		workflowRecords, err := s.repo.GetWorkflows(context.Background(), workflowIDs, accountID, consts.WorkflowStatusPublished)
		if err != nil {
			return nil, errno.ErrNotFound.AppendBizMessage(errors.New("查询工作流失败"))
		}

		validWorkflows := make([]uuid.UUID, 0)
		validWorkflowSet := make(map[uuid.UUID]bool)
		for _, w := range workflowRecords {
			validWorkflowSet[w.ID] = true
		}

		for _, id := range workflowIDs {
			if validWorkflowSet[id] {
				validWorkflows = append(validWorkflows, id)
			}
		}

		draftAppConfig["workflows"] = validWorkflows
	}

	// 8. 校验datasets知识库列表
	if datasets, exists := draftAppConfig["datasets"]; exists {
		datasetsSlice, ok := datasets.([]any)
		if !ok {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("绑定知识库列表参数格式错误"))
		}

		if len(datasetsSlice) > 5 {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("Agent绑定的知识库数量不能超过5个"))
		}

		datasetIDs := make([]uuid.UUID, 0, len(datasetsSlice))
		for _, item := range datasetsSlice {
			datasetID, ok := item.(uuid.UUID)
			if !ok {
				return nil, errno.ErrValidate.AppendBizMessage(errors.New("知识库列表参数必须是UUID"))
			}
			datasetIDs = append(datasetIDs, datasetID)
		}

		if len(datasetIDs) != len(util.UniqueUUID(datasetIDs)) {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("绑定知识库存在重复"))
		}

		datasetRecords, err := s.repo.GetDatasets(context.Background(), datasetIDs, accountID)
		if err != nil {
			return nil, errno.ErrNotFound.AppendBizMessage(errors.New("查询知识库失败"))
		}

		validDatasets := make([]uuid.UUID, 0)
		validDatasetSet := make(map[uuid.UUID]bool)
		for _, d := range datasetRecords {
			validDatasetSet[d.ID] = true
		}

		for _, id := range datasetIDs {
			if validDatasetSet[id] {
				validDatasets = append(validDatasets, id)
			}
		}

		draftAppConfig["datasets"] = validDatasets
	}

	// 9. 校验retrieval_config检索配置
	if retrievalConfig, exists := draftAppConfig["retrieval_config"]; exists {
		rc, ok := retrievalConfig.(map[string]any)
		if !ok || len(rc) == 0 {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("检索配置格式错误"))
		}

		requiredKeys := map[string]bool{
			"retrieval_strategy": true,
			"k":                  true,
			"score":              true,
		}

		for k := range rc {
			delete(requiredKeys, k)
		}
		if len(requiredKeys) > 0 {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("检索配置格式错误"))
		}

		strategy, ok := rc["retrieval_strategy"].(string)
		if !ok || (strategy != "semantic" && strategy != "full_text" && strategy != "hybrid") {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("检测策略格式错误"))
		}

		k, ok := rc["k"].(float64)
		if !ok || k < 0 || k > 10 {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("最大召回数量范围为0-10"))
		}

		score, ok := rc["score"].(float64)
		if !ok || score < 0 || score > 1 {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("最小匹配范围为0-1"))
		}
	}

	// 10. 校验long_term_memory长期记忆配置
	if longTermMemory, exists := draftAppConfig["long_term_memory"]; exists {
		ltm, ok := longTermMemory.(map[string]any)
		if !ok || len(ltm) == 0 {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("长期记忆设置格式错误"))
		}

		if len(ltm) != 1 {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("长期记忆设置格式错误"))
		}

		enable, ok := ltm["enable"]
		if !ok {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("长期记忆设置格式错误"))
		}
		if _, ok := enable.(bool); !ok {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("长期记忆设置格式错误"))
		}
	}

	// 11. 校验opening_statement对话开场白
	if openingStatement, exists := draftAppConfig["opening_statement"]; exists {
		openS, ok := openingStatement.(string)
		if !ok || len(openS) > 2000 {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("对话开场白的长度范围是0-2000"))
		}
	}

	// 12. 校验opening_questions开场建议问题列表
	if openingQuestions, exists := draftAppConfig["opening_questions"]; exists {
		oq, ok := openingQuestions.([]any)
		if !ok || len(oq) > 3 {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("开场建议问题不能超过3个"))
		}

		for _, question := range oq {
			if _, ok := question.(string); !ok {
				return nil, errno.ErrValidate.AppendBizMessage(errors.New("开场建议问题必须是字符串"))
			}
		}
	}

	// 13. 校验speech_to_text语音转文本
	if speechToText, exists := draftAppConfig["speech_to_text"]; exists {
		// 13.1 检查是否是 map 且非空
		stt, ok := speechToText.(map[string]interface{})
		if !ok || len(stt) == 0 {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("语音转文本设置格式错误"))
		}

		// 13.2 检查是否只有 "enable" 键
		if len(stt) != 1 {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("语音转文本设置格式错误"))
		}

		// 检查 enable 是否存在且是 bool 类型
		enable, exists := stt["enable"]
		if !exists {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("语音转文本设置格式错误"))
		}

		if _, ok := enable.(bool); !ok {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("语音转文本设置格式错误"))
		}
	}

	// 14. 校验text_to_speech文本转语音设置
	if textToSpeech, exists := draftAppConfig["text_to_speech"]; exists {
		tts, ok := textToSpeech.(map[string]any)
		if !ok {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("文本转语音设置格式错误"))
		}

		requiredKeys := map[string]bool{
			"enable":    true,
			"voice":     true,
			"auto_play": true,
		}

		for k := range tts {
			delete(requiredKeys, k)
		}
		if len(requiredKeys) > 0 {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("文本转语音设置格式错误"))
		}

		enable, ok := tts["enable"]
		if !ok {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("文本转语音设置格式错误"))
		}
		if _, ok := enable.(bool); !ok {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("文本转语音设置格式错误"))
		}

		voice, ok := tts["voice"].(string)
		if !ok || !util.Contains(consts.AllowedAudioVoices, voice) {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("文本转语音设置格式错误"))
		}

		autoPlay, ok := tts["auto_play"]
		if !ok {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("文本转语音设置格式错误"))
		}
		if _, ok := autoPlay.(bool); !ok {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("文本转语音设置格式错误"))
		}
	}

	// 15. 校验回答后生成建议问题
	if suggestedAfterAnswer, exists := draftAppConfig["suggested_after_answer"]; exists {
		saa, ok := suggestedAfterAnswer.(map[string]any)
		if !ok || len(saa) == 0 {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("回答后建议问题设置格式错误"))
		}

		if len(saa) != 1 {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("回答后建议问题设置格式错误"))
		}

		enable, ok := saa["enable"]
		if !ok {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("回答后建议问题设置格式错误"))
		}
		if _, ok := enable.(bool); !ok {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("回答后建议问题设置格式错误"))
		}
	}

	// 16. 校验review_config审核配置
	if reviewConfig, exists := draftAppConfig["review_config"]; exists {
		rc, ok := reviewConfig.(map[string]any)
		if !ok || len(rc) == 0 {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("审核配置格式错误"))
		}

		requiredKeys := map[string]bool{
			"enable":         true,
			"keywords":       true,
			"inputs_config":  true,
			"outputs_config": true,
		}

		for k := range rc {
			delete(requiredKeys, k)
		}
		if len(requiredKeys) > 0 {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("审核配置格式错误"))
		}

		enable, ok := rc["enable"].(bool)
		if !ok {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("review.enable格式错误"))
		}

		keywords, ok := rc["keywords"].([]any)
		if !ok || (enable && len(keywords) == 0) || len(keywords) > 100 {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("review.keywords非空且不能超过100个关键词"))
		}

		for _, keyword := range keywords {
			if _, ok := keyword.(string); !ok {
				return nil, errno.ErrValidate.AppendBizMessage(errors.New("review.keywords敏感词必须是字符串"))
			}
		}

		inputsConfig, ok := rc["inputs_config"].(map[string]any)
		if !ok || len(inputsConfig) == 0 {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("review.inputs_config必须是一个字典"))
		}

		if len(inputsConfig) != 2 {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("review.inputs_config必须是一个字典"))
		}

		inputEnable, ok := inputsConfig["enable"].(bool)
		if !ok {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("review.inputs_config必须是一个字典"))
		}

		presetResponse, ok := inputsConfig["preset_response"].(string)
		if !ok {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("review.inputs_config必须是一个字典"))
		}

		outputsConfig, ok := rc["outputs_config"].(map[string]any)
		if !ok || len(outputsConfig) == 0 {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("review.outputs_config格式错误"))
		}

		if len(outputsConfig) != 1 {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("review.outputs_config格式错误"))
		}

		outputEnable, ok := outputsConfig["enable"].(bool)
		if !ok {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("review.outputs_config格式错误"))
		}

		if enable {
			if !inputEnable && !outputEnable {
				return nil, errno.ErrValidate.AppendBizMessage(errors.New("输入审核和输出审核至少需要开启一项"))
			}

			if inputEnable && strings.TrimSpace(presetResponse) == "" {
				return nil, errno.ErrValidate.AppendBizMessage(errors.New("输入审核预设响应不能为空"))
			}
		}
	}

	return draftAppConfig, nil
}

func (s *AppService) generateDefaultToken(ctx context.Context, appID uuid.UUID) (string, error) {
	app, err := s.repo.GetAppByID(ctx, appID)
	if err != nil {
		return "", err
	}
	var token string
	if app.Status != consts.AppStatusPublished {
		if app.Token != "" {
			token = ""
		}
	}

	token = s.generateRandomString(16)

	return token, s.repo.UpdateApp(ctx, appID, map[string]any{
		"token": token,
	})
}

func (s *AppService) generateRandomString(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	rand.New(rand.NewSource(time.Now().UnixNano()))
	var sb strings.Builder
	for i := 0; i < length; i++ {
		sb.WriteByte(chars[rand.Intn(len(chars))])
	}

	return sb.String()
}

func (s *AppService) getParamWithDefault(params map[string]any, name string, defaultValue any) any {
	if val, ok := params[name]; ok {
		return val
	}
	return defaultValue
}

// getUserSessions 获取用户的会话map，如果不存在则创建
func (s *AppService) getUserSessions(userID string) map[string]context.CancelFunc {
	if sessions, exists := s.activeSessions.Load(userID); exists {
		return sessions.(map[string]context.CancelFunc)
	}

	// 创建新的用户会话map
	userSessions := make(map[string]context.CancelFunc)
	s.activeSessions.Store(userID, userSessions)
	return userSessions
}

// addUserSession 添加用户会话
func (s *AppService) addUserSession(userID, taskID string, cancel context.CancelFunc) {
	if sessions, exists := s.activeSessions.Load(userID); exists {
		userSessions := sessions.(map[string]context.CancelFunc)
		userSessions[taskID] = cancel
	} else {
		userSessions := make(map[string]context.CancelFunc)
		userSessions[taskID] = cancel
		s.activeSessions.Store(userID, userSessions)
	}
}

// removeUserSession 移除用户会话
func (s *AppService) removeUserSession(userID, taskID string) {
	if sessions, exists := s.activeSessions.Load(userID); exists {
		userSessions := sessions.(map[string]context.CancelFunc)
		delete(userSessions, taskID)

		// 如果用户没有活跃会话了，清理用户记录
		if len(userSessions) == 0 {
			s.activeSessions.Delete(userID)
		}
	}
}

// cancelUserSession 取消用户的特定会话
func (s *AppService) cancelUserSession(userID, taskID string) bool {
	if sessions, exists := s.activeSessions.Load(userID); exists {
		userSessions := sessions.(map[string]context.CancelFunc)
		if cancel, exists := userSessions[taskID]; exists {
			cancel()
			delete(userSessions, taskID)

			// 如果用户没有活跃会话了，清理用户记录
			if len(userSessions) == 0 {
				s.activeSessions.Delete(userID)
			}
			return true
		}
	}
	return false
}
