package analysis

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
)

const fooDoc = `
def foo(a, b, c, d):
  pass

foo%s
`

func TestSignatureHelp(t *testing.T) {
	tt := []struct {
		args   string
		active uint32
	}{
		{args: "(", active: 0},
		{args: "()", active: 0},
		{args: "(1,", active: 1},
		{args: "(1,)", active: 1},
		{args: "(1, 2", active: 1},
		{args: "(1, 2)", active: 1},
		{args: "(1, 2,", active: 2},
		{args: "(1, 2,)", active: 2},
		{args: "(b)", active: 0},
		{args: "(b=", active: 1},
		{args: "(b=)", active: 1},
		{args: "(1, d=", active: 3},
		{args: "(1, d=)", active: 3},
		{args: "(1, d=,)", active: 0},
		{args: "(1, d=True, c=)", active: 2},
	}

	for i, test := range tt {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			f := newFixture(t)
			doc := f.MainDoc(fmt.Sprintf(fooDoc, test.args))
			ch := uint32(3 + len(test.args))
			if strings.HasSuffix(test.args, ")") {
				ch -= 1
			}
			pos := protocol.Position{Line: 4, Character: ch}
			help := f.a.SignatureHelp(doc, pos)
			assert.NotNil(t, help)
			if help == nil {
				return
			}
			assert.Equal(t, 1, len(help.Signatures))
			assert.Equal(t, "(a, b, c, d)", help.Signatures[0].Label)
			assert.Equal(t, test.active, help.ActiveParameter)
		})
	}
}

func TestMethodSignatureHelp(t *testing.T) {
	f := newFixture(t)
	_ = WithStarlarkBuiltins()(f.a)
	doc := f.MainDoc(`"".endswith()`)
	help := f.a.SignatureHelp(doc, protocol.Position{Character: 12})
	assert.NotNil(t, help)
	if help == nil {
		return
	}
	assert.Equal(t, 1, len(help.Signatures))
	assert.Equal(t, "(suffix) -> bool", help.Signatures[0].Label)
	assert.Equal(t, uint32(0), help.ActiveParameter)
}

func TestTypedMethodSignatureHelp(t *testing.T) {
	f := newFixture(t)
	_ = WithStarlarkBuiltins()(f.a)
	doc := f.MainDoc(`d = {}
d.pop()`)
	pos := protocol.Position{Line: 1, Character: 6}
	help := f.a.SignatureHelp(doc, pos)
	assert.NotNil(t, help)
	if help == nil {
		return
	}
	assert.Equal(t, 1, len(help.Signatures))
	assert.Equal(t, "(key)", help.Signatures[0].Label)
	assert.Equal(t, uint32(0), help.ActiveParameter)
}
