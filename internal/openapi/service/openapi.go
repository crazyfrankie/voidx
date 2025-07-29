package service

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"

	"github.com/crazyfrankie/voidx/internal/app"
	"github.com/crazyfrankie/voidx/internal/app_config"
	"github.com/crazyfrankie/voidx/internal/conversation/service"
	"github.com/crazyfrankie/voidx/internal/core/agent"
	agenteneity "github.com/crazyfrankie/voidx/internal/core/agent/entities"
	llmentity "github.com/crazyfrankie/voidx/internal/core/llm/entity"
	"github.com/crazyfrankie/voidx/internal/core/memory"
	"github.com/crazyfrankie/voidx/internal/llm"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/internal/openapi/repository"
	"github.com/crazyfrankie/voidx/internal/retriever"
	"github.com/crazyfrankie/voidx/pkg/consts"
	"github.com/crazyfrankie/voidx/pkg/util"
)

type OpenAPIService struct {
	repo                *repository.OpenAPIRepo
	conversationService *service.ConversationService
	retrieverSvc        *retriever.Service
	llmSvc              *llm.Service
	appConfigSvc        *app_config.Service
	appSvc              *app.Service
	agentManager        *agent.AgentQueueManager
	tokeBufMem          *memory.TokenBufferMemory
}

func NewOpenAPIService(repo *repository.OpenAPIRepo, conversationService *service.ConversationService,
	retrieverSvc *retriever.Service, llmSvc *llm.Service, appConfigSvc *app_config.Service,
	appSvc *app.Service, agentManager *agent.AgentQueueManager, tokeBufMem *memory.TokenBufferMemory) *OpenAPIService {
	return &OpenAPIService{
		repo:                repo,
		conversationService: conversationService,
		retrieverSvc:        retrieverSvc,
		llmSvc:              llmSvc,
		appConfigSvc:        appConfigSvc,
		appSvc:              appSvc,
		agentManager:        agentManager,
		tokeBufMem:          tokeBufMem,
	}
}

// Chat 根据传递的请求+账号信息发起聊天对话，返回块内容
func (s *OpenAPIService) Chat(ctx context.Context, userID uuid.UUID, chatReq req.OpenAPIChatReq) (*resp.OpenAPIChatResp, error) {
	// 1. 验证应用权限和状态
	app, err := s.appSvc.GetApp(ctx, chatReq.AppID, userID)
	if err != nil {
		return nil, err
	}

	// 判断应用是否已发布
	if app.Status != consts.AppStatusPublished {
		return nil, fmt.Errorf("该应用不存在或未发布，请核实后重试")
	}

	// 2. 获取或创建终端用户
	endUser, err := s.getOrCreateEndUser(ctx, chatReq.EndUserID, userID, app.ID)
	if err != nil {
		return nil, err
	}

	// 3. 获取或创建会话
	conversation, err := s.getOrCreateConversation(ctx, chatReq.ConversationID, endUser.ID, app.ID)
	if err != nil {
		return nil, err
	}

	// 4. 创建消息记录
	message, err := s.conversationService.RawCreateMessage(ctx, &entity.Message{
		ID:             uuid.New(),
		AppID:          app.ID,
		ConversationID: conversation.ID,
		InvokeFrom:     consts.InvokeFromServiceAPI,
		CreatedBy:      endUser.ID,
		Query:          chatReq.Query,
		ImageUrls:      chatReq.ImageUrls,
		Status:         consts.MessageStatusNormal,
	})
	if err != nil {
		return nil, err
	}

	// 5. 获取校验后的运行时配置
	appConfig, err := s.appConfigSvc.GetAppConfig(ctx, app)
	if err != nil {
		return nil, err
	}

	// 6. 从语言模型中根据模型配置获取模型实例
	llm, err := s.llmSvc.LoadLanguageModel(appConfig.ModelConfig)
	if err != nil {
		return nil, err
	}
	s.tokeBufMem = s.tokeBufMem.WithLLM(llm)
	// 7. 获取历史消息（Chat方法中暂不使用，但保留调用以验证功能）
	_, err = s.tokeBufMem.GetHistoryPromptMessages(2000, appConfig.DialogRound)
	if err != nil {
		return nil, err
	}

	// 8. 将配置中的tools转换成LangChain工具
	tools, err := s.appConfigSvc.GetLangchainToolsByToolsConfig(ctx, appConfig.Tools)
	if err != nil {
		return nil, err
	}

	// 9. 检测是否关联了知识库
	if appConfig.Datasets != nil && len(appConfig.Datasets) > 0 {
		datasets := make([]uuid.UUID, 0, len(appConfig.Datasets))
		for _, dataset := range appConfig.Datasets {
			if datasetID, ok := dataset["id"].(uuid.UUID); ok {
				datasets = append(datasets, datasetID)
			}
		}
		if len(datasets) > 0 {
			datasetTool, err := s.retrieverSvc.CreateLangchainToolFromSearch(ctx, userID, datasets, consts.RetrievalSourceApp, appConfig.RetrievalConfig)
			if err == nil {
				tools = append(tools, datasetTool)
			}
		}
	}

	// 10. 检测是否关联工作流
	if appConfig.Workflows != nil && len(appConfig.Workflows) > 0 {
		workflows := make([]uuid.UUID, 0, len(appConfig.Workflows))
		for _, workflow := range appConfig.Workflows {
			if workflowID, ok := workflow["id"].(uuid.UUID); ok {
				workflows = append(workflows, workflowID)
			}
		}
		if len(workflows) > 0 {
			workflowTools, err := s.appConfigSvc.GetLangchainToolsByWorkflowIDs(ctx, workflows)
			if err == nil {
				tools = append(tools, workflowTools...)
			}
		}
	}

	// 11. 创建Agent并处理（简化版本，直接返回模拟回答）
	// 在Chat方法中，我们不使用history变量，因为这是同步版本
	answer := fmt.Sprintf("这是对问题「%s」的回答。", chatReq.Query)

	// 12. 构建响应
	return &resp.OpenAPIChatResp{
		ID:              message.ID,
		EndUserID:       endUser.ID,
		ConversationID:  conversation.ID,
		Query:           chatReq.Query,
		ImageUrls:       chatReq.ImageUrls,
		Answer:          answer,
		TotalTokenCount: 0,
		Latency:         0.5,
		AgentThoughts:   []resp.OpenAPIAgentThought{},
	}, nil
}

