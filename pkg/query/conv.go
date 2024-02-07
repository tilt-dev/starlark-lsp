package query

import (
	"fmt"
	"strconv"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/protocol"
)

// Unquote a Tree sitter string node into its string contents.
//
// Also accepts a parent module, block or expression statement containing a
// string node for convenience. If a non-string node is passed, Unquote will
// panic().
//
// Tree sitter parses a string literal into 2 or more child nodes
// representing the beginning/ending delimiters, and any number of escape
// sequences inside. Thus, the string
//
//    """hello\nTilted\nWorld"""
//
// gets parsed into a tree like:
//
//   module [0, 0] - [1, 0]
//     expression_statement [0, 0] - [0, 26]
//       string [0, 0] - [0, 26]
//         " [0, 0] - [0, 3]
//         escape_sequence [0, 8] - [0, 10]
//         escape_sequence [0, 16] - [0, 18]
//         " [0, 23] - [0, 26]
//
// Notably, there are no nodes to represent the contents in between the string
// delimiters and any escape sequences, so we have to extract those manually
// based on the start/end byte boundaries of the adjacent delimiter or escape
// sequence nodes.
func Unquote(input []byte, n *sitter.Node) string {
done:
	for {
		switch n.Type() {
		case NodeTypeModule,
			NodeTypeBlock,
			NodeTypeExpressionStatement:
			n = n.Child(0)
		case NodeTypeString:
			break done
		default:
			panic(fmt.Errorf("[Unquote:bug:unexpected node: %s: %s]", n.Type(), n.Content(input)))
		}
	}

	startDelim := n.Child(0)
	endDelim := n.Child(int(n.ChildCount() - 1))
	byteoffset := startDelim.EndByte()
	bytes := []byte{}

	for i := 1; i < int(n.ChildCount()-1); i++ {
		escape := n.Child(i)
		if byteoffset < escape.StartByte() {
			bytes = append(bytes, input[byteoffset:escape.StartByte()]...)
		}
		escseq := string(escape.Content(input))
		if escseq == "\\\n" {
			// ignore backslash-newline line continuation at the end of a line
			// per Starlark spec
			escseq = ""
		} else {
			// use Go Unquote to expand the escape sequence
			escseq, _ = strconv.Unquote(`"` + escseq + `"`)
		}
		bytes = append(bytes, []byte(escseq)...)
		byteoffset = escape.EndByte()
	}

	if byteoffset < endDelim.StartByte() {
		bytes = append(bytes, input[byteoffset:endDelim.StartByte()]...)
	}

	return string(bytes)
}

var strToProtocolMap = map[string]protocol.SymbolKind{
	"str":    protocol.SymbolKindString, // annotation
	"string": protocol.SymbolKindString, // node.type, annotation

	"bool":  protocol.SymbolKindBoolean, // annotation
	"true":  protocol.SymbolKindBoolean, // node.type
	"false": protocol.SymbolKindBoolean, // node.type

	"list":  protocol.SymbolKindArray, // node.type, annotation
	"tuple": protocol.SymbolKindArray, // annotation

	"dict":       protocol.SymbolKindObject, // annotation
	"dictionary": protocol.SymbolKindObject, // node.type
	"any":        protocol.SymbolKindObject, // annotation

	"int":     protocol.SymbolKindNumber, // annotation
	"float":   protocol.SymbolKindNumber, // node.type, annotation
	"integer": protocol.SymbolKindNumber, // node.type

	"none": protocol.SymbolKindNull, // node.type

	"callable":            protocol.SymbolKindFunction, // annotation
	"function_definition": protocol.SymbolKindFunction, // node.type, func def
	"call":                protocol.SymbolKindFunction, // node.type, assignment from func call
}

var KnownSymbolKinds map[protocol.SymbolKind]bool = map[protocol.SymbolKind]bool{
	protocol.SymbolKindBoolean: true,
	protocol.SymbolKindArray:   true, // list
	protocol.SymbolKindObject:  true, // dict
	protocol.SymbolKindNumber:  true, // float, int
	protocol.SymbolKindNull:    true,
	protocol.SymbolKindString:  true,
}

func StrToSymbolKind(s string) protocol.SymbolKind {
	if kind, found := strToProtocolMap[s]; found {
		return kind
	}
	return 0
}

func SymbolKindToBuiltinType(kind protocol.SymbolKind) string {
	switch kind {
	case protocol.SymbolKindString:
		return "String"
	case protocol.SymbolKindArray:
		return "List"
	case protocol.SymbolKindObject:
		return "Dict"
	}
	return ""
}

func StrToSymbolKindAndType(s string) (protocol.SymbolKind, string) {
	if s == "" {
		return 0, s
	}
	sl := strings.ToLower(s)
	kind := StrToSymbolKind(sl)
	if kind != 0 {
		if builtinType := SymbolKindToBuiltinType(kind); builtinType != "" {
			return kind, builtinType
		}
		return kind, sl // known symbol kind, return lowcase
	}
	return kind, s // return the original type
}

func NodeToSymbolKind(n *sitter.Node) protocol.SymbolKind {
	return StrToSymbolKind(n.Type())
}

// returns corresponding protocol.symbolKind and type parsed
func AnnotationNodeToSymbolKindAndType(doc DocumentContent, n *sitter.Node) (protocol.SymbolKind, string) {
	// if the type has a subscript like 'List[str]', use 'List' as the type
	if n.ChildCount() > 0 && n.Child(0).Type() == "subscript" {
		n = n.Child(0).ChildByFieldName("value")
	}
	return StrToSymbolKindAndType(doc.Content(n))
}
