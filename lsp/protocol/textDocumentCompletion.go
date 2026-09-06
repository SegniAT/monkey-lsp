package protocol

type CompletionItemKind uint8

const (
	Function CompletionItemKind = 3
	Variable CompletionItemKind = 6
	Keyword  CompletionItemKind = 14
)

type CompletionRequest struct {
	Request
	Params CompletionParams `json:"params"`
}

type CompletionParams struct {
	TextDocumentPositionParams
}

type CompletionResponse struct {
	Response
	Result []CompletionItem `json:"result"`
	//Error  any        `json:"error"`
}

type CompletionItem struct {
	Label         string              `json:"label"`
	Detail        string              `json:"detail"`
	Kind          *CompletionItemKind `json:"kind,omitzero"`
	Documentation MarkupContent       `json:"documentation"`
}
