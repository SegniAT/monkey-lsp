package server

import (
	"encoding/json"
	"log/slog"

	"github.com/SegniAT/monkey-lsp/lsp/protocol"
)

func (s *Server) handleInitialize(id int64, content json.RawMessage) {
	var request protocol.InitializeRequest
	if err := json.Unmarshal(content, &request); err != nil {
		slog.Error("Error unmarshalling initialize request", slog.String("err", err.Error()))
		return
	}

	if request.Params.ClientInfo != nil {
		slog.Info("Connected to client",
			slog.String("name", request.Params.ClientInfo.Name),
			slog.String("version", request.Params.ClientInfo.Version),
		)
	}

	s.writeMessage(protocol.NewInitializeResponse(id))
}

func (s *Server) handleTextDocumentHover(id int64, content json.RawMessage) {
	var request protocol.HoverRequest
	if err := json.Unmarshal(content, &request); err != nil {
		slog.Error("Error unmarshalling hover request", slog.String("err", err.Error()))
		return
	}

	result := s.state.Hover(request.Params.TextDocument.URI, request.Params.Position.Line+1, request.Params.Position.Character+1)
	s.writeMessage(protocol.HoverResponse{
		Response: protocol.Response{
			Message: protocol.Message{JSONRPC: "2.0"},
			ID:      &id,
		},
		Result: toProtocolHover(result),
	})
}

func (s *Server) handleTextDocumentDefinition(id int64, content json.RawMessage) {
	var request protocol.DefinitionRequest
	if err := json.Unmarshal(content, &request); err != nil {
		slog.Error("Error unmarshalling definition request", slog.String("err", err.Error()))
		return
	}

	result := s.state.Definition(request.Params.TextDocument.URI, request.Params.Position.Line+1, request.Params.Position.Character+1)
	s.writeMessage(protocol.DefinitionResponse{
		Response: protocol.Response{
			Message: protocol.Message{JSONRPC: "2.0"},
			ID:      &id,
		},
		Result: toProtocolLocation(result),
	})
}

func (s *Server) handleTextDocumentCompletion(id int64, content json.RawMessage) {
	var request protocol.CompletionRequest
	if err := json.Unmarshal(content, &request); err != nil {
		slog.Error("Error unmarshalling completion request", slog.String("err", err.Error()))
		return
	}

	result := s.state.Completion(request.Params.TextDocument.URI, request.Params.Position.Line, request.Params.Position.Character)
	s.writeMessage(protocol.CompletionResponse{
		Response: protocol.Response{
			Message: protocol.Message{JSONRPC: "2.0"},
			ID:      &id,
		},
		Result: toProtocolCompletionItems(result),
	})
}
