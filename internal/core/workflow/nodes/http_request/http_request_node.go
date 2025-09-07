package http_request

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
	"github.com/crazyfrankie/voidx/internal/core/workflow/nodes"
	"github.com/crazyfrankie/voidx/pkg/sonic"
)

// HttpRequestNode HTTP请求节点
type HttpRequestNode struct {
	*nodes.BaseNodeImpl
	nodeData *HttpRequestNodeData
}

// NewHttpRequestNode 创建新的HTTP请求节点
func NewHttpRequestNode(nodeData *HttpRequestNodeData) *HttpRequestNode {
	return &HttpRequestNode{
		BaseNodeImpl: nodes.NewBaseNodeImpl(nodeData.BaseNodeData),
		nodeData:     nodeData,
	}
}

// Invoke HTTP请求节点调用函数，向指定的URL发起请求并获取响应
func (h *HttpRequestNode) Invoke(state *entities.WorkflowState) (*entities.WorkflowState, error) {
	startAt := time.Now()

	// 1. 提取节点输入变量字典
	inputsDict := h.extractVariablesFromState(state)

	// 2. 提取数据，涵盖params、headers、body的数据
	params := make(map[string]string)
	headers := make(map[string]string)
	body := make(map[string]any)

	for _, input := range h.nodeData.Inputs {
		if inputType, exists := input.Meta["type"]; exists {
			if value, valueExists := inputsDict[input.Name]; valueExists {
				switch inputType {
				case string(HttpRequestInputTypeParams):
					params[input.Name] = fmt.Sprintf("%v", value)
				case string(HttpRequestInputTypeHeaders):
					headers[input.Name] = fmt.Sprintf("%v", value)
				case string(HttpRequestInputTypeBody):
					body[input.Name] = value
				}
			}
		}
	}

	// 3. 发起HTTP请求
	response, err := h.makeHttpRequest(params, headers, body)
	if err != nil {
		nodeResult := entities.NewNodeResult(h.nodeData.BaseNodeData)
		nodeResult.Status = entities.NodeStatusFailed
		nodeResult.Error = err.Error()
		nodeResult.Latency = time.Since(startAt)

		newState := &entities.WorkflowState{
			Inputs:      state.Inputs,
			Outputs:     state.Outputs,
			NodeResults: append(state.NodeResults, nodeResult),
		}

		return newState, err
	}

	// 4. 构建输出数据结构
	outputs := map[string]any{
		"text":        response.Text,
		"status_code": response.StatusCode,
	}

	// 5. 构建节点结果
	nodeResult := entities.NewNodeResult(h.nodeData.BaseNodeData)
	nodeResult.Status = entities.NodeStatusSucceeded
	nodeResult.Inputs = map[string]any{
		"params":  params,
		"headers": headers,
		"body":    body,
	}
	nodeResult.Outputs = outputs
	nodeResult.Latency = time.Since(startAt)

	// 6. 构建新状态
	newState := &entities.WorkflowState{
		Inputs:      state.Inputs,
		Outputs:     state.Outputs,
		NodeResults: append(state.NodeResults, nodeResult),
	}

	return newState, nil
}

// HttpResponse HTTP响应结构
type HttpResponse struct {
	Text       string
	StatusCode int
}

// makeHttpRequest 发起HTTP请求
func (h *HttpRequestNode) makeHttpRequest(params map[string]string, headers map[string]string, body map[string]any) (*HttpResponse, error) {
	// 构建URL和查询参数
	requestURL := h.nodeData.URL
	if len(params) > 0 {
		values := url.Values{}
		for key, value := range params {
			values.Add(key, value)
		}
		if strings.Contains(requestURL, "?") {
			requestURL += "&" + values.Encode()
		} else {
			requestURL += "?" + values.Encode()
		}
	}

	// 准备请求体
	var requestBody io.Reader
	if h.nodeData.Method != HttpRequestMethodGet && len(body) > 0 {
		bodyBytes, err := sonic.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %v", err)
		}
		requestBody = bytes.NewBuffer(bodyBytes)

		// 如果没有设置Content-Type，默认设置为application/json
		if _, exists := headers["Content-Type"]; !exists {
			headers["Content-Type"] = "application/json"
		}
	}

	// 创建HTTP请求
	req, err := http.NewRequest(string(h.nodeData.Method), requestURL, requestBody)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 发起请求
	client := &http.Client{
		Timeout: 30 * time.Second, // 30秒超时
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}

	return &HttpResponse{
		Text:       string(responseBody),
		StatusCode: resp.StatusCode,
	}, nil
}

// extractVariablesFromState 从状态中提取变量
func (h *HttpRequestNode) extractVariablesFromState(state *entities.WorkflowState) map[string]any {
	result := make(map[string]any)

	for _, input := range h.nodeData.Inputs {
		inputs := make(map[string]any)
		if err := sonic.UnmarshalString(state.Inputs, &inputs); err != nil {
			return nil
		}
		if val, exists := inputs[input.Name]; exists {
			result[input.Name] = val
		}
	}

	return result
}
