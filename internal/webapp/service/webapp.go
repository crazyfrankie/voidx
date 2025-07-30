package service

import (
	"context"
	"errors"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"

	"github.com/crazyfrankie/voidx/internal/app_config"
	"github.com/crazyfrankie/voidx/internal/conversation"
	"github.com/crazyfrankie/voidx/internal/core/agent"
	"github.com/crazyfrankie/voidx/internal/core/agent/entities"
	agententities "github.com/crazyfrankie/voidx/internal/core/agent/entities"
	llmentity "github.com/crazyfrankie/voidx/internal/core/llm/entity"
	"github.com/crazyfrankie/voidx/internal/core/memory"
	"github.com/crazyfrankie/voidx/internal/llm"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/internal/retriever"
	"github.com/crazyfrankie/voidx/internal/webapp/repository"
	"github.com/crazyfrankie/voidx/pkg/consts"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/util"
)

type WebAppService struct {
	agentManager    *agent.AgentQueueManager
	appConfigSvc    *app_config.Service
	conversationSvc *conversation.Service
	retrievalSvc    *retriever.Service
	llmSvc          *llm.Service
	tokenBufMem     *memory.TokenBufferMemory
	repo            *repository.WebAppRepo
}

func NewWebAppService(repo *repository.WebAppRepo, appConfigSvc *app_config.Service,
	conversationSvc *conversation.Service, llmSvc *llm.Service, retrievalSvc *retriever.Service,
	tokenBufMem *memory.TokenBufferMemory, agentManager *agent.AgentQueueManager) *WebAppService {
	return &WebAppService{
		repo:            repo,
		conversationSvc: conversationSvc,
		appConfigSvc:    appConfigSvc,
		retrievalSvc:    retrievalSvc,
		llmSvc:          llmSvc,
		tokenBufMem:     tokenBufMem,
		agentManager:    agentManager,
	}
}

func (s *WebAppService) GetWebAppInfo(ctx context.Context, token string) (*resp.WebAppInfoResp, error) {
	// 根据token获取应用信息
	app, err := s.repo.GetAppByToken(ctx, token)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("WebApp不存在或未发布"))
	}

	// 获取应用配置
	appConfig, err := s.appConfigSvc.GetAppConfig(ctx, app)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("应用配置不存在"))
	}

	// 从语言模型管理器中加载大语言模型以获取features
	languageModel, err := s.llmSvc.LoadLanguageModel(appConfig.ModelConfig)
	if err != nil {
		return nil, err
	}

	// 构建响应
	features := make([]string, 0, len(languageModel.GetFeatures()))
	for _, f := range languageModel.GetFeatures() {
		features = append(features, string(f))
	}
	info := &resp.WebAppInfoResp{
		ID:               app.ID,
		Name:             app.Name,
		Icon:             app.Icon,
		Description:      app.Description,
		OpeningStatement: appConfig.OpeningStatement,
		Features:         features,
	}

	// 安全地解析JSON字段
	if openingQuestions := appConfig.OpeningQuestions; openingQuestions != nil {
		info.OpeningQuestions = openingQuestions
	}
	if speechToText := appConfig.SpeechToText; speechToText != nil {
		info.SpeechToText = speechToText
	}
	if textToSpeech := appConfig.TextToSpeech; textToSpeech != nil {
		info.TextToSpeech = textToSpeech
	}
	if suggestedAfterAnswer := appConfig.SuggestedAfterAnswer; suggestedAfterAnswer != nil {
		info.SuggestedAfterAnswer = suggestedAfterAnswer
	}

	return info, nil
}

