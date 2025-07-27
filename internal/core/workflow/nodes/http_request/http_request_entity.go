package http_request

import (
	"fmt"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

// HttpRequestMethod HTTP请求方法类型枚举
type HttpRequestMethod string

const (
	HttpRequestMethodGet     HttpRequestMethod = "get"
	HttpRequestMethodPost    HttpRequestMethod = "post"
	HttpRequestMethodPut     HttpRequestMethod = "put"
	HttpRequestMethodPatch   HttpRequestMethod = "patch"
	HttpRequestMethodDelete  HttpRequestMethod = "delete"
	HttpRequestMethodHead    HttpRequestMethod = "head"
	HttpRequestMethodOptions HttpRequestMethod = "options"
)

// HttpRequestInputType HTTP请求输入变量类型
type HttpRequestInputType string

const (
	HttpRequestInputTypeParams  HttpRequestInputType = "params"  // query参数
	HttpRequestInputTypeHeaders HttpRequestInputType = "headers" // header请求头
	HttpRequestInputTypeBody    HttpRequestInputType = "body"    // body参数
)

// HttpRequestNodeData HTTP请求节点数据
type HttpRequestNodeData struct {
	*entities.BaseNodeData
	URL     string                     `json:"url"`     // 请求URL地址
	Method  HttpRequestMethod          `json:"method"`  // API请求方法
	Inputs  []*entities.VariableEntity `json:"inputs"`  // 输入变量列表
	Outputs []*entities.VariableEntity `json:"outputs"` // 输出变量列表
}

// NewHttpRequestNodeData 创建新的HTTP请求节点数据
func NewHttpRequestNodeData() *HttpRequestNodeData {
	baseData := entities.NewBaseNodeData()
	baseData.NodeType = entities.NodeTypeHTTPRequest

	// 默认输出变量
	outputs := []*entities.VariableEntity{
		{
			Name: "status_code",
			Type: entities.VariableTypeInt,
			Value: entities.VariableValue{
				Type:    entities.VariableValueTypeGenerated,
				Content: 0,
			},
		},
		{
			Name: "text",
			Type: entities.VariableTypeString,
			Value: entities.VariableValue{
				Type: entities.VariableValueTypeGenerated,
			},
		},
	}

	return &HttpRequestNodeData{
		BaseNodeData: baseData,
		Method:       HttpRequestMethodGet,
		Inputs:       make([]*entities.VariableEntity, 0),
		Outputs:      outputs,
	}
}

// ValidateInputs 校验输入列表数据
func (h *HttpRequestNodeData) ValidateInputs() error {
	// 校验判断输入变量列表中的类型信息
	for _, input := range h.Inputs {
		if inputType, exists := input.Meta["type"]; exists {
			switch inputType {
			case string(HttpRequestInputTypeParams),
				string(HttpRequestInputTypeHeaders),
				string(HttpRequestInputTypeBody):
				// 有效类型
			default:
				return fmt.Errorf("HTTP请求参数结构出错: 无效的输入类型 %v", inputType)
			}
		}
	}
	return nil
}
