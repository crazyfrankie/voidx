package handler

import (
	"errors"
	"io"

	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/voidx/internal/upload/service"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/response"
	"github.com/crazyfrankie/voidx/pkg/util"
)

type UploadFileHandler struct {
	svc *service.OssService
}

func NewUploadFileHandler(svc *service.OssService) *UploadFileHandler {
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
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage(errors.New("上传图片不能为空")))
			return
		}
		defer file.Close()

		// 检查文件大小（15MB限制）
		if header.Size > 15*1024*1024 {
			response.Error(c, errno.ErrValidate.AppendBizMessage(errors.New("上传文件最大不能超过15MB")))
		}
		data, err := io.ReadAll(file)
		if err != nil {
			response.Error(c, err)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, errno.ErrUnauthorized)
			return
		}

		res, err := h.svc.UploadFile(c.Request.Context(), data, false, header.Filename, userID)
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
			response.Error(c, errno.ErrValidate.AppendBizMessage(errors.New("上传图片不能为空")))
			return
		}
		defer file.Close()

		// 检查文件大小（15MB限制）
		if header.Size > 15*1024*1024 {
			response.Error(c, errno.ErrValidate.AppendBizMessage(errors.New("上传文件最大不能超过15MB")))
		}
		data, err := io.ReadAll(file)
		if err != nil {
			response.Error(c, err)
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, errno.ErrUnauthorized)
			return
		}

		res, err := h.svc.UploadFile(c.Request.Context(), data, true, header.Filename, userID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, gin.H{"image_url": res.URL})
	}
}
