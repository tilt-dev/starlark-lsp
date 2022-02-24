package analysis

import (
	"fmt"
	"strings"

	"go.lsp.dev/protocol"
	"go.uber.org/zap"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
	"github.com/tilt-dev/starlark-lsp/pkg/query"
)

func SymbolMatching(symbols []protocol.DocumentSymbol, name string) protocol.DocumentSymbol {
	for _, sym := range symbols {
		if sym.Name == name {
			return sym
		}
	}
	return protocol.DocumentSymbol{}
}

func SymbolsStartingWith(symbols []protocol.DocumentSymbol, prefix string) []protocol.DocumentSymbol {
	if prefix == "" {
		return symbols
	}
	result := []protocol.DocumentSymbol{}
	for _, sym := range symbols {
		if strings.HasPrefix(sym.Name, prefix) {
			result = append(result, sym)
		}
	}
	return result
}

func ToCompletionItemKind(k protocol.SymbolKind) protocol.CompletionItemKind {
	switch k {
	case protocol.SymbolKindField:
		return protocol.CompletionItemKindField
	case protocol.SymbolKindFunction:
		return protocol.CompletionItemKindFunction
	case protocol.SymbolKindMethod:
		return protocol.CompletionItemKindMethod
	default:
		return protocol.CompletionItemKindVariable
	}
}

func (a *Analyzer) Completion(doc document.Document, pos protocol.Position) *protocol.CompletionList {
	node, ok := query.NodeAtPosition(doc, pos)
	if !ok {
		return nil
	}
	node = adjustNodeForCompletion(node, query.PositionToPoint(pos))

	var symbols []protocol.DocumentSymbol
	if node != nil {
		content := doc.Content(node)
		a.logger.Debug("completion", zap.String("node", fmt.Sprintf("%.32s", content)), zap.String("type", node.Type()))
		symbols = a.completeAttributeExpression(doc, node, content, pos)
	}

	completionList := &protocol.CompletionList{
		Items: make([]protocol.CompletionItem, len(symbols)),
	}

	names := make([]string, len(symbols))
	for i, sym := range symbols {
		names[i] = sym.Name
		completionList.Items[i] = protocol.CompletionItem{
			Label:  sym.Name,
			Detail: sym.Detail,
			Kind:   ToCompletionItemKind(sym.Kind),
		}
	}

	a.logger.Debug("completion", zap.Strings("symbols", names))
	return completionList
}

func (a *Analyzer) completeAttributeExpression(doc document.Document, node *sitter.Node, content string, pos protocol.Position) []protocol.DocumentSymbol {
	// TODO(nicksieger): This is a naive way to parse an attribute expression
	// a.b.c. Parse the nodes instead.
	content = content[:pos.Character-node.StartPoint().Column]
	identifiers := strings.Split(content, ".")
	symbols := query.SymbolsInScope(doc, node)
	symbols = append(symbols, a.builtins.Symbols...)

	for i := 0; i < len(identifiers); i++ {
		if i < len(identifiers)-1 {
			sym := SymbolMatching(symbols, identifiers[i])
			symbols = sym.Children
			names := make([]string, len(symbols))
			for j, s := range symbols {
				names[j] = s.Name
			}
			a.logger.Debug("children", zap.String("id", identifiers[i]), zap.Strings("names", names))
		} else {
			symbols = SymbolsStartingWith(symbols, identifiers[i])
		}
	}

	return symbols
}

func adjustNodeForCompletion(node *sitter.Node, pt sitter.Point) *sitter.Node {
	for {
		reevaluate := false
		switch node.Type() {
		case query.NodeTypeString:
			// No completion inside a string
			node = nil
		case query.NodeTypeModule:
			// Advance to the first node that appears after the point
			if node.NamedChildCount() > 0 {
				for node = node.NamedChild(0); query.PointBefore(node.StartPoint(), pt); {
					next := node.NextNamedSibling()
					if next == nil {
						break
					}
					node = next
				}
			}
		case query.NodeTypeAttribute:
			switch node.Parent().Type() {
			case query.NodeTypeExpressionStatement:
				node = node.Parent()
			case query.NodeTypeAttribute:
				node = node.Parent()
				reevaluate = true
			}
		case query.NodeTypeIdentifier:
			// If inside an attribute expression, capture the larger expression for
			// completion.
			switch node.Parent().Type() {
			case query.NodeTypeExpressionStatement:
				node = node.Parent()
			case query.NodeTypeAttribute:
				node = node.Parent()
				reevaluate = true
			}
		case query.NodeTypeERROR:
			node = node.Parent()
			reevaluate = true
		}

		if !reevaluate {
			break
		}
	}
	return node
}
