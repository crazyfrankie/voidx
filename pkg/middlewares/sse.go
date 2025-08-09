package middlewares

import "github.com/gin-gonic/gin"

var paths = map[string]struct{}{
	"/api/ai/optimize-prompt":        {},
	"/api/assistant-agent/chat":      {},
	"/api/apps/:app_id/conversation": {},
	"/api/webapp/:token/chat":        {},
}

func SSEHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		fullPath := c.FullPath()
		if _, ok := paths[fullPath]; !ok {
			c.Next()
			return
		}

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Transfer-Encoding", "chunked")

		c.Next()
	}
}
