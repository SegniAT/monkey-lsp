package analysis

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/SegniAT/monkey-language-interpreter/ast"
	"github.com/SegniAT/monkey-language-interpreter/lexer"
	"github.com/SegniAT/monkey-language-interpreter/parser"
	"github.com/SegniAT/monkey-language-interpreter/token"
)

var builtinDocs = map[string]string{
	"len":   "Calculates and returns the length of a string or an array.\n\n**Returns:** `INTEGER`\n\n**Example:**\n```monkey\nlen(\"monkey\"); // 6\nlen([1, 2, 3]); // 3\n```",
	"first": "Returns the first element of an array. If the array is empty, returns `NULL`.\n\n**Returns:** `ANY`\n\n**Example:**\n```monkey\nlet arr = [1, 2, 3];\nfirst(arr); // 1\n```",
	"last":  "Returns the last element of an array. If the array is empty, returns `NULL`.\n\n**Returns:** `ANY`\n\n**Example:**\n```monkey\nlet arr = [1, 2, 3];\nlast(arr); // 3\n```",
	"rest":  "Returns a new array containing all elements of the given array except the first one. If the array is empty, returns `NULL`.\n\n**Returns:** `ARRAY`\n\n**Example:**\n```monkey\nlet arr = [1, 2, 3];\nrest(arr); // [2, 3]\n```",
	"push":  "Adds a new element to the end of an array and returns the newly allocated array.\n\n**Returns:** `ARRAY`\n\n**Example:**\n```monkey\nlet arr = [1, 2];\npush(arr, 3); // [1, 2, 3]\n```",
	"puts":  "Prints the given arguments to the standard output as strings.\n\n**Returns:** `NULL`\n\n**Example:**\n```monkey\nputs(\"Hello, \", \"World!\");\n```",
}

type Document struct {
	Version int
	URI     string
	Content string
	AST     *ast.Program
	parser  *parser.Parser

	rootScope      *symbolTable
	identifierUses map[*ast.Identifier]*symbol
	functionScopes map[*ast.FunctionLiteral]*symbolTable
}

func (d *Document) parse() {
	if d == nil {
		return
	}

	d.parser = parser.New(lexer.New(d.Content))
	d.AST = d.parser.ParseProgram()
}

type State struct {
	mu        sync.RWMutex
	Documents map[string]*Document
}

func NewState() *State {
	return &State{Documents: map[string]*Document{}}
}

func (s *State) DidOpen(version int, uri, text string) []token.Diagnostic {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc := &Document{Version: version, URI: uri, Content: text}
	s.Documents[uri] = doc
	doc.parse()
	return append(doc.parser.Diagnostics(), doc.analyze()...)
}

// contentChange is the full content of the text as specified in our server capabilities
func (s *State) DidChange(version int, uri string, contentChange string) []token.Diagnostic {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, ok := s.Documents[uri]
	if !ok {
		return nil
	}

	if contentChange == "" {
		return nil
	}

	doc.Version = version
	doc.Content = contentChange
	doc.parse()
	return append(doc.parser.Diagnostics(), doc.analyze()...)
}

