package analysis

import (
	"strings"
	"testing"

	"github.com/SegniAT/monkey-language-interpreter/token"
)

func didOpenHelper(t *testing.T, content string) []token.Diagnostic {
	t.Helper()

	state := NewState()
	return state.DidOpen(3, "file:///tmp/test.monkey", content)
}

func documentHelper(t *testing.T, content string) *Document {
	t.Helper()

	state := NewState()
	state.DidOpen(3, "file:///tmp/test.monkey", content)
	return state.Documents["file:///tmp/test.monkey"]
}

func TestUndefinedVariable(t *testing.T) {
	tests := map[string]struct {
		content      string
		wantCount    int
		wantContains string
	}{
		"valid program": {
			content: "let x = 5;\nlet y = x + 1;\ny;", wantCount: 0,
		},
		"undefined variable": {
			content: "let x = foo;\nx;", wantCount: 1, wantContains: "undefined variable: foo",
		},
		"builtin allowed": {
			content: "len(\"monkey\")", wantCount: 0,
		},
		"params and call": {
			content: "let add = fn(a, b) { a + b };\nadd(1, 2);", wantCount: 0,
		},
		"undefined inside function": {
			content: "let add = fn(a) { a + mystery };\nadd(1);", wantCount: 1, wantContains: "undefined variable: mystery",
		},
		"param shadows global": {
			content: "let x = 5;\nlet f = fn(x) { x };\nf(x);", wantCount: 0,
		},
		"inner scope not leaked": {
			content: "let f = fn() { let inner = 1; };\nlet x = inner;\nx;", wantCount: 3, wantContains: "undefined variable: inner",
		},
		"array index": {
			content: "let a = [1, 2];\nlet i = a[0] + bug;\ni;", wantCount: 1, wantContains: "undefined variable: bug",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			diags := didOpenHelper(t, tt.content)

			if len(diags) != tt.wantCount {
				t.Fatalf("expected %d diagnostics, got %d: %v", tt.wantCount, len(diags), diags)
			}

			if tt.wantContains != "" && !strings.Contains(diags[0].Message, tt.wantContains) {
				t.Errorf("expected diagnostic to contain %q, got %q", tt.wantContains, diags[0].Message)
			}
		})
	}
}

func TestRedeclaration(t *testing.T) {
	tests := map[string]struct {
		content      string
		wantCount    int
		wantContains string
	}{
		"same scope": {
			content: "let x = 1;\nlet x = 2;", wantCount: 1, wantContains: "redeclaration of x",
		},
		"shadowing let in function body": {
			content: "let f = fn(a) { let a = 5; };\nf(1);", wantCount: 1, wantContains: "redeclaration of a",
		},
		"duplicate params": {
			content: "let f = fn(a, a) { a; };\nf(1, 2);", wantCount: 1, wantContains: "redeclaration of a",
		},
		"builtin may be shadowed": {
			content: "let len = 5;\nlen;", wantCount: 0,
		},
		"outer scope is not redeclared": {
			content: "let x = 5;\nlet f = fn(x) { x };\nf(x);", wantCount: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			diags := didOpenHelper(t, tt.content)

			if len(diags) != tt.wantCount {
				t.Fatalf("expected %d diagnostics, got %d: %v", tt.wantCount, len(diags), diags)
			}

			if tt.wantContains != "" && !strings.Contains(diags[0].Message, tt.wantContains) {
				t.Errorf("expected diagnostic to contain %q, got %q", tt.wantContains, diags[0].Message)
			}
		})
	}
}

func TestUnusedVariable(t *testing.T) {
	tests := map[string]struct {
		content      string
		wantCount    int
		wantContains string
	}{
		"unused local": {
			content: "let x = 1;", wantCount: 1, wantContains: "unused variable: x",
		},
		"mixed": {
			content: "let x = 1;\nlet y = 2;\ny;", wantCount: 1, wantContains: "unused variable: x",
		},
		"unused function binding": {
			content: "let f = fn() { 1; };", wantCount: 1, wantContains: "unused variable: f",
		},
		"local in function body": {
			content: "let f = fn() { let inner = 1; };\nf();", wantCount: 1, wantContains: "unused variable: inner",
		},
		"unused parameter": {
			content: "let f = fn(a) { 5; };\nf(1);", wantCount: 1, wantContains: "unused parameter: a",
		},
		"used is not reported": {
			content: "let x = 1;\nx;", wantCount: 0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			diags := didOpenHelper(t, tt.content)

			if len(diags) != tt.wantCount {
				t.Fatalf("expected %d diagnostics, got %d: %v", tt.wantCount, len(diags), diags)
			}

			if tt.wantContains != "" && !strings.Contains(diags[0].Message, tt.wantContains) {
				t.Errorf("expected diagnostic to contain %q, got %q", tt.wantContains, diags[0].Message)
			}
		})
	}
}

func TestUsesResolved(t *testing.T) {
	t.Run("resolutions recorded per use site", func(t *testing.T) {
		doc := documentHelper(t, "let x = 5;\nlet y = x;\ny;")

		if len(doc.identifierUses) != 2 {
			t.Fatalf("expected 2 uses, got %d", len(doc.identifierUses))
		}

		for useNode, sym := range doc.identifierUses {
			if sym.Name != useNode.Value {
				t.Errorf("use of %q resolved to %q", useNode.Value, sym.Name)
			}
			if sym.Type != variable {
				t.Errorf("use of %q has type %v, want Variable", useNode.Value, sym.Type)
			}
		}
	})

	t.Run("builtins resolve", func(t *testing.T) {
		doc := documentHelper(t, "len(\"hi\");")

		if len(doc.identifierUses) != 1 {
			t.Fatalf("expected 1 use, got %d", len(doc.identifierUses))
		}

		for useNode, sym := range doc.identifierUses {
			if useNode.Value != "len" {
				t.Errorf("unexpected use %v", useNode)
			}
			if sym.Type != builtin {
				t.Errorf("len has type %v, want Builtin", sym.Type)
			}
		}
	})

	t.Run("undefined identifiers absent", func(t *testing.T) {
		doc := documentHelper(t, "mystery;")

		if len(doc.identifierUses) != 0 {
			t.Fatalf("expected no uses, got %d", len(doc.identifierUses))
		}
	})
}

func TestFunctionScopesRecorded(t *testing.T) {
	t.Run("function literals get their own scope", func(t *testing.T) {
		doc := documentHelper(t, "let f = fn(a) { a };")

		if len(doc.functionScopes) != 1 {
			t.Fatalf("expected 1 scope, got %d", len(doc.functionScopes))
		}

		for fnNode := range doc.functionScopes {
			if len(fnNode.Parameters) != 1 {
				t.Errorf("unexpected function literal parameters")
			}
		}
	})

	t.Run("nested functions each get a scope", func(t *testing.T) {
		doc := documentHelper(t, "let f = fn() { let g = fn() { 1; }; };")

		if len(doc.functionScopes) != 2 {
			t.Fatalf("expected 2 scopes, got %d", len(doc.functionScopes))
		}
	})
}
