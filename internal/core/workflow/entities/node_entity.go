package entities

import (
	"time"

	"github.com/google/uuid"
)

// NodeType 节点类型枚举
type NodeType string

const (
	NodeTypeStart              NodeType = "start"
	NodeTypeLLM                NodeType = "llm"
	NodeTypeTool               NodeType = "tool"
	NodeTypeCode               NodeType = "code"
	NodeTypeDatasetRetrieval   NodeType = "dataset_retrieval"
	NodeTypeHTTPRequest        NodeType = "http_request"
	NodeTypeTemplateTransform  NodeType = "template_transform"
	NodeTypeQuestionClassifier NodeType = "question_classifier"
	NodeTypeIteration          NodeType = "iteration"
	NodeTypeEnd                NodeType = "end"
)

// Position 节点坐标基础模型
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// BaseNodeData 基础节点数据
type BaseNodeData struct {
	ID          uuid.UUID `json:"id"`          // 节点id，数值必须唯一
	NodeType    NodeType  `json:"node_type"`   // 节点类型
	Title       string    `json:"title"`       // 节点标题，数据也必须唯一
	Description string    `json:"description"` // 节点描述信息
	Position    Position  `json:"position"`    // 节点对应的坐标信息
}

// NodeDataInterface 节点数据接口，用于获取输入输出变量
type NodeDataInterface interface {
	GetInputs() []*VariableEntity
	GetOutputs() []*VariableEntity
	GetBaseNodeData() *BaseNodeData
}

// NewBaseNodeData 创建新的基础节点数据
func NewBaseNodeData() *BaseNodeData {
	return &BaseNodeData{
		ID:       uuid.New(),
		Position: Position{X: 0, Y: 0},
	}
}

// NodeStatus 节点状态
type NodeStatus string

const (
	NodeStatusRunning   NodeStatus = "running"
	NodeStatusSucceeded NodeStatus = "succeeded"
	NodeStatusFailed    NodeStatus = "failed"
)

// NodeResult 节点运行结果
type NodeResult struct {
	NodeData *BaseNodeData          `json:"node_data"` // 节点基础数据
	Status   NodeStatus             `json:"status"`    // 节点运行状态
	Inputs   map[string]interface{} `json:"inputs"`    // 节点的输入数据
	Outputs  map[string]interface{} `json:"outputs"`   // 节点的输出数据
	Latency  time.Duration          `json:"latency"`   // 节点响应耗时
	Error    string                 `json:"error"`     // 节点运行错误信息
}

// NewNodeResult 创建新的节点结果
func NewNodeResult(nodeData *BaseNodeData) *NodeResult {
	return &NodeResult{
		NodeData: nodeData,
		Status:   NodeStatusRunning,
		Inputs:   make(map[string]interface{}),
		Outputs:  make(map[string]interface{}),
		Latency:  0,
		Error:    "",
	}
}
