package analysis

import (
	"fmt"
	"strings"

	"go.lsp.dev/protocol"
	"go.uber.org/zap"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/autokitteh/starlark-lsp/pkg/document"
	"github.com/autokitteh/starlark-lsp/pkg/query"
)

func SymbolMatching(symbols []query.Symbol, name string) query.Symbol {
	for _, sym := range symbols {
		if sym.Name == name {
			return sym
		}
	}
	return query.Symbol{}
}

// ak: deal with binded symbols (sym.Name -> sym.Detail) ---------------------
func akIsBindedSymbol(sym query.Symbol) bool {
	if len(sym.Tags) > 0 {
		for _, tag := range sym.Tags {
			if tag == query.Binded {
				return true
			}
		}
	}
	return false
}

// find and return either symbol or resolved binded symbol
func akSymbolMatching(symbols []query.Symbol, name string) query.Symbol {
	sym := SymbolMatching(symbols, name)
	if akIsBindedSymbol(sym) {
		return SymbolMatching(symbols, sym.Detail)
	}
	return sym
} // -------------------------------------------------------------------------

func SymbolsStartingWith(symbols []query.Symbol, prefix string) []query.Symbol {
	if prefix == "" {
		return symbols
	}
	result := []query.Symbol{}
	for _, sym := range symbols {
		if strings.HasPrefix(sym.Name, prefix) {
			result = append(result, sym)
		}
	}
	return result
}

func ToCompletionItemKind(k protocol.SymbolKind) protocol.CompletionItemKind {
	switch k {
	case protocol.SymbolKindField:
		return protocol.CompletionItemKindField
	case protocol.SymbolKindFunction:
		return protocol.CompletionItemKindFunction
	case protocol.SymbolKindMethod:
		return protocol.CompletionItemKindMethod
	default:
		return protocol.CompletionItemKindVariable
	}
}

func (a *Analyzer) Completion(doc document.Document, pos protocol.Position) *protocol.CompletionList {
	pt := query.PositionToPoint(pos)
	nodes, ok := a.nodesAtPointForCompletion(doc, pt)
	symbols := []query.Symbol{}

	if ok {
		symbols = a.completeExpression(doc, nodes, pt)
	}

	completionList := &protocol.CompletionList{
		Items: make([]protocol.CompletionItem, len(symbols)),
	}

	names := make([]string, len(symbols))
	for i, sym := range symbols {
		names[i] = sym.Name
		var sortText string
		if strings.HasSuffix(sym.Name, "=") {
			sortText = fmt.Sprintf("0%s", sym.Name)
		} else {
			sortText = fmt.Sprintf("1%s", sym.Name)
		}
		firstDetailLine := strings.SplitN(sym.Detail, "\n", 2)[0]
		completionList.Items[i] = protocol.CompletionItem{
			Label:    sym.Name,
			Detail:   firstDetailLine,
			Kind:     ToCompletionItemKind(sym.Kind),
			SortText: sortText,
		}
	}

	if len(names) > 0 {
		a.logger.Debug("completion result", zap.Strings("symbols", names))
	}
	return completionList
}

var availableSymbols []query.Symbol

