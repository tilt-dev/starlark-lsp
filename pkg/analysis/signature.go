package analysis

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/protocol"

	"github.com/autokitteh/starlark-lsp/pkg/document"
	"github.com/autokitteh/starlark-lsp/pkg/query"
)

func (a *Analyzer) signatureInformation(doc document.Document, node *sitter.Node, args callWithArguments) (query.Signature, bool) {
	var sig query.Signature
	var found bool
	fnName := args.fnName

	// first try to find non-typed functions, e.g. local/global scope and builtins
	// NOTE that functions imported from builtins have a dot in their name denoting imported path, e.g. dir.file.func, but those functiosn aren't methods
	for n := node; n != nil && !query.IsModuleScope(doc, n); n = n.Parent() {
		sig, found = query.Function(doc, n, fnName)
		if found {
			break
		}
	}

	if !found {
		sig, found = doc.Functions()[fnName]
	}

	if !found {
		sig, found = a.builtins.Functions[fnName]
	}

	if found {
		return sig, sig.Name != ""
	}

	ind := strings.LastIndex(fnName, ".")
	if ind == -1 {
		return sig, false
	}
	mName := fnName[ind+1:]

	// handle AK rebinded builtins -------------------------------------------
	preDotName := fnName[:ind]
	sym := SymbolMatching(doc.Symbols(), preDotName)
	if akIsBindedSymbol(sym) {
		buitinSym := SymbolMatching(a.builtins.Symbols, sym.Detail)
		sig, _ = a.builtins.Functions[buitinSym.Name+"."+mName]
		return sig, sig.Name != ""
	} // ---------------------------------------------------------------------

	// at last, try to find whether it's a method
	if !found && strings.Contains(fnName, ".") {
		sig = a.builtins.Methods[mName]
		if sig.Name != "" {
			method, found := a.findTypedMethodForNode(doc, node, mName, args)
			if found {
				sig = method
			}
		}
	}
	return sig, sig.Name != ""
}

func (a *Analyzer) findTypedMethod(typeName string, methodName string) (query.Signature, bool) {
	sig := query.Signature{}
	if typeName != "" && methodName != "" {
		if t, ok := a.builtins.Types[typeName]; ok {
			return t.FindMethod(methodName)
		}
	}
	return sig, false
}

func (a *Analyzer) findTypedMethodForNode(doc document.Document, node *sitter.Node, methodName string, args callWithArguments) (query.Signature, bool) {
	afterDot := node.EndPoint() // assume that node passed is the object node
	afterDot.Column += 1

	if args.argsNode != nil {
		afterDot = args.argsNode.StartPoint()
		afterDot.Column -= uint32(len(methodName))
		if query.PointAfterOrEqual(node.StartPoint(), afterDot) {
			node = node.Parent()
		}
	}

	expr := a.findObjectExpression([]*sitter.Node{node}, afterDot)
	typeName := a.analyzeType(doc, expr)
	return a.findTypedMethod(typeName, methodName)
}

func (a *Analyzer) SignatureHelp(doc document.Document, pos protocol.Position) *protocol.SignatureHelp {
	pt := query.PositionToPoint(pos)
	node, ok := query.NodeAtPoint(doc, pt)
	if !ok {
		return nil
	}

	args := possibleCallInfo(doc, node, pt)
	if args.fnName == "" {
		// avoid computing function defs
		return nil
	}

	sig, ok := a.signatureInformation(doc, node, args)
	if !ok {
		return nil
	}

	activeParam := uint32(0)

	if args.currentKeyword != "" {
		for i, param := range sig.Params {
			if param.Name == args.currentKeyword {
				activeParam = uint32(i)
				break
			}
		}
	} else if args.positional == args.total {
		activeParam = args.positional
	}

	if activeParam > uint32(len(sig.Params)-1) {
		activeParam = uint32(len(sig.Params) - 1)
	}

	return &protocol.SignatureHelp{
		Signatures:      []protocol.SignatureInformation{sig.SignatureInfo()},
		ActiveParameter: activeParam,
		ActiveSignature: 0,
	}
}

type callWithArguments struct {
	fnName            string
	positional, total uint32
	keywords          map[string]bool
	currentKeyword    string
	argsNode          *sitter.Node
}

// possibleCallInfo attempts to find the name of the function for a
// `call`.
//
// Currently, this supports two cases:
//
//	(1) Current node is inside of a `call`
//	(2) Current node is inside of an ERROR block where first child is an
//		`identifier`
func possibleCallInfo(doc document.Document, node *sitter.Node, pt sitter.Point) (args callWithArguments) {
	for n := node; n != nil; n = n.Parent() {
		if n.Type() == query.NodeTypeCall {
			argsList := n.ChildByFieldName("arguments")
			args = possibleActiveParam(doc, argsList.Child(0), pt)
			args.fnName = doc.Content(n.ChildByFieldName("function"))
			args.argsNode = argsList
			return args
		} else if n.Type() == query.NodeTypeArgList {
			continue
		} else if n.HasError() {
			// look for `foo(` and assume it's a function call - this could
			// happen if the closing `)` is not (yet) present or if there's
			// something invalid going on within the args, e.g. `foo(x#)`
			possibleCall := n.NamedChild(0)
			if possibleCall != nil {
				possibleParen := possibleCall.NextSibling()
				if possibleParen != nil && possibleParen.Type() == "(" {
					switch possibleCall.Type() {
					case query.NodeTypeIdentifier, query.NodeTypeAttribute:
						args = possibleActiveParam(doc, possibleParen.NextSibling(), pt)
						args.argsNode = possibleParen
						args.fnName = doc.Content(possibleCall)
						return args
					}
				}
			}
		}
	}
	return args
}

func possibleActiveParam(doc document.Document, node *sitter.Node, pt sitter.Point) callWithArguments {
	args := callWithArguments{keywords: make(map[string]bool)}
	for n := node; n != nil; n = n.NextSibling() {
		inRange := query.PointBeforeOrEqual(n.StartPoint(), pt) &&
			query.PointBeforeOrEqual(n.EndPoint(), pt)
		if !inRange {
			break
		}

		switch n.Type() {
		case ",":
			args.total++
			if len(args.keywords) == 0 {
				args.positional++
			}
			args.currentKeyword = ""
		case query.NodeTypeERROR:
			if doc.Content(n) != "=" {
				break
			}
			fallthrough
		case "=":
			if n.PrevSibling().Type() == query.NodeTypeIdentifier {
				name := doc.Content(n.PrevSibling())
				args.keywords[name] = true
				args.currentKeyword = name
			}
		case query.NodeTypeKeywordArgument:
			name := doc.Content(n.ChildByFieldName("name"))
			args.keywords[name] = true
			args.currentKeyword = name
		}
	}
	return args
}
