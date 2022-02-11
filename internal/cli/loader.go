package cli

import (
	"context"
	"fmt"
	"os"

	"go.lsp.dev/protocol"

	"github.com/tilt-dev/starlark-lsp/pkg/analysis"
	"github.com/tilt-dev/starlark-lsp/pkg/document"
)

type Builtins struct {
	Functions map[string]protocol.SignatureInformation
	Symbols   []protocol.SymbolInformation
}

func LoadBuiltins(ctx context.Context, paths ...string) (Builtins, error) {
	functions := make(map[string]protocol.SignatureInformation)
	// TODO(milas): fix Symbol analysis query and include
	var builtinSymbols []protocol.SymbolInformation

	for _, path := range paths {
		contents, err := os.ReadFile(path)
		if err != nil {
			return Builtins{}, err
		}

		tree, err := analysis.Parse(ctx, contents)
		if err != nil {
			return Builtins{}, fmt.Errorf("failed to parse %q: %v", path, err)
		}

		doc := document.NewDocument(contents, tree)
		docFunctions := analysis.Functions(doc, tree.RootNode())
		// symbols := analysis.DocumentSymbols(doc)

		for fn, sig := range docFunctions {
			if _, ok := functions[fn]; ok {
				return Builtins{}, fmt.Errorf("duplicate function %q found in %q", fn, path)
			}
			functions[fn] = sig
		}
	}

	builtins := Builtins{
		Functions: functions,
		Symbols:   builtinSymbols,
	}
	return builtins, nil
}