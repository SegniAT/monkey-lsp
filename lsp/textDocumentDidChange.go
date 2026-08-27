package lsp

type DidChangeTextDocumentNotification struct {
	Notification
	Params DidChangeTextDocumentParams `json:"params"`
}

type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier          `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeWholeDocument `json:"contentChanges,omitempty"`
}

type TextDocumentContentChangeWholeDocument struct {
	Text string `json:"text"`
}
