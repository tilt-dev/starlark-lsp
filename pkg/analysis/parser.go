package analysis

import (
	"context"
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
	"go.lsp.dev/protocol"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
	"github.com/tilt-dev/starlark-lsp/pkg/query"
)

const (
	NodeTypeArgList     = "argument_list"
	NodeTypeFunctionDef = "function_definition"
	NodeTypeIdentifier  = "identifier"

	FieldName       = "name"
	FieldParameters = "parameters"
)

var lang = python.GetLanguage()

func Parse(ctx context.Context, input []byte) (*sitter.Tree, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(lang)

	tree, err := parser.ParseCtx(ctx, nil, input)
	if err != nil {
		return nil, err
	}

	return tree, nil
}

// NodeAtPosition returns the most granular named descendant at a position.
func NodeAtPosition(doc document.Document, pos protocol.Position) (*sitter.Node, bool) {
	pt := PositionToPoint(pos)
	if doc.Tree == nil {
		return nil, false
	}
	node := doc.Tree.RootNode().NamedDescendantForPointRange(pt, pt)
	if node != nil {
		return node, true
	}
	return nil, false
}

// Query executes a Tree-sitter S-expression query against a subtree and invokes
// matchFn on each result.
func Query(node *sitter.Node, pattern []byte, matchFn func(q *sitter.Query, match *sitter.QueryMatch)) {
	q := mustQuery(pattern)
	qc := sitter.NewQueryCursor()
	defer qc.Close()

	qc.Exec(q, node)
	for m, hasMatch := qc.NextMatch(); hasMatch; m, hasMatch = qc.NextMatch() {
		if m == nil {
			panic("tree-sitter returned nil match")
		}
		matchFn(q, m)
	}
}

// DocumentSymbols returns all symbols with document-wide visibility.
// TODO(milas): this currently only looks for assignment expressions
func DocumentSymbols(doc document.Document) []protocol.SymbolInformation {
	var symbols []protocol.SymbolInformation
	for n := doc.Tree.RootNode().NamedChild(0); n != nil; n = n.NextNamedSibling() {
		var symbol protocol.SymbolInformation

		if n.Type() == "expression_statement" {
			assignment := n.NamedChild(0)
			if assignment == nil || assignment.Type() != "assignment" {
				continue
			}
			symbol.Name = assignment.ChildByFieldName("left").Content(doc.Contents)
			kind := nodeTypeToSymbolKind(assignment.ChildByFieldName("right"))
			if kind == 0 {
				kind = protocol.SymbolKindVariable
			}
			symbol.Kind = kind
			symbol.Location.Range = protocol.Range{
				Start: PointToPosition(n.StartPoint()),
				End:   PointToPosition(n.EndPoint()),
			}
		}

		if symbol.Name != "" {
			symbols = append(symbols, symbol)
		}
	}

	return symbols
}

func Functions(doc document.Document, node *sitter.Node) map[string]protocol.SignatureInformation {
	signatures := make(map[string]protocol.SignatureInformation)
	for _, n := range nodePathTopDown(node) {
		scopeSigs := functionsForDirectScope(doc, n)
		for fn, sig := range scopeSigs {
			signatures[fn] = sig
		}
	}
	return signatures
}

func nodePathBottomUp(node *sitter.Node) []*sitter.Node {
	var nodes []*sitter.Node
	for n := node; n != nil; n = n.Parent() {
		nodes = append(nodes, n)
	}
	return nodes
}

func nodePathTopDown(node *sitter.Node) []*sitter.Node {
	nodes := nodePathBottomUp(node)
	// we built the collection bottom-up, so need to reverse it
	for i, j := 0, len(nodes)-1; i < j; i, j = i+1, j-1 {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	}
	return nodes
}

func functionsForDirectScope(doc document.Document, node *sitter.Node) map[string]protocol.SignatureInformation {
	signatures := make(map[string]protocol.SignatureInformation)
	for n := node.NamedChild(0); n != nil; n = n.NextNamedSibling() {
		if n.Type() != NodeTypeFunctionDef {
			continue
		}

		fnName := n.ChildByFieldName(FieldName).Content(doc.Contents)
		var signature protocol.SignatureInformation
		var rawParams []string
		Query(n, query.FunctionParameters, func(q *sitter.Query, match *sitter.QueryMatch) {
			var param protocol.ParameterInformation

			for _, c := range match.Captures {
				// TODO(milas): use type + default (if available) in label
				switch q.CaptureNameForId(c.Index) {
				case "param":
					label := c.Node.Content(doc.Contents)
					param.Label = label
					rawParams = append(rawParams, label)
				}
			}

			signature.Parameters = append(signature.Parameters, param)
		})

		signature.Label = fmt.Sprintf("%s(%s)", fnName, strings.Join(rawParams, ", "))
		signatures[fnName] = signature
	}

	return signatures
}

func mustQuery(pattern []byte) *sitter.Query {
	q, err := sitter.NewQuery(pattern, lang)
	if err != nil {
		panic(fmt.Errorf("invalid query pattern\n-----%s\n-----\n", strings.TrimSpace(string(pattern))))
	}
	return q
}

func nodeTypeToSymbolKind(n *sitter.Node) protocol.SymbolKind {
	switch n.Type() {
	case "true":
		return protocol.SymbolKindBoolean
	case "false":
		return protocol.SymbolKindBoolean
	case "list":
		return protocol.SymbolKindArray
	case "dictionary":
		return protocol.SymbolKindObject
	case "integer":
		return protocol.SymbolKindNumber
	case "float":
		return protocol.SymbolKindNumber
	case "none":
		return protocol.SymbolKindNull
	case "string":
		return protocol.SymbolKindString
	}
	return 0
}
