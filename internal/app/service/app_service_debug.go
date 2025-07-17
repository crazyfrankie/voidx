package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"

	coreEntity "github.com/crazyfrankie/voidx/internal/core/llm/entity"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/errno"
)

// TaskManager 任务管理器，用于管理调试聊天任务
type TaskManager struct {
	tasks sync.Map // key: taskID, value: *TaskContext
}

// TaskContext 任务上下文
type TaskContext struct {
	TaskID     uuid.UUID
	AppID      uuid.UUID
	AccountID  uuid.UUID
	CancelFunc context.CancelFunc
	Status     string // "running", "stopped", "completed"
	CreatedAt  time.Time
}

// 全局任务管理器
var taskManager = &TaskManager{}

// StartTask 启动任务
func (tm *TaskManager) StartTask(taskID, appID, accountID uuid.UUID, cancelFunc context.CancelFunc) {
	tm.tasks.Store(taskID, &TaskContext{
		TaskID:     taskID,
		AppID:      appID,
		AccountID:  accountID,
		CancelFunc: cancelFunc,
		Status:     "running",
		CreatedAt:  time.Now(),
	})
}

// StopTask 停止任务
func (tm *TaskManager) StopTask(taskID uuid.UUID) bool {
	if value, ok := tm.tasks.Load(taskID); ok {
		taskCtx := value.(*TaskContext)
		taskCtx.Status = "stopped"
		taskCtx.CancelFunc()
		tm.tasks.Delete(taskID)
		return true
	}
	return false
}

// CompleteTask 完成任务
func (tm *TaskManager) CompleteTask(taskID uuid.UUID) {
	if value, ok := tm.tasks.Load(taskID); ok {
		taskCtx := value.(*TaskContext)
		taskCtx.Status = "completed"
		tm.tasks.Delete(taskID)
	}
}

// IsTaskRunning 检查任务是否正在运行
func (tm *TaskManager) IsTaskRunning(taskID uuid.UUID) bool {
	if value, ok := tm.tasks.Load(taskID); ok {
		taskCtx := value.(*TaskContext)
		return taskCtx.Status == "running"
	}
	return false
}

// GetTaskContext 获取任务上下文
func (tm *TaskManager) GetTaskContext(taskID uuid.UUID) (*TaskContext, bool) {
	if value, ok := tm.tasks.Load(taskID); ok {
		return value.(*TaskContext), true
	}
	return nil, false
}

// GetDebugConversationSummary 获取调试会话长期记忆
func (s *AppService) GetDebugConversationSummary(ctx context.Context, appID uuid.UUID) (string, error) {
	// 获取当前用户ID
	accountID, err := getCurrentUserID(ctx)
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

	// 获取草稿配置
	draftConfig, err := s.GetDraftAppConfig(ctx, appID)
	if err != nil {
		return "", err
	}

	// 检查长期记忆是否启用
	longTermMemory, ok := draftConfig["long_term_memory"].(map[string]interface{})
	if !ok || !longTermMemory["enable"].(bool) {
		return "", errno.ErrValidate.AppendBizMessage("该应用并未开启长期记忆，无法获取")
	}

	// 获取调试会话
	if app.DebugConversationID == nil {
		return "", nil
	}

	conversation, err := s.repo.GetConversation(ctx, *app.DebugConversationID)
	if err != nil {
		return "", err
	}

	return conversation.Summary, nil
}

// UpdateDebugConversationSummary 更新调试会话长期记忆
func (s *AppService) UpdateDebugConversationSummary(ctx context.Context, appID uuid.UUID, summary string) error {
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

	// 获取草稿配置
	draftConfig, err := s.GetDraftAppConfig(ctx, appID)
	if err != nil {
		return err
	}

	// 检查长期记忆是否启用
	longTermMemory, ok := draftConfig["long_term_memory"].(map[string]interface{})
	if !ok || !longTermMemory["enable"].(bool) {
		return errno.ErrValidate.AppendBizMessage("该应用并未开启长期记忆，无法更新")
	}

	// 获取或创建调试会话
	var conversation *entity.Conversation
	if app.DebugConversationID != nil {
		conversation, err = s.repo.GetConversation(ctx, *app.DebugConversationID)
		if err != nil {
			return err
		}
	} else {
		// 创建新的调试会话
		conversation, err = s.repo.CreateConversation(ctx, accountID, appID)
		if err != nil {
			return err
		}

		// 更新应用的调试会话ID
		app.DebugConversationID = &conversation.ID
		if err := s.repo.UpdateApp(ctx, app); err != nil {
			return err
		}
	}

	// 更新摘要
	conversation.Summary = summary
	if err := s.repo.UpdateConversation(ctx, conversation); err != nil {
		return err
	}

	return nil
}

