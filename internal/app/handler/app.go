package handler

import (
	"io"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/app/service"
	"github.com/crazyfrankie/voidx/internal/app_config"
	"github.com/crazyfrankie/voidx/internal/base/response"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/pkg/util"
	"github.com/crazyfrankie/voidx/types/errno"
)

type AppHandler struct {
	appService *service.AppService
	appCfgSvc  *app_config.Service
}

func NewAppHandler(appService *service.AppService, appCfgSvc *app_config.Service) *AppHandler {
	return &AppHandler{
		appService: appService,
		appCfgSvc:  appCfgSvc,
	}
}

// RegisterRoute 注册路由
func (h *AppHandler) RegisterRoute(r *gin.RouterGroup) {
	appGroup := r.Group("/apps")
	{
		appGroup.GET("/:app_id", h.GetApp())
		appGroup.POST("", h.CreateApp())
		appGroup.PUT("/:app_id", h.UpdateApp())
		appGroup.DELETE("/:app_id", h.DeleteApp())
		appGroup.POST("/:app_id", h.CopyApp())
		appGroup.GET("", h.GetAppsWithPage())
		appGroup.GET("/:app_id/draft-app-config", h.GetAppDraftConfig())
		appGroup.PUT("/:app_id/draft-app-config", h.UpdateAppDraftConfig())
		appGroup.GET("/:app_id/summary", h.GetDebugAppSummary())
		appGroup.PUT("/:app_id/summary", h.UpdateDebugAppSummary())
		appGroup.POST("/:app_id/conversation", h.DebugChat())
		appGroup.POST("/:app_id/conversation/tasks/:task_id/stop", h.StopDebugChat())
		appGroup.GET("/:app_id/conversation/messages", h.GetDebugConversationWithPage())
		appGroup.DELETE("/:app_id/debug-conversation")
		appGroup.POST("/:app_id/publish", h.PublishApp())
		appGroup.POST("/:app_id/unpublish", h.UnpublishApp())
		appGroup.GET("/:app_id/publish-histories", h.GetPublishedHistoryWithPage())
		appGroup.GET("/:app_id/fallback-history", h.FallBackHistory())
		appGroup.GET("/:app_id/published-config", h.GetPublishedConfig())
		appGroup.POST("/:app_id/published-config/regenerate-web-app-token", h.RegenerateToken())
	}
}

// CreateApp 创建应用
func (h *AppHandler) CreateApp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var createReq req.CreateAppReq
		if err := c.ShouldBindJSON(&createReq); err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		app, err := h.appService.CreateApp(c.Request.Context(), userID, &createReq)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Data(c, map[string]any{"id": app.ID})
	}
}

// GetApp 获取应用详情
func (h *AppHandler) GetApp() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		app, err := h.appService.GetApp(c.Request.Context(), appID, userID)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Data(c, app)
	}
}

// UpdateApp 更新应用基本信息
func (h *AppHandler) UpdateApp() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		var updateReq req.UpdateAppReq
		if err := c.ShouldBindJSON(&updateReq); err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		err = h.appService.UpdateApp(c.Request.Context(), appID, userID, &updateReq)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Success(c)
	}
}

// CopyApp 复制应用
func (h *AppHandler) CopyApp() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		err = h.appService.CopyApp(c.Request.Context(), appID, userID)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Success(c)
	}
}

// DeleteApp 删除应用
func (h *AppHandler) DeleteApp() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		err = h.appService.DeleteApp(c.Request.Context(), appID, userID)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Success(c)
	}
}

// GetAppsWithPage 获取应用分页列表
func (h *AppHandler) GetAppsWithPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var pageReq req.GetAppsWithPageReq
		if err := c.ShouldBindQuery(&pageReq); err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		// 设置默认值
		if pageReq.CurrentPage == 0 {
			pageReq.CurrentPage = 1
		}
		if pageReq.PageSize == 0 {
			pageReq.PageSize = 20
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		res, err := h.appService.GetAppsWithPage(c.Request.Context(), userID, pageReq)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Data(c, res)
	}
}

// GetAppDraftConfig 获取应用草稿配置
func (h *AppHandler) GetAppDraftConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		appConfig, err := h.appService.GetDraftAppConfig(c.Request.Context(), appID, userID)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Data(c, appConfig)
	}
}

