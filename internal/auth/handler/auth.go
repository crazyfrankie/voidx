package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/voidx/internal/auth/service"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/response"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) RegisterRoute(r *gin.RouterGroup) {
	authGroup := r.Group("auth")
	{
		authGroup.POST("login", h.Login())
		authGroup.GET("logout", h.Logout())
	}
}

func (h *AuthHandler) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginReq req.LoginReq
		if err := c.ShouldBind(&loginReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		tokens, err := h.svc.Login(c.Request.Context(), c.Request.UserAgent(), loginReq)
		if err != nil {
			response.Error(c, err)
			return
		}

		c.SetSameSite(http.SameSiteLaxMode)
		c.Header("x-access-token", tokens[0])
		c.SetCookie("llmops_refresh", tokens[1], int(time.Hour*24), "/", "", false, true)

		response.Success(c)
	}
}

func (h *AuthHandler) Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := h.svc.Logout(c.Request.Context(), c.Request.UserAgent())
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}
