package analysis_test

import (
	"strings"
	"testing"

	"github.com/SegniAT/monkey-language-interpreter/token"
	"github.com/SegniAT/monkey-lsp/analysis"
)

func TestDidOpen(t *testing.T) {
	state := analysis.NewState()
	version, uri, content := 3, "/home/jcarmack/projects/demo.monkey", "let game = \"cool\""
	state.DidOpen(version, uri, content)

	document, ok := state.Documents[uri]
	if !ok {
		t.Fatalf("Document with uri '%s' not saved correctly", uri)
	}

	if document.Version != version {
		t.Errorf("expected document version %d, got %d", version, document.Version)
	}

	if document.URI != uri {
		t.Errorf("expected document uri %s, got %s", uri, document.URI)
	}

	if document.Content != content {
		t.Errorf("expected document content %s, got %s", content, document.Content)
	}
}

func TestDidChange(t *testing.T) {
	state := analysis.NewState()
	uri := "/home/jcarmack/projects/demo.monkey"
	state.DidOpen(3, uri, "let game = \"cool\"")

	nuVersion, nuContent := 4, "let game = \"cool finished game\""
	state.DidChange(nuVersion, uri, nuContent)

	document, ok := state.Documents[uri]
	if !ok {
		t.Fatalf("Document with uri '%s' not saved correctly", uri)
	}

	if document.Version != nuVersion {
		t.Errorf("expected document version %d, got %d", nuVersion, document.Version)
	}

	if document.URI != uri {
		t.Errorf("expected document uri %s, got %s", uri, document.URI)
	}

	if document.Content != nuContent {
		t.Errorf("expected document content %s, got %s", nuContent, document.Content)
	}
}

func documentHelper(t *testing.T, content string) *analysis.Document {
	t.Helper()

	state := analysis.NewState()
	state.DidOpen(3, "file:///tmp/test.monkey", content)
	return state.Documents["file:///tmp/test.monkey"]
}

func TestFindASTNode(t *testing.T) {
	content := `let n = 100;
let adder = fn(num){
    return num + 1;
}

if(adder(10)==10){
	puts("true")
}else{
	let x = {
		1: "one",
		2: true,
    }
    puts("false")
}
`
	doc := documentHelper(t, content)

	tests := map[string]struct {
		line                uint
		character           uint
		nodeExists          bool
		expectedRange       token.Range
		expectedStringValue string
	}{
		"identifier": {
			line:       1,
			character:  5, // 'n'
			nodeExists: true,
			expectedRange: token.Range{
				Start: token.Position{Line: 1, Character: 5},
				End:   token.Position{Line: 1, Character: 5},
			},
			expectedStringValue: "n",
		},
		"integer literal": {
			line:       1,
			character:  9, // '1'
			nodeExists: true,
			expectedRange: token.Range{
				Start: token.Position{Line: 1, Character: 9},
				End:   token.Position{Line: 1, Character: 11},
			},
			expectedStringValue: "100",
		},
		"string literal": {
			line:       7,
			character:  9, // inside "true" string, on 'r'
			nodeExists: true,
			expectedRange: token.Range{
				Start: token.Position{Line: 7, Character: 7},
				End:   token.Position{Line: 7, Character: 12},
			},
			expectedStringValue: "true",
		},
		"empty space": {
			line:       5,
			character:  1,
			nodeExists: false,
		},
		"symbol (operator)": {
			line:       6,
			character:  14, // on "=="
			nodeExists: false,
		},
		"boolean": {
			line:       11,
			character:  8, // on true, "u" character
			nodeExists: true,
			expectedRange: token.Range{
				Start: token.Position{Line: 11, Character: 6},
				End:   token.Position{Line: 11, Character: 9},
			},
			expectedStringValue: "true",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			node := analysis.FindASTNode(doc.AST, test.line, test.character)

			if node == nil && test.nodeExists {
				t.Fatal("Node expected to be found, but not found")
			}

			if node != nil && !test.nodeExists {
				t.Fatal("Node expected to not be found, but found")
			}

			if !test.nodeExists {
				return
			}

			if node.TokenLiteral() != test.expectedStringValue {
				t.Fatalf("Expected value %s, got %s", test.expectedStringValue, node.String())
			}

			gotRange := token.Range{Start: node.Start(), End: node.End()}
			if gotRange != test.expectedRange {
				t.Errorf("Expected range %v, got %v", test.expectedRange, gotRange)
			}
		})
	}
}

