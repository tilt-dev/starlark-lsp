package query_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"

	"github.com/autokitteh/starlark-lsp/pkg/query"
)

func TestStarlarkTypes(t *testing.T) {
	code := `
class Student():
    """A sleepy student."""
    name: str
    """Student name."""
    def sleep(self):
        """Tell student to sleep."""
        pass
`
	f := newQueryFixture(t, "", code)
	doc := f.document()
	classes := query.Types(doc, f.root)
	assert.Equal(t, 1, len(classes))
	if len(classes) == 1 {
		class := classes[0]
		assert.Equal(t, "Student", class.Name)
		assert.Equal(t, 1, len(class.Methods))
		if len(class.Methods) == 1 {
			assert.Equal(t, "sleep", class.Methods[0].Name)
			assert.Equal(t, 0, len(class.Methods[0].Params))
			assert.Equal(t, "Tell student to sleep.", class.Methods[0].Docs.Description)
		}
		assert.Equal(t, 1, len(class.Fields))
		if len(class.Fields) == 1 {
			assert.Equal(t, "name", class.Fields[0].Name)
			assert.Equal(t, protocol.SymbolKindString, class.Fields[0].Kind)
			assert.Equal(t, "Student name.", class.Fields[0].Detail)
		}
	}
}

func TestTypesEmptyClass(t *testing.T) {
	code := `
class Student():
    """A sleepy student."""
    pass
`
	f := newQueryFixture(t, "", code)
	doc := f.document()
	classes := query.Types(doc, f.root)
	assert.Equal(t, 0, len(classes))
}
