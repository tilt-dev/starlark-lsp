package analysis

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
	"github.com/tilt-dev/starlark-lsp/pkg/query"
)

func TestBasicHover(t *testing.T) {
	for _, tc := range []struct {
		name                      string
		line                      uint32
		char                      uint32
		expectedHoverRangeContent string
		expectedHoverContent      string
	}{
		{"func", 0, 1, "foo", "desc1"},
		{"var", 0, 6, "hello", "desc2"},
		{"module func - hover over module", 1, 1, "baz.quu", "desc3"},
		{"module func - hover over dot", 1, 3, "baz.quu", "desc3"},
		{"module func - hover over func", 1, 5, "baz.quu", "desc3"},
		{"module var - hover over module", 1, 9, "qux.fd", "desc4"},
		{"module var - hover over dot", 1, 11, "qux.fd", "desc4"},
		{"module var - hover over var", 1, 12, "qux.fd", "desc4"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)

			f.AddFunction("foo", "desc1")
			f.AddSymbol("hello", "desc2")
			f.AddFunction("baz.quu", "desc3")
			f.AddSymbol("qux.fd", "desc4")

			doc := f.MainDoc(`foo(hello)
baz.quu(qux.fd)`)

			result := f.a.Hover(f.ctx, doc, protocol.Position{Character: tc.char, Line: tc.line})
			assertHoverResult(t, doc, tc.expectedHoverRangeContent, tc.expectedHoverContent, result)
		})
	}
}

func TestHoverFuncDefinedInFile(t *testing.T) {
	f := newFixture(t)

	doc := f.MainDoc(`
def foo():
  """
  foos a bar
  """
  pass

foo(hello)
`)
	result := f.a.Hover(f.ctx, doc, protocol.Position{Line: 7, Character: 2})
	assertHoverResult(t, doc, "foo", "foos a bar", result)
}

func TestHoverFuncDefinedWithArgInFile(t *testing.T) {
	f := newFixture(t)

	doc := f.MainDoc(`
def foo(name: str) -> str:
  """
  foos a bar

  Args:
    name: name of the bar
  Returns:
    name of the foo
  """
  pass

foo(hello)
`)
	result := f.a.Hover(f.ctx, doc, protocol.Position{Line: 12, Character: 2})
	assertHoverResult(t, doc, "foo", "foos a bar\n# Parameters\nname: name of the bar\n# Returns\nname of the foo", result)
}

func TestHoverNoMatch(t *testing.T) {
	f := newFixture(t)

	doc := f.MainDoc("foo(hello)")
	result := f.a.Hover(f.ctx, doc, protocol.Position{Character: 6})
	require.Nil(t, result)
}

func assertHoverResult(t *testing.T, doc document.Document, highlighted string, content string, result *protocol.Hover) {
	require.NotNil(t, result, "result")
	require.NotNil(t, result.Range, "result.range")

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
