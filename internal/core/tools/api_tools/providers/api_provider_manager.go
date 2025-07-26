package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/crazyfrankie/voidx/internal/core/tools/api_tools/entities"
)

// ApiProviderManager API工具提供者管理器，能根据传递的工具配置信息生成自定义API工具
type ApiProviderManager struct{}

// NewApiProviderManager 创建新的API提供者管理器
func NewApiProviderManager() *ApiProviderManager {
	return &ApiProviderManager{}
}

// ToolFunc API工具函数类型
type ToolFunc func(args map[string]any) (string, error)

// ToolSchema 工具模式定义
type ToolSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// Tool 工具接口
type Tool interface {
	GetName() string
	GetDescription() string
	GetSchema() *ToolSchema
	Execute(args map[string]any) (string, error)
}

// ApiTool API工具实现
type ApiTool struct {
	entity   *entities.ToolEntity
	toolFunc ToolFunc
	schema   *ToolSchema
}

// GetName 获取工具名称
func (t *ApiTool) GetName() string {
	return fmt.Sprintf("%s_%s", t.entity.ID, t.entity.Name)
}

// GetDescription 获取工具描述
func (t *ApiTool) GetDescription() string {
	return t.entity.Description
}

// GetSchema 获取工具模式
func (t *ApiTool) GetSchema() *ToolSchema {
	return t.schema
}

// Execute 执行工具
func (t *ApiTool) Execute(args map[string]any) (string, error) {
	return t.toolFunc(args)
}

// GetTool 根据传递的配置获取自定义API工具
func (m *ApiProviderManager) GetTool(toolEntity *entities.ToolEntity) Tool {
	toolFunc := m.createToolFuncFromToolEntity(toolEntity)
	schema := m.createSchemaFromParameters(toolEntity)

	return &ApiTool{
		entity:   toolEntity,
		toolFunc: toolFunc,
		schema:   schema,
	}
}

// createToolFuncFromToolEntity 根据传递的信息创建发起API请求的函数
func (m *ApiProviderManager) createToolFuncFromToolEntity(toolEntity *entities.ToolEntity) ToolFunc {
	return func(args map[string]any) (string, error) {
		// 1.定义变量存储来自path/query/header/cookie/request_body中的数据
		parameters := map[entities.ParameterIn]map[string]any{
			entities.ParameterInPath:        make(map[string]any),
			entities.ParameterInHeader:      make(map[string]any),
			entities.ParameterInQuery:       make(map[string]any),
			entities.ParameterInCookie:      make(map[string]any),
			entities.ParameterInRequestBody: make(map[string]any),
		}

		// 2.更改参数结构映射
		parameterMap := make(map[string]map[string]any)
		for _, parameter := range toolEntity.Parameters {
			if name, ok := parameter["name"].(string); ok {
				parameterMap[name] = parameter
			}
		}

		headerMap := make(map[string]string)
		for _, header := range toolEntity.Headers {
			headerMap[header.Key] = header.Value
		}

		// 3.循环遍历传递的所有字段并校验
		for key, value := range args {
			// 4.提取键值对关联的字段并校验
			parameter, exists := parameterMap[key]
			if !exists {
				continue
			}

			// 5.将参数存储到合适的位置上，默认在query上
			paramIn := entities.ParameterInQuery
			if inValue, ok := parameter["in"].(string); ok {
				paramIn = entities.ParameterIn(inValue)
			}
			parameters[paramIn][key] = value
		}

		// 6.构建request请求并返回采集的内容
		return m.makeHTTPRequest(toolEntity, parameters, headerMap)
	}
}

// makeHTTPRequest 发起HTTP请求
func (m *ApiProviderManager) makeHTTPRequest(
	toolEntity *entities.ToolEntity,
	parameters map[entities.ParameterIn]map[string]any,
	headerMap map[string]string,
) (string, error) {
	// 处理URL中的路径参数
	requestURL := toolEntity.URL
	for key, value := range parameters[entities.ParameterInPath] {
		placeholder := fmt.Sprintf("{%s}", key)
		requestURL = strings.ReplaceAll(requestURL, placeholder, fmt.Sprintf("%v", value))
	}

	// 构建查询参数
	if len(parameters[entities.ParameterInQuery]) > 0 {
		queryParams := url.Values{}
		for key, value := range parameters[entities.ParameterInQuery] {
			queryParams.Add(key, fmt.Sprintf("%v", value))
		}
		if strings.Contains(requestURL, "?") {
			requestURL += "&" + queryParams.Encode()
		} else {
			requestURL += "?" + queryParams.Encode()
		}
	}

	// 构建请求体
	var requestBody io.Reader
	if len(parameters[entities.ParameterInRequestBody]) > 0 {
		jsonData, err := json.Marshal(parameters[entities.ParameterInRequestBody])
		if err != nil {
			return "", fmt.Errorf("failed to marshal request body: %w", err)
		}
		requestBody = bytes.NewBuffer(jsonData)
	}

	// 创建HTTP请求
	req, err := http.NewRequest(strings.ToUpper(toolEntity.Method), requestURL, requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	for key, value := range headerMap {
		req.Header.Set(key, value)
	}
	for key, value := range parameters[entities.ParameterInHeader] {
		req.Header.Set(key, fmt.Sprintf("%v", value))
	}

	// 设置Cookie
	for key, value := range parameters[entities.ParameterInCookie] {
		req.AddCookie(&http.Cookie{
			Name:  key,
			Value: fmt.Sprintf("%v", value),
		})
	}

	// 如果有请求体，设置Content-Type
	if requestBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// 发起请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(responseBody), nil
}

// createSchemaFromParameters 根据传递的parameters参数创建工具模式
func (m *ApiProviderManager) createSchemaFromParameters(toolEntity *entities.ToolEntity) *ToolSchema {
	properties := make(map[string]any)
	var required []string

	for _, parameter := range toolEntity.Parameters {
		name, ok := parameter["name"].(string)
		if !ok {
			continue
		}

		paramType, ok := parameter["type"].(string)
		if !ok {
			paramType = "string"
		}

		description, ok := parameter["description"].(string)
		if !ok {
			description = ""
		}

		isRequired, ok := parameter["required"].(bool)
		if !ok {
			isRequired = true
		}

		// 转换参数类型
		jsonType := m.convertParameterType(paramType)

		properties[name] = map[string]any{
			"type":        jsonType,
			"description": description,
		}

		if isRequired {
			required = append(required, name)
		}
	}

	parameters := map[string]any{
		"type":       "object",
		"properties": properties,
	}

	if len(required) > 0 {
		parameters["required"] = required
	}

	return &ToolSchema{
		Name:        fmt.Sprintf("%s_%s", toolEntity.ID, toolEntity.Name),
		Description: toolEntity.Description,
		Parameters:  parameters,
	}
}

// convertParameterType 转换参数类型到JSON Schema类型
func (m *ApiProviderManager) convertParameterType(paramType string) string {
	switch paramType {
	case string(entities.ParameterTypeStr):
		return "string"
	case string(entities.ParameterTypeInt):
		return "integer"
	case string(entities.ParameterTypeFloat):
		return "number"
	case string(entities.ParameterTypeBool):
		return "boolean"
	default:
		return "string"
	}
}

// GetParameterTypeFromString 从字符串获取参数类型
func GetParameterTypeFromString(paramType string) reflect.Type {
	if t, exists := entities.ParameterTypeMap[entities.ParameterType(paramType)]; exists {
		return t
	}
	return reflect.TypeOf("")
}
