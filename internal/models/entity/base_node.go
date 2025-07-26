package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Position 节点位置
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// BaseNodeData 基础节点数据
type BaseNodeData struct {
	ID          string   `json:"id"`          // 节点ID
	Type        NodeType `json:"type"`        // 节点类型
	Name        string   `json:"name"`        // 节点名称
	Description string   `json:"description"` // 节点描述
	Position    Position `json:"position"`    // 节点位置
	Inputs      []string `json:"inputs"`      // 输入节点ID列表
	Outputs     []string `json:"outputs"`     // 输出节点ID列表
	Utime       int64    `json:"utime"`
	Ctime       int64    `json:"ctime"`
}

// NodeResult 节点执行结果
type NodeResult struct {
	NodeID    string      `json:"node_id"`    // 节点ID
	Status    NodeStatus  `json:"status"`     // 执行状态
	Output    interface{} `json:"output"`     // 输出结果
	Error     string      `json:"error"`      // 错误信息
	StartTime time.Time   `json:"start_time"` // 开始时间
	EndTime   time.Time   `json:"end_time"`   // 结束时间
}

// VariableEntity 变量实体
type VariableEntity struct {
	ID           uuid.UUID    `json:"id"`            // 变量ID
	Name         string       `json:"name"`          // 变量名称
	Description  string       `json:"description"`   // 变量描述
	Type         VariableType `json:"type"`          // 变量类型
	Required     bool         `json:"required"`      // 是否必填
	DefaultValue interface{}  `json:"default_value"` // 默认值
}

// BaseEdgeData 基础边数据
type BaseEdgeData struct {
	ID        string `json:"id"`         // 边ID
	Source    string `json:"source"`     // 源节点ID
	Target    string `json:"target"`     // 目标节点ID
	Label     string `json:"label"`      // 边标签
	Condition string `json:"condition"`  // 边条件
	Ctime     int64  `json:"created_at"` // 创建时间
	Utime     int64  `json:"updated_at"` // 更新时间
}

// WorkflowConfig 工作流配置
type WorkflowConfig struct {
	ID          uuid.UUID                  `json:"id"`          // 工作流ID
	Name        string                     `json:"name"`        // 工作流名称
	Description string                     `json:"description"` // 工作流描述
	Nodes       map[string]json.RawMessage `json:"nodes"`       // 节点配置
	Edges       []BaseEdgeData             `json:"edges"`       // 边配置
	Variables   []VariableEntity           `json:"variables"`   // 变量配置
	Utime       int64                      `gorm:"autoUpdateTime" json:"utime"`
	Ctime       int64                      `gorm:"autoCreateTime" json:"ctime"`
}

// WorkflowState 工作流状态
type WorkflowState struct {
	WorkflowID  uuid.UUID             `json:"workflow_id"`  // 工作流ID
	Status      NodeStatus            `json:"status"`       // 执行状态
	CurrentNode string                `json:"current_node"` // 当前执行节点
	Variables   map[string]any        `json:"variables"`    // 变量值
	NodeResults map[string]NodeResult `json:"node_results"` // 节点执行结果
	StartTime   time.Time             `json:"start_time"`   // 开始时间
	EndTime     *time.Time            `json:"end_time"`     // 结束时间
	Error       string                `json:"error"`        // 错误信息
}
