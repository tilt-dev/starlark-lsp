package server_test

import (
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

	doc, err := f.docManager.Read(uri.File("./test.star"))
	require.NoError(t, err, "Failed to read file from doc manager")
	require.Equal(t, fileData, string(doc.Contents), "File contents did not match")
	require.NotNil(t, doc.Tree, "Tree-sitter tree was nil")
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

	doc, err := f.docManager.Read(uri.File("./test.star"))
	require.NoError(t, err, "Failed to read file from doc manager")
	require.Equal(t, fileData, string(doc.Contents), "File contents did not match")
	require.NotNil(t, doc.Tree, "Tree-sitter tree was nil")
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

	doc, err := f.docManager.Read(uri.File("./test.star"))
	require.NoError(t, err, "Failed to read file from doc manager")
	require.Equal(t, fileData, string(doc.Contents), "File contents did not match")
	require.NotNil(t, doc.Tree, "Tree-sitter tree was nil")
}

func TestServer_DidClose(t *testing.T) {
	f := newFixture(t)

	const fileData = "foo + 1 = &^ 3"

	var resp jsonrpc2.Response
	f.mustEditorCall(protocol.MethodTextDocumentDidOpen, protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:  uri.File("./test.star"),
			Text: fileData,
		},
	}, &resp)

	doc, err := f.docManager.Read(uri.File("./test.star"))
	require.NoError(t, err, "Failed to read file from doc manager")
	require.Equal(t, fileData, string(doc.Contents), "File contents did not match")
	require.NotNil(t, doc.Tree, "Tree-sitter tree was nil")

	f.mustEditorCall(protocol.MethodTextDocumentDidClose, protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri.File("./test.star")},
	}, &resp)

	doc, err = f.docManager.Read(uri.File("./test.star"))
	require.EqualError(t, err, "file does not exist", "Document should no longer exist")
	require.Zero(t, doc, "Document was not zero-value")
}
