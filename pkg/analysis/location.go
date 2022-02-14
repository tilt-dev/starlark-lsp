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

func PointCmp(a, b sitter.Point) int {
	if a.Row < b.Row {
		return -1
	}

	if a.Row > b.Row {
		return 1
	}

	if a.Column < b.Column {
		return -1
	}

	if a.Column > b.Column {
		return 1
	}

	return 0
}

func PointBeforeOrEqual(a, b sitter.Point) bool {
	return PointCmp(a, b) <= 0
}

func PointBefore(a, b sitter.Point) bool {
	return PointCmp(a, b) < 0
}

func PointAfterOrEqual(a, b sitter.Point) bool {
	return PointCmp(a, b) >= 0
}

func PointAfter(a, b sitter.Point) bool {
	return PointCmp(a, b) > 0
}
