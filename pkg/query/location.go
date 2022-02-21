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

// NodeAtPosition returns the node (named or unnamed) with the smallest
// start/end range that covers the given position.
func NodeAtPosition(doc document.Document, pos protocol.Position) (*sitter.Node, bool) {
	namedNode, ok := NamedNodeAtPosition(doc, pos)
	if !ok {
		return nil, false
	}
	count := int(namedNode.ChildCount())
	for i := 0; i < count; i++ {
		child := namedNode.Child(i)
		startPoint := child.StartPoint()
		endPoint := child.EndPoint()
		if startPoint.Row < pos.Line || endPoint.Row > pos.Line {
			continue
		}
		if startPoint.Column <= pos.Character && endPoint.Column >= pos.Character {
			return child, true
		}
	}
	return namedNode, true
}
