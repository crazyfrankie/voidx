package entities

// ToolParamType represents the type of a tool parameter
type ToolParamType string

const (
	// ToolParamTypeString represents a string parameter type
	ToolParamTypeString ToolParamType = "string"

	// ToolParamTypeNumber represents a number parameter type
	ToolParamTypeNumber ToolParamType = "number"

	// ToolParamTypeBoolean represents a boolean parameter type
	ToolParamTypeBoolean ToolParamType = "boolean"

	// ToolParamTypeSelect represents a select parameter type
	ToolParamTypeSelect ToolParamType = "select"
)

// ToolParam represents a tool parameter configuration
type ToolParam struct {
	// Name is the actual parameter name
	Name string `json:"name"`

	// Label is the display label for the parameter
	Label string `json:"label"`

	// Type is the parameter type
	Type ToolParamType `json:"type"`

	// Required indicates if the parameter is mandatory
	Required bool `json:"required"`

	// Default is the default value for the parameter
	Default interface{} `json:"default,omitempty"`

	// Min is the minimum value for number parameters
	Min *float64 `json:"min,omitempty"`

	// Max is the maximum value for number parameters
	Max *float64 `json:"max,omitempty"`

	// Options is the list of options for select parameters
	Options []map[string]any `json:"options,omitempty"`
}

// ToolEntity represents a tool configuration
type ToolEntity struct {
	// Name is the tool's name
	Name string `json:"name"`

	// Label is the tool's display label
	Label string `json:"label"`

	// Description is the tool's description
	Description string `json:"description"`

	// Params is the list of tool parameters
	Params []ToolParam `json:"params"`
}
