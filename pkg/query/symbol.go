package query

import (
	"go.lsp.dev/protocol"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
)

// Get all symbols defined at the same level as the given node.
func SiblingSymbols(doc document.Document, n *sitter.Node) []protocol.DocumentSymbol {
	var symbols []protocol.DocumentSymbol
	for ; n != nil; n = n.NextNamedSibling() {
		var symbol protocol.DocumentSymbol

		if n.Type() == NodeTypeExpressionStatement {
			assignment := n.NamedChild(0)
			if assignment == nil || assignment.Type() != "assignment" {
				continue
			}
			symbol.Name = doc.Content(assignment.ChildByFieldName("left"))
			kind := nodeTypeToSymbolKind(assignment.ChildByFieldName("right"))
			if kind == 0 {
				kind = protocol.SymbolKindVariable
			}
			symbol.Kind = kind
			symbol.Range = protocol.Range{
				Start: PointToPosition(n.StartPoint()),
				End:   PointToPosition(n.EndPoint()),
			}
		}

		if n.Type() == NodeTypeFunctionDef {
			name, sigInfo := extractSignatureInformation(doc, n)
			symbol.Name = name
			symbol.Kind = protocol.SymbolKindFunction
			symbol.Detail = sigInfo.Label
			symbol.Range = protocol.Range{
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

// Get all symbols defined in scopes above the level of the given node.
func SymbolsInScope(doc document.Document, start *sitter.Node) []protocol.DocumentSymbol {
	var symbols []protocol.DocumentSymbol
	for n := start; n.Parent() != nil; n = n.Parent() {
		symbols = append(symbols, SiblingSymbols(doc, n.Parent().NamedChild(0))...)
	}
	return symbols
}

// DocumentSymbols returns all symbols with document-wide visibility.
func DocumentSymbols(doc document.Document) []protocol.DocumentSymbol {
	return SiblingSymbols(doc, doc.Tree().RootNode().NamedChild(0))
}
