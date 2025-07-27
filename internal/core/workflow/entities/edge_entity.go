package entities

import "github.com/google/uuid"

// BaseEdgeData 基础边数据
type BaseEdgeData struct {
	ID             uuid.UUID `json:"id"`               // 边的唯一标识
	Source         uuid.UUID `json:"source"`           // 起始节点ID
	Target         uuid.UUID `json:"target"`           // 目标节点ID
	SourceType     NodeType  `json:"source_type"`      // 起始节点类型
	TargetType     NodeType  `json:"target_type"`      // 目标节点类型
	SourceHandleID *string   `json:"source_handle_id"` // 起始节点句柄ID（用于意图识别等特殊节点）
}

// NewBaseEdgeData 创建新的基础边数据
func NewBaseEdgeData() *BaseEdgeData {
	return &BaseEdgeData{
		ID: uuid.New(),
	}
}
