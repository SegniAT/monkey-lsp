package analysis

import (
	"fmt"

	"github.com/SegniAT/monkey-language-interpreter/ast"
	"github.com/SegniAT/monkey-language-interpreter/lexer"
	"github.com/SegniAT/monkey-language-interpreter/parser"
	"github.com/SegniAT/monkey-language-interpreter/token"
)

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

func (d *Document) Parse() {
	if d == nil {
		return
	}

	d.parser = parser.New(lexer.New(d.Content))
	d.AST = d.parser.ParseProgram()
}

type State struct {
	Documents map[string]*Document
}

func NewState() *State {
	return &State{Documents: map[string]*Document{}}
}

func (s *State) DidOpen(version int, uri, text string) []token.Diagnostic {
	doc := &Document{Version: version, URI: uri, Content: text}
	s.Documents[uri] = doc
	doc.Parse()
	return append(doc.parser.Diagnostics(), doc.analyze()...)
}

// contentChange is the full content of the text as specified in our server capabilities
func (s *State) DidChange(version int, uri string, contentChange string) []token.Diagnostic {
	doc, ok := s.Documents[uri]
	if !ok {
		return nil
	}

	if contentChange == "" {
		return nil
	}

	doc.Version = version
	doc.Content = contentChange
	doc.Parse()
	return append(doc.parser.Diagnostics(), doc.analyze()...)
}

func (s *State) Hover(uri string, line, character uint) *Hover {
	document := s.Documents[uri]
	if document == nil {
		return nil
	}

	return &Hover{
		Contents: MarkupContent{
			Kind: MarkupMarkdown,
			Value: fmt.Sprintf(`# Sup nigga
				URI: %s
				Length: %d
				==faf==
				`, document.URI, len(document.Content)),
		},
		Range: &Range{
			Start: Position{Line: line, Character: character},
			End:   Position{Line: line, Character: character},
		},
	}
}

func (s *State) Definition(uri string, line, character uint) *Location {
	return &Location{
		URI: uri,
		Range: Range{
			Start: Position{
				Line:      line,
				Character: character,
			},
			End: Position{
				Line:      line,
				Character: character,
			},
		},
	}
}

func (s *State) Completion(uri string, line, character uint) []CompletionItem {
	return []CompletionItem{
		{
			Label:  "say what",
			Detail: "say whaaat",
			Documentation: MarkupContent{
				Kind: MarkupMarkdown,
				Value: `# brotha
			## how how how
			`,
			},
		},
	}
}
