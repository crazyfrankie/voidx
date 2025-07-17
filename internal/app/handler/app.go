package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/app/service"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/response"
)

type AppHandler struct {
	appService *service.AppService
}

func NewAppHandler(appService *service.AppService) *AppHandler {
	return &AppHandler{
		appService: appService,
	}
}

// RegisterRoute 注册路由
func (h *AppHandler) RegisterRoute(r *gin.RouterGroup) {
	appGroup := r.Group("/apps")
	{
		appGroup.POST("", h.CreateApp)
		appGroup.GET("/:appID", h.GetApp)
		appGroup.PUT("/:appID", h.UpdateApp)
		appGroup.POST("/:appID/copy", h.CopyApp)
		appGroup.DELETE("/:appID", h.DeleteApp)
		appGroup.GET("", h.GetAppsWithPage)

		// 草稿配置相关
		appGroup.GET("/:appID/draft-app-config", h.GetDraftAppConfig)
		appGroup.PUT("/:appID/draft-app-config", h.UpdateDraftAppConfig)

		// 发布相关
		appGroup.POST("/:appID/publish", h.Publish)
		appGroup.POST("/:appID/cancel-publish", h.CancelPublish)
		appGroup.POST("/:appID/fallback-history", h.FallbackHistoryToDraft)
		appGroup.GET("/:appID/publish-histories", h.GetPublishHistoriesWithPage)

		// 调试会话相关
		appGroup.GET("/:appID/summary", h.GetDebugConversationSummary)
		appGroup.PUT("/:appID/summary", h.UpdateDebugConversationSummary)
		appGroup.DELETE("/:appID/debug-conversation", h.DeleteDebugConversation)
		appGroup.POST("/:appID/conversation", h.DebugChat)
		appGroup.POST("/:appID/conversation/tasks/:taskID/stop", h.StopDebugChat)
		appGroup.GET("/:appID/conversation/messages", h.GetDebugConversationMessagesWithPage)

		// 已发布配置相关
		appGroup.GET("/:appID/published-config", h.GetPublishedConfig)
		appGroup.POST("/:appID/published-config/regenerate-web-app-token", h.RegenerateWebAppToken)
	}
}

// CreateApp 创建应用
func (h *AppHandler) CreateApp(c *gin.Context) {
	var createReq req.CreateAppReq
	if err := c.ShouldBindJSON(&createReq); err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
		return
	}

	appID, err := h.appService.CreateApp(c.Request.Context(), createReq)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, gin.H{"id": appID})
}

// GetApp 获取应用
func (h *AppHandler) GetApp(c *gin.Context) {
	appIDStr := c.Param("appID")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("应用ID格式不正确"))
		return
	}

	app, err := h.appService.GetApp(c.Request.Context(), appID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, app)
}

// UpdateApp 更新应用
func (h *AppHandler) UpdateApp(c *gin.Context) {
	appIDStr := c.Param("appID")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("应用ID格式不正确"))
		return
	}

	var updateReq req.UpdateAppReq
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
		return
	}

	if err := h.appService.UpdateApp(c.Request.Context(), appID, updateReq); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}

// CopyApp 拷贝应用
func (h *AppHandler) CopyApp(c *gin.Context) {
	appIDStr := c.Param("appID")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("应用ID格式不正确"))
		return
	}

	newAppID, err := h.appService.CopyApp(c.Request.Context(), appID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, gin.H{"id": newAppID})
}

// DeleteApp 删除应用
func (h *AppHandler) DeleteApp(c *gin.Context) {
	appIDStr := c.Param("appID")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("应用ID格式不正确"))
		return
	}

	if err := h.appService.DeleteApp(c.Request.Context(), appID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}

// GetAppsWithPage 获取应用分页列表
func (h *AppHandler) GetAppsWithPage(c *gin.Context) {
	var pageReq req.GetAppsWithPageReq
	if err := c.ShouldBindQuery(&pageReq); err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
		return
	}

	apps, paginator, err := h.appService.GetAppsWithPage(c.Request.Context(), pageReq)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, gin.H{
		"list":      apps,
		"paginator": paginator,
	})
}

// GetDraftAppConfig 获取应用的最新草稿配置
func (h *AppHandler) GetDraftAppConfig(c *gin.Context) {
	appIDStr := c.Param("appID")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("应用ID格式不正确"))
		return
	}

	draftConfig, err := h.appService.GetDraftAppConfig(c.Request.Context(), appID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, draftConfig)
}

// UpdateDraftAppConfig 更新应用的最新草稿配置
func (h *AppHandler) UpdateDraftAppConfig(c *gin.Context) {
	appIDStr := c.Param("appID")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("应用ID格式不正确"))
		return
	}

	var draftConfig map[string]interface{}
	if err := c.ShouldBindJSON(&draftConfig); err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
		return
	}

	if err := h.appService.UpdateDraftAppConfig(c.Request.Context(), appID, draftConfig); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}

// Publish 发布/更新特定的草稿配置信息
func (h *AppHandler) Publish(c *gin.Context) {
	appIDStr := c.Param("appID")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("应用ID格式不正确"))
		return
	}

	if err := h.appService.PublishDraftAppConfig(c.Request.Context(), appID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}

