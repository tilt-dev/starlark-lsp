package document

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"

	"github.com/tilt-dev/starlark-lsp/pkg/query"
)

type LoadStatement struct {
	File        string
	Symbols     [][2]string
	Range       protocol.Range
	Diagnostics []protocol.Diagnostic
}

type Document interface {
	Input() []byte
	Content(n *sitter.Node) string
	ContentRange(r sitter.Range) string

	Tree() *sitter.Tree
	Functions() map[string]protocol.SignatureInformation
	Symbols() []protocol.DocumentSymbol
	Diagnostics() []protocol.Diagnostic
	Loads() []LoadStatement

	Copy() Document

	Close()
}

type NewDocumentFunc func(u uri.URI, input []byte, tree *sitter.Tree) Document

func NewDocument(u uri.URI, input []byte, tree *sitter.Tree) Document {
	return &document{
		uri:   u,
		input: input,
		tree:  tree,
	}
}

func NewDocumentWithSymbols(u uri.URI, input []byte, tree *sitter.Tree) Document {
	doc := &document{
		uri:   u,
		input: input,
		tree:  tree,
	}
	doc.functions = query.Functions(doc, tree.RootNode())
	doc.symbols = query.DocumentSymbols(doc)
	doc.parseLoadStatements()
	return doc
}

func NodesToContent(doc Document, nodes []*sitter.Node) string {
	var content string
	if len(nodes) > 0 {
		content = doc.ContentRange(sitter.Range{
			StartByte: nodes[0].StartByte(),
			EndByte:   nodes[len(nodes)-1].EndByte(),
		})
	} else {
		content = doc.Content(doc.Tree().RootNode())
	}
	return content
}

type document struct {
	uri uri.URI

	// input is the file as it exists in the editor buffer.
	input []byte

	// tree represents the parsed version of the document.
	tree *sitter.Tree

	functions   map[string]protocol.SignatureInformation
	symbols     []protocol.DocumentSymbol
	diagnostics []protocol.Diagnostic
	loads       []LoadStatement
}

var _ Document = &document{}

func (d *document) Input() []byte {
	return d.input
}

func (d *document) Content(n *sitter.Node) string {
	return n.Content(d.input)
}

func (d *document) ContentRange(r sitter.Range) string {
	return string(d.input[r.StartByte:r.EndByte])
}

func (d *document) Tree() *sitter.Tree {
	return d.tree
}

func (d *document) Functions() map[string]protocol.SignatureInformation {
	return d.functions
}

func (d *document) Symbols() []protocol.DocumentSymbol {
	return d.symbols
}

func (d *document) Diagnostics() []protocol.Diagnostic {
	return d.diagnostics
}

func (d *document) Loads() []LoadStatement {
	return d.loads
}

func (d *document) Close() {
	d.tree.Close()
}

// Copy creates a shallow copy of the Document.
//
// The Contents byte slice is returned as-is.
// A shallow copy of the Tree is made, as Tree-sitter trees are not thread-safe.
func (d *document) Copy() Document {
	doc := &document{
		uri:         d.uri,
		input:       d.input,
		tree:        d.tree.Copy(),
		functions:   make(map[string]protocol.SignatureInformation),
		symbols:     append([]protocol.DocumentSymbol{}, d.symbols...),
		loads:       append([]LoadStatement{}, d.loads...),
		diagnostics: append([]protocol.Diagnostic{}, d.diagnostics...),
	}
	for fn := range d.functions {
		doc.functions[fn] = d.functions[fn]
	}
	return doc
}

func (d *document) processLoads(ctx context.Context, m *Manager) {
	for i, load := range d.loads {
		if load.File == "" {
			continue
		}
		path, err := resolvePath(load.File, d.uri)
		var dep Document
		if err == nil {
			dep, err = m.readAndParse(ctx, path)
		}
		if err != nil {
			diag := protocol.Diagnostic{
				Range:    load.Range,
				Severity: protocol.DiagnosticSeverityError,
				Message:  err.Error(),
			}
			d.loads[i].Diagnostics = append(d.loads[i].Diagnostics, diag)
			d.diagnostics = append(d.diagnostics, diag)
			continue
		}
		fns := dep.Functions()
		symMap := make(map[string]protocol.DocumentSymbol)
		for _, s := range dep.Symbols() {
			symMap[s.Name] = s
		}
		for _, v := range load.Symbols {
			if sym, found := symMap[v[1]]; found {
				sym.Name = v[0]
				sym.Range = load.Range
				d.symbols = append(d.symbols, sym)
				if f, ok := fns[v[1]]; ok {
					d.functions[v[0]] = f
				}
			} else {
				d.diagnostics = append(d.diagnostics, protocol.Diagnostic{
					Range:    load.Range,
					Severity: protocol.DiagnosticSeverityWarning,
					Message:  fmt.Sprintf("symbol '%s' not found in %s", v[1], load.File),
				})
			}
		}
		for _, depload := range dep.Loads() {
			for _, diag := range depload.Diagnostics {
				diag.Range = load.Range
				d.diagnostics = append(d.diagnostics, diag)
			}
		}
	}
}

