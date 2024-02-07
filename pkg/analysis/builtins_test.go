package analysis

import (
	"context"
	"crypto/sha1"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/exp/maps"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.uber.org/zap/zaptest"

	"github.com/autokitteh/starlark-lsp/pkg/docstring"
	"github.com/autokitteh/starlark-lsp/pkg/document"
	"github.com/autokitteh/starlark-lsp/pkg/query"
)

const envGetcwd = "environ = {}\ndef getcwd():\n  pass\n"

//go:embed test/*.py
var testFS embed.FS

type testType int

const (
	testTypeFunctions testType = iota
	testTypeSymbols
)

func TestLoadBuiltinsFromFile(t *testing.T) {
	f := newFixture(t)
	tests := []struct {
		code     string
		ttype    testType
		expected []string
	}{
		{code: "def foo():\n    pass\ndef bar(a, **b):\n  pass\n", ttype: testTypeFunctions, expected: []string{"foo", "bar"}},
		{code: "def foo():\n    pass\ndef bar(a, **b):\n  pass\n", ttype: testTypeSymbols, expected: []string{"foo", "bar"}},
		{code: "def foo():\n  def bar():\n    pass\n  pass\n", ttype: testTypeFunctions, expected: []string{"foo"}},
		{code: "foo = 1\n\ndef bar():\n  pass\n", ttype: testTypeSymbols, expected: []string{"foo", "bar"}},
	}
	for i, test := range tests {
		t.Run(strings.Join(test.expected, "-")+"-"+strconv.Itoa(i), func(t *testing.T) {
			name := fmt.Sprintf("test%x.py", sha1.Sum([]byte(test.code)))
			path := f.File(name, test.code)
			builtins, err := LoadBuiltinsFromFile(f.ctx, path, nil)
			require.NoError(t, err)
			switch test.ttype {
			case testTypeFunctions:
				assertContainsAll(t, test.expected, maps.Keys(builtins.Functions))
			case testTypeSymbols:
				assertContainsAll(t, test.expected, symbolNames(builtins.Symbols))
			}
		})
	}
}

func TestLoadBuiltinsFromFS(t *testing.T) {
	fixture := newFixture(t)
	dir := fixture.Dir("api")
	fixture.File("api/os.py", envGetcwd)
	builtins, err := LoadBuiltinsFromFS(fixture.ctx, os.DirFS(dir))
	require.NoError(t, err)

	assert.Equal(t, []string{"os.getcwd"}, maps.Keys(builtins.Functions))
	assert.Equal(t, []string{"os"}, symbolNames(builtins.Symbols))
	osSym := builtins.Symbols[0]
	assert.Equal(t, protocol.SymbolKindVariable, osSym.Kind)
	assert.Equal(t, 2, len(osSym.Children))
	environSym := osSym.Children[0]
	assert.Equal(t, "environ", environSym.Name)
	assert.Equal(t, protocol.SymbolKindField, environSym.Kind)
	getcwdSym := osSym.Children[1]
	assert.Equal(t, "getcwd", getcwdSym.Name)
	assert.Equal(t, protocol.SymbolKindMethod, getcwdSym.Kind)
}

func TestLoadBuiltinsFromFSEmbed(t *testing.T) {
	fixture := newFixture(t)
	testDir, err := fs.Sub(testFS, "test")
	require.NoError(t, err)
	builtins, err := LoadBuiltinsFromFS(fixture.ctx, testDir)
	require.NoError(t, err)

	assert.Equal(t, []string{"os.getcwd"}, maps.Keys(builtins.Functions))
	assert.Equal(t, []string{"os"}, symbolNames(builtins.Symbols))
	osSym := builtins.Symbols[0]
	assert.Equal(t, protocol.SymbolKindVariable, osSym.Kind)
	assert.Equal(t, 2, len(osSym.Children))
	environSym := osSym.Children[0]
	assert.Equal(t, "environ", environSym.Name)
	assert.Equal(t, protocol.SymbolKindField, environSym.Kind)
	getcwdSym := osSym.Children[1]
	assert.Equal(t, "getcwd", getcwdSym.Name)
	assert.Equal(t, protocol.SymbolKindMethod, getcwdSym.Kind)

	assert.Equal(t, 1, len(builtins.Methods))
	assert.Equal(t, []string{"curl"}, maps.Keys(builtins.Methods))
	assert.Equal(t, []string{"url", "curl"}, symbolNames(builtins.Members))

	assert.Equal(t, 1, len(builtins.Types))
	assert.Equal(t, []string{"Link"}, maps.Keys(builtins.Types))
}