// line and character are 0 based
func (s *State) Hover(uri string, line, character uint) *Hover {
	s.mu.RLock()
	defer s.mu.RUnlock()

	document := s.Documents[uri]
	if document == nil {
		return nil
	}

	node := FindASTNode(document.AST, line, character)
	if node == nil {
		return nil
	}

	var markup string

	switch node := node.(type) {
	case *ast.Identifier:
		identSymbol, ok := document.identifierUses[node]
		if !ok {
			return nil
		}

		var kind string
		switch identSymbol.Type {
		case variable:
			kind = "Variable"
		case parameter:
			kind = "Parameter"
		case builtin:
			kind = "Built-in Function"
		}

		if identSymbol.Type == builtin {
			desc := builtinDocs[identSymbol.Name]
			markup = fmt.Sprintf("```monkey\nBuilt-in Function\n```\n---\n**%s**\n\n%s", identSymbol.Name, desc)
		} else {
			markup = fmt.Sprintf("**%s**\n---\n```monkey\n%s\n```\n**Defined at:** Line `%d`, Col `%d`",
				kind,
				identSymbol.Name,
				identSymbol.Range.Start.Line,
				identSymbol.Range.Start.Character)
		}
	case *ast.IntegerLiteral:
		value := node.Value
		markup = fmt.Sprintf("```monkey\n%d\n```\n---\n**Hex:**\t`0x%X`  \n**Octal:**\t`%O`  \n**Binary:**\t`0b%b`",
			value, value, value, value)

	case *ast.Boolean:
		value := node.Value
		markup = fmt.Sprintf("```monkey\n%t\n```\n---\n**Type:** BOOLEAN", value) // Value is already at the top, so we just state the type below

	case *ast.StringLiteral:
		value := node.Value
		markup = fmt.Sprintf("```monkey\n\"%s\"\n```\n---\n**Length:** %d characters  \n**Size:** %d bytes",
			value, utf8.RuneCountInString(value), len(value))

	default:
		slog.Warn("Unknown node type",
			slog.String("node type", fmt.Sprintf("%T", node)))
		return nil
	}

	return &Hover{
		Contents: MarkupContent{
			Kind:  MarkupMarkdown,
			Value: markup,
		},
		Range: &Range{Position(node.Start()), Position(node.End())},
	}
}

func (s *State) Definition(uri string, line, character uint) *Location {
	s.mu.RLock()
	defer s.mu.RUnlock()

	document := s.Documents[uri]
	if document == nil {
		return nil
	}

	node := FindASTNode(document.AST, line, character)
	if node == nil {
		return nil
	}

	identifier, ok := node.(*ast.Identifier)
	if !ok {
		return nil
	}

	symbol, ok := document.identifierUses[identifier]
	if !ok {
		return nil
	}

	return &Location{
		URI:   uri,
		Range: symbol.Range,
	}
}

func (s *State) Completion(uri string, line, character uint) []CompletionItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	document := s.Documents[uri]
	if document == nil {
		return nil
	}

	// character will land on an empty space, so we should go one character to the left
	node := FindASTNode(document.AST, line, character-1)
	if node == nil {
		return nil
	}

	// Find the deepest function containing the node.
	var deepestFuncLiteral *ast.FunctionLiteral
	for funcLiteral := range document.functionScopes {
		if !nodeEncloses(funcLiteral, node) {
			continue
		}

		if deepestFuncLiteral == nil {
			deepestFuncLiteral = funcLiteral
			continue
		}

		if nodeEncloses(deepestFuncLiteral, funcLiteral) {
			deepestFuncLiteral = funcLiteral
		}
	}

	var deepestSymbolTable *symbolTable
	if deepestFuncLiteral != nil {
		deepestSymbolTable = document.functionScopes[deepestFuncLiteral]
	} else {
		deepestSymbolTable = document.rootScope
	}

	completionItems := []CompletionItem{}
	nodeStr := node.TokenLiteral()
	walker := deepestSymbolTable
	for walker != nil {
		for _, symbol := range walker.Symbols {
			if !strings.HasPrefix(symbol.Name, nodeStr) {
				continue
			}

			completionItems = append(completionItems, CompletionItem{
				Label:  symbol.Name,
				Detail: symbol.Name,
				Kind: func() CompletionItemKind {
					switch symbol.Type {
					case builtin:
						return Function
					case variable, parameter:
						return Variable
					default:
						panic("unknown symbol type")
					}
				}(),
				Documentation: MarkupContent{},
			})
		}

		walker = walker.Outer
	}

	// keywords
	for keyword := range token.Keywords {
		if !strings.HasPrefix(keyword, nodeStr) {
			continue
		}

		completionItems = append(completionItems, CompletionItem{
			Label:         keyword,
			Detail:        keyword,
			Kind:          Keyword,
			Documentation: MarkupContent{},
		})
	}

	return completionItems
}

