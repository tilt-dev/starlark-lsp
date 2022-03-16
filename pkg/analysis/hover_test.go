package analysis

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
	"github.com/tilt-dev/starlark-lsp/pkg/query"
)

func TestSimpleHover(t *testing.T) {
	f := newFixture(t)

	f.AddFunction("foo", "foos a bar")

	f.Document("foo(hello)")
	result := f.a.Hover(f.ctx, f.doc, protocol.Position{Character: 6})
	assertHoverResult(t, f.doc, "foo(hello)", "foos a bar", result)
}

func TestHoverFuncDefinedInFile(t *testing.T) {
	f := newFixture(t)

	f.Document(`
def foo():
  """
  foos a bar
  """
  pass

foo(hello)
`)
	result := f.a.Hover(f.ctx, f.doc, protocol.Position{Line: 7, Character: 2})
	assertHoverResult(t, f.doc, "foo(hello)", "foos a bar", result)
}

func TestHoverNoMatch(t *testing.T) {
	f := newFixture(t)

	f.Document("foo(hello)")
	result := f.a.Hover(f.ctx, f.doc, protocol.Position{Character: 6})
	require.Nil(t, result)
}

func assertHoverResult(t *testing.T, doc document.Document, highlighted string, content string, result *protocol.Hover) {
	require.NotNil(t, result)
	require.NotNil(t, result.Range)

	require.Equal(t, highlighted, contentByRange(doc, *result.Range), "highlighted document content")
	require.Equal(t, result.Contents.Value, content, "tooltip content")
}

func nodeWithRange(node *sitter.Node, r protocol.Range) *sitter.Node {
	if query.NodeRange(node) == r {
		return node
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		r := nodeWithRange(node.Child(i), r)
		if r != nil {
			return r
		}
	}
	return nil
}

// we don't have a good way to get content by protocol.Range:
//   a document can look up by byte index, but protocol.Range doesn't have that info
//   inside the analyzer, we still have the original node, so we don't need to do this
//   we'll just do this inefficiently since for now we only need to do this in tests
func contentByRange(doc document.Document, r protocol.Range) string {
	n := nodeWithRange(doc.Tree().RootNode(), r)
	if n == nil {
		return ""
	}
	return doc.Content(n)
}
