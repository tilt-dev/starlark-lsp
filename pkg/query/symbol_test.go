package query_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
	"github.com/tilt-dev/starlark-lsp/pkg/query"
)

func TestQueryDocumentSymbols(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	input := []byte(`
x = a(3)
y = None
z = True
`)

	tree, err := query.Parse(ctx, input)
	require.NoError(t, err)

	doc := document.NewDocument(input, tree)
	symbols := query.DocumentSymbols(doc)
	names := make([]string, len(symbols))
	for i, sym := range symbols {
		names[i] = sym.Name
	}
	assert.Equal(t, []string{"x", "y", "z"}, names)
}
