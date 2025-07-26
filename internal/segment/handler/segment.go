package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/segment/service"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/response"
)

type SegmentHandler struct {
	svc *service.SegmentService
}

func NewSegmentHandler(svc *service.SegmentService) *SegmentHandler {
	return &SegmentHandler{svc: svc}
}

func (h *SegmentHandler) RegisterRoute(r *gin.RouterGroup) {
	segmentGroup := r.Group("datasets/:dataset_id/documents/:document_id/segments")
	{
		segmentGroup.POST("", h.CreateSegment())
		segmentGroup.GET("", h.GetSegmentsWithPage())
		segmentGroup.GET("/:segment_id", h.GetSegment())
		segmentGroup.PUT("/:segment_id", h.UpdateSegment())
		segmentGroup.DELETE("/:segment_id", h.DeleteSegment())
		segmentGroup.PUT("/:segment_id/enabled", h.UpdateSegmentEnabled())
	}
}

func (h *SegmentHandler) CreateSegment() gin.HandlerFunc {
	return func(c *gin.Context) {
		datasetIDStr := c.Param("dataset_id")
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("知识库ID格式错误"))
			return
		}

		documentIDStr := c.Param("document_id")
		documentID, err := uuid.Parse(documentIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("文档ID格式错误"))
			return
		}

		var createReq req.CreateSegmentReq
		if err := c.ShouldBind(&createReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		segment, err := h.svc.CreateSegment(c.Request.Context(), datasetID, documentID, createReq)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, segment)
	}
}

func (h *SegmentHandler) GetSegmentsWithPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		datasetIDStr := c.Param("dataset_id")
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("知识库ID格式错误"))
			return
		}

		documentIDStr := c.Param("document_id")
		documentID, err := uuid.Parse(documentIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("文档ID格式错误"))
			return
		}

		var pageReq req.GetSegmentsWithPageReq
		if err := c.ShouldBindQuery(&pageReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		segments, paginator, err := h.svc.GetSegmentsWithPage(c.Request.Context(), datasetID, documentID, pageReq)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, gin.H{
			"list":      segments,
			"paginator": paginator,
		})
	}
}

func (h *SegmentHandler) GetSegment() gin.HandlerFunc {
	return func(c *gin.Context) {
		datasetIDStr := c.Param("dataset_id")
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("知识库ID格式错误"))
			return
		}

		documentIDStr := c.Param("document_id")
		documentID, err := uuid.Parse(documentIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("文档ID格式错误"))
			return
		}

		segmentIDStr := c.Param("segment_id")
		segmentID, err := uuid.Parse(segmentIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("片段ID格式错误"))
			return
		}

		segment, err := h.svc.GetSegment(c.Request.Context(), datasetID, documentID, segmentID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, segment)
	}
}

func (h *SegmentHandler) UpdateSegment() gin.HandlerFunc {
	return func(c *gin.Context) {
		datasetIDStr := c.Param("dataset_id")
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("知识库ID格式错误"))
			return
		}

		documentIDStr := c.Param("document_id")
		documentID, err := uuid.Parse(documentIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("文档ID格式错误"))
			return
		}

		segmentIDStr := c.Param("segment_id")
		segmentID, err := uuid.Parse(segmentIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("片段ID格式错误"))
			return
		}

		var updateReq req.UpdateSegmentReq
		if err := c.ShouldBind(&updateReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		err = h.svc.UpdateSegment(c.Request.Context(), datasetID, documentID, segmentID, updateReq)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}

func (h *SegmentHandler) DeleteSegment() gin.HandlerFunc {
	return func(c *gin.Context) {
		datasetIDStr := c.Param("dataset_id")
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("知识库ID格式错误"))
			return
		}

		documentIDStr := c.Param("document_id")
		documentID, err := uuid.Parse(documentIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("文档ID格式错误"))
			return
		}

		segmentIDStr := c.Param("segment_id")
		segmentID, err := uuid.Parse(segmentIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("片段ID格式错误"))
			return
		}

		err = h.svc.DeleteSegment(c.Request.Context(), datasetID, documentID, segmentID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}

func (h *SegmentHandler) UpdateSegmentEnabled() gin.HandlerFunc {
	return func(c *gin.Context) {
		datasetIDStr := c.Param("dataset_id")
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("知识库ID格式错误"))
			return
		}

		documentIDStr := c.Param("document_id")
		documentID, err := uuid.Parse(documentIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("文档ID格式错误"))
			return
		}

		segmentIDStr := c.Param("segment_id")
		segmentID, err := uuid.Parse(segmentIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("片段ID格式错误"))
			return
		}

		var enabledReq struct {
			Enabled bool `json:"enabled"`
		}
		if err := c.ShouldBind(&enabledReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		err = h.svc.UpdateSegmentEnabled(c.Request.Context(), datasetID, documentID, segmentID, enabledReq.Enabled)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}
