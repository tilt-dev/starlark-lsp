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
