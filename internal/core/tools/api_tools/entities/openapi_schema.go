package entities

import (
	"fmt"
	"reflect"
	"strings"
)

// ParameterType 参数支持的类型
type ParameterType string

const (
	ParameterTypeStr   ParameterType = "str"
	ParameterTypeInt   ParameterType = "int"
	ParameterTypeFloat ParameterType = "float"
	ParameterTypeBool  ParameterType = "bool"
)

// ParameterTypeMap 参数类型映射
var ParameterTypeMap = map[ParameterType]reflect.Type{
	ParameterTypeStr:   reflect.TypeOf(""),
	ParameterTypeInt:   reflect.TypeOf(0),
	ParameterTypeFloat: reflect.TypeOf(0.0),
	ParameterTypeBool:  reflect.TypeOf(false),
}

// ParameterIn 参数支持存放的位置
type ParameterIn string

const (
	ParameterInPath        ParameterIn = "path"
	ParameterInQuery       ParameterIn = "query"
	ParameterInHeader      ParameterIn = "header"
	ParameterInCookie      ParameterIn = "cookie"
	ParameterInRequestBody ParameterIn = "request_body"
)

// Parameter OpenAPI参数定义
type Parameter struct {
	Name        string      `json:"name"`
	In          ParameterIn `json:"in"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Type        string      `json:"type"`
}

// Operation OpenAPI操作定义
type Operation struct {
	Description string      `json:"description"`
	OperationID string      `json:"operationId"`
	Parameters  []Parameter `json:"parameters"`
}

// PathItem OpenAPI路径项定义
type PathItem struct {
	Get  *Operation `json:"get,omitempty"`
	Post *Operation `json:"post,omitempty"`
}

// OpenAPISchema OpenAPI规范的数据结构
type OpenAPISchema struct {
	Server      string              `json:"server"`
	Description string              `json:"description"`
	Paths       map[string]PathItem `json:"paths"`
}

// ValidateError 验证错误
type ValidateError struct {
	Message string
}

func (e *ValidateError) Error() string {
	return e.Message
}

// NewOpenAPISchema 创建并验证OpenAPI Schema
func NewOpenAPISchema(server, description string, paths map[string]any) (*OpenAPISchema, error) {
	schema := &OpenAPISchema{
		Server:      server,
		Description: description,
		Paths:       make(map[string]PathItem),
	}

	// 验证server
	if err := schema.validateServer(); err != nil {
		return nil, err
	}

	// 验证description
	if err := schema.validateDescription(); err != nil {
		return nil, err
	}

	// 验证并处理paths
	if err := schema.validateAndProcessPaths(paths); err != nil {
		return nil, err
	}

	return schema, nil
}

// validateServer 校验server数据
func (s *OpenAPISchema) validateServer() error {
	if s.Server == "" {
		return &ValidateError{Message: "server不能为空且为字符串"}
	}
	return nil
}

// validateDescription 校验description信息
func (s *OpenAPISchema) validateDescription() error {
	if s.Description == "" {
		return &ValidateError{Message: "description不能为空且为字符串"}
	}
	return nil
}

// validateAndProcessPaths 校验paths信息，涵盖：方法提取、operationId唯一标识，parameters校验
func (s *OpenAPISchema) validateAndProcessPaths(paths map[string]any) error {
	// 1.paths不能为空且类型为字典
	if paths == nil || len(paths) == 0 {
		return &ValidateError{Message: "openapi_schema中的paths不能为空且必须为字典"}
	}

	// 2.提取paths里的每一个元素，并获取元素下的get/post方法对应的值
	methods := []string{"get", "post"}
	var interfaces []map[string]any

	for path, pathItemInterface := range paths {
		pathItem, ok := pathItemInterface.(map[string]any)
		if !ok {
			continue
		}

		for _, method := range methods {
			// 3.检测是否存在特定的方法并提取信息
			if operation, exists := pathItem[method]; exists {
				interfaces = append(interfaces, map[string]any{
					"path":      path,
					"method":    method,
					"operation": operation,
				})
			}
		}
	}

	// 4.遍历提取到的所有接口并校验信息，涵盖operationId唯一标识，parameters参数
	var operationIDs []string
	processedPaths := make(map[string]PathItem)

	for _, interfaceItem := range interfaces {
		path := interfaceItem["path"].(string)
		method := interfaceItem["method"].(string)
		operationInterface := interfaceItem["operation"]

		operation, ok := operationInterface.(map[string]any)
		if !ok {
			return &ValidateError{Message: "operation必须是对象"}
		}

		// 5.校验description/operationId/parameters字段
		description, ok := operation["description"].(string)
		if !ok || description == "" {
			return &ValidateError{Message: "description不能为空且为字符串"}
		}

		operationID, ok := operation["operationId"].(string)
		if !ok || operationID == "" {
			return &ValidateError{Message: "operationId不能为空且为字符串"}
		}

		// 6.检测operationId是否是唯一的
		for _, existingID := range operationIDs {
			if existingID == operationID {
				return &ValidateError{Message: fmt.Sprintf("operationId必须唯一，%s出现重复", operationID)}
			}
		}
		operationIDs = append(operationIDs, operationID)

		// 7.处理parameters参数
		var parameters []Parameter
		if parametersInterface, exists := operation["parameters"]; exists {
			parametersList, ok := parametersInterface.([]any)
			if !ok {
				return &ValidateError{Message: "parameters必须是列表或者为空"}
			}

			for _, paramInterface := range parametersList {
				param, ok := paramInterface.(map[string]any)
				if !ok {
					return &ValidateError{Message: "parameter必须是对象"}
				}

				// 8.校验name/in/description/required/type参数是否存在，并且类型正确
				name, ok := param["name"].(string)
				if !ok || name == "" {
					return &ValidateError{Message: "parameter.name参数必须为字符串且不为空"}
				}

				paramDescription, ok := param["description"].(string)
				if !ok || paramDescription == "" {
					return &ValidateError{Message: "parameter.description参数必须为字符串且不为空"}
				}

				required, ok := param["required"].(bool)
				if !ok {
					return &ValidateError{Message: "parameter.required参数必须为布尔值且不为空"}
				}

				inValue, ok := param["in"].(string)
				if !ok || !isValidParameterIn(inValue) {
					validIns := []string{
						string(ParameterInPath),
						string(ParameterInQuery),
						string(ParameterInHeader),
						string(ParameterInCookie),
						string(ParameterInRequestBody),
					}
					return &ValidateError{Message: fmt.Sprintf("parameter.in参数必须为%s", strings.Join(validIns, "/"))}
				}

				paramType, ok := param["type"].(string)
				if !ok || !isValidParameterType(paramType) {
					validTypes := []string{
						string(ParameterTypeStr),
						string(ParameterTypeInt),
						string(ParameterTypeFloat),
						string(ParameterTypeBool),
					}
					return &ValidateError{Message: fmt.Sprintf("parameter.type参数必须为%s", strings.Join(validTypes, "/"))}
				}

				parameters = append(parameters, Parameter{
					Name:        name,
					In:          ParameterIn(inValue),
					Description: paramDescription,
					Required:    required,
					Type:        paramType,
				})
			}
		}

		// 9.组装数据并更新
		processedOperation := &Operation{
			Description: description,
			OperationID: operationID,
			Parameters:  parameters,
		}

		pathItem := processedPaths[path]
		if method == "get" {
			pathItem.Get = processedOperation
		} else if method == "post" {
			pathItem.Post = processedOperation
		}
		processedPaths[path] = pathItem
	}

	s.Paths = processedPaths
	return nil
}

// isValidParameterIn 检查参数位置是否有效
func isValidParameterIn(in string) bool {
	validIns := []ParameterIn{
		ParameterInPath,
		ParameterInQuery,
		ParameterInHeader,
		ParameterInCookie,
		ParameterInRequestBody,
	}
	for _, validIn := range validIns {
		if string(validIn) == in {
			return true
		}
	}
	return false
}

// isValidParameterType 检查参数类型是否有效
func isValidParameterType(paramType string) bool {
	validTypes := []ParameterType{
		ParameterTypeStr,
		ParameterTypeInt,
		ParameterTypeFloat,
		ParameterTypeBool,
	}
	for _, validType := range validTypes {
		if string(validType) == paramType {
			return true
		}
	}
	return false
}
