package resp

import "github.com/google/uuid"

// GetWorkflowResp 获取工作流详情响应
type GetWorkflowResp struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	ToolCallName  string    `json:"tool_call_name"`
	Icon          string    `json:"icon"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	IsDebugPassed bool      `json:"is_debug_passed"`
	NodeCount     int       `json:"node_count"`
	PublishedAt   int64     `json:"published_at"`
	Ctime         int64     `json:"ctime"`
	Utime         int64     `json:"utime"`
}

// GetWorkflowsWithPageResp 获取工作流分页列表数据响应
type GetWorkflowsWithPageResp struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	ToolCallName  string    `json:"tool_call_name"`
	Icon          string    `json:"icon"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	IsDebugPassed bool      `json:"is_debug_passed"`
	NodeCount     int       `json:"node_count"`
	PublishedAt   int64     `json:"published_at"`
	Ctime         int64     `json:"ctime"`
	Utime         int64     `json:"utime"`
}

// WorkflowDebugEvent 工作流调试事件响应（流式）
type WorkflowDebugEvent struct {
	ID          string         `json:"id"`
	NodeID      string         `json:"node_id"`
	NodeType    string         `json:"node_type"`
	Title       string         `json:"title"`
	Status      string         `json:"status"`
	Inputs      map[string]any `json:"inputs"`
	Outputs     map[string]any `json:"outputs"`
	Error       string         `json:"error"`
	ElapsedTime float64        `json:"elapsed_time"`
}
