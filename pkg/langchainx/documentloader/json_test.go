package documentloaders

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/textsplitter"
)

func TestJSONLoader_SingleObject(t *testing.T) {
	t.Parallel()
	file, err := os.Open("./testdata/test.json")
	require.NoError(t, err)
	defer file.Close()

	loader := NewJSON(file)

	ctx := context.Background()
	docs, err := loader.Load(ctx)
	require.NoError(t, err)
	require.Len(t, docs, 1)

	doc := docs[0]

	// Check that content contains key information
	assert.Contains(t, doc.PageContent, "author: John Doe")
	assert.Contains(t, doc.PageContent, "content: This is a sample JSON document")
	assert.Contains(t, doc.PageContent, "title: Sample Document")

	// Check metadata
	assert.Equal(t, "John Doe", doc.Metadata["author"])
	assert.Equal(t, "Sample Document", doc.Metadata["title"])
	assert.Equal(t, true, doc.Metadata["published"])
}

func TestJSONLoader_WithSchema(t *testing.T) {
	t.Parallel()
	file, err := os.Open("./testdata/test.json")
	require.NoError(t, err)
	defer file.Close()

	// Use "content" as main content and extract specific metadata
	loader := NewJSONWithSchema(file, "content", "title", "author", "created_at")

	ctx := context.Background()
	docs, err := loader.Load(ctx)
	require.NoError(t, err)
	require.Len(t, docs, 1)

	doc := docs[0]

	// Check that only the content field is used as PageContent
	expectedContent := "This is a sample JSON document with various fields for testing the JSON document loader."
	assert.Equal(t, expectedContent, doc.PageContent)

	// Check that only specified fields are in metadata
	assert.Equal(t, "Sample Document", doc.Metadata["title"])
	assert.Equal(t, "John Doe", doc.Metadata["author"])
	assert.Equal(t, "2024-01-15T10:30:00Z", doc.Metadata["created_at"])

	// These should not be in metadata since they weren't specified
	assert.NotContains(t, doc.Metadata, "published")
	assert.NotContains(t, doc.Metadata, "word_count")
}

func TestJSONLoader_Array(t *testing.T) {
	t.Parallel()
	file, err := os.Open("./testdata/test_array.json")
	require.NoError(t, err)
	defer file.Close()

	loader := NewJSON(file)

	ctx := context.Background()
	docs, err := loader.Load(ctx)
	require.NoError(t, err)
	require.Len(t, docs, 3)

	// Check first document
	doc1 := docs[0]
	assert.Contains(t, doc1.PageContent, "title: First Article")
	assert.Contains(t, doc1.PageContent, "author: Alice Smith")
	assert.Equal(t, 0, doc1.Metadata["array_index"])
	assert.Equal(t, float64(1), doc1.Metadata["id"]) // JSON numbers are float64

	// Check second document
	doc2 := docs[1]
	assert.Contains(t, doc2.PageContent, "title: Second Article")
	assert.Contains(t, doc2.PageContent, "author: Bob Johnson")
	assert.Equal(t, 1, doc2.Metadata["array_index"])
	assert.Equal(t, float64(2), doc2.Metadata["id"])

	// Check third document
	doc3 := docs[2]
	assert.Contains(t, doc3.PageContent, "title: Third Article")
	assert.Contains(t, doc3.PageContent, "author: Carol Williams")
	assert.Equal(t, 2, doc3.Metadata["array_index"])
	assert.Equal(t, float64(3), doc3.Metadata["id"])
}

func TestJSONLoader_ArrayWithSchema(t *testing.T) {
	t.Parallel()
	file, err := os.Open("./testdata/test_array.json")
	require.NoError(t, err)
	defer file.Close()

	// Use "content" as main content and extract specific metadata
	loader := NewJSONWithSchema(file, "content", "title", "author", "category")

	ctx := context.Background()
	docs, err := loader.Load(ctx)
	require.NoError(t, err)
	require.Len(t, docs, 3)

	// Check first document
	doc1 := docs[0]
	assert.Equal(t, "This is the content of the first article about technology and innovation.", doc1.PageContent)
	assert.Equal(t, "First Article", doc1.Metadata["title"])
	assert.Equal(t, "Alice Smith", doc1.Metadata["author"])
	assert.Equal(t, "Technology", doc1.Metadata["category"])
	assert.Equal(t, 0, doc1.Metadata["array_index"])

	// Check that non-specified fields are not in metadata (except array_index which is always added)
	assert.NotContains(t, doc1.Metadata, "id")
	assert.NotContains(t, doc1.Metadata, "published")
}

func TestJSONLoader_LoadAndSplit(t *testing.T) {
	t.Parallel()
	file, err := os.Open("./testdata/test.json")
	require.NoError(t, err)
	defer file.Close()

	loader := NewJSON(file)
	splitter := textsplitter.NewRecursiveCharacter()
	splitter.ChunkSize = 100
	splitter.ChunkOverlap = 20

	ctx := context.Background()
	docs, err := loader.LoadAndSplit(ctx, splitter)
	require.NoError(t, err)

	// Should have at least one document, possibly more if split
	assert.Greater(t, len(docs), 0)

	// All chunks should have content
	for _, doc := range docs {
		assert.NotEmpty(t, doc.PageContent)
	}
}

func TestJSONLoader_InvalidJSON(t *testing.T) {
	t.Parallel()
	invalidJSON := strings.NewReader(`{"invalid": json}`)

	loader := NewJSON(invalidJSON)

	ctx := context.Background()
	_, err := loader.Load(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse JSON")
}

func TestJSONLoader_PrimitiveValue(t *testing.T) {
	t.Parallel()
	primitiveJSON := strings.NewReader(`"just a string"`)

	loader := NewJSON(primitiveJSON)

	ctx := context.Background()
	docs, err := loader.Load(ctx)
	require.NoError(t, err)
	require.Len(t, docs, 1)

	doc := docs[0]
	assert.Equal(t, "just a string", doc.PageContent)
	assert.Equal(t, "string", doc.Metadata["type"])
}

func TestJSONLoader_NestedObject(t *testing.T) {
	t.Parallel()
	nestedJSON := strings.NewReader(`{
		"title": "Nested Example",
		"user": {
			"name": "John",
			"profile": {
				"age": 30,
				"city": "New York"
			}
		},
		"settings": {
			"theme": "dark",
			"notifications": true
		}
	}`)

	loader := NewJSON(nestedJSON)

	ctx := context.Background()
	docs, err := loader.Load(ctx)
	require.NoError(t, err)
	require.Len(t, docs, 1)

	doc := docs[0]

	// Check that nested objects are properly formatted
	assert.Contains(t, doc.PageContent, "title: Nested Example")
	assert.Contains(t, doc.PageContent, "user:")
	assert.Contains(t, doc.PageContent, "  name: John")
	assert.Contains(t, doc.PageContent, "  profile:")
	assert.Contains(t, doc.PageContent, "    age: 30")
	assert.Contains(t, doc.PageContent, "    city: New York")
}
