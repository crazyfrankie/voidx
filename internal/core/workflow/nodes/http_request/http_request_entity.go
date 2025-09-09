package http_request

import (
	"github.com/crazyfrankie/voidx/internal/core/workflow/entities"
)

// HTTPMethod represents HTTP request methods
type HTTPMethod string

const (
	HTTPMethodGET    HTTPMethod = "GET"
	HTTPMethodPOST   HTTPMethod = "POST"
	HTTPMethodPUT    HTTPMethod = "PUT"
	HTTPMethodDELETE HTTPMethod = "DELETE"
	HTTPMethodPATCH  HTTPMethod = "PATCH"
)

// HTTPRequestNodeData represents the data structure for HTTP request workflow nodes
type HTTPRequestNodeData struct {
	*entities.BaseNodeData
	URL     string                     `json:"url"`
	Method  HTTPMethod                 `json:"method"`
	Headers map[string]string          `json:"headers"`
	Body    string                     `json:"body"`
	Timeout int                        `json:"timeout"` // timeout in seconds
	Inputs  []*entities.VariableEntity `json:"inputs"`
	Outputs []*entities.VariableEntity `json:"outputs"`
}

// NewHTTPRequestNodeData creates a new HTTP request node data instance
func NewHTTPRequestNodeData() *HTTPRequestNodeData {
	return &HTTPRequestNodeData{
		BaseNodeData: &entities.BaseNodeData{NodeType: entities.NodeTypeHTTPRequest},
		Method:       HTTPMethodGET,
		Headers:      make(map[string]string),
		Timeout:      30,
		Inputs:       make([]*entities.VariableEntity, 0),
		Outputs:      make([]*entities.VariableEntity, 0),
	}
}

// GetBaseNodeData returns the base node data (implements NodeDataInterface)
func (h *HTTPRequestNodeData) GetBaseNodeData() *entities.BaseNodeData {
	return h.BaseNodeData
}
