package query

import (
	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/protocol"
)

const structs = `
(assignment
  left: (identifier) @name
  right: (call
    function: ((identifier) @fname
               (#eq? @fname "struct"))
    arguments: (argument_list
      (keyword_argument)+ @fields
    )
  )
)
`

func Struct(doc DocumentContent, node *sitter.Node) Symbol {
	symbol := Symbol{}
	Query(node, structs, func(q *sitter.Query, match *sitter.QueryMatch) bool {
		for _, c := range match.Captures {
			switch q.CaptureNameForId(c.Index) {
			case "name":
				symbol.Name = doc.Content(c.Node)
				symbol.Kind = protocol.SymbolKindStruct
			case "fields":
				for node := c.Node; node != nil; node = node.NextNamedSibling() {
					if node.Type() != "keyword_argument" {
						continue
					}
					field := Symbol{}
					fieldNode := node.ChildByFieldName("name")
					fieldName := doc.Content(fieldNode)
					field.Name = fieldName
					field.Kind = protocol.SymbolKindField
					field.Location = protocol.Location{
						Range: NodeRange(c.Node),
						URI:   doc.URI(),
					}
					symbol.Children = append(symbol.Children, field)
				}
			}
		}
		return true
	})
	return symbol
}