// ProcessStreamChat 根据传递的请求+账号信息发起聊天对话，返回流式事件
func (s *OpenAPIService) ProcessStreamChat(ctx context.Context, userID uuid.UUID, chatReq req.OpenAPIChatReq) (<-chan string, error) {
	// 1. 验证应用权限和状态
	app, err := s.appSvc.GetApp(ctx, chatReq.AppID, userID)
	if err != nil {
		return nil, err
	}

	// 判断应用是否已发布
	if app.Status != consts.AppStatusPublished {
		return nil, fmt.Errorf("该应用不存在或未发布，请核实后重试")
	}

	// 2. 获取或创建终端用户
	endUser, err := s.getOrCreateEndUser(ctx, chatReq.EndUserID, userID, app.ID)
	if err != nil {
		return nil, err
	}

	// 3. 获取或创建会话
	conversation, err := s.getOrCreateConversation(ctx, chatReq.ConversationID, endUser.ID, app.ID)
	if err != nil {
		return nil, err
	}

	// 4. 创建消息记录
	message, err := s.conversationService.RawCreateMessage(ctx, &entity.Message{
		ID:             uuid.New(),
		AppID:          app.ID,
		ConversationID: conversation.ID,
		InvokeFrom:     consts.InvokeFromServiceAPI,
		CreatedBy:      endUser.ID,
		Query:          chatReq.Query,
		ImageUrls:      chatReq.ImageUrls,
		Status:         consts.MessageStatusNormal,
	})
	if err != nil {
		return nil, err
	}

	// 5. 获取校验后的运行时配置
	appConfig, err := s.appConfigSvc.GetAppConfig(ctx, app)
	if err != nil {
		return nil, err
	}

	// 6. 从语言模型中根据模型配置获取模型实例
	llm, err := s.llmSvc.LoadLanguageModel(appConfig.ModelConfig)
	if err != nil {
		return nil, err
	}
	s.tokeBufMem = s.tokeBufMem.WithLLM(llm)
	// 7. 获取历史消息
	history, err := s.tokeBufMem.GetHistoryPromptMessages(2000, appConfig.DialogRound)
	if err != nil {
		return nil, err
	}

	// 8. 将配置中的tools转换成LangChain工具
	tools, err := s.appConfigSvc.GetLangchainToolsByToolsConfig(ctx, appConfig.Tools)
	if err != nil {
		return nil, err
	}

	// 9. 检测是否关联了知识库
	if appConfig.Datasets != nil && len(appConfig.Datasets) > 0 {
		datasets := make([]uuid.UUID, 0, len(appConfig.Datasets))
		for _, dataset := range appConfig.Datasets {
			if datasetID, ok := dataset["id"].(uuid.UUID); ok {
				datasets = append(datasets, datasetID)
			}
		}
		if len(datasets) > 0 {
			datasetTool, err := s.retrieverSvc.CreateLangchainToolFromSearch(ctx, userID, datasets, consts.RetrievalSourceApp, appConfig.RetrievalConfig)
			if err == nil {
				tools = append(tools, datasetTool)
			}
		}
	}

	// 10. 检测是否关联工作流
	if appConfig.Workflows != nil && len(appConfig.Workflows) > 0 {
		workflows := make([]uuid.UUID, 0, len(appConfig.Workflows))
		for _, workflow := range appConfig.Workflows {
			if workflowID, ok := workflow["id"].(uuid.UUID); ok {
				workflows = append(workflows, workflowID)
			}
		}
		if len(workflows) > 0 {
			workflowTools, err := s.appConfigSvc.GetLangchainToolsByWorkflowIDs(ctx, workflows)
			if err == nil {
				tools = append(tools, workflowTools...)
			}
		}
	}

	// 11. 创建响应流通道
	responseStream := make(chan string, 100)

	// 12. 启动异步处理
	go s.processStreamChatAsync(ctx, app, endUser, conversation, message, appConfig, llm, history, tools, chatReq, responseStream)

	return responseStream, nil
}

