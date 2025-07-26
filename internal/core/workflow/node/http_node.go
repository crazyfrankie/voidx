package node

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/crazyfrankie/voidx/internal/models/entity"
)

// HTTPNodeData HTTP请求节点数据
type HTTPNodeData struct {
	*entity.BaseNodeData
	Method     string            `json:"method"`     // HTTP方法
	URL        string            `json:"url"`        // 请求URL
	Headers    map[string]string `json:"headers"`    // 请求头
	Body       string            `json:"body"`       // 请求体模板
	BodyType   string            `json:"body_type"`  // 请求体类型：json, form, raw
	OutputKey  string            `json:"output_key"` // 输出键
	Timeout    int               `json:"timeout"`    // 超时时间（秒）
	InputKeys  map[string]string `json:"input_keys"` // 输入键映射
	Parameters map[string]any    `json:"parameters"` // 静态参数
}

// HTTPNode HTTP请求节点
type HTTPNode struct {
	BaseNode
	Data   *HTTPNodeData
	client *http.Client
}

// NewHTTPNode 创建HTTP请求节点
func NewHTTPNode(data *HTTPNodeData) *HTTPNode {
	timeout := time.Duration(data.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second // 默认30秒超时
	}

	return &HTTPNode{
		BaseNode: BaseNode{Data: data.BaseNodeData},
		Data:     data,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Invoke 执行HTTP请求节点
func (n *HTTPNode) Invoke(ctx context.Context, state map[string]any) (map[string]any, error) {
	// 复制当前状态
	result := make(map[string]any)
	for k, v := range state {
		result[k] = v
	}

	// 准备请求参数
	requestData := make(map[string]any)

	// 从状态中获取输入参数
	for paramName, stateKey := range n.Data.InputKeys {
		if value, ok := state[stateKey]; ok {
			requestData[paramName] = value
		}
	}

	// 合并静态参数
	for k, v := range n.Data.Parameters {
		requestData[k] = v
	}

	// 准备请求体
	var requestBody io.Reader
	if n.Data.Body != "" {
		// 这里可以实现模板替换逻辑
		bodyStr := n.Data.Body

		// 简单的参数替换（实际项目中应该使用模板引擎）
		if n.Data.BodyType == "json" {
			bodyBytes, err := json.Marshal(requestData)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request data: %w", err)
			}
			requestBody = bytes.NewReader(bodyBytes)
		} else {
			requestBody = bytes.NewReader([]byte(bodyStr))
		}
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, n.Data.Method, n.Data.URL, requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// 设置请求头
	for key, value := range n.Data.Headers {
		req.Header.Set(key, value)
	}

	// 设置Content-Type
	if n.Data.BodyType == "json" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// 发送请求
	resp, err := n.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 构建响应结果
	httpResult := map[string]any{
		"status_code": resp.StatusCode,
		"headers":     resp.Header,
		"body":        string(responseBody),
		"success":     resp.StatusCode >= 200 && resp.StatusCode < 300,
	}

	// 尝试解析JSON响应
	var jsonResponse interface{}
	if err := json.Unmarshal(responseBody, &jsonResponse); err == nil {
		httpResult["json"] = jsonResponse
	}

	// 设置输出
	outputKey := n.Data.OutputKey
	if outputKey == "" {
		outputKey = "http_result"
	}
	result[outputKey] = httpResult

	return result, nil
}

// Validate 验证HTTP请求节点配置
func (n *HTTPNode) Validate() error {
	if n.Data.Type != entity.NodeTypeHTTPRequest {
		return errors.New("invalid node type for HTTP request node")
	}

	if n.Data.URL == "" {
		return errors.New("url is required for HTTP request node")
	}

	if n.Data.Method == "" {
		n.Data.Method = "GET" // 默认GET方法
	}

	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	valid := false
	for _, method := range validMethods {
		if n.Data.Method == method {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid HTTP method: %s", n.Data.Method)
	}

	if n.Data.Headers == nil {
		n.Data.Headers = make(map[string]string)
	}

	if n.Data.InputKeys == nil {
		n.Data.InputKeys = make(map[string]string)
	}

	if n.Data.Parameters == nil {
		n.Data.Parameters = make(map[string]any)
	}

	if n.Data.Timeout <= 0 {
		n.Data.Timeout = 30 // 默认30秒
	}

	return nil
}
