package server

import (
	"context"

	"go.lsp.dev/protocol"
)

func (s *Server) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) (err error) {
	uri := params.TextDocument.URI
	contents := []byte(params.TextDocument.Text)
	return s.docs.Write(ctx, uri, contents)
}

func (s *Server) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) (err error) {
	if len(params.ContentChanges) == 0 {
		return nil
	}

	uri := params.TextDocument.URI
	contents := []byte(params.ContentChanges[0].Text)
	return s.docs.Write(ctx, uri, contents)
}

func (s *Server) DidSave(ctx context.Context, params *protocol.DidSaveTextDocumentParams) (err error) {
	uri := params.TextDocument.URI
	contents := []byte(params.Text)
	return s.docs.Write(ctx, uri, contents)
}

func (s *Server) DidClose(_ context.Context, params *protocol.DidCloseTextDocumentParams) (err error) {
	s.docs.Remove(params.TextDocument.URI)
	return nil
}
