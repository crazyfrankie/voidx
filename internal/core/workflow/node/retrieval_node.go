package node

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

// RetrievalNodeData 检索节点数据
type RetrievalNodeData struct {
	*entity.BaseNodeData
	DatasetID      uuid.UUID `json:"dataset_id"`      // 数据集ID
	QueryKey       string    `json:"query_key"`       // 查询键
	OutputKey      string    `json:"output_key"`      // 输出键
	K              int       `json:"k"`               // 返回结果数量
	ScoreThreshold float64   `json:"score_threshold"` // 分数阈值
	RetrievalType  string    `json:"retrieval_type"`  // 检索类型：semantic, keyword, hybrid
}

// RetrievalNode 检索节点
type RetrievalNode struct {
	BaseNode
	Data *RetrievalNodeData
	// 实际项目中需要注入检索服务
	// retrievalService *service.RetrievalService
}

// NewRetrievalNode 创建检索节点
func NewRetrievalNode(data *RetrievalNodeData) *RetrievalNode {
	return &RetrievalNode{
		BaseNode: BaseNode{Data: data.BaseNodeData},
		Data:     data,
	}
}

// Invoke 执行检索节点
func (n *RetrievalNode) Invoke(ctx context.Context, state map[string]any) (map[string]any, error) {
	// 复制当前状态
	result := make(map[string]any)
	for k, v := range state {
		result[k] = v
	}

	// 获取查询内容
	var query string
	if n.Data.QueryKey != "" {
		if queryValue, ok := state[n.Data.QueryKey]; ok {
			if strValue, ok := queryValue.(string); ok {
				query = strValue
			} else {
				return nil, fmt.Errorf("query value for key %s is not a string", n.Data.QueryKey)
			}
		} else {
			return nil, fmt.Errorf("query key %s not found in state", n.Data.QueryKey)
		}
	}

	// 在实际项目中，这里应该调用检索服务
	// 这里只是模拟检索结果
	retrievalResults := []map[string]any{
		{
			"content": fmt.Sprintf("Retrieved content for query: %s", query),
			"score":   0.95,
			"metadata": map[string]any{
				"source": "document_1",
			},
		},
		{
			"content": fmt.Sprintf("Another relevant content for: %s", query),
			"score":   0.87,
			"metadata": map[string]any{
				"source": "document_2",
			},
		},
	}

	// 设置输出
	outputKey := n.Data.OutputKey
	if outputKey == "" {
		outputKey = "retrieval_results"
	}
	result[outputKey] = retrievalResults

	return result, nil
}

// Validate 验证检索节点配置
func (n *RetrievalNode) Validate() error {
	if n.Data.Type != entity.NodeTypeRetrieval {
		return errors.New("invalid node type for retrieval node")
	}

	if n.Data.DatasetID == uuid.Nil {
		return errors.New("dataset_id is required for retrieval node")
	}

	if n.Data.QueryKey == "" {
		return errors.New("query_key is required for retrieval node")
	}

	if n.Data.K <= 0 {
		n.Data.K = 5 // 默认返回5个结果
	}

	validTypes := []string{"semantic", "keyword", "hybrid"}
	if n.Data.RetrievalType == "" {
		n.Data.RetrievalType = "semantic" // 默认语义检索
	} else {
		valid := false
		for _, t := range validTypes {
			if n.Data.RetrievalType == t {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid retrieval_type: %s, must be one of %v", n.Data.RetrievalType, validTypes)
		}
	}

	return nil
}
