package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/base/response"
	"github.com/crazyfrankie/voidx/internal/dataset/service"
	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/types/errno"
)

type DatasetHandler struct {
	svc *service.DatasetService
}

func NewDatasetHandler(svc *service.DatasetService) *DatasetHandler {
	return &DatasetHandler{svc: svc}
}

func (h *DatasetHandler) RegisterRoute(r *gin.RouterGroup) {
	datasetGroup := r.Group("datasets")
	{
		datasetGroup.POST("", h.CreateDataset())
		datasetGroup.GET("", h.GetDatasetsWithPage())
		datasetGroup.GET("/:dataset_id", h.GetDataset())
		datasetGroup.PUT("/:dataset_id", h.UpdateDataset())
		datasetGroup.DELETE("/:dataset_id", h.DeleteDataset())
		datasetGroup.POST("/:dataset_id/hit", h.Hit())
		datasetGroup.GET("/:dataset_id/queries", h.GetDatasetQueries())
	}
}

func (h *DatasetHandler) CreateDataset() gin.HandlerFunc {
	return func(c *gin.Context) {
		var createReq req.CreateDatasetReq
		if err := c.ShouldBind(&createReq); err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		err := h.svc.CreateDataset(c.Request.Context(), createReq)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Success(c)
	}
}

func (h *DatasetHandler) GetDatasetsWithPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var pageReq req.GetDatasetsWithPageReq
		if err := c.ShouldBindQuery(&pageReq); err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		datasets, paginator, err := h.svc.GetDatasetsWithPage(c.Request.Context(), pageReq)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Data(c, gin.H{
			"list":      datasets,
			"paginator": paginator,
		})
	}
}

func (h *DatasetHandler) GetDataset() gin.HandlerFunc {
	return func(c *gin.Context) {
		datasetIDStr := c.Param("dataset_id")
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		dataset, err := h.svc.GetDataset(c.Request.Context(), datasetID)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Data(c, dataset)
	}
}

func (h *DatasetHandler) UpdateDataset() gin.HandlerFunc {
	return func(c *gin.Context) {
		datasetIDStr := c.Param("dataset_id")
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		var updateReq req.UpdateDatasetReq
		if err := c.ShouldBind(&updateReq); err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		err = h.svc.UpdateDataset(c.Request.Context(), datasetID, updateReq)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Success(c)
	}
}

func (h *DatasetHandler) DeleteDataset() gin.HandlerFunc {
	return func(c *gin.Context) {
		datasetIDStr := c.Param("dataset_id")
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		err = h.svc.DeleteDataset(c.Request.Context(), datasetID)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Success(c)
	}
}

func (h *DatasetHandler) Hit() gin.HandlerFunc {
	return func(c *gin.Context) {
		datasetIDStr := c.Param("dataset_id")
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		var hitReq req.HitReq
		if err := c.ShouldBind(&hitReq); err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		result, err := h.svc.Hit(c.Request.Context(), datasetID, hitReq)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Data(c, result)
	}
}

func (h *DatasetHandler) GetDatasetQueries() gin.HandlerFunc {
	return func(c *gin.Context) {
		datasetIDStr := c.Param("dataset_id")
		datasetID, err := uuid.Parse(datasetIDStr)
		if err != nil {
			response.InvalidParamRequestResponse(c, errno.ErrValidate)
			return
		}

		queries, err := h.svc.GetDatasetQueries(c.Request.Context(), datasetID)
		if err != nil {
			response.InternalServerErrorResponse(c, err)
			return
		}

		response.Data(c, queries)
	}
}