func (s *WebAppService) WebAppChat(ctx context.Context, token string, chatReq req.WebAppChatReq, accountID uuid.UUID) (<-chan string, error) {
	// 1. 获取WebApp应用并校验应用是否发布
	app, err := s.repo.GetAppByToken(ctx, token)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("WebApp不存在或未发布"))
	}

	// 2. 检测是否传递了会话id，如果传递了需要校验会话的归属信息
	var convers *entity.Conversation
	if chatReq.ConversationID != "" {
		conversationID, err := uuid.Parse(chatReq.ConversationID)
		if err != nil {
			return nil, errno.ErrValidate.AppendBizMessage(errors.New("会话ID格式错误"))
		}

		// 验证会话归属
		convers, err = s.repo.GetConversationByID(ctx, conversationID)
		if err != nil || convers == nil ||
			convers.AppID != app.ID ||
			convers.InvokeFrom != consts.InvokeFromWebApp ||
			convers.CreatedBy != accountID {
			return nil, errno.ErrForbidden.AppendBizMessage(errors.New("该会话不存在，或者不属于当前应用/用户/调用方式"))
		}
	} else {
		// 3. 如果没传递conversation_id表示新会话，这时候需要创建一个会话
		createConvReq := req.CreateConversationReq{
			AppID:      app.ID,
			Name:       "New Conversation",
			InvokeFrom: string(consts.InvokeFromWebApp),
		}
		convers, err = s.conversationSvc.CreateConversation(ctx, accountID, createConvReq)
		if err != nil {
			return nil, err
		}
	}

	// 4. 获取校验后的运行时配置
	appConfig, err := s.appConfigSvc.GetAppConfig(ctx, app)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("应用配置不存在"))
	}

	// 5. 新建一条消息记录
	createMsgReq := req.CreateMessageReq{
		ConversationID: convers.ID,
		Query:          chatReq.Query,
		ImageUrls:      chatReq.ImageUrls,
		InvokeFrom:     "web_app",
	}
	message, err := s.conversationSvc.CreateMessage(ctx, accountID, createMsgReq)
	if err != nil {
		return nil, err
	}

	// 6. 从语言模型管理器中加载大语言模型
	languageModel, err := s.llmSvc.LoadLanguageModel(appConfig.ModelConfig)
	if err != nil {
		return nil, err
	}

	// 7. 实例化TokenBufferMemory用于提取短期记忆
	tokenBufferMemory := s.tokenBufMem.WithConversationID(convers.ID).WithLLM(languageModel)

	// 解析对话轮数配置
	var dialogRound = 10 // 默认值
	if appConfig.DialogRound != 0 {
		dialogRound = appConfig.DialogRound
	}

	history, err := tokenBufferMemory.GetHistoryPromptMessages(2000, dialogRound)
	if err != nil {
		return nil, err
	}

	// 8. 将草稿配置中的tools转换成工具
	tools, err := s.appConfigSvc.GetLangchainToolsByToolsConfig(ctx, appConfig.Tools)
	if err != nil {
		return nil, err
	}

	// 9. 检测是否关联了知识库
	if appConfig.Datasets != nil && len(appConfig.Datasets) > 0 {
		// 10. 构建知识库检索工具
		var datasetIDs []uuid.UUID
		for _, dataset := range appConfig.Datasets {
			datasetIDs = append(datasetIDs, dataset["id"].(uuid.UUID))
		}

		datasetTool, err := s.retrievalSvc.CreateLangchainToolFromSearch(ctx, accountID, datasetIDs, consts.RetrievalSourceApp, appConfig.RetrievalConfig)
		if err != nil {
			return nil, err
		}
		tools = append(tools, datasetTool)
	}

	// 11. 检测是否关联工作流，如果关联了工作流则将工作流构建成工具添加到tools中
	if appConfig.Workflows != nil && len(appConfig.Workflows) > 0 {
		var workflowIDs []uuid.UUID
		for _, workflow := range appConfig.Workflows {
			workflowIDs = append(workflowIDs, workflow["id"].(uuid.UUID))
		}

		workflowTools, err := s.appConfigSvc.GetLangchainToolsByWorkflowIDs(ctx, workflowIDs)
		if err != nil {
			return nil, err
		}
		tools = append(tools, workflowTools...)
	}

	// 12. 根据LLM是否支持tool_call决定使用不同的Agent
	agentConfig := &entities.AgentConfig{
		UserID:               accountID,
		InvokeFrom:           consts.InvokeFromWebApp,
		PresetPrompt:         appConfig.PresetPrompt,
		EnableLongTermMemory: appConfig.LongTermMemory != nil && appConfig.LongTermMemory["enable"].(bool),
		Tools:                tools,
	}
	if err := util.ConvertViaJSON(&agentConfig.ReviewConfig, appConfig.ReviewConfig); err != nil {
		return nil, err
	}

	var agentInstance agent.BaseAgent
	features := languageModel.GetFeatures()
	for _, f := range features {
		if f == llmentity.FeatureFunctionCall {
			agentInstance = agent.NewFunctionCallAgent(languageModel, *agentConfig, s.agentManager)
		}
	}
	agentInstance = agent.NewReACTAgent(languageModel, *agentConfig, s.agentManager)

	// 13. 创建响应流通道
	responseStream := make(chan string, 100)

	// 14. 启动异步处理
	go s.processWebAppChat(ctx, agentInstance, convers, message, chatReq, history, accountID, responseStream)

	return responseStream, nil
}

