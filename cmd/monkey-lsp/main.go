package main

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/SegniAT/monkey-lsp/lsp/server"
)

func main() {
	file, err := os.OpenFile(filepath.Join(os.TempDir(), "monkey-lsp.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		panic(err)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(file, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})))

	server := server.NewServer(os.Stdin, os.Stdout)
	slog.Info("Started LSP server")
	err = server.Run()
	if err != nil {
		// TODO: Send an error to the client?
		log.Fatalf("Server stopped: %v", err)
	}

}
