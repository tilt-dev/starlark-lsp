package analysis

import (
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"

	"github.com/tilt-dev/starlark-lsp/pkg/query"
)

func (f *fixture) builtinSymbols() {
	_ = WithStarlarkBuiltins()(f.a)
}

func (f *fixture) osSysSymbols() {
	f.Symbols("os", "sys")
	f.builtins.Symbols[0].Children = []protocol.DocumentSymbol{
		f.Symbol("environ"),
		f.Symbol("name"),
	}
	f.builtins.Symbols[1].Children = []protocol.DocumentSymbol{
		f.Symbol("argv"),
		f.Symbol("executable"),
	}
}

func assertCompletionResult(t *testing.T, names []string, result *protocol.CompletionList) {
	labels := make([]string, len(result.Items))
	for i, item := range result.Items {
		labels[i] = item.Label
	}
	assert.ElementsMatch(t, names, labels)
}

func TestSimpleCompletion(t *testing.T) {
	f := newFixture(t)

	f.Symbols("foo", "bar", "baz")

	f.Document("")
	result := f.a.Completion(f.doc, protocol.Position{})
	assertCompletionResult(t, []string{"foo", "bar", "baz"}, result)

	f.Document("ba")
	result = f.a.Completion(f.doc, protocol.Position{Character: 2})
	assertCompletionResult(t, []string{"bar", "baz"}, result)
}

func TestSimpleAttributeCompletion(t *testing.T) {
	f := newFixture(t)
	f.osSysSymbols()

	f.Document("")
	result := f.a.Completion(f.doc, protocol.Position{})
	assertCompletionResult(t, []string{"os", "sys"}, result)

	f.Document("os.")
	result = f.a.Completion(f.doc, protocol.Position{Character: 3})
	assertCompletionResult(t, []string{"environ", "name"}, result)

	f.Document("os.e")
	result = f.a.Completion(f.doc, protocol.Position{Character: 4})
	assertCompletionResult(t, []string{"environ"}, result)
}

func TestCompletionMiddleOfDocument(t *testing.T) {
	f := newFixture(t)
	f.osSysSymbols()
	f.Document(`
def f1():
    pass

s = "a string"

def f2():
    # <- position 2
	return False

# <- position 1

#^- position 3

if True:
    # position 4
	pass

t = 1234
`)
	result := f.a.Completion(f.doc, protocol.Position{Line: 10}) // position 1
	assertCompletionResult(t, []string{"f1", "s", "f2", "os", "sys"}, result)

	result = f.a.Completion(f.doc, protocol.Position{Line: 7, Character: 4}) // position 2
	assertCompletionResult(t, []string{"f1", "s", "f2", "t", "os", "sys"}, result)

	result = f.a.Completion(f.doc, protocol.Position{Line: 11}) // position 3
	assertCompletionResult(t, []string{"f1", "s", "f2", "os", "sys"}, result)

	result = f.a.Completion(f.doc, protocol.Position{Line: 15, Character: 4}) // position 4
	assertCompletionResult(t, []string{"f1", "s", "f2", "os", "sys"}, result)
}

func TestCompletionWithAnErrorNode(t *testing.T) {
	f := newFixture(t)
	f.osSysSymbols()
	f.Document(`
def foo():
  pass

f(

def quux():
  pass
`)
	result := f.a.Completion(f.doc, protocol.Position{Line: 4, Character: 1})
	assertCompletionResult(t, []string{"foo"}, result)
}

func TestCompletionInsideAString(t *testing.T) {
	f := newFixture(t)
	f.osSysSymbols()
	f.Document(`f = "abc123"`)

	result := f.a.Completion(f.doc, protocol.Position{Line: 0, Character: 5})
	assertCompletionResult(t, []string{}, result)
}

func TestCompletionStarlarkBuiltins(t *testing.T) {
	f := newFixture(t)
	f.builtinSymbols()
	f.Document(`f`)

	result := f.a.Completion(f.doc, protocol.Position{Line: 0, Character: 1})
	assertCompletionResult(t, []string{"float", "fail"}, result)
}

func TestCompletionNoneTrueFalse(t *testing.T) {
	f := newFixture(t)
	f.builtinSymbols()

	f.Document(`N`)
	result := f.a.Completion(f.doc, protocol.Position{Line: 0, Character: 1})
	assertCompletionResult(t, []string{"None"}, result)

	f.Document(`T`)
	result = f.a.Completion(f.doc, protocol.Position{Line: 0, Character: 1})
	assertCompletionResult(t, []string{"True"}, result)

	f.Document(`F`)
	result = f.a.Completion(f.doc, protocol.Position{Line: 0, Character: 1})
	assertCompletionResult(t, []string{"False"}, result)
}

func TestIdentifierCompletion(t *testing.T) {
	f := newFixture(t)

	tests := []struct {
		doc      string
		col      uint32
		expected []string
	}{
		{doc: "", col: 0, expected: []string{""}},
		{doc: "os", col: 2, expected: []string{"os"}},
		{doc: "os.", col: 3, expected: []string{"os", ""}},
		{doc: "os.e", col: 4, expected: []string{"os", "e"}},
		{doc: "os.path.", col: 8, expected: []string{"os", "path", ""}},
		{doc: "os.path.e", col: 9, expected: []string{"os", "path", "e"}},
		{doc: "[os]", col: 3, expected: []string{"os"}},
		{doc: "[os.]", col: 4, expected: []string{"os", ""}},
		{doc: "[os.e]", col: 5, expected: []string{"os", "e"}},
		{doc: "x = [os]", col: 7, expected: []string{"os"}},
		{doc: "x = [os.]", col: 8, expected: []string{"os", ""}},
		{doc: "x = [os.e]", col: 9, expected: []string{"os", "e"}},
		{doc: "x = [os.path.]", col: 13, expected: []string{"os", "path", ""}},
		{doc: "x = [os.path.e]", col: 14, expected: []string{"os", "path", "e"}},
	}

	for _, tt := range tests {
		t.Run(tt.doc, func(t *testing.T) {
			f.Document(tt.doc)
			nodes := nodesAtPointForCompletion(f.doc, sitter.Point{Column: tt.col})
			ids := query.ExtractIdentifiers(f.doc, nodes, nil)
			assert.ElementsMatch(t, tt.expected, ids)
		})
	}
}
