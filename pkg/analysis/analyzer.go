package analysis

import (
	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/protocol"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
	"github.com/tilt-dev/starlark-lsp/pkg/query"
)

type Analyzer struct {
	builtinFunctions map[string]protocol.SignatureInformation
	builtinSymbols   []protocol.SymbolInformation
}

type AnalyzerOption func(*Analyzer)

func NewAnalyzer(opts ...AnalyzerOption) *Analyzer {
	analyzer := Analyzer{
		builtinFunctions: make(map[string]protocol.SignatureInformation),
	}
	for _, opt := range opts {
		opt(&analyzer)
	}
	return &analyzer
}

func WithBuiltinFunctions(sigs map[string]protocol.SignatureInformation) AnalyzerOption {
	return func(analyzer *Analyzer) {
		for fn, sig := range sigs {
			analyzer.builtinFunctions[fn] = sig
		}
	}
}

func WithBuiltinSymbols(symbols []protocol.SymbolInformation) AnalyzerOption {
	return func(analyzer *Analyzer) {
		analyzer.builtinSymbols = append(analyzer.builtinSymbols, symbols...)
	}
}

func (a *Analyzer) SignatureHelp(doc document.Document, pos protocol.Position) *protocol.SignatureHelp {
	node, ok := query.NamedNodeAtPosition(doc, pos)
	if !ok {
		return nil
	}

	fnName := possibleCallFunctionName(doc, node)
	if fnName == "" {
		// avoid computing function defs
		return nil
	}

	for n := node; n != nil; n = n.Parent() {
		sig, ok := query.Function(doc, n, fnName)
		if ok {
			// TODO(milas): determine active parameter based on position
			return &protocol.SignatureHelp{
				Signatures:      []protocol.SignatureInformation{sig},
				ActiveSignature: 0,
			}
		}
	}

	if sig, ok := a.builtinFunctions[fnName]; ok {
		// TODO(milas): determine active parameter based on position
		return &protocol.SignatureHelp{
			Signatures:      []protocol.SignatureInformation{sig},
			ActiveSignature: 0,
		}
	}

	return nil
}

// possibleCallFunctionName attempts to find the name of the function for a
// `call`.
//
// Currently, this supports two cases:
// 	(1) Current node is inside of a `call`
// 	(2) Current node is inside of an ERROR block where first child is an
// 		`identifier`
func possibleCallFunctionName(doc document.Document, node *sitter.Node) string {
	for n := node; n != nil; n = n.Parent() {
		if n.Type() == "call" {
			return n.ChildByFieldName("function").Content(doc.Contents)
		}
		if n.HasError() {
			// look for `foo(` and assume it's a function call - this could
			// happen if the closing `)` is not (yet) present or if there's
			// something invalid going on within the params
			possibleCall := n.NamedChild(0)
			if possibleCall != nil && possibleCall.Type() == query.NodeTypeIdentifier {
				possibleParen := possibleCall.NextSibling()
				if possibleParen != nil && possibleParen.Content(doc.Contents) == "(" {
					return possibleCall.Content(doc.Contents)
				}
			}
			break
		}
	}
	return ""
}
