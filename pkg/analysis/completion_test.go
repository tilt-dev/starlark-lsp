package analysis

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
)

func TestSimpleCompletion(t *testing.T) {
	f := newFixture(t)

	f.Symbols("foo", "bar", "baz")

	f.Document("")
	result := f.a.Completion(f.doc, protocol.Position{})
	assert.Equal(t, 3, len(result.Items))
	assert.Equal(t, "foo", result.Items[0].Label)
	assert.Equal(t, "bar", result.Items[1].Label)
	assert.Equal(t, "baz", result.Items[2].Label)

	f.Document("ba")
	result = f.a.Completion(f.doc, protocol.Position{Character: 2})
	assert.Equal(t, 2, len(result.Items))
	assert.Equal(t, "bar", result.Items[0].Label)
	assert.Equal(t, "baz", result.Items[1].Label)
}
