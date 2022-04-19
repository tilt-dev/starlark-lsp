package query

import (
	sitter "github.com/smacker/go-tree-sitter"
)

const methodsAndFields = `
(class_definition
  name: (identifier) @name
  body: (block ([
    (expression_statement (assignment)) @field
    (function_definition) @method
    (_)
  ])*)
)
`

type Class struct {
	Name    string
	Methods []Signature
	Fields  []Symbol
}

func Classes(doc DocumentContent, node *sitter.Node) []Class {
	classes := []Class{}
	Query(node, []byte(methodsAndFields), func(q *sitter.Query, match *sitter.QueryMatch) bool {
		curr := Class{}
		for _, c := range match.Captures {
			switch q.CaptureNameForId(c.Index) {
			case "name":
				curr.Name = doc.Content(c.Node)
			case "field":
				curr.Fields = append(curr.Fields, ExtractAssignment(doc, c.Node))
			case "method":
				meth := ExtractSignature(doc, c.Node)
				// Remove Python "self" parameter if present
				if len(meth.Params) > 0 && meth.Params[0].Content == "self" {
					meth.Params = meth.Params[1:]
				}
				curr.Methods = append(curr.Methods, meth)
			}
		}
		if curr.Name != "" && (len(curr.Methods) > 0 || len(curr.Fields) > 0) {
			classes = append(classes, curr)
		}
		return true
	})

	return classes
}
