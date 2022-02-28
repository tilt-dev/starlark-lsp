package analysis

import (
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
	pt := query.PositionToPoint(pos)
	nodes := nodesAtPointForCompletion(doc, pt)

	var symbols []protocol.DocumentSymbol
	if len(nodes) > 0 {
		symbols = a.completeExpression(doc, nodes, pt)
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

	if len(names) > 0 {
		a.logger.Debug("completion result", zap.Strings("symbols", names))
	}
	return completionList
}

func (a *Analyzer) completeExpression(doc document.Document, nodes []*sitter.Node, pt sitter.Point) []protocol.DocumentSymbol {
	symbols := append(query.SymbolsInScope(doc, nodes[len(nodes)-1]), a.builtins.Symbols...)
	identifiers := query.ExtractIdentifiers(doc, nodes, &pt)

	a.logger.Debug("completion attempt",
		zap.String("code", doc.ContentRange(sitter.Range{
			StartByte: nodes[0].StartByte(),
			EndByte:   nodes[len(nodes)-1].EndByte(),
		})),
		zap.Strings("nodes", func() []string {
			types := make([]string, len(nodes))
			for i, n := range nodes {
				types[i] = n.Type()
			}
			return types
		}()),
		zap.Strings("identifiers", identifiers),
	)

	for i, id := range identifiers {
		if i < len(identifiers)-1 {
			sym := SymbolMatching(symbols, id)
			symbols = sym.Children
			a.logger.Debug("children",
				zap.String("id", id),
				zap.Strings("names", func() []string {
					names := make([]string, len(symbols))
					for j, s := range symbols {
						names[j] = s.Name
					}
					return names
				}()))
		} else {
			symbols = SymbolsStartingWith(symbols, id)
		}
	}

	return symbols
}

func nodesAtPointForCompletion(doc document.Document, pt sitter.Point) []*sitter.Node {
	node, ok := query.NodeAtPoint(doc, pt)
	if !ok {
		return []*sitter.Node{}
	}
	return nodesForCompletion(node, pt)
}

// Zoom in or out from the node to include adjacent attribute expressions, so we can
// complete starting from the top-most attribute expression.
func nodesForCompletion(node *sitter.Node, pt sitter.Point) []*sitter.Node {
	nodes := []*sitter.Node{}
	switch node.Type() {
	case query.NodeTypeString, query.NodeTypeComment:
		if query.PointCovered(pt, node) {
			// No completion inside a string or comment
			return nodes
		}
	case query.NodeTypeModule:
		// Sometimes the top-level module is the most granular node due to
		// location of the point being between children, in this case, advance
		// to the first child node that appears after the point
		if node.NamedChildCount() > 0 {
			for node = node.NamedChild(0); node != nil && query.PointBefore(node.StartPoint(), pt); {
				next := node.NextNamedSibling()
				if next == nil {
					break
				}
				node = next
			}
		}
	case query.NodeTypeAttribute, query.NodeTypeIdentifier:
		// If inside an attribute expression, capture the larger expression for
		// completion.
		switch node.Parent().Type() {
		case query.NodeTypeAttribute:
			nodes = nodesForCompletion(node.Parent(), pt)
		}
	case query.NodeTypeERROR:
		if node.PrevNamedSibling() != nil {
			nodes = nodesForCompletion(node.PrevNamedSibling(), pt)
			nodes = append(nodes, node)
		} else {
			nodes = nodesForCompletion(node.Parent(), pt)
		}
	}

	if len(nodes) == 0 {
		nodes = append(nodes, node)
	}
	return nodes
}