func (a *Analyzer) completeExpression(doc document.Document, nodes []*sitter.Node, pt sitter.Point) []query.Symbol {
	var nodeAtPoint *sitter.Node
	if len(nodes) > 0 {
		nodeAtPoint = nodes[len(nodes)-1]
	}
	symbols := a.availableSymbols(doc, nodeAtPoint, pt)
	availableSymbols = symbols // FIXME:
	identifiers := query.ExtractIdentifiers(doc, nodes, &pt)

	a.logger.Debug("completion attempt",
		zap.String("code", document.NodesToContent(doc, nodes)),
		zap.Strings("nodes", func() []string {
			types := make([]string, len(nodes))
			for i, n := range nodes {
				types[i] = n.Type()
			}
			return types
		}()),
		zap.Strings("identifiers", identifiers),
	)

	for i, id := range identifiers {
		if i < len(identifiers)-1 {
			sym := akSymbolMatching(symbols, id)
			symbols = sym.Children
			a.logger.Debug("children",
				zap.String("id", id),
				zap.Strings("names", func() []string {
					names := make([]string, len(symbols))
					for j, s := range symbols {
						names[j] = s.Name
					}
					return names
				}()))
		} else {
			symbols = SymbolsStartingWith(symbols, id)
		}
	}
	if len(symbols) != 0 {
		return symbols
	}

	lastId := identifiers[len(identifiers)-1]
	pt = sitter.Point{Row: pt.Row, Column: pt.Column - uint32(len(lastId))}
	expr := a.findAttrObjectExpression(nodes, pt)

	if expr != nil {
		a.logger.Debug("dot object completion", zap.Strings("objs", identifiers[:len(identifiers)-1]))

		if expr.Type() == query.NodeTypeIdentifier { // FIXME: remove from here? better place? AnalyzeTYpe?
			if sym, found := a.FindDefinition(doc, expr, doc.Content(expr)); found {
				if sym.Kind == protocol.SymbolKindVariable {
					origSym := a.resolveSymbolTypeAndKind(doc, sym, availableSymbols)
					if origSym.Kind != protocol.SymbolKindVariable { // FIXME: better way to know that we resolved
						return SymbolsStartingWith(a.availableMembersForType(origSym.GetType()), lastId)
					}
				}
			}
		}

		return SymbolsStartingWith(a.availableMembersForNode(doc, expr), lastId)
	}

	return symbols
}

// Returns a list of available symbols for completion as follows:
//   - If in a function argument list, include keyword args for that function
//   - Add symbols in scope for the node at point, excluding symbols at the module
//     level (document symbols), because the document already has those computed
//   - Add document symbols
//   - Add builtins
func (a *Analyzer) availableSymbols(doc document.Document, nodeAtPoint *sitter.Node, pt sitter.Point) []query.Symbol {
	symbols := []query.Symbol{}
	if nodeAtPoint != nil {
		if args := keywordArgContext(doc, nodeAtPoint, pt); args.fnName != "" {
			if fn, ok := a.signatureInformation(doc, nodeAtPoint, args); ok {
				symbols = append(symbols, a.keywordArgSymbols(fn, args)...)
			}
		}
		symbols = append(symbols, query.SymbolsInScope(doc, nodeAtPoint)...)
	}
	docAndBuiltin := append(doc.Symbols(), a.builtins.Symbols...)
	for _, sym := range docAndBuiltin {
		found := false
		for _, s := range symbols {
			if sym.Name == s.Name {
				found = true
				break
			}
		}
		if !found {
			symbols = append(symbols, sym)
		}
	}

	return symbols
}

func (a *Analyzer) nodesAtPointForCompletion(doc document.Document, pt sitter.Point) ([]*sitter.Node, bool) {
	node, ok := query.NodeAtPoint(doc, pt)
	if !ok {
		return []*sitter.Node{}, false
	}
	a.logger.Debug("node at point", zap.String("node", node.Type()))
	return a.nodesForCompletion(doc, node, pt)
}

// Zoom in or out from the node to include adjacent attribute expressions, so we can
// complete starting from the top-most attribute expression.
func (a *Analyzer) nodesForCompletion(doc document.Document, node *sitter.Node, pt sitter.Point) ([]*sitter.Node, bool) {
	nodes := []*sitter.Node{}
	switch node.Type() {
	case query.NodeTypeString, query.NodeTypeComment:
		if query.PointCovered(pt, node) {
			// No completion inside a string or comment
			return nodes, false
		}
	case query.NodeTypeModule, query.NodeTypeBlock:
		// Sometimes the top-level module is the most granular node due to
		// location of the point being between children, in this case, advance
		// to the first child node that appears after the point
		if node.NamedChildCount() > 0 {
			for node = node.NamedChild(0); node != nil && query.PointBefore(node.StartPoint(), pt); {
				next := node.NextNamedSibling()
				if next == nil {
					break
				}
				node = next
			}
			return a.nodesForCompletion(doc, node, pt)
		}

	case query.NodeTypeIfStatement,
		query.NodeTypeExpressionStatement,
		query.NodeTypeForStatement,
		query.NodeTypeAssignment:
		if node.NamedChildCount() == 1 {
			return a.nodesForCompletion(doc, node.NamedChild(0), pt)
		}

		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			if query.PointBefore(child.EndPoint(), pt) {
				return a.leafNodesForCompletion(doc, child, pt)
			}
		}

	case query.NodeTypeAttribute, query.NodeTypeIdentifier:
		// If inside an attribute expression, capture the larger expression for
		// completion.
		switch node.Parent().Type() {
		case query.NodeTypeAttribute:
			nodes, _ = a.nodesForCompletion(doc, node.Parent(), pt)
		}

	case query.NodeTypeERROR, query.NodeTypeArgList:
		leafNodes, ok := a.leafNodesForCompletion(doc, node, pt)
		if len(leafNodes) > 0 {
			return leafNodes, ok
		}
		node = node.Child(int(node.ChildCount()) - 1)
	}

	if len(nodes) == 0 {
		nodes = append(nodes, node)
	}
	return nodes, true
}

