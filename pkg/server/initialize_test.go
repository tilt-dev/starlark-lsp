package server_test

import (
	"testing"

	"go.lsp.dev/protocol"
)

func TestInitialize(t *testing.T) {
	f := newFixture(t)

	var resp protocol.InitializeResult
	f.mustEditorCall(protocol.MethodInitialize, protocol.InitializeParams{}, &resp)

	expected := protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			// N.B. this field is interface{} so we need to compare by JSON
			TextDocumentSync: protocol.TextDocumentSyncOptions{
				Change:    protocol.TextDocumentSyncKindFull,
				OpenClose: true,
				Save: &protocol.SaveOptions{
					IncludeText: true,
				},
			},
			SignatureHelpProvider: &protocol.SignatureHelpOptions{
				TriggerCharacters:   []string{"("},
				RetriggerCharacters: []string{","},
			},
			DocumentSymbolProvider: true,
		},
	}
	requireJsonEqual(t, expected, resp)
}
