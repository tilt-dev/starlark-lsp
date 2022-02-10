package server

import (
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
	"github.com/tilt-dev/starlark-lsp/pkg/middleware"
)

type Server struct {
	// FallbackServer stubs out protocol.Server, returning "not found" errors
	// for all methods; overridden methods on this object provide real
	// implementations
	FallbackServer

	// docs tracks open files for the editor including their contents and parse tree
	docs *document.Manager
	// notifier can send broadcasts to the editor (e.g. diagnostics)
	notifier protocol.Client
}

func NewServer(docManager *document.Manager, notifier protocol.Client) *Server {
	return &Server{
		docs:     docManager,
		notifier: notifier,
	}
}

func (s *Server) Handler(middlewares ...middleware.Middleware) jsonrpc2.Handler {
	serverHandler := protocol.ServerHandler(s, jsonrpc2.MethodNotFoundHandler)
	return middleware.WrapHandler(serverHandler, middlewares...)
}
