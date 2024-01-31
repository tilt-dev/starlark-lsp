package query_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tilt-dev/starlark-lsp/pkg/query"
)

func TestStructs(t *testing.T) {
	code := `
x = struct(a=1, b=2)
`
	f := newQueryFixture(t, "", code)
	doc := f.document()
	sym := query.Struct(doc, f.root)
	assert.Equal(t, "x", sym.Name)
	assert.Equal(t, 2, len(sym.Children))
	assert.Equal(t, "a", sym.Children[0].Name)
	assert.Equal(t, "b", sym.Children[1].Name)
}
