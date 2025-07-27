package service

import (
	"context"
	"errors"
	"github.com/crazyfrankie/voidx/internal/core/agent"
	"github.com/crazyfrankie/voidx/internal/core/memory"
	"gorm.io/gorm"
	"net/http"
	"strings"

	"github.com/crazyfrankie/voidx/pkg/consts"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/officialaccount"
	offConfig "github.com/silenceper/wechat/v2/officialaccount/config"
	"github.com/silenceper/wechat/v2/officialaccount/message"

	"github.com/crazyfrankie/voidx/internal/app_config"
	"github.com/crazyfrankie/voidx/internal/conversation"
	"github.com/crazyfrankie/voidx/internal/llm"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/internal/retriever"
	"github.com/crazyfrankie/voidx/internal/wechat/repository"
)

type WechatService struct {
	wec             *wechat.Wechat
	repo            *repository.WechatRepository
	retrievalSvc    *retriever.Service
	appConfigSvc    *app_config.Service
	conversationSvc *conversation.Service
	llmSvc          *llm.Service
	agentManager    *agent.AgentQueueManager
	tokenBufMem     *memory.TokenBufferMemory
}

func NewWechatService(wec *wechat.Wechat, repo *repository.WechatRepository, retrievalSvc *retriever.Service, appConfigSvc *app_config.Service,
	conversationSvc *conversation.Service, llmSvc *llm.Service, tokenBufMem *memory.TokenBufferMemory,
	agentManager *agent.AgentQueueManager) *WechatService {
	return &WechatService{
		wec:             wec,
		repo:            repo,
		retrievalSvc:    retrievalSvc,
		appConfigSvc:    appConfigSvc,
		conversationSvc: conversationSvc,
		llmSvc:          llmSvc,
		tokenBufMem:     tokenBufMem,
		agentManager:    agentManager,
	}
}

// Wechat 微信公众号校验与消息推送处理
func (s *WechatService) Wechat(c *gin.Context, appID uuid.UUID) (string, error) {
	// 1. 获取应用信息并校验状态
	app, err := s.repo.GetApp(c.Request.Context(), appID)
	if err != nil {
		return "", err
	}

	if app == nil || app.Status != consts.AppStatusPublished {
		if c.Request.Method == http.MethodGet {
			return "", errors.New("该应用未发布或不存在，无法使用，请核实后重试")
		}
		return "该应用未发布或不存在，无法使用，请核实后重试", nil
	}

	// 2. 获取微信配置
	wechatConfig, err := s.repo.GetWechatConfig(c.Request.Context(), appID)
	if wechatConfig == nil || wechatConfig.Status != consts.WechatConfigStatusConfigured {
		if c.Request.Method == http.MethodGet {
			return "", errors.New("该应用未发布到微信公众号，无法使用，请核实后重试")
		}
		return "该应用未发布到微信公众号，无法使用，请核实后重试", nil
	}

	// 3. 创建微信公众号实例
	oa := s.wec.GetOfficialAccount(&offConfig.Config{
		AppID:     wechatConfig.WechatAppID,
		AppSecret: wechatConfig.WechatAppSecret,
		Token:     wechatConfig.WechatToken,
	})

	// 4. 处理GET请求（微信服务器验证）
	if c.Request.Method == http.MethodGet {
		return s.handleVerification(c.Request, oa)
	}

	// 5. 处理POST请求（消息推送）
	return s.handleMessage(c, oa, app)
}

// handleVerification 处理微信服务器验证
func (s *WechatService) handleVerification(req *http.Request, oa *officialaccount.OfficialAccount) (string, error) {
	server := oa.GetServer(req, nil)

	echostr := req.URL.Query().Get("echostr")
	if server.Validate() {
		return echostr, nil
	}

	return "", errors.New("微信公众号服务器配置接入失败")
}

// handleMessage 处理微信消息
func (s *WechatService) handleMessage(c *gin.Context, oa *officialaccount.OfficialAccount, app *entity.App) (string, error) {
	server := oa.GetServer(c.Request, c.Writer)

	// 设置消息处理器
	server.SetMessageHandler(func(mixedMsg *message.MixMessage) *message.Reply {
		// 只支持文本消息
		if mixedMsg.MsgType != message.MsgTypeText {
			return &message.Reply{
				MsgType: message.MsgTypeText,
				MsgData: message.NewText("抱歉，该Agent目前暂时只支持文本消息。"),
			}
		}

		content := mixedMsg.Content
		openID := mixedMsg.FromUserName

		// 获取或创建微信终端用户
		wechatEndUser, err := s.getOrCreateWechatEndUser(c.Request.Context(), string(openID), app.ID, app.AccountID)
		if err != nil {
			return &message.Reply{
				MsgType: message.MsgTypeText,
				MsgData: message.NewText("系统错误，请稍后重试"),
			}
		}

		// 处理特殊消息"1"（获取结果）
		if strings.TrimSpace(content) == "1" {
			reply := s.handleResultQuery(c.Request.Context(), wechatEndUser.ID)
			if reply != "" {
				return &message.Reply{
					MsgType: message.MsgTypeText,
					MsgData: message.NewText(reply),
				}
			}
		}

		// 处理普通消息
		return s.handleNormalMessage(c.Request.Context(), content, wechatEndUser, app)
	})

	// 处理消息
	err := server.Serve()
	if err != nil {
		return "", err
	}

	return "", nil
}

// getOrCreateWechatEndUser 获取或创建微信终端用户
func (s *WechatService) getOrCreateWechatEndUser(ctx context.Context, openID string, appID, accountID uuid.UUID) (*entity.WechatEndUser, error) {
	// 查询现有用户
	wechatEndUser, err := s.repo.GetWechatEndUser(ctx, openID, appID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 如果用户不存在，创建新用户
	if wechatEndUser == nil {
		// 创建终端用户
		endUser := &entity.EndUser{
			ID:       uuid.New(),
			TenantID: accountID,
			AppID:    appID,
		}
		if err := s.repo.CreateEndUser(ctx, endUser); err != nil {
			return nil, err
		}

		// 创建微信终端用户
		wechatEndUser = &entity.WechatEndUser{
			ID:        uuid.New(),
			OpenID:    openID,
			AppID:     appID,
			EndUserID: endUser.ID,
		}
		if err := s.repo.CreateWechatEndUser(ctx, wechatEndUser); err != nil {
			return nil, err
		}
	}

	return wechatEndUser, nil
}
