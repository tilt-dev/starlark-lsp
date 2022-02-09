package main

import (
	"context"
	"io"
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
	defer func() {
		_ = logger.Sync()
	}()
	ctx = protocol.WithLogger(ctx, logger)

	logger.Debug("starlark-lsp launched")

	doneCh := make(chan struct{}, 1)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			if sig == os.Interrupt {
				doneCh <- struct{}{}
				// TODO(milas): give open conns a grace period to close gracefully
				cancel()
				os.Exit(0)
			}
		}
	}()

	if len(os.Args) > 1 && os.Args[1] == "--stdio" {
		logger.Debug("running in stdio mode")
		stdio := Stdio{stdin: os.Stdin, stdout: os.Stdout}
		setupConn(ctx, stdio, logger)
		<-doneCh
	} else {
		logger.Debug("running in socket mode", zap.String("addr", addr))
		var lc net.ListenConfig
		listener, err := lc.Listen(ctx, "tcp4", addr)
		if err != nil {
			panic(err)
		}
		defer func() {
			_ = listener.Close()
		}()
		for {
			conn, err := listener.Accept()
			if err != nil {
				panic(err)
			}
			setupConn(ctx, conn, logger)
		}
	}
}

func setupConn(ctx context.Context, conn io.ReadWriteCloser, logger *zap.Logger) {
	stream := jsonrpc2.NewStream(conn)
	jsonConn := jsonrpc2.NewConn(stream)

	client := protocol.ClientDispatcher(jsonConn, logger.Named("client"))

	docManager := document.NewDocumentManager()
	s := server.NewServer(docManager, client)
	h := s.Handler(server.StandardMiddleware...)

	jsonConn.Go(ctx, h)
}

type Stdio struct {
	stdin  *os.File
	stdout *os.File
}

func (s Stdio) Read(p []byte) (n int, err error) {
	return s.stdin.Read(p)
}

func (s Stdio) Write(p []byte) (n int, err error) {
	return s.stdout.Write(p)
}

func (s Stdio) Close() error {
	_ = s.stdin.Close()
	_ = s.stdout.Close()
	return nil
}