// nodeEncloses helps us find if a node is enclosed in another node
func nodeEncloses(outer, inner ast.Node) bool {
	if outer == nil || inner == nil {
		return false
	}

	outerStart, outerEnd := outer.Start(), outer.End()
	innerStart, innerEnd := inner.Start(), inner.End()

	// outer node starts after the inner
	if outerStart.Line > innerStart.Line {
		return false
	}

	// outer node ends before the inner
	if outerEnd.Line < innerEnd.Line {
		return false
	}

	// if both start on the same line but outer's character comes AFTER inner's
	if outerStart.Line == innerStart.Line && outerStart.Character > innerStart.Character {
		return false
	}

	// if both end on the same line but outer's character comes BEFORE inner's
	if outerEnd.Line == innerEnd.Line && outerEnd.Character < innerEnd.Character {
		return false
	}

	return true
}

// positionIsInRange helps us figure out if a position is in a node.
func positionIsInRange(node ast.Node, line, character uint) bool {
	if node == nil {
		return false
	}

	start, end := node.Start(), node.End()

	// Outside the upper and bottom line
	if line < start.Line || line > end.Line {
		return false
	}
	// On the starting line, but before the starting character
	if line == start.Line && character < start.Character {
		return false
	}
	// On the ending line, but after the ending character
	if line == end.Line && character > end.Character {
		return false
	}

	return true
}

/*
FindASTNode finds AST node at given line and character, if one exists.
It returns either of these leaf nodes or nil: *ast.Identifier, *ast.IntegerLiteral, *ast.Boolean or *ast.StringLiteral
*/
func FindASTNode(node ast.Node, line, character uint) ast.Node {
	if node == nil {
		return nil
	}

	switch node := node.(type) {
	case *ast.Program:
		for _, stmt := range node.Statements {
			if n := FindASTNode(stmt, line, character); n != nil {
				return n
			}
		}

	case *ast.BlockStatement:
		// We keep the bounds check here because blocks can be massive,
		// and we know BlockStatement has a precise EndToken from the parser.
		if !positionIsInRange(node, line, character) {
			return nil
		}
		for _, stmt := range node.Statements {
			if n := FindASTNode(stmt, line, character); n != nil {
				return n
			}
		}

	case *ast.LetStatement:
		if n := FindASTNode(node.Name, line, character); n != nil {
			return n
		}
		return FindASTNode(node.Value, line, character)

	case *ast.FunctionLiteral:
		for _, param := range node.Parameters {
			if n := FindASTNode(param, line, character); n != nil {
				return n
			}
		}
		return FindASTNode(node.Body, line, character)

	case *ast.ExpressionStatement:
		return FindASTNode(node.Expression, line, character)

	case *ast.ReturnStatement:
		return FindASTNode(node.ReturnValue, line, character)

	case *ast.IfExpression:
		if n := FindASTNode(node.Condition, line, character); n != nil {
			return n
		}
		if n := FindASTNode(node.Consequence, line, character); n != nil {
			return n
		}
		return FindASTNode(node.Alternative, line, character)

	case *ast.PrefixExpression:
		return FindASTNode(node.Right, line, character)

	case *ast.InfixExpression:
		if n := FindASTNode(node.Left, line, character); n != nil {
			return n
		}
		return FindASTNode(node.Right, line, character)

	case *ast.CallExpression:
		if n := FindASTNode(node.Function, line, character); n != nil {
			return n
		}
		for _, arg := range node.Arguments {
			if n := FindASTNode(arg, line, character); n != nil {
				return n
			}
		}

	case *ast.ArrayLiteral:
		for _, el := range node.Elements {
			if n := FindASTNode(el, line, character); n != nil {
				return n
			}
		}

	case *ast.IndexExpression:
		if n := FindASTNode(node.Left, line, character); n != nil {
			return n
		}
		return FindASTNode(node.Index, line, character)

	case *ast.HashLiteral:
		for keyExpr, valExpr := range node.Pairs {
			if n := FindASTNode(keyExpr, line, character); n != nil {
				return n
			}
			if n := FindASTNode(valExpr, line, character); n != nil {
				return n
			}
		}

	default:
		// For leaf nodes (Identifier, IntegerLiteral, Boolean, StringLiteral),
		// we perform the strict spatial boundary check to confirm the cursor is exactly here.
		if positionIsInRange(node, line, character) {
			return node
		}
	}

	return nil
}

func (s *State) Close(uri string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Documents, uri)
}
