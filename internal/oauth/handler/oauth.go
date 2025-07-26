package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/oauth/service"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/response"
	"github.com/crazyfrankie/voidx/pkg/util"
)

type OAuthHandler struct {
	svc *service.OAuthService
}

func NewOAuthHandler(svc *service.OAuthService) *OAuthHandler {
	return &OAuthHandler{svc: svc}
}

func (h *OAuthHandler) RegisterRoute(r *gin.RouterGroup) {
	oauthGroup := r.Group("oauth")
	{
		oauthGroup.GET("/:provider_name", h.Provider())
		oauthGroup.POST("/authorize/:provider_name", h.Authorize())
	}
}

func (h *OAuthHandler) Provider() gin.HandlerFunc {
	return func(c *gin.Context) {
		providerName := c.Param("provider_name")

		oauth, err := h.svc.GetOAuthByProviderName(c.Request.Context(), providerName)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, gin.H{"redirect_url": oauth.GetAuthorizationURL()})
	}
}

func (h *OAuthHandler) Authorize() gin.HandlerFunc {
	return func(c *gin.Context) {
		providerName := c.Param("provider_name")

		var authReq req.AuthorizeReq
		if err := c.ShouldBind(&authReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("code代码不能为空"))
			return
		}

		newCtx := context.WithValue(c.Request.Context(), "last_login_ip", c.Request.RemoteAddr)
		tokens, err := h.svc.OAuthLogin(newCtx, providerName, authReq.Code, c.Request.UserAgent())
		if err != nil {
			response.Error(c, err)
			return
		}

		util.SetAuthorization(c, tokens[0], tokens[1])

		response.Success(c)
	}
}