func TestLoadBuiltinsFromFSInit(t *testing.T) {
	fixture := newFixture(t)
	dir := fixture.Dir("api")
	fixture.File("api/__init__.py", envGetcwd)
	builtins, err := LoadBuiltinsFromFS(fixture.ctx, os.DirFS(dir))

	require.NoError(t, err)
	assert.Equal(t, []string{"getcwd"}, maps.Keys(builtins.Functions))
	assertContainsAll(t, []string{"environ", "getcwd"}, symbolNames(builtins.Symbols))
}

func TestLoadBuiltinsFromFSDirectory(t *testing.T) {
	fixture := newFixture(t)
	dir := fixture.Dir("api")
	fixture.Dir("api/os")
	fixture.File("api/os/__init__.py", envGetcwd)
	builtins, err := LoadBuiltinsFromFS(fixture.ctx, os.DirFS(dir))
	require.NoError(t, err)

	assert.Equal(t, []string{"os.getcwd"}, maps.Keys(builtins.Functions))
	assert.Equal(t, []string{"os"}, symbolNames(builtins.Symbols))
	osSym := builtins.Symbols[0]
	assert.Equal(t, protocol.SymbolKindVariable, osSym.Kind)
	assert.Equal(t, 2, len(osSym.Children))
	environSym := osSym.Children[0]
	assert.Equal(t, "environ", environSym.Name)
	assert.Equal(t, protocol.SymbolKindField, environSym.Kind)
	getcwdSym := osSym.Children[1]
	assert.Equal(t, "getcwd", getcwdSym.Name)
	assert.Equal(t, protocol.SymbolKindMethod, getcwdSym.Kind)
}

func TestLoadBuiltinsFromFSEmptyDirectories(t *testing.T) {
	fixture := newFixture(t)
	dir := fixture.Dir("api")
	fixture.Dir("api/os")
	builtins, err := LoadBuiltinsFromFS(fixture.ctx, os.DirFS(dir))
	require.NoError(t, err)
	assert.True(t, builtins.IsEmpty())
}

func TestLoadBuiltinsFromFSMultipleModules(t *testing.T) {
	fixture := newFixture(t)
	dir := fixture.Dir("api")
	fixture.Dir("api/os")
	fixture.File("api/os.py", `name: str = ""`)
	fixture.File("api/os/__init__.py", envGetcwd)
	builtins, err := LoadBuiltinsFromFS(fixture.ctx, os.DirFS(dir))
	require.NoError(t, err)

	assert.Equal(t, []string{"os"}, symbolNames(builtins.Symbols))
	osSym := builtins.Symbols[0]
	assert.Equal(t, protocol.SymbolKindVariable, osSym.Kind)
	assert.Equal(t, 3, len(osSym.Children))
	environSym := osSym.Children[0]
	assert.Equal(t, "environ", environSym.Name)
	assert.Equal(t, protocol.SymbolKindField, environSym.Kind)
	getcwdSym := osSym.Children[1]
	assert.Equal(t, "getcwd", getcwdSym.Name)
	assert.Equal(t, protocol.SymbolKindMethod, getcwdSym.Kind)
	nameSym := osSym.Children[2]
	assert.Equal(t, "name", nameSym.Name)
	assert.Equal(t, protocol.SymbolKindField, nameSym.Kind)
}