// DeleteDebugConversation 删除调试会话
func (s *AppService) DeleteDebugConversation(ctx context.Context, appID uuid.UUID) error {
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

	// 检查调试会话ID是否存在
	if app.DebugConversationID == nil {
		return nil
	}

	// 更新应用的调试会话ID
	app.DebugConversationID = nil
	if err := s.repo.UpdateApp(ctx, app); err != nil {
		return err
	}

	return nil
}

// DebugChat 发起调试对话
func (s *AppService) DebugChat(ctx context.Context, appID uuid.UUID, req req.DebugChatReq, w http.ResponseWriter) error {
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

	// 获取草稿配置
	draftConfig, err := s.GetDraftAppConfig(ctx, appID)
	if err != nil {
		return err
	}

	// 获取或创建调试会话
	var conversation *entity.Conversation
	if app.DebugConversationID != nil {
		conversation, err = s.repo.GetConversation(ctx, *app.DebugConversationID)
		if err != nil {
			return err
		}
	} else {
		// 创建新的调试会话
		conversation, err = s.repo.CreateConversation(ctx, accountID, appID)
		if err != nil {
			return err
		}

		// 更新应用的调试会话ID
		app.DebugConversationID = &conversation.ID
		if err := s.repo.UpdateApp(ctx, app); err != nil {
			return err
		}
	}

	// 创建消息记录
	createdMessage, err := s.repo.CreateMessage(ctx, appID, conversation.ID, accountID, "debugger", req.Query, req.ImageUrls)
	if err != nil {
		return err
	}

	// 生成任务ID
	taskID := uuid.New()

	// 设置SSE响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 发送事件到客户端
	sendSSEEvent := func(event string, data interface{}) {
		jsonData, _ := json.Marshal(data)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, string(jsonData))
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}

	// 创建可取消的上下文
	taskCtx, cancel := context.WithCancel(ctx)

	// 注册任务到任务管理器
	taskManager.StartTask(taskID, appID, accountID, cancel)

	// 异步处理聊天逻辑
	go func() {
		defer func() {
			// 完成任务
			taskManager.CompleteTask(taskID)

			// 发送结束事件
			sendSSEEvent("message_end", map[string]interface{}{
				"id":              taskID.String(),
				"conversation_id": conversation.ID.String(),
				"message_id":      createdMessage.ID.String(),
				"task_id":         taskID.String(),
			})
		}()

		// 检查任务是否被取消
		select {
		case <-taskCtx.Done():
			// 任务被取消
			createdMessage.Status = "stopped"
			s.repo.UpdateMessage(taskCtx, createdMessage)

			sendSSEEvent("message_stop", map[string]interface{}{
				"id":              taskID.String(),
				"conversation_id": conversation.ID.String(),
				"message_id":      createdMessage.ID.String(),
				"task_id":         taskID.String(),
			})
			return
		default:
		}

		// 发送开始事件
		sendSSEEvent("message", map[string]interface{}{
			"id":                taskID.String(),
			"conversation_id":   conversation.ID.String(),
			"message_id":        createdMessage.ID.String(),
			"task_id":           taskID.String(),
			"event":             "message_start",
			"answer":            "",
			"thought":           "",
			"observation":       "",
			"tool":              nil,
			"tool_input":        nil,
			"total_token_count": 0,
			"total_price":       0.0,
			"latency":           0.0,
		})

		// 获取模型配置
		modelConfig, ok := draftConfig["model_config"].(map[string]interface{})
		if !ok {
			sendSSEEvent("error", map[string]interface{}{
				"error": "模型配置不正确",
			})
			return
		}

		// 检查审核配置
		reviewConfig := make(map[string]interface{})
		if rc, exists := draftConfig["review_config"]; exists {
			if rcMap, ok := rc.(map[string]interface{}); ok {
				reviewConfig = rcMap
			}
		}

		// 验证输入内容
		if passed, response := s.validateReviewConfig(taskCtx, req.Query, reviewConfig); !passed {
			// 输入未通过审核，返回预设响应
			sendSSEEvent("message", map[string]interface{}{
				"id":                taskID.String(),
				"conversation_id":   conversation.ID.String(),
				"message_id":        createdMessage.ID.String(),
				"task_id":           taskID.String(),
				"event":             "agent_message",
				"answer":            response,
				"thought":           "输入内容未通过审核",
				"observation":       "",
				"tool":              nil,
				"tool_input":        nil,
				"total_token_count": len([]rune(response)),
				"total_price":       0.0,
				"latency":           0.0,
			})

			// 更新消息状态
			createdMessage.Answer = response
			createdMessage.Status = "normal"
			s.repo.UpdateMessage(taskCtx, createdMessage)

			// 保存审核结果
			s.saveAgentThought(taskCtx, &entity.AgentThought{
				MessageID:       createdMessage.ID,
				Event:           "agent_message",
				Thought:         "输入内容未通过审核",
				Answer:          response,
				TotalTokenCount: len([]rune(response)),
				TotalPrice:      0.0,
				Latency:         0.0,
				Ctime:           time.Now().Unix(),
			})
			return
		}

		// 使用LLM服务加载语言模型
		llmModel, err := s.llmService.LoadLanguageModel(modelConfig)
		if err != nil {
			sendSSEEvent("error", map[string]interface{}{
				"error": fmt.Sprintf("加载语言模型失败: %v", err),
			})
			return
		}

		// 发送思考事件
		sendSSEEvent("message", map[string]interface{}{
			"id":                taskID.String(),
			"conversation_id":   conversation.ID.String(),
			"message_id":        createdMessage.ID.String(),
			"task_id":           taskID.String(),
			"event":             "agent_thought",
			"thought":           "正在分析用户问题...",
			"answer":            "",
			"observation":       "",
			"tool":              nil,
			"tool_input":        nil,
			"total_token_count": 0,
			"total_price":       0.0,
			"latency":           0.0,
		})

		// 构建完整的提示消息
		promptMessage, err := s.buildPromptMessage(taskCtx, req.Query, draftConfig, conversation)
		if err != nil {
			sendSSEEvent("error", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// 发送生成回答事件
		sendSSEEvent("message", map[string]interface{}{
			"id":                taskID.String(),
			"conversation_id":   conversation.ID.String(),
			"message_id":        createdMessage.ID.String(),
			"task_id":           taskID.String(),
			"event":             "agent_thought",
			"thought":           "正在生成回答...",
			"answer":            "",
			"observation":       "",
			"tool":              nil,
			"tool_input":        nil,
			"total_token_count": 0,
			"total_price":       0.0,
			"latency":           0.0,
		})

		// 调用LLM生成响应
		startTime := time.Now()
		answer, err := s.generateResponseWithLLM(taskCtx, llmModel, promptMessage, req.ImageUrls)
		if err != nil {
			// 检查是否是因为任务被取消
			if errors.Is(taskCtx.Err(), context.Canceled) {
				sendSSEEvent("message_stop", map[string]interface{}{
					"id":              taskID.String(),
					"conversation_id": conversation.ID.String(),
					"message_id":      createdMessage.ID.String(),
					"task_id":         taskID.String(),
				})
				return
			}

			sendSSEEvent("error", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// 计算延迟
		latency := time.Since(startTime).Seconds()

		// 验证输出内容
		if !s.validateOutputContent(taskCtx, answer, reviewConfig) {
			// 输出未通过审核，返回提示信息
			auditResponse := "抱歉，生成的回答包含不当内容，请重新尝试。"

			sendSSEEvent("message", map[string]interface{}{
				"id":                taskID.String(),
				"conversation_id":   conversation.ID.String(),
				"message_id":        createdMessage.ID.String(),
				"task_id":           taskID.String(),
				"event":             "agent_message",
				"answer":            auditResponse,
				"thought":           "生成的回答未通过审核",
				"observation":       "",
				"tool":              nil,
				"tool_input":        nil,
				"total_token_count": len([]rune(auditResponse)),
				"total_price":       0.0,
				"latency":           latency,
			})

			// 更新消息状态
			createdMessage.Answer = auditResponse
			createdMessage.Status = "normal"
			s.repo.UpdateMessage(taskCtx, createdMessage)

			// 保存审核结果
			s.saveAgentThought(taskCtx, &entity.AgentThought{
				MessageID:       createdMessage.ID,
				Event:           "agent_message",
				Thought:         "生成的回答未通过审核",
				Answer:          auditResponse,
				TotalTokenCount: len([]rune(auditResponse)),
				TotalPrice:      0.0,
				Latency:         latency,
				Ctime:           time.Now().Unix(),
			})
			return
		}

		// 流式发送答案
		s.streamAnswerWithStats(taskCtx, answer, taskID, conversation.ID, createdMessage.ID, latency, sendSSEEvent)

		// 更新消息状态
		createdMessage.Answer = answer
		createdMessage.Status = "normal"
		s.repo.UpdateMessage(taskCtx, createdMessage)

		// 保存Agent思考过程到数据库
		s.saveAgentThought(taskCtx, &entity.AgentThought{
			MessageID:       createdMessage.ID,
			Event:           "agent_message",
			Thought:         "分析用户问题并生成回答",
			Answer:          answer,
			TotalTokenCount: len([]rune(answer)), // 简单的token计算
			TotalPrice:      0.0,                 // 实际应该根据模型定价计算
			Latency:         latency,
			Ctime:           time.Now().Unix(),
		})
	}()

	return nil
}

// StopDebugChat 停止调试聊天任务
func (s *AppService) StopDebugChat(ctx context.Context, appID, taskID uuid.UUID) error {
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

	// 获取任务上下文
	taskCtx, exists := taskManager.GetTaskContext(taskID)
	if !exists {
		return errno.ErrNotFound.AppendBizMessage("任务不存在或已完成")
	}

	// 检查任务所属权限
	if taskCtx.AccountID != accountID || taskCtx.AppID != appID {
		return errno.ErrForbidden.AppendBizMessage("无权停止该任务")
	}

	// 停止任务
	if !taskManager.StopTask(taskID) {
		return errno.ErrValidate.AppendBizMessage("任务无法停止")
	}

	return nil
}

// GetDebugConversationMessagesWithPage 获取调试会话分页消息列表
func (s *AppService) GetDebugConversationMessagesWithPage(ctx context.Context, appID uuid.UUID, req req.GetDebugConversationMessagesWithPageReq) ([]*resp.DebugConversationMessageResp, *resp.Paginator, error) {
	// 获取当前用户ID
	accountID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, nil, err
	}

	// 获取应用
	app, err := s.repo.GetAppByID(ctx, appID)
	if err != nil {
		return nil, nil, err
	}

	// 检查权限
	if app.AccountID != accountID {
		return nil, nil, errno.ErrForbidden.AppendBizMessage("无权访问该应用")
	}

	// 检查调试会话是否存在
	if app.DebugConversationID == nil {
		return []*resp.DebugConversationMessageResp{}, &resp.Paginator{
			CurrentPage: req.Page,
			PageSize:    req.PageSize,
			TotalPage:   0,
			TotalRecord: 0,
		}, nil
	}

	// 获取消息列表
	messages, total, err := s.repo.GetDebugConversationMessagesWithPage(ctx, *app.DebugConversationID, req.Page, req.PageSize, req.Ctime)
	if err != nil {
		return nil, nil, err
	}

	// 构建响应
	var messageResps []*resp.DebugConversationMessageResp
	for _, message := range messages {
		messageResp := &resp.DebugConversationMessageResp{
			ID:             message.ID,
			ConversationID: message.ConversationID,
			Query:          message.Query,
			Answer:         message.Answer,
			Ctime:          message.Ctime,
		}

		// 添加智能体思考过程
		for _, thought := range message.AgentThoughts {
			messageResp.AgentThoughts = append(messageResp.AgentThoughts, resp.AgentThought{
				ID:              thought.ID,
				Event:           thought.Event,
				Thought:         thought.Thought,
				Observation:     thought.Observation,
				Tool:            thought.Tool,
				ToolInput:       thought.ToolInput,
				Answer:          thought.Answer,
				TotalTokenCount: thought.TotalTokenCount,
				TotalPrice:      thought.TotalPrice,
				Latency:         thought.Latency,
				Ctime:           thought.Ctime,
			})
		}

		messageResps = append(messageResps, messageResp)
	}

	// 计算分页信息
	totalPage := int(total) / req.PageSize
	if int(total)%req.PageSize != 0 {
		totalPage++
	}

	paginator := &resp.Paginator{
		CurrentPage: req.Page,
		PageSize:    req.PageSize,
		TotalPage:   totalPage,
		TotalRecord: int(total),
	}

	return messageResps, paginator, nil
}

// buildPromptMessage 构建完整的提示消息
func (s *AppService) buildPromptMessage(ctx context.Context, query string, draftConfig map[string]interface{}, conversation *entity.Conversation) (string, error) {
	var promptBuilder strings.Builder

	// 添加预设提示
	if presetPrompt, exists := draftConfig["preset_prompt"]; exists {
		if prompt, ok := presetPrompt.(string); ok && prompt != "" {
			promptBuilder.WriteString(prompt)
			promptBuilder.WriteString("\n\n")
		}
	}

	// 添加长期记忆
	if longTermMemory, exists := draftConfig["long_term_memory"]; exists {
		if memoryConfig, ok := longTermMemory.(map[string]interface{}); ok {
			if enable, ok := memoryConfig["enable"].(bool); ok && enable {
				if conversation.Summary != "" {
					promptBuilder.WriteString("长期记忆：")
					promptBuilder.WriteString(conversation.Summary)
					promptBuilder.WriteString("\n\n")
				}
			}
		}
	}

	// 添加历史对话上下文
	if dialogRound, exists := draftConfig["dialog_round"]; exists {
		if rounds, ok := dialogRound.(float64); ok && rounds > 0 {
			historyMessages, err := s.repo.GetRecentMessages(ctx, conversation.ID, int(rounds))
			if err == nil && len(historyMessages) > 0 {
				promptBuilder.WriteString("历史对话：\n")
				for _, msg := range historyMessages {
					promptBuilder.WriteString(fmt.Sprintf("用户：%s\n", msg.Query))
					if msg.Answer != "" {
						promptBuilder.WriteString(fmt.Sprintf("助手：%s\n", msg.Answer))
					}
				}
				promptBuilder.WriteString("\n")
			}
		}
	}

	// 添加当前用户问题
	promptBuilder.WriteString("用户：")
	promptBuilder.WriteString(query)

	return promptBuilder.String(), nil
}

// streamAnswer 流式发送答案
func (s *AppService) streamAnswer(ctx context.Context, answer string, taskID, conversationID, messageID uuid.UUID, sendSSEEvent func(string, interface{})) {
	// 按字符流式发送
	runes := []rune(answer)
	chunkSize := 5 // 每次发送5个字符

	for i := 0; i < len(runes); i += chunkSize {
		// 检查上下文是否被取消
		select {
		case <-ctx.Done():
			return
		default:
		}

		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}

		chunk := string(runes[i:end])
		sendSSEEvent("message", map[string]interface{}{
			"id":                taskID.String(),
			"conversation_id":   conversationID.String(),
			"message_id":        messageID.String(),
			"task_id":           taskID.String(),
			"event":             "agent_message",
			"answer":            chunk,
			"thought":           "",
			"observation":       "",
			"tool":              nil,
			"tool_input":        nil,
			"total_token_count": 0,
			"total_price":       0.0,
			"latency":           0.0,
		})

		// 模拟流式效果
		time.Sleep(50 * time.Millisecond)
	}
}

// generateResponseWithLLM 使用LLM生成响应
func (s *AppService) generateResponseWithLLM(ctx context.Context, llmModel coreEntity.BaseLanguageModel, prompt string, imageUrls []string) (string, error) {
	// 检查上下文是否被取消
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// 构建消息内容
	messageContent := llmModel.ConvertToHumanMessage(prompt, imageUrls)

	// 调用LLM生成内容
	response, err := llmModel.GenerateContent(ctx, []llms.MessageContent{messageContent})
	if err != nil {
		return "", fmt.Errorf("LLM生成内容失败: %w", err)
	}

	// 提取生成的文本
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("LLM未返回任何响应")
	}

	choice := response.Choices[0]

	// 直接获取Content内容 - 修复类型错误
	return choice.Content, nil
}

