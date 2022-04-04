package document

import (
	"context"
	"fmt"
	"os"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/uri"

	"github.com/tilt-dev/starlark-lsp/pkg/query"
)

type ManagerOpt func(manager *Manager)
type ReadDocumentFunc func(uri.URI) ([]byte, error)

// Manager provides simplified file read/write operations for the LSP server.
type Manager struct {
	mu          sync.Mutex
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
	defer func() {
		// recover from non-file URI in uri.Filename()
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	return os.ReadFile(u.Filename())
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
func (m *Manager) Write(ctx context.Context, uri uri.URI, input []byte) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.removeAndCleanup(uri)
	m.parseSetup()
	_, err = m.parse(ctx, uri, input)
	m.parseCleanup(err)
	if err != nil {
		return fmt.Errorf("could not parse file %q: %v", uri, err)
	}
	return err
}

func (m *Manager) Remove(uri uri.URI) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.removeAndCleanup(uri)
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
	contents, err = m.readDocFunc(u)
	if err != nil {
		return nil, err
	}
	return m.parse(ctx, u, contents)
}

func (m *Manager) parse(ctx context.Context, uri uri.URI, input []byte) (doc Document, err error) {
	if err == nil {
		var tree *sitter.Tree
		tree, err = query.Parse(ctx, input)
		if err == nil {
			doc = m.newDocFunc(input, tree)
			m.parseState[uri] = doc
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
