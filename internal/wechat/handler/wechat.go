package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/wechat/service"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/response"
)

type WechatHandler struct {
	svc *service.WechatService
}

func NewWechatHandler(svc *service.WechatService) *WechatHandler {
	return &WechatHandler{svc: svc}
}

func (h *WechatHandler) RegisterRoute(r *gin.RouterGroup) {
	wechatGroup := r.Group("wechat")
	{
		wechatGroup.POST("/:app_id", h.Wechat())
	}
}

func (h *WechatHandler) Wechat() gin.HandlerFunc {
	return func(c *gin.Context) {
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage(errors.New("无效的应用ID格式")))
			return
		}

		res, err := h.svc.Wechat(c, appID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, res)
	}
}
