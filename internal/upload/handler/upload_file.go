package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/voidx/internal/upload/service"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/response"
)

type UploadFileHandler struct {
	svc *service.UploadFileService
}

func NewUploadFileHandler(svc *service.UploadFileService) *UploadFileHandler {
	return &UploadFileHandler{svc: svc}
}

func (h *UploadFileHandler) RegisterRoute(r *gin.RouterGroup) {
	uploadGroup := r.Group("upload-files")
	{
		uploadGroup.POST("file", h.UploadFile())
		uploadGroup.POST("image", h.UploadImage())
	}
}

// UploadFile 上传文件/文档
func (h *UploadFileHandler) UploadFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取上传的文件
		header, err := c.FormFile("file")
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("上传文件不能为空"))
			return
		}

		// 调用服务上传文件
		res, err := h.svc.UploadFile(c.Request.Context(), header, false)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, res)
	}
}

// UploadImage 上传图片
func (h *UploadFileHandler) UploadImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取上传的文件
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("上传图片不能为空"))
			return
		}
		defer file.Close()

		// 调用服务上传图片
		res, err := h.svc.UploadFile(c.Request.Context(), header, true)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, res)
	}
}
