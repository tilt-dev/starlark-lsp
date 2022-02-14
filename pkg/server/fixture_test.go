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
	"go.lsp.dev/uri"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/tilt-dev/starlark-lsp/pkg/analysis"
	"github.com/tilt-dev/starlark-lsp/pkg/document"
	"github.com/tilt-dev/starlark-lsp/pkg/middleware"
	"github.com/tilt-dev/starlark-lsp/pkg/query"
	"github.com/tilt-dev/starlark-lsp/pkg/server"
)

type fixture struct {
	t            testing.TB
	ctx          context.Context
	docManager   *document.Manager
	editorConn   jsonrpc2.Conn
	editorEvents chan jsonrpc2.Request
}

func newFixture(t testing.TB) *fixture {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	logger := zaptest.NewLogger(t)
	t.Cleanup(func() {
		_ = logger.Sync()
	})
	ctx = protocol.WithLogger(ctx, logger)

	editorConn, serverConn := net.Pipe()

	serverStream := jsonrpc2.NewStream(serverConn)
	serverJsonConn := jsonrpc2.NewConn(serverStream)
	// the client is the _server_ client used to broadcast events to the editor,
	// so it needs to write to the server connection, not the editor connection
	notifier := protocol.ClientDispatcher(serverJsonConn, logger.Named("notify"))

	docManager := document.NewDocumentManager()
	analyzer := analysis.NewAnalyzer()
	s := server.NewServer(notifier, docManager, analyzer)

	// TODO(milas): AsyncHandler does not stop if the server is shut down which
	// 	can cause panics in tests (due to logs being emitted after tests are
	// 	done); should upstream a patch
	testMiddleware := []middleware.Middleware{
		middleware.Recover,
		middleware.Error,
		protocol.CancelHandler,
		// jsonrpc2.AsyncHandler,
		jsonrpc2.ReplyHandler,
	}

	h := s.Handler(testMiddleware...)
	serverJsonConn.Go(protocol.WithLogger(ctx, logger.Named("server")), h)

	editorStream := jsonrpc2.NewStream(editorConn)
	editorJsonConn := jsonrpc2.NewConn(editorStream)
	editorChan := make(chan jsonrpc2.Request, 20)
	editorJsonConn.Go(protocol.WithLogger(ctx, logger.Named("editor")),
		func(ctx context.Context, _ jsonrpc2.Replier, req jsonrpc2.Request) error {
			protocol.LoggerFromContext(ctx).Debug("received message",
				zap.String("method", req.Method()),
				zap.Int("len", len(req.Params())))
			select {
			case editorChan <- req:
			default:
				panic("editor channel was full")
			}
			return nil
		})

	t.Cleanup(func() {
		// this will close down the chain for us
		// jsonrpc2.Conn -> jsonrpc2.Stream -> io.ReadWriteCloser (net.Pipe)
		_ = editorJsonConn.Close()
		_ = serverJsonConn.Close()

		close(editorChan)

		cancel()

		<-editorJsonConn.Done()
		<-serverJsonConn.Done()
	})

	return &fixture{
		t:            t,
		ctx:          ctx,
		docManager:   docManager,
		editorConn:   editorJsonConn,
		editorEvents: editorChan,
	}
}

func (f *fixture) loadDocument(path string, source string) {
	f.t.Helper()
	contents := []byte(source)
	tree, err := query.Parse(f.ctx, contents)
	require.NoErrorf(f.t, err, "Failed to parse document %q", path)

	doc := document.NewDocument(contents, tree)
	f.docManager.Write(uri.File(path), doc)
}

func (f *fixture) mustEditorCall(method string, params interface{}, result interface{}) jsonrpc2.ID {
	f.t.Helper()
	id, err := f.editorConn.Call(f.ctx, method, params, result)
	require.NoErrorf(f.t, err, "RPC call %q returned an error", method)
	return id
}

// requireNextEditorEvent fails the test if the next event broadcast to the
// editor from the server is not for the given method, fails deserialization,
// or is not received within a reasonable amount of time.
//
// Tests can assert further on the unmarshalled parameters.
func (f *fixture) requireNextEditorEvent(method string, params interface{}) {
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
