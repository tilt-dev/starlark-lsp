package query_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
	"github.com/tilt-dev/starlark-lsp/pkg/query"
)

func TestQueryDocumentSymbols(t *testing.T) {
	f := newQueryFixture(t, []byte{}, `
x = a(3)
y = None
z = True
`)

	doc := document.NewDocument(f.input, f.tree)
	symbols := query.DocumentSymbols(doc)
	names := make([]string, len(symbols))
	for i, sym := range symbols {
		names[i] = sym.Name
	}
	assert.Equal(t, []string{"x", "y", "z"}, names)
}

func TestQuerySiblingSymbols(t *testing.T) {
	f := newQueryFixture(t, []byte{}, `
def foo():
  bar = 1
  def baz():
    pass
  # position
  pass

def start():
  pass
`)

	doc := document.NewDocument(f.input, f.tree)
	n, ok := query.NamedNodeAtPosition(doc, protocol.Position{Line: 5, Character: 2})
	assert.True(t, ok)
	symbols := query.SiblingSymbols(doc, n.Parent().NamedChild(0), nil)
	names := make([]string, len(symbols))
	for i, sym := range symbols {
		names[i] = sym.Name
	}
	assert.Equal(t, []string{"bar", "baz"}, names)
}

func TestSymbolsInScope(t *testing.T) {
	f := newQueryFixture(t, []byte{}, `
def foo():
  bar = 1
  def baz():
    pass
  # position
  pass

def start():
  pass
`)

	doc := document.NewDocument(f.input, f.tree)
	n, ok := query.NamedNodeAtPosition(doc, protocol.Position{Line: 5, Character: 2})
	assert.True(t, ok)
	symbols := query.SymbolsInScope(doc, n)
	names := make([]string, len(symbols))
	for i, sym := range symbols {
		names[i] = sym.Name
	}
	assert.Equal(t, []string{"bar", "baz"}, names)
}

func TestSymbolsInScopeExcludesFollowingSiblings(t *testing.T) {
	f := newQueryFixture(t, []byte{}, `
def foo():
  bar = 1
  def baz():
    pass
  # position
  quux = True
  return

def start():
  pass
`)

	doc := document.NewDocument(f.input, f.tree)
	n, ok := query.NamedNodeAtPosition(doc, protocol.Position{Line: 5, Character: 2})
	assert.True(t, ok)
	symbols := query.SymbolsInScope(doc, n)
	names := make([]string, len(symbols))
	for i, sym := range symbols {
		names[i] = sym.Name
	}
	assert.Equal(t, []string{"bar", "baz"}, names)
}
