package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/req"
	"github.com/crazyfrankie/voidx/internal/workflow/service"
	"github.com/crazyfrankie/voidx/pkg/errno"
	"github.com/crazyfrankie/voidx/pkg/response"
	"github.com/crazyfrankie/voidx/pkg/util"
)

type WorkflowHandler struct {
	svc *service.WorkflowService
}

func NewWorkflowHandler(svc *service.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{svc: svc}
}

func (h *WorkflowHandler) RegisterRoute(r *gin.RouterGroup) {
	workflowGroup := r.Group("workflows")
	{
		workflowGroup.POST("", h.CreateWorkflow())
		workflowGroup.GET("", h.GetWorkflowsWithPage())
		workflowGroup.GET(":workflow_id", h.GetWorkflow())
		workflowGroup.PUT(":workflow_id", h.UpdateWorkflow())
		workflowGroup.DELETE(":workflow_id", h.DeleteWorkflow())
		workflowGroup.PUT(":workflow_id/draft-graph", h.UpdateDraftGraph())
		workflowGroup.GET(":workflow_id/draft-graph", h.GetDraftGraph())
		workflowGroup.POST(":workflow_id/debug", h.DebugWorkflow())
		workflowGroup.POST(":workflow_id/publish", h.PublishWorkflow())
		workflowGroup.POST(":workflow_id/cancel-publish", h.CancelPublishWorkflow())
	}
}

// CreateWorkflow 新增工作流
func (h *WorkflowHandler) CreateWorkflow() gin.HandlerFunc {
	return func(c *gin.Context) {
		var createReq req.CreateWorkflowReq
		if err := c.ShouldBindJSON(&createReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		workflow, err := h.svc.CreateWorkflow(c.Request.Context(), userID, createReq)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, map[string]any{"id": workflow.ID})
	}
}

// GetWorkflowsWithPage 获取当前登录账号下的工作流分页列表数据
func (h *WorkflowHandler) GetWorkflowsWithPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var pageReq req.GetWorkflowsWithPageReq
		if err := c.ShouldBindQuery(&pageReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		// 设置默认值
		if pageReq.Page == 0 {
			pageReq.Page = 1
		}
		if pageReq.PageSize == 0 {
			pageReq.PageSize = 20
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		workflows, paginator, err := h.svc.GetWorkflowsWithPage(c.Request.Context(), userID, pageReq)
		if err != nil {
			response.Error(c, err)
			return
		}

		result := map[string]any{
			"list":      workflows,
			"paginator": paginator,
		}

		response.SuccessWithData(c, result)
	}
}

// GetWorkflow 根据传递的工作流id获取工作流详情
func (h *WorkflowHandler) GetWorkflow() gin.HandlerFunc {
	return func(c *gin.Context) {
		workflowIDStr := c.Param("workflow_id")
		workflowID, err := uuid.Parse(workflowIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("无效的工作流ID格式"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		workflow, err := h.svc.GetWorkflow(c.Request.Context(), workflowID, userID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, workflow)
	}
}

// UpdateWorkflow 根据传递的工作流id更新工作流基础信息
func (h *WorkflowHandler) UpdateWorkflow() gin.HandlerFunc {
	return func(c *gin.Context) {
		workflowIDStr := c.Param("workflow_id")
		workflowID, err := uuid.Parse(workflowIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("无效的工作流ID格式"))
			return
		}

		var updateReq req.UpdateWorkflowReq
		if err := c.ShouldBindJSON(&updateReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		err = h.svc.UpdateWorkflow(c.Request.Context(), workflowID, userID, updateReq)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}

// DeleteWorkflow 根据传递的工作流id删除指定的工作流
func (h *WorkflowHandler) DeleteWorkflow() gin.HandlerFunc {
	return func(c *gin.Context) {
		workflowIDStr := c.Param("workflow_id")
		workflowID, err := uuid.Parse(workflowIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("无效的工作流ID格式"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		err = h.svc.DeleteWorkflow(c.Request.Context(), workflowID, userID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}

// UpdateDraftGraph 根据传递的工作流id+请求信息更新工作流草稿图配置
func (h *WorkflowHandler) UpdateDraftGraph() gin.HandlerFunc {
	return func(c *gin.Context) {
		workflowIDStr := c.Param("workflow_id")
		workflowID, err := uuid.Parse(workflowIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("无效的工作流ID格式"))
			return
		}

		var draftGraphReq req.UpdateDraftGraphReq
		if err := c.ShouldBindJSON(&draftGraphReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		draftGraph := map[string]any{
			"nodes": draftGraphReq.Nodes,
			"edges": draftGraphReq.Edges,
		}

		err = h.svc.UpdateDraftGraph(c.Request.Context(), workflowID, userID, draftGraph)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}

// GetDraftGraph 根据传递的工作流id获取该工作流的草稿配置信息
func (h *WorkflowHandler) GetDraftGraph() gin.HandlerFunc {
	return func(c *gin.Context) {
		workflowIDStr := c.Param("workflow_id")
		workflowID, err := uuid.Parse(workflowIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("无效的工作流ID格式"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		draftGraph, err := h.svc.GetDraftGraph(c.Request.Context(), workflowID, userID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.SuccessWithData(c, draftGraph)
	}
}

// DebugWorkflow 根据传递的变量字典+工作流id调试指定的工作流
func (h *WorkflowHandler) DebugWorkflow() gin.HandlerFunc {
	return func(c *gin.Context) {
		workflowIDStr := c.Param("workflow_id")
		workflowID, err := uuid.Parse(workflowIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("无效的工作流ID格式"))
			return
		}

		var debugReq req.DebugWorkflowReq
		if err := c.ShouldBindJSON(&debugReq); err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("请求参数验证失败"))
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
		eventChan, err := h.svc.DebugWorkflow(c.Request.Context(), workflowID, userID, debugReq.Inputs)
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

				eventData, _ := json.Marshal(event)
				fmt.Fprintf(w, "event: workflow\ndata: %s\n\n", string(eventData))

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

// PublishWorkflow 根据传递的工作流id发布指定的工作流
func (h *WorkflowHandler) PublishWorkflow() gin.HandlerFunc {
	return func(c *gin.Context) {
		workflowIDStr := c.Param("workflow_id")
		workflowID, err := uuid.Parse(workflowIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("无效的工作流ID格式"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		err = h.svc.PublishWorkflow(c.Request.Context(), workflowID, userID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}

// CancelPublishWorkflow 根据传递的工作流id取消发布指定的工作流
func (h *WorkflowHandler) CancelPublishWorkflow() gin.HandlerFunc {
	return func(c *gin.Context) {
		workflowIDStr := c.Param("workflow_id")
		workflowID, err := uuid.Parse(workflowIDStr)
		if err != nil {
			response.Error(c, errno.ErrValidate.AppendBizMessage("无效的工作流ID格式"))
			return
		}

		userID, err := util.GetCurrentUserID(c.Request.Context())
		if err != nil {
			response.Error(c, err)
			return
		}

		err = h.svc.CancelPublishWorkflow(c.Request.Context(), workflowID, userID)
		if err != nil {
			response.Error(c, err)
			return
		}

		response.Success(c)
	}
}
