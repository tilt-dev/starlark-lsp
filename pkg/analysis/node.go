package analysis

import (
	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/protocol"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
)

const (
	NodeTypeArgList     = "argument_list"
	NodeTypeFunctionDef = "function_definition"
	NodeTypeIdentifier  = "identifier"

	FieldName       = "name"
	FieldParameters = "parameters"
)

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

// nodePathBottomUp returns the path starting from the passed node and ending
// with the root node.
func nodePathBottomUp(node *sitter.Node) []*sitter.Node {
	var nodes []*sitter.Node
	for n := node; n != nil; n = n.Parent() {
		nodes = append(nodes, n)
	}
	return nodes
}

// nodePathTopDown returns the path starting from the root node and ending with
// with the passed node.
func nodePathTopDown(node *sitter.Node) []*sitter.Node {
	nodes := nodePathBottomUp(node)
	// we built the collection bottom-up, so need to reverse it
	for i, j := 0, len(nodes)-1; i < j; i, j = i+1, j-1 {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	}
	return nodes
}
