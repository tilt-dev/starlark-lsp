package document

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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
	if len(syms) == 1 {
		assert.Equal(t, "foo", syms[0].Name)
	}
}

func TestReadWithUnsupportedURI(t *testing.T) {
	f := newFixture(t)
	require.NoError(t, os.WriteFile("doc", []byte(`load("ext://doc2", "foo")`), 0644))
	doc, err := f.m.Read(f.ctx, uri.File("doc"))
	require.NoError(t, err)
	assert.Equal(t, 1, len(doc.Loads()))
	diags := doc.Diagnostics()
	assert.Equal(t, 1, len(diags))
	if len(diags) == 1 {
		assert.Equal(t, "only file URIs are supported, got ext", diags[0].Message)
	}
}

func TestNestedLoad(t *testing.T) {
	cases := []struct {
		code     string
		expected string
	}{
		{code: `if True:
  %s
`, expected: "if statement"},
		{code: "x = %s", expected: "assignment"},
		{code: "x = lambda: %s", expected: "lambda"},
		{code: `def fn():
  %s`, expected: "function definition"},
	}

	for i, tt := range cases {
		t.Run(fmt.Sprintf("%s-%d", tt.expected, i), func(t *testing.T) {
			f := newFixture(t)
			require.NoError(t, os.WriteFile("doc1", []byte(fmt.Sprintf(tt.code, `load("doc2","foo")`)), 0644))
			doc, err := f.m.Read(f.ctx, uri.File("doc1"))
			require.NoError(t, err)
			diags := doc.Diagnostics()
			assert.Equal(t, 1, len(diags))
			if len(diags) == 1 {
				assert.True(t, strings.HasSuffix(doc.Diagnostics()[0].Message, tt.expected))
			}
		})
	}
}

func TestCircularLoad(t *testing.T) {
	f := newFixture(t)
	require.NoError(t, os.WriteFile("doc1", []byte(`
load("doc2", "foo")
bar = True
`), 0644))
	require.NoError(t, os.WriteFile("doc2", []byte(`
load("doc1", "bar")
foo = True
`), 0644))
	doc, err := f.m.Read(f.ctx, uri.File("doc1"))
	require.NoError(t, err)
	diags := doc.Diagnostics()
	assert.Equal(t, 1, len(diags))
	if len(diags) == 1 {
		assert.True(t, strings.Contains(diags[0].Message, "circular load"),
			"message was: %s", diags[0].Message)
	}
}

func TestNotCircularLoad(t *testing.T) {
	f := newFixture(t)
	require.NoError(t, os.WriteFile("a.tiltfile", []byte(`
load('ext.tiltfile', 'flag')
load('b.tiltfile', 'hello')
hello()
`), 0644))
	require.NoError(t, os.WriteFile("b.tiltfile", []byte(`
load('ext.tiltfile', 'flag')

def hello():
  print("in b: " + flag)
`), 0644))
	require.NoError(t, os.WriteFile("ext.tiltfile", []byte(`
flag = True
`), 0644))
	doc, err := f.m.Read(f.ctx, uri.File("a.tiltfile"))
	require.NoError(t, err)
	diags := doc.Diagnostics()
	fmt.Printf("diags: %v\n", diags)
	assert.Equal(t, 0, len(diags))
}

func TestResolveURI(t *testing.T) {
	f := newFixture(t)
	t.Run("default resolve function", func(t *testing.T) {
		file := uri.File("f.txt")
		u, err := f.m.Resolve(file)
		require.NoError(t, err)
		assert.Equal(t, file, u)

		_, err = f.m.Resolve("ext://some/file")
		require.Error(t, err)
	})

	t.Run("overridden resolve function", func(t *testing.T) {
		cwd, err := os.Getwd()
		require.NoError(t, err)
		extsPath := filepath.Join(cwd, "exts")
		WithResolveURIFunc(func(u uri.URI) (string, error) {
			parsed, err := url.Parse(string(u))
			if err != nil {
				return "", err
			}
			if parsed.Scheme == "ext" {
				return filepath.Join(extsPath, parsed.Host, parsed.Path), nil
			}
			return ResolveURI(u)
		})(f.m)

		ext := uri.URI("ext://some/ext")
		u, err := f.m.Resolve(ext)
		require.NoError(t, err)
		assert.Equal(t, uri.File(filepath.Join(extsPath, "some", "ext")), u)
	})
}

func TestReadLoadResolve(t *testing.T) {
	f := newFixture(t)
	require.NoError(t, os.Mkdir("exts", 0755))
	cwd, err := os.Getwd()
	require.NoError(t, err)
	extsPath := filepath.Join(cwd, "exts")
	resolve := func(u uri.URI) (string, error) {
		parsed, err := url.Parse(string(u))
		if err != nil {
			return "", err
		}
		if parsed.Scheme == "ext" {
			return filepath.Join(extsPath, parsed.Host, parsed.Path), nil
		}
		return ResolveURI(u)
	}
	WithResolveURIFunc(resolve)(f.m)
	WithReadDocumentFunc(func(u uri.URI) ([]byte, error) {
		file, err := resolve(u)
		if err != nil {
			return nil, err
		}
		return ReadDocument(uri.File(file))
	})(f.m)

	hello := filepath.Join(extsPath, "hello")
	require.NoError(t, os.WriteFile(hello, []byte(`hello = lambda: print('Hi')`), 0644))
	require.NoError(t, os.WriteFile("doc", []byte("load('ext://hello', 'hello')\nhello()"), 0644))

	doc, err := f.m.Read(f.ctx, uri.File("doc"))
	require.NoError(t, err)
	syms := doc.Symbols()
	assert.Equal(t, 1, len(syms))
	assert.Equal(t, "hello", syms[0].Name)
	assert.Equal(t, uri.File(hello), syms[0].Location.URI)
}

func TestURIfilename(t *testing.T) {
	var fn string
	var err error
	fn, err = filename(uri.URI("file:///mod"))
	require.NoError(t, err)
	assert.Equal(t, "/mod", fn)
	fn, err = filename(uri.URI("ext://mod"))
	require.Error(t, err)
	assert.Equal(t, "", fn)
	assert.Equal(t, "only file URIs are supported, got ext", err.Error())
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
