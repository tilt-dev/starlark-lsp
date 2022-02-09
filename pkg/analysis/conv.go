package analysis

import (
	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/protocol"
)

// PositionToPoint converts an LSP protocol file location to a Tree-sitter file location.
func PositionToPoint(pos protocol.Position) sitter.Point {
	return sitter.Point{
		Row:    pos.Line,
		Column: pos.Character,
	}
}

// PointToPosition converts a Tree-sitter file location to an LSP protocol file location.
func PointToPosition(point sitter.Point) protocol.Position {
	return protocol.Position{
		Line:      point.Row,
		Character: point.Column,
	}
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
	case "function_definition":
		return protocol.SymbolKindFunction
	}
	return 0
}
