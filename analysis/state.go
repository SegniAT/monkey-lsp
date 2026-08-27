package analysis

import (
	"fmt"

	"github.com/SegniAT/monkey-language-interpreter/ast"
	"github.com/SegniAT/monkey-lsp/lsp"
)

type State struct {
	Documents map[string]*Document
}

type Document struct {
	Version int
	URI     string
	Content string
	AST     *ast.Program
}

func NewState() *State {
	return &State{Documents: map[string]*Document{}}
}

func (s *State) DidOpen(params lsp.DidOpenTextDocumentParams) []lsp.Diagnostic {
	uri := params.TextDocument.URI
	text := params.TextDocument.Text

	doc := &Document{
		URI:     uri,
		Content: text,
		//AST:     ast.parse(text),
	}

	s.Documents[uri] = doc

	return nil
}

func (s *State) DidChange(params lsp.DidChangeTextDocumentParams) []lsp.Diagnostic {
	doc, ok := s.Documents[params.TextDocument.URI]
	if !ok {
		return nil
	}

	if len(params.ContentChanges) == 0 {
		return nil
	}

	change := params.ContentChanges[len(params.ContentChanges)-1]

	doc.Version = params.TextDocument.Version
	doc.Content = change.Text

	// Parse the new content.
	// doc.AST = parser.Parse(doc.Content)
	return nil
}

// TODO: just bad, don't couple them if possible
func (s *State) Hover(request lsp.HoverRequest) lsp.HoverResponse {
	document := s.Documents[request.Params.TextDocument.URI]

	return lsp.HoverResponse{
		Response: lsp.Response{
			Message: lsp.Message{
				JSONRPC: "2.0",
			},
			ID: &request.ID,
		},
		Result: lsp.Hover{
			Contents: lsp.MarkupContent{
				Kind: lsp.MarkupMarkdown,
				Value: fmt.Sprintf(`# Sup nigga
					URI: %s
					Length: %d
					==faf==
					`, document.URI, len(document.Content)),
			},
			Range: &lsp.Range{Start: request.Params.Position, End: request.Params.Position},
		},
	}
}

func (s *State) Definition(request lsp.DefinitionRequest) lsp.DefinitionResponse {
	return lsp.DefinitionResponse{
		Response: lsp.Response{
			Message: lsp.Message{
				JSONRPC: "2.0",
			},
			ID: &request.ID,
		},
		Result: lsp.Location{
			URI: request.Params.TextDocument.URI,
			Range: lsp.Range{
				Start: lsp.Position{Line: request.Params.Position.Line - 1, Character: request.Params.Position.Character - 1},
				End:   lsp.Position{Line: request.Params.Position.Line - 1, Character: request.Params.Position.Character - 1},
			},
		},
	}
}

func (s *State) Completion(request lsp.CompletionRequest) lsp.CompletionResponse {
	return lsp.CompletionResponse{
		Response: lsp.Response{
			Message: lsp.Message{
				JSONRPC: "2.0",
			},
			ID: &request.ID,
		},
		Result: []lsp.CompletionItem{
			{
				Label:  "say what",
				Detail: "say whaaat",
				Documentation: lsp.MarkupContent{
					Kind: lsp.MarkupMarkdown,
					Value: `# brotha
					## how how how
					`,
				},
			},
		},
	}
}
