package dataset_retrieval

import (
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

// DatasetRetrievalNodeData represents the data structure for dataset retrieval workflow nodes
type DatasetRetrievalNodeData struct {
	*entities.BaseNodeData
	DatasetIDs      []uuid.UUID                `json:"dataset_ids"`
	RetrievalConfig map[string]interface{}     `json:"retrieval_config"`
	Inputs          []*entities.VariableEntity `json:"inputs"`
	Outputs         []*entities.VariableEntity `json:"outputs"`
}

// NewDatasetRetrievalNodeData creates a new dataset retrieval node data instance
func NewDatasetRetrievalNodeData() *DatasetRetrievalNodeData {
	return &DatasetRetrievalNodeData{
		BaseNodeData:    &entities.BaseNodeData{NodeType: entities.NodeTypeDatasetRetrieval},
		DatasetIDs:      make([]uuid.UUID, 0),
		RetrievalConfig: make(map[string]interface{}),
		Inputs:          make([]*entities.VariableEntity, 0),
		Outputs:         make([]*entities.VariableEntity, 0),
	}
}

// GetBaseNodeData returns the base node data (implements NodeDataInterface)
func (d *DatasetRetrievalNodeData) GetBaseNodeData() *entities.BaseNodeData {
	return d.BaseNodeData
}
