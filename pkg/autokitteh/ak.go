package autokitteh

import (
	"embed"
	"io/fs"
)

//go:embed builtins/*.py
var akBuiltins embed.FS

type BuiltinFSProvider = func() fs.FS

func BuiltinsFSProvider() (BuiltinFSProvider, error) {
	builtinFSProvider, err := fs.Sub(akBuiltins, "builtins")
	return func() fs.FS { return builtinFSProvider }, err
}