// streamAnswerWithStats 流式发送答案并包含统计信息
func (s *AppService) streamAnswerWithStats(ctx context.Context, answer string, taskID, conversationID, messageID uuid.UUID, latency float64, sendSSEEvent func(string, interface{})) {
	// 按字符流式发送
	runes := []rune(answer)
	chunkSize := 5 // 每次发送5个字符
	totalTokens := len(runes)

	var currentAnswer strings.Builder

	for i := 0; i < len(runes); i += chunkSize {
		// 检查上下文是否被取消
		select {
		case <-ctx.Done():
			return
		default:
		}

		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}

		chunk := string(runes[i:end])
		currentAnswer.WriteString(chunk)

		// 计算进度
		progress := float64(end) / float64(totalTokens)

		sendSSEEvent("message", map[string]interface{}{
			"id":                taskID.String(),
			"conversation_id":   conversationID.String(),
			"message_id":        messageID.String(),
			"task_id":           taskID.String(),
			"event":             "agent_message",
			"answer":            chunk,
			"thought":           "",
			"observation":       "",
			"tool":              nil,
			"tool_input":        nil,
			"total_token_count": totalTokens,
			"total_price":       0.0, // 实际应该根据模型定价计算
			"latency":           latency,
			"progress":          progress,
		})

		// 模拟流式效果
		time.Sleep(50 * time.Millisecond)
	}
}

