package query

import (
	"fmt"
	"strings"

	"go.lsp.dev/protocol"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/tilt-dev/starlark-lsp/pkg/docstring"
)

// Functions finds all function definitions that are direct children of the provided sitter.Node.
func Functions(doc DocumentContent, node *sitter.Node) map[string]protocol.SignatureInformation {
	signatures := make(map[string]protocol.SignatureInformation)

	// N.B. we don't use a query here for a couple reasons:
	// 	(1) Tree-sitter doesn't support bounding the depth, and we only want
	//		direct descendants (to avoid matching on functions in nested scopes)
	//		See https://github.com/tree-sitter/tree-sitter/issues/1212.
	//	(2) function_definition nodes have named fields for what we care about,
	//		which makes it easy to get the data without using a query to help
	//		massage/standardize it (for example, we do this for params since
	//		there are multiple type of param values)
	for n := node.NamedChild(0); n != nil; n = n.NextNamedSibling() {
		if n.Type() != NodeTypeFunctionDef {
			continue
		}
		sig := extractSignature(doc, n)
		signatures[sig.name] = sig.signatureInfo()
	}

	return signatures
}

// Function finds a function definition for the given function name that is a direct child of the provided sitter.Node.
func Function(doc DocumentContent, node *sitter.Node, fnName string) (protocol.SignatureInformation, bool) {
	for n := node.NamedChild(0); n != nil; n = n.NextNamedSibling() {
		if n.Type() != NodeTypeFunctionDef {
			continue
		}
		curFuncName := doc.Content(n.ChildByFieldName(FieldName))
		if curFuncName == fnName {
			sig := extractSignature(doc, n)
			return sig.signatureInfo(), true
		}
	}
	return protocol.SignatureInformation{}, false
}

type signature struct {
	name       string
	params     []parameter
	returnType string
	docs       docstring.Parsed
	node       *sitter.Node
}

func (s signature) signatureInfo() protocol.SignatureInformation {
	params := make([]protocol.ParameterInformation, len(s.params))
	for i, param := range s.params {
		params[i] = param.paramInfo(s.docs)
	}
	sigInfo := protocol.SignatureInformation{
		Label:      s.label(),
		Parameters: params,
	}
	if s.docs.Description != "" {
		sigInfo.Documentation = protocol.MarkupContent{
			Kind:  protocol.PlainText,
			Value: s.docs.Description,
		}
	}

	return sigInfo
}

// Label produces a human-readable label for a function signature.
//
// It's modeled to behave similarly to VSCode Python signature labels.
func (s signature) label() string {
	var sb strings.Builder
	sb.WriteRune('(')
	for i := range s.params {
		sb.WriteString(s.params[i].content)
		if i != len(s.params)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString(")")
	if s.returnType != "" {
		sb.WriteString(" -> ")
		sb.WriteString(s.returnType)
	}
	return sb.String()
}

func (s signature) symbol() protocol.DocumentSymbol {
	return protocol.DocumentSymbol{
		Name:   s.name,
		Kind:   protocol.SymbolKindFunction,
		Detail: s.label(),
		Range:  NodeRange(s.node),
	}
}

func extractSignature(doc DocumentContent, n *sitter.Node) signature {
	if n.Type() != NodeTypeFunctionDef {
		panic(fmt.Errorf("invalid node type: %s", n.Type()))
	}

	fnName := doc.Content(n.ChildByFieldName(FieldName))
	fnDocs := extractDocstring(doc, n.ChildByFieldName(FieldBody))

	// params might be empty but a node for `()` will still exist
	params := extractParameters(doc, fnDocs, n.ChildByFieldName(FieldParameters))
	// unlike name + params, returnType is optional
	var returnType string
	if rtNode := n.ChildByFieldName(FieldReturnType); rtNode != nil {
		returnType = doc.Content(rtNode)
	}

	return signature{
		name:       fnName,
		params:     params,
		returnType: returnType,
		docs:       fnDocs,
		node:       n,
	}
}

func extractDocstring(doc DocumentContent, n *sitter.Node) docstring.Parsed {
	if n.Type() != NodeTypeBlock {
		panic(fmt.Errorf("invalid node type: %s", n.Type()))
	}

	if exprNode := n.NamedChild(0); exprNode != nil && exprNode.Type() == NodeTypeExpressionStatement {
		if docStringNode := exprNode.NamedChild(0); docStringNode != nil && docStringNode.Type() == NodeTypeString {
			return docstring.Parse(Unquote(doc.Input(), docStringNode))
		}
	}

	// we don't return any sort of bool about success because even if there's
	// a string in the right place in the syntax tree, it might not even be a
	// valid docstring, so this is all on a best effort basis
	return docstring.Parsed{}
}
