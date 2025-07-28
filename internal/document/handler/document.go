package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/document/service"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/response"
)

type DocumentHandler struct {
	svc *service.DocumentService
}

func NewDocumentHandler(svc *service.DocumentService) *DocumentHandler {
	return &DocumentHandler{svc: svc}
}

func (h *DocumentHandler) RegisterRoute(r *gin.RouterGroup) {
	documentGroup := r.Group("datasets/:dataset_id/documents")
	{
		documentGroup.POST("", h.CreateDocument())
		documentGroup.GET("", h.GetDocumentsWithPage())
		documentGroup.GET("/:document_id", h.GetDocument())
		documentGroup.PUT("/:document_id", h.UpdateDocument())
		documentGroup.DELETE("/:document_id", h.DeleteDocument())
		documentGroup.PUT("/:document_id/enabled", h.UpdateDocumentEnabled())
		documentGroup.POST("/:document_id/processing", h.ProcessDocument())
	}
}

func (h *DocumentHandler) CreateDocument() gin.HandlerFunc {
	return func(c *gin.Context) {
		datasetIDStr := c.Param("dataset_id")
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("知识库ID格式错误"))
			return
		}

		var createReq req.CreateDocumentsReq
		if err := c.ShouldBind(&createReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		document, batch, err := h.svc.CreateDocuments(c.Request.Context(), datasetID, createReq)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, gin.H{
			"documents": document,
			"batch":     batch,
		})
	}
}

func (h *DocumentHandler) GetDocumentsWithPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		datasetIDStr := c.Param("dataset_id")
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("知识库ID格式错误"))
			return
		}

		var pageReq req.GetDocumentsWithPageReq
		if err := c.ShouldBindQuery(&pageReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		documents, paginator, err := h.svc.GetDocumentsWithPage(c.Request.Context(), datasetID, pageReq)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, gin.H{
			"list":      documents,
			"paginator": paginator,
		})
	}
}

func (h *DocumentHandler) GetDocument() gin.HandlerFunc {
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

		document, err := h.svc.GetDocument(c.Request.Context(), datasetID, documentID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, document)
	}
}

func (h *DocumentHandler) UpdateDocument() gin.HandlerFunc {
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

		var updateReq req.UpdateDocumentReq
		if err := c.ShouldBind(&updateReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		err = h.svc.UpdateDocument(c.Request.Context(), datasetID, documentID, updateReq)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}

func (h *DocumentHandler) DeleteDocument() gin.HandlerFunc {
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

		err = h.svc.DeleteDocument(c.Request.Context(), datasetID, documentID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}

func (h *DocumentHandler) UpdateDocumentEnabled() gin.HandlerFunc {
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

		var enabledReq struct {
			Enabled bool `json:"enabled"`
		}
		if err := c.ShouldBind(&enabledReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		err = h.svc.UpdateDocumentEnabled(c.Request.Context(), datasetID, documentID, enabledReq.Enabled)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}

func (h *DocumentHandler) ProcessDocument() gin.HandlerFunc {
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

		err = h.svc.ProcessDocument(c.Request.Context(), datasetID, documentID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}
