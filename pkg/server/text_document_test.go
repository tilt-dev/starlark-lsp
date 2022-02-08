package server_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestServer_DidChange(t *testing.T) {
	f := newFixture(t)

	const fileData = "foo + 1 = &^ 3"

	var resp jsonrpc2.Response
	f.mustEditorCall(protocol.MethodTextDocumentDidChange, protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: uri.File("./test.star"),
			},
			Version: 1,
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{
			{Text: fileData},
		},
	}, &resp)

	data, err := f.docManager.Read(uri.File("./test.star"))
	require.NoError(t, err, "Failed to read file from doc manager")
	require.Equal(t, fileData, string(data), "File contents did not match")
}
