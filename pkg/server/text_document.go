package server

import (
	"context"
	"strings"

	"go.lsp.dev/protocol"
	"go.starlark.net/syntax"
	"go.uber.org/zap"
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

	go func() {
		var diagnostics []protocol.Diagnostic

		// TODO(milas): use the actual filename (need to share the URI fix-up logic)
		f, err := syntax.Parse("file.star", strings.NewReader(params.ContentChanges[0].Text), 0644)
		if parseErr, ok := err.(syntax.Error); ok {
			pos := protocol.Position{
				Line:      uint32(parseErr.Pos.Line - 1),
				Character: uint32(parseErr.Pos.Col - 1),
			}

			diag := protocol.Diagnostic{
				Range: protocol.Range{
					Start: pos,
					End:   pos,
				},
				Severity: protocol.DiagnosticSeverityError,
				Source:   "starlark",
				Message:  parseErr.Msg,
			}
			diagnostics = append(diagnostics, diag)
		} else if err != nil {
			protocol.LoggerFromContext(ctx).Error("Internal parse error", zap.Error(err))
		} else {
			protocol.LoggerFromContext(ctx).Info("Parse success", zap.Int("stmt_count", len(f.Stmts)))
		}

		err = s.client.PublishDiagnostics(ctx, &protocol.PublishDiagnosticsParams{
			URI:         params.TextDocument.URI,
			Version:     uint32(params.TextDocument.Version),
			Diagnostics: diagnostics,
		})
		if err != nil {
			protocol.LoggerFromContext(ctx).Error("Failed to publish diagnostics", zap.Error(err))
		}
	}()

	return nil
}