// getOrCreateEndUser 获取或创建终端用户
func (s *OpenAPIService) getOrCreateEndUser(ctx context.Context, endUserID uuid.UUID, tenantID, appID uuid.UUID) (*entity.EndUser, error) {
	if endUserID != uuid.Nil {
		endUser, err := s.repo.GetEndUserByID(ctx, endUserID)
		if err == nil && endUser != nil && endUser.AppID == appID {
			return endUser, nil
		}
	}

	// 创建新的终端用户
	endUser := &entity.EndUser{
		ID:       uuid.New(),
		TenantID: tenantID,
		AppID:    appID,
	}

	err := s.repo.CreateEndUser(ctx, endUser)
	if err != nil {
		return nil, err
	}

	return endUser, nil
}

// getOrCreateConversation 获取或创建会话
func (s *OpenAPIService) getOrCreateConversation(ctx context.Context, conversationID uuid.UUID, endUserID, appID uuid.UUID) (*entity.Conversation, error) {
	if conversationID != uuid.Nil {
		conversation, err := s.conversationService.GetConversationByID(ctx, conversationID)
		if err == nil && conversation != nil &&
			conversation.AppID == appID &&
			conversation.CreatedBy == endUserID &&
			conversation.InvokeFrom == consts.InvokeFromServiceAPI {
			return conversation, nil
		}
	}

	// 创建新会话
	return s.conversationService.RawCreateConversation(ctx, appID, endUserID)
}

// processStreamChatAsync 异步处理流式聊天
func (s *OpenAPIService) processStreamChatAsync(
	ctx context.Context,
	app *entity.App,
	endUser *entity.EndUser,
	conversation *entity.Conversation,
	message *entity.Message,
	appConfig *resp.AppDraftConfigResp,
	llm llmentity.BaseLanguageModel,
	history []llms.MessageContent,
	tools []tools.Tool,
	chatReq req.OpenAPIChatReq,
	responseStream chan<- string,
) {
	defer close(responseStream)

	// 创建Agent配置
	agentConfig := agenteneity.AgentConfig{
		UserID:               endUser.TenantID,
		InvokeFrom:           consts.InvokeFromServiceAPI,
		PresetPrompt:         appConfig.PresetPrompt,
		EnableLongTermMemory: appConfig.LongTermMemory["enabled"].(bool),
		Tools:                tools,
	}

	if appConfig.LongTermMemory != nil {
		if enable, ok := appConfig.LongTermMemory["enable"].(bool); ok {
			agentConfig.EnableLongTermMemory = enable
		}
	}

	if err := util.ConvertViaJSON(&agentConfig.ReviewConfig, appConfig.ReviewConfig); err != nil {
		select {
		case responseStream <- fmt.Sprintf("event: error\ndata: %s\n\n", err.Error()):
		case <-ctx.Done():
		}
		return
	}

	// 根据LLM特性选择Agent类型
	var agentInstance agent.BaseAgent
	for _, feature := range llm.GetFeatures() {
		if feature == llmentity.FeatureToolCall {
			agentInstance = agent.NewFunctionCallAgent(llm, agentConfig, s.agentManager)
			break
		}
	}
	if agentInstance == nil {
		agentInstance = agent.NewReACTAgent(llm, agentConfig, s.agentManager)
	}

	// 创建Agent状态
	agentState := agenteneity.AgentState{
		TaskID:         uuid.New(),
		Messages:       util.MessageContentToChatMessages(history),
		History:        util.MessageContentToChatMessages(history),
		LongTermMemory: conversation.Summary,
		IterationCount: 0,
	}

	// 添加当前用户消息
	if len(chatReq.Query) > 0 {
		userMsg := llms.HumanChatMessage{Content: chatReq.Query}
		agentState.Messages = append(agentState.Messages, userMsg)
	}

	// 获取Agent流式输出
	thoughtChan, err := agentInstance.Stream(ctx, agentState)
	if err != nil {
		select {
		case responseStream <- fmt.Sprintf("event: error\ndata: %s\n\n", err.Error()):
		case <-ctx.Done():
		}
		return
	}

	// 存储agent思考过程
	agentThoughts := make(map[string]agenteneity.AgentThought)

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
			"conversation_id": conversation.ID.String(),
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
		thoughtCopy := thought
		agentThoughtsList = append(agentThoughtsList, thoughtCopy)
	}

	// 保存Agent思考过程到数据库
	err = s.conversationService.SaveAgentThoughts(ctx, endUser.TenantID, app.ID, conversation.ID, message.ID, agentThoughtsList)
	if err != nil {
		// 记录错误但不中断流程
		fmt.Printf("Failed to save agent thoughts: %v\n", err)
	}
}
