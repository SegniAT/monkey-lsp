package protocol

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

type ErrorCode int

const (
	// Defined by JSON-RPC
	ErrParseError     ErrorCode = -32700
	ErrInvalidRequest ErrorCode = -32600
	ErrMethodNotFound ErrorCode = -32601
	ErrInvalidParams  ErrorCode = -32602
	ErrInternalError  ErrorCode = -32603

	ErrRequestFailed ErrorCode = -32803
)

type ResponseError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Data    any       `json:"data,omitempty"`
}

type Notification struct {
	Message
	Method string `json:"method"`
}
