package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/voidx/internal/auth/service"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/response"
	"github.com/crazyfrankie/voidx/pkg/util"
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
		authGroup.POST("logout", h.Logout())
	}
}

func (h *AuthHandler) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var loginReq req.LoginReq
		if err := c.ShouldBind(&loginReq); err != nil {
			response.Error(c, errno.ErrValidate)
			return
		}

		tokens, err := h.svc.Login(c.Request.Context(), c.Request.UserAgent(), loginReq)
		if err != nil {
			response.Error(c, err)
			return
		}

		util.SetAuthorization(c, tokens[0], tokens[1])

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

		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie("llmops_refresh", "", int(time.Hour*24), "/", "", false, true)

		response.Success(c)
	}
}
