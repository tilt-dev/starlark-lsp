package analysis

import (
	"context"
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
	"go.uber.org/zap/zaptest"
)

const envGetcwd = "environ = {}\ndef getcwd():\n  pass\n"

func TestLoadBuiltinsFromFile(t *testing.T) {
	fixture := newFixture(t)
	tests := []builtinTest{
		{code: "def foo():\n    pass\ndef bar(a, **b):\n  pass\n", ttype: testTypeFunctions, expectedSymbols: []string{"foo", "bar"}},
		{code: "def foo():\n    pass\ndef bar(a, **b):\n  pass\n", ttype: testTypeSymbols, expectedSymbols: []string{"foo", "bar"}},
		{code: "def foo():\n  def bar():\n    pass\n  pass\n", ttype: testTypeFunctions, expectedSymbols: []string{"foo"}},
		{code: "foo = 1\n\ndef bar():\n  pass\n", ttype: testTypeSymbols, expectedSymbols: []string{"foo", "bar"}},
	}
	for i, test := range tests {
		test.Run(fixture, strings.Join(test.expectedSymbols, "-")+"-"+strconv.Itoa(i))
	}
}

func TestLoadBuiltinModule(t *testing.T) {
	fixture := newFixture(t)
	dir := fixture.Dir("api")
	fixture.File("api/os.py", envGetcwd)
	builtins, err := LoadBuiltinModule(fixture.ctx, dir)
	require.NoError(t, err)

	assert.Equal(t, []string{"os.getcwd"}, builtins.FunctionNames())
	assert.Equal(t, []string{"os"}, builtins.SymbolNames())
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

func TestLoadBuiltinModuleInit(t *testing.T) {
	fixture := newFixture(t)
	dir := fixture.Dir("api")
	fixture.File("api/__init__.py", envGetcwd)
	builtins, err := LoadBuiltinModule(fixture.ctx, dir)

	require.NoError(t, err)
	assert.Equal(t, []string{"getcwd"}, builtins.FunctionNames())
	assertContainsAll(t, []string{"environ", "getcwd"}, builtins.SymbolNames())
}

func TestLoadBuiltinModuleDirectory(t *testing.T) {
	fixture := newFixture(t)
	dir := fixture.Dir("api")
	fixture.Dir("api/os")
	fixture.File("api/os/__init__.py", envGetcwd)
	builtins, err := LoadBuiltinModule(fixture.ctx, dir)
	require.NoError(t, err)

	assert.Equal(t, []string{"os.getcwd"}, builtins.FunctionNames())
	assert.Equal(t, []string{"os"}, builtins.SymbolNames())
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

func TestLoadBuiltinModuleDirectoryFile(t *testing.T) {
	fixture := newFixture(t)
	dir := fixture.Dir("api")
	fixture.Dir("api/os")
	fixture.File("api/os/fns.py", envGetcwd)
	builtins, err := LoadBuiltinModule(fixture.ctx, dir)
	require.NoError(t, err)

	assert.Equal(t, []string{"os.fns.getcwd"}, builtins.FunctionNames())
	assert.Equal(t, []string{"os"}, builtins.SymbolNames())
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
	dir      string
	builtins *Builtins
}

type testType int

const (
	testTypeFunctions testType = iota
	testTypeSymbols
)

type builtinTest struct {
	code            string
	ttype           testType
	expectedSymbols []string
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

func (b builtinTest) Run(f *fixture, name string) {

	f.t.Run(name, func(t *testing.T) {
		name := fmt.Sprintf("test%x.py", sha1.Sum([]byte(b.code)))
		path := f.File(name, b.code)
		builtins, err := LoadBuiltinsFromFile(f.ctx, path)
		require.NoError(t, err)
		switch b.ttype {
		case testTypeFunctions:
			assertContainsAll(t, b.expectedSymbols, builtins.FunctionNames())
		case testTypeSymbols:
			assertContainsAll(t, b.expectedSymbols, builtins.SymbolNames())
		}
	})
}

func (f *fixture) File(name, contents string) string {
	path := filepath.Join(f.dir, name)
	_ = os.WriteFile(path, []byte(contents), 0644)
	return path
}

func (f *fixture) Dir(name string) string {
	path := filepath.Join(f.dir, name)
	_ = os.Mkdir(path, 0755)
	return path
}

func newFixture(t *testing.T) *fixture {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	logger := zaptest.NewLogger(t)
	t.Cleanup(func() {
		_ = logger.Sync()
	})
	ctx = protocol.WithLogger(ctx, logger)

	return &fixture{
		ctx: ctx,
		t:   t,
		dir: t.TempDir(),
		builtins: &Builtins{
			Functions: make(map[string]protocol.SignatureInformation),
			Symbols:   []protocol.DocumentSymbol{},
		},
	}
}
