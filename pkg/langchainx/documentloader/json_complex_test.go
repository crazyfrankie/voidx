package documentloaders

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONLoader_ComplexNested(t *testing.T) {
	t.Parallel()
	file, err := os.Open("./testdata/test_complex.json")
	require.NoError(t, err)
	defer file.Close()

	loader := NewJSON(file)

	ctx := context.Background()
	docs, err := loader.Load(ctx)
	require.NoError(t, err)
	require.Len(t, docs, 1) // Current behavior: treats as single document

	doc := docs[0]
	t.Logf("Generated content:\n%s", doc.PageContent)
	t.Logf("Metadata keys: %v", getMapKeys(doc.Metadata))

	// Current implementation will flatten everything into readable text
	assert.Contains(t, doc.PageContent, "sections:")
	assert.Contains(t, doc.PageContent, "section_title: Introduction")
	assert.Contains(t, doc.PageContent, "subsections:")
}

func getMapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
