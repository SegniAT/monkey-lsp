package protocol

type InitializeRequest struct {
	Request
	Params InitializeParams `json:"params"`
}

type InitializeParams struct {
	ClientInfo *ClientInfo `json:"clientinfo,omitempty"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResponse struct {
	Response
	Result InitializeResult `json:"result"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   ServerInfo         `json:"serverInfo"`
}

type ServerCapabilities struct {
	TextDocumentSync   int  `json:"textDocumentSync"`
	HoverProvider      bool `json:"hoverProvider"`
	DefinitionProvider bool `json:"definitionProvider"`
	// CodeActionProvider bool `json:"codeActionProvider"`
	CompletionProvider map[string]any `json:"completionProvider"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func NewInitializeResponse(id int64) InitializeResponse {
	return InitializeResponse{
		Response: Response{
			ID: &id,
			Message: Message{
				JSONRPC: "2.0",
			},
		},
		Result: InitializeResult{
			Capabilities: ServerCapabilities{
				TextDocumentSync:   1, // Documents are synced by always sending the full content of the document.
				HoverProvider:      true,
				DefinitionProvider: true,
				//CodeActionProvider: true,
				CompletionProvider: map[string]any{},
			},
			ServerInfo: ServerInfo{
				Name:    "monkey-lsp",
				Version: "0.0.1",
			},
		},
	}
}
