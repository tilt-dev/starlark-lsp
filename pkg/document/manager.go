package document

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"

	"github.com/tilt-dev/starlark-lsp/pkg/query"
)

type ManagerOpt func(manager *Manager)
type ReadDocumentFunc func(uri.URI) ([]byte, error)

// Manager provides simplified file read/write operations for the LSP server.
type Manager struct {
	mu          sync.Mutex
	root        uri.URI
	docs        map[uri.URI]Document
	newDocFunc  NewDocumentFunc
	readDocFunc ReadDocumentFunc
	// map of documents created during parsing; load statements in a file will
	// trigger additional reads/parses and could create multiple documents.
	parseState map[uri.URI]Document
}

func NewDocumentManager(opts ...ManagerOpt) *Manager {
	m := Manager{
		docs:        make(map[uri.URI]Document),
		newDocFunc:  NewDocumentWithSymbols,
		readDocFunc: ReadDocument,
	}

	for _, opt := range opts {
		opt(&m)
	}

	return &m
}

func WithNewDocumentFunc(newDocFunc NewDocumentFunc) ManagerOpt {
	return func(manager *Manager) {
		manager.newDocFunc = newDocFunc
	}
}

func WithReadDocumentFunc(readDocFunc ReadDocumentFunc) ManagerOpt {
	return func(manager *Manager) {
		manager.readDocFunc = readDocFunc
	}
}

func ReadDocument(u uri.URI) (contents []byte, err error) {
	fn, err := filename(u)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(fn)
}

func filename(u uri.URI) (fn string, err error) {
	defer func() {
		// recover from non-file URI in uri.Filename()
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	return u.Filename(), err
}

func canonicalURI(u uri.URI, base uri.URI) uri.URI {
	fn, err := filename(u)
	if err != nil {
		return u
	}
	if !filepath.IsAbs(fn) && base != "" {
		basepath, err := filename(base)
		if err != nil {
			return u
		}
		fn = filepath.Join(basepath, fn)
	}
	fn, err = filepath.EvalSymlinks(fn)
	if err != nil {
		return u
	}
	return uri.File(fn)
}

func (m *Manager) Initialize(params *protocol.InitializeParams) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(params.WorkspaceFolders) > 0 {
		m.root = uri.URI(params.WorkspaceFolders[0].URI)
	} else {
		dir, err := os.Getwd()
		if err == nil {
			m.root = uri.File(dir)
		}
	}
}

// Read returns the contents of the file for the given URI.
//
// If no file exists at the path or the URI is of an invalid type, an error is
// returned.
func (m *Manager) Read(ctx context.Context, u uri.URI) (doc Document, err error) {
	m.mu.Lock()
	defer func() {
		if err == nil {
			// Always return a copy of the document
			doc = doc.Copy()
		}
		m.mu.Unlock()
	}()
	u = canonicalURI(u, m.root)

	// TODO(siegs): check staleness for files read from disk?
	var found bool
	if doc, found = m.docs[u]; !found {
		m.parseSetup()
		doc, err = m.readAndParse(ctx, u)
		m.parseCleanup(err)
	}

	if os.IsNotExist(err) {
		err = os.ErrNotExist
	}

	return doc, err
}

// Write creates or replaces the contents of the file for the given URI.
func (m *Manager) Write(ctx context.Context, u uri.URI, input []byte) (diags []protocol.Diagnostic, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u = canonicalURI(u, m.root)
	m.removeAndCleanup(u)
	m.parseSetup()
	doc, err := m.parse(ctx, u, input)
	m.parseCleanup(err)
	if err != nil {
		return nil, fmt.Errorf("could not parse file %q: %v", u, err)
	}
	return doc.Diagnostics(), err
}

func (m *Manager) Remove(u uri.URI) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u = canonicalURI(u, m.root)
	m.removeAndCleanup(u)
}

func (m *Manager) Keys() []uri.URI {
	m.mu.Lock()
	defer m.mu.Unlock()
	keys := make([]uri.URI, 0, len(m.docs))
	for k := range m.docs {
		keys = append(keys, k)
	}
	return keys
}

func (m *Manager) parseSetup() {
	m.parseState = make(map[uri.URI]Document)
}

func (m *Manager) parseCleanup(err error) {
	if err == nil {
		for u, d := range m.parseState {
			m.docs[u] = d
		}
	}
	m.parseState = nil
}

func (m *Manager) readAndParse(ctx context.Context, u uri.URI) (doc Document, err error) {
	var contents []byte
	u = canonicalURI(u, m.root)
	contents, err = m.readDocFunc(u)
	if err != nil {
		return nil, err
	}
	return m.parse(ctx, u, contents)
}

func (m *Manager) parse(ctx context.Context, uri uri.URI, input []byte) (doc Document, err error) {
	if _, found := m.parseState[uri]; found {
		return nil, fmt.Errorf("circular load: %v", uri)
	}
	tree, err := query.Parse(ctx, input)
	if err == nil {
		doc = m.newDocFunc(uri, input, tree)
		m.parseState[uri] = doc
		if docx, ok := doc.(*document); ok {
			docx.processLoads(ctx, m)
		}
	}
	return doc, err
}

// removeAndCleanup removes a Document and frees associated resources.
func (m *Manager) removeAndCleanup(uri uri.URI) {
	if existing, ok := m.docs[uri]; ok {
		existing.Close()
	}
	delete(m.docs, uri)
}
