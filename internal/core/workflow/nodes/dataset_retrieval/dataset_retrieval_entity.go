package dataset_retrieval

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

// RetrievalStrategy 检索策略枚举
type RetrievalStrategy string

const (
	RetrievalStrategySemantic RetrievalStrategy = "semantic" // 语义检索
	RetrievalStrategyKeyword  RetrievalStrategy = "keyword"  // 关键词检索
	RetrievalStrategyHybrid   RetrievalStrategy = "hybrid"   // 混合检索
)

// RetrievalConfig 检索配置
type RetrievalConfig struct {
	RetrievalStrategy RetrievalStrategy `json:"retrieval_strategy"` // 检索策略
	K                 int               `json:"k"`                  // 最大召回数量
	Score             float64           `json:"score"`              // 得分阈值
}

// NewRetrievalConfig 创建默认检索配置
func NewRetrievalConfig() *RetrievalConfig {
	return &RetrievalConfig{
		RetrievalStrategy: RetrievalStrategySemantic,
		K:                 4,
		Score:             0.0,
	}
}

// DatasetRetrievalNodeData 知识库检索节点数据
type DatasetRetrievalNodeData struct {
	*entities.BaseNodeData
	DatasetIDs      []uuid.UUID                `json:"dataset_ids"`      // 关联的知识库id列表
	RetrievalConfig *RetrievalConfig           `json:"retrieval_config"` // 检索配置
	Inputs          []*entities.VariableEntity `json:"inputs"`           // 输入变量信息
	Outputs         []*entities.VariableEntity `json:"outputs"`          // 输出变量信息
}

// NewDatasetRetrievalNodeData 创建新的知识库检索节点数据
func NewDatasetRetrievalNodeData() *DatasetRetrievalNodeData {
	baseData := entities.NewBaseNodeData()
	baseData.NodeType = entities.NodeTypeDatasetRetrieval

	// 默认输出变量
	outputs := []*entities.VariableEntity{
		{
			Name: "combine_documents",
			Type: entities.VariableTypeString,
			Value: entities.VariableValue{
				Type: entities.VariableValueTypeGenerated,
			},
		},
	}

	return &DatasetRetrievalNodeData{
		BaseNodeData:    baseData,
		DatasetIDs:      make([]uuid.UUID, 0),
		RetrievalConfig: NewRetrievalConfig(),
		Inputs:          make([]*entities.VariableEntity, 0),
		Outputs:         outputs,
	}
}

// ValidateInputs 校验输入变量信息
func (d *DatasetRetrievalNodeData) ValidateInputs() error {
	// 1. 判断是否只有一个输入变量，如果有多个则抛出错误
	if len(d.Inputs) != 1 {
		return fmt.Errorf("知识库节点输入变量信息出错")
	}

	// 2. 判断输入变量的类型及字段名称是否出错
	queryInput := d.Inputs[0]
	if queryInput.Name != "query" || queryInput.Type != entities.VariableTypeString || !queryInput.Required {
		return fmt.Errorf("知识库节点输入变量名字/变量类型/必填属性出错")
	}

	return nil
}

// GetInputs 实现NodeDataInterface接口
func (d *DatasetRetrievalNodeData) GetInputs() []*entities.VariableEntity {
	return d.Inputs
}

// GetOutputs 实现NodeDataInterface接口
func (d *DatasetRetrievalNodeData) GetOutputs() []*entities.VariableEntity {
	return d.Outputs
}

// GetBaseNodeData 实现NodeDataInterface接口
func (d *DatasetRetrievalNodeData) GetBaseNodeData() *entities.BaseNodeData {
	return d.BaseNodeData
}