// Look at all leaf nodes for the node and its previous sibling in a
// flattened slice, in order of appearance. Take all consecutive trailing
// identifiers or '.' as the attribute expression to complete.
func (a *Analyzer) leafNodesForCompletion(doc document.Document, node *sitter.Node, pt sitter.Point) ([]*sitter.Node, bool) {
	leafNodes := []*sitter.Node{}

	if node.PrevNamedSibling() != nil {
		leafNodes = append(leafNodes, query.LeafNodes(node.PrevNamedSibling())...)
	}

	leafNodes = append(leafNodes, query.LeafNodes(node)...)

	// count number of trailing id/'.' nodes, if any
	trailingCount := 0
	leafCount := len(leafNodes)
	for i := 0; i < leafCount && i == trailingCount; i++ {
		switch leafNodes[leafCount-1-i].Type() {
		case query.NodeTypeIdentifier, ".":
			trailingCount++
		}
	}
	nodes := make([]*sitter.Node, trailingCount)
	for j := 0; j < len(nodes); j++ {
		nodes[j] = leafNodes[leafCount-trailingCount+j]
	}

	return nodes, true
}

func (a *Analyzer) keywordArgSymbols(fn query.Signature, args callWithArguments) []query.Symbol {
	symbols := []query.Symbol{}
	for i, param := range fn.Params {
		if i < int(args.positional) {
			continue
		}
		kwarg := param.Name
		if used := args.keywords[kwarg]; !used {
			symbols = append(symbols, query.Symbol{
				Name:   kwarg + "=",
				Detail: param.Content,
				Kind:   protocol.SymbolKindVariable,
			})
		}
	}
	return symbols
}

// Find the object part of an attribute expression that has a dot '.' immediately before the given point.
func (a *Analyzer) findAttrObjectExpression(nodes []*sitter.Node, pt sitter.Point) *sitter.Node {
	if pt.Column == 0 {
		return nil
	}

	var dot *sitter.Node
	searchRange := sitter.Range{StartPoint: sitter.Point{Row: pt.Row, Column: pt.Column - 1}, EndPoint: pt}
	var parentNode *sitter.Node
	for i := len(nodes) - 1; i >= 0; i-- {
		parentNode = nodes[i]
		dot = query.FindChildNode(parentNode, func(n *sitter.Node) int {
			if query.PointBeforeOrEqual(n.EndPoint(), searchRange.StartPoint) {
				return -1
			}
			if n.StartPoint() == searchRange.StartPoint &&
				n.EndPoint() == searchRange.EndPoint &&
				n.Type() == "." {
				return 0
			}
			if query.PointBeforeOrEqual(n.StartPoint(), searchRange.StartPoint) &&
				query.PointAfterOrEqual(n.EndPoint(), searchRange.EndPoint) {
				return 0
			}
			return 1
		})
		if dot != nil {
			break
		}
	}

	// ak: FIXME: add parent as well?
	if dot == nil {
		// FIXME? should we check the range?
		n := nodes[len(nodes)-1]
		if n.Type() == "." {
			dot = n
			parentNode = n
		}

	}

	if dot != nil {
		expr := parentNode.PrevSibling()
		for n := dot; n != parentNode; n = n.Parent() {
			if n.PrevSibling() != nil {
				expr = n.PrevSibling()
				break
			}
		}

		if expr != nil {
			a.logger.Debug("dot completion",
				zap.String("dot", dot.String()),
				zap.String("expr", expr.String()))
			return expr
		}
	}
	return nil
}

