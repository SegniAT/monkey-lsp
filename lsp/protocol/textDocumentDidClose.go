package protocol

type DidCloseTextDocumentNotification struct {
	Notification
	Params DidCloseTextDocumentParams `json:"params"`
}

type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}
