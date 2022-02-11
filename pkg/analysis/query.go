package analysis

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

func mustQuery(pattern []byte) *sitter.Query {
	q, err := sitter.NewQuery(pattern, lang)
	if err != nil {
		panic(fmt.Errorf("invalid query pattern\n-----%s\n-----\n", strings.TrimSpace(string(pattern))))
	}
	return q
}

type QueryMatchFunc func(q *sitter.Query, match *sitter.QueryMatch) bool

// Query executes a Tree-sitter S-expression query against a subtree and invokes
// matchFn on each result.
func Query(node *sitter.Node, pattern []byte, matchFn QueryMatchFunc) {
	q := mustQuery(pattern)
	qc := sitter.NewQueryCursor()
	defer qc.Close()

	qc.Exec(q, node)
	for m, hasMatch := qc.NextMatch(); hasMatch; m, hasMatch = qc.NextMatch() {
		if m == nil {
			panic("tree-sitter returned nil match")
		}
		if !matchFn(q, m) {
			return
		}
	}
}
