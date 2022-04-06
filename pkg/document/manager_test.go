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

func TestReadWithLoad(t *testing.T) {
	f := newFixture(t)
	require.NoError(t, os.WriteFile("doc1", []byte(`load("doc2", "foo")`), 0644))
	require.NoError(t, os.WriteFile("doc2", []byte(`foo = True`), 0644))
	doc, err := f.m.Read(f.ctx, uri.File("doc1"))
	require.NoError(t, err)
	assert.Equal(t, 1, len(doc.Loads()))
	assert.Equal(t, 0, len(doc.Diagnostics()))
	syms := doc.Symbols()
	assert.Equal(t, 1, len(syms))
	//assert.Equal(t, "foo", syms[0].Name)
}

func TestNestedLoad(t *testing.T) {
	f := newFixture(t)
	code := `
if True:
  load("doc2", "foo")
`
	require.NoError(t, os.WriteFile("doc1", []byte(code), 0644))
	doc, err := f.m.Read(f.ctx, uri.File("doc1"))
	require.NoError(t, err)
	assert.Equal(t, 0, len(doc.Symbols()))
	assert.Equal(t, 1, len(doc.Diagnostics()))
}

func TestCircularLoad(t *testing.T) {
	f := newFixture(t)
	require.NoError(t, os.WriteFile("doc1", []byte(`load("doc2", "foo")`), 0644))
	require.NoError(t, os.WriteFile("doc2", []byte(`load("doc1", "bar")`), 0644))
	doc, err := f.m.Read(f.ctx, uri.File("doc1"))
	require.NoError(t, err)
	assert.Equal(t, 1, len(doc.Diagnostics()))
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