func (s *WebAppService) StopWebAppChat(ctx context.Context, token, taskID string) error {
	// 验证应用
	_, err := s.repo.GetAppByToken(ctx, token)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("WebApp不存在或未发布"))
	}

	task, err := uuid.Parse(taskID)
	if err != nil {
		return err
	}
	uid, err := util.GetCurrentUserID(ctx)
	if err != nil {
		return err
	}

	return s.agentManager.SetStopFlag(task, consts.InvokeFromWebApp, uid)
}

func (s *WebAppService) GetConversations(ctx context.Context, token string, getReq req.GetWebAppConversationsReq) ([]resp.WebAppConversationResp, error) {
	// 验证应用
	app, err := s.repo.GetAppByToken(ctx, token)
	if err != nil {
		return nil, errno.ErrNotFound.AppendBizMessage(errors.New("WebApp不存在或未发布"))
	}

	// 使用ConversationService获取会话列表
	convReq := req.GetConversationsReq{
		AppID:    app.ID,
		IsPinned: &getReq.IsPinned,
	}
	conversations, err := s.conversationSvc.GetConversations(ctx, app.AccountID, convReq)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	conversationResps := make([]resp.WebAppConversationResp, len(conversations))
	for i, conv := range conversations {
		conversationResps[i] = resp.WebAppConversationResp{
			ID:         conv.ID,
			Name:       conv.Name,
			IsPinned:   conv.IsPinned,
			InvokeFrom: string(conv.InvokeFrom),
			Ctime:      conv.Ctime,
			Utime:      conv.Utime,
		}
	}

	return conversationResps, nil
}

