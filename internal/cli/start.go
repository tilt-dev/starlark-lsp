package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/spf13/cobra"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"

	"github.com/tilt-dev/starlark-lsp/pkg/analysis"
	"github.com/tilt-dev/starlark-lsp/pkg/document"
	"github.com/tilt-dev/starlark-lsp/pkg/server"
)

type startCmd struct {
	*cobra.Command
	address         string
	builtinDefPaths []string
}

func newStartCmd() *startCmd {
	cmd := startCmd{
		Command: &cobra.Command{
			Use:   "start",
			Short: "Start the Starlark LSP server",
			Long: `Start the Starlark LSP server.

By default, the server will run in stdio mode: requests should be written to
stdin and responses will be written to stdout. (All logging is _always_ done
to stderr.)

For socket mode, pass the --address option.
`,
			Example: `
# Launch in stdio mode with extra logging
starlark-lsp start --verbose

# Listen on all interfaces on port 8765
starlark-lsp start --address=":8765"

# Provide type-stub style files to parse and treat as additional language
# built-ins
starlark-lsp start --builtin-paths "foo.py" --builtin-paths "/tmp/bar.py"
`,
		},
	}

	cmd.Command.RunE = func(cc *cobra.Command, args []string) error {
		var err error
		ctx := cc.Context()
		analyzer, err := createAnalyzer(ctx, cmd.builtinDefPaths)
		if err != nil {
			return fmt.Errorf("failed to create analyzer: %v", err)
		}
		if cmd.address != "" {
			err = runSocketServer(ctx, cmd.address, analyzer)
		} else {
			err = runStdioServer(ctx, analyzer)
		}
		if err == context.Canceled {
			err = nil
		}
		return err
	}

	cmd.Flags().StringVar(&cmd.address, "address", "",
		"Address (hostname:port) to listen on")
	cmd.Flags().StringArrayVar(&cmd.builtinDefPaths, "builtin-paths", nil,
		"Paths to files to parse and treat as additional language builtins")

	return &cmd
}

func runStdioServer(ctx context.Context, analyzer *analysis.Analyzer) error {
	ctx, cancel := context.WithCancel(ctx)
	logger := protocol.LoggerFromContext(ctx)
	logger.Debug("running in stdio mode")
	stdio := struct {
		io.ReadCloser
		io.Writer
	}{
		os.Stdin,
		os.Stdout,
	}

	conn := launchHandler(ctx, cancel, stdio, analyzer)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-conn.Done():
		if ctx.Err() == nil {
			// only propagate connection error if context is still valid
			return conn.Err()
		}
	}
	return nil
}

func runSocketServer(ctx context.Context, addr string, analyzer *analysis.Analyzer) error {
	ctx, cancel := context.WithCancel(ctx)
	var lc net.ListenConfig
	listener, err := lc.Listen(ctx, "tcp4", addr)
	if err != nil {
		cancel()
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
				cancel()
				return nil
			}
			logger.Warn("failed to accept connection", zap.Error(err))
		}
		logger.Debug("accepted connection",
			zap.String("remote_addr", conn.RemoteAddr().String()))
		jsonConn := launchHandler(ctx, cancel, conn, analyzer)

		select {
		case <-ctx.Done():
			_ = jsonConn.Close()
			return ctx.Err()
		case <-jsonConn.Done():
			if ctx.Err() == nil {
				if errors.Unwrap(jsonConn.Err()) != io.EOF {
					// only propagate connection error if context is still valid
					return jsonConn.Err()
				}
			}
		}
	}
}

func initializeConn(conn io.ReadWriteCloser, logger *zap.Logger) (jsonrpc2.Conn, protocol.Client) {
	stream := jsonrpc2.NewStream(conn)
	jsonConn := jsonrpc2.NewConn(stream)
	notifier := protocol.ClientDispatcher(jsonConn, logger.Named("notify"))

	return jsonConn, notifier
}

func createHandler(cancel context.CancelFunc, notifier protocol.Client, analyzer *analysis.Analyzer) jsonrpc2.Handler {
	docManager := document.NewDocumentManager()
	s := server.NewServer(cancel, notifier, docManager, analyzer)
	h := s.Handler(server.StandardMiddleware...)
	return h
}

func launchHandler(ctx context.Context, cancel context.CancelFunc, conn io.ReadWriteCloser, analyzer *analysis.Analyzer) jsonrpc2.Conn {
	logger := protocol.LoggerFromContext(ctx)
	jsonConn, notifier := initializeConn(conn, logger)
	h := createHandler(cancel, notifier, analyzer)
	jsonConn.Go(ctx, h)
	return jsonConn
}

func createAnalyzer(ctx context.Context, builtinDefPaths []string) (*analysis.Analyzer, error) {
	var opts []analysis.AnalyzerOption

	builtins, err := LoadBuiltins(ctx, builtinDefPaths...)
	if err != nil {
		return nil, err
	}

	logger := protocol.LoggerFromContext(ctx)
	if len(builtins.Functions) != 0 {
		logger.Debug("registered built-in functions",
			zap.Int("count", len(builtins.Functions)))
		opts = append(opts, analysis.WithBuiltinFunctions(builtins.Functions))
	}

	if len(builtins.Symbols) != 0 {
		logger.Debug("registered built-in symbols",
			zap.Int("count", len(builtins.Symbols)))
		opts = append(opts, analysis.WithBuiltinSymbols(builtins.Symbols))
	}

	return analysis.NewAnalyzer(opts...), nil
}
