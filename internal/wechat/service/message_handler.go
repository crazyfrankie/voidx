package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	lcllms "github.com/tmc/langchaingo/llms"

	"github.com/crazyfrankie/voidx/internal/core/agent"
	agenteneity "github.com/crazyfrankie/voidx/internal/core/agent/entities"
	llmentity "github.com/crazyfrankie/voidx/internal/core/llm/entity"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/internal/wechat/repository"
	"github.com/crazyfrankie/voidx/pkg/logs"
	"github.com/crazyfrankie/voidx/pkg/util"
	"github.com/crazyfrankie/voidx/types/consts"
)

// handleResultQuery 处理结果查询（消息"1"）
func (s *WechatService) handleResultQuery(ctx context.Context, wechatEndUserID uuid.UUID) string {
	// 查询最新的微信消息记录
	wechatMessage, err := s.repo.GetLatestWechatMessage(ctx, wechatEndUserID)
	if err != nil {
		if errors.Is(err, repository.ErrMessageNotFound) {
			return ""
		}
		return "系统错误，请稍后重试"
	}

	// 检查消息是否已推送
	if wechatMessage.IsPushed {
		return ""
	}

	// 获取关联的Agent消息
	msg, err := s.repo.GetMessage(ctx, wechatMessage.MessageID)
	if err != nil {
		return "系统错误，请稍后重试"
	}

	// 根据消息状态返回不同内容
	switch msg.Status {
	case consts.MessageStatusNormal, consts.MessageStatusStop:
		if strings.TrimSpace(msg.Answer) != "" {
			// 标记消息已推送
			if err := s.repo.UpdateWechatMessage(ctx, wechatMessage.ID, map[string]interface{}{
				"is_pushed": true,
			}); err != nil {
				return ""
			}
			return strings.TrimSpace(msg.Answer)
		}
		return "该Agent智能体任务正在处理中，请稍后重新回复`1`获取结果。"
	case consts.MessageStatusTimeout:
		return "该Agent智能体处理任务超时，请重新发起提问。"
	case consts.MessageStatusError:
		return "该Agent智能体处理任务出错，请重新发起提问，错误信息: " + msg.Error + "。"
	default:
		return "该Agent智能体任务正在处理中，请稍后重新回复`1`获取结果。"
	}
}

// handleNormalMessage 处理普通消息
func (s *WechatService) handleNormalMessage(ctx context.Context, content string, wechatEndUser *entity.WechatEndUser, app *entity.App) *message.Reply {
	// 获取应用配置
	appConfig, err := s.appConfigSvc.GetAppConfig(ctx, app)
	if err != nil {
		return &message.Reply{
			MsgType: message.MsgTypeText,
			MsgData: message.NewText("系统错误，请稍后重试"),
		}
	}

	// 创建消息记录
	convers, err := s.repo.GetConversation(ctx, app.ID, wechatEndUser.EndUserID)
	if err != nil {
		return nil
	}
	msg := &entity.Message{
		ID:             uuid.New(),
		AppID:          app.ID,
		ConversationID: convers.ID,
		InvokeFrom:     consts.InvokeFromServiceAPI,
		CreatedBy:      wechatEndUser.EndUserID,
		Query:          content,
		ImageUrls:      []string{},
		Status:         consts.MessageStatusNormal,
	}

	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		return &message.Reply{
			MsgType: message.MsgTypeText,
			MsgData: message.NewText("系统错误，请稍后重试"),
		}
	}

	// 创建微信消息记录
	wechatMsg := &entity.WechatMessage{
		ID:              uuid.New(),
		WechatEndUserID: wechatEndUser.ID,
		MessageID:       msg.ID,
		IsPushed:        false,
	}

	if err := s.repo.CreateWechatMessage(ctx, wechatMsg); err != nil {
		return &message.Reply{
			MsgType: message.MsgTypeText,
			MsgData: message.NewText("系统错误，请稍后重试"),
		}
	}

	// 启动异步处理
	go s.processMessageAsync(ctx, app, appConfig, msg.ID, convers.ID, content)

	return &message.Reply{
		MsgType: message.MsgTypeText,
		MsgData: message.NewText("思考中，请回复\"1\"获取结果。"),
	}
}