// Perform some rudimentary type analysis to determine the Starlark type of the node
func (a *Analyzer) analyzeType(doc document.Document, node *sitter.Node) string {
	if node == nil {
		return ""
	}
	nodeT := node.Type()
	switch nodeT {
	case query.NodeTypeString, query.NodeTypeDictionary, query.NodeTypeList:
		return query.SymbolKindToBuiltinType(query.StrToSymbolKind(nodeT))

	case query.NodeTypeIdentifier:
		if sym, found := a.FindDefinition(doc, node, doc.Content(node)); found {
			return sym.GetType()
		}

	case query.NodeTypeCall:
		fnName := doc.Content(node.ChildByFieldName("function"))
		args := node.ChildByFieldName("arguments")
		sig, found := a.signatureInformation(doc, node, callWithArguments{fnName: fnName, argsNode: args})
		if found && sig.ReturnType != "" {
			if _, t := query.StrToSymbolKindAndType(sig.ReturnType); t != "" {
				return t
			}
			return sig.ReturnType
		}
	}
	return ""
}

func (a *Analyzer) resolveSymbolTypeAndKind(doc document.Document, sym query.Symbol, symbols []query.Symbol) query.Symbol {
	symbolsToResolve := []query.Symbol{}

	for i := 0; i < len(symbols); i++ { // limit just to be sure  // FIXME: map? to avoid infinite loops? instead of counter
		symbolsToResolve = append(symbolsToResolve, sym)

		if strings.HasSuffix(sym.Type, "()") { // function assignment
			break
		}
		if _, knownKind := query.KnownKinds[sym.Kind]; knownKind {
			break
		}

		for _, s := range symbols {
			if s.Name == sym.Type {
				sym = s
				break
			}
		}
	}

	if len(symbolsToResolve) == 0 {
		return sym
	}

	sourceSymbol := symbolsToResolve[len(symbolsToResolve)-1]

	// FIXME? see findDefinitionType
	if _, knownKind := query.KnownKinds[sourceSymbol.Kind]; knownKind {
		return sourceSymbol
	}
	if strings.HasSuffix(sourceSymbol.Type, "()") { // function assignment
		// FIXME: maybe to pass node? and not symbols? signatureINfo will search first for and local scope functions
		// FIXME: split signatureInfo to 2 methods. buitins and all other
		// FIXME: move seacrh for name + funcname before searching in typeMethods

		funcName := sourceSymbol.Type[:len(sourceSymbol.Type)-2] // remove '()'
		sig, found := a.signatureInformation(doc, doc.Tree().RootNode(), callWithArguments{fnName: funcName})
		if found && sig.ReturnType != "" {
			sourceSymbol.Type = sig.ReturnType
			sourceSymbol.Kind = protocol.SymbolKindStruct
		}
		return sourceSymbol

	}
	return sym
}

func (a *Analyzer) availableMembersForType(t string) []query.Symbol {
	if t != "" {
		if class, found := a.builtins.Types[t]; found {
			return class.Members
		}
		switch t {
		case "None", "bool", "int", "float":
			return []query.Symbol{}
		}
	}
	return a.builtins.Members
}

func (a *Analyzer) availableMembersForNode(doc document.Document, node *sitter.Node) []query.Symbol {
	t := a.analyzeType(doc, node)
	return a.availableMembersForType(t)
}

func (a *Analyzer) FindDefinition(doc document.Document, node *sitter.Node, name string) (query.Symbol, bool) {
	for _, sym := range query.SymbolsInScope(doc, node) {
		if sym.Name == name {
			return sym, true
		}
	}
	for _, sym := range doc.Symbols() {
		if sym.Name == name {
			return sym, true
		}
	}
	for _, sym := range a.builtins.Symbols {
		if sym.Name == name {
			return sym, true
		}
	}
	return query.Symbol{}, false
}

func keywordArgContext(doc document.Document, node *sitter.Node, pt sitter.Point) callWithArguments {
	if node.Type() == "=" ||
		query.HasAncestor(node, func(anc *sitter.Node) bool {
			return anc.Type() == query.NodeTypeKeywordArgument
		}) {
		return callWithArguments{}
	}
	return possibleCallInfo(doc, node, pt)
}
