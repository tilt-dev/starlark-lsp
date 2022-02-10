package cli

import (
	"context"
	"io"
	"net"
	"os"

	"github.com/spf13/cobra"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
	"github.com/tilt-dev/starlark-lsp/pkg/server"
)

type startCmd struct {
	*cobra.Command
	address string
}

func newStartCmd() *startCmd {
	cmd := startCmd{
		Command: &cobra.Command{
			Use: "start",
		},
	}

	cmd.RunE = func(cc *cobra.Command, args []string) error {
		ctx := cc.Context()
		if cmd.address != "" {
			return runSocketServer(ctx, cmd.address)
		}
		return runStdioServer(ctx)
	}

	cmd.Flags().StringVar(&cmd.address, "address", "", "Address (hostname:port) to listen on")

	return &cmd
}

func runStdioServer(ctx context.Context) error {
	logger := protocol.LoggerFromContext(ctx)
	logger.Debug("running in stdio mode")
	stdio := struct {
		io.ReadCloser
		io.Writer
	}{
		os.Stdin,
		os.Stdout,
	}
	conn := launchServer(ctx, stdio)
	select {
	case <-ctx.Done():
	case <-conn.Done():
		if ctx.Err() == nil {
			// only propagate connection error if context is still valid
			return conn.Err()
		}
	}
	return nil
}

func runSocketServer(ctx context.Context, addr string) error {
	var lc net.ListenConfig
	listener, err := lc.Listen(ctx, "tcp4", addr)
	if err != nil {
		return err
	}
	defer func() {
		_ = listener.Close()
	}()

	logger := protocol.LoggerFromContext(ctx).
		With(zap.String("local_addr", listener.Addr().String()))
	ctx = protocol.WithLogger(ctx, logger)
	logger.Debug("running in socket mode")

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			logger.Warn("failed to accept connection", zap.Error(err))
		}
		logger.Debug("accepted connection",
			zap.String("remote_addr", conn.RemoteAddr().String()))
		launchServer(ctx, conn)
	}
}

func launchServer(ctx context.Context, conn io.ReadWriteCloser) jsonrpc2.Conn {
	stream := jsonrpc2.NewStream(conn)
	jsonConn := jsonrpc2.NewConn(stream)

	logger := protocol.LoggerFromContext(ctx)
	notifier := protocol.ClientDispatcher(jsonConn, logger.Named("notifier"))

	docManager := document.NewDocumentManager()
	s := server.NewServer(docManager, notifier)
	h := s.Handler(server.StandardMiddleware...)

	jsonConn.Go(ctx, h)
	return jsonConn
}
