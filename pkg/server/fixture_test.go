package server_test

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
	"github.com/tilt-dev/starlark-lsp/pkg/server"
)

type fixture struct {
	t            testing.TB
	ctx          context.Context
	client       protocol.Client
	editorConn   jsonrpc2.Conn
	editorEvents chan jsonrpc2.Request
}

func newFixture(t testing.TB) *fixture {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	logger := zaptest.NewLogger(t)
	ctx = protocol.WithLogger(ctx, logger)

	editorConn, serverConn := net.Pipe()

	serverStream := jsonrpc2.NewStream(serverConn)
	serverJsonConn := jsonrpc2.NewConn(serverStream)
	// the client is the _server_ client used to broadcast events to the editor,
	// so it needs to write to the server connection, not the editor connection
	client := protocol.ClientDispatcher(serverJsonConn, logger.Named("client"))

	docManager := document.NewDocumentManager()
	s := server.NewServer(docManager, client)
	h := s.Handler(server.StandardMiddleware...)
	serverJsonConn.Go(protocol.WithLogger(ctx, logger.Named("server")), h)

	editorStream := jsonrpc2.NewStream(editorConn)
	editorJsonConn := jsonrpc2.NewConn(editorStream)
	editorChan := make(chan jsonrpc2.Request)
	editorJsonConn.Go(protocol.WithLogger(ctx, logger.Named("editor")),
		func(ctx context.Context, _ jsonrpc2.Replier, req jsonrpc2.Request) error {
			protocol.LoggerFromContext(ctx).Debug("received message",
				zap.String("method", req.Method()),
				zap.Int("len", len(req.Params())))
			editorChan <- req
			return nil
		})

	t.Cleanup(func() {
		// this will close down the chain for us
		// jsonrpc2.Conn -> jsonrpc2.Stream -> net.Conn
		_ = editorJsonConn.Close()
		_ = serverJsonConn.Close()

		close(editorChan)

		cancel()

		_ = logger.Sync()
	})

	return &fixture{
		t:            t,
		ctx:          ctx,
		editorConn:   editorJsonConn,
		editorEvents: editorChan,
	}
}

func (f *fixture) mustEditorCall(method string, params interface{}, result interface{}) jsonrpc2.ID {
	f.t.Helper()
	id, err := f.editorConn.Call(f.ctx, method, params, result)
	require.NoErrorf(f.t, err, "RPC call %q returned an error", method)
	return id
}

func (f *fixture) nextEditorEvent(method string, params interface{}) {
	f.t.Helper()

	select {
	case <-f.ctx.Done():
		return
	case <-time.After(time.Second):
		require.Failf(f.t, "Timed out waiting for %s event", method)
	case event := <-f.editorEvents:
		require.NotNil(f.t, event, "Received nil event")
		require.Equal(f.t, method, event.Method(), "Event was for unexpected method")
		require.NoErrorf(f.t, json.Unmarshal(event.Params(), params),
			"Could not unmarshal params for %s event", method)
		return
	}
}
