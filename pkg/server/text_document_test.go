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

	var resp jsonrpc2.Response
	f.mustEditorCall(protocol.MethodTextDocumentDidChange, protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: uri.File("./test.star"),
			},
			Version: 1,
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{
			{Text: "foo + 1 = &^ 3"},
		},
	}, &resp)

	var params protocol.PublishDiagnosticsParams
	f.nextEditorEvent(protocol.MethodTextDocumentPublishDiagnostics, &params)
	require.Equal(t, uri.File("./test.star"), params.URI)
}
