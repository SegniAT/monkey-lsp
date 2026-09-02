package analysis_test

import (
	"testing"

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
