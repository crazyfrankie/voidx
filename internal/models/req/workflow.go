package req

// CreateWorkflowReq 创建工作流基础请求
type CreateWorkflowReq struct {
	Name         string `json:"name" binding:"required,max=50"`
	ToolCallName string `json:"tool_call_name" binding:"required,max=50,alphanum"`
	Icon         string `json:"icon" binding:"required,url"`
	Description  string `json:"description" binding:"required,max=1024"`
}

// UpdateWorkflowReq 更新工作流基础请求
type UpdateWorkflowReq struct {
	Name         string `json:"name" binding:"required,max=50"`
	ToolCallName string `json:"tool_call_name" binding:"required,max=50,alphanum"`
	Icon         string `json:"icon" binding:"required,url"`
	Description  string `json:"description" binding:"required,max=1024"`
}

// GetWorkflowsWithPageReq 获取工作流分页列表数据请求
type GetWorkflowsWithPageReq struct {
	Page       int    `form:"page" binding:"min=1"`
	PageSize   int    `form:"page_size" binding:"min=1,max=100"`
	Status     string `form:"status"`
	SearchWord string `form:"search_word"`
}

// UpdateDraftGraphReq 更新工作流草稿图配置请求
type UpdateDraftGraphReq struct {
	Nodes []map[string]any `json:"nodes"`
	Edges []map[string]any `json:"edges"`
}

// DebugWorkflowReq 调试工作流请求
type DebugWorkflowReq struct {
	Inputs map[string]any `json:"inputs"`
}
