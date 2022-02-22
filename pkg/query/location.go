package query

import (
	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/protocol"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
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

// Returns true if point a occurs before point b
func PointBefore(a, b sitter.Point) bool {
	return a.Row < b.Row ||
		a.Row == b.Row && a.Column <= b.Column
}

// NamedNodeAtPosition returns the most granular named descendant at a position.
func NamedNodeAtPosition(doc document.Document, pos protocol.Position) (*sitter.Node, bool) {
	pt := PositionToPoint(pos)
	if doc.Tree() == nil {
		return nil, false
	}
	node := doc.Tree().RootNode().NamedDescendantForPointRange(pt, pt)
	if node != nil {
		return node, true
	}
	return nil, false
}

func ChildNodeAtPosition(doc document.Document, pt sitter.Point, node *sitter.Node) (*sitter.Node, bool) {
	count := int(node.NamedChildCount())
	for i := 0; i < count; i++ {
		child := node.NamedChild(i)
		if PointBefore(child.StartPoint(), pt) && PointBefore(pt, child.EndPoint()) {
			return ChildNodeAtPosition(doc, pt, child)
		}
	}
	return node, true
}

// NodeAtPosition returns the node (named or unnamed) with the smallest
// start/end range that covers the given position.
func NodeAtPosition(doc document.Document, pos protocol.Position) (*sitter.Node, bool) {
	pt := PositionToPoint(pos)
	namedNode, ok := NamedNodeAtPosition(doc, pos)
	if !ok {
		return nil, false
	}
	return ChildNodeAtPosition(doc, pt, namedNode)
}
