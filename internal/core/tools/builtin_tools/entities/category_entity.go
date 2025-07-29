package entities

import (
	"fmt"
	"strings"
)

// CategoryEntity represents a tool category
type CategoryEntity struct {
	// Category is the unique identifier for the category
	Category string `yaml:"category" json:"category"`

	// Name is the display name of the category
	Name string `yaml:"name" json:"name"`

	// Icon is the name of the category's icon file
	Icon string `yaml:"icon" json:"icon"`
}

// Validate checks if the CategoryEntity is valid
func (c *CategoryEntity) Validate() error {
	// Check if icon has .svg extension
	if !strings.HasSuffix(c.Icon, ".svg") {
		return fmt.Errorf("category icon must be in SVG format")
	}
	return nil
}
