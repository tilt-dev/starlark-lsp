package query

import (
	"go.lsp.dev/protocol"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
)

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