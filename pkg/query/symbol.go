package query

import (
	"strings"

	"go.lsp.dev/protocol"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
)

// Get all symbols defined at the same level as the given node.
func SiblingSymbols(doc document.Document, begin, end *sitter.Node) []protocol.DocumentSymbol {
	var symbols []protocol.DocumentSymbol
	for n := begin; n != nil && n != end; n = n.NextNamedSibling() {
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
			// Look for possible docstring for the assigned variable
			if n.NextNamedSibling() != nil && n.NextNamedSibling().Type() == NodeTypeExpressionStatement {
				if ch := n.NextNamedSibling().NamedChild(0); ch != nil && ch.Type() == NodeTypeString {
					symbol.Detail = strings.Trim(doc.Content(ch), `"'`)
				}
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
		// We only ignore symbols following the start node at the same level as
		// the start node. In parent scopes, all symbols are visible. Hence, we
		// pass `start` to SiblingSymbols, not `n`.
		symbols = append(symbols, SiblingSymbols(doc, n.Parent().NamedChild(0), start)...)
	}
	return symbols
}

// DocumentSymbols returns all symbols with document-wide visibility.
func DocumentSymbols(doc document.Document) []protocol.DocumentSymbol {
	return SiblingSymbols(doc, doc.Tree().RootNode().NamedChild(0), nil)
}
