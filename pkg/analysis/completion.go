package analysis

import (
	"strings"

	"go.lsp.dev/protocol"
	"go.uber.org/zap"

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

	content := doc.Content(node)
	identifiers := strings.Split(content, ".")

	symbols := query.DocumentSymbols(doc)
	symbols = append(symbols, a.builtins.Symbols...)

	for i := 0; i < len(identifiers); i++ {
		if i < len(identifiers)-1 {
			sym := SymbolMatching(symbols, identifiers[i])
			symbols = sym.Children
		} else {
			symbols = SymbolsStartingWith(symbols, identifiers[i])
		}
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
