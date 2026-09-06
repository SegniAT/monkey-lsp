package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync/atomic"

	"github.com/SegniAT/monkey-lsp/analysis"
	"github.com/SegniAT/monkey-lsp/lsp/protocol"
	"github.com/SegniAT/monkey-lsp/rpc"
)

type Server struct {
	initialized      atomic.Bool
	shutdownReceived atomic.Bool

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
	defer func() {
		if r := recover(); r != nil {
			slog.Error("recovered from panic in message handler",
				slog.Any("panic", r))
		}
	}()

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

		isRequest := msg.ID != nil

		// If we receive a request before the server is initialized, we should return error with -32002 code
		if isRequest && !s.initialized.Load() &&
			msg.Method != "exit" && msg.Method != "initialize" {
			_ = s.writeError(*msg.ID, &protocol.ResponseError{
				Code: protocol.ErrServerNotInitialized})
			continue
		}

		if isRequest {
			// Spin up a goroutine for read-only requests
			go func(reqID int64, reqMethod string, reqContent []byte) {
				defer func() {
					if r := recover(); r != nil {
						slog.Error("recovered from panic in request handler",
							slog.Any("panic", r),
							slog.String("method", reqMethod))
					}
				}()

				s.HandleRequest(reqID, reqMethod, reqContent)
			}(*msg.ID, msg.Method, content)
		} else {
			// Run notifications synchronously to guarantee state mutation order
			s.HandleNotification(msg.Method, content)
		}
	}
}

func (s *Server) HandleRequest(id int64, method string, content []byte) {
	switch method {
	case "initialize":
		s.handleInitialize(id, content)
	case "shutdown":
		s.handleShutdown()
	case "exit":
		s.handleExit()
	case "textDocument/hover":
		s.handleTextDocumentHover(id, content)
	case "textDocument/definition":
		s.handleTextDocumentDefinition(id, content)
	case "textDocument/completion":
		s.handleTextDocumentCompletion(id, content)
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
	case "textDocument/didClose":
		s.handleTextDocumentDidClose(content)
	case "exit":
		s.handleExit()
	default:
		slog.Warn("Unsupported notification", slog.String("method", method))
	}
}

func (s *Server) writeMessage(v any) error {
	content, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("Error marshalling message: %w", err)
	}

	_, err = s.writer.WriteMessage(content)
	if err != nil {
		return fmt.Errorf("Error writing message: %v", err)
	}

	return nil
}

func (s *Server) writeError(id int64, respErr *protocol.ResponseError) error {
	return s.writeMessage(protocol.Response{
		Message: protocol.Message{JSONRPC: "2.0"},
		ID:      &id,
		Error:   respErr,
	})
}
