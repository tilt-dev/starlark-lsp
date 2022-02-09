package document

import sitter "github.com/smacker/go-tree-sitter"

type Document struct {
	// Contents of the file as they exist in the editor buffer.
	Contents []byte

	// Tree represents the parsed version of the document.
	Tree *sitter.Tree
}

func NewDocument(contents []byte, tree *sitter.Tree) Document {
	return Document{
		Contents: contents,
		Tree:     tree,
	}
}

func (d Document) Close() {
	d.Tree.Close()
}

// shallowClone creates a shallow copy of the Document.
//
// The Contents byte slice is returned as-is.
// A shallow copy of the Tree is made, as Tree-sitter trees are not thread-safe.
func (d Document) shallowClone() Document {
	return Document{
		Contents: d.Contents,
		Tree:     d.Tree.Copy(),
	}
}
