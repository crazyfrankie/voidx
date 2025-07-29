package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/app/service"
	"github.com/crazyfrankie/voidx/internal/app_config"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/response"
	"github.com/crazyfrankie/voidx/pkg/util"
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
		appGroup.POST("", h.CreateApp())                                //
		appGroup.GET("", h.GetAppsWithPage())                           //
		appGroup.GET("/:app_id", h.GetApp())                            //
		appGroup.PUT("/:app_id", h.UpdateApp())                         //
		appGroup.POST("/:app_id", h.CopyApp())                          //
		appGroup.DELETE("/:app_id", h.DeleteApp())                      //
		appGroup.GET("/:app_id/draft-config", h.GetAppDraftConfig())    //
		appGroup.PUT("/:app_id/draft-config", h.UpdateAppDraftConfig()) //
		appGroup.POST("/:app_id/publish", h.PublishApp())               //
		appGroup.POST("/:app_id/unpublish", h.UnpublishApp())           //
		appGroup.POST("/:app_id/summary", h.UpdateAppSummary())
		appGroup.GET("/:app_id/published-config", h.GetPublishedConfig())
		appGroup.POST("/:app_id/conversations", h.DebugChat())
	}
}

// CreateApp 创建应用
func (h *AppHandler) CreateApp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var createReq req.CreateAppReq
		if err := c.ShouldBindJSON(&createReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		app, err := h.appService.CreateApp(c.Request.Context(), userID, createReq)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, map[string]any{"id": app.ID})
	}
}

// GetApp 获取应用详情
func (h *AppHandler) GetApp() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("无效的应用ID格式"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		app, err := h.appService.GetApp(c.Request.Context(), appID, userID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, map[string]any{
			"id":          app.ID,
			"name":        app.Name,
			"icon":        app.Icon,
			"description": app.Description,
			"status":      app.Status,
			"token":       app.Token,
			"updated_at":  app.Utime,
			"created_at":  app.Ctime,
		})
	}
}

// UpdateApp 更新应用基本信息
func (h *AppHandler) UpdateApp() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("无效的应用ID格式"))
			return
		}

		var updateReq req.UpdateAppReq
		if err := c.ShouldBindJSON(&updateReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		err = h.appService.UpdateApp(c.Request.Context(), appID, userID, updateReq)
		if err != nil {
			response.Error(c, err)
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
			response.Error(c, errno.ErrValidate.AppendBizMessage("无效的应用ID格式"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		err = h.appService.CopyApp(c.Request.Context(), appID, userID)
		if err != nil {
			response.Error(c, err)
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
			response.Error(c, errno.ErrValidate.AppendBizMessage("无效的应用ID格式"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		err = h.appService.DeleteApp(c.Request.Context(), appID, userID)
		if err != nil {
			response.Error(c, err)
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
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
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
			response.Error(c, err)
			return
		}

		res, err := h.appService.GetAppsWithPage(c.Request.Context(), userID, pageReq)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, res)
	}
}

// GetAppDraftConfig 获取应用草稿配置
func (h *AppHandler) GetAppDraftConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("无效的应用ID格式"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		appConfig, err := h.appService.GetDraftAppConfig(c.Request.Context(), appID, userID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, appConfig)
	}
}

// UpdateAppDraftConfig 更新应用草稿配置
func (h *AppHandler) UpdateAppDraftConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("无效的应用ID格式"))
			return
		}

		updateReq := make(map[string]any)
		if err := c.ShouldBindJSON(&updateReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		err = h.appService.UpdateDraftAppConfig(c.Request.Context(), appID, userID, updateReq)
		if err != nil {
			response.Error(c, err)
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
			response.Error(c, errno.ErrValidate.AppendBizMessage("无效的应用ID格式"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		err = h.appService.PublishDraftAppConfig(c.Request.Context(), appID, userID)
		if err != nil {
			response.Error(c, err)
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
			response.Error(c, errno.ErrValidate.AppendBizMessage("无效的应用ID格式"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		err = h.appService.CancelPublishAppConfig(c.Request.Context(), appID, userID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}

// UpdateAppSummary 更新应用长记忆
func (h *AppHandler) UpdateAppSummary() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("无效的应用ID格式"))
			return
		}

		var summaryReq req.UpdateAppSummaryReq
		if err := c.ShouldBindJSON(&summaryReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		err = h.appService.UpdateDebugConversationSummary(c.Request.Context(), appID, userID, summaryReq.Summary)
		if err != nil {
			response.Error(c, err)
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
			response.Error(c, errno.ErrValidate.AppendBizMessage("无效的应用ID格式"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		res, err := h.appService.GetPublishedConfig(c.Request.Context(), appID, userID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, res)
	}
}

// DebugChat 发起调试对话
func (h *AppHandler) DebugChat() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.Param("app_id")
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

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		// 调用服务发起调试对话，并将结果流式返回
		_, err = h.appService.DebugChat(c.Request.Context(), appID, userID, chatReq)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}
