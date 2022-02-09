package server

import (
	"go.lsp.dev/protocol"
	"go.uber.org/zap"
)

func positionField(pos protocol.Position) zap.Field {
	return zap.Uint32s("pos", []uint32{pos.Line, pos.Character})
}
