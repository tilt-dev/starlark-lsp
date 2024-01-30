package query_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"

	"github.com/autokitteh/starlark-lsp/pkg/query"
)

func TestQueryDocumentSymbols(t *testing.T) {
	f := newQueryFixture(t, "", `
x = a(3)
y = None
z = True
`)

	doc := f.document()
	symbols := query.DocumentSymbols(doc)
	names := make([]string, len(symbols))
	for i, sym := range symbols {
		names[i] = sym.Name
	}
	assert.Equal(t, []string{"x", "y", "z"}, names)
}

func TestQuerySiblingSymbols(t *testing.T) {
	f := newQueryFixture(t, "", `
def foo():
  bar = 1
  def baz():
    pass
  # position
  pass

def start():
  pass
`)

	doc := f.document()
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
	f := newQueryFixture(t, "", `
def foo():
  bar = 1
  def baz():
    pass
  # position
  pass

def start():
  pass
`)

	doc := f.document()
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
	f := newQueryFixture(t, "", `
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

	doc := f.document()
	n, ok := query.NamedNodeAtPosition(doc, protocol.Position{Line: 5, Character: 2})
	assert.True(t, ok)
	symbols := query.SymbolsInScope(doc, n)
	names := make([]string, len(symbols))
	for i, sym := range symbols {
		names[i] = sym.Name
	}
	assert.Equal(t, []string{"bar", "baz"}, names)
}

func TestSymbolsInScopeIncludesFunctionArguments1(t *testing.T) {
	f := newQueryFixture(t, "", `
def foo(a, b=True, c=None):
  bar = 1
  def baz(d):
    pass
  # position
  pass
`)

	doc := f.document()
	n, ok := query.NamedNodeAtPosition(doc, protocol.Position{Line: 5, Character: 2})
	assert.True(t, ok)
	symbols := query.SymbolsInScope(doc, n)
	names := make([]string, len(symbols))
	for i, sym := range symbols {
		names[i] = sym.Name
	}
	assert.Equal(t, []string{"bar", "baz", "a", "b", "c"}, names)
}

func TestSymbolsInScopeIncludesFunctionArguments2(t *testing.T) {
	f := newQueryFixture(t, "", `
def foo(a, b=True, c=None):
  bar = 1
  def baz(d):
    # position
    pass
  pass
`)

	doc := f.document()
	n, ok := query.NamedNodeAtPosition(doc, protocol.Position{Line: 4, Character: 4})
	assert.True(t, ok)
	symbols := query.SymbolsInScope(doc, n)
	names := make([]string, len(symbols))
	for i, sym := range symbols {
		names[i] = sym.Name
	}
	assert.Equal(t, []string{"d", "bar", "baz", "a", "b", "c"}, names)
}
