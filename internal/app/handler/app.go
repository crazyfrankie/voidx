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
		appGroup.POST("/:appID/conversation", h.DebugChat)
	}
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
	res, err := h.appService.DebugChat(c.Request.Context(), appID, chatReq)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithData(c, res)
}
