package analysis

import (
	"context"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"
	"go.lsp.dev/protocol"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
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
