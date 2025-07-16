package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

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

	// 检查草稿配置ID是否存在
	if app.DraftAppConfigID == nil {
		return "", errno.ErrNotFound.AppendBizMessage("应用草稿配置不存在")
	}

	// 获取草稿配置
	draftAppConfigVersion, err := s.repo.GetAppConfigVersion(ctx, *app.DraftAppConfigID)
	if err != nil {
		return "", err
	}

	// 检查长期记忆是否启用
	longTermMemory, ok := draftAppConfigVersion.Config["long_term_memory"].(map[string]interface{})
	if !ok {
		return "", errno.ErrValidate.AppendBizMessage("长期记忆配置格式错误")
	}

	enable, ok := longTermMemory["enable"].(bool)
	if !ok || !enable {
		return "", errno.ErrValidate.AppendBizMessage("该应用并未开启长期记忆")
	}

	// 检查调试会话ID是否存在
	if app.DebugConversationID == nil {
		return "", errors.New("会话不存在")
	}

	// 获取调试会话
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

	// 检查草稿配置ID是否存在
	if app.DraftAppConfigID == nil {
		return errno.ErrNotFound.AppendBizMessage("应用草稿配置不存在")
	}

	// 获取草稿配置
	draftAppConfigVersion, err := s.repo.GetAppConfigVersion(ctx, *app.DraftAppConfigID)
	if err != nil {
		return err
	}

	// 检查长期记忆是否启用
	longTermMemory, ok := draftAppConfigVersion.Config["long_term_memory"].(map[string]interface{})
	if !ok {
		return errno.ErrValidate.AppendBizMessage("长期记忆配置格式错误")
	}

	enable, ok := longTermMemory["enable"].(bool)
	if !ok || !enable {
		return errno.ErrValidate.AppendBizMessage("该应用并未开启长期记忆")
	}

	// 检查调试会话ID是否存在
	if app.DebugConversationID == nil {
		// 创建调试会话
		conversation, err := s.repo.CreateConversation(ctx, accountID, appID)
		if err != nil {
			return err
		}

		// 更新应用的调试会话ID
		app.DebugConversationID = &conversation.ID
		if err := s.repo.UpdateApp(ctx, app); err != nil {
			return err
		}

		// 更新调试会话的摘要
		conversation.Summary = summary
		if err := s.repo.UpdateConversation(ctx, conversation); err != nil {
			return err
		}
	} else {
		// 获取调试会话
		conversation, err := s.repo.GetConversation(ctx, *app.DebugConversationID)
		if err != nil {
			return err
		}

		// 更新调试会话的摘要
		conversation.Summary = summary
		if err := s.repo.UpdateConversation(ctx, conversation); err != nil {
			return err
		}
	}

	return nil
}

// DeleteDebugConversation 清空该应用的调试会话记录
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
func (s *AppService) DebugChat(ctx context.Context, appID uuid.UUID, req interface{}, w http.ResponseWriter) error {
	// 这里应该实现完整的调试对话逻辑
	// 为了简化，我们只返回一个简单的响应

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

	// 简单实现，实际项目中应该使用流式响应
	w.Write([]byte("event: agent_message\ndata:{\"id\":\"1\",\"event\":\"agent_message\",\"answer\":\"这是一个测试回复\",\"thought\":\"思考中...\",\"observation\":\"\",\"tool\":\"\",\"tool_input\":\"\",\"total_token_count\":10,\"total_price\":0.0001,\"latency\":0.5,\"conversation_id\":\"1\",\"message_id\":\"1\",\"task_id\":\"1\"}\n\n"))
	w.Write([]byte("event: end\ndata:{\"id\":\"2\",\"event\":\"end\",\"answer\":\"这是一个测试回复\",\"thought\":\"思考完毕\",\"observation\":\"\",\"tool\":\"\",\"tool_input\":\"\",\"total_token_count\":10,\"total_price\":0.0001,\"latency\":0.5,\"conversation_id\":\"1\",\"message_id\":\"1\",\"task_id\":\"1\"}\n\n"))

	return nil
}

// StopDebugChat 停止某个应用的指定调试会话
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

	// 实际项目中应该实现停止调试会话的逻辑
	// 这里只是一个简单的实现

	return nil
}
