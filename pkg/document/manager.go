package document

import (
	"io"

	"github.com/spf13/afero"
	"go.lsp.dev/uri"
)

// Manager provides simplified file read/write operations for the LSP server.
type Manager struct {
	fs afero.Fs
}

func NewDocumentManager() Manager {
	return Manager{
		fs: afero.NewMemMapFs(),
	}
}

// Read returns the contents of the file for the given URI.
//
// If no file exists at the path or the URI is of an invalid type, an error is
// returned.
func (m *Manager) Read(uri uri.URI) ([]byte, error) {
	filename, err := uriToFilename(uri)
	if err != nil {
		return nil, err
	}
	return afero.ReadFile(m.fs, filename)
}

// Write creates or replaces the contents of the file for the given URI.
//
// If the URI is of an invalid type, an error is returned.
func (m *Manager) Write(uri uri.URI, reader io.Reader) error {
	filename, err := uriToFilename(uri)
	if err != nil {
		return err
	}
	return afero.WriteReader(m.fs, filename, reader)
}
