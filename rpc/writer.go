package rpc

import (
	"bytes"
	"fmt"
	"io"
)

type Writer struct {
	w io.Writer
}

func NewWriter(writer io.Writer) *Writer {
	return &Writer{
		w: writer,
	}
}

func (w *Writer) WriteMessage(content []byte) (int, error) {
	contentLength := len(content)

	var buf bytes.Buffer
	buf.Grow(32 + contentLength)

	fmt.Fprintf(&buf, "Content-Length: %d\r\n\r\n", contentLength)
	buf.Write(content)

	return w.w.Write(buf.Bytes())
}