// processWebAppChat 处理WebApp对话的异步逻辑
func (s *WebAppService) processWebAppChat(ctx context.Context, agentInstance agent.BaseAgent,
	conversation *entity.Conversation, message *entity.Message, chatReq req.WebAppChatReq,
	history []llms.MessageContent, accountID uuid.UUID, responseStream chan<- string) {
	defer close(responseStream)

	// 构建Agent输入
	convertHistory := make([]llms.ChatMessage, 0, len(history))
	for _, h := range history {
		convertHistory = append(convertHistory, util.MessageContentToChatMessage(h))
	}

	agentInput := agententities.AgentState{
		Messages:       []llms.ChatMessage{},
		History:        convertHistory,
		LongTermMemory: conversation.Summary,
	}

	// 定义字典存储推理过程
	agentThoughts := make(map[string]entities.AgentThought)

	// 调用智能体获取消息流
	thoughtStream, err := agentInstance.Stream(ctx, agentInput)
	if err != nil {
		s.sendErrorEvent(responseStream, conversation, message, err)
		return
	}

	// 处理Agent思考流
	for agentThought := range thoughtStream {
		eventID := agentThought.ID.String()

		// 将数据填充到agent_thought，便于存储到数据库服务中
		if agentThought.Event != agententities.EventPing {
			// 除了agent_message数据为叠加，其他均为覆盖
			if agentThought.Event == entities.EventAgentMessage {
				if existingThought, exists := agentThoughts[eventID]; exists {
					// 叠加智能体消息
					existingThought.Thought += agentThought.Thought
					existingThought.Answer += agentThought.Answer
					existingThought.MessageTokenCount = agentThought.MessageTokenCount
					existingThought.AnswerTokenCount = agentThought.AnswerTokenCount
					existingThought.TotalTokenCount = agentThought.TotalTokenCount
					existingThought.TotalPrice = agentThought.TotalPrice
					existingThought.Latency = agentThought.Latency
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
			"id":                eventID,
			"conversation_id":   conversation.ID.String(),
			"message_id":        message.ID.String(),
			"task_id":           agentThought.TaskID.String(),
			"event":             string(agentThought.Event),
			"thought":           agentThought.Thought,
			"observation":       agentThought.Observation,
			"tool":              agentThought.Tool,
			"tool_input":        agentThought.ToolInput,
			"answer":            agentThought.Answer,
			"total_token_count": agentThought.TotalTokenCount,
			"total_price":       agentThought.TotalPrice,
			"latency":           agentThought.Latency,
		}

		// 序列化并发送事件
		jsonData, err := sonic.Marshal(data)
		if err != nil {
			continue
		}

		eventStr := "event: " + string(agentThought.Event) + "\ndata:" + string(jsonData) + "\n\n"
		select {
		case responseStream <- eventStr:
		case <-ctx.Done():
			return
		}
	}

	// 将消息以及推理过程添加到数据库
	var thoughtList []entities.AgentThought
	for _, thought := range agentThoughts {
		thoughtList = append(thoughtList, thought)
	}

	err = s.conversationSvc.SaveAgentThoughts(ctx, accountID, conversation.AppID, conversation.ID, message.ID, thoughtList)
	if err != nil {
		// 记录错误但不中断流程
		// TODO: 添加日志记录
	}
}

// sendErrorEvent 发送错误事件
func (s *WebAppService) sendErrorEvent(responseStream chan<- string, conversation *entity.Conversation, message *entity.Message, err error) {
	errorData := map[string]any{
		"id":              uuid.New().String(),
		"conversation_id": conversation.ID.String(),
		"message_id":      message.ID.String(),
		"task_id":         uuid.New().String(),
		"event":           "error",
		"answer":          "抱歉，处理您的请求时出现错误: " + err.Error(),
		"latency":         0,
	}

	jsonData, _ := sonic.Marshal(errorData)
	eventStr := "event: error\ndata:" + string(jsonData) + "\n\n"

	select {
	case responseStream <- eventStr:
	default:
	}
}

func (s *WebAppService) GetConversationMessages(ctx context.Context, token, conversationID string, pageReq req.GetWebAppConversationMessagesReq) ([]resp.MessageResp, resp.Paginator, error) {
	// 验证应用
	_, err := s.repo.GetAppByToken(ctx, token)
	if err != nil {
		return nil, resp.Paginator{}, errno.ErrNotFound.AppendBizMessage(errors.New("WebApp不存在或未发布"))
	}

	// 解析会话ID
	convID, err := uuid.Parse(conversationID)
	if err != nil {
		return nil, resp.Paginator{}, errno.ErrValidate.AppendBizMessage(errors.New("会话ID格式错误"))
	}

	// 使用ConversationService获取消息列表
	msgReq := req.GetConversationMessagesWithPageReq{
		CurrentPage: pageReq.Page,
		PageSize:    pageReq.PageSize,
	}
	messages, paginator, err := s.conversationSvc.GetConversationMessagesWithPage(ctx, convID, msgReq)
	if err != nil {
		return nil, resp.Paginator{}, err
	}

	// 转换为响应格式
	messageResps := make([]resp.MessageResp, len(messages))
	for i, msg := range messages {
		messageResps[i] = resp.MessageResp{
			ID:             msg.ID,
			ConversationID: msg.ConversationID,
			AppID:          msg.AppID,
			InvokeFrom:     string(msg.InvokeFrom),
			CreatedBy:      msg.CreatedBy,
			Query:          msg.Query,
			ImageUrls:      msg.ImageUrls,
			Answer:         msg.Answer,
			Status:         string(msg.Status),
			Ctime:          msg.Ctime,
		}
	}

	return messageResps, paginator, nil
}

func (s *WebAppService) DeleteConversation(ctx context.Context, token, conversationID string) error {
	// 验证应用
	_, err := s.repo.GetAppByToken(ctx, token)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("WebApp不存在或未发布"))
	}

	// 解析会话ID
	convID, err := uuid.Parse(conversationID)
	if err != nil {
		return errno.ErrValidate.AppendBizMessage(errors.New("会话ID格式错误"))
	}

	// 使用ConversationService删除会话
	return s.conversationSvc.DeleteConversation(ctx, convID)
}

func (s *WebAppService) UpdateConversationName(ctx context.Context, token, conversationID, name string) error {
	// 验证应用
	_, err := s.repo.GetAppByToken(ctx, token)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("WebApp不存在或未发布"))
	}

	// 解析会话ID
	convID, err := uuid.Parse(conversationID)
	if err != nil {
		return errno.ErrValidate.AppendBizMessage(errors.New("会话ID格式错误"))
	}

	// 使用ConversationService更新会话名称
	return s.conversationSvc.UpdateConversationName(ctx, convID, name)
}

func (s *WebAppService) UpdateConversationPin(ctx context.Context, token, conversationID string, isPinned bool) error {
	// 验证应用
	_, err := s.repo.GetAppByToken(ctx, token)
	if err != nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("WebApp不存在或未发布"))
	}

	// 解析会话ID
	convID, err := uuid.Parse(conversationID)
	if err != nil {
		return errno.ErrValidate.AppendBizMessage(errors.New("会话ID格式错误"))
	}

	// 使用ConversationService更新会话置顶状态
	return s.conversationSvc.UpdateConversationIsPinned(ctx, convID, isPinned)
}
