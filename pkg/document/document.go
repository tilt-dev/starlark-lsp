package document

import (
	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/protocol"

	"github.com/tilt-dev/starlark-lsp/pkg/query"
)

type Document interface {
	Content(n *sitter.Node) string
	ContentRange(r sitter.Range) string

	Tree() *sitter.Tree
	Functions() map[string]protocol.SignatureInformation
	Symbols() []protocol.DocumentSymbol

	Copy() Document

	Close()
}

type NewDocumentFunc func(input []byte, tree *sitter.Tree) Document

func NewDocument(input []byte, tree *sitter.Tree) Document {
	return document{
		input: input,
		tree:  tree,
	}
}

func NewDocumentWithSymbols(input []byte, tree *sitter.Tree) Document {
	doc := document{
		input: input,
		tree:  tree,
	}
	doc.functions = query.Functions(doc, tree.RootNode())
	doc.symbols = query.DocumentSymbols(doc)
	return doc
}

func NodesToContent(doc Document, nodes []*sitter.Node) string {
	var content string
	if len(nodes) > 0 {
		content = doc.ContentRange(sitter.Range{
			StartByte: nodes[0].StartByte(),
			EndByte:   nodes[len(nodes)-1].EndByte(),
		})
	} else {
		content = doc.Content(doc.Tree().RootNode())
	}
	return content
}

type document struct {
	// input is the file as it exists in the editor buffer.
	input []byte

	// tree represents the parsed version of the document.
	tree *sitter.Tree

	functions map[string]protocol.SignatureInformation
	symbols   []protocol.DocumentSymbol
}

var _ Document = document{}

func (d document) Content(n *sitter.Node) string {
	return n.Content(d.input)
}

func (d document) ContentRange(r sitter.Range) string {
	return string(d.input[r.StartByte:r.EndByte])
}

func (d document) Tree() *sitter.Tree {
	return d.tree
}

func (d document) Functions() map[string]protocol.SignatureInformation {
	return d.functions
}

func (d document) Symbols() []protocol.DocumentSymbol {
	return d.symbols
}

func (d document) Close() {
	d.tree.Close()
}

// Copy creates a shallow copy of the Document.
//
// The Contents byte slice is returned as-is.
// A shallow copy of the Tree is made, as Tree-sitter trees are not thread-safe.
func (d document) Copy() Document {
	doc := document{
		input:     d.input,
		tree:      d.tree.Copy(),
		functions: make(map[string]protocol.SignatureInformation),
		symbols:   append([]protocol.DocumentSymbol{}, d.symbols...),
	}
	for fn := range d.functions {
		doc.functions[fn] = d.functions[fn]
	}
	return doc
}
