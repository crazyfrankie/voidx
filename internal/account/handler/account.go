package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/account/service"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/response"
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
		id := c.Request.Context().Value("user_id").(string)
		uid, _ := uuid.Parse(id)

		resp, err := h.svc.GetAccountByID(c.Request.Context(), uid)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, resp)
	}
}

func (h *AccountHandler) UpdatePassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var updateReq req.UpdatePasswdReq
		if err := c.ShouldBind(&updateReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		id := c.Request.Context().Value("user_id").(string)
		uid, _ := uuid.Parse(id)

		err := h.svc.UpdatePassword(c.Request.Context(), uid, updateReq.Password)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}

func (h *AccountHandler) UpdateName() gin.HandlerFunc {
	return func(c *gin.Context) {
		var updateReq req.UpdateNameReq
		if err := c.ShouldBind(&updateReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		id := c.Request.Context().Value("user_id").(string)
		uid, _ := uuid.Parse(id)

		err := h.svc.UpdateName(c.Request.Context(), uid, updateReq.Name)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}

func (h *AccountHandler) UpdateAvatar() gin.HandlerFunc {
	return func(c *gin.Context) {
		var updateReq req.UpdateAvatarReq
		if err := c.ShouldBind(&updateReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		id := c.Request.Context().Value("user_id").(string)
		uid, _ := uuid.Parse(id)

		err := h.svc.UpdateAvatar(c.Request.Context(), uid, updateReq.Avatar)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}
