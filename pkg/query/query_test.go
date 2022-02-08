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

	"github.com/tilt-dev/starlark-lsp/pkg/query"
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

type queryFixture struct {
	t     testing.TB
	q     *sitter.Query
	input []byte
	root  *sitter.Node
}

func newQueryFixture(t testing.TB, queryPattern []byte, src string) *queryFixture {
	t.Helper()

	lang := python.GetLanguage()
	q, err := sitter.NewQuery(queryPattern, lang)
	t.Cleanup(q.Close)
	require.NoError(t, err, "Error creating query %q", string(queryPattern))

	parseCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	srcBytes := []byte(src)
	root, err := sitter.ParseCtx(parseCtx, srcBytes, lang)
	require.NoError(t, err, "Error parsing source:\n-----%s\n-----", src)

	f := queryFixture{
		t:     t,
		q:     q,
		input: srcBytes,
		root:  root,
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
		captureValue := c.Node.Content(f.input)
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
