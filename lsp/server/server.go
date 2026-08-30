package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/SegniAT/monkey-lsp/analysis"
	"github.com/SegniAT/monkey-lsp/lsp/protocol"
	"github.com/SegniAT/monkey-lsp/rpc"
)

type Server struct {
	state  *analysis.State
	reader *rpc.Reader
	writer *rpc.Writer
}

func NewServer(reader io.Reader, writer io.Writer) *Server {
	return &Server{
		state:  analysis.NewState(),
		reader: rpc.NewReader(reader),
		writer: rpc.NewWriter(writer),
	}
}

func (s *Server) Run() error {
	for {
		content, err := s.reader.ReadMessage()
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, rpc.ErrSeparatorNotFound) {
				slog.Info("Input closed, shutting down")
				return nil
			}
			slog.Error("Error reading content", slog.String("err", err.Error()))
			continue
		}

		var msg struct {
			ID     *int64 `json:"id,omitempty"`
			Method string `json:"method"`
		}

		err = json.Unmarshal(content, &msg)
		if err != nil {
			slog.Error("Error unmarshalling message", slog.String("err", err.Error()))
			continue
		}

		if msg.ID != nil {
			s.HandleRequest(*msg.ID, msg.Method, content)
		} else {
			s.HandleNotification(msg.Method, content)
		}
	}
}

func (s *Server) HandleRequest(id int64, method string, content []byte) {
	switch method {
	case "initialize":
		s.handleInitialize(id, content)
	// case "shutdown":
	// 	s.handleShutdown(id, content)
	case "textDocument/hover":
		s.handleTextDocumentHover(id, content)
	case "textDocument/definition":
		s.handleTextDocumentDefinition(id, content)
	case "textDocument/completion":
		s.handleTextDocumentCompletion(id, content)
	// case "textDocument/documentHighlight":
	default:
		s.writeError(id, &protocol.ResponseError{
			Code:    protocol.ErrMethodNotFound,
			Message: fmt.Sprintf("method not found: %s", method),
		})
	}
}

func (s *Server) HandleNotification(method string, content []byte) {
	switch method {
	case "initialized":
		s.handleInitialized(content)
	case "textDocument/didOpen":
		s.handleTextDocumentDidOpen(content)
	case "textDocument/didChange":
		s.handleTextDocumentDidChange(content)
	// case "textDocument/didClose":
	// 	s.handleTextDocumentDidClose(content)
	default:
		slog.Warn("Unsupported notification", slog.String("method", method))
	}
}

func (s *Server) writeMessage(v any) {
	content, err := json.Marshal(v)
	if err != nil {
		slog.Error("Error marshalling message", slog.String("err", err.Error()))
		return
	}

	_, err = s.writer.WriteMessage(content)
	if err != nil {
		slog.Error("Error writing message", slog.String("err", err.Error()))
	}
}

func (s *Server) writeError(id int64, respErr *protocol.ResponseError) {
	s.writeMessage(protocol.Response{
		Message: protocol.Message{JSONRPC: "2.0"},
		ID:      &id,
		Error:   respErr,
	})
}
