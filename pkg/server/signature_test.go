package server_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

func TestServer_SignatureHelp(t *testing.T) {
	f := newFixture(t)

	docURI := uri.File("./test.star")

	src := `
def foo():
  def foo(a, b: str, c=None, d: int=5) -> List[str]:
    foo(a,,)
`

	f.mustWriteDocument("./test.star", src)

	var resp protocol.SignatureHelp
	f.mustEditorCall(protocol.MethodTextDocumentSignatureHelp, protocol.SignatureHelpParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 3, Character: 10},
		},
	}, &resp)

	require.Len(t, resp.Signatures, 1)
	assert.Equal(t, uint32(0), resp.ActiveSignature)

	expected := protocol.SignatureHelp{
		Signatures: []protocol.SignatureInformation{
			{
				Label: `(a, b: str, c=None, d: int=5) -> List[str]`,
				Parameters: []protocol.ParameterInformation{
					{Label: "a"},
					{Label: "b: str"},
					{Label: "c=None"},
					{Label: "d: int=5"},
				},
			},
		},
		ActiveParameter: 1,
		ActiveSignature: 0,
	}
	requireJsonEqual(t, expected, resp)
}

func TestServer_SignatureHelp_ErrorAtCursor(t *testing.T) {
	f := newFixture(t)

	docURI := uri.File("./test.star")

	src := `
def foo(a, b):
  pass

foo(a,
`

	f.mustWriteDocument("./test.star", src)

	var resp protocol.SignatureHelp
	f.mustEditorCall(protocol.MethodTextDocumentSignatureHelp, protocol.SignatureHelpParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			// TODO(milas): change char 3 -> 6 so that this is actually within
			// 	the invalid "call" and assert that `b` is the active param
			Position: protocol.Position{Line: 4, Character: 3},
		},
	}, &resp)

	require.Len(t, resp.Signatures, 1)
	assert.Equal(t, uint32(0), resp.ActiveSignature)

	expected := protocol.SignatureHelp{
		Signatures: []protocol.SignatureInformation{
			{
				Label: `(a, b)`,
				Parameters: []protocol.ParameterInformation{
					{Label: "a"},
					{Label: "b"},
				},
			},
		},
		ActiveSignature: 0,
	}
	requireJsonEqual(t, expected, resp)
}

func TestServer_SignatureHelp_OutOfScope(t *testing.T) {
	f := newFixture(t)

	docURI := uri.File("./test.star")

	src := `
def foo():
  def bar():
    pass

bar()
`

	f.mustWriteDocument("./test.star", src)

	var resp protocol.SignatureHelp
	f.mustEditorCall(protocol.MethodTextDocumentSignatureHelp, protocol.SignatureHelpParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: docURI},
			Position:     protocol.Position{Line: 4, Character: 4},
		},
	}, &resp)

	// bar is not in scope for cursor position, so we shouldn't suggest it
	require.Empty(t, resp.Signatures, "No signatures should have been suggested")
}
