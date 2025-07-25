package time

import (
	"time"
)

// CurrentTimeTool represents a tool for getting current time
type CurrentTimeTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// NewCurrentTimeTool creates a new CurrentTimeTool instance
func NewCurrentTimeTool() *CurrentTimeTool {
	return &CurrentTimeTool{
		Name:        "current_time",
		Description: "一个用于获取当前时间的工具",
	}
}

// Run executes the current time tool
func (t *CurrentTimeTool) Run(args map[string]interface{}) (interface{}, error) {
	return time.Now().Format("2006-01-02 15:04:05 MST"), nil
}

// CurrentTime is the exported function for dynamic loading
func CurrentTime(args map[string]interface{}) (interface{}, error) {
	tool := NewCurrentTimeTool()
	return tool.Run(args)
}
