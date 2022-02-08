package server

import (
	"context"
	"strings"

	"go.lsp.dev/protocol"
)

func (s *Server) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) (err error) {
	return s.docs.Write(params.TextDocument.URI, strings.NewReader(params.TextDocument.Text))
}

func (s *Server) DidSave(ctx context.Context, params *protocol.DidSaveTextDocumentParams) (err error) {
	return s.docs.Write(params.TextDocument.URI, strings.NewReader(params.Text))
}

func (s *Server) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) (err error) {
	if len(params.ContentChanges) == 0 {
		return nil
	}
	if err := s.docs.Write(params.TextDocument.URI, strings.NewReader(params.ContentChanges[0].Text)); err != nil {
		return err
	}

	return nil
}
