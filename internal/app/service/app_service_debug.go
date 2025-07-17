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
			"event":             "message",
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

		// 构建完整的提示消息
		promptMessage, err := s.buildPromptMessage(taskCtx, req.Query, draftConfig, conversation)
		if err != nil {
			sendSSEEvent("error", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// 调用LLM生成响应
		answer, err := s.generateResponse(taskCtx, promptMessage, modelConfig)
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

		// 流式发送答案
		s.streamAnswer(taskCtx, answer, taskID, conversation.ID, createdMessage.ID, sendSSEEvent)

		// 更新消息状态
		createdMessage.Answer = answer
		createdMessage.Status = "normal"
		s.repo.UpdateMessage(taskCtx, createdMessage)
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

// generateResponse 生成响应
func (s *AppService) generateResponse(ctx context.Context, prompt string, modelConfig map[string]interface{}) (string, error) {
	// 检查上下文是否被取消
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// 获取模型配置
	modelName, ok := modelConfig["model"].(string)
	if !ok {
		return "", errno.ErrValidate.AppendBizMessage("模型名称配置不正确")
	}

	// 获取模型参数
	temperature := 0.7
	if temp, ok := modelConfig["temperature"].(float64); ok {
		temperature = temp
	}

	maxTokens := 2048
	if m, ok := modelConfig["max_tokens"].(float64); ok {
		maxTokens = int(m)
	}

	// 模拟调用LLM生成响应
	// 这里应该替换为实际的LLM调用，比如调用OpenAI、Azure OpenAI等
	// 示例中我们模拟一个简单的回答

	// 检查是否被取消
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// 模拟LLM处理时间
	time.Sleep(1 * time.Second)

	// 根据模型和提示生成简单的回答
	answer := fmt.Sprintf("这是使用模型 %s (温度: %.2f, 最大token: %d) 对您的问题的回答。\n\n您的问题：%s\n\n我理解您的问题，并会尽力为您提供帮助。",
		modelName, temperature, maxTokens, prompt)

	return answer, nil
}
