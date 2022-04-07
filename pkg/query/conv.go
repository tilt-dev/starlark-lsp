package query

import (
	"strconv"

	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/protocol"
)

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
			return ""
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
			byteoffset = escape.EndByte()
		}
		escseq := string(input[escape.StartByte():escape.EndByte()])
		if escseq == "\\\n" {
			// ignore backslash-newline line continuation at the end of a line per Starlark spec
			escseq = ""
		} else {
			// use Go Unquote to expand the escape sequence
			escseq, _ = strconv.Unquote(`"` + escseq + `"`)
		}
		bytes = append(bytes, []byte(escseq)...)
	}

	if byteoffset < endDelim.StartByte() {
		bytes = append(bytes, input[byteoffset:endDelim.StartByte()]...)
	}

	return string(bytes)
}

func nodeTypeToSymbolKind(n *sitter.Node) protocol.SymbolKind {
	switch n.Type() {
	case "true":
		return protocol.SymbolKindBoolean
	case "false":
		return protocol.SymbolKindBoolean
	case "list":
		return protocol.SymbolKindArray
	case "dictionary":
		return protocol.SymbolKindObject
	case "integer":
		return protocol.SymbolKindNumber
	case "float":
		return protocol.SymbolKindNumber
	case "none":
		return protocol.SymbolKindNull
	case "string":
		return protocol.SymbolKindString
	case "function_definition":
		return protocol.SymbolKindFunction
	}
	return 0
}
