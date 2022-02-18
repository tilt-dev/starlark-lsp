package analysis

import (
	"context"
	"fmt"
	"os"

	"go.lsp.dev/protocol"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
	"github.com/tilt-dev/starlark-lsp/pkg/query"
)

type Builtins struct {
	Functions map[string]protocol.SignatureInformation
	Symbols   []protocol.SymbolInformation
}

func (b *Builtins) Update(other *Builtins) {
	if len(other.Functions) > 0 {
		for name, sig := range other.Functions {
			b.Functions[name] = sig
		}
	}
	if len(other.Symbols) > 0 {
		b.Symbols = append(b.Symbols, other.Symbols...)
	}
}

func WithBuiltinPaths(paths []string) AnalyzerOption {
	return func(analyzer *Analyzer) error {
		builtins, err := LoadBuiltins(analyzer.context, paths)
		if err != nil {
			return err
		}
		analyzer.builtins.Update(builtins)
		return nil
	}
}

func WithBuiltinModulePaths(paths []string) AnalyzerOption {
	return func(analyzer *Analyzer) error {
		builtins, err := LoadBuiltinModules(analyzer.context, paths)
		if err != nil {
			return err
		}
		analyzer.builtins.Update(builtins)
		return nil
	}
}

func WithBuiltinFunctions(sigs map[string]protocol.SignatureInformation) AnalyzerOption {
	return func(analyzer *Analyzer) error {
		analyzer.builtins.Update(&Builtins{Functions: sigs})
		return nil
	}
}

func WithBuiltinSymbols(symbols []protocol.SymbolInformation) AnalyzerOption {
	return func(analyzer *Analyzer) error {
		analyzer.builtins.Update(&Builtins{Symbols: symbols})
		return nil
	}
}

func LoadBuiltinsFromFile(ctx context.Context, path string) (*Builtins, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return &Builtins{}, err
	}

	tree, err := query.Parse(ctx, contents)
	if err != nil {
		return &Builtins{}, fmt.Errorf("failed to parse %q: %v", path, err)
	}

	functions := make(map[string]protocol.SignatureInformation)
	symbols := []protocol.SymbolInformation{}
	doc := document.NewDocument(contents, tree)
	docFunctions := query.Functions(doc, tree.RootNode())
	// symbols := analysis.DocumentSymbols(doc)
	doc.Close()

	for fn, sig := range docFunctions {
		if _, ok := functions[fn]; ok {
			return &Builtins{}, fmt.Errorf("duplicate function %q found in %q", fn, path)
		}
		functions[fn] = sig
	}

	return &Builtins{
		Functions: functions,
		Symbols:   symbols,
	}, nil
}

func LoadBuiltins(ctx context.Context, filePaths []string) (*Builtins, error) {
	functions := make(map[string]protocol.SignatureInformation)
	symbols := []protocol.SymbolInformation{}

	for _, path := range filePaths {
		fileBuiltins, err := LoadBuiltinsFromFile(ctx, path)
		if err != nil {
			return &Builtins{}, err
		}
		for name, sig := range fileBuiltins.Functions {
			functions[name] = sig
		}
		symbols = append(symbols, fileBuiltins.Symbols...)
	}

	return &Builtins{
		Functions: functions,
		Symbols:   symbols,
	}, nil
}

func LoadBuiltinModules(ctx context.Context, moduleDirs []string) (*Builtins, error) {
	return &Builtins{}, nil
}
