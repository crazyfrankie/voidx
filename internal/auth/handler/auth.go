package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/voidx/internal/auth/service"
	"github.com/crazyfrankie/voidx/internal/base/response"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/pkg/util"
	"github.com/crazyfrankie/voidx/types/errno"
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
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		tokens, err := h.svc.Login(c.Request.Context(), c.Request.UserAgent(), loginReq)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
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
			response.InternalServerErrorResponse(c, err)
			return
		}

		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie("llmops_refresh", "", int(time.Hour*24), "/", "", false, true)

		response.Success(c)
	}
}
