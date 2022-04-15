package server_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"runtime/debug"
	"strings"
	"sync"
	"testing"
	"time"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/assert"
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

	docManager := newDocumentManager(t)
	analyzer, _ := analysis.NewAnalyzer(ctx)
	s := server.NewServer(cancel, notifier, docManager, analyzer)

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

func (f *fixture) mustWriteDocument(path string, source string) {
	f.t.Helper()
	contents := []byte(source)
	_, err := f.docManager.Write(f.ctx, uri.File(path), contents)
	require.NoErrorf(
		f.t,
		err,
		"Failed to parse document %q",
		path,
	)
}

func (f *fixture) requireDocContents(path string, input string) {
	f.t.Helper()
	doc, err := f.docManager.Read(f.ctx, uri.File(path))
	require.NoErrorf(f.t, err, "Failed to read document %q", path)
	defer doc.Close()
	require.NotNil(f.t, doc.Tree(), "Document tree was nil")
	require.NotNil(f.t, doc.Tree().RootNode(),
		"Document root node was nil")
	require.Equal(f.t, input, doc.Content(doc.Tree().RootNode()),
		"Document contents did not match")
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

func newDocumentManager(t testing.TB) *document.Manager {
	t.Helper()

	var mu sync.Mutex
	openDocs := make(map[*testDocument][]byte)
	copyFunc := func(testDoc *testDocument) {
		t.Helper()
		mu.Lock()
		defer mu.Unlock()
		if existing, ok := openDocs[testDoc]; ok {
			require.Fail(t, "Copied document already exists in open document list, was created at\n%s\n", existing)
		}
		openDocs[testDoc] = debug.Stack()
	}
	closeFunc := func(testDoc *testDocument) {
		t.Helper()
		mu.Lock()
		defer mu.Unlock()
		_, ok := openDocs[testDoc]
		if !ok {
			require.Fail(t, "Close for document that was not tracked")
		}
		delete(openDocs, testDoc)
	}

	newDocFunc := func(u uri.URI, input []byte, tree *sitter.Tree) document.Document {
		testDoc := &testDocument{
			doc:     document.NewDocument(u, input, tree),
			onCopy:  copyFunc,
			onClose: closeFunc,
		}
		mu.Lock()
		defer mu.Unlock()
		openDocs[testDoc] = debug.Stack()
		return testDoc
	}

	mgr := document.NewDocumentManager(
		document.WithNewDocumentFunc(newDocFunc),
	)

	t.Cleanup(func() {
		t.Helper()

		for _, key := range mgr.Keys() {
			mgr.Remove(key)
		}

		mu.Lock()
		defer mu.Unlock()

		if len(openDocs) == 0 {
			return
		}

		var stacks []string
		for _, stack := range openDocs {
			lines := strings.Split(string(stack), "\n")
			if len(lines) > 7 {
				// stack trace should start with some things we don't care about:
				// 	1. goroutine line
				// 	2. 2x lines for `debug.Stack()` (call site + file info)
				// 	3. 2x lines for our anonymous new/copy func (call site + file info)
				// 	4. 2x lines for Document::Copy (call site + file info)
				lines = lines[7:]
			}
			stacks = append(stacks, strings.TrimSpace(strings.Join(lines, "\n")))
		}

		const divider = "\n------------------------------\n"
		assert.Failf(t,
			fmt.Sprintf("%d document(s) were not closed", len(stacks)),
			"Stack traces for document creation:%s%s", divider, strings.Join(stacks, divider))
	})

	return mgr
}

type testDocument struct {
	doc     document.Document
	onCopy  func(*testDocument)
	onClose func(*testDocument)
}

var _ document.Document = &testDocument{}

func (t *testDocument) Input() []byte {
	return t.doc.Input()
}

func (t *testDocument) Content(n *sitter.Node) string {
	return t.doc.Content(n)
}

func (t *testDocument) ContentRange(r sitter.Range) string {
	return t.doc.ContentRange(r)
}

func (t *testDocument) Tree() *sitter.Tree {
	return t.doc.Tree()
}

func (t *testDocument) FunctionSignatures() map[string]query.Signature {
	return t.doc.FunctionSignatures()
}

func (t *testDocument) Functions() map[string]protocol.SignatureInformation {
	return t.doc.Functions()
}

func (t *testDocument) Symbols() []protocol.DocumentSymbol {
	return t.doc.Symbols()
}

func (t *testDocument) Diagnostics() []protocol.Diagnostic {
	return t.doc.Diagnostics()
}

func (t *testDocument) Loads() []document.LoadStatement {
	return t.doc.Loads()
}

func (t *testDocument) Copy() document.Document {
	copiedDoc := &testDocument{
		doc:     t.doc.Copy(),
		onCopy:  t.onCopy,
		onClose: t.onClose,
	}
	if copiedDoc.onCopy != nil {
		copiedDoc.onCopy(copiedDoc)
	}
	return copiedDoc
}

func (t *testDocument) Close() {
	if t.onClose != nil {
		t.onClose(t)
	}
	t.doc.Close()
}
