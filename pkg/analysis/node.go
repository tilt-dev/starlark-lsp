package analysis

import (
	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/protocol"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
)

const (
	NodeTypeArgList             = "argument_list"
	NodeTypeFunctionDef         = "function_definition"
	NodeTypeParameters          = "parameters"
	NodeTypeIdentifier          = "identifier"
	NodeTypeExpressionStatement = "expression_statement"
	NodeTypeString              = "string"
	NodeTypeBlock               = "block"

	FieldName       = "name"
	FieldParameters = "parameters"
	FieldReturnType = "return_type"
	FieldBody       = "body"
)

// NamedNodeAtPosition returns the most granular named descendant at a position.
func NamedNodeAtPosition(doc document.Document, pos protocol.Position) (*sitter.Node, bool) {
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
