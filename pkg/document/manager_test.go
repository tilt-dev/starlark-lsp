package document

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/uri"
)

func TestManagerRead(t *testing.T) {
	f := newFixture(t)
	_, err := f.m.Read(f.ctx, uri.File("doesnotexist"))
	require.ErrorIs(t, os.ErrNotExist, err)
	_, err = f.m.Read(f.ctx, uri.URI("https://example.com/"))
	require.EqualError(t, err, "only file URIs are supported, got https")
	_ = os.WriteFile("doc", []byte(""), 0644)
	doc, err := f.m.Read(f.ctx, uri.File("doc"))
	require.NoError(t, err)
	assert.Equal(t, "", doc.Content(doc.Tree().RootNode()))
}

type fixture struct {
	ctx context.Context
	m   *Manager
}

func newFixture(t *testing.T) *fixture {
	wd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(wd) })
	dir := t.TempDir()
	require.NoError(t, os.Chdir(dir))
	return &fixture{
		ctx: context.Background(),
		m:   NewDocumentManager(),
	}
}