// saveAgentThought 保存Agent思考过程到数据库
func (s *AppService) saveAgentThought(ctx context.Context, thought *entity.AgentThought) error {
	// 使用repository保存思考过程
	_, err := s.repo.CreateAgentThought(ctx,
		thought.MessageID,
		thought.Event,
		thought.Thought,
		thought.Observation,
		thought.Tool,
		thought.ToolInput,
		thought.Answer,
		thought.TotalTokenCount,
		thought.TotalPrice,
		thought.Latency,
	)
	return err
}

// calculateTokenPrice 计算token价格
func (s *AppService) calculateTokenPrice(llmModel coreEntity.BaseLanguageModel, inputTokens, outputTokens int) float64 {
	inputPrice, outputPrice, unit := llmModel.GetPricing()

	// 计算总价格
	totalPrice := (float64(inputTokens)*inputPrice + float64(outputTokens)*outputPrice) / unit

	return totalPrice
}

// processTools 处理工具配置，转换为LangChain工具
func (s *AppService) processTools(ctx context.Context, draftConfig map[string]interface{}) ([]tools.Tool, error) {
	var allTools []tools.Tool

	// 处理内置工具和API工具
	if toolsConfig, exists := draftConfig["tools"]; exists {
		if toolsList, ok := toolsConfig.([]interface{}); ok {
			for _, toolItem := range toolsList {
				if toolMap, ok := toolItem.(map[string]interface{}); ok {
					tool, err := s.createToolFromConfig(toolMap)
					if err != nil {
						continue // 忽略错误的工具配置
					}
					if tool != nil {
						allTools = append(allTools, tool)
					}
				}
			}
		}
	}

	// 处理知识库检索工具
	if datasets, exists := draftConfig["datasets"]; exists {
		if datasetsList, ok := datasets.([]interface{}); ok && len(datasetsList) > 0 {
			// 创建知识库检索工具
			// 这里应该调用检索服务来创建工具
			// 暂时跳过，实际实现时需要集成检索服务
		}
	}

	// 处理工作流工具
	if workflows, exists := draftConfig["workflows"]; exists {
		if workflowsList, ok := workflows.([]interface{}); ok && len(workflowsList) > 0 {
			for _, workflowItem := range workflowsList {
				if workflowMap, ok := workflowItem.(map[string]interface{}); ok {
					if workflowID, exists := workflowMap["id"]; exists {
						// 创建工作流工具
						// 这里应该调用工作流服务来创建工具
						// 暂时跳过，实际实现时需要集成工作流服务
						_ = workflowID
					}
				}
			}
		}
	}

	return allTools, nil
}

