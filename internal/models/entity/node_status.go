package entity

// NodeStatus 节点状态枚举
type NodeStatus string

const (
	NodeStatusWaiting   NodeStatus = "waiting"   // 等待执行
	NodeStatusRunning   NodeStatus = "running"   // 正在执行
	NodeStatusCompleted NodeStatus = "completed" // 执行完成
	NodeStatusError     NodeStatus = "error"     // 执行错误
	NodeStatusSkipped   NodeStatus = "skipped"   // 跳过执行
)

// IsValid 检查节点状态是否有效
func (s NodeStatus) IsValid() bool {
	switch s {
	case NodeStatusWaiting, NodeStatusRunning, NodeStatusCompleted, NodeStatusError, NodeStatusSkipped:
		return true
	default:
		return false
	}
}

// String 返回节点状态的字符串表示
func (s NodeStatus) String() string {
	return string(s)
}
