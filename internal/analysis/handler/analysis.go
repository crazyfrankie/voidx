package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/analysis/service"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/response"
	"github.com/crazyfrankie/voidx/pkg/util"
)

type AnalysisHandler struct {
	svc *service.AnalysisService
}

func NewAnalysisHandler(svc *service.AnalysisService) *AnalysisHandler {
	return &AnalysisHandler{svc: svc}
}

func (h *AnalysisHandler) RegisterRoute(r *gin.RouterGroup) {
	analysisGroup := r.Group("analysis")
	{
		analysisGroup.GET("apps/:app_id", h.GetAppAnalysis())
	}
}

// GetAppAnalysis 根据传递的应用id获取应用的统计信息
func (h *AnalysisHandler) GetAppAnalysis() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取路径参数中的应用ID
		appIDStr := c.Param("app_id")
		appID, err := uuid.Parse(appIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("无效的应用ID格式"))
			return
		}

		// 获取当前用户ID
		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		// 调用服务获取应用分析数据
		analysis, err := h.svc.GetAppAnalysis(c.Request.Context(), appID, userID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, analysis)
	}
}
