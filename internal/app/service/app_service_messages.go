package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/errno"
)

// GetDebugConversationMessagesWithPage 获取该应用的调试会话分页列表记录
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

	// 检查调试会话ID是否存在
	if app.DebugConversationID == nil {
		return []*resp.MessageResp{}, &resp.Paginator{
			CurrentPage: pageReq.Page,
			PageSize:    pageReq.PageSize,
			TotalPage:   0,
			TotalRecord: 0,
		}, nil
	}

	// 获取调试会话消息分页列表
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
