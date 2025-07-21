package documentloaders

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

// JSON represents a JSON document loader that can handle various JSON structures.
type JSON struct {
	r            io.Reader
	jqSchema     string   // JSONPath-like schema for extracting specific fields
	contentField string   // Primary field to use as content
	metadataKeys []string // Keys to extract as metadata
}

var _ documentloaders.Loader = &JSON{}

// NewJSON creates a new JSON loader with an io.Reader.
// By default, it will convert the entire JSON to a readable text format.
func NewJSON(r io.Reader) JSON {
	return JSON{
		r: r,
	}
}

// NewJSONWithSchema creates a new JSON loader with specific field extraction.
// contentField: the JSON field to use as main content (e.g., "content", "text", "body")
// metadataKeys: JSON fields to extract as metadata
func NewJSONWithSchema(r io.Reader, contentField string, metadataKeys ...string) JSON {
	return JSON{
		r:            r,
		contentField: contentField,
		metadataKeys: metadataKeys,
	}
}

// Load reads from the io.Reader and returns documents based on JSON structure.
func (j *JSON) Load(_ context.Context) ([]schema.Document, error) {
	data, err := io.ReadAll(j.r)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON data: %w", err)
	}

	// Try to parse as JSON
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return j.processJSONData(jsonData)
}

// processJSONData converts JSON data to documents based on its structure
func (j *JSON) processJSONData(data interface{}) ([]schema.Document, error) {
	switch v := data.(type) {
	case []interface{}:
		// Handle JSON array - create one document per array element
		return j.processJSONArray(v)
	case map[string]interface{}:
		// Handle JSON object - create one document
		return j.processJSONObject(v)
	default:
		// Handle primitive values
		return []schema.Document{
			{
				PageContent: fmt.Sprintf("%v", v),
				Metadata:    map[string]any{"type": reflect.TypeOf(v).String()},
			},
		}, nil
	}
}

// processJSONArray processes a JSON array, creating one document per element
func (j *JSON) processJSONArray(arr []interface{}) ([]schema.Document, error) {
	var docs []schema.Document

	for i, item := range arr {
		switch itemData := item.(type) {
		case map[string]interface{}:
			doc, err := j.createDocumentFromObject(itemData)
			if err != nil {
				return nil, fmt.Errorf("failed to process array item %d: %w", i, err)
			}
			// Add array index to metadata
			doc.Metadata["array_index"] = i
			docs = append(docs, doc)
		default:
			// Handle primitive array elements
			docs = append(docs, schema.Document{
				PageContent: fmt.Sprintf("%v", itemData),
				Metadata: map[string]any{
					"array_index": i,
					"type":        reflect.TypeOf(itemData).String(),
				},
			})
		}
	}

	return docs, nil
}

// processJSONObject processes a single JSON object
func (j *JSON) processJSONObject(obj map[string]interface{}) ([]schema.Document, error) {
	doc, err := j.createDocumentFromObject(obj)
	if err != nil {
		return nil, err
	}
	return []schema.Document{doc}, nil
}

// createDocumentFromObject creates a document from a JSON object
func (j *JSON) createDocumentFromObject(obj map[string]interface{}) (schema.Document, error) {
	var content string
	metadata := make(map[string]any)

	// If contentField is specified, use it as primary content
	if j.contentField != "" {
		if val, exists := obj[j.contentField]; exists {
			content = fmt.Sprintf("%v", val)
		}
	}

	// Extract specified metadata fields
	for _, key := range j.metadataKeys {
		if val, exists := obj[key]; exists {
			metadata[key] = val
		}
	}

	// If no specific content field or it's empty, convert entire object to readable format
	if content == "" {
		content = j.objectToReadableText(obj)
	}

	// If no specific metadata keys specified, add all non-content fields as metadata
	if len(j.metadataKeys) == 0 {
		for key, val := range obj {
			if key != j.contentField {
				metadata[key] = val
			}
		}
	}

	return schema.Document{
		PageContent: content,
		Metadata:    metadata,
	}, nil
}

// objectToReadableText converts a JSON object to a human-readable text format
func (j *JSON) objectToReadableText(obj map[string]interface{}) string {
	var lines []string

	// Sort keys for consistent output
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := obj[key]
		line := j.formatKeyValue(key, value)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// formatKeyValue formats a key-value pair for readable output
func (j *JSON) formatKeyValue(key string, value interface{}) string {
	switch v := value.(type) {
	case map[string]interface{}:
		// For nested objects, format them with indentation
		nested := j.objectToReadableText(v)
		indented := strings.ReplaceAll(nested, "\n", "\n  ")
		return fmt.Sprintf("%s:\n  %s", key, indented)
	case []interface{}:
		// For arrays, check if they contain objects that should be expanded
		return j.formatArray(key, v)
	default:
		return fmt.Sprintf("%s: %v", key, v)
	}
}

// formatArray intelligently formats arrays based on their content
func (j *JSON) formatArray(key string, arr []interface{}) string {
	if len(arr) == 0 {
		return fmt.Sprintf("%s: []", key)
	}

	// Check if array contains objects that should be expanded
	hasComplexObjects := false
	for _, item := range arr {
		if obj, ok := item.(map[string]interface{}); ok {
			// Consider it complex if it has more than 2 fields or contains nested structures
			if len(obj) > 2 || j.hasNestedStructures(obj) {
				hasComplexObjects = true
				break
			}
		}
	}

	if hasComplexObjects {
		// Format as expanded multi-line structure
		var lines []string
		lines = append(lines, fmt.Sprintf("%s:", key))
		for i, item := range arr {
			if obj, ok := item.(map[string]interface{}); ok {
				lines = append(lines, fmt.Sprintf("  [%d]:", i))
				nested := j.objectToReadableText(obj)
				indented := strings.ReplaceAll(nested, "\n", "\n    ")
				lines = append(lines, fmt.Sprintf("    %s", indented))
			} else {
				lines = append(lines, fmt.Sprintf("  [%d]: %v", i, item))
			}
		}
		return strings.Join(lines, "\n")
	} else {
		// Format as simple comma-separated list
		var elements []string
		for _, item := range arr {
			if obj, ok := item.(map[string]interface{}); ok {
				// For simple objects, create a compact representation
				elements = append(elements, j.compactObjectString(obj))
			} else {
				elements = append(elements, fmt.Sprintf("%v", item))
			}
		}
		return fmt.Sprintf("%s: [%s]", key, strings.Join(elements, ", "))
	}
}

// hasNestedStructures checks if an object contains nested objects or arrays
func (j *JSON) hasNestedStructures(obj map[string]interface{}) bool {
	for _, value := range obj {
		switch value.(type) {
		case map[string]interface{}, []interface{}:
			return true
		}
	}
	return false
}

// compactObjectString creates a compact string representation for simple objects
func (j *JSON) compactObjectString(obj map[string]interface{}) string {
	var parts []string

	// Sort keys for consistency
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := obj[key]
		// Only include simple values in compact format
		switch v := value.(type) {
		case string, int, int64, float64, bool:
			parts = append(parts, fmt.Sprintf("%s:%v", key, v))
		}
	}

	if len(parts) > 0 {
		return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
	}
	return fmt.Sprintf("%v", obj)
}

// LoadAndSplit reads JSON data from the io.Reader and splits it into multiple
// documents using a text splitter.
func (j *JSON) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := j.Load(ctx)
	if err != nil {
		return nil, err
	}

	return textsplitter.SplitDocuments(splitter, docs)
}
