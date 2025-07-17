package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/errno"
)

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

	// 模拟流式响应
	go func() {
		defer func() {
			// 发送结束事件
			sendSSEEvent("message_end", map[string]interface{}{
				"id":              taskID.String(),
				"conversation_id": conversation.ID.String(),
				"message_id":      createdMessage.ID.String(),
				"task_id":         taskID.String(),
			})
		}()

		// 获取模型配置
		modelConfig, ok := draftConfig["model_config"].(map[string]interface{})
		if !ok {
			sendSSEEvent("error", map[string]interface{}{
				"error": "模型配置不正确",
			})
			return
		}

		// 调用LLM生成响应
		answer, err := s.generateResponse(ctx, req.Query, modelConfig)
		if err != nil {
			sendSSEEvent("error", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		// 分块发送答案
		chunkSize := 10
		for i := 0; i < len(answer); i += chunkSize {
			end := i + chunkSize
			if end > len(answer) {
				end = len(answer)
			}

			chunk := answer[i:end]
			sendSSEEvent("message", map[string]interface{}{
				"id":                taskID.String(),
				"conversation_id":   conversation.ID.String(),
				"message_id":        createdMessage.ID.String(),
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
			time.Sleep(100 * time.Millisecond)
		}

		// 更新消息状态
		createdMessage.Answer = answer
		createdMessage.Status = "normal"
		s.repo.UpdateMessage(ctx, createdMessage)
	}()

	return nil
}

// StopDebugChat 停止调试对话
func (s *AppService) StopDebugChat(ctx context.Context, appID uuid.UUID, taskID uuid.UUID) error {
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

	// 这里可以实现停止任务的逻辑
	// 由于是简化实现，我们只是记录停止请求
	fmt.Printf("停止调试对话任务: %s\n", taskID.String())

	return nil
}

// GetDebugConversationMessagesWithPage 获取调试会话消息分页列表
func (s *AppService) GetDebugConversationMessagesWithPage(ctx context.Context, appID uuid.UUID, pageReq req.GetDebugConversationMessagesWithPageReq) ([]*resp.MessageResp, *resp.Paginator, error) {
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
		return []*resp.MessageResp{}, &resp.Paginator{
			CurrentPage: pageReq.Page,
			PageSize:    pageReq.PageSize,
			TotalPage:   0,
			TotalRecord: 0,
		}, nil
	}

	// 获取消息分页列表
	messages, total, err := s.repo.GetDebugConversationMessagesWithPage(ctx, *app.DebugConversationID, pageReq.Page, pageReq.PageSize, pageReq.Ctime)
	if err != nil {
		return nil, nil, err
	}

	// 转换为响应
	messageResps := make([]*resp.MessageResp, 0, len(messages))
	for _, message := range messages {
		// 转换智能体思考过程
		agentThoughts := make([]resp.AgentThought, 0, len(message.AgentThoughts))
		for _, agentThought := range message.AgentThoughts {
			agentThoughts = append(agentThoughts, resp.AgentThought{
				ID:              agentThought.ID,
				MessageID:       agentThought.MessageID,
				Event:           agentThought.Event,
				Thought:         agentThought.Thought,
				Observation:     agentThought.Observation,
				Tool:            agentThought.Tool,
				ToolInput:       agentThought.ToolInput,
				Answer:          agentThought.Answer,
				TotalTokenCount: agentThought.TotalTokenCount,
				TotalPrice:      agentThought.TotalPrice,
				Latency:         agentThought.Latency,
				Ctime:           agentThought.Ctime,
			})
		}

		messageResps = append(messageResps, &resp.MessageResp{
			ID:             message.ID,
			ConversationID: message.ConversationID,
			AppID:          message.AppID,
			InvokeFrom:     message.InvokeFrom,
			CreatedBy:      message.CreatedBy,
			Query:          message.Query,
			ImageUrls:      message.ImageUrls,
			Answer:         message.Answer,
			Status:         message.Status,
			AgentThoughts:  agentThoughts,
			Ctime:          message.Ctime,
			Utime:          message.Utime,
		})
	}

	// 构建分页器
	paginator := &resp.Paginator{
		CurrentPage: pageReq.Page,
		PageSize:    pageReq.PageSize,
		TotalPage:   int(total) / pageReq.PageSize,
		TotalRecord: int(total),
	}
	if int(total)%pageReq.PageSize > 0 {
		paginator.TotalPage++
	}

	return messageResps, paginator, nil
}

// generateResponse 生成响应（简化实现）
func (s *AppService) generateResponse(ctx context.Context, query string, modelConfig map[string]interface{}) (string, error) {
	// 这里应该调用真正的LLM模型
	// 为了简化，我们返回一个模拟的响应

	// 在真实实现中，你需要：
	// 1. 根据modelConfig配置LLM
	// 2. 准备历史对话记录
	// 3. 调用LLM生成响应
	// 4. 处理工具调用
	// 5. 处理知识库检索

	return fmt.Sprintf("这是对问题 '%s' 的回答。这是一个模拟的响应，用于演示SSE流式输出功能。", query), nil
}
