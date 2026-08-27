package rpc

import "errors"

var (
	ErrSeparatorNotFound          = errors.New("content separator not found")
	ErrInvalidContentLength       = errors.New("invalid content length")
	ErrIncompleteMessage          = errors.New("incomplete content")
	ErrMissingHeaderContentLength = errors.New("missing Content-Length header")
)
