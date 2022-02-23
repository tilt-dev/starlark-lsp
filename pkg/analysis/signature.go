package analysis

import (
	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/protocol"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
	"github.com/tilt-dev/starlark-lsp/pkg/query"
)

func (a *Analyzer) SignatureHelp(doc document.Document, pos protocol.Position) *protocol.SignatureHelp {
	// TODO(milas): this doesn't work right for ERROR states because we're only
	// 	looking for named nodes
	node, ok := query.NamedNodeAtPosition(doc, pos)
	if !ok {
		return nil
	}

	fnName, activeParam := possibleCallInfo(doc, node, pos)
	if fnName == "" {
		// avoid computing function defs
		return nil
	}

	var sig protocol.SignatureInformation
	for n := node; n != nil; n = n.Parent() {
		var found bool
		sig, found = query.Function(doc, n, fnName)
		if found {
			break
		}
	}

	if sig.Label == "" {
		sig = a.builtins.Functions[fnName]
	}

	if sig.Label == "" {
		return nil
	}

	if activeParam > uint32(len(sig.Parameters)-1) {
		activeParam = uint32(len(sig.Parameters) - 1)
	}

	return &protocol.SignatureHelp{
		Signatures:      []protocol.SignatureInformation{sig},
		ActiveParameter: activeParam,
		ActiveSignature: 0,
	}
}

// possibleCallInfo attempts to find the name of the function for a
// `call`.
//
// Currently, this supports two cases:
// 	(1) Current node is inside of a `call`
// 	(2) Current node is inside of an ERROR block where first child is an
// 		`identifier`
func possibleCallInfo(doc document.Document, node *sitter.Node,
	pos protocol.Position) (fnName string, argIndex uint32) {
	for n := node; n != nil; n = n.Parent() {
		if n.Type() == "call" {
			fnName = doc.Content(n.ChildByFieldName("function"))
			argIndex = possibleActiveParam(doc, n.ChildByFieldName("arguments").Child(0), pos)
			return fnName, argIndex
		} else if n.HasError() {
			// look for `foo(` and assume it's a function call - this could
			// happen if the closing `)` is not (yet) present or if there's
			// something invalid going on within the args, e.g. `foo(x#)`
			possibleCall := n.NamedChild(0)
			if possibleCall != nil && possibleCall.Type() == query.NodeTypeIdentifier {
				possibleParen := possibleCall.NextSibling()
				if possibleParen != nil && !possibleParen.IsNamed() && doc.Content(possibleParen) == "(" {
					fnName = doc.Content(possibleCall)
					argIndex = possibleActiveParam(doc, possibleParen.NextSibling(), pos)
					return fnName, argIndex
				}
			}
		}
	}
	return "", 0
}

func possibleActiveParam(doc document.Document, node *sitter.Node, pos protocol.Position) uint32 {
	pt := query.PositionToPoint(pos)
	argIndex := uint32(0)
	for n := node; n != nil; n = n.NextSibling() {
		inRange := query.PointBeforeOrEqual(n.StartPoint(), pt) &&
			query.PointBeforeOrEqual(n.EndPoint(), pt)
		if !inRange {
			break
		}

		if !n.IsNamed() && doc.Content(n) == "," {
			argIndex++
		}
	}
	return argIndex
}
