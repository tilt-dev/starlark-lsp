package query

import (
	sitter "github.com/smacker/go-tree-sitter"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
)

func ExtractIdentifiers(doc document.Document, nodes []*sitter.Node, limit *sitter.Point) []string {
	identifiers := []string{}
	for _, node := range nodes {
		Query(node, Identifiers, func(q *sitter.Query, match *sitter.QueryMatch) bool {
			for _, c := range match.Captures {
				switch q.CaptureNameForId(c.Index) {
				case "id":
					if limit != nil && PointAfter(c.Node.StartPoint(), *limit) {
						identifiers = append(identifiers, "")
					} else {
						identifiers = append(identifiers, doc.Content(c.Node))
					}
				case "trailing-dot":
					identifiers = append(identifiers, "")
				case "module":
					if c.Node.ChildCount() == 0 {
						identifiers = append(identifiers, "")
					}
				}
			}
			return true
		})
	}
	return identifiers
}
