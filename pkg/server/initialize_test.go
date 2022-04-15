package server_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
				RetriggerCharacters: []string{",", "="},
			},
			DocumentSymbolProvider: true,
			CompletionProvider: &protocol.CompletionOptions{
				TriggerCharacters: []string{"."},
			},
			HoverProvider: true,
		},
	}
	requireJsonEqual(t, expected, resp)

	var logParams protocol.LogMessageParams
	f.requireNextEditorEvent(protocol.MethodWindowLogMessage, &logParams)
	assert.Equal(t, protocol.MessageTypeLog, logParams.Type)
	assert.Equal(t, "Starlark LSP server initialized", logParams.Message)
}
