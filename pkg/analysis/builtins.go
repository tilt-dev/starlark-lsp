package analysis

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.lsp.dev/protocol"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
	"github.com/tilt-dev/starlark-lsp/pkg/query"
)

type Builtins struct {
	Functions map[string]protocol.SignatureInformation
	Symbols   []protocol.DocumentSymbol
}

func NewBuiltins() *Builtins {
	return &Builtins{
		Functions: make(map[string]protocol.SignatureInformation),
		Symbols:   []protocol.DocumentSymbol{},
	}
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
	builtins := NewBuiltins()

	for _, path := range filePaths {
		fileBuiltins, err := LoadBuiltinsFromFile(ctx, path)
		if err != nil {
			return &Builtins{}, err
		}
		builtins.Update(fileBuiltins)
	}

	return builtins, nil
}

func LoadBuiltinModule(ctx context.Context, dir string) (*Builtins, error) {
	builtins := NewBuiltins()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		entryName := entry.Name()

		if !entry.IsDir() && !strings.HasSuffix(entryName, ".py") {
			continue
		}

		if entryName == "__init__.py" {
			initBuiltins, err := LoadBuiltinsFromFile(ctx, filepath.Join(dir, entryName))
			if err != nil {
				return nil, err
			}
			builtins.Update(initBuiltins)
			continue
		}

		var modName string
		if entry.IsDir() {
			modName = entryName
		} else {
			modName = entryName[:len(entryName)-3]
		}

		modSym := protocol.DocumentSymbol{
			Name:     modName,
			Kind:     protocol.SymbolKindVariable,
			Children: []protocol.DocumentSymbol{},
		}

		var modBuiltins *Builtins
		if entry.IsDir() {
			modBuiltins, err = LoadBuiltinModule(ctx, filepath.Join(dir, entryName))
		} else {
			modBuiltins, err = LoadBuiltinsFromFile(ctx, filepath.Join(dir, entryName))
		}

		if err != nil {
			return nil, err
		}

		for name, fn := range modBuiltins.Functions {
			builtins.Functions[modName+"."+name] = fn
		}
		for _, sym := range modBuiltins.Symbols {
			var kind protocol.SymbolKind
			switch sym.Kind {
			case protocol.SymbolKindFunction:
				kind = protocol.SymbolKindMethod
			default:
				kind = protocol.SymbolKindField
			}
			modSym.Children = append(modSym.Children, protocol.DocumentSymbol{
				Name:   sym.Name,
				Kind:   kind,
				Detail: sym.Detail,
			})
		}
		if len(modSym.Children) > 0 {
			builtins.Symbols = append(builtins.Symbols, modSym)
		}
	}
	return builtins, nil
}

func LoadBuiltinModules(ctx context.Context, moduleDirs []string) (*Builtins, error) {
	builtins := NewBuiltins()
	for _, dir := range moduleDirs {
		modBuiltins, err := LoadBuiltinModule(ctx, dir)
		if err != nil {
			return nil, err
		}
		builtins.Update(modBuiltins)
	}
	return builtins, nil
}
