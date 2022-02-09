package server

import (
	"context"

	"go.lsp.dev/protocol"
	"go.uber.org/zap"

	"github.com/tilt-dev/starlark-lsp/pkg/analysis"
)

func (s *Server) SignatureHelp(ctx context.Context,
	params *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error) {

	uri := params.TextDocument.URI
	logger := protocol.LoggerFromContext(ctx).
		With(zap.String("uri", string(uri)))

	doc, err := s.docs.Read(uri)
	if err != nil {
		return nil, err
	}
	defer doc.Close()

	node, ok := analysis.NodeAtPosition(doc, params.Position)
	if !ok {
		logger.Debug("no node at current editor position", positionField(params.Position))
	}

	// determine the function name we want a signature for; if we can't find
	// one, return early to avoid unnecessarily determining all functions in
	// scope
	// currently, this supports two cases:
	// 	(1) current node is inside of a `call`
	// 	(2) current node is inside of an ERROR block where first child is an
	// 		`identifier`
	var candidateFunctionName string
	for n := node; n != nil; n = n.Parent() {
		if n.Type() == "call" {
			candidateFunctionName = n.ChildByFieldName("function").Content(doc.Contents)
			break
		}
		if n.HasError() {
			// look for `foo(` and assume it's a function call - this could
			// happen if the closing `)` is not (yet) present or if there's
			// something invalid going on within the params
			possibleCall := n.NamedChild(0)
			if possibleCall != nil && possibleCall.Type() == analysis.NodeTypeIdentifier {
				possibleParen := possibleCall.NextSibling()
				if possibleParen != nil && possibleParen.Content(doc.Contents) == "(" {
					candidateFunctionName = possibleCall.Content(doc.Contents)
				}
			}
			break
		}
	}
	if candidateFunctionName == "" {
		logger.Debug("no call node in current tree")
		return &protocol.SignatureHelp{}, nil
	}

	sigs := analysis.Functions(doc, node)
	sig, ok := sigs[candidateFunctionName]
	if !ok {
		logger.Debug("no signature found", zap.String("function", candidateFunctionName))
		return &protocol.SignatureHelp{}, nil
	}

	// TODO(milas): determine active parameter based on position
	resp := &protocol.SignatureHelp{
		Signatures:      []protocol.SignatureInformation{sig},
		ActiveSignature: 0,
	}

	return resp, nil
}
