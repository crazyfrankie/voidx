package handler

import (
	"github.com/crazyfrankie/voidx/internal/base/response"
	"github.com/crazyfrankie/voidx/types/errno"
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/voidx/internal/account/service"
	"github.com/crazyfrankie/voidx/internal/models/req"
)

type AccountHandler struct {
	svc *service.AccountService
}

func NewAccountHandler(svc *service.AccountService) *AccountHandler {
	return &AccountHandler{svc: svc}
}

func (h *AccountHandler) RegisterRoute(r *gin.RouterGroup) {
	accountGroup := r.Group("account")
	{
		accountGroup.GET("", h.GetAccount())
		accountGroup.PUT("password", h.UpdatePassword())
		accountGroup.PUT("name", h.UpdateName())
		accountGroup.PUT("avatar", h.UpdateAvatar())
	}
}

func (h *AccountHandler) GetAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		resp, err := h.svc.GetAccountByID(c.Request.Context())
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Data(c, resp)
	}
}

func (h *AccountHandler) UpdatePassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var updateReq req.UpdatePasswdReq
		if err := c.ShouldBind(&updateReq); err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		err := h.svc.UpdatePassword(c.Request.Context(), updateReq.Password)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Success(c)
	}
}

func (h *AccountHandler) UpdateName() gin.HandlerFunc {
	return func(c *gin.Context) {
		var updateReq req.UpdateNameReq
		if err := c.ShouldBind(&updateReq); err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		err := h.svc.UpdateName(c.Request.Context(), updateReq.Name)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Success(c)
	}
}

func (h *AccountHandler) UpdateAvatar() gin.HandlerFunc {
	return func(c *gin.Context) {
		var updateReq req.UpdateAvatarReq
		if err := c.ShouldBind(&updateReq); err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		err := h.svc.UpdateAvatar(c.Request.Context(), updateReq.Avatar)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Success(c)
	}
}