// processMessageAsync 异步处理消息
func (s *WechatService) processMessageAsync(ctx context.Context, app *entity.App,
	appConfig *resp.AppDraftConfigResp, messageID uuid.UUID, conversationID uuid.UUID, query string) {
	// 1. 加载语言模型
	llm, err := s.llmSvc.LoadLanguageModel(appConfig.ModelConfig)
	if err != nil {
		logs.Errorf("Failed to load language model: %v", err)
		return
	}
	// 2. 获取对话历史
	conversation, err := s.repo.GetConversationByID(ctx, conversationID)
	if err != nil {
		logs.Errorf("Failed to get conversation by ID %s: %v", conversationID, err)
		return
	}
	s.tokenBufMem = s.tokenBufMem.WithLLM(llm)
	history, err := s.tokenBufMem.GetHistoryPromptMessages(2000, appConfig.DialogRound)
	if err != nil {
		logs.Errorf("Failed to get conversation history: %v", err)
		return
	}

	// 3. 构建工具链
	tools, err := s.appConfigSvc.GetLangchainToolsByToolsConfig(ctx, appConfig.Tools)
	if err != nil {
		logs.Errorf("Failed to get langchain tools by config: %v", err)
		return
	}

	// 4. 检测是否关联了知识库
	if appConfig.Datasets != nil {
		datasetIDs := make([]uuid.UUID, 0, len(appConfig.Datasets))
		for _, dataset := range appConfig.Datasets {
			datasetIDs = append(datasetIDs, dataset["id"].(uuid.UUID))
		}
		// 5.构建LangChain知识库检索工具
		datasetRetrieval, err := s.retrievalSvc.CreateLangchainToolFromSearch(ctx, app.AccountID,
			datasetIDs, consts.RetrievalSourceApp, appConfig.RetrievalConfig)
		if err != nil {
			logs.Errorf("Failed to create dataset retrieval tool: %v", err)
			return
		}
		tools = append(tools, datasetRetrieval)
	}

	// 6.检测是否关联工作流，如果关联了工作流则将工作流构建成工具添加到tools中
	if appConfig.Workflows != nil {
		workflowIDs := make([]uuid.UUID, 0, len(appConfig.Workflows))
		for _, workflow := range appConfig.Workflows {
			workflowIDs = append(workflowIDs, workflow["id"].(uuid.UUID))
		}
		// 5.构建LangChain知识库检索工具
		workflowTool, err := s.appConfigSvc.GetLangchainToolsByWorkflowIDs(ctx, workflowIDs)
		if err != nil {
			logs.Errorf("Failed to get workflow tools: %v", err)
			return
		}
		tools = append(tools, workflowTool...)
	}

	// 7.根据LLM是否支持tool_call决定使用不同的Agent
	agentCfg := agenteneity.AgentConfig{
		UserID:               app.AccountID,
		InvokeFrom:           consts.InvokeFromDebugger,
		PresetPrompt:         appConfig.PresetPrompt,
		EnableLongTermMemory: appConfig.LongTermMemory["enabled"].(bool),
		Tools:                tools,
	}
	if err := util.ConvertViaJSON(&agentCfg.ReviewConfig, appConfig.ReviewConfig); err != nil {
		logs.Errorf("Failed to convert review config: %v", err)
		return
	}
	agentIns := agent.NewFunctionCallAgent(llm, agentCfg, s.agentManager)
	for _, f := range llm.GetFeatures() {
		if f == llmentity.FeatureToolCall {
			agentIns = agent.NewFunctionCallAgent(llm, agentCfg, s.agentManager)
			break
		}
	}

	// 8.定义智能体状态基础数据
	agentState := agenteneity.AgentState{
		History:        util.MessageContentToChatMessages(history),
		LongTermMemory: conversation.Summary,
		Messages:       []lcllms.ChatMessage{util.MessageContentToChatMessage(llm.ConvertToHumanMessage(query, nil))},
	}

	// 9.调用智能体获取执行结果
	agentResult, err := agentIns.Invoke(ctx, agentState)
	if err != nil {
		logs.Errorf("Agent invocation failed: %v", err)
		return
	}

	// 10.将数据存储到数据库中，包含会话、消息、推理过程
	err = s.conversationSvc.SaveAgentThoughts(ctx, app.AccountID, app.ID, conversationID, messageID, agentResult.AgentThoughts)
	if err != nil {
		logs.Errorf("Failed to save agent thoughts: %v", err)
	}
}
