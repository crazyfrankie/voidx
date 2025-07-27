package dataset_retrieval

import (
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes"
)

// RetrievalTool 检索工具接口
type RetrievalTool interface {
	Invoke(query string) (string, error)
}

// DatasetRetrievalNode 知识库检索节点
type DatasetRetrievalNode struct {
	*nodes.BaseNodeImpl
	nodeData      *DatasetRetrievalNodeData
	retrievalTool RetrievalTool
}

// NewDatasetRetrievalNode 创建新的知识库检索节点
func NewDatasetRetrievalNode(nodeData *DatasetRetrievalNodeData, accountID uuid.UUID) *DatasetRetrievalNode {
	node := &DatasetRetrievalNode{
		BaseNodeImpl: nodes.NewBaseNodeImpl(nodeData.BaseNodeData),
		nodeData:     nodeData,
	}

	// 初始化检索工具
	node.retrievalTool = node.createRetrievalTool(accountID)

	return node
}

// createRetrievalTool 创建检索工具
func (d *DatasetRetrievalNode) createRetrievalTool(accountID uuid.UUID) RetrievalTool {
	// 这里应该实现真正的检索服务工具创建逻辑
	// 目前提供一个模拟实现
	return &MockRetrievalTool{
		datasetIDs:      d.nodeData.DatasetIDs,
		accountID:       accountID,
		retrievalConfig: d.nodeData.RetrievalConfig,
	}
}

// Invoke 知识库检索节点调用函数，执行响应的知识库检索后返回
func (d *DatasetRetrievalNode) Invoke(state *entities.WorkflowState) (*entities.WorkflowState, error) {
	startAt := time.Now()

	// 1. 提取节点输入变量字典映射
	inputsDict := d.extractVariablesFromState(state)

	// 2. 获取查询字符串
	query := ""
	if queryValue, exists := inputsDict["query"]; exists {
		query = fmt.Sprintf("%v", queryValue)
	}

	// 3. 调用知识库检索工具
	combineDocuments, err := d.retrievalTool.Invoke(query)
	if err != nil {
		nodeResult := entities.NewNodeResult(d.nodeData.BaseNodeData)
		nodeResult.Status = entities.NodeStatusFailed
		nodeResult.Error = fmt.Sprintf("知识库检索失败: %v", err)
		nodeResult.Latency = time.Since(startAt)

		newState := &entities.WorkflowState{
			Inputs:      state.Inputs,
			Outputs:     state.Outputs,
			NodeResults: append(state.NodeResults, nodeResult),
		}

		return newState, err
	}

	// 4. 提取并构建输出数据结构
	outputs := make(map[string]any)
	if len(d.nodeData.Outputs) > 0 {
		outputs[d.nodeData.Outputs[0].Name] = combineDocuments
	} else {
		outputs["combine_documents"] = combineDocuments
	}

	// 5. 构建节点结果
	nodeResult := entities.NewNodeResult(d.nodeData.BaseNodeData)
	nodeResult.Status = entities.NodeStatusSucceeded
	nodeResult.Inputs = inputsDict
	nodeResult.Outputs = outputs
	nodeResult.Latency = time.Since(startAt)

	// 6. 构建新状态
	newState := &entities.WorkflowState{
		Inputs:      state.Inputs,
		Outputs:     state.Outputs,
		NodeResults: append(state.NodeResults, nodeResult),
	}

	return newState, nil
}

// extractVariablesFromState 从状态中提取变量
func (d *DatasetRetrievalNode) extractVariablesFromState(state *entities.WorkflowState) map[string]any {
	result := make(map[string]any)

	for _, input := range d.nodeData.Inputs {
		inputs := make(map[string]any)
		if err := sonic.UnmarshalString(state.Inputs, &inputs); err != nil {
			return nil
		}
		if val, exists := inputs[input.Name]; exists {
			result[input.Name] = val
		}
	}

	return result
}

// MockRetrievalTool 模拟检索工具实现
type MockRetrievalTool struct {
	datasetIDs      []uuid.UUID
	accountID       uuid.UUID
	retrievalConfig *RetrievalConfig
}

func (m *MockRetrievalTool) Invoke(query string) (string, error) {
	// 模拟知识库检索
	if query == "" {
		return "没有提供查询内容", nil
	}

	// 模拟检索结果
	result := fmt.Sprintf(`基于查询 "%s" 的检索结果：

文档1: 这是一个关于 %s 的相关文档内容，包含了详细的信息和解释。

文档2: 另一个与 %s 相关的文档片段，提供了补充信息。

文档3: 第三个相关文档，进一步阐述了 %s 的相关概念。

检索配置: 策略=%s, K=%d, 得分阈值=%.2f
数据集数量: %d`,
		query, query, query, query,
		m.retrievalConfig.RetrievalStrategy,
		m.retrievalConfig.K,
		m.retrievalConfig.Score,
		len(m.datasetIDs))

	return result, nil
}