// CancelPublish 取消发布指定的应用配置信息
func (h *AppHandler) CancelPublish(c *gin.Context) {
	appIDStr := c.Param("appID")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("应用ID格式不正确"))
		return
	}

	if err := h.appService.CancelPublishAppConfig(c.Request.Context(), appID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}

// FallbackHistoryToDraft 退回指定版本到草稿中
func (h *AppHandler) FallbackHistoryToDraft(c *gin.Context) {
	appIDStr := c.Param("appID")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("应用ID格式不正确"))
		return
	}

	var fallbackReq req.FallbackHistoryToDraftReq
	if err := c.ShouldBindJSON(&fallbackReq); err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
		return
	}

	if err := h.appService.FallbackHistoryToDraft(c.Request.Context(), appID, fallbackReq.AppConfigVersionID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}

// GetPublishHistoriesWithPage 获取应用发布历史列表
func (h *AppHandler) GetPublishHistoriesWithPage(c *gin.Context) {
	appIDStr := c.Param("appID")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("应用ID格式不正确"))
		return
	}

	var pageReq req.GetPublishHistoriesWithPageReq
	if err := c.ShouldBindQuery(&pageReq); err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
		return
	}

	histories, paginator, err := h.appService.GetPublishHistoriesWithPage(c.Request.Context(), appID, pageReq)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, gin.H{
		"list":      histories,
		"paginator": paginator,
	})
}

// GetDebugConversationSummary 获取调试会话长期记忆
func (h *AppHandler) GetDebugConversationSummary(c *gin.Context) {
	appIDStr := c.Param("appID")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("应用ID格式不正确"))
		return
	}

	summary, err := h.appService.GetDebugConversationSummary(c.Request.Context(), appID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, gin.H{"summary": summary})
}

// UpdateDebugConversationSummary 更新调试会话长期记忆
func (h *AppHandler) UpdateDebugConversationSummary(c *gin.Context) {
	appIDStr := c.Param("appID")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("应用ID格式不正确"))
		return
	}

	var summaryReq req.UpdateDebugConversationSummaryReq
	if err := c.ShouldBindJSON(&summaryReq); err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
		return
	}

	if err := h.appService.UpdateDebugConversationSummary(c.Request.Context(), appID, summaryReq.Summary); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}

// DeleteDebugConversation 清空该应用的调试会话记录
func (h *AppHandler) DeleteDebugConversation(c *gin.Context) {
	appIDStr := c.Param("appID")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("应用ID格式不正确"))
		return
	}

	if err := h.appService.DeleteDebugConversation(c.Request.Context(), appID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}

// DebugChat 发起调试对话
func (h *AppHandler) DebugChat(c *gin.Context) {
	appIDStr := c.Param("appID")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("应用ID格式不正确"))
		return
	}

	var chatReq req.DebugChatReq
	if err := c.ShouldBindJSON(&chatReq); err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
		return
	}

	// 调用服务发起调试对话，并将结果流式返回
	if err := h.appService.DebugChat(c.Request.Context(), appID, chatReq, c.Writer); err != nil {
		// 错误处理已经在流式响应中完成
		return
	}
}

// StopDebugChat 停止某个应用的指定调试会话
func (h *AppHandler) StopDebugChat(c *gin.Context) {
	appIDStr := c.Param("appID")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("应用ID格式不正确"))
		return
	}

	taskIDStr := c.Param("taskID")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("任务ID格式不正确"))
		return
	}

	if err := h.appService.StopDebugChat(c.Request.Context(), appID, taskID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c)
}

// GetDebugConversationMessagesWithPage 获取该应用的调试会话分页列表记录
func (h *AppHandler) GetDebugConversationMessagesWithPage(c *gin.Context) {
	appIDStr := c.Param("appID")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("应用ID格式不正确"))
		return
	}

	var pageReq req.GetDebugConversationMessagesWithPageReq
	if err := c.ShouldBindQuery(&pageReq); err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
		return
	}

	messages, paginator, err := h.appService.GetDebugConversationMessagesWithPage(c.Request.Context(), appID, pageReq)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, gin.H{
		"list":      messages,
		"paginator": paginator,
	})
}

// GetPublishedConfig 获取已发布的配置
func (h *AppHandler) GetPublishedConfig(c *gin.Context) {
	appIDStr := c.Param("appID")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("应用ID格式不正确"))
		return
	}
	publishedConfig, err := h.appService.GetPublishedConfig(c.Request.Context(), appID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessWithData(c, publishedConfig)
}

// RegenerateWebAppToken 重新生成WebApp令牌
func (h *AppHandler) RegenerateWebAppToken(c *gin.Context) {
	appIDStr := c.Param("appID")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		response.Error(c, errno.ErrValidate.AppendBizMessage("应用ID格式不正确"))
		return
	}

	token, err := h.appService.RegenerateWebAppToken(c.Request.Context(), appID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.SuccessWithData(c, gin.H{"token": token})
}
