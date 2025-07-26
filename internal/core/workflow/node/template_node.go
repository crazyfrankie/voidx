package node

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"text/template"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

// TemplateNodeData 模板节点数据
type TemplateNodeData struct {
	*entity.BaseNodeData
	Template  string `json:"template"`   // 模板内容
	InputKey  string `json:"input_key"`  // 输入键
	OutputKey string `json:"output_key"` // 输出键
}

// TemplateNode 模板节点
type TemplateNode struct {
	BaseNode
	Data     *TemplateNodeData
	template *template.Template
}

// NewTemplateNode 创建模板节点
func NewTemplateNode(data *TemplateNodeData) (*TemplateNode, error) {
	// 解析模板
	tmpl, err := template.New("template").Parse(data.Template)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return &TemplateNode{
		BaseNode: BaseNode{Data: data.BaseNodeData},
		Data:     data,
		template: tmpl,
	}, nil
}

// Invoke 执行模板节点
func (n *TemplateNode) Invoke(ctx context.Context, state map[string]any) (map[string]any, error) {
	// 复制当前状态
	result := make(map[string]any)
	for k, v := range state {
		result[k] = v
	}

	// 执行模板
	var buf bytes.Buffer
	if err := n.template.Execute(&buf, state); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// 设置输出
	outputKey := n.Data.OutputKey
	if outputKey == "" {
		outputKey = "template_output"
	}
	result[outputKey] = buf.String()

	return result, nil
}

// Validate 验证模板节点配置
func (n *TemplateNode) Validate() error {
	if n.Data.Type != entity.NodeTypeTemplate {
		return errors.New("invalid node type for template node")
	}

	if n.Data.Template == "" {
		return errors.New("template is required for template node")
	}

	// 尝试解析模板
	_, err := template.New("template").Parse(n.Data.Template)
	if err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}

	return nil
}