func TestLoadBuiltinsFromFSDirectoryFile(t *testing.T) {
	fixture := newFixture(t)
	dir := fixture.Dir("api")
	fixture.Dir("api/os")
	fixture.File("api/os/fns.py", envGetcwd)
	builtins, err := LoadBuiltinsFromFS(fixture.ctx, os.DirFS(dir))
	require.NoError(t, err)

	assert.Equal(t, []string{"os.fns.getcwd"}, maps.Keys(builtins.Functions))
	assert.Equal(t, []string{"os"}, symbolNames(builtins.Symbols))
	osSym := builtins.Symbols[0]
	assert.Equal(t, protocol.SymbolKindVariable, osSym.Kind)
	assert.Equal(t, 1, len(osSym.Children))
	fnsSym := osSym.Children[0]
	assert.Equal(t, protocol.SymbolKindField, fnsSym.Kind)
	assert.Equal(t, 2, len(fnsSym.Children))
	environSym := fnsSym.Children[0]
	assert.Equal(t, "environ", environSym.Name)
	assert.Equal(t, protocol.SymbolKindField, environSym.Kind)
	getcwdSym := fnsSym.Children[1]
	assert.Equal(t, "getcwd", getcwdSym.Name)
	assert.Equal(t, protocol.SymbolKindMethod, getcwdSym.Kind)
}

type fixture struct {
	t        *testing.T
	ctx      context.Context
	a        *Analyzer
	dir      string
	builtins *Builtins
	docs     *document.Manager
}

func assertContainsAll(t *testing.T, expected []string, actual []string) {
	for _, exp := range expected {
		found := false
		for _, act := range actual {
			if exp == act {
				found = true
				break
			}
		}
		if !found {
			assert.Fail(t, fmt.Sprintf("\"%s\" not found in %v", exp, actual))
		}
	}
}

func (f *fixture) File(name, contents string) string {
	path := filepath.Join(f.dir, name)
	_ = os.WriteFile(path, []byte(contents), 0o644)
	return path
}

func (f *fixture) Dir(name string) string {
	path := filepath.Join(f.dir, name)
	_ = os.Mkdir(path, 0o755)
	return path
}

func (f *fixture) Symbols(names ...string) {
	for _, name := range names {
		f.builtins.Symbols = append(f.builtins.Symbols, f.Symbol(name))
	}
}

func (f *fixture) AddFunction(name string, content string) {
	f.builtins.Functions[name] = query.Signature{
		Name: name,
		Docs: docstring.Parsed{Description: content},
	}
	f.AddSymbol(name, content)
}

func (f *fixture) AddSymbol(name string, content string) {
	ids := strings.Split(name, ".")
	var cur query.Symbol
	for i := len(ids) - 1; i >= 0; i-- {
		s := query.Symbol{Name: ids[i]}
		if i == len(ids)-1 {
			s.Detail = content
		} else {
			s.Children = []query.Symbol{cur}
		}
		cur = s
	}
	f.builtins.Symbols = append(f.builtins.Symbols, cur)
}

func (f *fixture) Symbol(name string) query.Symbol {
	return query.Symbol{
		Name: name,
		Kind: protocol.SymbolKindVariable,
	}
}

func (f *fixture) Document(name string, content string) document.Document {
	u := uri.File(name)
	_, err := f.docs.Write(f.ctx, u, []byte(content))
	require.NoError(f.t, err)
	doc, err := f.docs.Read(f.ctx, u)
	require.NoError(f.t, err)
	f.t.Cleanup(func() { doc.Close() })
	return doc
}

func (f *fixture) MainDoc(content string) document.Document {
	return f.Document("Tiltfile.test", content)
}

func (f *fixture) ParseBuiltins(content string) {
	builtins, err := LoadBuiltinsFromSource(f.ctx, []byte(content), "__test__")
	require.NoError(f.t, err)
	f.a.builtins.Update(builtins)
	f.builtins = f.a.builtins

	// f.a.builtins = builtins
	// f.builtins = builtins
}

func newFixture(t *testing.T) *fixture {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	logger := zaptest.NewLogger(t)
	t.Cleanup(func() {
		_ = logger.Sync()
	})
	ctx = protocol.WithLogger(ctx, logger)

	builtins := NewBuiltins()
	a, _ := NewAnalyzer(ctx)
	a.builtins = builtins

	f := &fixture{
		ctx:      ctx,
		t:        t,
		dir:      t.TempDir(),
		builtins: builtins,
		a:        a,
	}

	f.docs = document.NewDocumentManager()

	return f
}
