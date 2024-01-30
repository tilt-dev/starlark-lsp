package query_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/autokitteh/starlark-lsp/pkg/document"
	"github.com/autokitteh/starlark-lsp/pkg/query"
)

func TestFunctionParameters(t *testing.T) {
	type param struct {
		name  string
		value string
		typ   string
	}

	type tc struct {
		source string
		params []param
	}
	tcs := []tc{
		{source: `foo()`, params: nil},
		{source: `foo(a)`, params: []param{{name: "a"}}},
		{source: `foo(a, b, c)`, params: []param{{name: "a"}, {name: "b"}, {name: "c"}}},
		{
			source: `foo(a, b: str, c=None, d: int=-1)`,
			params: []param{
				{name: "a"},
				{name: "b", typ: "str"},
				{name: "c", value: "None"},
				{name: "d", typ: "int", value: "-1"},
			},
		},
	}

	for _, tt := range tcs {
		t.Run(tt.source, func(t *testing.T) {
			f := newQueryFixture(t, query.FunctionParameters, fmt.Sprintf("def %s:\n\tpass\n", tt.source))
			qc := f.exec()

			if len(tt.params) == 0 {
				m, ok := qc.NextMatch()
				require.False(t, ok, "Unexpected match found")
				require.Nil(t, m, "Match object is non-nil")
				// nothing more to assert
				return
			}

			for _, param := range tt.params {
				m, ok := qc.NextMatch()
				require.True(t, ok, "No match found for param %q", param.name)
				require.NotNil(t, m, "Match object is nil for param %q", param.name)
				f.assertCapture("name", param.name, m.Captures)
				if param.value != "" {
					f.assertCapture("value", param.value, m.Captures)
				} else {
					f.assertNoCapture("value", m.Captures)
				}
				if param.typ != "" {
					f.assertCapture("type", param.typ, m.Captures)
				} else {
					f.assertNoCapture("type", m.Captures)
				}
			}
		})
	}
}

func TestLeafNodes(t *testing.T) {
	tests := []struct {
		src   string
		nodes []string
		types []string
	}{
		{
			src:   "a or b",
			nodes: []string{"a", "or", "b"},
			types: []string{"identifier", "or", "identifier"},
		},
		{
			src:   "a = b.",
			nodes: []string{"a", "=", "b", "."},
			types: []string{"identifier", "=", "identifier", "."},
		},
		{
			src:   "a = b.c.",
			nodes: []string{"a", "=", "b", ".", "c", "."},
			types: []string{"identifier", "=", "identifier", ".", "identifier", "."},
		},
		{
			src:   "if a or b: pass",
			nodes: []string{"if", "a", "or", "b", ":", "pass"},
			types: []string{"if", "identifier", "or", "identifier", ":", "pass"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			q := newQueryFixture(t, "", tt.src)
			nodes := query.LeafNodes(q.root)
			names := make([]string, len(nodes))
			types := make([]string, len(nodes))
			for i, n := range nodes {
				names[i] = q.nodeContents(n)
				types[i] = n.Type()
			}
			assert.ElementsMatch(t, tt.nodes, names)
			assert.ElementsMatch(t, tt.types, types)
		})
	}
}

func TestLoadStatements(t *testing.T) {
	q := newQueryFixture(t, "", `
load("file1.star", "foo", "bar")
if foo():
  load("file2.star", q="quux") # This is not a legal load statement but we still want to capture it
quux()
`)
	nodes := query.LoadStatements(q.input, q.tree)
	assert.Equal(t, 2, len(nodes))

	fnNode := nodes[0].ChildByFieldName("function")
	assert.Equal(t, "load", q.nodeContents(fnNode))
	argsNode := nodes[0].ChildByFieldName("arguments")
	args := make([]string, argsNode.NamedChildCount())
	for i := 0; i < len(args); i++ {
		args[i] = q.nodeContents(argsNode.NamedChild(i))
	}
	assert.ElementsMatch(t, []string{`"file1.star"`, `"foo"`, `"bar"`}, args)

	fnNode = nodes[1].ChildByFieldName("function")
	assert.Equal(t, "load", q.nodeContents(fnNode))
	argsNode = nodes[1].ChildByFieldName("arguments")
	args = make([]string, argsNode.NamedChildCount())
	for i := 0; i < len(args); i++ {
		args[i] = q.nodeContents(argsNode.NamedChild(i))
	}
	assert.ElementsMatch(t, []string{`"file2.star"`, `q="quux"`}, args)
}

type queryFixture struct {
	t     testing.TB
	q     *sitter.Query
	input []byte
	tree  *sitter.Tree
	root  *sitter.Node
}

func newQueryFixture(t testing.TB, queryPattern string, src string) *queryFixture {
	t.Helper()

	lang := python.GetLanguage()

	var q *sitter.Query
	var err error
	if len(queryPattern) > 0 {
		q, err = sitter.NewQuery([]byte(queryPattern), lang)
		t.Cleanup(q.Close)
		require.NoError(t, err, "Error creating query %q", string(queryPattern))
	}

	parseCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	srcBytes := []byte(src)
	tree, err := query.Parse(parseCtx, srcBytes)
	require.NoError(t, err, "Error parsing source:\n-----%s\n-----", src)

	f := queryFixture{
		t:     t,
		q:     q,
		input: srcBytes,
		tree:  tree,
		root:  tree.RootNode(),
	}
	return &f
}

func (f *queryFixture) exec() *sitter.QueryCursor {
	f.t.Helper()
	qc := sitter.NewQueryCursor()
	f.t.Cleanup(qc.Close)
	qc.Exec(f.q, f.root)
	return qc
}

func (f *queryFixture) assertCapture(name string, value string, captures []sitter.QueryCapture) bool {
	f.t.Helper()

	captureValues := make(map[string]string, len(captures))
	for _, c := range captures {
		captureName := f.q.CaptureNameForId(c.Index)
		captureValue := f.nodeContents(c.Node)
		captureValues[captureName] = captureValue
	}

	actualValue, ok := captureValues[name]
	if !ok {
		return assert.Failf(f.t, "Capture missing",
			"Capture name: %s\nCaptures: %v", name, captureValues)
	}
	return assert.Equalf(f.t, value, actualValue, "Wrong capture value for %q: %v", name, captureValues)
}

func (f *queryFixture) assertNoCapture(name string, captures []sitter.QueryCapture) bool {
	f.t.Helper()

	captureValues := make(map[string]string, len(captures))
	for _, c := range captures {
		captureName := f.q.CaptureNameForId(c.Index)
		captureValue := c.Node.Content(f.input)
		captureValues[captureName] = captureValue
	}

	_, ok := captureValues[name]
	if ok {
		return assert.Failf(f.t, "Unexpected capture value",
			"Capture name: %s\nCaptures: %v", name, captureValues)
	}
	return true
}

func (f *queryFixture) nodeContents(n *sitter.Node) string {
	return n.Content(f.input)
}

func (f *queryFixture) document() query.DocumentContent {
	doc := document.NewDocument("", f.input, f.tree)
	f.t.Cleanup(doc.Close)
	return doc
}
