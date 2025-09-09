package workflow

import (
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/google/uuid"

	"github.com/crazyfrankie/voidx/internal/core/retrievers"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes"
)

// WorkflowManager manages workflow creation and execution
type WorkflowManager struct {
	llmModel         model.BaseChatModel
	retrieverService *retrievers.RetrieverService
	nodeFactory      *nodes.NodeFactory
	toolManager      map[string]tool.InvokableTool
}

// NewWorkflowManager creates a new workflow manager
func NewWorkflowManager(llmModel model.BaseChatModel, retrieverService *retrievers.RetrieverService) *WorkflowManager {
	nodeFactory := nodes.NewNodeFactory(llmModel, retrieverService)

	return &WorkflowManager{
		llmModel:         llmModel,
		retrieverService: retrieverService,
		nodeFactory:      nodeFactory,
		toolManager:      make(map[string]tool.InvokableTool),
	}
}

// RegisterTool registers a tool with the workflow manager
func (wm *WorkflowManager) RegisterTool(name string, tool tool.InvokableTool) {
	wm.toolManager[name] = tool
	wm.nodeFactory.RegisterTool(name, tool)
}

// CreateWorkflow creates a new workflow from configuration
func (wm *WorkflowManager) CreateWorkflow(values map[string]interface{}, accountID uuid.UUID) (*Workflow, error) {
	// Create workflow
	wf, err := NewWorkflow(values)
	if err != nil {
		return nil, err
	}

	// Set the node factory and account ID
	wf.nodeFactory = wm.nodeFactory
	wf.accountID = accountID

	return wf, nil
}