// createToolFromConfig 根据配置创建工具
func (s *AppService) createToolFromConfig(toolConfig map[string]interface{}) (tools.Tool, error) {
	toolType, ok := toolConfig["type"].(string)
	if !ok {
		return nil, fmt.Errorf("工具类型配置错误")
	}

	switch toolType {
	case "builtin_tool":
		return s.createBuiltinTool(toolConfig)
	case "api_tool":
		return s.createAPITool(toolConfig)
	default:
		return nil, fmt.Errorf("不支持的工具类型: %s", toolType)
	}
}

// createBuiltinTool 创建内置工具
func (s *AppService) createBuiltinTool(toolConfig map[string]interface{}) (tools.Tool, error) {
	// 这里应该根据内置工具配置创建相应的工具
	// 由于需要集成内置工具管理器，暂时返回nil
	// 实际实现时需要：
	// 1. 获取工具提供者ID和工具ID
	// 2. 调用内置工具管理器创建工具实例
	// 3. 设置工具参数

	return nil, nil
}

// createAPITool 创建API工具
func (s *AppService) createAPITool(toolConfig map[string]interface{}) (tools.Tool, error) {
	// 这里应该根据API工具配置创建相应的工具
	// 由于需要集成API工具管理器，暂时返回nil
	// 实际实现时需要：
	// 1. 获取API工具的配置信息
	// 2. 创建HTTP调用工具
	// 3. 设置请求参数和响应处理

	return nil, nil
}

