package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/crazyfrankie/voidx/internal/core/tools/entities"
)

// ParameterIn represents where the parameter should be placed in the request
type ParameterIn string

const (
	ParameterInPath        ParameterIn = "path"
	ParameterInQuery       ParameterIn = "query"
	ParameterInHeader      ParameterIn = "header"
	ParameterInCookie      ParameterIn = "cookie"
	ParameterInRequestBody ParameterIn = "request_body"
)

// ParameterTypeMap maps parameter types to Go types
var ParameterTypeMap = map[string]string{
	"string":  "string",
	"integer": "int",
	"number":  "float64",
	"boolean": "bool",
}

// APIProviderManager API工具提供者管理器，能根据传递的工具配置信息生成自定义工具
type APIProviderManager struct {
	httpClient *http.Client
}

// NewAPIProviderManager 创建API工具提供者管理器
func NewAPIProviderManager() *APIProviderManager {
	return &APIProviderManager{
		httpClient: &http.Client{},
	}
}

// GetTool 根据传递的配置获取自定义API工具
func (apm *APIProviderManager) GetTool(ctx context.Context, toolEntity *entities.APIToolEntity) (tool.InvokableTool, error) {
	return &APITool{
		entity:     toolEntity,
		httpClient: apm.httpClient,
	}, nil
}

// APITool 实现了eino的InvokableTool接口的API工具
type APITool struct {
	entity     *entities.APIToolEntity
	httpClient *http.Client
}

// Info 返回工具信息
func (at *APITool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	// Convert parameters to eino schema
	params := make(map[string]*schema.ParameterInfo)

	for _, param := range at.entity.Parameters {
		paramName, _ := param["name"].(string)
		paramType, _ := param["type"].(string)
		paramDesc, _ := param["description"].(string)
		paramRequired, _ := param["required"].(bool)

		// Map parameter type to eino DataType
		var dataType schema.DataType
		switch paramType {
		case "string":
			dataType = schema.String
		case "integer":
			dataType = schema.Integer
		case "number":
			dataType = schema.Number
		case "boolean":
			dataType = schema.Boolean
		default:
			dataType = schema.String
		}

		params[paramName] = &schema.ParameterInfo{
			Type:     dataType,
			Desc:     paramDesc,
			Required: paramRequired,
		}
	}

	var paramsOneOf *schema.ParamsOneOf
	if len(params) > 0 {
		paramsOneOf = schema.NewParamsOneOfByParams(params)
	}

	return &schema.ToolInfo{
		Name:        fmt.Sprintf("%s_%s", at.entity.ID, at.entity.Name),
		Desc:        at.entity.Description,
		ParamsOneOf: paramsOneOf,
	}, nil
}

// InvokableRun 执行API调用
func (at *APITool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	// Parse arguments
	var args map[string]interface{}
	if argumentsInJSON != "" {
		if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
			return "", fmt.Errorf("failed to parse arguments: %w", err)
		}
	}

	// Organize parameters by location
	parameters := map[ParameterIn]map[string]interface{}{
		ParameterInPath:        make(map[string]interface{}),
		ParameterInQuery:       make(map[string]interface{}),
		ParameterInHeader:      make(map[string]interface{}),
		ParameterInCookie:      make(map[string]interface{}),
		ParameterInRequestBody: make(map[string]interface{}),
	}

	// Create parameter map for easy lookup
	parameterMap := make(map[string]map[string]interface{})
	for _, param := range at.entity.Parameters {
		if name, ok := param["name"].(string); ok {
			parameterMap[name] = param
		}
	}

	// Organize arguments by parameter location
	for key, value := range args {
		if param, exists := parameterMap[key]; exists {
			location := ParameterInQuery // default location
			if in, ok := param["in"].(string); ok {
				location = ParameterIn(in)
			}
			parameters[location][key] = value
		}
	}

	// Build headers
	headers := make(map[string]string)
	for _, header := range at.entity.Headers {
		headers[header.Key] = header.Value
	}
	for key, value := range parameters[ParameterInHeader] {
		headers[key] = fmt.Sprintf("%v", value)
	}

	// Build URL with path parameters
	url := at.entity.URL
	for key, value := range parameters[ParameterInPath] {
		placeholder := fmt.Sprintf("{%s}", key)
		url = strings.ReplaceAll(url, placeholder, fmt.Sprintf("%v", value))
	}

	// Create request
	var req *http.Request
	var err error

	if len(parameters[ParameterInRequestBody]) > 0 {
		// If there's a request body, marshal it to JSON
		bodyData, err := json.Marshal(parameters[ParameterInRequestBody])
		if err != nil {
			return "", fmt.Errorf("failed to marshal request body: %w", err)
		}
		req, err = http.NewRequestWithContext(ctx, strings.ToUpper(at.entity.Method), url, strings.NewReader(string(bodyData)))
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}
		headers["Content-Type"] = "application/json"
	} else {
		req, err = http.NewRequestWithContext(ctx, strings.ToUpper(at.entity.Method), url, nil)
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}
	}

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Add query parameters
	if len(parameters[ParameterInQuery]) > 0 {
		q := req.URL.Query()
		for key, value := range parameters[ParameterInQuery] {
			q.Add(key, fmt.Sprintf("%v", value))
		}
		req.URL.RawQuery = q.Encode()
	}

	// Add cookies
	for key, value := range parameters[ParameterInCookie] {
		req.AddCookie(&http.Cookie{
			Name:  key,
			Value: fmt.Sprintf("%v", value),
		})
	}

	// Execute request
	resp, err := at.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	var res []byte
	if res, err = io.ReadAll(resp.Body); err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(res), nil
}
