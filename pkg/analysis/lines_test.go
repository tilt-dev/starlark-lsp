package analysis_test

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"

	"github.com/tilt-dev/starlark-lsp/pkg/analysis"
)

func TestLinesInfo(t *testing.T) {
	input := `12345
7890
abcd`

	type tc struct {
		offset uint32
		line   uint32
		col    uint32
	}

	tcs := []tc{
		{offset: 0, line: 0, col: 0},
		{offset: 1, line: 0, col: 1},
		{offset: 8, line: 1, col: 3},
	}

	lines := analysis.NewLineOffsets([]byte(input))
	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("Offset object: %s", spew.Sdump(lines))
		}
	})

	for _, tt := range tcs {
		pos := lines.PositionForOffset(tt.offset)
		ok := assert.Equalf(t,
			protocol.Position{Line: tt.line, Character: tt.col},
			pos,
			"Wrong position for offset: %d", tt.offset)

		if ok {
			offset := lines.OffsetForPosition(pos)
			assert.Equalf(t, tt.offset, offset,
				"Wrong offset for position: (line=%d, char=%d)",
				pos.Line, pos.Character)
		}

		pt := lines.PointForOffset(tt.offset)
		ok = assert.Equalf(t,
			sitter.Point{Row: tt.line, Column: tt.col},
			pt,
			"Wrong point for offset: %d", tt.offset)

		if ok {
			offset := lines.OffsetForPoint(pt)
			assert.Equalf(t, tt.offset, offset,
				"Wrong offset for point: (row=%d, col=%d)",
				pt.Row, pt.Column)
		}
	}
}
