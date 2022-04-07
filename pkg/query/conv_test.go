package query_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tilt-dev/starlark-lsp/pkg/query"
)

func TestUnquote(t *testing.T) {
	cases := [][2]string{
		{`"hello"`, "hello"},
		{`r"hello"`, "hello"},
		{`r"he\nllo"`, "he\\nllo"},
		{`b"hello"`, "hello"},
		{`rb"hello"`, "hello"},
		{`"""hello"""`, "hello"},
		{`'''hello'''`, "hello"},
		{`r'''hello'''`, "hello"},
		{`b'''hello'''`, "hello"},
		{`"he\nllo"`, "he\nllo"},
		{`"he\"llo"`, "he\"llo"},
		{`"he\rllo"`, "he\rllo"},
		{`"he\tllo"`, "he\tllo"},
		{`"he\\llo"`, "he\\llo"},
		{`"he\nll\no"`, "he\nll\no"},
		{`"he\u0034llo"`, "he\u0034llo"},
		{`"he\x32llo"`, "he\x32llo"},
		{`"he\001llo"`, "he\001llo"},
		{`"hello\
world"`, "helloworld"},
		{`"hello\n\n"`, "hello\n\n"},
		{`"\n\t\v\b"`, "\n\t\v\b"},
		{`""`, ""},
	}
	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			q := newQueryFixture(t, nil, tt[0])
			v := query.Unquote(q.input, q.root)
			assert.Equal(t, tt[1], v)
		})
	}
}

func TestUnquotePanic(t *testing.T) {
	q := newQueryFixture(t, nil, `hello`)
	defer func() {
		e := recover().(error)
		assert.NotNil(t, e)
		assert.Equal(t, "[Unquote:bug:unexpected node: identifier: hello]", e.Error())
	}()
	query.Unquote(q.input, q.root)
}
