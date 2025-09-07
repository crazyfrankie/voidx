package tool

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/crazyfrankie/voidx/types/errno"
	"github.com/tmc/langchaingo/tools"
	"gorm.io/gorm"

	toolsentity "github.com/crazyfrankie/voidx/internal/core/tools/api_tools/entities"
	api "github.com/crazyfrankie/voidx/internal/core/tools/api_tools/providers"
	builtin "github.com/crazyfrankie/voidx/internal/core/tools/builtin_tools/providers"
	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes"
	"github.com/crazyfrankie/voidx/internal/models/entity"
	"github.com/crazyfrankie/voidx/pkg/logs"
)

// ToolNode 扩展插件节点
type ToolNode struct {
	*nodes.BaseNodeImpl
	nodeData            *ToolNodeData
	tool                tools.Tool
	builtinToolProvider *builtin.BuiltinProviderManager
	apiToolProvider     *api.ApiProviderManager
	db                  *gorm.DB
}

// NewToolNode 创建新的工具节点
func NewToolNode(nodeData *ToolNodeData,
	builtinToolProvider *builtin.BuiltinProviderManager,
	apiToolProvider *api.ApiProviderManager, db *gorm.DB) *ToolNode {
	node := &ToolNode{
		BaseNodeImpl:        nodes.NewBaseNodeImpl(nodeData.BaseNodeData),
		nodeData:            nodeData,
		builtinToolProvider: builtinToolProvider,
		apiToolProvider:     apiToolProvider,
		db:                  db,
	}

	// 初始化工具
	if err := node.initializeTool(); err != nil {
		// 在实际应用中，这里应该记录错误日志
		logs.Errorf("Failed to initialize tool: %v", err)
	}

	return node
}

// initializeTool 初始化工具
func (t *ToolNode) initializeTool() error {
	switch t.nodeData.ToolType {
	case ToolTypeBuiltin:
		return t.initializeBuiltinTool()
	case ToolTypeAPI:
		return t.initializeAPITool()
	default:
		return fmt.Errorf("不支持的工具类型: %s", t.nodeData.ToolType)
	}
}

// initializeBuiltinTool 初始化内置工具
func (t *ToolNode) initializeBuiltinTool() error {
	tool := t.builtinToolProvider.GetTool(t.nodeData.ProviderID, t.nodeData.ToolID)
	t.tool = tool.(tools.Tool)

	return nil
}

// initializeAPITool 初始化API工具
func (t *ToolNode) initializeAPITool() error {
	var apiTool *entity.ApiTool
	if err := t.db.Model(&entity.ApiTool{}).
		Where("provider_id = ? AND name = ?", t.nodeData.ProviderID, t.nodeData.ToolID).
		Find(&apiTool).Error; err != nil {
		return fmt.Errorf("获取API工具失败: %v", err)
	}
	if apiTool == nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("该API扩展插件不存在，请核实重试"))
	}

	var apiToolProvider *entity.ApiToolProvider
	if err := t.db.Model(&entity.ApiToolProvider{}).
		Where("id = ?", t.nodeData.ProviderID).
		Find(&apiToolProvider).Error; err != nil {
		return fmt.Errorf("获取API提供者失败: %v", err)
	}
	if apiToolProvider == nil {
		return errno.ErrNotFound.AppendBizMessage(errors.New("该API提供商不存在，请核实重试"))
	}

	t.tool = t.apiToolProvider.GetTool(&toolsentity.ToolEntity{
		ID:          apiTool.ID.String(),
		Name:        apiTool.Name,
		URL:         apiTool.URL,
		Method:      apiTool.Method,
		Description: apiTool.Description,
		Headers:     apiToolProvider.Headers,
		Parameters:  apiTool.Parameters,
	})

	return nil
}

// Invoke 扩展插件执行节点，根据传递的信息调用预设的插件，涵盖内置插件及API插件
func (t *ToolNode) Invoke(state *entities.WorkflowState) (*entities.WorkflowState, error) {
	startAt := time.Now()

	// 调用插件并获取结果
	result, err := t.tool.Call(context.Background(), state.Inputs)
	if err != nil {
		nodeResult := entities.NewNodeResult(t.nodeData.BaseNodeData)
		nodeResult.Status = entities.NodeStatusFailed
		nodeResult.Error = fmt.Sprintf("扩展插件执行失败: %v", err)
		nodeResult.Latency = time.Since(startAt)

		newState := &entities.WorkflowState{
			Inputs:      state.Inputs,
			Outputs:     state.Outputs,
			NodeResults: append(state.NodeResults, nodeResult),
		}

		return newState, err
	}

	state.Outputs = result

	// 构建节点结果
	nodeResult := entities.NewNodeResult(t.nodeData.BaseNodeData)
	nodeResult.Status = entities.NodeStatusSucceeded
	nodeResult.Inputs, _ = state.GetInputsAsMap()
	nodeResult.Outputs, _ = state.GetOutputsAsMap()
	nodeResult.Latency = time.Since(startAt)

	// 构建新状态
	newState := &entities.WorkflowState{
		Inputs:      state.Inputs,
		Outputs:     state.Outputs,
		NodeResults: append(state.NodeResults, nodeResult),
	}

	return newState, nil
}
