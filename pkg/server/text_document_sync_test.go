package server_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestServer_DidOpen(t *testing.T) {
	f := newFixture(t)

	const fileData = "foo + 1 = &^ 3"

	var resp jsonrpc2.Response
	f.mustEditorCall(protocol.MethodTextDocumentDidOpen, protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  uri.File("./test.star"),
			Text: fileData,
		},
	}, &resp)

	f.requireDocContents("./test.star", fileData)
}

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

	f.requireDocContents("./test.star", fileData)
}

func TestServer_DidSave(t *testing.T) {
	f := newFixture(t)

	const fileData = "foo + 1 = &^ 3"

	var resp jsonrpc2.Response
	f.mustEditorCall(protocol.MethodTextDocumentDidSave, protocol.DidSaveTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{
			URI: uri.File("./test.star"),
		},
		Text: fileData,
	}, &resp)

	f.requireDocContents("./test.star", fileData)
}

func TestServer_DidClose(t *testing.T) {
	f := newFixture(t)

	const fileData = "foo + 1 = &^ 3"

	f.mustWriteDocument("./test.star", fileData)

	var resp jsonrpc2.Response
	f.mustEditorCall(protocol.MethodTextDocumentDidClose, protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri.File("./test.star")},
	}, &resp)

	doc, err := f.docManager.Read(f.ctx, uri.File("./test.star"))
	require.ErrorIs(t, os.ErrNotExist, err, "file does not exist", "Document should no longer exist")
	require.Zero(t, doc, "Document was not zero-value")
}
