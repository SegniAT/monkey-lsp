package protocol

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
	Label         string        `json:"label"`
	Detail        string        `json:"detail"`
	Documentation MarkupContent `json:"documentation"`
}
