package query

import (
	"fmt"

	"go.lsp.dev/protocol"

	sitter "github.com/smacker/go-tree-sitter"
)

const (
	Binded protocol.SymbolTag = 1 << iota // The symbol is binded to some other symbol
)

var KnownKinds map[protocol.SymbolKind]bool = map[protocol.SymbolKind]bool{
	protocol.SymbolKindBoolean: true,
	protocol.SymbolKindArray:   true, // list
	protocol.SymbolKindObject:  true, // dict
	protocol.SymbolKindNumber:  true, // float, int
	protocol.SymbolKindNull:    true,
	protocol.SymbolKindString:  true,
}

// Get all symbols defined at the same level as the given node.
// If before != nil, only include symbols that appear before that node.
func SiblingSymbols(doc DocumentContent, node, before *sitter.Node) []Symbol {
	var symbols []Symbol
	for n := node; n != nil && NodeBefore(n, before); n = n.NextNamedSibling() {
		var symbol Symbol

		switch n.Type() {
		case NodeTypeExpressionStatement:
			symbol = ExtractVariableAssignment(doc, n)
		case NodeTypeFunctionDef:
			sig := ExtractSignature(doc, n)
			symbol = sig.Symbol()
		}

		if symbol.Name != "" {
			symbols = append(symbols, symbol)
		}
	}
	return symbols
}

// assignment expression could appear in various forms
// i: int  # varable definition (with or w/o hint/annotation)
// i = 1  # direct assignment of well known kind, e.g. int
// i = j  # from other var
// i = foo()  # from function res
func ExtractVariableAssignment(doc DocumentContent, n *sitter.Node) Symbol {
	if n.Type() != NodeTypeExpressionStatement {
		panic(fmt.Errorf("invalid node type: %s", n.Type()))
	}

	var symbol Symbol
	assignment := n.NamedChild(0)
	if assignment == nil || assignment.Type() != "assignment" {
		return symbol
	}
	symbol.Name = doc.Content(assignment.ChildByFieldName("left"))
	val := assignment.ChildByFieldName("right")
	typeNode := assignment.ChildByFieldName("type")

	var kind protocol.SymbolKind
	var typeStr string

	if typeNode != nil {
		kind, typeStr = AnnotationNodeToSymbolKindAndType(doc, typeNode)
	} else if val != nil {
		// set exact kind and corresponding type for protocol-defined kinds, e.g. String, Array(list), Object(Dict), Number, Bool, etc..
		// for others set their kind to Variable and type to Rval, i.e. `func()` or for assigned symbol name
		kind = NodeToSymbolKind(val)

		if kind == protocol.SymbolKindFunction || kind == 0 {
			kind = protocol.SymbolKindVariable
			typeStr = doc.Content(val)
		} else {
			typeStr = SymbolKindToBuiltinType(kind) // String, List, Dict, etc .. or ""
		}
	}

	if kind == 0 {
		kind = protocol.SymbolKindVariable
	}
	symbol.Kind = kind
	symbol.Type = typeStr

	symbol.Location = protocol.Location{
		Range: NodeRange(n),
		URI:   doc.URI(),
	}

	// Look for possible docstring for the assigned variable
	if n.NextNamedSibling() != nil && n.NextNamedSibling().Type() == NodeTypeExpressionStatement {
		if ch := n.NextNamedSibling().NamedChild(0); ch != nil && ch.Type() == NodeTypeString {
			symbol.Detail = Unquote(doc.Input(), ch)
		}
	}
	return symbol
}

// A node is in the scope of the top level module if there are no function
// definitions in the ancestry of the node.
func IsModuleScope(doc DocumentContent, node *sitter.Node) bool {
	for n := node.Parent(); n != nil; n = n.Parent() {
		if n.Type() == NodeTypeFunctionDef {
			return false
		} else if n.Type() == NodeTypeERROR && n.ChildCount() >= 2 {
			// upon dot completion inside the function, if the current node is "." (dot), then tree-sitter will fail
			// to parse valid syntax, therefore instead of functionDefinition node there will be ERROR node.
			// Let's check its direct named kids. It should have identifier (errored function definition) and parameters
			if child1, child2 := n.NamedChild(0), n.NamedChild(1); child1 != nil && child2 != nil {
				if child1.Type() == NodeTypeIdentifier && child2.Type() == NodeTypeParameters {
					return false
				}
			}
		}
	}
	return true
}

// Get all symbols defined in scopes at or above the level of the given node,
// excluding symbols from the top-level module (document symbols).
func SymbolsInScope(doc DocumentContent, node *sitter.Node) []Symbol {
	var symbols []Symbol

	appendParameters := func(fnNode *sitter.Node) {
		sig := ExtractSignature(doc, fnNode)
		for _, p := range sig.Params {
			symbols = append(symbols, p.Symbol())
		}
	}

	// While we are in the current scope, only include symbols defined before
	// the provided node.
	before := node
	n := node
	for ; n.Parent() != nil && !IsModuleScope(doc, n); n = n.Parent() {
		// A function definition creates an enclosing scope, where all symbols
		// in the parent scope are visible. After that point, don't specify a
		// before node.
		if n.Type() == NodeTypeFunctionDef {
			before = nil
			appendParameters(n)
		}

		symbols = append(symbols, SiblingSymbols(doc, n.Parent().NamedChild(0), before)...)
	}
	// Append parameters of parent function that's in module scope
	if n.Type() == NodeTypeFunctionDef {
		appendParameters(n)
	}
	return symbols
}

// DocumentSymbols returns all symbols with document-wide visibility.
func DocumentSymbols(doc DocumentContent) []Symbol {
	return SiblingSymbols(doc, doc.Tree().RootNode().NamedChild(0), nil)
}

// Returns only the symbols that occur before the node given if any, otherwise return all symbols.
func SymbolsBefore(symbols []protocol.DocumentSymbol, before *sitter.Node) []protocol.DocumentSymbol {
	if before == nil {
		return symbols
	}
	result := []protocol.DocumentSymbol{}
	for _, sym := range symbols {
		symStart := PositionToPoint(sym.Range.Start)
		if PointBefore(symStart, before.StartPoint()) {
			result = append(result, sym)
		}
	}
	return result
}

type Symbol struct {
	Name           string
	Detail         string
	Kind           protocol.SymbolKind
	Tags           []protocol.SymbolTag
	Location       protocol.Location
	SelectionRange protocol.Range
	Children       []Symbol
	Type           string
}

// builtins (e.g., `False`) have no location
func (s Symbol) HasLocation() bool {
	return s.Location.URI != ""
}

func (s Symbol) GetType() string {
	if t := SymbolKindToBuiltinType(s.Kind); t != "" {
		return t
	}
	return s.Type
}
