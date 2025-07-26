package entities

// CategoryEntity represents a category for built-in applications
type CategoryEntity struct {
	// Category represents the unique identifier of the category
	Category string `json:"category"`

	// Name represents the display name of the category
	Name string `json:"name"`
}