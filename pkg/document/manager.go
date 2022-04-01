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
}

func NewDocumentManager(opts ...ManagerOpt) *Manager {
	m := Manager{
		docs:        make(map[uri.URI]Document),
		newDocFunc:  NewDocument,
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
func (m *Manager) Read(ctx context.Context, uri uri.URI) (Document, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if doc, ok := m.docs[uri]; ok {
		// TODO(siegs): check staleness for files read from disk?
		return doc.Copy(), nil
	}

	contents, err := m.readDocFunc(uri)
	if err == nil {
		var tree *sitter.Tree
		tree, err = query.Parse(ctx, contents)
		if err == nil {
			doc := m.newDocFunc(contents, tree)
			m.docs[uri] = doc
			return doc.Copy(), nil
		}
	}
	if os.IsNotExist(err) {
		err = os.ErrNotExist
	}

	return nil, err
}

// Write creates or replaces the contents of the file for the given URI.
func (m *Manager) Write(ctx context.Context, uri uri.URI, input []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.removeAndCleanup(uri)
	tree, err := query.Parse(ctx, input)
	if err != nil {
		return fmt.Errorf("could not parse file %q: %v", uri, err)
	}

	m.docs[uri] = m.newDocFunc(input, tree)
	return nil
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

// removeAndCleanup removes a Document and frees associated resources.
func (m *Manager) removeAndCleanup(uri uri.URI) {
	if existing, ok := m.docs[uri]; ok {
		existing.Close()
	}
	delete(m.docs, uri)
}
