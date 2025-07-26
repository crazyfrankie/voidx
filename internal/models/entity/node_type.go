package entity

// NodeType 节点类型枚举
type NodeType string

const (
	// 基础节点类型
	NodeTypeStart NodeType = "start" // 开始节点
	NodeTypeEnd   NodeType = "end"   // 结束节点

	// 处理节点类型
	NodeTypeLLM         NodeType = "llm"          // LLM节点
	NodeTypeTemplate    NodeType = "template"     // 模板转换节点
	NodeTypeRetrieval   NodeType = "retrieval"    // 数据集检索节点
	NodeTypeCode        NodeType = "code"         // 代码节点
	NodeTypeTool        NodeType = "tool"         // 工具节点
	NodeTypeHTTPRequest NodeType = "http_request" // HTTP请求节点
	NodeTypeClassifier  NodeType = "classifier"   // 问题分类节点
	NodeTypeIteration   NodeType = "iteration"    // 迭代节点
)

// IsValid 检查节点类型是否有效
func (t NodeType) IsValid() bool {
	switch t {
	case NodeTypeStart, NodeTypeEnd,
		NodeTypeLLM, NodeTypeTemplate, NodeTypeRetrieval,
		NodeTypeCode, NodeTypeTool, NodeTypeHTTPRequest,
		NodeTypeClassifier, NodeTypeIteration:
		return true
	default:
		return false
	}
}

// String 返回节点类型的字符串表示
func (t NodeType) String() string {
	return string(t)
}
