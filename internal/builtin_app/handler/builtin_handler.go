package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/voidx/internal/base/response"
	"github.com/crazyfrankie/voidx/internal/builtin_app/service"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/pkg/util"
	"github.com/crazyfrankie/voidx/types/errno"
)

type BuiltinAppHandler struct {
	svc *service.BuiltinService
}

func NewBuiltinAppHandler(svc *service.BuiltinService) *BuiltinAppHandler {
	return &BuiltinAppHandler{svc: svc}
}

func (h *BuiltinAppHandler) RegisterRoute(r *gin.RouterGroup) {
	builtinAppGroup := r.Group("builtin-apps")
	{
		builtinAppGroup.GET("categories", h.GetBuiltinAppCategories())
		builtinAppGroup.GET("", h.GetBuiltinApps())
		builtinAppGroup.POST("add-builtin-app-to-space", h.AddBuiltinAppToSpace())
	}
}

func (h *BuiltinAppHandler) GetBuiltinAppCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		res := h.svc.GetBuiltinAppCategories(c.Request.Context())

		response.Data(c, res)
	}
}

func (h *BuiltinAppHandler) GetBuiltinApps() gin.HandlerFunc {
	return func(c *gin.Context) {
		res := h.svc.GetBuiltinApps(c.Request.Context())

		response.Data(c, res)
	}
}

func (h *BuiltinAppHandler) AddBuiltinAppToSpace() gin.HandlerFunc {
	return func(c *gin.Context) {
		var addReq req.AddBuiltinAppReq
		if err := c.ShouldBind(&addReq); err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		app, err := h.svc.AddBuiltinAppToSpace(c.Request.Context(), userID, addReq)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Data(c, gin.H{"id": app.ID})
	}
}
