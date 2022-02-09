package document

import (
	"os"
	"sync"

	"go.lsp.dev/uri"
)

// Manager provides simplified file read/write operations for the LSP server.
type Manager struct {
	mu   sync.Mutex
	docs map[uri.URI]Document
}

func NewDocumentManager() *Manager {
	return &Manager{
		docs: make(map[uri.URI]Document),
	}
}

// Read returns the contents of the file for the given URI.
//
// If no file exists at the path or the URI is of an invalid type, an error is
// returned.
func (m *Manager) Read(uri uri.URI) (Document, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if doc, ok := m.docs[uri]; ok {
		return doc.shallowClone(), nil
	}
	return Document{}, os.ErrNotExist
}

// Write creates or replaces the contents of the file for the given URI.
func (m *Manager) Write(uri uri.URI, doc Document) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.removeAndCleanup(uri)
	m.docs[uri] = doc
}

func (m *Manager) Remove(uri uri.URI) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.removeAndCleanup(uri)
}

// removeAndCleanup removes a Document and frees associated resources.
func (m *Manager) removeAndCleanup(uri uri.URI) {
	if existing, ok := m.docs[uri]; ok {
		existing.Close()
	}
	delete(m.docs, uri)
}
