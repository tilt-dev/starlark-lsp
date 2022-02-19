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

func TestLoadBuiltinsFromFile(t *testing.T) {
	fixture := newFixture(t)
	tests := []builtinTest{
		{code: "def foo():\n    pass\n", ttype: testTypeFunctions, expectedSymbols: []string{"foo"}},
		{code: "def foo():\n  def bar():\n    pass\n  pass\n", ttype: testTypeFunctions, expectedSymbols: []string{"foo"}},
		{code: "foo = 1", ttype: testTypeSymbols, expectedSymbols: []string{"foo"}},
	}
	for i, test := range tests {
		test.Run(fixture, strings.Join(test.expectedSymbols, "-")+"-"+strconv.Itoa(i))
	}
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

func (b builtinTest) Run(f *fixture, name string) {
	f.t.Run(name, func(t *testing.T) {
		name := fmt.Sprintf("test%x.py", sha1.Sum([]byte(b.code)))
		path := f.File(name, b.code)
		builtins, err := LoadBuiltinsFromFile(f.ctx, path)
		require.NoError(t, err)
		switch b.ttype {
		case testTypeFunctions:
			assert.Equal(t, b.expectedSymbols, builtins.FunctionNames())
		case testTypeSymbols:
			assert.Equal(t, b.expectedSymbols, builtins.SymbolNames())
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
