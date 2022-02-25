package query_test

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/assert"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
	"github.com/tilt-dev/starlark-lsp/pkg/query"
)

func TestIdentifiers(t *testing.T) {
	tests := []struct {
		doc      string
		expected []string
		limit    *sitter.Point
	}{
		{doc: "", expected: []string{""}},
		{doc: "os", expected: []string{"os"}},
		{doc: "os.", expected: []string{"os", ""}},
		{doc: "os.e", expected: []string{"os", "e"}},
		{doc: "os.path.", expected: []string{"os", "path", ""}},
		{doc: "os.path.e", expected: []string{"os", "path", "e"}},
		{doc: "[os]", expected: []string{"os"}},
		{doc: "[os.]", expected: []string{"os", ""}},
		{doc: "[os.e]", expected: []string{"os", "e"}},
		{doc: "x = [os]", expected: []string{"x", "os"}},
		{doc: "x = [os.]", expected: []string{"x", "os", ""}},
		{doc: "x = [os.e]", expected: []string{"x", "os", "e"}},
		{doc: "x = [os.path.]", expected: []string{"x", "os", "path", ""}},
		{doc: "x = [os.path.e]", expected: []string{"x", "os", "path", "e"}},
		{doc: `os.path.dirname("blah").strip()`, expected: []string{"os", "path", "dirname", "strip"}},
		{doc: `os.path.
print("")`, expected: []string{"os", "path", "print"}},
		{doc: `os.path.
print("")`, expected: []string{"os", "path", ""}, limit: &sitter.Point{Column: 8}},
	}

	for _, tt := range tests {
		t.Run(tt.doc, func(t *testing.T) {
			f := newQueryFixture(t, []byte{}, tt.doc)
			doc := document.NewDocument(f.input, f.tree)
			ids := query.ExtractIdentifiers(doc, []*sitter.Node{f.tree.RootNode()}, tt.limit)
			assert.ElementsMatch(t, tt.expected, ids)
		})
	}
}
