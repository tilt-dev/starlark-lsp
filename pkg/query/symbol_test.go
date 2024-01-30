package query_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"

	"github.com/autokitteh/starlark-lsp/pkg/query"
)

func names(symbols []query.Symbol) []string {
	names := make([]string, len(symbols))
	for i, sym := range symbols {
		names[i] = sym.Name
	}
	return names
}

func TestQueryDocumentSymbols(t *testing.T) {
	f := newQueryFixture(t, "", `
x = a(3)
y = None
z = True
`)

	doc := f.document()
	symbols := query.DocumentSymbols(doc)
	assert.Equal(t, []string{"x", "y", "z"}, names(symbols))
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
	assert.Equal(t, []string{"bar", "baz"}, names(symbols))
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
	assert.Equal(t, []string{"bar", "baz"}, names(symbols))
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
	assert.Equal(t, []string{"bar", "baz"}, names(symbols))
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
	assert.Equal(t, []string{"bar", "baz", "a", "b", "c"}, names(symbols))
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
	assert.Equal(t, []string{"d", "bar", "baz", "a", "b", "c"}, names(symbols))
}

func TestSymbolsInScopeDotCompletion1(t *testing.T) {
	f := newQueryFixture(t, "", `
def foo():
  bar = 1
  bar.
  # position is dot
`)
	// module [0, 0] - [3, 0]
	//  ERROR [0, 0] - [2, 6]
	//    identifier [0, 4] - [0, 7]
	//    parameters [0, 7] - [0, 9]
	//    expression_statement [1, 2] - [1, 9]
	//      assignment [1, 2] - [1, 9]
	//        left: identifier [1, 2] - [1, 5]
	//        right: integer [1, 8] - [1, 9]
	//    identifier [2, 2] - [2, 5]

	doc := f.document()
	n, ok := query.NamedNodeAtPosition(doc, protocol.Position{Line: 3, Character: 5}) // dot
	assert.Equal(t, n.Type(), query.NodeTypeERROR)
	assert.True(t, ok)
	assert.True(t, query.IsModuleScope(doc, n)) // ERROR is below module (failed to parse function definition) therefore the scope is module

	n, ok = query.NamedNodeAtPosition(doc, protocol.Position{Line: 3, Character: 4}) // bar
	assert.True(t, ok)
	symbols := query.SymbolsInScope(doc, n)
	assert.False(t, query.IsModuleScope(doc, n)) // bar is below highest ERROR + identifier and parameters, assume this is ERROR function definition
	assert.Equal(t, []string{"bar"}, names(symbols))
}

func TestSymbolsInScopeDotCompletion2(t *testing.T) {
	f := newQueryFixture(t, "", `
def foo():
	bar = 1
	bar.
	
def foo2():
	pass
`)
	// after dot there is more chars, therefore tree sitter will try to parse it as a ERROR call expression

	//module [0, 0] - [6, 0]
	//  function_definition [0, 0] - [5, 6]
	//    name: identifier [0, 4] - [0, 7]
	//    parameters: parameters [0, 7] - [0, 9]
	//    body: block [1, 2] - [5, 6]
	//      expression_statement [1, 2] - [1, 9]
	//        assignment [1, 2] - [1, 9]
	//          left: identifier [1, 2] - [1, 5]
	//          right: integer [1, 8] - [1, 9]
	//      ERROR [2, 2] - [4, 9]
	//        call [2, 2] - [4, 8]
	//          function: attribute [2, 2] - [4, 3]         <- dot
	//            object: identifier [2, 2] - [2, 5]        <- bar
	//            attribute: identifier [4, 0] - [4, 3]
	//          ERROR [4, 4] - [4, 6]
	//            identifier [4, 4] - [4, 6]
	//          arguments: argument_list [4, 6] - [4, 8]
	//      pass_statement [5, 2] - [5, 6]
	//

	doc := f.document()
	n, ok := query.NamedNodeAtPosition(doc, protocol.Position{Line: 3, Character: 5}) // dot
	assert.Equal(t, n.Type(), query.NodeTypeAttribute)
	assert.True(t, ok)
	assert.False(t, query.IsModuleScope(doc, n)) // this time there is function defintion on top level

	n, ok = query.NamedNodeAtPosition(doc, protocol.Position{Line: 3, Character: 4}) // bar
	assert.True(t, ok)
	symbols := query.SymbolsInScope(doc, n)
	assert.False(t, query.IsModuleScope(doc, n))
	assert.Equal(t, []string{"bar"}, names(symbols))
}
