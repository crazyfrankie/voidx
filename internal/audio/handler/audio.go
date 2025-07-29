package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/voidx/internal/audio/service"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/models/resp"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/response"
	"github.com/crazyfrankie/voidx/pkg/util"
)

type AudioHandler struct {
	svc *service.AudioService
}

func NewAudioHandler(svc *service.AudioService) *AudioHandler {
	return &AudioHandler{svc: svc}
}

func (h *AudioHandler) RegisterRoute(r *gin.RouterGroup) {
	audioGroup := r.Group("audio")
	{
		audioGroup.POST("audio-to-text", h.AudioToText())
		audioGroup.POST("text-to-audio", h.MessageToAudio())
	}
}

// AudioToText 将语音转换成文本
func (h *AudioHandler) AudioToText() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取上传的文件
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("音频文件不能为空"))
			return
		}
		defer file.Close()

		// 检查文件大小（25MB限制）
		if header.Size > 25*1024*1024 {
			response.Error(c, errno.ErrValidate.AppendBizMessage("音频文件不能超过25MB"))
			return
		}

		// 读取文件内容
		fileContent, err := io.ReadAll(file)
		if err != nil {
			response.Error(c, errno.ErrInternalServer.AppendBizMessage("读取音频文件失败"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		// 调用服务转换语音
		text, err := h.svc.AudioToText(c.Request.Context(), userID, fileContent, header.Filename)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, resp.AudioToTextResp{Text: text})
	}
}

// MessageToAudio 将消息转换成流式输出音频
func (h *AudioHandler) MessageToAudio() gin.HandlerFunc {
	return func(c *gin.Context) {
		var audioReq req.MessageToAudioReq
		if err := c.ShouldBindUri(&audioReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("消息ID格式错误"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		// 设置SSE响应头
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Access-Control-Allow-Origin", "*")

		// 获取流式响应
		eventChan, err := h.svc.MessageToAudio(c.Request.Context(), userID, audioReq.MessageID)
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

				eventData, _ := sonic.Marshal(event)
				fmt.Fprintf(w, "event: tts_message\ndata: %s\n\n", string(eventData))

				// 刷新缓冲区
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				return true
			case <-c.Request.Context().Done():
				return false
			}
		})
	}
}