func (d *document) parseLoadStatements() {
	nodes := query.LoadStatements(d.input, d.tree)
	for _, n := range nodes {
		parent := n.Parent()
	parentloop:
		for parent != nil {
			switch parent.Type() {
			case query.NodeTypeBlock, query.NodeTypeExpressionStatement:
				parent = parent.Parent()
			default:
				break parentloop
			}
		}

		if parent == d.tree.RootNode() {
			load, diagnostics := loadStatement(d.input, n)
			d.loads = append(d.loads, load)
			d.diagnostics = append(d.diagnostics, diagnostics...)
		} else {
			d.diagnostics = append(d.diagnostics, protocol.Diagnostic{
				Range:    query.NodeRange(n),
				Severity: protocol.DiagnosticSeverityError,
				Message:  fmt.Sprintf("load statement not allowed in %s", withArticle(strings.ReplaceAll(parent.Type(), "_", " "))),
			})
		}
	}
}

func loadStatement(input []byte, n *sitter.Node) (LoadStatement, []protocol.Diagnostic) {
	load := LoadStatement{Range: query.NodeRange(n)}
	diagnostics := []protocol.Diagnostic{}
	argsNode := n.ChildByFieldName("arguments")
	notAString := func(nn *sitter.Node) protocol.Diagnostic {
		return protocol.Diagnostic{
			Range:    query.NodeRange(nn),
			Severity: protocol.DiagnosticSeverityError,
			Message:  fmt.Sprintf("load parameter must be a literal string, found '%s'", nn.Content(input)),
		}
	}
	args := make([]*sitter.Node, argsNode.NamedChildCount())
	for i := range args {
		args[i] = argsNode.NamedChild(i)
	}

	if len(args) > 0 {
		fileArg := args[0]
		if fileArg.Type() == query.NodeTypeString {
			load.File = query.Unquote(input, fileArg)
		} else {
			diagnostics = append(diagnostics, notAString(fileArg))
		}
	}

	if len(args) > 1 {
		for _, va := range args[1:] {
			switch va.Type() {
			case query.NodeTypeString:
				s := query.Unquote(input, va)
				load.Symbols = append(load.Symbols, [2]string{s, s})
			case query.NodeTypeKeywordArgument:
				alias := va.ChildByFieldName("name").Content(input)
				nameNode := va.ChildByFieldName("value")
				if nameNode.Type() == query.NodeTypeString {
					load.Symbols = append(load.Symbols, [2]string{alias, query.Unquote(input, nameNode)})
				} else {
					diagnostics = append(diagnostics, notAString(nameNode))
				}
			default:
				diagnostics = append(diagnostics, notAString(va))
			}
		}
	} else {
		diagnostics = append(diagnostics, protocol.Diagnostic{
			Range:    query.NodeRange(n),
			Severity: protocol.DiagnosticSeverityWarning,
			Message:  "load statement did not specify any symbols to import",
		})
	}
	return load, diagnostics
}

// Resolve the given (possible relative) path from the parent directory of the
// relativeTo URI.
func resolvePath(path string, relativeTo uri.URI) (uri.URI, error) {
	var err error
	if strings.Contains(path, "://") {
		var url *url.URL
		url, err = url.Parse(path)
		if err == nil {
			if url.Scheme != "file" {
				// The provided ReadDocumentFunc must handle this scheme
				return uri.URI(path), nil
			} else {
				path = filepath.FromSlash(url.Path)
			}
		}
	}

	if err != nil {
		return "", err
	}

	if filepath.IsAbs(path) {
		return uri.File(path), nil
	}

	relPath, err := filename(relativeTo)
	if err != nil {
		return "", err
	}
	relPath = filepath.Dir(relPath)
	return uri.File(filepath.Join(relPath, path)), nil
}

func withArticle(s string) string {
	article := "a"
	if strings.ContainsAny(s[0:1], "aeiou") {
		article = "an"
	}
	return fmt.Sprintf("%s %s", article, s)
}
