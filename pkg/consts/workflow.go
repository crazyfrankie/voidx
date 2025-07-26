package consts

// Workflow相关常量定义

// WorkflowStatus 工作流状态类型枚举
type WorkflowStatus string

const (
	WorkflowStatusDraft     WorkflowStatus = "draft"
	WorkflowStatusPublished WorkflowStatus = "published"
)

// WorkflowResultStatus 工作流运行结果状态
type WorkflowResultStatus string

const (
	WorkflowResultStatusRunning   WorkflowResultStatus = "running"
	WorkflowResultStatusSucceeded WorkflowResultStatus = "succeeded"
	WorkflowResultStatusFailed    WorkflowResultStatus = "failed"
)

// DefaultWorkflowConfig 工作流默认配置信息，默认添加一个空的工作流
var DefaultWorkflowConfig = map[string]any{
	"graph": map[string]any{},
	"draft_graph": map[string]any{
		"nodes": []any{},
		"edges": []any{},
	},
}
