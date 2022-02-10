package server

import (
	"context"

	"go.lsp.dev/protocol"
)

func (s *Server) SignatureHelp(_ context.Context,
	params *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error) {

	doc, err := s.docs.Read(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}
	defer doc.Close()

	return s.analyzer.SignatureHelp(doc, params.Position), nil
}
