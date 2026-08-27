package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/SegniAT/monkey-lsp/analysis"
	"github.com/SegniAT/monkey-lsp/lsp"
	"github.com/SegniAT/monkey-lsp/rpc"
)

func main() {
	file, err := os.OpenFile(filepath.Join(os.TempDir(), "monkey-lsp.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		panic(err)
	}

	a := bufio.Reader{}

	slog.SetDefault(slog.New(slog.NewTextHandler(file, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})))

	writer := os.Stdout

	state := analysis.NewState()

	slog.Info("LSP started")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(rpc.SplitFunc)

	for {
		scanned := scanner.Scan()
		if !scanned {
			// Check if the scanner stopped due to an error rather than just EOF
			if err := scanner.Err(); err != nil {
				slog.Error("Error reading input", slog.String("err", err.Error()))
			}

			return
		}

		method, content, err := rpc.DecodeMessage(scanner.Bytes())
		if err != nil {
			slog.Error("Error decoding message", slog.String("err", err.Error()))
			continue
		}

		handleMessage(writer, state, method, content)
	}
}

func handleMessage(writer io.Writer, state *analysis.State, method string, content []byte) {
	slog.Info("Message received", slog.String("method", method))

	switch method {
	case "initialize":
		var request lsp.InitializeRequest
		err := json.Unmarshal(content, &request)
		if err != nil {
			slog.Error("Error unmarshalling content", slog.String("method", method), slog.String("err", err.Error()))
		}

		slog.Info("Connected to client",
			slog.String("name", request.Params.ClientInfo.Name),
			slog.String("version", request.Params.ClientInfo.Version),
		)

		response := lsp.NewInitializeResponse(request.ID)
		encodedResponse := rpc.EncodeMessage(response)
		n, err := fmt.Fprint(writer, encodedResponse)
		if err != nil {
			slog.Error("Error responding to intitialize", slog.String("err", err.Error()))
			return
		}

		slog.Info("Response sent",
			slog.String("method", method),
			slog.Int("bytes written", n),
		)

	case "textDocument/didOpen":
		var notification lsp.DidOpenTextDocumentNotification
		err := json.Unmarshal(content, &notification)
		if err != nil {
			slog.Error("Error unmarshalling content", slog.String("method", method), slog.String("err", err.Error()))
		}

		diagnostics := state.DidOpen(notification.Params)
		if len(diagnostics) > 0 {
			response := lsp.PublishDiagnosticsNotification{
				Notification: lsp.Notification{
					Message: lsp.Message{JSONRPC: "2.0"},
					Method:  "textDocument/publishDiagnostics",
				},
				Params: lsp.PublishDiagnosticsParams{
					URI:         notification.Params.TextDocument.URI,
					Version:     notification.Params.TextDocument.Version,
					Diagnostics: diagnostics,
				},
			}

			n, err := fmt.Fprint(writer, response)
			if err != nil {
				slog.Error("Error responding to didChange", slog.String("err", err.Error()))
				return
			}

			slog.Info("Response sent",
				slog.String("method", method),
				slog.Int("bytes written", n),
			)
		}

		slog.Info("document opened", slog.String("URI", notification.Params.TextDocument.URI))

	case "textDocument/didChange":
		var notification lsp.DidChangeTextDocumentNotification
		err := json.Unmarshal(content, &notification)
		if err != nil {
			slog.Error("Error unmarshalling content", slog.String("method", method), slog.String("err", err.Error()))
		}

		diagnostics := state.DidChange(notification.Params)
		if len(diagnostics) > 0 {
			response := lsp.PublishDiagnosticsNotification{
				Notification: lsp.Notification{
					Message: lsp.Message{JSONRPC: "2.0"},
					Method:  "textDocument/publishDiagnostics",
				},
				Params: lsp.PublishDiagnosticsParams{
					URI:         notification.Params.TextDocument.URI,
					Version:     notification.Params.TextDocument.Version,
					Diagnostics: diagnostics,
				},
			}

			n, err := fmt.Fprint(writer, response)
			if err != nil {
				slog.Error("Error responding to didChange", slog.String("err", err.Error()))
				return
			}

			slog.Info("Response sent",
				slog.String("method", method),
				slog.Int("bytes written", n),
			)
		}

		slog.Info("document changed", slog.String("URI", notification.Params.TextDocument.URI))

	case "textDocument/hover":
		var request lsp.HoverRequest
		err := json.Unmarshal(content, &request)
		if err != nil {
			slog.Error("Error unmarshalling content", slog.String("method", method), slog.String("err", err.Error()))
		}

		response := state.Hover(request)
		rpc.EncodeMessage(response)
		encodedResponse := rpc.EncodeMessage(response)
		n, err := fmt.Fprint(writer, encodedResponse)
		if err != nil {
			slog.Error("Error responding to hover", slog.String("err", err.Error()))
			return
		}

		slog.Info("Response sent",
			slog.String("method", method),
			slog.Int("bytes written", n),
		)

	case "textDocument/definition":
		var request lsp.DefinitionRequest
		err := json.Unmarshal(content, &request)
		if err != nil {
			slog.Error("Error unmarshalling content", slog.String("method", method), slog.String("err", err.Error()))
		}

		response := state.Definition(request)
		rpc.EncodeMessage(response)
		encodedResponse := rpc.EncodeMessage(response)
		n, err := fmt.Fprint(writer, encodedResponse)
		if err != nil {
			slog.Error("Error responding to definition", slog.String("err", err.Error()))
			return
		}

		slog.Info("Response sent",
			slog.String("method", method),
			slog.Int("bytes written", n),
		)

	case "textDocument/completion":
		var request lsp.CompletionRequest
		err := json.Unmarshal(content, &request)
		if err != nil {
			slog.Error("Error unmarshalling content", slog.String("method", method), slog.String("err", err.Error()))
		}

		response := state.Completion(request)
		rpc.EncodeMessage(response)
		encodedResponse := rpc.EncodeMessage(response)
		n, err := fmt.Fprint(writer, encodedResponse)
		if err != nil {
			slog.Error("Error responding to completion", slog.String("err", err.Error()))
			return
		}

		slog.Info("Response sent",
			slog.String("method", method),
			slog.Int("bytes written", n),
		)
	}
}
