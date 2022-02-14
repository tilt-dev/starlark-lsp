package server

import (
	"context"

	"go.lsp.dev/protocol"

	"github.com/tilt-dev/starlark-lsp/pkg/query"
)

func (s *Server) DocumentSymbol(ctx context.Context,
	params *protocol.DocumentSymbolParams) ([]interface{}, error) {

	doc, err := s.docs.Read(params.TextDocument.URI)
	if err != nil {
		return nil, err
	}

	symbols := query.DocumentSymbols(doc)
	result := make([]interface{}, len(symbols))
	for i := range symbols {
		symbols[i].Location.URI = params.TextDocument.URI
		result[i] = symbols[i]
	}
	return result, nil
}
