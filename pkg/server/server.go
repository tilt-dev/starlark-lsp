package server

import (
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
	"github.com/tilt-dev/starlark-lsp/pkg/middleware"
)

type Server struct {
	FallbackServer

	docs   *document.Manager
	client protocol.Client
}

func NewServer(docManager *document.Manager, client protocol.Client) *Server {
	return &Server{
		docs:   docManager,
		client: client,
	}
}

func (s *Server) Handler(middlewares ...middleware.Middleware) jsonrpc2.Handler {
	serverHandler := protocol.ServerHandler(s, jsonrpc2.MethodNotFoundHandler)
	return middleware.WrapHandler(serverHandler, middlewares...)
}
