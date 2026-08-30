package server

import (
	"encoding/json"
	"log/slog"

	"github.com/SegniAT/monkey-language-interpreter/token"
	"github.com/SegniAT/monkey-lsp/lsp/protocol"
)

func (s *Server) handleInitialized(_ json.RawMessage) {
	slog.Info("initialized")
}

func (s *Server) handleTextDocumentDidOpen(content json.RawMessage) {
	var notification protocol.DidOpenTextDocumentNotification
	if err := json.Unmarshal(content, &notification); err != nil {
		slog.Error("Error unmarshalling didOpen notification", slog.String("err", err.Error()))
		return
	}

	doc := notification.Params.TextDocument
	diagnostics := s.state.DidOpen(doc.Version, doc.URI, doc.Text)
	s.publishDiagnostics(doc.URI, doc.Version, diagnostics)

	slog.Info("document opened", slog.String("URI", doc.URI))
}

func (s *Server) handleTextDocumentDidChange(content json.RawMessage) {
	var notification protocol.DidChangeTextDocumentNotification
	if err := json.Unmarshal(content, &notification); err != nil {
		slog.Error("Error unmarshalling didChange notification", slog.String("err", err.Error()))
		return
	}

	doc := notification.Params.TextDocument
	for _, change := range notification.Params.ContentChanges {
		diagnostics := s.state.DidChange(doc.Version, doc.URI, change.Text)
		s.publishDiagnostics(doc.URI, doc.Version, diagnostics)
	}

	slog.Info("document changed", slog.String("URI", doc.URI))
}

func (s *Server) publishDiagnostics(uri string, version int, diagnostics []token.Diagnostic) {
	notif := protocol.PublishDiagnosticsNotification{
		Notification: protocol.Notification{
			Message: protocol.Message{JSONRPC: "2.0"},
			Method:  "textDocument/publishDiagnostics",
		},
		Params: protocol.PublishDiagnosticsParams{
			URI:     uri,
			Version: version,
			Diagnostics: func() []protocol.Diagnostic {
				if len(diagnostics) == 0 {
					return []protocol.Diagnostic{}
				}
				return toProtocolDiagnostics(diagnostics)
			}(),
		},
	}

	s.writeMessage(notif)
}

func toProtocolDiagnostics(diagnostics []token.Diagnostic) []protocol.Diagnostic {
	result := make([]protocol.Diagnostic, 0, len(diagnostics))
	for _, d := range diagnostics {
		result = append(result, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      d.Range.Start.Line - 1,
					Character: d.Range.Start.Character - 1,
				},
				End: protocol.Position{
					Line:      d.Range.End.Line - 1,
					Character: d.Range.End.Character - 1,
				},
			},
			Severity: int(d.Severity),
			Source:   "monkey",
			Message:  d.Message,
		})
	}
	return result
}