func TestHover(t *testing.T) {
	content := `let num = 42;
let myFunc = fn(param) {
	let str = "hello";
	let b = true;
	puts(str);
	return param + num;
};`

	state := analysis.NewState()
	uri := "file:///tmp/hover_test.monkey"
	state.DidOpen(1, uri, content)

	tests := map[string]struct {
		uri          string
		line         uint
		character    uint
		wantNil      bool
		wantContains []string
	}{
		"unknown URI": {
			uri:     "file:///tmp/unknown.monkey",
			line:    0,
			wantNil: true,
		},
		"empty space": {
			uri:       uri,
			line:      1,
			character: 4,
			wantNil:   true,
		},
		"integer literal": {
			uri:       uri,
			line:      1,
			character: 11, // on '4' in 42
			wantContains: []string{
				"```monkey\n42\n```",
				"**Hex:**\t`0x2A`",
			},
		},
		"string literal": {
			uri:       uri,
			line:      3,
			character: 14, // on 'e' in "hello"
			wantContains: []string{
				"```monkey\n\"hello\"\n```",
				"**Length:** 5 characters",
				"**Size:** 5 bytes",
			},
		},
		"boolean literal": {
			uri:       uri,
			line:      4,
			character: 10, // on 't' in true
			wantContains: []string{
				"```monkey\ntrue\n```",
				"**Type:** BOOLEAN",
			},
		},
		"builtin function": {
			uri:       uri,
			line:      5,
			character: 2, // on 'p' in puts
			wantContains: []string{
				"```monkey\nBuilt-in Function\n```",
				"**puts**",
				"Prints the given arguments",
			},
		},
		"variable identifier": {
			uri:       uri,
			line:      5,
			character: 7, // on 's' in str inside puts(str)
			wantContains: []string{
				"Variable",
				"```monkey\nstr\n```",
			},
		},
		"parameter identifier": {
			uri:       uri,
			line:      6,
			character: 9, // on 'p' in param + num
			wantContains: []string{
				"Parameter",
				"```monkey\nparam\n```",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			hover := state.Hover(test.uri, test.line, test.character)

			if test.wantNil {
				if hover != nil {
					t.Fatalf("expected nil hover response, got %+v", hover)
				}
				return
			}

			if hover == nil {
				t.Fatalf("expected hover response, got nil")
			}

			if hover.Contents.Kind != analysis.MarkupMarkdown {
				t.Errorf("expected markup kind %q, got %q", analysis.MarkupMarkdown, hover.Contents.Kind)
			}

			for _, expectedStr := range test.wantContains {
				if !strings.Contains(hover.Contents.Value, expectedStr) {
					t.Errorf("markup missing expected content.\nExpected to contain:\n%s\n\nFull Markup Got:\n%s", expectedStr, hover.Contents.Value)
				}
			}
		})
	}
}

func TestCompletion(t *testing.T) {
	uri := "file:///tmp/completion_test.monkey"
	tests := map[string]struct {
		uri             string
		content         string
		line            uint
		character       uint
		completionItems []struct {
			label string
			kind  analysis.CompletionItemKind
		}
	}{
		"no suggestions": {
			uri:             uri,
			content:         `ak`,
			line:            1,
			character:       3, // past k
			completionItems: nil,
		},
		"builtin function and keyword suggestions": {
			uri:       uri,
			content:   `re`,
			line:      1,
			character: 3, // past e
			completionItems: []struct {
				label string
				kind  analysis.CompletionItemKind
			}{
				{label: "rest", kind: analysis.Function},
				{label: "return", kind: analysis.Keyword},
			},
		},
		"variable suggestion": {
			uri:       uri,
			content:   `let myVariable=3; myV`,
			line:      1,
			character: 22, // past V
			completionItems: []struct {
				label string
				kind  analysis.CompletionItemKind
			}{
				{label: "myVariable", kind: analysis.Variable},
			},
		},
		"parameter suggestion": {
			uri:       uri,
			content:   `fn(paramOne){ para }`,
			line:      1,
			character: 19, // past ra
			completionItems: []struct {
				label string
				kind  analysis.CompletionItemKind
			}{
				{label: "paramOne", kind: analysis.Variable},
			},
		},
		"access to parent scopes": {
			uri:       uri,
			content:   `let myVar = 1; fn(){let myVarOne = 2; fn(myVarTwo){ myVa }}`,
			line:      1,
			character: 57, // past Va
			completionItems: []struct {
				label string
				kind  analysis.CompletionItemKind
			}{
				{label: "myVar", kind: analysis.Variable},
				{label: "myVarOne", kind: analysis.Variable},
				{label: "myVarTwo", kind: analysis.Variable},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			state := analysis.NewState()
			state.DidOpen(1, test.uri, test.content)

			gotCompletionItems := state.Completion(test.uri, test.line, test.character)
			gotLen, expectedLen := len(gotCompletionItems), len(test.completionItems)
			if gotLen != expectedLen {
				t.Fatalf("Expected completion items %d, got %d", expectedLen, gotLen)
			}

			for _, expectedCompletionItem := range test.completionItems {
				found := false
				for _, gotCompletionItem := range gotCompletionItems {
					if expectedCompletionItem.label == gotCompletionItem.Label &&
						expectedCompletionItem.kind == gotCompletionItem.Kind {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("Expected competion item with label '%s' and kind '%s', not found", expectedCompletionItem.label, expectedCompletionItem.kind.String())
				}
			}
		})
	}
}
