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
	Symbols   []protocol.DocumentSymbol
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

func (b *Builtins) FunctionNames() []string {
	names := make([]string, len(b.Functions))
	i := 0
	for name := range b.Functions {
		names[i] = name
		i++
	}
	return names
}

func (b *Builtins) SymbolNames() []string {
	names := make([]string, len(b.Symbols))
	for i, sym := range b.Symbols {
		names[i] = sym.Name
	}
	return names
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

func WithBuiltinSymbols(symbols []protocol.DocumentSymbol) AnalyzerOption {
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
	doc := document.NewDocument(contents, tree)
	docFunctions := query.Functions(doc, tree.RootNode())
	symbols := query.DocumentSymbols(doc)
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
	builtins := &Builtins{
		Functions: make(map[string]protocol.SignatureInformation),
		Symbols:   []protocol.DocumentSymbol{},
	}

	for _, path := range filePaths {
		fileBuiltins, err := LoadBuiltinsFromFile(ctx, path)
		if err != nil {
			return &Builtins{}, err
		}
		builtins.Update(fileBuiltins)
	}

	return builtins, nil
}

func LoadBuiltinModule(ctx context.Context, name, dir string) (*Builtins, error) {
	return &Builtins{}, nil
}

func LoadBuiltinModules(ctx context.Context, moduleDirs []string) (*Builtins, error) {
	return &Builtins{}, nil
}
