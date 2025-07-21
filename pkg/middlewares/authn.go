package middlewares

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/jwt"
	"github.com/crazyfrankie/voidx/pkg/response"
)

type AuthnHandler struct {
	ignore map[string]struct{}
	token  *jwt.TokenService
}

func NewAuthnHandler(t *jwt.TokenService) *AuthnHandler {
	return &AuthnHandler{token: t, ignore: make(map[string]struct{})}
}

func (h *AuthnHandler) IgnorePath(path string) *AuthnHandler {
	h.ignore[path] = struct{}{}
	return h
}

func (h *AuthnHandler) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := h.ignore[c.Request.URL.Path]; ok {
			c.Next()
			return
		}

		access, err := h.token.GetAccessToken(c)
		if err == nil {
			if claims, err := h.token.ParseToken(access); err == nil {
				c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "user_id", claims.UID))
				c.Next()
				return
			}
		}

		refresh, err := c.Cookie("llmops_refresh")
		if err != nil {
			response.Error(c, errno.ErrUnauthorized)
			return
		}
		tokens, err := h.token.TryRefresh(refresh, c.Request.UserAgent())
		if err != nil {
			response.Error(c, errno.ErrUnauthorized)
			return
		}

		c.SetSameSite(http.SameSiteLaxMode)
		c.Header("x-access-token", tokens[0])
		c.SetCookie("llmops_refresh", tokens[1], int(time.Hour*24), "/", "", false, true)

		c.Next()
	}
}
