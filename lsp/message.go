package lsp

type Message struct {
	JSONRPC string `json:"jsonrpc"`
}

type Request struct {
	Message
	ID     int64  `json:"id"`
	Method string `json:"method"`
}

type Response struct {
	Message
	ID    *int64         `json:"id,omitempty"`
	Error *ResponseError `json:"error,omitempty"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type Notification struct {
	Message
	Method string `json:"method"`
}