// validateReviewConfig 验证审核配置
func (s *AppService) validateReviewConfig(ctx context.Context, input string, reviewConfig map[string]interface{}) (bool, string) {
	// 检查审核是否启用
	enable, ok := reviewConfig["enable"].(bool)
	if !ok || !enable {
		return true, "" // 审核未启用，直接通过
	}

	// 检查输入审核
	if inputsConfig, exists := reviewConfig["inputs_config"]; exists {
		if inputsMap, ok := inputsConfig.(map[string]interface{}); ok {
			if inputEnable, ok := inputsMap["enable"].(bool); ok && inputEnable {
				// 执行输入审核
				if keywords, exists := reviewConfig["keywords"]; exists {
					if keywordsList, ok := keywords.([]interface{}); ok {
						for _, keyword := range keywordsList {
							if keywordStr, ok := keyword.(string); ok {
								if strings.Contains(input, keywordStr) {
									// 发现敏感词，返回预设响应
									if presetResponse, exists := inputsMap["preset_response"]; exists {
										if response, ok := presetResponse.(string); ok {
											return false, response
										}
									}
									return false, "您的输入包含敏感内容，请重新输入。"
								}
							}
						}
					}
				}
			}
		}
	}

	return true, ""
}

// validateOutputContent 验证输出内容
func (s *AppService) validateOutputContent(ctx context.Context, output string, reviewConfig map[string]interface{}) bool {
	// 检查审核是否启用
	enable, ok := reviewConfig["enable"].(bool)
	if !ok || !enable {
		return true // 审核未启用，直接通过
	}

	// 检查输出审核
	if outputsConfig, exists := reviewConfig["outputs_config"]; exists {
		if outputsMap, ok := outputsConfig.(map[string]interface{}); ok {
			if outputEnable, ok := outputsMap["enable"].(bool); ok && outputEnable {
				// 执行输出审核
				if keywords, exists := reviewConfig["keywords"]; exists {
					if keywordsList, ok := keywords.([]interface{}); ok {
						for _, keyword := range keywordsList {
							if keywordStr, ok := keyword.(string); ok {
								if strings.Contains(output, keywordStr) {
									return false // 发现敏感词，输出不通过
								}
							}
						}
					}
				}
			}
		}
	}

	return true
}
