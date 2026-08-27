package rpc_test

import (
	"bytes"
	"io"
	"testing"
	"testing/iotest"

	"github.com/SegniAT/monkey-lsp/rpc"
)

func TestWriteMessage(t *testing.T) {
	tests := map[string]struct {
		setup    func() (out *bytes.Buffer, dest io.Writer)
		input    []byte
		expected []byte
	}{
		"Basic string": {
			setup: func() (*bytes.Buffer, io.Writer) {
				buf := bytes.NewBuffer(nil)
				return buf, buf
			},
			input:    []byte("hello"),
			expected: []byte("Content-Length: 5\r\n\r\nhello"),
		},
		"Truncated": {
			setup: func() (*bytes.Buffer, io.Writer) {
				buf := bytes.NewBuffer(nil)
				return buf, iotest.TruncateWriter(buf, 3)
			},
			input:    []byte("hello"),
			expected: []byte("Con"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			buf, dest := test.setup()

			writer := rpc.NewWriter(dest)
			_, err := writer.WriteMessage(test.input)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(buf.Bytes(), test.expected) {
				t.Errorf("expected: %q\ngot: %q", string(test.expected), buf.String())
			}
		})
	}
}
