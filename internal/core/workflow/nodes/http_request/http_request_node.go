package http_request

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
	"github.com/crazyfrankie/voidx/pkg/sonic"
)

// HTTPRequestNode represents an HTTP request workflow node
type HTTPRequestNode struct {
	nodeData   *HTTPRequestNodeData
	httpClient *http.Client
}

// NewHTTPRequestNode creates a new HTTP request node instance
func NewHTTPRequestNode(nodeData *HTTPRequestNodeData) *HTTPRequestNode {
	return &HTTPRequestNode{
		nodeData: nodeData,
		httpClient: &http.Client{
			Timeout: time.Duration(nodeData.Timeout) * time.Second,
		},
	}
}

// Execute executes the HTTP request node with the given workflow state
func (n *HTTPRequestNode) Execute(ctx context.Context, state *entities.WorkflowState) (*entities.NodeResult, error) {
	startTime := time.Now()

	// Create node result
	result := entities.NewNodeResult(n.nodeData.BaseNodeData)
	result.StartTime = startTime.Unix()

	// Extract input variables from state
	inputsDict, err := n.extractVariablesFromState(state)
	if err != nil {
		result.Status = entities.NodeStatusFailed
		result.Error = fmt.Sprintf("failed to extract input variables: %v", err)
		result.EndTime = time.Now().Unix()
		return result, err
	}
	result.Inputs = inputsDict

	// Render URL template
	url, err := n.renderTemplate(n.nodeData.URL, inputsDict)
	if err != nil {
		result.Status = entities.NodeStatusFailed
		result.Error = fmt.Sprintf("failed to render URL template: %v", err)
		result.EndTime = time.Now().Unix()
		return result, err
	}

	// Render body template
	body, err := n.renderTemplate(n.nodeData.Body, inputsDict)
	if err != nil {
		result.Status = entities.NodeStatusFailed
		result.Error = fmt.Sprintf("failed to render body template: %v", err)
		result.EndTime = time.Now().Unix()
		return result, err
	}

	// Make HTTP request
	response, err := n.makeHTTPRequest(ctx, url, body, inputsDict)
	if err != nil {
		result.Status = entities.NodeStatusFailed
		result.Error = fmt.Sprintf("HTTP request failed: %v", err)
		result.EndTime = time.Now().Unix()
		return result, err
	}

	// Build output data structure
	outputs := make(map[string]interface{})
	if len(n.nodeData.Outputs) > 0 {
		outputs[n.nodeData.Outputs[0].Name] = response
	} else {
		outputs["response"] = response
	}
	result.Outputs = outputs

	// Set success status
	result.Status = entities.NodeStatusSucceeded
	result.EndTime = time.Now().Unix()

	return result, nil
}

// extractVariablesFromState extracts input variables from the workflow state
func (n *HTTPRequestNode) extractVariablesFromState(state *entities.WorkflowState) (map[string]interface{}, error) {
	inputsDict := make(map[string]interface{})

	for _, input := range n.nodeData.Inputs {
		var value interface{}
		var found bool

		// Check if it's a reference to another node's output
		if input.Value.Type == entities.VariableValueTypeRef {
			if content, ok := input.Value.Content.(*entities.VariableContent); ok {
				if content.RefNodeID != nil {
					// Find the referenced node's output in state
					for _, nodeResult := range state.NodeResults {
						if nodeResult.NodeID == *content.RefNodeID {
							if refValue, exists := nodeResult.Outputs[content.RefVarName]; exists {
								value = refValue
								found = true
								break
							}
						}
					}
				}
			}
		} else {
			// It's a constant value
			value = input.Value.Content
			found = true
		}

		if !found && input.Required {
			return nil, fmt.Errorf("required input variable %s not found", input.Name)
		}

		if found {
			inputsDict[input.Name] = value
		}
	}

	// Also include workflow inputs
	for key, value := range state.Inputs {
		if _, exists := inputsDict[key]; !exists {
			inputsDict[key] = value
		}
	}

	return inputsDict, nil
}

// renderTemplate renders a template string with the given variables
func (n *HTTPRequestNode) renderTemplate(templateStr string, variables map[string]interface{}) (string, error) {
	if templateStr == "" {
		return "", nil
	}

	// Use Go's text/template to render the template
	tmpl, err := template.New("http").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, variables); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return result.String(), nil
}

// makeHTTPRequest makes the actual HTTP request
func (n *HTTPRequestNode) makeHTTPRequest(ctx context.Context, url, body string, variables map[string]interface{}) (map[string]interface{}, error) {
	// Create request body
	var requestBody io.Reader
	if body != "" {
		requestBody = bytes.NewBufferString(body)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, string(n.nodeData.Method), url, requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	for key, value := range n.nodeData.Headers {
		// Render header value template
		renderedValue, err := n.renderTemplate(value, variables)
		if err != nil {
			return nil, fmt.Errorf("failed to render header template for %s: %w", key, err)
		}
		req.Header.Set(key, renderedValue)
	}

	// Set default content type for POST/PUT/PATCH requests
	if (n.nodeData.Method == HTTPMethodPOST || n.nodeData.Method == HTTPMethodPUT || n.nodeData.Method == HTTPMethodPATCH) && body != "" {
		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	// Make the request
	resp, err := n.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response body as JSON if possible
	var responseData interface{}
	if err := sonic.Unmarshal(responseBody, &responseData); err != nil {
		// If JSON parsing fails, use raw string
		responseData = string(responseBody)
	}

	// Build response object
	response := map[string]interface{}{
		"status_code": resp.StatusCode,
		"headers":     resp.Header,
		"body":        responseData,
		"url":         url,
		"method":      string(n.nodeData.Method),
	}

	return response, nil
}

// GetNodeData returns the node data
func (n *HTTPRequestNode) GetNodeData() *HTTPRequestNodeData {
	return n.nodeData
}
