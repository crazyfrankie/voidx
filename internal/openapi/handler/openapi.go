package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/openapi/service"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/response"
	"github.com/crazyfrankie/voidx/pkg/util"
)

type OpenAPIHandler struct {
	svc *service.OpenAPIService
}

func NewOpenAPIHandler(svc *service.OpenAPIService) *OpenAPIHandler {
	return &OpenAPIHandler{svc: svc}
}

func (h *OpenAPIHandler) RegisterRoute(r *gin.RouterGroup) {
	openAPIGroup := r.Group("openapi")
	{
		openAPIGroup.POST("chat", h.Chat())
	}
}

// Chat 开放Chat对话接口
func (h *OpenAPIHandler) Chat() gin.HandlerFunc {
	return func(c *gin.Context) {
		var chatReq req.OpenAPIChatReq
		if err := c.ShouldBindJSON(&chatReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		// 根据stream参数决定响应类型
		if chatReq.Stream {
			// 设置SSE响应头
			c.Header("Content-Type", "text/event-stream")
			c.Header("Cache-Control", "no-cache")
			c.Header("Connection", "keep-alive")
			c.Header("Access-Control-Allow-Origin", "*")

			// 获取流式响应
			eventChan, err := h.svc.ProcessStreamChat(c.Request.Context(), userID, chatReq)
			if err != nil {
				response.Error(c, err)
				return
			}

			// 流式输出
			c.Stream(func(w io.Writer) bool {
				select {
				case event, ok := <-eventChan:
					if !ok {
						return false
					}

					fmt.Fprint(w, event)

					// 刷新缓冲区
					if f, ok := w.(http.Flusher); ok {
						f.Flush()
					}
					return true
				case <-c.Request.Context().Done():
					return false
				}
			})
		} else {
			// 块内容输出
			chatResp, err := h.svc.Chat(c.Request.Context(), userID, chatReq)
			if err != nil {
				response.Error(c, err)
				return
			}

			response.SuccessWithData(c, chatResp)
		}
	}
}
