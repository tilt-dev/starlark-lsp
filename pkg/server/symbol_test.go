package server_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestServer_DocumentSymbol(t *testing.T) {
	f := newFixture(t)

	docURI := uri.File("./test.star")
	doc := `
x = a(3)
y = None
z = True
`

	f.mustWriteDocument("./test.star", doc)

	var resp []protocol.DocumentSymbol
	f.mustEditorCall(protocol.MethodTextDocumentDocumentSymbol, protocol.DocumentSymbolParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
	}, &resp)

	require.Len(t, resp, 3)
}