// UpdateAppDraftConfig 更新应用草稿配置
func (h *AppHandler) UpdateAppDraftConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		updateReq := make(map[string]any)
		if err := c.ShouldBindJSON(&updateReq); err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		err = h.appService.UpdateDraftAppConfig(c.Request.Context(), appID, userID, updateReq)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Success(c)
	}
}

// PublishApp 发布应用
func (h *AppHandler) PublishApp() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		err = h.appService.PublishDraftAppConfig(c.Request.Context(), appID, userID)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Success(c)
	}
}

// UnpublishApp 取消发布应用
func (h *AppHandler) UnpublishApp() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		err = h.appService.CancelPublishAppConfig(c.Request.Context(), appID, userID)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Success(c)
	}
}

// UpdateDebugAppSummary 更新应用长记忆
func (h *AppHandler) UpdateDebugAppSummary() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		var summaryReq req.UpdateAppSummaryReq
		if err := c.ShouldBindJSON(&summaryReq); err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		err = h.appService.UpdateDebugConversationSummary(c.Request.Context(), appID, userID, summaryReq.Summary)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Success(c)
	}
}

// GetPublishedConfig 获取已发布应用的配置
func (h *AppHandler) GetPublishedConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		res, err := h.appService.GetPublishedConfig(c.Request.Context(), appID, userID)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Data(c, res)
	}
}

// DebugChat 发起调试对话
func (h *AppHandler) DebugChat() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		var chatReq req.DebugChatReq
		if err := c.ShouldBindJSON(&chatReq); err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		// 调用服务发起调试对话，并将结果流式返回
		res, err := h.appService.DebugChat(c.Request.Context(), appID, userID, chatReq)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		// 流式输出
		c.Stream(func(w io.Writer) bool {
			select {
			case resp, ok := <-res:
				if !ok {
					return false
				}
				c.SSEvent("message", resp)
				return true
			case <-c.Request.Context().Done():
				return false
			}
		})
	}
}

// GetDebugAppSummary 获取应用调试长记忆
func (h *AppHandler) GetDebugAppSummary() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		res, err := h.appService.GetDebugConversationSummary(c.Request.Context(), appID, userID)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Data(c, gin.H{"summary": res})
	}
}

// StopDebugChat 停止某次应用的调试会话
func (h *AppHandler) StopDebugChat() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		taskIdStr := c.Param("task_id")
		taskID, err := uuid.Parse(taskIdStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		err = h.appService.StopDebugChat(c.Request.Context(), appID, taskID, userID)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Success(c)
	}
}

// GetDebugConversationWithPage 获取应用的调试会话消息列表
func (h *AppHandler) GetDebugConversationWithPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var pageReq req.GetDebugConversationMessagesWithPageReq
		if err := c.ShouldBindQuery(&pageReq); err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrUnauthorized)
			return
		}

		list, paginator, err := h.appService.GetDebugConversationMessagesWithPage(c.Request.Context(), appID, pageReq, userID)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Data(c, gin.H{"list": list, "paginator": paginator})
	}
}

// DeleteDebugConversation 清空应用的调试会话记录
func (h *AppHandler) DeleteDebugConversation() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		err = h.appService.DeleteDebugConversation(c.Request.Context(), appID, userID)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Success(c)
	}
}

// GetPublishedHistoryWithPage 获取应用的发布历史列表信息
func (h *AppHandler) GetPublishedHistoryWithPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var pageReq req.GetPublishHistoriesWithPageReq
		if err := c.ShouldBindQuery(&pageReq); err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, errno.ErrUnauthorized)
			return
		}

		list, paginator, err := h.appService.GetPublishHistoriesWithPage(c.Request.Context(), appID, pageReq, userID)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Data(c, gin.H{"list": list, "paginator": paginator})
	}
}

// FallBackHistory 回退指定的历史配置到草稿
func (h *AppHandler) FallBackHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		var fallbackReq req.FallbackHistoryToDraftReq
		if err := c.ShouldBind(&fallbackReq); err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		_, err = h.appService.FallbackHistoryToDraft(c.Request.Context(), appID, fallbackReq.AppConfigVersionID, userID)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}
	}
}

// RegenerateToken 重新生成 WebApp 的凭证标识
func (h *AppHandler) RegenerateToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		res, err := h.appService.RegenerateWebAppToken(c.Request.Context(), appID, userID)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Data(c, gin.H{"token": res})
	}
}
