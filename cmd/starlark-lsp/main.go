package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
	"github.com/tilt-dev/starlark-lsp/pkg/server"
)

func main() {
	// TODO(milas): add flags to specify options (e.g. port)
	const addr = "localhost:7654"

	ctx, cancel := context.WithCancel(context.Background())

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	ctx = protocol.WithLogger(ctx, logger)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			if sig == os.Interrupt {
				// TODO(milas): give open conns a grace period to close gracefully
				cancel()
				os.Exit(0)
			}
		}
	}()

	var lc net.ListenConfig
	listener, err := lc.Listen(ctx, "tcp4", addr)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = listener.Close()
	}()

	fmt.Printf("listening on %s\n", addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		stream := jsonrpc2.NewStream(conn)
		jsonConn := jsonrpc2.NewConn(stream)

		client := protocol.ClientDispatcher(jsonConn, logger.Named("client"))

		docManager := document.NewDocumentManager()
		s := server.NewServer(docManager, client)
		h := s.Handler(server.StandardMiddleware...)

		jsonConn.Go(ctx, h)
	}
}
