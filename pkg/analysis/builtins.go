package analysis

import (
	"context"
	_ "embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
	"github.com/tilt-dev/starlark-lsp/pkg/query"
)

type Builtins struct {
	Signatures map[string]query.Signature
	Functions  map[string]protocol.SignatureInformation
	Symbols    []protocol.DocumentSymbol
}

//go:embed builtins.py
var StarlarkBuiltins []byte

func NewBuiltins() *Builtins {
	return &Builtins{
		Signatures: make(map[string]query.Signature),
		Functions:  make(map[string]protocol.SignatureInformation),
		Symbols:    []protocol.DocumentSymbol{},
	}
}

func (b *Builtins) IsEmpty() bool {
	return len(b.Signatures) == 0 && len(b.Symbols) == 0
}

func (b *Builtins) Update(other *Builtins) {
	if len(other.Signatures) > 0 {
		for name, sig := range other.Signatures {
			b.Signatures[name] = sig
			b.Functions[name] = other.Functions[name]
		}
	}
	if len(other.Symbols) > 0 {
		b.Symbols = append(b.Symbols, other.Symbols...)
	}
}

func (b *Builtins) FunctionNames() []string {
	names := make([]string, len(b.Signatures))
	i := 0
	for name := range b.Signatures {
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
		for _, path := range paths {
			builtins, err := LoadBuiltins(analyzer.context, path)
			if err != nil {
				return err
			}
			analyzer.builtins.Update(builtins)
		}
		return nil
	}
}

func WithBuiltins(f fs.FS) AnalyzerOption {
	return func(analyzer *Analyzer) error {
		builtins, err := LoadBuiltinsFromFS(analyzer.context, f)
		if err != nil {
			return err
		}
		analyzer.builtins.Update(builtins)
		return nil
	}
}

func WithStarlarkBuiltins() AnalyzerOption {
	return func(analyzer *Analyzer) error {
		builtins, err := LoadBuiltinsFromSource(analyzer.context, StarlarkBuiltins, "builtins.py")
		if err != nil {
			return errors.Wrapf(err, "loading builtins from builtins.py")
		}
		analyzer.builtins.Update(&Builtins{
			Symbols: []protocol.DocumentSymbol{
				{Name: "False", Kind: protocol.SymbolKindBoolean},
				{Name: "None", Kind: protocol.SymbolKindNull},
				{Name: "True", Kind: protocol.SymbolKindBoolean},
			},
		})
		analyzer.builtins.Update(builtins)
		return nil
	}
}

func LoadBuiltinsFromSource(ctx context.Context, contents []byte, path string) (*Builtins, error) {
	tree, err := query.Parse(ctx, contents)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse %q", path)
	}

	doc := document.NewDocument(uri.File(path), contents, tree)
	docSignatures := doc.FunctionSignatures()
	docFunctions := doc.Functions()
	symbols := doc.Symbols()
	doc.Close()

	return &Builtins{
		Signatures: docSignatures,
		Functions:  docFunctions,
		Symbols:    symbols,
	}, nil
}

func LoadBuiltinsFromFile(ctx context.Context, path string, f fs.FS) (*Builtins, error) {
	var contents []byte
	var err error
	if f != nil {
		contents, err = fs.ReadFile(f, path)
	} else {
		contents, err = os.ReadFile(path)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "reading %s", path)
	}
	return LoadBuiltinsFromSource(ctx, contents, path)
}

func loadBuiltinsWalker(ctx context.Context, f fs.FS) (map[string]*Builtins, fs.WalkDirFunc) {
	builtins := make(map[string]*Builtins)
	return builtins, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		entryName := entry.Name()

		if !entry.IsDir() && !strings.HasSuffix(entryName, ".py") {
			return nil
		}

		modPath := path

		if entry.IsDir() {
			builtins[modPath] = NewBuiltins()
			return nil
		}

		if entryName == "__init__.py" {
			modPath = filepath.Dir(modPath)
		} else {
			modPath = path[:len(path)-len(".py")]
		}

		modBuiltins, err := LoadBuiltinsFromFile(ctx, path, f)
		if err != nil {
			return errors.Wrapf(err, "loading builtins from %s", path)
		}

		if b, ok := builtins[modPath]; ok {
			b.Update(modBuiltins)
		} else {
			builtins[modPath] = modBuiltins
		}
		return nil
	}
}

func LoadBuiltinsFromFS(ctx context.Context, f fs.FS) (*Builtins, error) {
	root := "."

	builtinsMap, walker := loadBuiltinsWalker(ctx, f)
	err := fs.WalkDir(f, root, walker)

	if err != nil {
		return nil, errors.Wrapf(err, "walking %s", root)
	}

	modulePaths := make([]string, len(builtinsMap))
	i := 0
	for modPath := range builtinsMap {
		modulePaths[i] = modPath
		i++
	}
	sort.Sort(sort.Reverse(sort.StringSlice(modulePaths)))

	for _, modPath := range modulePaths {
		mod := builtinsMap[modPath]
		if mod.IsEmpty() || modPath == root {
			continue
		}

		modName := filepath.Base(modPath)
		parentModPath := filepath.Dir(modPath)
		parentMod, ok := builtinsMap[parentModPath]
		if !ok {
			return nil, fmt.Errorf("no entry for parent %s", parentModPath)
		}

		copyBuiltinsToParent(mod, parentMod, modName)
	}

	builtins, ok := builtinsMap[root]
	if !ok {
		return nil, fmt.Errorf("no entry for root %s", root)
	}
	return builtins, nil
}

func copyBuiltinsToParent(mod, parentMod *Builtins, modName string) {
	for name, fn := range mod.Signatures {
		parentMod.Signatures[modName+"."+name] = fn
		parentMod.Functions[modName+"."+name] = mod.Functions[name]
	}

	children := []protocol.DocumentSymbol{}
	for _, sym := range mod.Symbols {
		var kind protocol.SymbolKind
		switch sym.Kind {
		case protocol.SymbolKindFunction:
			kind = protocol.SymbolKindMethod
		default:
			kind = protocol.SymbolKindField
		}
		childSym := sym
		childSym.Kind = kind
		children = append(children, childSym)
	}

	if len(children) > 0 {
		existingIndex := -1
		for i, sym := range parentMod.Symbols {
			if sym.Name == modName {
				existingIndex = i
				break
			}
		}

		if existingIndex >= 0 {
			parentMod.Symbols[existingIndex].Children = append(parentMod.Symbols[existingIndex].Children, children...)
		} else {
			parentMod.Symbols = append(parentMod.Symbols, protocol.DocumentSymbol{
				Name:     modName,
				Kind:     protocol.SymbolKindVariable,
				Children: children,
			})
		}
	}
}

func LoadBuiltins(ctx context.Context, path string) (*Builtins, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, errors.Wrapf(err, "statting %s", path)
	}

	var result *Builtins
	if fileInfo.IsDir() {
		result, err = LoadBuiltinsFromFS(ctx, os.DirFS(path))
	} else {
		result, err = LoadBuiltinsFromFile(ctx, path, nil)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "loading builtins from %s", path)
	}

	return result, nil
}
